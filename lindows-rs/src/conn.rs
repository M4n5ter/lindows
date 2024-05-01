use std::{collections::HashMap, sync::Arc};

use futures_util::{SinkExt, StreamExt};
use serde::{Deserialize, Serialize};
use tokio::{
    net::{TcpListener, TcpStream},
    sync::Mutex,
};
use tokio_tungstenite::{accept_async, tungstenite, WebSocketStream};
use tracing::{error, info, instrument, trace};
use webrtc::{
    api::{
        interceptor_registry::register_default_interceptors, media_engine::MediaEngine, APIBuilder,
    },
    ice_transport::{
        ice_candidate::{RTCIceCandidate, RTCIceCandidateInit},
        ice_server::RTCIceServer,
    },
    interceptor::registry::Registry,
    peer_connection::{
        configuration::RTCConfiguration, sdp::session_description::RTCSessionDescription,
        RTCPeerConnection,
    },
};

#[derive(Debug)]
pub struct Conn {
    pub peer_connection: Arc<Mutex<RTCPeerConnection>>,
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

        let peer_connection = Arc::new(Mutex::new(
            api.new_peer_connection(config)
                .await
                .expect("Failed to create peer connection"),
        ));

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

    pub async fn close(&self) -> anyhow::Result<()> {
        self.peer_connection.lock().await.close().await?;
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
    pub async fn set_on_ice_candidate(&self) {
        let pc = Arc::downgrade(&self.peer_connection);
        let pending_candidates = self.pending_candidates.clone();
        let ws_streams = self.ws_streams.clone();

        self.peer_connection.lock().await.on_ice_candidate(Box::new(
            move |c: Option<RTCIceCandidate>| {
                let pc1 = pc.clone();
                let pending_candidates1 = pending_candidates.clone();
                let ws_streams1 = ws_streams.clone();

                Box::pin(async move {
                    if let Some(candidate) = c {
                        if let Some(pc) = pc1.upgrade() {
                            let desc = pc.lock().await.remote_description().await;
                            if desc.is_none() {
                                pending_candidates1.lock().await.push(candidate);
                            } else {
                                for ws_stream in ws_streams1.lock().await.values() {
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
            },
        ))
    }
}

#[instrument]
async fn handle_connection(
    ws_stream: Arc<Mutex<WebSocketStream<TcpStream>>>,
    peer_connection: Arc<Mutex<RTCPeerConnection>>,
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
                        info!("Received offer event");

                        let sdp_str = ws_message.payload;
                        let sdp = match serde_json::from_str::<RTCSessionDescription>(&sdp_str) {
                            Ok(sdp) => sdp,
                            Err(_) => RTCSessionDescription::offer(sdp_str)?,
                        };
                        info!("Parsed offer successfully");

                        // Set remote description, create answer, set local description
                        {
                            let peer_connection = peer_connection.lock().await;
                            peer_connection.set_remote_description(sdp).await?;
                            info!("Set remote description successfully");

                            let answer = peer_connection.create_answer(None).await?;
                            let ws_message = WSMessage {
                                event: "answer".to_owned(),
                                payload: answer.unmarshal()?.marshal(),
                            };
                            let ws_message = serde_json::to_string(&ws_message)?;
                            ws_stream1
                                .send(tungstenite::Message::Text(ws_message))
                                .await?;
                            info!("Sent answer successfully");

                            peer_connection.set_local_description(answer).await?;
                            info!("Set local description successfully");
                        }

                        // Send pending candidates
                        {
                            let mut pending_candidates = pending_candidates.lock().await;
                            for candidate in pending_candidates.iter() {
                                let ws_stream2 = ws_stream.clone();
                                signal_candidate(candidate.clone(), ws_stream2).await?;
                            }
                            pending_candidates.clear();

                            info!("Signaled pending candidates successfully");
                        }
                    }
                    "candidate" => {
                        info!("Received candidate event");

                        let candidate_init = RTCIceCandidateInit {
                            candidate: ws_message.payload,
                            ..Default::default()
                        };
                        let peer_connection = peer_connection.lock().await;
                        peer_connection.add_ice_candidate(candidate_init).await?;
                        info!("Added ICE candidate successfully");
                    }
                    "ping" => {
                        info!("Received ping event");

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
