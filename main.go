package main

import (
	"log"
	"net"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

var mainHub *Hub

func main() {
	mainHub = newHub()

	http.HandleFunc("/ws", wsHandler)

	// Load variables from .env if present
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8001"
	}

	addr := ":" + port
	log.Printf("Attempting to start server on %s...", addr)

	// Try to bind to the address first so we can distinguish startup errors
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to bind to %s: %v", addr, err)
	}

	log.Printf("Server successfully started and listening on %s", addr)

	if err := http.Serve(ln, nil); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server stopped with error: %v", err)
	}
}
