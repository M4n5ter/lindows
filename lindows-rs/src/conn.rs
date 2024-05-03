use std::{collections::HashMap, sync::Arc};

use futures_util::{SinkExt, StreamExt};
use lazy_static::lazy_static;
use serde::{Deserialize, Serialize};
use tokio::{
    net::{TcpListener, TcpStream},
    sync::Mutex,
};
use tokio_tungstenite::{accept_async, tungstenite, WebSocketStream};
use tracing::{debug, error, info, instrument, trace};
use webrtc::{
    api::{
        interceptor_registry::register_default_interceptors,
        media_engine::{MediaEngine, MIME_TYPE_VP8},
        APIBuilder,
    },
    data_channel::data_channel_message::DataChannelMessage,
    ice_transport::{
        ice_candidate::{RTCIceCandidate, RTCIceCandidateInit},
        ice_server::RTCIceServer,
    },
    interceptor::registry::Registry,
    peer_connection::{
        configuration::RTCConfiguration, peer_connection_state::RTCPeerConnectionState,
        sdp::session_description::RTCSessionDescription, RTCPeerConnection,
    },
    rtp_transceiver::{rtp_codec::RTCRtpCodecCapability, RTCPFeedback},
    track::track_local::{track_local_static_sample::TrackLocalStaticSample, TrackLocal},
};

lazy_static! {
    pub static ref VIDEO_TRACK: Arc<Mutex<Option<Arc<TrackLocalStaticSample>>>> =
        Arc::new(Mutex::new(None));
}

#[derive(Debug)]
pub struct Conn {
    pub peer_connection: Arc<RTCPeerConnection>,
    pub pending_candidates: Arc<Mutex<Vec<RTCIceCandidate>>>,
    pub tcp_listener: TcpListener,
    ws_streams: Arc<Mutex<StreamsMap>>,
}

type StreamsMap = HashMap<i32, Arc<Mutex<WebSocketStream<TcpStream>>>>;

impl Conn {
    pub async fn new(tcp_listener: TcpListener) -> Self {
        let config = RTCConfiguration {
            ice_servers: vec![RTCIceServer {
                urls: vec!["stun:stun.syncthing.net:3478".to_owned()],
                ..Default::default()
            }],
            ..Default::default()
        };

        let mut m = MediaEngine::default();
        m.register_default_codecs()
            .expect("Failed to register default codecs");

        let mut registry = Registry::new();

        registry = register_default_interceptors(registry, &mut m)
            .expect("Failed to register default interceptors");

        let api = APIBuilder::new()
            .with_media_engine(m)
            .with_interceptor_registry(registry)
            .build();

        let peer_connection = Arc::new(
            api.new_peer_connection(config)
                .await
                .expect("Failed to create peer connection"),
        );

        let video_rtcp_feedback = vec![
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
        ];
        let video_track = Arc::new(TrackLocalStaticSample::new(
            RTCRtpCodecCapability {
                mime_type: MIME_TYPE_VP8.to_owned(),
                clock_rate: 90000,
                channels: 0,
                sdp_fmtp_line: "".to_owned(),
                rtcp_feedback: video_rtcp_feedback,
            },
            "video".to_owned(),
            "stream".to_owned(),
        ));

        let rtp_sender = peer_connection
            .add_track(Arc::clone(&video_track) as Arc<dyn TrackLocal + Send + Sync>)
            .await
            .expect("Failed to add track");
        tokio::spawn(async move {
            let mut rtcp_buf = vec![0u8; 1500];
            while let Ok((_, _)) = rtp_sender.read(&mut rtcp_buf).await {}
            anyhow::Result::<()>::Ok(())
        });
        *VIDEO_TRACK.lock().await = Some(video_track);

        let pending_candidates = Arc::new(Mutex::new(vec![]));

        let ws_streams = Arc::new(Mutex::new(HashMap::<
            i32,
            Arc<Mutex<WebSocketStream<TcpStream>>>,
        >::new()));

        Self {
            peer_connection,
            pending_candidates,
            tcp_listener,
            ws_streams,
        }
    }

    pub async fn send_offer(
        pc: Arc<RTCPeerConnection>,
        ws_streams: Arc<Mutex<StreamsMap>>,
    ) -> anyhow::Result<()> {
        let offer = pc.create_offer(None).await?;
        pc.set_local_description(offer.clone()).await?;

        let ws_streams = ws_streams.lock().await;
        for ws_stream in ws_streams.values() {
            let ws_stream = ws_stream.clone();
            let ws_message = WSMessage {
                event: "offer".to_owned(),
                payload: offer.unmarshal()?.marshal(),
            };
            let ws_message = serde_json::to_string(&ws_message)?;
            ws_stream
                .lock()
                .await
                .send(tungstenite::Message::Text(ws_message))
                .await?;
        }

        Ok(())
    }

    pub async fn close(&self) -> anyhow::Result<()> {
        self.peer_connection.close().await?;
        Ok(())
    }

