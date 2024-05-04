use anyhow::Context;
use bytes::Bytes;
use std::{
    sync::{mpsc, Arc},
    time::{self, Duration, SystemTime},
};
use tokio::{net::TcpListener, sync::Mutex};
use tokio_tungstenite::accept_async;
use tracing::{error, info, instrument};
use webrtc::media::Sample;

use crate::{
    conn::{self, handle_connection, Conn},
    rtc,
};
use crabgrab::{
    capturable_content::{CapturableContent, CapturableContentFilter, CapturableWindowFilter},
    capture_stream::{CaptureConfig, CaptureStream, StreamEvent},
    feature::bitmap::{FrameBitmap, VideoFrameBitmap},
};
use lindows_utils::common::{
    self,
    codec::{self, EncoderApi},
    vpxcodec, PixelBuffer,
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

    #[inline]
    pub async fn serve(&mut self) -> anyhow::Result<()> {
        let start = time::Instant::now();

        let (sender, receiver) = mpsc::channel::<Sample>();
        tokio::spawn(async move {
            while let Ok(sample) = receiver.recv() {
                conn::VIDEO_TRACK
                    .clone()
                    .lock()
                    .await
                    .write_sample(&sample)
                    .await
                    .expect("Failed to write sample");
            }
        });

        tokio::spawn(async move {
            // let runtime = tokio::runtime::Builder::new_multi_thread().build().unwrap();

            let token = match CaptureStream::test_access(false) {
                Some(token) => token,
                None => CaptureStream::request_access(false)
                    .await
                    .expect("Expected capture access"),
            };
            let window_filter = CapturableWindowFilter {
                desktop_windows: false,
                onscreen_only: true,
            };
            let filter = CapturableContentFilter {
                windows: Some(window_filter),
                displays: false,
            };
            let content = CapturableContent::new(filter).await.unwrap();
            let window = content.windows().find(|window| {
                let app_identifier = window.application().identifier();
                !window.title().is_empty()
                    && (app_identifier.to_lowercase().contains("finder")
                        || app_identifier.to_lowercase().contains("qq"))
            });

            match window {
                Some(window) => {
                    println!("capturing window: {}", window.title());
                    let (width, height) = (window.rect().size.width, window.rect().size.height);

                    let config = CaptureConfig::with_window(
                        window,
                        CaptureStream::supported_pixel_formats()[0],
                    )
                    .unwrap();

                    let mut vpx = vpxcodec::VpxEncoder::new(
                        codec::EncoderCfg::VPX(vpxcodec::VpxEncoderConfig {
                            width: width as u32,
                            height: height as u32,
                            quality: codec::Quality::Balanced,
                            codec: vpxcodec::VpxVideoCodecId::VP8,
                            keyframe_interval: None,
                        }),
                        false,
                    )
                    .expect("Failed to create vpx encoder");

                    let mut last_capture_time = time::Instant::now();
                    let mut stream =
                        CaptureStream::new(token, config, move |stream_event| match stream_event {
                            Ok(event) => {
                                if let StreamEvent::Video(frame) = event {
                                    let mut yuv = Vec::<u8>::new();
                                    let mut mid_data = Vec::<u8>::new();

                                    let now = time::Instant::now();
                                    let duration = now.duration_since(last_capture_time);
                                    last_capture_time = now;
                                    let time = now - start;
                                    let ms = (time.as_secs() * 1000 + time.subsec_millis() as u64)
                                        as i64;

                                    if let FrameBitmap::BgraUnorm8x4(bgra8888) =
                                        frame.get_bitmap().expect("Failed to get bitmap")
                                    {
                                        let data: Vec<u8> = bgra8888
                                            .data
                                            .iter()
                                            .flat_map(|pixel| pixel.to_vec())
                                            .collect();

                                        let frame = common::Frame::PixelBuffer(PixelBuffer::new(
                                            &data,
                                            bgra8888.width,
                                            bgra8888.height,
                                        ));

                                        let frame = frame
                                            .to(vpx.yuvfmt(), &mut yuv, &mut mid_data)
                                            .expect("Failed to convert frame");

                                        let mut frames = vec![];
                                        for ref frame in vpx
                                            .encode(
                                                ms,
                                                frame.yuv().expect("Shoud have yuv data"),
                                                common::STRIDE_ALIGN,
                                            )
                                            .expect("Failed to encode frame")
                                        {
                                            let data = Bytes::copy_from_slice(frame.data);
                                            let sample = Sample {
                                                data,
                                                timestamp: SystemTime::now(),
                                                duration,
                                                ..Default::default()
                                            };
                                            frames.push(sample);
                                        }
                                        for ref frame in vpx
                                            .flush()
                                            .with_context(|| "Failed to flush")
                                            .expect("Failed to flush")
                                        {
                                            let data = Bytes::copy_from_slice(frame.data);
                                            let sample = Sample {
                                                data,
                                                timestamp: SystemTime::now(),
                                                duration,
                                                ..Default::default()
                                            };
                                            frames.push(sample);
                                        }
                                        for frame in frames {
                                            sender.send(frame).unwrap();
                                        }
                                    };
                                }
                            }
                            Err(error) => {
                                println!("Stream error: {:?}", error);
                            }
                        })
                        .unwrap();
                    println!("stream created!");
                    tokio::task::block_in_place(|| {
                        std::thread::sleep(Duration::from_millis(600000))
                    });
                    stream.stop().unwrap();
                }
                None => {
                    println!("Failed to find window");
                }
            }
        });

        let tcp_listener = TcpListener::bind(&self.addr)
            .await
            .expect("Failed to bind to port 11111");
        self.listen_and_serve(tcp_listener).await;
        Ok(())
    }

    #[instrument]
    async fn listen_and_serve(&mut self, listener: TcpListener) {
        while let Ok((stream, _)) = listener.accept().await {
            info!("Accepted new connection");

            if let Ok(ws_stream) = accept_async(stream).await {
                let pc = rtc::new_peer_connection().await;
                let ws_stream = Arc::new(Mutex::new(ws_stream));
                let mut conn = Conn::new(pc.clone(), ws_stream.clone());
                conn.add_transceiver()
                    .await
                    .expect("Failed to add transceiver");
                // conn.set_on_data_channel().await;
                conn.set_on_ice_candidate().await;
                conn.set_on_peer_connection_state_change().await;
                conn.set_on_ice_connection_state_change().await;
                conn.set_on_ice_gathering_state_change().await;
                conn.set_on_negotiation_needed().await;
                conn.set_on_signaling_state_change().await;

                self.connections.push(conn);

                tokio::spawn(async move {
                    if let Err(err) = handle_connection(ws_stream, pc).await {
                        error!("Failed to handle connection: {:?}", err);
                    }
                });
            };
        }

        error!("TCP listener closed");
    }
}
