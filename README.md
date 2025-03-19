# playing-around-with-rust

Este projeto demonstra uma comunicação via WebSocket utilizando um servidor em Rust e um cliente em Go.

## Visão Geral

- **Servidor (Rust):**  
  Utiliza o framework [Warp](https://github.com/seanmonstar/warp) para criar um servidor WebSocket que escuta em `ws://127.0.0.1:3030/ws`. O servidor gerencia múltiplos clientes, envia mensagens de "ping" periodicamente e responde a comandos específicos, como:
  - `cmd:get_time`: Retorna a hora atual do servidor.
  - `cmd:random_num`: Retorna um número aleatório entre 1 e 100.

- **Cliente (Go):**  
  Após a conexão, o cliente envia seu nome e periodicamente:
  - Envia pings para manter a conexão ativa.
  - Envia comandos aleatórios (`cmd:get_time` ou `cmd:random_num`) para o servidor.
  - Trata mensagens recebidas do servidor e exibe no console.

## Estrutura do Projeto

```
├── .gitignore
├── .vscode
│   ├── launch.json          # Configurações de depuração para Rust e Go
│   └── tasks.json
├── Cargo.lock
├── Cargo.toml
├── client
│   ├── client/go.mod
│   ├── client/go.sum
│   └── client/main.go       # Código do cliente em Go
├── docker-compose.yml       # Configuração opcional para orquestração via Docker
├── prometheus.yml           # Configuração opcional para monitoramento com Prometheus
└── src
├── src/main.rs         # Ponto de entrada do servidor em Rust
└── src/server
└── src/server/mod.rs  # Lógica do servidor WebSocket
```

## Como Executar

### Servidor em Rust

1. Certifique-se de que o [Rust](https://www.rust-lang.org/tools/install) esteja instalado.
2. Na raiz do projeto, execute:
   ```sh
   cargo run --bin playing-around-with-rust
   ```

Ou utilize a configuração de depuração do VS Code.

Cliente em Go
	1.	Navegue até a pasta client:

cd client

	2.	Certifique-se de que o Go esteja instalado.
	3.	Execute:

```sh
go run main.go
```

Ou utilize a configuração de depuração do VS Code.

Depuração Conjunta

As configurações do VS Code (arquivos launch.json e tasks.json) permitem iniciar o servidor e o cliente simultaneamente através de um compound launch, facilitando o processo de depuração.

