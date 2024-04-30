use std::time::{self, Duration};

use crabgrab::prelude::*;
use lindows_utils::common::{
    self,
    codec::{self, EncoderApi},
    vpxcodec, PixelBuffer,
};

fn main() {
    let runtime = tokio::runtime::Builder::new_multi_thread().build().unwrap();
    let future = runtime.spawn(async {
        let start = time::Instant::now();
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
                    || app_identifier.to_lowercase().contains("edge"))
        });

        match window {
            Some(window) => {
                println!("capturing window: {}", window.title());
                let (width, height) = (window.rect().size.width, window.rect().size.height);

                let config =
                    CaptureConfig::with_window(window, CaptureStream::supported_pixel_formats()[0])
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

                let mut stream =
                    CaptureStream::new(token, config, move |stream_event| match stream_event {
                        Ok(event) => {
                            if let StreamEvent::Video(frame) = event {
                                let mut yuv = Vec::<u8>::new();
                                let mut mid_data = Vec::<u8>::new();

                                let now = time::Instant::now();
                                let time = now - start;
                                let ms =
                                    (time.as_secs() * 1000 + time.subsec_millis() as u64) as i64;

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
                                    for ref frame in vpx
                                        .encode(
                                            ms,
                                            frame.yuv().expect("Shoud have yuv data"),
                                            crate::common::STRIDE_ALIGN,
                                        )
                                        .expect("Failed to encode frame")
                                    {
                                        println!("frame encoded: {}", frame.data.len());
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
                tokio::task::block_in_place(|| std::thread::sleep(Duration::from_millis(4000)));
                stream.stop().unwrap();
            }
            None => {
                println!("Failed to find window");
            }
        }
    });
    runtime.block_on(future).unwrap();
    runtime.shutdown_timeout(Duration::from_millis(10000));
}
