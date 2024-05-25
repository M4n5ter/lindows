use hbb_common::message_proto::EncodedVideoFrame;
use std::{sync::Arc, thread, time::Duration};
use tokio::{net::TcpListener, sync::mpsc};
use tokio_tungstenite::accept_async;
use tracing::{error, info, instrument};
use webrtc::media::Sample;

use crate::{
    conn::{self, handle_connection, Conn},
    record, rtc,
};

#[derive(Debug)]
pub struct Server<'a> {
    addr: &'a str,
    connections: Vec<Conn>,
}

impl<'a> Server<'a> {
    pub fn new(addr: &'a str) -> Self {
        Self {
            addr,
            connections: vec![],
        }
    }

    pub async fn serve(&mut self) -> anyhow::Result<()> {
        let (sender, mut receiver) = mpsc::unbounded_channel::<(EncodedVideoFrame, u64)>();
        let (record_signal_sender, record_signal_receiver) = mpsc::channel::<()>(1);
        tokio::spawn(async move {
            while let Some((frame, duration)) = receiver.recv().await {
                let sample = Sample {
                    data: frame.data,
                    duration: Duration::from_millis(duration),
                    // duration: Duration::from_secs(1),
                    ..Default::default()
                };

                Arc::clone(&conn::VIDEO_TRACK)
                    .write_sample(&sample)
                    .await
                    .expect("Failed to write sample");
            }
        });
        thread::spawn(move || record::record(sender, record_signal_receiver));

        let tcp_listener = TcpListener::bind(&self.addr)
            .await
            .expect("Failed to bind to port 11111");
        self.listen_and_serve(tcp_listener, record_signal_sender)
            .await;
        Ok(())
    }

    #[instrument]
    async fn listen_and_serve(
        &mut self,
        listener: TcpListener,
        record_signal_sender: mpsc::Sender<()>,
    ) {
        while let Ok((stream, _)) = listener.accept().await {
            info!("Accepted new connection");

            if let Ok(ws_stream) = accept_async(stream).await {
                let pc = rtc::new_peer_connection().await;
                let (ws_sender, ws_receiver) = tokio::sync::mpsc::unbounded_channel();
                let mut conn = Conn::new(pc.clone(), ws_sender.clone());
                conn.add_transceiver()
                    .await
                    .expect("Failed to add transceiver");
                // conn.set_on_data_channel().await;
                conn.set_on_ice_candidate().await;
                conn.set_on_peer_connection_state_change().await;
                conn.set_on_ice_connection_state_change(record_signal_sender.clone())
                    .await;
                conn.set_on_ice_gathering_state_change().await;
                conn.set_on_negotiation_needed().await;
                conn.set_on_signaling_state_change().await;

                self.connections.push(conn);

                tokio::spawn(async move {
                    if let Err(err) = handle_connection(ws_receiver, ws_sender, ws_stream, pc).await
                    {
                        error!("Failed to handle connection: {:?}", err);
                    }
                });
            };
        }

        error!("TCP listener closed");
    }
}
