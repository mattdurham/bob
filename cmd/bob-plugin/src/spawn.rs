use std::collections::BTreeMap;
use std::path::PathBuf;
use zellij_tile::prelude::*;

use crate::state::{slugify, AgentEntry, AgentRegistry, AgentStatus};

// Embed the approval hook at compile time — single source of truth.
// Any edits to scripts/hooks/stop.sh are automatically reflected here.
const APPROVAL_HOOK: &str = include_str!("../../../scripts/hooks/stop.sh");

/// Spawn a new agent: create worktree, install hook, open a floating Claude pane.
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

    let wt_path = worktree_path(repo_root, &slug);

    // Create worktree
    run_command_with_env_variables_and_cwd(
        &["git", "worktree", "add", &wt_path, "-b", &slug],
        BTreeMap::new(),
        PathBuf::from(repo_root.trim_end_matches('/')),
        BTreeMap::new(),
    );

    // Verify worktree was actually created before registering the agent.
    // run_command_with_env_variables_and_cwd is fire-and-forget in Zellij's API;
    // we check the path to detect silent failures (branch exists, disk full, etc.).
    if !std::path::Path::new(&wt_path).exists() {
        // Worktree creation failed — don't register a ghost agent.
        // TODO: surface an error message in the plugin UI.
        return;
    }

    // Write approval hook using std::fs::write — no shell involved, no injection surface.
    let hook_dir = format!("{}/.claude/hooks", wt_path);
    let hook_path = format!("{}/stop.sh", hook_dir);
    if std::fs::create_dir_all(&hook_dir).is_ok() {
        if std::fs::write(&hook_path, APPROVAL_HOOK).is_ok() {
            // Set executable bit via chmod (cross-platform via run_command)
            run_command(&["chmod", "+x", &hook_path], BTreeMap::<String, String>::new());
        }
    }

    // Determine tab index for the new agent
    let tab_index = registry.agents.len();

    // Open a floating pane running claude in the new worktree
    open_command_pane_floating(
        CommandToRun {
            path: PathBuf::from("claude"),
            args: vec!["--model".to_string(), model.to_string()],
            cwd: Some(PathBuf::from(&wt_path)),
        },
        None,
        BTreeMap::new(),
    );

    registry.agents.push(AgentEntry {
        worktree: slug.clone(),
        worktree_path: wt_path,
        model: model.to_string(),
        status: AgentStatus::Running,
        last_action: "starting...".to_string(),
        zellij_tab_index: tab_index,
        ..Default::default()
    });
    registry.selected = registry.agents.len() - 1;
}

/// Kill the selected agent: remove it from the registry and remove the worktree.
pub fn kill_selected(registry: &mut AgentRegistry) {
    if registry.agents.is_empty() {
        return;
    }
    let agent = registry.agents.remove(registry.selected);
    // Clamp selected to the new valid range (handles the last-agent-at-0 case)
    registry.selected = registry.selected.min(registry.agents.len().saturating_sub(1));

    run_command(
        &["git", "worktree", "remove", "--force", &agent.worktree_path],
        BTreeMap::<String, String>::new(),
    );
}

/// Compute the worktree path for a given repo root and feature slug.
/// Trims trailing slashes from repo_root before deriving repo_name.
pub fn worktree_path(repo_root: &str, slug: &str) -> String {
    let root = repo_root.trim_end_matches('/');
    let repo_name = root.rsplit('/').next().unwrap_or("repo");
    format!("{}/../{}-worktrees/{}", root, repo_name, slug)
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::state::{slugify, AgentEntry, AgentRegistry};

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
    fn test_approval_hook_includes_jq_safe_construction() {
        // Verify the embedded hook uses jq -n for JSON construction, not bare printf
        assert!(APPROVAL_HOOK.starts_with("#!/bin/bash"));
        assert!(APPROVAL_HOOK.contains("bob-approval.json"));
        assert!(APPROVAL_HOOK.contains("jq -n --arg"));
        assert!(APPROVAL_HOOK.contains("tostring"));
    }

    #[test]
    fn test_kill_selected_last_agent() {
        let mut reg = AgentRegistry::default();
        reg.agents.push(AgentEntry {
            worktree: "only-agent".to_string(),
            worktree_path: "/nonexistent".to_string(),
            ..Default::default()
        });
        reg.selected = 0;
        reg.agents.remove(0);
        // Simulate the new clamp logic
        reg.selected = reg.selected.min(reg.agents.len().saturating_sub(1));
        assert_eq!(reg.selected, 0); // saturating_sub(1) on 0 = 0; safe
        assert!(reg.agents.is_empty());
    }

    #[test]
    fn test_kill_selected_middle_agent() {
        let mut reg = AgentRegistry::default();
        for name in &["a", "b", "c"] {
            reg.agents.push(AgentEntry {
                worktree: name.to_string(),
                worktree_path: "/nonexistent".to_string(),
                ..Default::default()
            });
        }
        reg.selected = 2; // select "c"
        reg.agents.remove(2);
        reg.selected = reg.selected.min(reg.agents.len().saturating_sub(1));
        assert_eq!(reg.selected, 1); // clamped to new last index
    }
}
