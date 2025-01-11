package main

import (
	"fmt"

	config "github.com/MyelinBots/pigeonbot-go/config"
	"github.com/MyelinBots/pigeonbot-go/internal/services/actions"
	"github.com/MyelinBots/pigeonbot-go/internal/services/commands"
	"github.com/MyelinBots/pigeonbot-go/internal/services/game"
	"github.com/MyelinBots/pigeonbot-go/internal/services/pigeon"
	"github.com/MyelinBots/pigeonbot-go/internal/services/player"
	"github.com/MyelinBots/pigeonbot-go/internal/services/timer"
)

func main() {

	actions.Doactions()
	commands.Docommands()
	config.LoadConfigOrPanic()
	game.NewGame()
	pigeon.NewPigeon()
	player.NewPlayer()
	timer.NewRepeatedTimer()

	<-quit
}

// Mock implementation of IRC
type MockIRC struct {
	channel string
}

func (m *MockIRC) Privmsg(channel, message string) {
	fmt.Printf("[%s] %s\n", channel, message)
}

func (m *MockIRC) Config() config.Config {
	return config.LoadConfigOrPanic()
}
