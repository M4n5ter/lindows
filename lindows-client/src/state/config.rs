use leptos::{provide_context, RwSignal};

pub fn provide_config() {
    provide_context(LindowsConfig::new());
}

#[derive(Clone, Copy, Debug)]
pub struct LindowsConfig {
    pub answer_addr: RwSignal<String>,
}

impl LindowsConfig {
    pub fn new() -> Self {
        Self {
            answer_addr: RwSignal::new("ws://192.168.1.109:8080/ws".to_string()),
        }
    }
}

impl Default for LindowsConfig {
    fn default() -> Self {
        Self::new()
    }
}