    #[instrument]
    pub async fn serve(&self) {
        let mut stream_id = 0;

        while let Ok((stream, _)) = self.tcp_listener.accept().await {
            info!("Accepted new connection");

            stream_id += 1;
            let peer_connection = self.peer_connection.clone();
            let pending_candidates = self.pending_candidates.clone();
            if let Ok(ws_stream) = accept_async(stream).await {
                let ws_stream = Arc::new(Mutex::new(ws_stream));
                let ws_stream_cloned = ws_stream.clone();
                let streams = self.ws_streams.clone();
                tokio::spawn(async move {
                    if let Err(err) = handle_connection(
                        ws_stream_cloned,
                        peer_connection,
                        pending_candidates,
                        stream_id,
                        streams,
                    )
                    .await
                    {
                        error!("Failed to handle connection: {:?}", err);
                    }
                });
                self.ws_streams.lock().await.insert(stream_id, ws_stream);
            };
        }

        error!("TCP listener closed");
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
        let pending_candidates = Arc::downgrade(&self.pending_candidates);
        let ws_streams = Arc::downgrade(&self.ws_streams);

        self.peer_connection
            .on_ice_candidate(Box::new(move |c: Option<RTCIceCandidate>| {
                info!("ICE candidate: {:?}", c);

                let pc1 = pc.clone();
                let pending_candidates1 = pending_candidates.clone();
                let ws_streams1 = ws_streams.clone();

                Box::pin(async move {
                    if let Some(candidate) = c {
                        if let Some(pc) = pc1.upgrade() {
                            let desc = pc.remote_description().await;
                            if desc.is_none() {
                                if let Some(pending_candidates) = pending_candidates1.upgrade() {
                                    pending_candidates.lock().await.push(candidate);
                                }
                            } else if let Some(ws_streams) = ws_streams1.upgrade() {
                                for ws_stream in ws_streams.lock().await.values() {
                                    if let Err(e) =
                                        signal_candidate(candidate.clone(), ws_stream.clone()).await
                                    {
                                        error!("Failed to signal candidate: {:?}", e);
                                    }
                                }
                            }
                        }
                    }
                })
            }))
    }

    pub async fn set_on_peer_connection_state_change(&mut self) {
        let peer_connection = Arc::downgrade(&self.peer_connection);
        let ws_streams = Arc::downgrade(&self.ws_streams);
        self.peer_connection
            .on_peer_connection_state_change(Box::new(move |s| {
                info!("Peer connection state changed: {:?}", s);

                let peer_connection = peer_connection.clone();
                let ws_streams = ws_streams.clone();

                if s == RTCPeerConnectionState::Connected {
                    Box::pin(async move {
                        let video_rtcp_feedback = vec![
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
                        ];
                        let video_track = Some(Arc::new(TrackLocalStaticSample::new(
                            RTCRtpCodecCapability {
                                mime_type: MIME_TYPE_VP8.to_owned(),
                                clock_rate: 90000,
                                channels: 0,
                                sdp_fmtp_line: "".to_owned(),
                                rtcp_feedback: video_rtcp_feedback,
                            },
                            "video".to_owned(),
                            "stream".to_owned(),
                        )));

                        if let Some(peer_connection) = peer_connection.upgrade() {
                            peer_connection
                                .add_track(video_track.as_ref().unwrap().clone())
                                .await
                                .expect("Failed to add track");

                            *VIDEO_TRACK.lock().await = video_track;

                            if let Some(ws_streams) = ws_streams.upgrade() {
                                Self::send_offer(peer_connection, ws_streams)
                                    .await
                                    .expect("Failed to send offer")
                            };
                        };
                    })
                } else {
                    Box::pin(async {})
                }
            }));
    }
}

#[instrument]
async fn handle_connection(
    ws_stream: Arc<Mutex<WebSocketStream<TcpStream>>>,
    peer_connection: Arc<RTCPeerConnection>,
    pending_candidates: Arc<Mutex<Vec<RTCIceCandidate>>>,
    stream_id: i32,
    streams: Arc<Mutex<StreamsMap>>,
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
                        let sdp = match serde_json::from_str::<RTCSessionDescription>(&sdp_str) {
                            Ok(sdp) => sdp,
                            Err(_) => RTCSessionDescription::offer(sdp_str)?,
                        };

                        // Set remote description, create answer, set local description
                        {
                            peer_connection.set_remote_description(sdp).await?;

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
                            let mut pending_candidates = pending_candidates.lock().await;
                            for candidate in pending_candidates.iter() {
                                let ws_stream2 = ws_stream.clone();
                                signal_candidate(candidate.clone(), ws_stream2).await?;
                            }
                            pending_candidates.clear();
                        }
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
                ws_stream1.close(None).await?;
                streams.lock().await.remove(&stream_id);
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
struct WSMessage {
    pub event: String,
    pub payload: String,
}
