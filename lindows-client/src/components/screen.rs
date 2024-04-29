use leptos::ev::{MouseEvent, WheelEvent};
use leptos::*;
use wasm_bindgen::prelude::*;

use crate::electron;
use crate::message_generated::lindows_msg;
use crate::state::session::Session;
use crate::user_event::{mapping_key_event_to_code, Event};

#[allow(non_snake_case)]
#[component]
pub fn Screen() -> impl IntoView {
    let session = expect_context::<RwSignal<Session>>();
    let key_data_channel = session.get_untracked().key_data_channel.clone();
    let mouse_data_channel = session.get_untracked().mouse_data_channel.clone();
    let common_data_channel = session.get_untracked().common_data_channel.clone();

    // 鼠标移动事件处理
    let mouse_data_channel_cloned = mouse_data_channel.clone();
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

        // 用 i32 表示鼠标移动的比例，精度为 10000
        let x_ratio = (x / width * 10000.0) as i32;
        let y_ratio = (y / height * 10000.0) as i32;

        let mut builder = flatbuffers::FlatBufferBuilder::new();

        let payload = lindows_msg::Payload::create(
            &mut builder,
            &lindows_msg::PayloadArgs {
                p1: x_ratio,
                p2: y_ratio,
                p3: 0,
                p4: None,
            },
        );

        let msg = lindows_msg::Message::create(
            &mut builder,
            &lindows_msg::MessageArgs {
                event: Event::MOUSEEVENTF_MOVE as u8,
                payload: Some(payload),
            },
        );

        builder.finish(msg, None);
        let buf = builder.finished_data();
        mouse_data_channel_cloned
            .send_with_u8_array(buf)
            .expect("Send mouse data");
    };

    // 鼠标按下事件处理
    let mouse_data_channel_cloned = mouse_data_channel.clone();
    let handle_mousedown = move |event: MouseEvent| {
        event.prevent_default();

        let mut builder = flatbuffers::FlatBufferBuilder::new();

        match event.button() {
            0 => {
                let msg = lindows_msg::Message::create(
                    &mut builder,
                    &lindows_msg::MessageArgs {
                        event: Event::MOUSEEVENTF_LEFTDOWN as u8,
                        payload: None,
                    },
                );
                builder.finish(msg, None);
            }

            1 => {
                let msg = lindows_msg::Message::create(
                    &mut builder,
                    &lindows_msg::MessageArgs {
                        event: Event::MOUSEEVENTF_MIDDLEDOWN as u8,
                        payload: None,
                    },
                );
                builder.finish(msg, None);
            }

            2 => {
                let msg = lindows_msg::Message::create(
                    &mut builder,
                    &lindows_msg::MessageArgs {
                        event: Event::MOUSEEVENTF_RIGHTDOWN as u8,
                        payload: None,
                    },
                );
                builder.finish(msg, None);
            }

            3 => {
                let payload = lindows_msg::Payload::create(
                    &mut builder,
                    &lindows_msg::PayloadArgs {
                        p1: 1, // p1 == 1 表示 XBUTTON1
                        p2: 0,
                        p3: 0,
                        p4: None,
                    },
                );

                let msg = lindows_msg::Message::create(
                    &mut builder,
                    &lindows_msg::MessageArgs {
                        event: Event::MOUSEEVENTF_XDOWN as u8,
                        payload: Some(payload),
                    },
                );
                builder.finish(msg, None);
            }

            4 => {
                let payload = lindows_msg::Payload::create(
                    &mut builder,
                    &lindows_msg::PayloadArgs {
                        p1: 2, // p1 == 2 表示 XBUTTON2
                        p2: 0,
                        p3: 0,
                        p4: None,
                    },
                );

                let msg = lindows_msg::Message::create(
                    &mut builder,
                    &lindows_msg::MessageArgs {
                        event: Event::MOUSEEVENTF_XDOWN as u8,
                        payload: Some(payload),
                    },
                );
                builder.finish(msg, None);
            }

            _ => {}
        }

        let buf = builder.finished_data();
        mouse_data_channel_cloned
            .send_with_u8_array(buf)
            .expect("Send mouse data");
    };

    // 鼠标松开事件处理
    let mouse_data_channel_cloned = mouse_data_channel.clone();
    let handle_mouseup = move |event: MouseEvent| {
        event.prevent_default();

        let mut builder = flatbuffers::FlatBufferBuilder::new();

        match event.button() {
            0 => {
                let msg = lindows_msg::Message::create(
                    &mut builder,
                    &lindows_msg::MessageArgs {
                        event: Event::MOUSEEVENTF_LEFTUP as u8,
                        payload: None,
                    },
                );
                builder.finish(msg, None);
            }

            1 => {
                let msg = lindows_msg::Message::create(
                    &mut builder,
                    &lindows_msg::MessageArgs {
                        event: Event::MOUSEEVENTF_MIDDLEUP as u8,
                        payload: None,
                    },
                );
                builder.finish(msg, None);
            }

            2 => {
                let msg = lindows_msg::Message::create(
                    &mut builder,
                    &lindows_msg::MessageArgs {
                        event: Event::MOUSEEVENTF_RIGHTUP as u8,
                        payload: None,
                    },
                );
                builder.finish(msg, None);
            }

            3 => {
                let payload = lindows_msg::Payload::create(
                    &mut builder,
                    &lindows_msg::PayloadArgs {
                        p1: 1, // p1 == 1 表示 XBUTTON1
                        p2: 0,
                        p3: 0,
                        p4: None,
                    },
                );

                let msg = lindows_msg::Message::create(
                    &mut builder,
                    &lindows_msg::MessageArgs {
                        event: Event::MOUSEEVENTF_XUP as u8,
                        payload: Some(payload),
                    },
                );
                builder.finish(msg, None);
            }

            4 => {
                let payload = lindows_msg::Payload::create(
                    &mut builder,
                    &lindows_msg::PayloadArgs {
                        p1: 2, // p1 == 2 表示 XBUTTON2
                        p2: 0,
                        p3: 0,
                        p4: None,
                    },
                );

                let msg = lindows_msg::Message::create(
                    &mut builder,
                    &lindows_msg::MessageArgs {
                        event: Event::MOUSEEVENTF_XUP as u8,
                        payload: Some(payload),
                    },
                );
                builder.finish(msg, None);
            }

            _ => {}
        }

        let buf = builder.finished_data();
        mouse_data_channel_cloned
            .send_with_u8_array(buf)
            .expect("Send mouse data");
    };

    // 鼠标滚轮事件处理
    let mouse_data_channel_cloned = mouse_data_channel.clone();
    let handle_wheel = move |event: WheelEvent| {
        event.prevent_default();

        let delta_x = event.delta_x();
        let delta_y = event.delta_y();
        let delta_z = event.delta_z();

        let builder = &mut flatbuffers::FlatBufferBuilder::new();

        let payload = lindows_msg::Payload::create(
            builder,
            &lindows_msg::PayloadArgs {
                p1: delta_x as i32,
                p2: delta_y as i32,
                p3: delta_z as i32,
                p4: None,
            },
        );

        let msg = lindows_msg::Message::create(
            builder,
            &lindows_msg::MessageArgs {
                event: Event::MOUSEEVENTF_WHEEL as u8,
                payload: Some(payload),
            },
        );

        builder.finish(msg, None);
        let buf = builder.finished_data();
        mouse_data_channel_cloned
            .send_with_u8_array(buf)
            .expect("Send mouse data");

        // if delta_x > 0.0 {
        //     console_log("Wheel right");
        // } else if delta_x < 0.0 {
        //     console_log("Wheel left");
        // }

        // if delta_y > 0.0 {
        //     console_log("Wheel down");
        // } else if delta_y < 0.0 {
        //     console_log("Wheel up");
        // }

        // if delta_z > 0.0 {
        //     console_log("Wheel z down");
        // } else if delta_z < 0.0 {
        //     console_log("Wheel z up");
        // }
    };

    // Focus the video when clicked
    let handle_click = move |event: MouseEvent| {
        event.prevent_default();
        let target = event.target().unwrap();
        let element = target
            .dyn_into::<web_sys::HtmlVideoElement>()
            .expect("Video should cast to HtmlElement");
        element.focus().expect("Focus video");
    };

    // 键盘按下事件处理
    let key_data_channel_cloned = key_data_channel.clone();
    let handle_keydown = move |event: web_sys::KeyboardEvent| {
        event.prevent_default();

        let mut builder = flatbuffers::FlatBufferBuilder::new();

        let key = event.key();

        // TODO: 处理其它剪贴板操作
        if event.ctrl_key() || key == "v" {
            let common_data_channel_cloned = common_data_channel.clone();

            let text = electron::clipboard::read_text();
            if !text.is_empty() {
                let mut builder = flatbuffers::FlatBufferBuilder::new();

                let text = builder.create_string(&text);

                let payload = lindows_msg::Payload::create(
                    &mut builder,
                    &lindows_msg::PayloadArgs {
                        p1: 0,
                        p2: 0,
                        p3: 0,
                        p4: Some(text),
                    },
                );

                let msg = lindows_msg::Message::create(
                    &mut builder,
                    &lindows_msg::MessageArgs {
                        event: Event::VK_V as u8,
                        payload: Some(payload),
                    },
                );

                builder.finish(msg, None);
                let buf = builder.finished_data();
                common_data_channel_cloned
                    .send_with_u8_array(buf)
                    .expect("Send key data");
            }
        }

        let code = mapping_key_event_to_code(key);
        let payload = lindows_msg::Payload::create(
            &mut builder,
            &lindows_msg::PayloadArgs {
                p1: 0,
                p2: 0,
                p3: 0, // p3 == 0 表示按下
                p4: None,
            },
        );

        let msg = lindows_msg::Message::create(
            &mut builder,
            &lindows_msg::MessageArgs {
                event: code as u8,
                payload: Some(payload),
            },
        );

        builder.finish(msg, None);
        let buf = builder.finished_data();
        key_data_channel_cloned
            .send_with_u8_array(buf)
            .expect("Send key data");
    };

    // 键盘松开事件处理
    let key_data_channel_cloned = key_data_channel.clone();
    let handle_keyup = move |event: web_sys::KeyboardEvent| {
        event.prevent_default();

        let mut builder = flatbuffers::FlatBufferBuilder::new();

        let key = event.key();

        let code = mapping_key_event_to_code(key);
        let payload = lindows_msg::Payload::create(
            &mut builder,
            &lindows_msg::PayloadArgs {
                p1: 0,
                p2: 0,
                p3: 1, // p3 == 1 表示释放
                p4: None,
            },
        );

        let msg = lindows_msg::Message::create(
            &mut builder,
            &lindows_msg::MessageArgs {
                event: code as u8,
                payload: Some(payload),
            },
        );

        builder.finish(msg, None);
        let buf = builder.finished_data();
        key_data_channel_cloned
            .send_with_u8_array(buf)
            .expect("Send key data");
    };

    view! {
        <video
            id="screen"
            tabindex=0
            controls=true
            autoplay=true
            muted=false
            playsinline=true
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
