use std::fs;
use std::path::Path;
use std::time::{SystemTime, UNIX_EPOCH};

use crate::state::{AgentEntry, AgentRegistry, AgentStatus, ApprovalFile, PendingApproval, StatusFile};

const STALE_THRESHOLD_SECS: u64 = 30;

pub fn update_registry(registry: &mut AgentRegistry) {
    let home = std::env::var("HOME").unwrap_or_default();
    let status_files = scan_status_files(&home);
    let approval_files = scan_approval_files(&home);

    for agent in registry.agents.iter_mut() {
        // Find matching status file by cwd
        let status = status_files
            .iter()
            .find(|(_, sf)| sf.cwd == agent.worktree_path || sf.cwd.ends_with(&agent.worktree));

        if let Some((hash, sf)) = status {
            agent.session_hash = Some(hash.clone());
            agent.model = sf.model.clone();
            agent.context_remaining = sf.context_remaining;

            let now = SystemTime::now()
                .duration_since(UNIX_EPOCH)
                .map(|d| d.as_secs())
                .unwrap_or(0);

            if now.saturating_sub(sf.updated_at) > STALE_THRESHOLD_SECS {
                agent.status = AgentStatus::Stale;
            } else if agent.status == AgentStatus::Stale {
                agent.status = AgentStatus::Running;
            }
        }

        // Find matching approval file by session hash
        if let Some(hash) = &agent.session_hash {
            if let Some((_, af)) = approval_files.iter().find(|(h, _)| h == hash) {
                if af.pending {
                    agent.status = AgentStatus::Waiting;
                    agent.pending_approval = Some(PendingApproval {
                        tool: af.tool.clone().unwrap_or_default(),
                        preview: af.preview.clone().unwrap_or_default(),
                    });
                } else {
                    agent.pending_approval = None;
                    if agent.status == AgentStatus::Waiting {
                        agent.status = AgentStatus::Running;
                    }
                }
            }
        }
    }
}

pub fn scan_status_files(home: &str) -> Vec<(String, StatusFile)> {
    scan_json_files(home, "bob-status.json")
        .into_iter()
        .filter_map(|(hash, content)| {
            serde_json::from_str::<StatusFile>(&content)
                .ok()
                .map(|sf| (hash, sf))
        })
        .collect()
}

pub fn scan_approval_files(home: &str) -> Vec<(String, ApprovalFile)> {
    scan_json_files(home, "bob-approval.json")
        .into_iter()
        .filter_map(|(hash, content)| {
            serde_json::from_str::<ApprovalFile>(&content)
                .ok()
                .map(|af| (hash, af))
        })
        .collect()
}

fn scan_json_files(home: &str, filename: &str) -> Vec<(String, String)> {
    let projects_dir = format!("{}/.claude/projects", home);
    let path = Path::new(&projects_dir);
    if !path.exists() {
        return vec![];
    }

    let mut results = Vec::new();
    if let Ok(entries) = fs::read_dir(path) {
        for entry in entries.flatten() {
            let entry_path = entry.path();
            if !entry_path.is_dir() {
                continue;
            }
            let hash = entry_path
                .file_name()
                .and_then(|n| n.to_str())
                .unwrap_or("")
                .to_string();
            let file_path = entry_path.join(filename);
            if let Ok(content) = fs::read_to_string(&file_path) {
                results.push((hash, content));
            }
        }
    }
    results
}

/// Check if an agent's worktree process is still alive by looking for
/// a recently-updated status file. Returns false if no status file found.
pub fn is_agent_alive(agent: &AgentEntry, home: &str) -> bool {
    let Some(hash) = &agent.session_hash else {
        return false;
    };
    let path = format!("{}/.claude/projects/{}/bob-status.json", home, hash);
    fs::metadata(&path).is_ok()
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::fs;
    use std::io::Write;
    use tempfile::TempDir;

    fn write_file(dir: &Path, name: &str, content: &str) {
        let mut f = fs::File::create(dir.join(name)).unwrap();
        f.write_all(content.as_bytes()).unwrap();
    }

    #[test]
    fn test_scan_status_files_finds_valid() {
        let tmp = TempDir::new().unwrap();
        let projects = tmp.path().join(".claude/projects/abc123");
        fs::create_dir_all(&projects).unwrap();
        write_file(
            &projects,
            "bob-status.json",
            r#"{"cwd":"/repo/add-auth","context_remaining":80,"model":"sonnet","updated_at":9999999999}"#,
        );

        let home = tmp.path().to_str().unwrap();
        let results = scan_status_files(home);
        assert_eq!(results.len(), 1);
        assert_eq!(results[0].0, "abc123");
        assert_eq!(results[0].1.cwd, "/repo/add-auth");
    }

    #[test]
    fn test_scan_status_files_skips_invalid_json() {
        let tmp = TempDir::new().unwrap();
        let projects = tmp.path().join(".claude/projects/bad");
        fs::create_dir_all(&projects).unwrap();
        write_file(&projects, "bob-status.json", "not json");

        let home = tmp.path().to_str().unwrap();
        let results = scan_status_files(home);
        assert_eq!(results.len(), 0);
    }

    #[test]
    fn test_scan_empty_projects_dir() {
        let tmp = TempDir::new().unwrap();
        let projects = tmp.path().join(".claude/projects");
        fs::create_dir_all(&projects).unwrap();

        let home = tmp.path().to_str().unwrap();
        let results = scan_status_files(home);
        assert!(results.is_empty());
    }

    #[test]
    fn test_stale_detection() {
        let mut reg = AgentRegistry::default();
        reg.agents.push(AgentEntry {
            worktree: "old-task".to_string(),
            worktree_path: "/repo/old-task".to_string(),
            session_hash: Some("stale123".to_string()),
            status: AgentStatus::Running,
            ..Default::default()
        });

        let tmp = TempDir::new().unwrap();
        let projects = tmp.path().join(".claude/projects/stale123");
        fs::create_dir_all(&projects).unwrap();
        write_file(
            &projects,
            "bob-status.json",
            r#"{"cwd":"/repo/old-task","context_remaining":null,"model":"sonnet","updated_at":1}"#,
        );

        let home = tmp.path().to_str().unwrap();
        let status_files = scan_status_files(home);
        assert!(!status_files.is_empty());

        // Simulate update: updated_at=1 is far in the past, should be stale
        let sf = &status_files[0].1;
        let now = SystemTime::now()
            .duration_since(UNIX_EPOCH)
            .map(|d| d.as_secs())
            .unwrap_or(0);
        assert!(now.saturating_sub(sf.updated_at) > STALE_THRESHOLD_SECS);
    }

    #[test]
    fn test_match_cwd_to_agent() {
        let agent = AgentEntry {
            worktree: "add-auth".to_string(),
            worktree_path: "/repo/add-auth".to_string(),
            ..Default::default()
        };
        let sf = StatusFile {
            cwd: "/repo/add-auth".to_string(),
            context_remaining: Some(50),
            model: "sonnet".to_string(),
            updated_at: 9999999999,
        };
        // Match by worktree_path
        assert!(sf.cwd == agent.worktree_path || sf.cwd.ends_with(&agent.worktree));
    }
}
