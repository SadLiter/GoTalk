package main

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"

	"github.com/fatih/color"
)

// deriveKey creates a 32-byte AES key from the passphrase using SHA-256.
func deriveKey(passphrase string) []byte {
	hash := sha256.Sum256([]byte(passphrase))
	return hash[:]
}

// encrypt encrypts the plaintext using AES-GCM and returns a base64-encoded string.
func encrypt(plaintext []byte, gcm cipher.AEAD) (string, error) {
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decrypt decodes the base64 input and decrypts it using AES-GCM.
func decrypt(encoded string, gcm cipher.AEAD) ([]byte, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: client <server address:port>")
		os.Exit(1)
	}
	serverAddr := os.Args[1]
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// Prompt user for passphrase
	fmt.Print("Enter passphrase: ")
	reader := bufio.NewReader(os.Stdin)
	passphrase, _ := reader.ReadString('\n')
	passphrase = strings.TrimSpace(passphrase)
	key := deriveKey(passphrase)

	block, err := aes.NewCipher(key)
	if err != nil {
		log.Fatal(err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		log.Fatal(err)
	}

	// Setup colored output for CLI
	peerColor := color.New(color.FgGreen).SprintFunc()
	selfColor := color.New(color.FgCyan).SprintFunc()

	// Goroutine to read and decrypt messages from server.
	go func() {
		serverReader := bufio.NewReader(conn)
		for {
			line, err := serverReader.ReadString('\n')
			if err != nil {
				log.Println("Error reading from server:", err)
				os.Exit(1)
			}
			line = strings.TrimSpace(line)
			decrypted, err := decrypt(line, gcm)
			if err != nil {
				log.Println("Error decrypting message:", err)
				continue
			}
			fmt.Printf("%s: %s\n", peerColor("Peer"), string(decrypted))
		}
	}()

	// Main loop: read user input, encrypt and send.
	for {
		fmt.Print(selfColor("You: "))
		text, err := reader.ReadString('\n')
		if err != nil {
			log.Println("Error reading input:", err)
			continue
		}
		text = strings.TrimSpace(text)
		if text == "" {
			continue
		}
		encrypted, err := encrypt([]byte(text), gcm)
		if err != nil {
			log.Println("Error encrypting message:", err)
			continue
		}
		// Send the encrypted message (newline as delimiter)
		_, err = conn.Write([]byte(encrypted + "\n"))
		if err != nil {
			log.Println("Error sending message:", err)
			continue
		}

		fmt.Printf("\n%s: %s\n\n%s", peerColor("Peer"), string(decrypted), selfColor("You: "))
	}
}
