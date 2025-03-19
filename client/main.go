package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

// Configurações globais
const (
	serverURL    = "ws://127.0.0.1:3030/ws"
	clientName   = "Microservice1"
	retryDelay   = 5 * time.Second // Tempo de espera antes de tentar reconectar
	pingInterval = 8 * time.Second // Tempo entre cada ping
	cmdInterval  = 8 * time.Second // Tempo entre cada comando aleatório
)

func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	for {
		fmt.Println("Tentando conectar ao servidor...")
		conn, err := connectWebSocket()
		if err != nil {
			fmt.Println("Falha ao conectar. Tentando novamente em", retryDelay)
			time.Sleep(retryDelay)
			continue
		}

		// Se conectou com sucesso, começa a rotina de mensagens
		if startClient(conn, interrupt) {
			fmt.Println("Tentando reconectar...")
			time.Sleep(retryDelay) // Aguarda antes de tentar reconectar
		} else {
			break
		}
	}

	fmt.Println("Cliente encerrado.")
}

// Conecta ao WebSocket e retorna a conexão ativa
func connectWebSocket() (*websocket.Conn, error) {
	conn, _, err := websocket.DefaultDialer.Dial(serverURL, nil)
	if err != nil {
		return nil, err
	}

	fmt.Println("Conectado ao servidor WebSocket!")

	// Enviar nome do cliente após a conexão
	err = conn.WriteMessage(websocket.TextMessage, []byte("name:"+clientName))
	if err != nil {
		fmt.Println("Erro ao enviar nome:", err)
		conn.Close()
		return nil, err
	}

	return conn, nil
}

// Gerencia a comunicação com o servidor WebSocket
func startClient(conn *websocket.Conn, interrupt chan os.Signal) bool {
	defer conn.Close()

	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				fmt.Println("Conexão perdida:", err)
				return
			}
			fmt.Printf("Mensagem recebida: %s\n", message)
		}
	}()

	// Configura os timers para ping e comandos aleatórios
	pingTicker := time.NewTicker(pingInterval)
	defer pingTicker.Stop()
	cmdTicker := time.NewTicker(cmdInterval)
	defer cmdTicker.Stop()

	for {
		select {
		case <-done:
			return true // Conexão foi fechada, tentar reconectar

		case <-pingTicker.C:
			fmt.Println("Enviando ping para o servidor")
			err := conn.WriteMessage(websocket.TextMessage, []byte("ping"))
			if err != nil {
				fmt.Println("Erro ao enviar ping:", err)
				return true
			}

		case <-cmdTicker.C:
			commands := []string{"cmd:get_time", "cmd:random_num"}
			cmd := commands[rand.Intn(len(commands))]
			fmt.Printf("Enviando comando: %s\n", cmd)
			err := conn.WriteMessage(websocket.TextMessage, []byte(cmd))
			if err != nil {
				fmt.Println("Erro ao enviar comando:", err)
				return true
			}

		case <-interrupt:
			fmt.Println("\nEncerrando conexão...")
			err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				fmt.Println("Erro ao fechar conexão:", err)
			}
			return false
		}
	}
}
