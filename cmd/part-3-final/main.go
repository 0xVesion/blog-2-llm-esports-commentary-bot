package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/net/websocket"
)

var ignoredEvents = map[string]bool{
	"player-purchased-item":          true,
	"player-completed-increaseLevel": true,
	"player-used-item":               true,
	"player-lost-item":               true,
	"game-set-npcRespawnClock":       true,
	"team-picked-character":          true,
}

const systemPrompt = "You're an esports match commentator. I'll feed you with events that happen during an esports round. For example a player-killed-player event. Please always respond with at most 120 characters e.g the length of a tweet. Only talk about the most recent event don't talk about the series at a whole. DO NOT describe the JSON structure. JUST put out a tweet about what happened in the event. Here is the most recent event:\n\n"

func main() {
	seriesId := 2
	apiKey := os.Getenv("GRID_API_KEY")
	url := fmt.Sprintf("wss://api.grid.gg/live-data-feed/series/%d?fromSequenceNumber=250&fromSessionSequenceNumber=&key=%s&useConfig=false", seriesId, apiKey)

	msgChannel, err := connectWs(url)
	if err != nil {
		slog.Error("Couldn't connect to GRID API", "error", err)
		return
	}

	for rawMsg := range msgChannel {
		var msg GridMessage
		err := json.Unmarshal([]byte(rawMsg), &msg)
		if err != nil {
			slog.Error("Couldn't unmarshal message", "error", err)
			return
		}

		for _, event := range msg.Events {
			if ignoredEvents[event.Type] {
				continue
			}

			response, err := promptLlm(systemPrompt + rawMsg)
			if err != nil {
				slog.Error("Couldn't prompt LLM", "error", err)
				return
			}

			slog.Info(event.Type + ": " + response)
		}
	}
}

type GridMessage struct {
	Events []GridEvent `json:"events"`
}

type GridEvent struct {
	Type string `json:"type"`
}

func promptLlm(prompt string) (string, error) {
	cmd := exec.Command("ollama", "run", "llama3.1", prompt)

	responseBytes, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("couldn't invoke ollama: %w", err)
	}

	return string(responseBytes), nil
}

func connectWs(url string) (<-chan string, error) {
	origin := strings.Replace(url, "ws", "http", 1)

	ws, err := websocket.Dial(url, "", origin)

	if err != nil {
		return nil, fmt.Errorf("couldn't connect to series ws: %w", err)
	}

	msgChannel := make(chan string)

	go func() {
		for {
			var msg string
			err := websocket.Message.Receive(ws, &msg)
			if err != nil {
				slog.Error("couldn't read ws message: %v", err)
				continue
			}

			msgChannel <- msg
		}
	}()

	return msgChannel, nil
}
