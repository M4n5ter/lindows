use leptos::*;

use crate::components::screen::Screen;
use crate::components::sidebar::Sidebar;

#[component]
pub fn App() -> impl IntoView {
    view! {
        <main class="flex fullwindow">
            <div class="w-24">
                <Sidebar/>
            </div>

            <div class="w-1 divider divider-horizontal"></div>

            <div class="flex flex-grow">
                <div class="flex skeleton rounded-box w-full items-center justify-center">
                    <Screen src="https://www.w3schools.com/html/mov_bbb.mp4"/>
                </div>
            </div>
        </main>
    }
}
