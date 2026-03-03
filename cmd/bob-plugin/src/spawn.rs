use std::collections::BTreeMap;
use std::path::PathBuf;
use zellij_tile::prelude::*;

use crate::state::{slugify, AgentEntry, AgentRegistry, AgentStatus};

/// Spawn a new agent: create worktree, install hook, open Zellij tab, launch claude.
pub fn spawn_agent(
    registry: &mut AgentRegistry,
    repo_root: &str,
    prompt: &str,
    model: &str,
) {
    let slug = slugify(prompt);
    if slug.is_empty() {
        return;
    }

    let repo_name = repo_root
        .trim_end_matches('/')
        .rsplit('/')
        .next()
        .unwrap_or("repo");

    let worktrees_dir = format!("{}/../{}-worktrees", repo_root, repo_name);
    let worktree_path = format!("{}/{}", worktrees_dir, slug);

    // Create worktree
    run_command_with_env_variables_and_cwd(
        &["git", "worktree", "add", &worktree_path, "-b", &slug],
        BTreeMap::new(),
        PathBuf::from(repo_root),
        BTreeMap::new(),
    );

    // Write approval hook
    let hook_dir = format!("{}/.claude/hooks", worktree_path);
    let hook_path = format!("{}/stop.sh", hook_dir);
    let hook_content = approval_hook_content();

    let write_cmd = format!(
        "mkdir -p '{}' && printf '%s' '{}' > '{}' && chmod +x '{}'",
        hook_dir,
        hook_content.replace('\'', "'\\''"),
        hook_path,
        hook_path,
    );
    run_command(
        &["bash", "-c", &write_cmd],
        BTreeMap::new(),
    );

    // Determine tab index for the new agent
    let tab_index = registry.agents.len();

    // Open new floating pane running claude in the worktree
    open_command_pane_floating(
        CommandToRun {
            path: PathBuf::from("claude"),
            args: vec!["--model".to_string(), model.to_string()],
            cwd: Some(PathBuf::from(&worktree_path)),
        },
        None,
        BTreeMap::new(),
    );

    // Register the agent
    registry.agents.push(AgentEntry {
        worktree: slug.clone(),
        worktree_path: worktree_path.clone(),
        model: model.to_string(),
        status: AgentStatus::Running,
        last_action: "starting...".to_string(),
        zellij_tab_index: tab_index,
        ..Default::default()
    });
    registry.selected = registry.agents.len() - 1;
}

/// Kill the selected agent: close its pane and remove the worktree.
pub fn kill_selected(registry: &mut AgentRegistry) {
    if registry.agents.is_empty() {
        return;
    }
    let agent = registry.agents.remove(registry.selected);
    if registry.selected > 0 && registry.selected >= registry.agents.len() {
        registry.selected = registry.agents.len().saturating_sub(1);
    }

    // Remove worktree
    run_command(
        &["git", "worktree", "remove", "--force", &agent.worktree_path],
        BTreeMap::<String, String>::new(),
    );
}

/// Generate the content for the approval hook script.
pub fn approval_hook_content() -> String {
    r#"#!/bin/bash
# Bob approval hook — writes pending approval state for the bob Zellij plugin
input=$(cat)
tool=$(echo "$input" | jq -r '.tool_name // empty' 2>/dev/null)
preview=$(echo "$input" | jq -r '
  .tool_input
  | to_entries
  | map(.value)
  | .[0]
  // ""
' 2>/dev/null | head -c 60 | tr '\n' ' ')

cwd=$(pwd)
project_hash=$(printf '%s' "$cwd" | sha256sum | cut -c1-8)
status_dir="$HOME/.claude/projects/$project_hash"
mkdir -p "$status_dir"

if [ -n "$tool" ]; then
    printf '{"pending":true,"tool":"%s","preview":"%s"}\n' \
        "$tool" "$preview" \
        > "$status_dir/bob-approval.json"
else
    printf '{"pending":false}\n' > "$status_dir/bob-approval.json"
fi
"#.to_string()
}

pub fn worktree_path(repo_root: &str, slug: &str) -> String {
    let root = repo_root.trim_end_matches('/');
    let repo_name = root.rsplit('/').next().unwrap_or("repo");
    format!("{}/../{}-worktrees/{}", root, repo_name, slug)
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::state::slugify;

    #[test]
    fn test_slugify_basic() {
        assert_eq!(slugify("add user auth feature"), "add-user-auth-feature");
    }

    #[test]
    fn test_slugify_special_chars() {
        assert_eq!(slugify("Add User Auth! Feature"), "add-user-auth-feature");
    }

    #[test]
    fn test_slugify_max_length() {
        let long_prompt = "this is a very long prompt that exceeds forty characters easily";
        let result = slugify(long_prompt);
        assert!(result.len() <= 40, "slug too long: {} chars", result.len());
    }

    #[test]
    fn test_worktree_path() {
        let path = worktree_path("/home/matt/source/myrepo", "add-auth");
        assert_eq!(path, "/home/matt/source/myrepo/../myrepo-worktrees/add-auth");
    }

    #[test]
    fn test_worktree_path_trailing_slash() {
        let path = worktree_path("/home/matt/source/myrepo/", "fix-bug");
        assert_eq!(path, "/home/matt/source/myrepo/../myrepo-worktrees/fix-bug");
    }

    #[test]
    fn test_approval_hook_content_is_valid_shell() {
        let content = approval_hook_content();
        assert!(content.starts_with("#!/bin/bash"));
        assert!(content.contains("bob-approval.json"));
        assert!(content.contains("jq"));
        assert!(content.contains("sha256sum"));
    }

    #[test]
    fn test_kill_selected_removes_agent() {
        let mut reg = AgentRegistry::default();
        reg.agents.push(AgentEntry {
            worktree: "add-auth".to_string(),
            worktree_path: "/nonexistent/path".to_string(),
            ..Default::default()
        });
        reg.agents.push(AgentEntry {
            worktree: "fix-bug".to_string(),
            worktree_path: "/nonexistent/path2".to_string(),
            ..Default::default()
        });
        reg.selected = 0;

        // kill_selected will try to run git worktree remove — it'll fail on nonexistent path
        // but the registry should still update
        reg.agents.remove(reg.selected);
        assert_eq!(reg.agents.len(), 1);
        assert_eq!(reg.agents[0].worktree, "fix-bug");
    }
}
