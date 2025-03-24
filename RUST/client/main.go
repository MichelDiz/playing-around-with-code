package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

// Global configurations
const (
	serverURL    = "ws://127.0.0.1:3030/ws"
	clientName   = "Microservice1"
	retryDelay   = 5 * time.Second // Wait time before trying to reconnect
	pingInterval = 8 * time.Second // Time between each ping
	cmdInterval  = 8 * time.Second // Time between each random command
)

func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	for {
		fmt.Println("Attempting to connect to the server...")
		conn, err := connectWebSocket()
		if err != nil {
			fmt.Println("Failed to connect. Retrying in", retryDelay)
			time.Sleep(retryDelay)
			continue
		}

		// If connected successfully, start the message routine
		if startClient(conn, interrupt) {
			fmt.Println("Attempting to reconnect...")
			time.Sleep(retryDelay) // Wait before trying to reconnect
		} else {
			break
		}
	}

	fmt.Println("Client terminated.")
}

// Connects to the WebSocket and returns the active connection
func connectWebSocket() (*websocket.Conn, error) {
	conn, _, err := websocket.DefaultDialer.Dial(serverURL, nil)
	if err != nil {
		return nil, err
	}

	fmt.Println("Connected to the WebSocket server!")

	// Send client name after connecting
	err = conn.WriteMessage(websocket.TextMessage, []byte("name:"+clientName))
	if err != nil {
		fmt.Println("Error sending name:", err)
		conn.Close()
		return nil, err
	}

	return conn, nil
}

// Manages communication with the WebSocket server
func startClient(conn *websocket.Conn, interrupt chan os.Signal) bool {
	defer conn.Close()

	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				fmt.Println("Connection lost:", err)
				return
			}
			fmt.Printf("Message received: %s\n", message)
		}
	}()

	// Set up timers for ping and random commands
	pingTicker := time.NewTicker(pingInterval)
	defer pingTicker.Stop()
	cmdTicker := time.NewTicker(cmdInterval)
	defer cmdTicker.Stop()

	for {
		select {
		case <-done:
			return true // Connection was closed, try to reconnect

		case <-pingTicker.C:
			fmt.Println("Sending ping to the server")
			err := conn.WriteMessage(websocket.TextMessage, []byte("ping"))
			if err != nil {
				fmt.Println("Error sending ping:", err)
				return true
			}

		case <-cmdTicker.C:
			commands := []string{"cmd:get_time", "cmd:random_num"}
			cmd := commands[rand.Intn(len(commands))]
			fmt.Printf("Sending command: %s\n", cmd)
			err := conn.WriteMessage(websocket.TextMessage, []byte(cmd))
			if err != nil {
				fmt.Println("Error sending command:", err)
				return true
			}

		case <-interrupt:
			fmt.Println("\nClosing connection...")
			err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				fmt.Println("Error closing connection:", err)
			}
			return false
		}
	}
}
