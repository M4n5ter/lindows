use leptos::leptos_dom::logging::console_log;
use leptos::{*};
use leptos::ev::MouseEvent;
use serde::{Deserialize, Serialize};
use wasm_bindgen::prelude::*;

use crate::components::sidebar::Sidebar;

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
    let handle_mousemove = move |event: MouseEvent| {
        let target = event.target().unwrap();
        let element = target.dyn_into::<web_sys::Element>().unwrap();
        let rect = element.get_bounding_client_rect();
        let x = event.client_x() as f64 - rect.left();
        let y = event.client_y() as f64 - rect.top();

        let width = rect.width();
        let height = rect.height();

        let x_ratio = x / width;
        let y_ratio = y / height;

        console_log(&format!("Mouse moved to ({}, {})", x_ratio, y_ratio));
    };

    view! {
        <main class="flex fullwindow">
            <div class="w-24">
                <Sidebar/>
            </div>

            <div class="w-1 divider divider-horizontal"></div>

            <div class="flex flex-grow">
                <div class="flex skeleton rounded-box w-full items-center justify-center">
                    <video
                        controls=false
                        autoplay=true
                        playsinline=true
                        src="https://www.w3schools.com/html/mov_bbb.mp4"
                        on:mousemove=handle_mousemove
                    >
                        // <source src="https://www.w3schools.com/html/mov_bbb.mp4" type="video/mp4"/>
                        "Your browser does not support the video tag."
                    </video>
                </div>
            </div>
        </main>
    }
}
