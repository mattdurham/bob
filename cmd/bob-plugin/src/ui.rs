use zellij_tile::prelude::*;

use crate::state::{AgentEntry, AgentRegistry, AgentStatus};

/// Rows occupied by each agent entry in the sidebar (name + status + approval + divider).
/// Must match the render loop in render_sidebar and row_to_agent_index in state.rs.
pub const ROWS_PER_AGENT: usize = 4;

const MODELS: &[(&str, &str)] = &[
    ("sonnet", "claude-sonnet-4-6"),
    ("opus", "claude-opus-4-6"),
    ("haiku", "claude-haiku-4-5-20251001"),
];

#[derive(Debug, Default)]
pub struct ModalState {
    pub prompt: String,
    pub model_index: usize,
}

impl ModalState {
    pub fn model(&self) -> &str {
        MODELS[self.model_index % MODELS.len()].1
    }

    pub fn model_label(&self) -> &str {
        MODELS[self.model_index % MODELS.len()].0
    }

    pub fn cycle_model(&mut self) {
        self.model_index = (self.model_index + 1) % MODELS.len();
    }

    pub fn push_char(&mut self, c: char) {
        self.prompt.push(c);
    }

    pub fn backspace(&mut self) {
        self.prompt.pop();
    }
}

/// Render the agent sidebar into the plugin pane.
pub fn render_sidebar(registry: &AgentRegistry, rows: usize, cols: usize) {
    // Need at least ROWS_PER_AGENT + 1 rows (1 for [+ new]) to render anything useful
    if rows < ROWS_PER_AGENT + 1 {
        return;
    }

    // Bottom: 1 row for [+ new]
    let visible_rows = rows.saturating_sub(1);
    let agents_visible = visible_rows / ROWS_PER_AGENT;

    // Scroll window to keep selected in view
    let scroll_offset = if registry.agents.is_empty() {
        0
    } else {
        let selected = registry.selected;
        let window_start = (selected / agents_visible.max(1)) * agents_visible.max(1);
        window_start
    };

    let divider = "─".repeat(cols.saturating_sub(1));

    let mut row = 0usize;
    for (i, agent) in registry.agents.iter().enumerate().skip(scroll_offset) {
        if row + ROWS_PER_AGENT > visible_rows {
            break;
        }
        let is_selected = i == registry.selected;
        render_agent_entry(agent, is_selected, cols, row);
        // Divider
        print_text_with_coordinates(
            Text::new(&divider).color_range(2, 0..divider.len()),
            0,
            row + 3,
            None,
            None,
        );
        row += ROWS_PER_AGENT;
    }

    // [+ new] at bottom
    let new_label = "  [+ new]";
    print_text_with_coordinates(
        Text::new(new_label).color_range(3, 2..new_label.len()),
        0,
        rows.saturating_sub(1),
        None,
        None,
    );
}

fn render_agent_entry(agent: &AgentEntry, selected: bool, cols: usize, row: usize) {
    let prefix = if selected { "> " } else { "  " };
    let name = truncate(&agent.worktree, cols.saturating_sub(2));
    let name_line = format!("{}{}", prefix, name);

    // Status line: "model • last_action" or "model • status"
    let action = if agent.last_action.is_empty() {
        agent.status.label().to_string()
    } else {
        agent.last_action.clone()
    };
    let status_content = format!("  {} • {}", agent.model_short(), action);
    let status_line = truncate(&status_content, cols);

    // Approval line
    let approval_line = if let Some(ref ap) = agent.pending_approval {
        let text = format!("  ⚠ {} {}", ap.tool, ap.preview);
        truncate(&text, cols)
    } else {
        String::new()
    };

    // Render name (selected = bold/highlighted)
    let name_text = if selected {
        Text::new(&name_line).color_range(0, 0..name_line.len())
    } else {
        Text::new(&name_line)
    };
    print_text_with_coordinates(name_text, 0, row, None, None);

    // Render status
    let status_color = match agent.status {
        AgentStatus::Running => 2,
        AgentStatus::Idle => 3,
        AgentStatus::Waiting => 1,
        AgentStatus::Error => 1,
        AgentStatus::Stale => 3,
    };
    print_text_with_coordinates(
        Text::new(&status_line).color_range(status_color, 0..status_line.len()),
        0,
        row + 1,
        None,
        None,
    );

    // Render approval (row + 2)
    if !approval_line.is_empty() {
        print_text_with_coordinates(
            Text::new(&approval_line).color_range(1, 0..approval_line.len()),
            0,
            row + 2,
            None,
            None,
        );
    }
}

