use js_sys::Promise;
use wasm_bindgen::prelude::*;

pub mod clipboard {
    use crate::electron::{electronReadText, electronWriteText};

    pub fn write_text(text: &str) {
        electronWriteText(text);
    }

    pub fn read_text() -> String {
        electronReadText()
    }
}

pub mod window {
    use wasm_bindgen_futures::JsFuture;

    use crate::electron::{
        closeWindow, createWindow, getMainWindowId, isMainWindow, maximizeWindow, minimizeWindow,
        unmaximizeWindow,
    };

    pub async fn get_main_window_id() -> i32 {
        let promise = getMainWindowId();
        let id = JsFuture::from(promise).await.expect("Get main window id");
        id.as_f64().expect("Get main window id") as i32
    }

    pub async fn is_main_window() -> bool {
        let promise = isMainWindow();
        let is_main = JsFuture::from(promise).await.expect("Is main window");
        is_main.as_bool().expect("Is main window")
    }

    pub async fn create(width: i32, height: i32) -> i32 {
        let promise = createWindow(width, height);
        let id = JsFuture::from(promise).await.expect("Create window");
        id.as_f64().expect("Create window") as i32
    }

    pub fn minimize(id: i32) {
        minimizeWindow(id)
    }

    pub fn maximize(id: i32) {
        maximizeWindow(id)
    }

    pub fn unmaximize(id: i32) {
        unmaximizeWindow(id)
    }

    pub fn close(id: i32) {
        closeWindow(id)
    }
}

#[wasm_bindgen]
extern "C" {
    // Clipboard

    #[wasm_bindgen(js_namespace = ["window", "clipboard"], js_name="writeText")]
    pub(self) fn electronWriteText(text: &str);

    #[wasm_bindgen(js_namespace = ["window", "clipboard"], js_name="readText")]
    pub(self) fn electronReadText() -> String;

    // Window

    #[wasm_bindgen(js_namespace = ["window","electronWindow"], js_name="createWindow")]
    pub(self) fn createWindow(width: i32, height: i32) -> Promise;

    #[wasm_bindgen(js_namespace = ["window","electronWindow"], js_name="minimizeWindow")]
    pub(self) fn minimizeWindow(id: i32);

    #[wasm_bindgen(js_namespace = ["window","electronWindow"], js_name="maximizeWindow")]
    pub(self) fn maximizeWindow(id: i32);

    #[wasm_bindgen(js_namespace = ["window","electronWindow"], js_name="unmaximizeWindow")]
    pub(self) fn unmaximizeWindow(id: i32);

    #[wasm_bindgen(js_namespace = ["window","electronWindow"], js_name="closeWindow")]
    pub(self) fn closeWindow(id: i32);

    #[wasm_bindgen(js_namespace = ["window","electronWindow"], js_name="getMainWindowId")]
    pub(self) fn getMainWindowId() -> Promise;

    #[wasm_bindgen(js_namespace = ["window","electronWindow"], js_name="isMainWindow")]
    pub(self) fn isMainWindow() -> Promise;

}
