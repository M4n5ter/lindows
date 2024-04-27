use base64::Engine;
use crabgrab::prelude::*;
use tauri::{AppHandle, Manager};
use tokio::sync::mpsc::{self,Sender,channel};
use lazy_static::lazy_static;
use parking_lot::Mutex;

lazy_static! {
    static ref FRAME_REQUEST: Mutex<Option<Sender<VideoFrame>>> = Mutex::new(None);
    static ref ACTIVE_STREAM: Mutex<Option<CaptureStream>> = Mutex::new(None);
}

#[tauri::command]
pub async fn begin_capture(title: String) -> Result<(), String> {
	let token = match CaptureStream::test_access(false) {
        Some(token) => token,
        None => CaptureStream::request_access(false).await.expect("Expected capture access")
    };

	// let mut active_stream = ACTIVE_STREAM.lock();
	// let window_map = WINDOW_MAP.lock();

    let window_filter = CapturableWindowFilter{
        desktop_windows: false,
        onscreen_only: true,
    };

    let filter = CapturableContentFilter{ windows: Some(window_filter), displays: false };
    let content = CapturableContent::new(filter).await.expect("Expected capturable content");
    let window = content.windows().find(|window|{
        let app_identifier = window.application().identifier();
        !window.title().is_empty() && (app_identifier.to_lowercase().contains(&title))
    });

    match window {
        Some(window) => {
            let config = CaptureConfig::with_window(window, CapturePixelFormat::Bgra8888).expect("Expected capture config");
            let stream = CaptureStream::new(token, config, |stream_event|{
                if let Ok(StreamEvent::Video(frame)) = stream_event {
                    let mut frame_req = FRAME_REQUEST.lock();
                    if let Some(frame_req) = &mut *frame_req {
                        let req = frame_req.clone();
                        tauri::async_runtime::spawn(async move {
                            req.send(frame).await.expect("Send frame");
                        });
                    }
                }
            }).expect("Expected capture stream");
            let mut active_stream = ACTIVE_STREAM.lock();
            *active_stream = Some(stream);
        }
        None => { return Err("No QQ window found".to_string());}
    }

    Ok(())
}

#[tauri::command]
pub fn end_capture() -> Result<(), String> {
	{
		let mut active_stream = ACTIVE_STREAM.lock();
		if let Some(mut stream) = active_stream.take() {
			// todo... finish recording
			let _ = stream.stop();
		}
	}
	//let app_window = app_handle.get_window("main").unwrap();
	//app_window.eval("window.location.replace('main.html')").unwrap();
	Ok(())
}

pub fn capture_frames(app_handle: AppHandle) {
    let (frame_tx, mut frame_rx) = channel(100); // 创建一个mpsc通道，缓冲区大小为100
    *FRAME_REQUEST.lock() = Some(frame_tx.clone()); // 将发送端存储在全局变量中
    
    tauri::async_runtime::spawn(async move {
        loop {
            if let Some(frame) = frame_rx.recv().await {
                if let Ok(FrameBitmap::BgraUnorm8x4(image_bitmap_bgra8888)) = frame.get_bitmap() {
                    let base64_png = make_scaled_base64_png_from_bitmap(image_bitmap_bgra8888, 3840, 2160);
                    let payload = format!("data:image/png;base64,{}", base64_png);
                    app_handle.emit("stream://video-frame", payload).expect("Expected emit")
                }
            }
        }
    });
}



fn make_scaled_base64_png_from_bitmap(
    bitmap: FrameBitmapBgraUnorm8x4,
    max_width: usize,
    max_height: usize,
) -> String {
    let (mut height, mut width) = (bitmap.width, bitmap.height);
    if width > max_width {
        width = max_width;
        height = ((max_width as f64 / bitmap.width as f64) * bitmap.height as f64).ceil() as usize;
    };

    if height > max_height {
        height = max_height;
        width = ((max_height as f64 / bitmap.height as f64) * bitmap.width as f64).ceil() as usize;
    };

    let mut write_vec = vec![0u8; 0];
    {
        let mut encoder = png::Encoder::new(&mut write_vec, width as u32, height as u32);
        encoder.set_color(png::ColorType::Rgba);
        encoder.set_depth(png::BitDepth::Eight);
        encoder.set_source_gamma(png::ScaledFloat::from_scaled(45455)); // 1.0 / 2.2, scaled by 100000
        encoder.set_source_gamma(png::ScaledFloat::new(1.0 / 2.2)); // 1.0 / 2.2, unscaled, but rounded
        let source_chromaticities = png::SourceChromaticities::new(
            // Using unscaled instantiation here
            (0.31270, 0.32900),
            (0.64000, 0.33000),
            (0.30000, 0.60000),
            (0.15000, 0.06000),
        );
        encoder.set_source_chromaticities(source_chromaticities);
        let mut writer = encoder.write_header().unwrap();
        let mut image_data = vec![0u8; width * height * 4];
        for y in 0..height {
            let sample_y = (bitmap.height * y) / height;
            for x in 0..width {
                let sample_x = (bitmap.width * x) / width;
                let [b, g, r, a] = bitmap.data[sample_x + sample_y * bitmap.width];
                image_data[(x + y * width) * 4] = r;
                image_data[(x + y * width) * 4 + 1] = g;
                image_data[(x + y * width) * 4 + 2] = b;
                image_data[(x + y * width) * 4 + 3] = a;
            }
        }
        writer.write_image_data(&image_data).unwrap();
    }
    base64::prelude::BASE64_STANDARD.encode(write_vec)
}
