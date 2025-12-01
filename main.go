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

	// Load variables from .env if present
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8001"
	}

	addr := "0.0.0.0:" + port
	log.Printf("Attempting to start server on %s...", addr)

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to bind to %s: %v", addr, err)
	}

	http.HandleFunc("/", withLogging(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Go backend is running!"))
	}))

	http.HandleFunc("/ws", withLogging(wsHandler))

	log.Printf("Server successfully started and listening on %s", addr)
	if err := http.Serve(ln, nil); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server stopped with error: %v", err)
	}
}

func withLogging(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("HTTP Request: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
		next(w, r)
	}
}
