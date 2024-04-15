use serde::{Deserialize, Serialize};
use wasm_bindgen::prelude::*;



#[derive(Serialize, Deserialize)]
pub struct GreetArgs<'a> {
    name: &'a str,
}

pub async fn greet(name: &str) -> String {
    let args = GreetArgs { name };
    let result = invoke("greet", serde_wasm_bindgen::to_value(&args).unwrap()).await;
    result.as_string().unwrap()
}

pub mod clipboard {
    use js_sys::Promise;
    use serde::{Deserialize, Serialize};
    use wasm_bindgen_futures::JsFuture;

    use super::{invoke, invoke_without_args};

    #[derive(Serialize, Deserialize)]
    #[serde(rename_all = "camelCase")]
    pub enum ClipboardContents {
        PlainText {
            text: String,
        },
        Image {
            bytes: Vec<u8>,
            width: usize,
            height: usize,
        },
    }

    #[derive(Serialize, Deserialize)]
    #[serde(rename_all = "camelCase")]
    pub enum ClipKind {
        PlainText {
            label: Option<String>,
            text: String,
        },
        Image {
            // tauri::image::JsImage 不支持 Serialize
            // image: tauri::image::JsImage,
            image: Vec<u8>,
        },
        Html {
            html: String,
            alt_html: Option<String>,
        },
    }

    pub async fn read_text() -> String {
        let clip_response = invoke_without_args("plugin:clipboard-manager|read_text").await;
        let clipboard_contents: ClipboardContents = serde_wasm_bindgen::from_value(clip_response).unwrap();
        match clipboard_contents {
            ClipboardContents::PlainText { text } => text,
            _ => panic!("Unexpected clipboard contents"),
        }
    }

    // pub async fn read_image() -> super::Image {
    //     let promise: Promise = invoke_without_args("plugin:clipboard-manager|read_image").await.into();
    //     let rid = JsFuture::from(promise).await.unwrap().as_f64().unwrap() as u32;
    //     super::Image::new(rid)
    // }

    pub async fn write_text(text: &str) {
        let clip_kind = ClipKind::PlainText {
            label: None,
            text: text.to_string(),
        };

        let data = WriteTextData {
            data: clip_kind,
        };

        invoke("plugin:clipboard-manager|write_text", serde_wasm_bindgen::to_value(&data).unwrap()).await;
    }

    // https://github.com/tauri-apps/plugins-workspace/blob/1f9e7ab4a0f97d2fa9d6f139e39e20b5a737cd46/plugins/clipboard-manager/guest-js/index.ts#L33
    #[derive(Serialize)]
    #[serde(rename_all = "camelCase")]
    struct WriteTextData {
        data: ClipKind
    }

    pub async fn write_image(image: Vec<u8>) {
        let clip_kind = ClipKind::Image {
            image,
        };
        invoke("plugin:clipboard-manager|write_image", serde_wasm_bindgen::to_value(&clip_kind).unwrap()).await;
    }
}

#[wasm_bindgen]
extern "C" {
    #[wasm_bindgen(js_namespace = ["window", "__TAURI__", "core"], js_name="invoke")]
    pub async fn invoke(cmd: &str, args: JsValue) -> JsValue;

    #[wasm_bindgen(js_namespace = ["window", "__TAURI__", "core"], js_name="invoke")]
    pub async fn invoke_without_args(cmd: &str) -> JsValue;

    // https://github.com/tauri-apps/tauri/blob/dev/tooling/api/src/image.ts
    #[wasm_bindgen(js_namespace = ["window", "__TAURI__", "api", "image"], js_name="transformImage")]
    pub async fn transform_image(args: JsValue) -> JsValue;
}

// // https://github.com/tauri-apps/tauri/blob/dev/tooling/api/src/core.ts
// // 定义 Resource 和 Image 结构体
// #[wasm_bindgen]
// extern "C" {
//     pub type Resource;
    
//     #[wasm_bindgen(structural, method, getter)]
//     pub fn rid(this: &Resource) -> u32;

//     #[wasm_bindgen(method, js_name = "close")]
//     pub async fn close_js(this: &Resource);
// }

// // https://github.com/tauri-apps/tauri/blob/dev/tooling/api/src/image.ts
// #[wasm_bindgen(extends = Resource)]
// extern "C" {
//     pub type Image;
    
//     #[wasm_bindgen(constructor)]
//     pub fn new(rid: u32) -> Image;
// }