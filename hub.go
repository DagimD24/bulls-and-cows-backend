package main

import (
	"encoding/json"
	"log"
	"sync"
)

type Hub struct {
	Game map[string]*Game
	mu   sync.RWMutex
}

func newHub() *Hub {
	return &Hub{
		Game: make(map[string]*Game),
	}
}

// processMessage is the Hub's main entry point for all actions
func (h *Hub) processMessage(player *Player, msg *Message) {
	// Log the incoming message
	log.Printf("WS Message: Type=%s from Player=%p Payload=%s", msg.Type, player, string(msg.Payload))

	// 1. LOCK THE HUB
	// We lock the entire hub because we are accessing and modifying global game state.
	h.mu.Lock()
	defer h.mu.Unlock()

	switch msg.Type {

	// CASE 1: CREATE GAME

	case "create_game":
		gameID := "GAME" + generateRandomString(4)

		game := &Game{
			GameID:    gameID,
			Player1:   player,
			GameState: "waiting_for_join",
		}

		// Store game in the Hub
		h.Game[gameID] = game
		player.Game = game

		// Notify the creator
		reply := map[string]string{
			"type":    "game_created",
			"game_id": gameID,
		}
		sendJSON(player, reply)

	case "join_game":
		// 1. Parse Payload
		var payload struct {
			GameID string `json:"game_id"`
		}
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return // Invalid payload
		}

		// 2. Find Game
		game, ok := h.Game[payload.GameID]
		if !ok || game.GameState != "waiting_for_join" {
			sendError(player, "Game not found or full")
			return
		}

		// 3. Add Player
		game.Player2 = player
		player.Game = game
		game.GameState = "waiting_for_ready"

		// 4. Notify Both
		sendJSON(player, map[string]string{"type": "join_success"})
		sendJSON(game.Player1, map[string]string{"type": "player_joined"})

	// CASE 3: PLAYER READY (Secrets are set on client)

	case "player_ready":
		game := player.Game
		if game == nil {
			return
		}

		// Parse username
		var payload struct {
			Username string `json:"username"`
		}
		json.Unmarshal(msg.Payload, &payload)
		player.Username = payload.Username
		player.IsReady = true

		// Notify opponent
		var opponent *Player
		if player == game.Player1 {
			opponent = game.Player2
		} else {
			opponent = game.Player1
		}
		sendJSON(opponent, map[string]string{"type": "opponent_ready", "username": player.Username})

		// Check if both are ready
		if game.Player1.IsReady && game.Player2.IsReady {
			game.GameState = "waiting_for_guesses"
			// Broadcast start
			msg := map[string]string{
				"type":        "game_start",
				"p1_username": game.Player1.Username,
				"p2_username": game.Player2.Username,
			}
			sendJSON(game.Player1, msg)
			sendJSON(game.Player2, msg)
		}

	// CASE 4: MAKE GUESS (Simultaneous Turns)

	case "make_guess":
		game := player.Game
		// Strict State Check
		if game == nil || game.GameState != "waiting_for_guesses" {
			return
		}

		var payload struct {
			Guess string `json:"guess"`
		}
		json.Unmarshal(msg.Payload, &payload)

		// Store the guess
		if player == game.Player1 {
			game.p1Guess = payload.Guess
		} else {
			game.p2Guess = payload.Guess
		}

		// CHECK: Do we have BOTH guesses?
		if game.p1Guess != "" && game.p2Guess != "" {
			// Transition state
			game.GameState = "waiting_for_results"

			// SWAP GUESSES: Send P2's guess to P1, and P1's guess to P2
			// P1 needs to grade P2's guess
			sendJSON(game.Player1, map[string]string{
				"type":  "opponent_guess",
				"guess": game.p2Guess,
			})
			// P2 needs to grade P1's guess
			sendJSON(game.Player2, map[string]string{
				"type":  "opponent_guess",
				"guess": game.p1Guess,
			})
		}

	// CASE 5: SUBMIT RESULT (The Grading)

	case "submit_result":
		game := player.Game
		if game == nil || game.GameState != "waiting_for_results" {
			return
		}

		var result ResultPayload
		json.Unmarshal(msg.Payload, &result)

		// Logic: If Player 1 sends a result, they are grading Player 2's guess.
		// So we store it in p2Result.
		if player == game.Player1 {
			game.p2Result = &result
		} else {
			game.p1Result = &result
		}

		// CHECK: Do we have BOTH results?
		if game.p1Result != nil && game.p2Result != nil {

			// 1. Construct the broadcast message
			// We send one big message containing the outcome of the round
			roundUpdate := map[string]interface{}{
				"type": "round_result",
				"p1": map[string]interface{}{
					"guess": game.p1Guess,
					"bulls": game.p1Result.Bulls,
					"cows":  game.p1Result.Cows,
				},
				"p2": map[string]interface{}{
					"guess": game.p2Guess,
					"bulls": game.p2Result.Bulls,
					"cows":  game.p2Result.Cows,
				},
			}

			sendJSON(game.Player1, roundUpdate)
			sendJSON(game.Player2, roundUpdate)

			// 2. Check Win Condition
			// Your Rule: "You win when you get 4 Cows"
			p1Wins := game.p1Result.Cows == 4
			p2Wins := game.p2Result.Cows == 4

			if p1Wins || p2Wins {
				game.GameState = "game_over"
				winner := ""
				if p1Wins && p2Wins {
					winner = "draw"
				} else if p1Wins {
					winner = game.Player1.Username
				} else {
					winner = game.Player2.Username
				}

				gameOverMsg := map[string]string{
					"type":   "game_over",
					"winner": winner,
				}
				sendJSON(game.Player1, gameOverMsg)
				sendJSON(game.Player2, gameOverMsg)
			} else {
				// 3. No winner, reset for next round
				game.p1Guess = ""
				game.p2Guess = ""
				game.p1Result = nil
				game.p2Result = nil
				game.GameState = "waiting_for_guesses"
			}
		}
	}
}

// unregister handles player disconnects
func (h *Hub) unregister(player *Player) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// See if the player was in a game
	game := player.Game
	if game != nil {
		// Find the *other* player
		var opponent *Player
		if game.Player1 == player {
			opponent = game.Player2
		} else {
			opponent = game.Player1
		}

		// Tell the opponent their friend left
		if opponent != nil {
			msg, _ := json.Marshal(map[string]string{"type": "opponent_disconnected"})
			opponent.Send <- msg
		}

		delete(h.Game, game.GameID)
	}

	// Close the player's send channel
	close(player.Send)
}
