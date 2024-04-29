mod app;
pub mod components;
pub mod electron;
pub mod message_generated;
pub mod state;
pub mod tauri;
pub mod user_event;

use app::*;
use leptos::*;

fn main() {
    console_error_panic_hook::set_once();
    mount_to_body(|| {
        view! { <App/> }
    })
}
