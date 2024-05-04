use std::sync::Arc;

use futures_util::SinkExt;
use tokio::{net::TcpStream, sync::Mutex};
use tokio_tungstenite::{tungstenite, WebSocketStream};
use webrtc::{
    api::{
        interceptor_registry::register_default_interceptors,
        media_engine::{MediaEngine, MIME_TYPE_VP8},
        APIBuilder,
    },
    ice_transport::ice_server::RTCIceServer,
    interceptor::registry::Registry,
    peer_connection::{configuration::RTCConfiguration, RTCPeerConnection},
    rtp_transceiver::{rtp_codec::RTCRtpCodecCapability, RTCPFeedback},
    track::track_local::track_local_static_sample::TrackLocalStaticSample,
};

use crate::conn::WSMessage;

#[inline]
pub async fn add_track_on_peer_connection(
    pc: Arc<RTCPeerConnection>,
    track: Arc<TrackLocalStaticSample>,
) -> anyhow::Result<()> {
    let rtp_sender = pc.add_track(track).await?;
    tokio::spawn(async move {
        let mut rtcp_buf = vec![0u8; 1500];
        while let Ok((_, _)) = rtp_sender.read(&mut rtcp_buf).await {}
        anyhow::Result::<()>::Ok(())
    });
    Ok(())
}

#[inline]
pub async fn new_peer_connection() -> Arc<RTCPeerConnection> {
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

    Arc::new(
        api.new_peer_connection(config)
            .await
            .expect("Failed to create peer connection"),
    )
}

#[inline]
pub async fn new_video_track() -> Arc<TrackLocalStaticSample> {
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
    Arc::new(TrackLocalStaticSample::new(
        RTCRtpCodecCapability {
            mime_type: MIME_TYPE_VP8.to_owned(),
            clock_rate: 90000,
            channels: 0,
            sdp_fmtp_line: "".to_owned(),
            rtcp_feedback: video_rtcp_feedback,
        },
        "video".to_owned(),
        "stream".to_owned(),
    ))
}

#[inline]
pub async fn send_offer(
    pc: Arc<RTCPeerConnection>,
    ws_stream: Arc<Mutex<WebSocketStream<TcpStream>>>,
) -> anyhow::Result<()> {
    let offer = pc.create_offer(None).await?;
    let offer_sdp = offer.sdp.clone();
    let ws_message = WSMessage {
        event: "offer".to_owned(),
        payload: offer_sdp,
    };
    let ws_message = serde_json::to_string(&ws_message)?;
    pc.set_local_description(offer).await?;

    ws_stream
        .clone()
        .lock()
        .await
        .send(tungstenite::Message::Text(ws_message))
        .await?;
    Ok(())
}
