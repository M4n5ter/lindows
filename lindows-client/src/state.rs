pub mod config;

pub fn provide_state() {
    config::provide_config();
}
