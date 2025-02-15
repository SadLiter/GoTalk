package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"sync"
)

type Client struct {
	conn net.Conn
	ch   chan string
}

var (
	clients   = make(map[net.Conn]*Client)
	clientsMu sync.Mutex
)

// handleConnection manages a new client connection.
func handleConnection(conn net.Conn) {
	defer conn.Close()
	client := &Client{
		conn: conn,
		ch:   make(chan string, 10),
	}
	clientsMu.Lock()
	clients[conn] = client
	clientsMu.Unlock()

	// Goroutine for sending messages to the client.
	go func() {
		for msg := range client.ch {
			_, _ = conn.Write([]byte(msg + "\n"))
		}
	}()

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		msg := scanner.Text()
		broadcast(msg, conn)
	}

	clientsMu.Lock()
	delete(clients, conn)
	clientsMu.Unlock()
}

// broadcast sends the given message to all connected clients except the sender.
func broadcast(message string, sender net.Conn) {
	clientsMu.Lock()
	defer clientsMu.Unlock()
	for conn, client := range clients {
		if conn != sender {
			client.ch <- message
		}
	}
}

func main() {
	listener, err := net.Listen("tcp", ":9000")
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()
	fmt.Println("Server started on :9000")
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting connection:", err)
			continue
		}
		go handleConnection(conn)
	}
}
