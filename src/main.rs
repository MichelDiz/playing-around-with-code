mod server;

#[tokio::main]
async fn main() {
    server::start_websocket_server().await;
}
