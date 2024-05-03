use std::{rc::Rc, sync::Mutex};

use js_sys::Reflect;
use leptos::{
    create_rw_signal, expect_context, leptos_dom::logging::console_log, provide_context,
    spawn_local, window, RwSignal, SignalGetUntracked as _,
};
use serde::{Deserialize, Serialize};
use wasm_bindgen::{closure::Closure, JsCast as _};
use wasm_bindgen_futures::JsFuture;
use web_sys::{
    MediaStream, MessageEvent, RtcConfiguration, RtcDataChannel, RtcIceCandidate,
    RtcIceCandidateInit, RtcIceServer, RtcPeerConnection, RtcPeerConnectionIceEvent, RtcSdpType,
    RtcSessionDescription, RtcSessionDescriptionInit, RtcTrackEvent, WebSocket,
};

use crate::state::config::LindowsConfig;

pub fn provide_session() {
    let session = Session::new();
    session.set_peer_callback();
    session.set_data_channel();
    session.set_ws_callback();
    provide_context(create_rw_signal(session));
}

#[derive(Clone)]
pub struct Session {
    pub pending_candidates: Rc<Mutex<Vec<RtcIceCandidate>>>,
    pub ws: Rc<WebSocket>,
    pub peer: Rc<RtcPeerConnection>,
    pub key_data_channel: Rc<RtcDataChannel>,
    pub mouse_data_channel: Rc<RtcDataChannel>,
    pub common_data_channel: Rc<RtcDataChannel>,
    pub address: RwSignal<String>,
}

impl Session {
    pub fn new() -> Self {
        let pending_candidates = Rc::new(Mutex::new(Vec::<RtcIceCandidate>::new()));

        let mut config = RtcConfiguration::new();
        let mut syncthing_ice_server = RtcIceServer::new();
        syncthing_ice_server.url("stun:stun.syncthing.net:3478");

        let ice_servers = js_sys::Array::new();
        ice_servers.push(&syncthing_ice_server);

        config.ice_servers(&ice_servers);

        // peer connection
        let peer = Rc::new(
            RtcPeerConnection::new_with_configuration(&config).expect("Create peer connection"),
        );

        // websocket
        let url = expect_context::<LindowsConfig>()
            .answer_addr
            .get_untracked();
        let ws = Rc::new(WebSocket::new(&url).expect("Create WebSocket"));

        let peer_cloned = peer.clone();
        Self {
            pending_candidates,
            ws,
            peer,
            key_data_channel: Rc::new(peer_cloned.create_data_channel("key")),
            mouse_data_channel: Rc::new(peer_cloned.create_data_channel("mouse")),
            common_data_channel: Rc::new(peer_cloned.create_data_channel("common")),
            address: RwSignal::new(url),
        }
    }

    pub fn connect(&mut self) {
        console_log(&format!("Connecting to {}", self.address.get_untracked()));
        let _ = self.ws.close();
        self.ws = Rc::new(WebSocket::new(&self.address.get_untracked()).expect("Create WebSocket"));
        self.set_ws_callback();
        console_log("WebSocket connected")
    }

    pub async fn send_offer(&self) {
        let offer = JsFuture::from(self.peer.create_offer())
            .await
            .expect("Create offer");
        let sdp = Reflect::get(&offer, &"sdp".into())
            .expect("Get sdp")
            .as_string()
            .expect("Sdp as string");

        let mut session_description_init = RtcSessionDescriptionInit::new(RtcSdpType::Offer);
        session_description_init.sdp(&sdp);
        JsFuture::from(self.peer.set_local_description(&session_description_init))
            .await
            .expect("Set local description");

        let sdp = RtcSessionDescription::new_with_description_init_dict(&session_description_init).expect("Offer");
        let offer_msg = WSMessage {
            event: "offer".to_string(),
            payload: sdp.sdp(),
        };
        let offer_msg_str = serde_json::to_string(&offer_msg).expect("Serialize offer message");

        self.ws.send_with_str(&offer_msg_str).expect("Send offer");
    }

    pub fn set_peer_callback(&self) {
        // on track 事件回调
        set_peer_ontrack(self.peer.clone());

        // on icecandidate 事件回调
        set_peer_onicecandidate(
            self.peer.clone(),
            self.ws.clone(),
            self.pending_candidates.clone(),
        );

        // on connectionstatechange 事件回调
        set_peer_onconnectionstatechange(self.peer.clone());

        // on iceconnectionstatechange 事件回调
        set_peer_oniceconnectionstatechange(self.peer.clone());
    }

