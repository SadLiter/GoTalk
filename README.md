Encrypted CLI Chat

A secure, encrypted chat application written in Go that guarantees end-to-end encryption between two users. It runs in the command line with colored output to distinguish between your messages and your peer's messages.

## Project Structure

```
encrypted-chat/
├── cmd
│   ├── server.go      # TCP server that relays encrypted messages
│   └── client.go      # CLI client that encrypts and decrypts messages
├── go.mod             # Go module file
└── README.md          # This file
```

## Requirements

- Go 1.20 or later

## How to Run

1. **Start the Server:**

   Open a terminal and run:
   ```bash
   go run cmd/server.go
   ```
   The server listens on port 9000.

2. **Start the Clients:**

   Open two separate terminals (for two users) and run:
   ```bash
   go run cmd/client.go localhost:9000
   ```
3. **Chat:**

   Enter the same passphrase in both clients to enable secure communication. Then, type your messages and enjoy the chat!