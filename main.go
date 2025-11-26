package main

import (
	"log"
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
	log.Printf("Server starting on port %s...", port)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
