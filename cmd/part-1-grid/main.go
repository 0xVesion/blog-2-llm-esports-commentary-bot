package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"golang.org/x/net/websocket"
)

func main() {
	seriesId := 2
	apiKey := os.Getenv("GRID_API_KEY")
	url := fmt.Sprintf("wss://api.grid.gg/live-data-feed/series/%d?fromSequenceNumber=0&fromSessionSequenceNumber=&key=%s&useConfig=false", seriesId, apiKey)

	msgChannel, err := connectWs(url)
	if err != nil {
		slog.Error("couldn't connect to ws", "error", err)
		return
	}

	for rawMsg := range msgChannel {
		var msg GridMessage
		err := json.Unmarshal([]byte(rawMsg), &msg)
		if err != nil {
			slog.Error("couldn't unmarshal message", "error", err)
			return
		}

		for _, event := range msg.Events {
			slog.Info("New event", "type", event.Type)
		}
	}
}

type GridMessage struct {
	Events []GridEvent `json:"events"`
}

type GridEvent struct {
	Type string `json:"type"`
}

func connectWs(url string) (<-chan string, error) {
	ws, err := websocket.Dial(url, "", url)

	if err != nil {
		return nil, fmt.Errorf("couldn't connect to series ws: %w", err)
	}

	msgChannel := make(chan string)

	go func() {
		for {
			var msg string
			err := websocket.Message.Receive(ws, &msg)
			if err != nil {
				slog.Error("couldn't read ws message", "error", err)
				continue
			}

			msgChannel <- msg
		}
	}()

	return msgChannel, nil
}
