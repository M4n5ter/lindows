use leptos::{leptos_dom::logging::console_log, *};
use wasm_bindgen::JsCast as _;

use crate::state::session::Session;

#[allow(non_snake_case)]
#[component]
pub fn Sidebar() -> impl IntoView {
    let session = expect_context::<RwSignal<Session>>();

    view! {
        <div class="flex flex-row">
            <div class="bg-base-100">
                <ul class="menu space-y-2">
                    <li>
                        <button class="btn btn-secondary" onclick="set_remote_address.showModal()">
                            <span class="text-white">"设置"</span>
                        </button>
                        <dialog id="set_remote_address" class="modal">
                            <div class="modal-box w-11/12 max-w-5xl m-4">
                                <h3 class="font-bold text-lg">"Lindows 设置"</h3>
                                <p class="py-4">"对等点 websocket 地址:"</p>
                                <input
                                    type="text"
                                    class="input input-bordered input-primary w-full max-w-xs"
                                    value=move || session.get_untracked().address.get_untracked()
                                    on:input=move |event| {
                                        session
                                            .get_untracked()
                                            .address
                                            .set_untracked(event_target_value(&event))
                                    }
                                />

                                <div class="modal-action">
                                    <form method="dialog">
                                        <button class="btn">"确定"</button>
                                    </form>
                                </div>
                            </div>
                        </dialog>
                    </li>
                    <li>
                        <button
                            class="btn btn-primary"
                            on:click=move |_| {
                                console_log("发起 Offer");
                                spawn_local(async move {
                                    session.get_untracked().send_offer().await;
                                });
                            }
                        >

                            <span class="text-white">"刷新"</span>
                        </button>

                    </li>
                    <li>
                        <button
                            class="btn btn-primary"
                            on:click=move |_| {
                                console_log("连接");
                                session.get_untracked().connect();
                            }
                        >

                            <span class="text-white">"连接"</span>
                        </button>
                    </li>

                    <li>
                        <button
                            class="btn btn-primary"
                            on:click=move |_| {
                                let window = window();
                                let document = window.document().expect("should have a Document");
                                let video_element = document
                                    .get_element_by_id("screen")
                                    .expect("Get video element");
                                let video_html_element = video_element
                                    .dyn_into::<web_sys::HtmlVideoElement>()
                                    .expect("Video element");
                                video_html_element
                                    .request_fullscreen()
                                    .expect("Request full screen");
                            }
                        >

                            <span class="text-white">"全屏"</span>
                        </button>
                    </li>
                </ul>
            </div>

        </div>
    }
}
