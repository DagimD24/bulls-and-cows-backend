package main

import (
	"encoding/json"
	"log"

	"github.com/gorilla/websocket"
)

func (p *Player) writeLoop() {
	// When this loop exits, we must close the connection
	defer p.Conn.Close()

	for {
		message, ok := <-p.Send
		if !ok {
			// The Hub closed the channel, so we're done
			p.Conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}
		// Write the message to the socket
		if err := p.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
			log.Printf("Write error: %v", err)
			return // Stop on write error
		}
	}
}

func (p *Player) readLoop() {
	// When this loop exits, unregister the player and close the connection
	defer func() {
		mainHub.unregister(p)
		p.Conn.Close()
	}()

	for {
		// Read a message from the socket
		_, rawMessage, err := p.Conn.ReadMessage()
		if err != nil {
			// This is the normal way a client disconnects
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Read error: %v", err)
			}
			break // Exit loop, which triggers the defer
		}

		// Parse the message
		var msg Message
		if err := json.Unmarshal(rawMessage, &msg); err != nil {
			log.Printf("JSON parse error: %v", err)
			continue // Keep looping
		}

		// Send the parsed message to the Hub for processing
		mainHub.processMessage(p, &msg)
	}
}
