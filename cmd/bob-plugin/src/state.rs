use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, PartialEq)]
pub enum AgentStatus {
    Running,
    Idle,
    Waiting,
    Error,
    Stale,
}

impl Default for AgentStatus {
    fn default() -> Self {
        AgentStatus::Idle
    }
}

impl AgentStatus {
    pub fn label(&self) -> &str {
        match self {
            AgentStatus::Running => "running",
            AgentStatus::Idle => "idle",
            AgentStatus::Waiting => "waiting",
            AgentStatus::Error => "error",
            AgentStatus::Stale => "stale",
        }
    }
}

#[derive(Debug, Clone, Default)]
pub struct PendingApproval {
    pub tool: String,
    pub preview: String,
}

#[derive(Debug, Clone, Default)]
pub struct AgentEntry {
    pub worktree: String,
    pub worktree_path: String,
    pub model: String,
    pub status: AgentStatus,
    pub last_action: String,
    pub context_remaining: Option<u8>,
    pub pending_approval: Option<PendingApproval>,
    pub zellij_tab_index: usize,
    pub session_hash: Option<String>,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct StatusFile {
    pub cwd: String,
    pub context_remaining: Option<u8>,
    pub model: String,
    pub updated_at: u64,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct ApprovalFile {
    pub pending: bool,
    pub tool: Option<String>,
    pub preview: Option<String>,
}

#[derive(Debug, Default)]
pub struct AgentRegistry {
    pub agents: Vec<AgentEntry>,
    pub selected: usize,
}

impl AgentRegistry {
    pub fn select_next(&mut self) {
        if !self.agents.is_empty() {
            self.selected = (self.selected + 1).min(self.agents.len() - 1);
        }
    }

    pub fn select_prev(&mut self) {
        self.selected = self.selected.saturating_sub(1);
    }

    pub fn selected_agent(&self) -> Option<&AgentEntry> {
        self.agents.get(self.selected)
    }

    pub fn selected_agent_mut(&mut self) -> Option<&mut AgentEntry> {
        self.agents.get_mut(self.selected)
    }

    /// Map a sidebar row number to an agent index.
    /// Each agent occupies 4 rows: name, status, approval/blank, divider.
    pub fn row_to_agent_index(&self, row: usize) -> Option<usize> {
        let idx = row / 4;
        if idx < self.agents.len() {
            Some(idx)
        } else {
            None
        }
    }
}

/// Derive a URL/filesystem-safe slug from a prompt string.
/// Takes first 40 chars, lowercases, replaces non-alphanumeric with `-`,
/// collapses consecutive dashes, trims trailing dashes.
pub fn slugify(prompt: &str) -> String {
    let slug: String = prompt
        .chars()
        .take(60)
        .map(|c| if c.is_alphanumeric() { c.to_ascii_lowercase() } else { '-' })
        .collect();

    // Collapse consecutive dashes
    let mut result = String::new();
    let mut last_dash = false;
    for c in slug.chars() {
        if c == '-' {
            if !last_dash {
                result.push(c);
            }
            last_dash = true;
        } else {
            result.push(c);
            last_dash = false;
        }
    }

    let result = result.trim_matches('-').to_string();
    // Truncate to 40 chars at a dash boundary if possible
    if result.len() <= 40 {
        return result;
    }
    let truncated = &result[..40];
    if let Some(pos) = truncated.rfind('-') {
        truncated[..pos].to_string()
    } else {
        truncated.to_string()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_agent_entry_default_status() {
        let entry = AgentEntry::default();
        assert_eq!(entry.status, AgentStatus::Idle);
    }

    #[test]
    fn test_registry_add_remove() {
        let mut reg = AgentRegistry::default();
        assert_eq!(reg.agents.len(), 0);

        reg.agents.push(AgentEntry {
            worktree: "add-auth".to_string(),
            ..Default::default()
        });
        reg.agents.push(AgentEntry {
            worktree: "fix-bug".to_string(),
            ..Default::default()
        });
        assert_eq!(reg.agents.len(), 2);

        reg.agents.remove(0);
        assert_eq!(reg.agents.len(), 1);
        assert_eq!(reg.agents[0].worktree, "fix-bug");
    }

    #[test]
    fn test_select_next_prev() {
        let mut reg = AgentRegistry::default();
        for i in 0..3 {
            reg.agents.push(AgentEntry {
                worktree: format!("agent-{}", i),
                ..Default::default()
            });
        }
        assert_eq!(reg.selected, 0);
        reg.select_next();
        assert_eq!(reg.selected, 1);
        reg.select_next();
        assert_eq!(reg.selected, 2);
        reg.select_next(); // at end, should not overflow
        assert_eq!(reg.selected, 2);
        reg.select_prev();
        assert_eq!(reg.selected, 1);
        reg.select_prev();
        assert_eq!(reg.selected, 0);
        reg.select_prev(); // at start, should not underflow
        assert_eq!(reg.selected, 0);
    }

    #[test]
    fn test_slug_from_prompt() {
        assert_eq!(slugify("add user auth feature"), "add-user-auth-feature");
        assert_eq!(slugify("Add User Auth! Feature"), "add-user-auth-feature");
        assert_eq!(slugify("  leading spaces  "), "leading-spaces");
    }

    #[test]
    fn test_slug_truncates_at_40() {
        let long = "a".repeat(50);
        assert!(slugify(&long).len() <= 40);
    }

    #[test]
    fn test_slug_collapses_dashes() {
        assert_eq!(slugify("hello   world"), "hello-world");
        assert_eq!(slugify("hello---world"), "hello-world");
    }

    #[test]
    fn test_status_file_parse_valid() {
        let json = r#"{
            "cwd": "/home/user/repo-worktrees/add-auth",
            "context_remaining": 72,
            "model": "claude-sonnet-4-6",
            "updated_at": 1709500000
        }"#;
        let sf: StatusFile = serde_json::from_str(json).unwrap();
        assert_eq!(sf.cwd, "/home/user/repo-worktrees/add-auth");
        assert_eq!(sf.context_remaining, Some(72));
        assert_eq!(sf.model, "claude-sonnet-4-6");
        assert_eq!(sf.updated_at, 1709500000);
    }

    #[test]
    fn test_status_file_parse_missing_context() {
        let json = r#"{
            "cwd": "/tmp/test",
            "context_remaining": null,
            "model": "unknown",
            "updated_at": 0
        }"#;
        let sf: StatusFile = serde_json::from_str(json).unwrap();
        assert_eq!(sf.context_remaining, None);
    }

    #[test]
    fn test_row_to_agent_index() {
        let mut reg = AgentRegistry::default();
        for i in 0..3 {
            reg.agents.push(AgentEntry {
                worktree: format!("a-{}", i),
                ..Default::default()
            });
        }
        assert_eq!(reg.row_to_agent_index(0), Some(0));
        assert_eq!(reg.row_to_agent_index(3), Some(0));
        assert_eq!(reg.row_to_agent_index(4), Some(1));
        assert_eq!(reg.row_to_agent_index(8), Some(2));
        assert_eq!(reg.row_to_agent_index(12), None);
    }
}
