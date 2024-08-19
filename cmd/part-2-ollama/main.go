package main

import (
	"fmt"
	"log/slog"
	"os/exec"
)

const systemPrompt = "You're an esports match commentator. I'll feed you with events that happen during an esports game. Please always respond with at most 120 characters e.g the length of a tweet. Only talk about the most recent event don't talk about the series as a whole. Here is the most recent event:\n\n"

func main() {
	response, err := promptLlm(systemPrompt + `{"type": "player-killed-player", "actor": "player-a", "target": "player-b"}`)
	if err != nil {
		slog.Error("Couldn't prompt LLM", "error", err)
		return
	}

	slog.Info(response)
}

func promptLlm(prompt string) (string, error) {
	cmd := exec.Command("ollama", "run", "llama3.1", prompt)

	responseBytes, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error whilst retrieving ollama output: %w", err)
	}

	return string(responseBytes), nil
}
