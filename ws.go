package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal("failed to upgrade the connection: ", err)
		return
	}

	player := &Player{
		Conn: conn,
		Send: make(chan []byte),
	}

	go player.readLoop()
	go player.writeLoop()

}
