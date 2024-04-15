use leptos::ev::{MouseEvent, WheelEvent};
use leptos::leptos_dom::logging::console_log;
use leptos::*;
use wasm_bindgen::prelude::*;

use crate::tauri::{self};

#[component]
pub fn Screen(#[prop(into)] src: String) -> impl IntoView {
    let handle_mousemove = move |event: MouseEvent| {
        let target = event.target().unwrap();
        let element = target
            .dyn_into::<web_sys::Element>()
            .expect("Video element");
        let rect = element.get_bounding_client_rect();
        let x = event.client_x() as f64 - rect.left();
        let y = event.client_y() as f64 - rect.top();

        let width = rect.width();
        let height = rect.height();

        let x_ratio = x / width;
        let y_ratio = y / height;

        console_log(&format!("Mouse moved to ({}, {})", x_ratio, y_ratio));
    };

    let handle_mousedown = move |event: MouseEvent| {
        event.prevent_default();
        let button = event.button();
        if button == 0 {
            console_log("Left button pressed");
        } else if button == 1 {
            console_log("Middle button pressed");
        } else if button == 2 {
            console_log("Right button pressed");
        }
    };

    let handle_mouseup = move |event: MouseEvent| {
        event.prevent_default();
        let button = event.button();
        if button == 0 {
            console_log("Left button released");
        } else if button == 1 {
            console_log("Middle button released");
        } else if button == 2 {
            console_log("Right button released");
        }
    };

    let handle_wheel = move |event: WheelEvent| {
        let delta_x = event.delta_x();
        let delta_y = event.delta_y();
        let delta_z = event.delta_z();

        if delta_x > 0.0 {
            console_log("Wheel right");
        } else if delta_x < 0.0 {
            console_log("Wheel left");
        }

        if delta_y > 0.0 {
            console_log("Wheel down");
        } else if delta_y < 0.0 {
            console_log("Wheel up");
        }

        if delta_z > 0.0 {
            console_log("Wheel z down");
        } else if delta_z < 0.0 {
            console_log("Wheel z up");
        }
    };

    // Focus the video when clicked
    let handle_click = move |event: MouseEvent| {
        event.prevent_default();
        let target = event.target().unwrap();
        let element = target
            .dyn_into::<web_sys::HtmlVideoElement>()
            .expect("Video should cast to HtmlElement");
        element.focus().expect("Focus video");
        console_log("Video focused");
    };

    let handle_keydown = move |event: web_sys::KeyboardEvent| {
        let key = event.key();
        console_log(&format!("Key down: {}", key));

        spawn_local(async move {
            if key == "c" {
                let text = tauri::clipboard::read_text().await;
                console_log(&format!("Clipboard text: {}", text));
            }

            if key == "v" {
                let text = "Hello from Leptos!";
                tauri::clipboard::write_text(text).await;
                console_log("Text copied to clipboard");
            }
        });
    };

    let handle_keyup = move |event: web_sys::KeyboardEvent| {
        let key = event.key();
        console_log(&format!("Key up: {}", key));
    };

    view! {
        <video
            tabindex=0
            controls=false
            autoplay=true
            playsinline=true
            src=src
            on:mousemove=handle_mousemove
            on:mousedown=handle_mousedown
            on:mouseup=handle_mouseup
            on:wheel=handle_wheel
            on:click=handle_click
            on:keydown=handle_keydown
            on:keyup=handle_keyup
        >

            "Your browser does not support the video tag."
        </video>
    }
}
