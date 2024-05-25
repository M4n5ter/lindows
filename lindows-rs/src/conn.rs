use std::sync::{Arc, Weak};

use futures_util::{SinkExt, StreamExt};
use lazy_static::lazy_static;
use serde::{Deserialize, Serialize};
use tokio::{
    net::TcpStream,
    sync::{
        mpsc::{self, UnboundedReceiver, UnboundedSender},
        Mutex,
    },
};
use tokio_tungstenite::{tungstenite, WebSocketStream};
use tracing::{debug, error, info, instrument, trace};
use webrtc::{
    api::media_engine::MIME_TYPE_VP8,
    data_channel::data_channel_message::DataChannelMessage,
    ice_transport::{
        ice_candidate::{RTCIceCandidate, RTCIceCandidateInit},
        ice_connection_state::RTCIceConnectionState,
    },
    peer_connection::{
        peer_connection_state::RTCPeerConnectionState,
        sdp::session_description::RTCSessionDescription, RTCPeerConnection,
    },
    rtp_transceiver::rtp_codec::{RTCRtpCodecCapability, RTPCodecType},
    track::track_local::track_local_static_sample::TrackLocalStaticSample,
};

use crate::rtc;

lazy_static! {
    pub static ref VIDEO_TRACK: Arc<TrackLocalStaticSample> =
        Arc::new(TrackLocalStaticSample::new(
            RTCRtpCodecCapability {
                mime_type: MIME_TYPE_VP8.to_owned(),
                ..Default::default()
            },
            "video".to_owned(),
            "stream".to_owned(),
        ));
    static ref PENDING_CANDIDATES: Arc<Mutex<Vec<RTCIceCandidate>>> = Arc::new(Mutex::new(vec![]));
}

#[derive(Debug)]
pub struct Conn {
    pub peer_connection: Arc<RTCPeerConnection>,
    pub ws_sender: UnboundedSender<tungstenite::Message>,
}

impl Conn {
    pub fn new(
        peer_connection: Arc<RTCPeerConnection>,
        ws_sender: UnboundedSender<tungstenite::Message>,
    ) -> Self {
        Self {
            peer_connection,
            ws_sender,
        }
    }

    pub async fn close(&self) -> anyhow::Result<()> {
        self.peer_connection.close().await?;
        Ok(())
    }

    pub async fn add_transceiver(&self) -> anyhow::Result<()> {
        self.peer_connection
            .add_transceiver_from_kind(RTPCodecType::Video, None)
            .await?;
        Ok(())
    }

    #[instrument]
    pub async fn set_on_data_channel(&self) {
        self.peer_connection.on_data_channel(Box::new(move |d| {
            let d_label = d.label().to_owned();

            Box::pin(async move {
                let d1 = Arc::clone(&d);
                let d_label1 = d_label.clone();
                let label = d_label1.to_owned();
                let label = label.as_str();

                d1.on_open(Box::new(move || {
                    debug!("Data channel opened: {d_label1}");

                    Box::pin(async move { if d_label1 == "common" {} })
                }));

                match label {
                    "key" => d1.on_message(Box::new(move |_msg: DataChannelMessage| {
                        debug!("Received key: {:?}", _msg.data);
                        Box::pin(async {})
                    })),
                    "mouse" => {
                        d1.on_message(Box::new(move |_msg: DataChannelMessage| Box::pin(async {})))
                    }
                    "common" => {
                        d1.on_message(Box::new(move |_msg: DataChannelMessage| Box::pin(async {})))
                    }
                    _ => d1.on_message(Box::new(move |msg: DataChannelMessage| {
                        debug!("Received unknown: {:?}", msg.data);
                        Box::pin(async {})
                    })),
                }
            })
        }));
    }

