mod state;
mod poll;
mod ui;
mod spawn;

use state::AgentRegistry;
use zellij_tile::prelude::*;

#[derive(Default)]
struct BobPlugin {
    registry: AgentRegistry,
    modal: Option<ui::ModalState>,
    repo_root: String,
}

register_plugin!(BobPlugin);

impl ZellijPlugin for BobPlugin {
    fn load(&mut self, configuration: std::collections::BTreeMap<String, String>) {
        self.repo_root = configuration
            .get("repo_root")
            .cloned()
            .unwrap_or_default();

        subscribe(&[
            EventType::Timer,
            EventType::Key,
            EventType::Mouse,
        ]);

        set_timeout(0.5);
    }

    fn update(&mut self, event: Event) -> bool {
        match event {
            Event::Timer(_) => {
                poll::update_registry(&mut self.registry);
                set_timeout(0.5);
                true
            }
            Event::Key(key) => self.handle_key(key),
            Event::Mouse(mouse) => self.handle_mouse(mouse),
            _ => false,
        }
    }

    fn render(&mut self, rows: usize, cols: usize) {
        if let Some(modal) = &self.modal {
            ui::render_modal(modal, rows, cols);
        } else {
            ui::render_sidebar(&self.registry, rows, cols);
        }
    }
}

impl BobPlugin {
    fn handle_key(&mut self, key: KeyWithModifier) -> bool {
        if self.modal.is_some() {
            return self.handle_modal_key(key);
        }
        match key.bare_key {
            BareKey::Up | BareKey::Char('k') => {
                self.registry.select_prev();
                true
            }
            BareKey::Down | BareKey::Char('j') => {
                self.registry.select_next();
                true
            }
            BareKey::Char('n') => {
                self.modal = Some(ui::ModalState::default());
                true
            }
            BareKey::Char('x') => {
                spawn::kill_selected(&mut self.registry);
                true
            }
            BareKey::Char('r') => {
                poll::update_registry(&mut self.registry);
                true
            }
            _ => false,
        }
    }

    fn handle_modal_key(&mut self, key: KeyWithModifier) -> bool {
        let modal = match self.modal.as_mut() {
            Some(m) => m,
            None => return false,
        };
        match key.bare_key {
            BareKey::Esc => {
                self.modal = None;
                true
            }
            BareKey::Enter => {
                let prompt = modal.prompt.clone();
                let model = modal.model().to_string();
                self.modal = None;
                if !prompt.is_empty() {
                    spawn::spawn_agent(&mut self.registry, &self.repo_root, &prompt, &model);
                }
                true
            }
            BareKey::Tab => {
                modal.cycle_model();
                true
            }
            BareKey::Backspace => {
                modal.backspace();
                true
            }
            BareKey::Char(c) => {
                modal.push_char(c);
                true
            }
            _ => false,
        }
    }

    fn handle_mouse(&mut self, mouse: Mouse) -> bool {
        if let Mouse::LeftClick(row, _col) = mouse {
            let idx = self.registry.row_to_agent_index(row as usize);
            if let Some(i) = idx {
                self.registry.selected = i;
                return true;
            }
        }
        false
    }
}
