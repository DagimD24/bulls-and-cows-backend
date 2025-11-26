package main

import (
	"encoding/json"
	"math/rand"
	"time"
)

// Helper to send JSON easier
func sendJSON(p *Player, data interface{}) {
	js, _ := json.Marshal(data)
	p.Send <- js
}

func sendError(p *Player, msg string) {
	sendJSON(p, map[string]string{"type": "error", "message": msg})
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func generateRandomString(length int) string {
	b := make([]byte, length)
	for i := range b {
		// Pick a random character from the charset
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
