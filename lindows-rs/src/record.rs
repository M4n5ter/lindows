use std::time::{Duration, Instant};

use crabgrab::{
    capturable_content::{CapturableContent, CapturableContentFilter},
    capture_stream::{CaptureConfig, CaptureStream, StreamEvent},
    feature::bitmap::VideoFrameBitmap,
};
use hbb_common::message_proto::EncodedVideoFrame;
use scrap::codec::{EncoderApi, EncoderCfg, Quality as Q};
use scrap::{vpxcodec as vpx_encode, Frame, PixelBuffer};
use tokio::sync::{mpsc::Receiver, mpsc::UnboundedSender};
use tracing::{info, instrument};

#[instrument]
pub fn record(frame_sender: UnboundedSender<(EncodedVideoFrame, u64)>, mut signal: Receiver<()>) {
    let runtime = tokio::runtime::Builder::new_multi_thread().build().unwrap();
    let future = runtime.spawn(async move {
        tokio::select! {
            _ = signal.recv() => {
                println!("Received signal to start recording");
            }
        }

        let mut start = Instant::now();

        let token = match CaptureStream::test_access(true) {
            Some(token) => token,
            None => CaptureStream::request_access(true)
                .await
                .expect("Expected capture access"),
        };
        let filter = CapturableContentFilter::NORMAL_WINDOWS;
        let content = CapturableContent::new(filter).await.unwrap();
        let window = content.windows().find(|window| {
            let app_identifier = window.application().identifier();
            app_identifier.to_lowercase().contains("qq")
        });

        match window {
            Some(window) => {
                info!("capturing window {}", window.title(),);

                let config =
                    CaptureConfig::with_window(window, CaptureStream::supported_pixel_formats()[0])
                        .unwrap();

                let mut yuv = Vec::new();
                let mut mid_data = Vec::new();
                let mut stream =
                    CaptureStream::new(token, config, move |stream_event| match stream_event {
                        Ok(event) => {
                            let now = Instant::now();
                            let time = now - start;
                            let ms = (time.as_secs() * 1000 + time.subsec_millis() as u64) as i64;
                            start = now;

                            if let StreamEvent::Video(frame) = event {
                                let quality = Q::Balanced;
                                let mut vpx = vpx_encode::VpxEncoder::new(
                                    EncoderCfg::VPX(vpx_encode::VpxEncoderConfig {
                                        width: frame.size().width as u32,
                                        height: frame.size().height as u32,
                                        quality,
                                        codec: vpx_encode::VpxVideoCodecId::VP8,
                                        keyframe_interval: None,
                                    }),
                                    false,
                                )
                                .unwrap();

                                if let crabgrab::feature::bitmap::FrameBitmap::BgraUnorm8x4(
                                    bitmap,
                                ) = frame.get_bitmap().expect("Failed to get bitmap")
                                {
                                    let bitmap_data = bitmap
                                        .data
                                        .iter()
                                        .flat_map(|pixel| pixel.to_vec())
                                        .collect::<Vec<_>>();
                                    let frame_pixelbuf = Frame::PixelBuffer(PixelBuffer::new(
                                        &bitmap_data,
                                        frame.size().width as usize,
                                        frame.size().height as usize,
                                    ));
                                    let frame = frame_pixelbuf
                                        .to(vpx.yuvfmt(), &mut yuv, &mut mid_data)
                                        .expect("Failed to convert pixel buffer to yuv");

                                    if let Ok(mut video_frame) = vpx.encode_to_message(frame, ms) {
                                        video_frame.take_vp8s().frames.into_iter().for_each(
                                            |vps| {
                                                frame_sender.send((vps, ms as u64)).unwrap();
                                            },
                                        );
                                    };
                                }
                            }
                        }
                        Err(error) => {
                            println!("Stream error: {:?}", error);
                        }
                    })
                    .unwrap();

                tokio::task::block_in_place(|| std::thread::sleep(Duration::from_secs(2000)));
                stream.stop().unwrap();
            }
            None => {
                println!("Failed to find window");
            }
        }
    });
    runtime.block_on(future).unwrap();
    runtime.shutdown_timeout(Duration::from_secs(100000));
}
