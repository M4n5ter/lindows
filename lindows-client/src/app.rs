use leptos::{*};
use serde::{Deserialize, Serialize};
use wasm_bindgen::prelude::*;

use crate::components::sidebar::Sidebar;
use crate::components::screen::Screen;

#[wasm_bindgen]
extern "C" {
    #[wasm_bindgen(js_namespace = ["window", "__TAURI__", "core"])]
    async fn invoke(cmd: &str, args: JsValue) -> JsValue;
}

#[derive(Serialize, Deserialize)]
struct GreetArgs<'a> {
    name: &'a str,
}

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
