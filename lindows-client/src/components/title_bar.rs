use crate::{components::svgs, electron, state::window::MainWindow};
use leptos::*;

#[allow(non_snake_case)]
#[component]
pub fn TitleBar() -> impl IntoView {
    let mut window_state = expect_context::<MainWindow>();

    let close_on_click = move |_| {
        electron::window::close(1);
    };
    let minimize_on_click = move |_| {
        electron::window::minimize(1);
    };
    let maximize_on_click = move |_| {
        if window_state.is_maximized {
            electron::window::unmaximize(1);
            window_state.is_maximized = false;
        } else {
            electron::window::maximize(1);
            window_state.is_maximized = true;
        }
    };

    view! {
        <div
            data-tauri-drag-region
            class="titlebar w-full flex flex-row-reverse flex-none justify-center bg-base-100 top-0 pt-2 pb-2 pr-2"
        >
            <div class="flex flex-1 flex-row-reverse">

                // close button
                <button
                    class="btn btn-ghost btn-sm flex items-center justify-center btn-square hover:bg-red-500 "
                    on:click=close_on_click
                >
                    <svgs::Close></svgs::Close>
                </button>

                // maximize button
                <button
                    class=" btn btn-ghost btn-sm flex items-center justify-center btn-square"
                    on:click=maximize_on_click
                >
                    <svgs::Maximize></svgs::Maximize>
                </button>

                // minimize button
                <button
                    class="btn btn-ghost btn-sm flex items-center justify-center btn-square"
                    on:click=minimize_on_click
                >
                    <svgs::Minimize></svgs::Minimize>
                </button>

            </div>
        </div>
    }
}
