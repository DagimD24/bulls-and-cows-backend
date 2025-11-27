package main

import (
	"encoding/json"

	"github.com/gorilla/websocket"
)

type Message struct {
	Type    string
	Payload json.RawMessage
}

type Player struct {
	Conn     *websocket.Conn
	Game     *Game
	Send     chan []byte
	Username string
	IsReady  bool
}

type Game struct {
	GameID    string  `json:"game_id"`
	Player1   *Player `json:"-"`
	Player2   *Player `json:"-"`
	GameState string  `json:"game_state"`

	p1Guess  string
	p2Guess  string
	p1Result *ResultPayload
	p2Result *ResultPayload
}

// ResultPayload is the struct for the "submit_result" payload
type ResultPayload struct {
	Bulls int  `json:"bulls"`
	Cows  int  `json:"cows"`
	IsWin bool `json:"is_win"`
}