    #[instrument]
    pub async fn set_on_ice_candidate(&self) {
        let pc_weak = Arc::downgrade(&self.peer_connection);
        let ws_sender = self.ws_sender.clone();

        self.peer_connection
            .on_ice_candidate(Box::new(move |c: Option<RTCIceCandidate>| {
                info!("ICE candidate: {:?}", c);

                let pc_weak = Weak::clone(&pc_weak);
                let pending_candidates = Arc::clone(&PENDING_CANDIDATES);

                let ws_sender = ws_sender.clone();
                Box::pin(async move {
                    if let Some(candidate) = c {
                        if let Some(pc) = pc_weak.upgrade() {
                            let desc = pc.remote_description().await;
                            if desc.is_none() {
                                pending_candidates.lock().await.push(candidate);
                            } else if let Err(e) = signal_candidate(candidate, ws_sender) {
                                error!("Failed to signal candidate: {:?}", e);
                            }
                        } else {
                            error!("Peer connection dropped");
                        }
                    }
                })
            }))
    }

    pub async fn set_on_ice_connection_state_change(
        &mut self,
        record_signal_sender: mpsc::Sender<()>,
    ) {
        let pc_weak = Arc::downgrade(&self.peer_connection);
        let ws_sender = self.ws_sender.clone();

        self.peer_connection
            .on_ice_connection_state_change(Box::new(move |s| {
                info!("ICE connection state changed: {:?}", s);
                let pc_weak = Weak::clone(&pc_weak);
                let ws_sender = ws_sender.clone();
                let record_signal_sender = record_signal_sender.clone();

                Box::pin(async move {
                    if let Some(pc) = pc_weak.upgrade() {
                        if s == RTCIceConnectionState::Connected {
                            rtc::add_track_on_peer_connection(
                                Arc::clone(&pc),
                                Arc::clone(&VIDEO_TRACK),
                            )
                            .await
                            .unwrap();
                            rtc::send_offer(Arc::clone(&pc), ws_sender).await.unwrap();

                            record_signal_sender
                                .send(())
                                .await
                                .expect("Failed to send signal");
                        }
                    }
                })
            }));
    }

    pub async fn set_on_peer_connection_state_change(&mut self) {
        self.peer_connection
            .on_peer_connection_state_change(Box::new(move |s| {
                info!("Peer connection state changed: {:?}", s);

                if s == RTCPeerConnectionState::Connected {
                    Box::pin(async move {})
                } else {
                    Box::pin(async {})
                }
            }));
    }

    pub async fn set_on_ice_gathering_state_change(&mut self) {
        self.peer_connection
            .on_ice_gathering_state_change(Box::new(move |s| {
                info!("ICE gathering state changed: {:?}", s);

                Box::pin(async {})
            }));
    }

    pub async fn set_on_negotiation_needed(&mut self) {
        self.peer_connection
            .on_negotiation_needed(Box::new(move || {
                info!("Negotiation needed");

                Box::pin(async {})
            }));
    }

    pub async fn set_on_signaling_state_change(&mut self) {
        self.peer_connection
            .on_signaling_state_change(Box::new(move |s| {
                info!("Signaling state changed: {:?}", s);

                Box::pin(async {})
            }));
    }
}

