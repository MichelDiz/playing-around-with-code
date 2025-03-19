use chrono::Utc;
use tokio::sync::Mutex;
use tokio::time::{self, Duration};
use warp::ws::{WebSocket, Ws, Message};
use warp::Filter;
use futures_util::{StreamExt, SinkExt};
use std::collections::HashMap;
use std::sync::Arc;
use rand::random;
use uuid::Builder;
use rand::Rng;

#[derive(Debug, Clone)]
struct ClientInfo {
    name: String,
    sender: tokio::sync::mpsc::Sender<String>,
}

#[derive(Debug, Clone)]
struct AppState {
    clients: Arc<Mutex<HashMap<String, ClientInfo>>>,
}

pub async fn start_websocket_server() {
    let state = AppState {
        clients: Arc::new(Mutex::new(HashMap::new())),
    };

    let ws_route = warp::path("ws")
        .and(warp::ws())
        .and(warp::any().map(move || state.clone()))
        .map(|ws: Ws, state| ws.on_upgrade(move |socket| handle_connection(socket, state)));

    let routes = warp::get().and(ws_route);

    println!("WebSocket Server running on ws://127.0.0.1:3030/ws");
    warp::serve(routes).run(([127, 0, 0, 1], 3030)).await;
}

async fn handle_connection(ws: WebSocket, state: AppState) {
    let (mut sender, mut receiver) = ws.split();
    let (tx, mut rx) = tokio::sync::mpsc::channel::<String>(10);

    let uuid = Builder::from_bytes(random()).into_uuid();
    let client_id = uuid.to_string();
    println!("New client connected: {}", client_id);

    let mut client_name = String::from("Unknown");

    // Wait for the client's first message to capture the name
    if let Some(Ok(msg)) = receiver.next().await {
        if let Ok(text) = msg.to_str() {
            if text.starts_with("name:") {
                client_name = text[5..].to_string();
                println!("Client {} set the name to: {}", client_id, client_name);
            }
        }
    }

    state.clients.lock().await.insert(client_id.clone(), ClientInfo { name: client_name.clone(), sender: tx });

    let state_clone = state.clone();
    let client_id_clone = client_id.clone();
    let _client_name_clone = client_name.clone();
    let clients_for_ping = state_clone.clients.clone();
    let ping_task = tokio::spawn(async move {
        let mut interval = time::interval(Duration::from_secs(10));
        loop {
            interval.tick().await;
            let clients = clients_for_ping.lock().await;
            if let Some(client_info) = clients.get(&client_id_clone) {
                let _ = client_info.sender.send("ping".to_string()).await;
            } else {
                break;
            }
        }
    });

    let client_id_clone = client_id.clone();
    let client_name_clone = client_name.clone();
    let state_clone_for_receiver = state_clone.clients.clone();
    let receiver_task = tokio::spawn(async move {
        while let Some(Ok(msg)) = receiver.next().await {
            if let Ok(text) = msg.to_str() {
                println!("Received from {} (UID {}): {}", client_name_clone, client_id_clone, text);

                if text == "ping" {
                    continue;
                }

                let response = match text {
                    "cmd:get_time" => {
                        let now = Utc::now();
                        format!("Server Time: {}", now)
                    }
                    "cmd:random_num" => {
                        let num: u32 = rand::rng().random_range(1..=100);
                        format!("Random Number: {}", num)
                    }
                    _ => "Command not recognized".to_string(),
                };

                if let Some(client) = state_clone_for_receiver.lock().await.get(&client_id_clone) {
                    let _ = client.sender.send(response).await;
                }
            }
        }
        println!("Client {} ({}) disconnected", client_name_clone, client_id_clone);
    });


    let _client_id_clone = client_id.clone();
    let _client_name_clone = client_name.clone();
    let sender_task = tokio::spawn(async move {
        while let Some(msg) = rx.recv().await {
            if sender.send(Message::text(msg)).await.is_err() {
                break;
            }
        }
    });

    tokio::select! {
        _ = receiver_task => {}
        _ = sender_task => {}
        _ = ping_task => {}
    }

    let mut clients = state.clients.lock().await;
    if let Some(client) = clients.remove(&client_id) {
        println!("Client removed from state: {} (UID {})", client.name, client_id);
    }
}
