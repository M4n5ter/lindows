use std::sync::Arc;

use futures_util::{SinkExt, StreamExt};
use lazy_static::lazy_static;
use serde::{Deserialize, Serialize};
use tokio::{net::TcpStream, sync::Mutex};
use tokio_tungstenite::{tungstenite, WebSocketStream};
use tracing::{debug, error, info, instrument, trace};
use webrtc::{
    api::media_engine::MIME_TYPE_VP8,
    data_channel::data_channel_message::DataChannelMessage,
    ice_transport::ice_candidate::{RTCIceCandidate, RTCIceCandidateInit},
    peer_connection::{
        peer_connection_state::RTCPeerConnectionState,
        sdp::session_description::RTCSessionDescription, RTCPeerConnection,
    },
    rtp_transceiver::{
        rtp_codec::{RTCRtpCodecCapability, RTPCodecType},
        RTCPFeedback,
    },
    track::track_local::track_local_static_sample::TrackLocalStaticSample,
};

use crate::rtc;

lazy_static! {
    pub static ref VIDEO_TRACK: Arc<Mutex<Arc<TrackLocalStaticSample>>> =
        Arc::new(Mutex::new(Arc::new(TrackLocalStaticSample::new(
            RTCRtpCodecCapability {
                mime_type: MIME_TYPE_VP8.to_owned(),
                clock_rate: 90000,
                channels: 0,
                sdp_fmtp_line: "".to_owned(),
                rtcp_feedback: vec![
                    RTCPFeedback {
                        typ: "goog-remb".to_owned(),
                        parameter: "".to_owned(),
                    },
                    RTCPFeedback {
                        typ: "ccm".to_owned(),
                        parameter: "fir".to_owned(),
                    },
                    RTCPFeedback {
                        typ: "nack".to_owned(),
                        parameter: "".to_owned(),
                    },
                    RTCPFeedback {
                        typ: "nack".to_owned(),
                        parameter: "pli".to_owned(),
                    },
                ],
            },
            "video".to_owned(),
            "stream".to_owned(),
        ))));
    static ref PENDING_CANDIDATES: Arc<Mutex<Vec<RTCIceCandidate>>> = Arc::new(Mutex::new(vec![]));
}

#[derive(Debug)]
pub struct Conn {
    pub peer_connection: Arc<RTCPeerConnection>,
    ws_stream: Arc<Mutex<WebSocketStream<TcpStream>>>,
}

impl Conn {
    pub fn new(
        peer_connection: Arc<RTCPeerConnection>,
        ws_stream: Arc<Mutex<WebSocketStream<TcpStream>>>,
    ) -> Self {
        Self {
            peer_connection,
            ws_stream,
        }
    }

    pub async fn close(&self) -> anyhow::Result<()> {
        self.peer_connection.close().await?;
        Ok(())
    }

    pub async fn add_video_track(&self) -> anyhow::Result<()> {
        rtc::add_track_on_peer_connection(
            self.peer_connection.clone(),
            VIDEO_TRACK.clone().lock().await.clone(),
        )
        .await?;
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
                let d1 = d.clone();
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
        let pc = Arc::downgrade(&self.peer_connection);
        let ws_stream = Arc::downgrade(&self.ws_stream);

        self.peer_connection
            .on_ice_candidate(Box::new(move |c: Option<RTCIceCandidate>| {
                info!("ICE candidate: {:?}", c);

                let pc1 = pc.clone();
                let pending_candidates = PENDING_CANDIDATES.clone();
                let ws_stream1 = ws_stream.clone();

                Box::pin(async move {
                    if let Some(candidate) = c {
                        if let Some(pc) = pc1.upgrade() {
                            let desc = pc.remote_description().await;
                            if desc.is_none() {
                                pending_candidates.lock().await.push(candidate);
                            } else if let Some(ws_stream) = ws_stream1.upgrade() {
                                if let Err(e) = signal_candidate(candidate, ws_stream).await {
                                    error!("Failed to signal candidate: {:?}", e);
                                }
                            }
                        }
                    }
                })
            }))
    }

