use leptos::*;

use crate::components::screen::Screen;
use crate::components::sidebar::Sidebar;
use crate::state::provide_state;

#[component]
pub fn App() -> impl IntoView {
    provide_state();

    view! {
        <main class="flex fullwindow">
            <div class="w-24">
                <Sidebar/>
            </div>

            <div class="w-1 divider divider-horizontal"></div>

            <div class="flex flex-grow">
                <div class="flex skeleton rounded-box w-full items-center justify-center">
                    <Screen/>
                </div>
            </div>
        </main>
    }
}
