use leptos::*;

use crate::components::screen::Screen;
use crate::components::sidebar::Sidebar;
use crate::components::title_bar::TitleBar;
use crate::state::provide_state;

#[allow(non_snake_case)]
#[component]
pub fn App() -> impl IntoView {
    provide_state();
    view! {
        <main class="flex flex-col fullwindow">
            <TitleBar/>

            <div class="flex flex-1 h-full w-full">
                <div class="w-24">
                    <Sidebar/>
                </div>

                <div class="divider divider-horizontal w-auto py-4"></div>

                <div class="flex flex-grow">
                    <div class="flex skeleton rounded-box w-full items-center justify-center">
                        <Screen/>
                    </div>
                </div>
            </div>
        </main>
    }
}
