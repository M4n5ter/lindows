use leptos::*;

#[component(transparent)]
pub fn Close() -> impl IntoView {
    view! {
        <svg
            xmlns="http://www.w3.org/2000/svg"
            fill="none"
            viewBox="0 0 24 24"
            class="inline-block w-4 h-4 stroke-current"
        >
            <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M6 18L18 6M6 6l12 12"
            ></path>
        </svg>
    }
}

#[component(transparent)]
pub fn Maximize() -> impl IntoView {
    view! {
        <svg
            class="h-5 w-5"
            fill="currentColor"
            xmlns="http://www.w3.org/2000/svg"
            viewBox="0 0 1024 1024"
        >
            <path d="M864 920H156.64a56 56 0 0 1-56-56V156.64a56 56 0 0 1 56-56H864a56 56 0 0 1 56 56V864a56 56 0 0 1-56 56zM156.64 148.64a8 8 0 0 0-8 8V864a8 8 0 0 0 8 8H864a8 8 0 0 0 8-8V156.64a8 8 0 0 0-8-8z"></path>
        </svg>
    }
}

#[component(transparent)]
pub fn Minimize() -> impl IntoView {
    view! {
        <svg
            class="h-5 w-5"
            fill="currentColor"
            xmlns="http://www.w3.org/2000/svg"
            viewBox="0 0 1024 1024"
        >
            <path d="M920 544H112a24 24 0 0 1 0-48h808a24 24 0 0 1 0 48z"></path>
        </svg>
    }
}
