use liblindows::conn::Conn;
use tokio::net::TcpListener;
use tracing::Level;
use tracing_subscriber::FmtSubscriber;

#[tokio::main]
async fn main() {
    liblindows::cnsoft();

    let subscriber = FmtSubscriber::builder()
        .with_max_level(Level::DEBUG)
        .finish();

    tracing::subscriber::set_global_default(subscriber).expect("setting default subscriber failed");

    let tcp_listener = TcpListener::bind("0.0.0.0:11111")
        .await
        .expect("Failed to bind to port 11111");

    let conn = Conn::new(tcp_listener).await;
    conn.set_on_ice_candidate().await;
    conn.serve().await;
}
