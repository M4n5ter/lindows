pub mod config;
pub mod session;
pub mod window;

pub fn provide_state() {
    config::provide_config();
    session::provide_session();
    window::provide_window();
}
