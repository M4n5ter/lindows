use leptos::*;

#[component]
pub fn Sidebar() -> impl IntoView {
    view! {
        <div class="flex flex-row">
            <div class="bg-base-100">
                <ul class="menu">
                    <li>
                        <a href="#" class="menu-item">
                            <i class="icon icon-home"></i>
                            <span>"Home"</span>
                        </a>
                    </li>
                    <li>
                        <a href="#" class="menu-item">
                            <i class="icon icon-user"></i>
                            <span>"Profile"</span>
                        </a>
                    </li>
                    <li>
                        <a href="#" class="menu-item">
                            <i class="icon icon-settings"></i>
                            <span>"Settings"</span>
                        </a>
                    </li>
                </ul>
            </div>

        </div>
    }
}