    pub fn set_data_channel(&self) {
        let common_data_channel = self.common_data_channel.clone();
        {
            let data_channel_cloned = common_data_channel.clone();
            let onopen_callback = Closure::wrap(Box::new(move || {
                console_log("Data channel opened");
                data_channel_cloned
                    .send_with_str("Hello from lindows-client!")
                    .expect("Send data");
            }) as Box<dyn FnMut()>);

            common_data_channel.set_onopen(Some(onopen_callback.as_ref().unchecked_ref()));
            onopen_callback.forget();

            let onmessage_callback = Closure::wrap(Box::new(move |event: MessageEvent| {
                if let Ok(text) = event.data().dyn_into::<js_sys::JsString>() {
                    if let Some(text_str) = text.as_string() {
                        console_log(&format!("Data channel message: {}", text_str));
                    }
                }
            }) as Box<dyn FnMut(MessageEvent)>);

            common_data_channel.set_onmessage(Some(onmessage_callback.as_ref().unchecked_ref()));
            onmessage_callback.forget();
        }
    }

    pub fn set_ws_callback(&self) {
        let ws_cloned = self.ws.clone();
        let peer = self.peer.clone();
        let pending_candidates = self.pending_candidates.clone();
        let onmessage_callback = Closure::wrap(Box::new(move |e: MessageEvent| {
            // 检查消息类型是否为文本
            if let Ok(text) = e.data().dyn_into::<js_sys::JsString>() {
                // 将JS字符串转换为Rust字符串
                if let Some(text_str) = text.as_string() {
                    // 尝试反序列化文本到WSMessage结构体
                    match serde_json::from_str::<WSMessage>(&text_str) {
                        Ok(message) => {
                            // 如果反序列化成功，则处理消息
                            handle_ws_message(
                                message,
                                peer.clone(),
                                ws_cloned.clone(),
                                pending_candidates.clone(),
                            );
                        }
                        Err(e) => {
                            // 如果反序列化失败，则处理错误
                            handle_serde_json_error(e);
                        }
                    }
                }
            }
        }) as Box<dyn FnMut(MessageEvent)>);

        self.ws
            .set_onmessage(Some(onmessage_callback.as_ref().unchecked_ref()));
        onmessage_callback.forget();

        // on close 事件回调
        let onclose_callback = Closure::wrap(Box::new(move || {
            console_log("WebSocket closed");
        }) as Box<dyn FnMut()>);

        self.ws
            .set_onclose(Some(onclose_callback.as_ref().unchecked_ref()));
        onclose_callback.forget();

        // on error 事件回调
        let onerror_callback = Closure::wrap(Box::new(move || {
            console_log("WebSocket error");
        }) as Box<dyn FnMut()>);

        self.ws
            .set_onerror(Some(onerror_callback.as_ref().unchecked_ref()));
        onerror_callback.forget();

        // on open 事件回调
        let onopen_callback = Closure::wrap(Box::new(move || {
            console_log("WebSocket opened");
        }) as Box<dyn FnMut()>);

        self.ws
            .set_onopen(Some(onopen_callback.as_ref().unchecked_ref()));
        onopen_callback.forget();

        console_log("WebSocket callbacks set");

        // keep alive
        // let ws = self.ws.clone();
        // spawn_local(async move {
        //     loop {
        //         ws.send_with_str(
        //             &serde_json::to_string(&WSMessage {
        //                 event: "ping".to_string(),
        //                 payload: "".to_string(),
        //             })
        //             .expect("Serialize ping message"),
        //         )
        //         .expect("Send ping");
                
        //         wasm_timer::Delay::new(std::time::Duration::from_secs(5))
        //         .await
        //         .expect("Delay");
        //     }
        // });

    }
}

impl Default for Session {
    fn default() -> Self {
        Self::new()
    }
}

#[derive(Debug, Serialize, Deserialize)]
struct WSMessage {
    pub event: String,
    pub payload: String,
}

fn set_peer_ontrack(peer: Rc<RtcPeerConnection>) {
    let ontrack_callback = Closure::wrap(Box::new(move |event: RtcTrackEvent| {
        console_log("Track event received");

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
            console_log("Video track received");

            let tracks = js_sys::Array::new();
            tracks.push(&track);

            let video_stream = MediaStream::new_with_tracks(&tracks).expect("Create video stream");
            video_html_element.set_src_object(Some(&video_stream));

            console_log("Video stream set");
        } else if event.track().kind() == "audio" {
            console_log("Audio track received");
        }
    }) as Box<dyn FnMut(_)>);

    peer.set_ontrack(Some(ontrack_callback.as_ref().unchecked_ref()));
    ontrack_callback.forget();
}

