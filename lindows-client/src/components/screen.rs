use std::rc::Rc;
use std::sync::Mutex;

use leptos::ev::{MouseEvent, WheelEvent};
use leptos::leptos_dom::logging::console_log;
use leptos::*;
use serde::{Deserialize, Serialize};
use wasm_bindgen::prelude::*;
use wasm_bindgen_futures::JsFuture;
use web_sys::{
    MediaStream, MessageEvent, RtcIceCandidate, RtcIceCandidateInit, RtcPeerConnection,
    RtcPeerConnectionIceEvent, RtcSdpType, RtcSessionDescriptionInit, RtcTrackEvent, WebSocket,
};

use crate::state::config::LindowsConfig;
use crate::tauri::{self};

#[component]
pub fn Screen() -> impl IntoView {
    let pending_candidates = Rc::new(Mutex::new(Vec::<RtcIceCandidate>::new()));

    // let mut config = RtcConfiguration::new();
    // let mut syncthing_ice_server = RtcIceServer::new();
    // syncthing_ice_server.url("stun:stun.syncthing.net:3478");
    // config.ice_servers(&syncthing_ice_server);

    // peer connection
    // let peer = Rc::new(
    //     RtcPeerConnection::new_with_configuration(&config).expect("Create peer connection"),
    // );
    let peer = Rc::new(RtcPeerConnection::new().expect("Create peer connection"));

    // websocket
    let url = expect_context::<LindowsConfig>()
        .answer_addr
        .get_untracked();
    let ws = Rc::new(WebSocket::new(&url).expect("Create WebSocket"));

    {
        let pending_candidates_cloned = pending_candidates.clone();
        let ws_cloned = ws.clone();
        let peer_cloned = peer.clone();
        let onmessage_callback = Closure::wrap(Box::new(move |e: MessageEvent| {
            // 检查消息类型是否为文本
            if let Ok(text) = e.data().dyn_into::<js_sys::JsString>() {
                // 将JS字符串转换为Rust字符串
                if let Some(text_str) = text.as_string() {
                    // 尝试反序列化文本到WSMessage结构体
                    match serde_json::from_str::<WSMessage>(&text_str) {
                        Ok(message) => {
                            // 如果反序列化成功，则打印消息
                            console_log(&format!("Received message: {:?}", message));
                            if message.event == "answer" {
                                let sdp = message.payload;
                                let peer_cloned = peer_cloned.clone();
                                spawn_local(async move {
                                    let mut session_description_init =
                                        RtcSessionDescriptionInit::new(RtcSdpType::Answer);
                                    session_description_init.sdp(&sdp);
                                    JsFuture::from(
                                        peer_cloned
                                            .set_remote_description(&session_description_init),
                                    )
                                    .await
                                    .expect("Set remote description");
                                });

                                if let Ok(mut pending_candidates) = pending_candidates_cloned.lock()
                                {
                                    for candidate in pending_candidates.iter() {
                                        let candidate_msg = WSMessage {
                                            event: "candidate".to_string(),
                                            payload: candidate.candidate(),
                                        };
                                        let candidate_msg_str =
                                            serde_json::to_string(&candidate_msg)
                                                .expect("Serialize candidate message");
                                        ws_cloned
                                            .send_with_str(&candidate_msg_str)
                                            .expect("Send candidate");
                                    }
                                    pending_candidates.clear();
                                }
                            } else if message.event == "candidate" {
                                let candidate = message.payload;
                                let mut candidate_init = RtcIceCandidateInit::new(&candidate);
                                candidate_init.sdp_mid(Some(""));
                                candidate_init.sdp_m_line_index(Some(0));
                                let add_ice_candidate_promise = peer_cloned
                                    .add_ice_candidate_with_opt_rtc_ice_candidate_init(Some(
                                        &candidate_init,
                                    ));
                                spawn_local(async move {
                                    let result = JsFuture::from(add_ice_candidate_promise).await;
                                    if let Err(error) = result {
                                        console_log(&format!(
                                            "Failed to add ICE candidate: {:?}",
                                            error
                                        ));
                                    } else {
                                        console_log("ICE candidate added");
                                    }
                                });
                            }
                        }
                        Err(e) => {
                            // 如果反序列化失败，则打印错误
                            console_log(&format!("Error deserializing message: {:?}", e));
                        }
                    }
                }
            }
        }) as Box<dyn FnMut(MessageEvent)>);

        ws.set_onmessage(Some(onmessage_callback.as_ref().unchecked_ref()));
        onmessage_callback.forget();
    }

    // on track 事件回调
    {
        let ontrack_callback = Closure::wrap(Box::new(move |event: RtcTrackEvent| {
            // 获取 window 对象
            let window = window();
            // 获取 document 对象
            let document = window.document().expect("should have a Document");
            let video_element = document
                .get_element_by_id("screen")
                .expect("Get video element");
            let video_html_element = video_element
                .dyn_into::<web_sys::HtmlVideoElement>()
                .expect("Video element");

            let track = event.track();
            if track.kind() == "video" {
                let video_stream =
                    MediaStream::new_with_tracks(&track).expect("Create video stream");
                video_html_element.set_src_object(Some(&video_stream));
            } else if event.track().kind() == "audio" {
            }
        }) as Box<dyn FnMut(_)>);

        peer.set_ontrack(Some(ontrack_callback.as_ref().unchecked_ref()));
        ontrack_callback.forget();
    }

    // on icecandidate 事件回调
    {
        let pending_candidates_clone = pending_candidates.clone();
        let ws_cloned = ws.clone();
        let peer_cloned = peer.clone();
        let onicecandidate_callback =
            Closure::wrap(Box::new(move |event: RtcPeerConnectionIceEvent| {
                let candidate = event.candidate();
                if let Some(candidate) = candidate {
                    let desc = peer_cloned.remote_description();
                    if desc.is_none() {
                        if let Ok(mut pending_candidates) = pending_candidates_clone.lock() {
                            pending_candidates.push(candidate);
                        }
                    } else if let Some(candidate_string) = candidate.to_json().as_string() {
                        let candidate_msg = WSMessage {
                            event: "candidate".to_string(),
                            payload: candidate_string,
                        };
                        let candidate_msg_str = serde_json::to_string(&candidate_msg)
                            .expect("Serialize candidate message");

                        ws_cloned
                            .send_with_str(&candidate_msg_str)
                            .expect("Send candidate");
                    };
                }
            }) as Box<dyn FnMut(_)>);

        let peer_cloned = peer.clone();
        peer_cloned.set_onicecandidate(Some(onicecandidate_callback.as_ref().unchecked_ref()));
        onicecandidate_callback.forget();
    }

    // on connectionstatechange 事件回调
    {
        let peer_cloned = peer.clone();
        let onconnectionstatechange_callback = Closure::wrap(Box::new(move || {
            let state = peer_cloned.connection_state();
            console_log(&format!("Connection state: {:?}", state));
        }) as Box<dyn FnMut()>);

        peer.set_onconnectionstatechange(Some(
            onconnectionstatechange_callback.as_ref().unchecked_ref(),
        ));
        onconnectionstatechange_callback.forget();
    }

    // on iceconnectionstatechange 事件回调
    {
        let peer_cloned = peer.clone();
        let oniceconnectionstatechange_callback = Closure::wrap(Box::new(move || {
            let state = peer_cloned.ice_connection_state();
            console_log(&format!("ICE connection state: {:?}", state));
        }) as Box<dyn FnMut()>);

        peer.set_oniceconnectionstatechange(Some(
            oniceconnectionstatechange_callback.as_ref().unchecked_ref(),
        ));
        oniceconnectionstatechange_callback.forget();
    }

    // 创建 DataChannel
    let peer_cloned = peer.clone();
    let data_channel = Rc::new(peer_cloned.create_data_channel("hello"));
    {
        let data_channel_cloned = data_channel.clone();
        let onopen_callback = Closure::wrap(Box::new(move || {
            console_log("Data channel opened");
            data_channel_cloned
                .send_with_str("Hello from lindows-client!")
                .expect("Send data");
        }) as Box<dyn FnMut()>);

        data_channel.set_onopen(Some(onopen_callback.as_ref().unchecked_ref()));
        onopen_callback.forget();

        let onmessage_callback = Closure::wrap(Box::new(move |event: MessageEvent| {
            if let Ok(text) = event.data().dyn_into::<js_sys::JsString>() {
                if let Some(text_str) = text.as_string() {
                    console_log(&format!("Data channel message: {}", text_str));
                }
            }
        }) as Box<dyn FnMut(MessageEvent)>);

        data_channel.set_onmessage(Some(onmessage_callback.as_ref().unchecked_ref()));
        onmessage_callback.forget();
    }

    // offer
    let ws_cloned = ws.clone();
    spawn_local(async move {
        let offer = JsFuture::from(peer.create_offer())
            .await
            .expect("Create offer");
        let offer = offer
            .dyn_into::<web_sys::RtcSessionDescription>()
            .expect("Offer");
        let sdp_type = offer.type_();
        let sdp = offer.sdp();
        let mut session_description_init = RtcSessionDescriptionInit::new(sdp_type);
        session_description_init.sdp(&sdp);
        JsFuture::from(peer.set_local_description(&session_description_init))
            .await
            .expect("Set local description");

        let offer_msg = WSMessage {
            event: "offer".to_string(),
            payload: sdp,
        };
        let offer_msg_str = serde_json::to_string(&offer_msg).expect("Serialize offer message");

        ws_cloned.send_with_str(&offer_msg_str).expect("Send offer");
    });

    // 键盘鼠标事件处理
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
            id="screen"
            tabindex=0
            controls=false
            autoplay=true
            playsinline=true
            // src=src
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

#[derive(Debug, Serialize, Deserialize)]
struct WSMessage {
    pub event: String,
    pub payload: String,
}