    pub async fn set_on_ice_connection_state_change(&mut self) {
        self.peer_connection
            .on_ice_connection_state_change(Box::new(move |s| {
                info!("ICE connection state changed: {:?}", s);

                Box::pin(async {})
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
    ws_stream: Arc<Mutex<WebSocketStream<TcpStream>>>,
    peer_connection: Arc<RTCPeerConnection>,
) -> anyhow::Result<()> {
    let mut ws_stream1 = ws_stream.lock().await;
    while let Some(msg) = ws_stream1.next().await {
        let msg = msg?;

        match msg {
            tungstenite::Message::Text(text) => {
                trace!("Received message: {:?}", text);
                let ws_message: WSMessage = serde_json::from_str(&text)?;
                match ws_message.event.as_str() {
                    "offer" => {
                        let sdp_str = ws_message.payload;
                        info!("Received offer: {:?}", sdp_str);
                        let sdp = match serde_json::from_str::<RTCSessionDescription>(&sdp_str) {
                            Ok(sdp) => sdp,
                            Err(_) => RTCSessionDescription::offer(sdp_str)?,
                        };

                        // Set remote description, create answer, set local description
                        {
                            peer_connection.set_remote_description(sdp).await?;

                            rtc::add_track_on_peer_connection(
                                peer_connection.clone(),
                                VIDEO_TRACK.clone().lock().await.clone(),
                            )
                            .await?;

                            let answer = peer_connection.create_answer(None).await?;

                            let ws_message = WSMessage {
                                event: "answer".to_owned(),
                                payload: answer.unmarshal()?.marshal(),
                            };
                            let ws_message = serde_json::to_string(&ws_message)?;
                            ws_stream1
                                .send(tungstenite::Message::Text(ws_message))
                                .await?;

                            peer_connection.set_local_description(answer).await?;
                        }

                        // Send pending candidates
                        {
                            let mut pending_candidates = PENDING_CANDIDATES.lock().await;
                            for candidate in pending_candidates.iter() {
                                let ws_stream2 = ws_stream.clone();
                                signal_candidate(candidate.clone(), ws_stream2).await?;
                            }
                            pending_candidates.clear();
                        }

                        // tokio::time::sleep(tokio::time::Duration::from_secs(1)).await;
                        // rtc::send_offer(peer_connection.clone(), ws_stream.clone()).await?;
                    }
                    "answer" => {
                        let sdp_str = ws_message.payload;
                        let sdp = match serde_json::from_str::<RTCSessionDescription>(&sdp_str) {
                            Ok(sdp) => sdp,
                            Err(_) => RTCSessionDescription::answer(sdp_str)?,
                        };
                        peer_connection.set_remote_description(sdp).await?;
                    }
                    "candidate" => {
                        let candidate_init = RTCIceCandidateInit {
                            candidate: ws_message.payload,
                            ..Default::default()
                        };
                        peer_connection.add_ice_candidate(candidate_init).await?;
                    }
                    "ping" => {
                        ws_stream1
                            .send(tungstenite::Message::Text(serde_json::to_string(
                                &WSMessage {
                                    event: "pong".to_owned(),
                                    payload: "".to_owned(),
                                },
                            )?))
                            .await?;
                    }
                    _ => {
                        error!("Unknown event: {:?}", ws_message.event);
                    }
                }
            }
            tungstenite::Message::Close(_) => {
                info!("Connection closed by client");
                break;
            }
            tungstenite::Message::Ping(_) => {
                ws_stream1.send(tungstenite::Message::Pong(vec![])).await?;
            }
            tungstenite::Message::Pong(_) => {}
            _ => {
                error!("Unsupported message: {:?}", msg);
            }
        }
    }

    info!("Connection closed");
    Ok(())
}

async fn signal_candidate(
    candidate: RTCIceCandidate,
    ws_stream: Arc<Mutex<WebSocketStream<TcpStream>>>,
) -> anyhow::Result<()> {
    let payload = candidate.to_json()?.candidate;
    let ws_message = WSMessage {
        event: "candidate".to_owned(),
        payload,
    };
    let ws_message = serde_json::to_string(&ws_message)?;
    ws_stream
        .lock()
        .await
        .send(tungstenite::Message::Text(ws_message))
        .await?;
    Ok(())
}

#[derive(Debug, Serialize, Deserialize)]
pub struct WSMessage {
    pub event: String,
    pub payload: String,
}
