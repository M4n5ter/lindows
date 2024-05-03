use anyhow::Context;
use bytes::Bytes;
use std::{
    sync::mpsc,
    time::{self, Duration, SystemTime},
};
use tokio::net::TcpListener;
use webrtc::media::Sample;

use crate::conn::{self, Conn};
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

pub struct Server {
    conn: Conn,
}

impl Server {
    pub async fn new() -> Self {
        let tcp_listener = TcpListener::bind("0.0.0.0:11111")
            .await
            .expect("Failed to bind to port 11111");
        let conn = Conn::new(tcp_listener).await;
        Self { conn }
    }

    #[inline]
    pub async fn serve(&mut self) -> anyhow::Result<()> {
        let start = time::Instant::now();
        // self.conn.set_video_track().await?;
        self.conn.set_on_data_channel().await;
        self.conn.set_on_ice_candidate().await;
        self.conn.set_on_peer_connection_state_change().await;

        let (sender, receiver) = mpsc::channel::<Sample>();

        let conn_video_track = conn::VIDEO_TRACK.clone();

        tokio::spawn(async move {
            while let Ok(sample) = receiver.recv() {
                let mut video_track = conn_video_track.lock().await;
                if let Some(vt) = video_track.take() {
                    let _ = vt.write_sample(&sample).await;
                }
            }
        });

        // tokio::spawn(async move {
        // let mut child = Command::new("ffplay")
        //     .arg("-vcodec")
        //     .arg("vp8")
        //     .arg("-video_size")
        //     .arg("1280x720")
        //     .arg("-framerate")
        //     .arg("30")
        //     .arg("-")
        //     .stdin(std::process::Stdio::piped())
        //     .spawn()
        //     .expect("Failed to spawn ffplay");

        // let stdin = child.stdin.as_mut().expect("Failed to open stdin");
        // while let Ok(sample) = receiver.recv() {
        // info!("sending video sample: {}", sample.len());
        // stdin.write_all(&sample).expect("Failed to write to stdin");
        // stdin.flush().expect("Failed to flush stdin");
        // }
        // child.wait().expect("Failed to wait for ffplay");
        // });

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
                        std::thread::sleep(Duration::from_millis(60000))
                    });
                    stream.stop().unwrap();
                }
                None => {
                    println!("Failed to find window");
                }
            }
        });

        self.conn.serve().await;
        Ok(())
    }
}
