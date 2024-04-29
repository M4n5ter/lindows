use leptos::provide_context;

pub fn provide_window() {
    provide_context(MainWindow {
        is_maximized: false,
    });
}

#[derive(Clone, Copy, Debug)]
pub struct MainWindow {
    pub is_maximized: bool,
}