#[instrument]
pub async fn handle_connection(
    mut ws_receiver: UnboundedReceiver<tungstenite::Message>,
    ws_sender: UnboundedSender<tungstenite::Message>,
    ws_stream: WebSocketStream<TcpStream>,
    peer_connection: Arc<RTCPeerConnection>,
) -> anyhow::Result<()> {
    let (mut ws_stream_sender, mut ws_stream_receiver) = ws_stream.split();
    tokio::spawn(async move {
        while let Some(message) = ws_receiver.recv().await {
            ws_stream_sender
                .send(message)
                .await
                .expect("Failed to send message");
        }
    });

    while let Some(msg) = ws_stream_receiver.next().await {
        let peer_connection = Arc::clone(&peer_connection);
        let ws_sender = ws_sender.clone();
        tokio::spawn(async move {
            let msg = msg.expect("Failed to receive message");
            match msg {
                tungstenite::Message::Text(text) => {
                    trace!("Received message: {:?}", text);
                    let ws_message: WSMessage =
                        serde_json::from_str(&text).expect("Failed to parse message");
                    match ws_message.event.as_str() {
                        "offer" => {
                            let sdp_str = ws_message.payload;
                            info!("Received offer: {:?}", sdp_str);
                            let sdp = match serde_json::from_str::<RTCSessionDescription>(&sdp_str)
                            {
                                Ok(sdp) => sdp,
                                Err(_) => RTCSessionDescription::offer(sdp_str)
                                    .expect("Failed to create offer"),
                            };

                            // Set remote description, create answer, set local description
                            {
                                peer_connection
                                    .set_remote_description(sdp)
                                    .await
                                    .expect("Failed to set remote description");

                                let answer = peer_connection
                                    .create_answer(None)
                                    .await
                                    .expect("Failed to create answer");
                                let ws_message = WSMessage {
                                    event: "answer".to_owned(),
                                    payload: answer
                                        .unmarshal()
                                        .expect("Unmarshal answer failed")
                                        .marshal(),
                                };
                                peer_connection
                                    .set_local_description(answer)
                                    .await
                                    .expect("Failed to set local description");

                                let ws_message = serde_json::to_string(&ws_message)
                                    .expect("Failed to serialize message");
                                ws_sender
                                    .clone()
                                    .send(tungstenite::Message::Text(ws_message))
                                    .expect("Failed to send message");
                            }

                            // Send pending candidates
                            {
                                let mut pending_candidates = PENDING_CANDIDATES.lock().await;
                                for candidate in pending_candidates.iter() {
                                    signal_candidate(candidate.clone(), ws_sender.clone())
                                        .expect("Failed to signal candidate");
                                }
                                pending_candidates.clear();
                            }

                            // tokio::time::sleep(tokio::time::Duration::from_secs(1)).await;
                            // rtc::send_offer(peer_connection.clone(), ws_stream.clone()).await?;
                        }
                        "answer" => {
                            let sdp_str = ws_message.payload;
                            let sdp = match serde_json::from_str::<RTCSessionDescription>(&sdp_str)
                            {
                                Ok(sdp) => sdp,
                                Err(_) => RTCSessionDescription::answer(sdp_str)
                                    .expect("Failed to create answer"),
                            };
                            peer_connection
                                .set_remote_description(sdp)
                                .await
                                .expect("Failed to set remote description");
                        }
                        "candidate" => {
                            let candidate_init = RTCIceCandidateInit {
                                candidate: ws_message.payload,
                                ..Default::default()
                            };
                            peer_connection
                                .add_ice_candidate(candidate_init)
                                .await
                                .expect("Failed to add ICE candidate");
                        }
                        "ping" => {
                            ws_sender
                                .clone()
                                .send(tungstenite::Message::Text(
                                    serde_json::to_string(&WSMessage {
                                        event: "pong".to_owned(),
                                        payload: "".to_owned(),
                                    })
                                    .expect("Failed to serialize message"),
                                ))
                                .expect("Failed to send message");
                        }
                        _ => {
                            error!("Unknown event: {:?}", ws_message.event);
                        }
                    }
                }
                tungstenite::Message::Close(_) => {
                    info!("Connection closed by client");
                }
                tungstenite::Message::Ping(_) => {
                    ws_sender
                        .clone()
                        .send(tungstenite::Message::Pong(vec![]))
                        .expect("Failed to send message");
                }
                tungstenite::Message::Pong(_) => {}
                _ => {
                    error!("Unsupported message: {:?}", msg);
                }
            }
        });
    }

    info!("Connection closed");
    Ok(())
}

#[inline]
fn signal_candidate(
    candidate: RTCIceCandidate,
    ws_sender: UnboundedSender<tungstenite::Message>,
) -> anyhow::Result<()> {
    let payload = candidate.to_json()?.candidate;
    let ws_message = WSMessage {
        event: "candidate".to_owned(),
        payload,
    };
    let ws_message = serde_json::to_string(&ws_message)?;
    ws_sender.send(tungstenite::Message::Text(ws_message))?;
    Ok(())
}

#[derive(Debug, Serialize, Deserialize)]
pub struct WSMessage {
    pub event: String,
    pub payload: String,
}