/// Render the spawn modal as a centered overlay.
pub fn render_modal(modal: &ModalState, rows: usize, cols: usize) {
    let modal_width = (cols * 2 / 3).max(30).min(60);
    let modal_height = 10usize;
    let start_col = (cols.saturating_sub(modal_width)) / 2;
    let start_row = (rows.saturating_sub(modal_height)) / 2;

    let border_top = format!("╔{}╗", "═".repeat(modal_width.saturating_sub(2)));
    let border_bot = format!("╚{}╝", "═".repeat(modal_width.saturating_sub(2)));
    let empty_row = format!("║{}║", " ".repeat(modal_width.saturating_sub(2)));

    print_text_with_coordinates(Text::new(&border_top), start_col, start_row, None, None);
    print_text_with_coordinates(
        Text::new(&format!("║  New Agent{}║", " ".repeat(modal_width.saturating_sub(13)))),
        start_col,
        start_row + 1,
        None,
        None,
    );
    print_text_with_coordinates(Text::new(&empty_row), start_col, start_row + 2, None, None);

    let model_line = format!(
        "║  Model:  [ {:<8} ]{}║",
        modal.model_label(),
        " ".repeat(modal_width.saturating_sub(24))
    );
    print_text_with_coordinates(Text::new(&model_line), start_col, start_row + 3, None, None);
    print_text_with_coordinates(Text::new(&empty_row), start_col, start_row + 4, None, None);

    let hint = "║  Tab=model  Enter=spawn  Esc=cancel";
    let hint_line = format!("{}{} ║", hint, " ".repeat(modal_width.saturating_sub(hint.len() + 2).max(0)));
    print_text_with_coordinates(Text::new(&hint_line), start_col, start_row + 5, None, None);

    let inner_width = modal_width.saturating_sub(4);
    let prompt_display = if modal.prompt.len() <= inner_width {
        format!("{}_", modal.prompt)
    } else {
        let start = modal.prompt.len().saturating_sub(inner_width);
        format!("{}_", &modal.prompt[start..])
    };
    let prompt_line = format!(
        "║  {:<width$}  ║",
        prompt_display,
        width = inner_width
    );
    print_text_with_coordinates(
        Text::new(&prompt_line).color_range(0, 3..3 + prompt_display.len()),
        start_col,
        start_row + 6,
        None,
        None,
    );

    print_text_with_coordinates(Text::new(&empty_row), start_col, start_row + 7, None, None);

    let spawn_label = "[ Spawn ]  [ Cancel ]";
    let spawn_line = format!(
        "║  {}{}║",
        spawn_label,
        " ".repeat(modal_width.saturating_sub(spawn_label.len() + 4))
    );
    print_text_with_coordinates(Text::new(&spawn_line), start_col, start_row + 8, None, None);
    print_text_with_coordinates(Text::new(&border_bot), start_col, start_row + 9, None, None);
}

fn truncate(s: &str, max_len: usize) -> String {
    if s.chars().count() <= max_len {
        s.to_string()
    } else {
        let truncated: String = s.chars().take(max_len.saturating_sub(1)).collect();
        format!("{}…", truncated)
    }
}

// Extension trait for display model name
trait ModelShort {
    fn model_short(&self) -> String;
}

impl ModelShort for AgentEntry {
    fn model_short(&self) -> String {
        if self.model.contains("sonnet") {
            "sonnet".to_string()
        } else if self.model.contains("opus") {
            "opus".to_string()
        } else if self.model.contains("haiku") {
            "haiku".to_string()
        } else if self.model.is_empty() {
            "?".to_string()
        } else {
            self.model.chars().take(8).collect()
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::state::{AgentEntry, PendingApproval};

    #[test]
    fn test_truncate_short_string() {
        assert_eq!(truncate("hello", 10), "hello");
    }

    #[test]
    fn test_truncate_long_string() {
        let result = truncate("hello world this is long", 10);
        // chars().count() because '…' is multi-byte; truncated to 9 chars + ellipsis = 10
        assert!(result.chars().count() <= 10);
        assert!(result.ends_with('…'));
    }

    #[test]
    fn test_truncate_exact_length() {
        assert_eq!(truncate("hello", 5), "hello");
    }

    #[test]
    fn test_modal_default() {
        let m = ModalState::default();
        assert_eq!(m.prompt, "");
        assert_eq!(m.model_index, 0);
        assert_eq!(m.model_label(), "sonnet");
    }

    #[test]
    fn test_modal_cycle_model() {
        let mut m = ModalState::default();
        assert_eq!(m.model_label(), "sonnet");
        m.cycle_model();
        assert_eq!(m.model_label(), "opus");
        m.cycle_model();
        assert_eq!(m.model_label(), "haiku");
        m.cycle_model();
        assert_eq!(m.model_label(), "sonnet"); // wraps around
    }

    #[test]
    fn test_modal_push_backspace() {
        let mut m = ModalState::default();
        m.push_char('h');
        m.push_char('i');
        assert_eq!(m.prompt, "hi");
        m.backspace();
        assert_eq!(m.prompt, "h");
        m.backspace();
        assert_eq!(m.prompt, "");
        m.backspace(); // no panic on empty
        assert_eq!(m.prompt, "");
    }

    #[test]
    fn test_model_short_sonnet() {
        let entry = AgentEntry {
            model: "claude-sonnet-4-6".to_string(),
            ..Default::default()
        };
        assert_eq!(entry.model_short(), "sonnet");
    }

    #[test]
    fn test_model_short_unknown() {
        let entry = AgentEntry {
            model: "".to_string(),
            ..Default::default()
        };
        assert_eq!(entry.model_short(), "?");
    }

    #[test]
    fn test_approval_indicator_format() {
        let ap = PendingApproval {
            tool: "Bash".to_string(),
            preview: "git push origin main".to_string(),
        };
        let line = format!("  ⚠ {} {}", ap.tool, ap.preview);
        assert!(line.contains("⚠"));
        assert!(line.contains("Bash"));
        assert!(line.contains("git push"));
    }
}