fn set_peer_onicecandidate(
    peer: Rc<RtcPeerConnection>,
    ws: Rc<WebSocket>,
    pending_candidates: Rc<Mutex<Vec<RtcIceCandidate>>>,
) {
    let ws = ws.clone();
    let peer_cloned = peer.clone();
    let onicecandidate_callback = Closure::wrap(Box::new(move |event: RtcPeerConnectionIceEvent| {
        let candidate = event.candidate();
        if let Some(candidate) = candidate {
            let desc = peer_cloned.remote_description();
            if desc.is_none() {
                if let Ok(mut pending_candidates) = pending_candidates.lock() {
                    pending_candidates.push(candidate);
                }
            } else if let Some(candidate_string) = candidate.to_json().as_string() {
                let candidate_msg = WSMessage {
                    event: "candidate".to_string(),
                    payload: candidate_string,
                };
                let candidate_msg_str =
                    serde_json::to_string(&candidate_msg).expect("Serialize candidate message");

                ws.send_with_str(&candidate_msg_str)
                    .expect("Send candidate");
            };
        }
    }) as Box<dyn FnMut(_)>);

    let peer = peer.clone();
    peer.set_onicecandidate(Some(onicecandidate_callback.as_ref().unchecked_ref()));
    onicecandidate_callback.forget();
}

fn set_peer_onconnectionstatechange(peer: Rc<RtcPeerConnection>) {
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

fn set_peer_oniceconnectionstatechange(peer: Rc<RtcPeerConnection>) {
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

fn handle_ws_message(
    message: WSMessage,
    peer: Rc<RtcPeerConnection>,
    ws: Rc<WebSocket>,
    pending_candidates: Rc<Mutex<Vec<RtcIceCandidate>>>,
) {
    // 根据消息类型执行相应的操作
    match message.event.as_str() {
        "answer" => {
            let sdp = message.payload;
            console_log(&format!("Received answer: {}", sdp));
            let peer = peer.clone();
            spawn_local(async move {
                let mut session_description_init =
                    RtcSessionDescriptionInit::new(RtcSdpType::Answer);
                session_description_init.sdp(&sdp);
                JsFuture::from(peer.set_remote_description(&session_description_init))
                    .await
                    .expect("Set remote description");
            });

            // 发送所有待处理的ICE候选
            send_pending_candidates(pending_candidates.clone(), ws.clone());
        }
        "offer" => {
            let sdp = message.payload;
            console_log(&format!("Received offer: {}", sdp));
            let peer = peer.clone();
            spawn_local(async move {
                let mut session_description_init =
                    RtcSessionDescriptionInit::new(RtcSdpType::Offer);
                session_description_init.sdp(&sdp);
                JsFuture::from(peer.set_remote_description(&session_description_init))
                    .await
                    .expect("Set remote description");

                let answer = JsFuture::from(peer.create_answer())
                    .await
                    .expect("Create answer");
                let answer = answer
                    .dyn_into::<web_sys::RtcSessionDescription>()
                    .expect("Answer");
                let sdp_type = answer.type_();
                let sdp = answer.sdp();
                let mut session_description_init = RtcSessionDescriptionInit::new(sdp_type);
                session_description_init.sdp(&sdp);
                JsFuture::from(peer.set_local_description(&session_description_init))
                    .await
                    .expect("Set local description");

                let answer_msg = WSMessage {
                    event: "answer".to_string(),
                    payload: sdp,
                };
                let answer_msg_str =
                    serde_json::to_string(&answer_msg).expect("Serialize answer message");
                ws.send_with_str(&answer_msg_str).expect("Send answer");

                // 发送所有待处理的ICE候选
                send_pending_candidates(pending_candidates.clone(), ws.clone());
            });
        }
        "candidate" => {
            let candidate = message.payload;
            let mut candidate_init = RtcIceCandidateInit::new(&candidate);
            candidate_init.sdp_mid(Some(""));
            candidate_init.sdp_m_line_index(Some(0));
            let add_ice_candidate_promise =
                peer.add_ice_candidate_with_opt_rtc_ice_candidate_init(Some(&candidate_init));
            spawn_local(async move {
                let result = JsFuture::from(add_ice_candidate_promise).await;
                if let Err(error) = result {
                    console_log(&format!("Failed to add ICE candidate: {:?}", error));
                } else {
                    console_log("ICE candidate added");
                }
            });
        }
        "ping" => {
            ws.send_with_str(
                &serde_json::to_string(&WSMessage {
                    event: "pong".to_string(),
                    payload: "".to_string(),
                })
                .expect("Serialize pong message"),
            )
            .expect("Send pong");
        }
        _ => {
            console_log("Unknown message event");
        }
    }
}

fn handle_serde_json_error(e: serde_json::Error) {
    // 打印错误信息
    console_log(&format!("Error deserializing message: {:?}", e));
}

fn send_pending_candidates(pending_candidates: Rc<Mutex<Vec<RtcIceCandidate>>>, ws: Rc<WebSocket>) {
    // 发送所有待处理的ICE候选
    if let Ok(mut pending_candidates) = pending_candidates.lock() {
        for candidate in pending_candidates.iter() {
            let candidate_msg = WSMessage {
                event: "candidate".to_string(),
                payload: candidate.candidate(),
            };
            let candidate_msg_str =
                serde_json::to_string(&candidate_msg).expect("Serialize candidate message");
            ws.send_with_str(&candidate_msg_str)
                .expect("Send candidate");
        }
        pending_candidates.clear();
    }
}
