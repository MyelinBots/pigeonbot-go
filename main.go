package main

import (
	"fmt"
	actions "pigeongo/actions"
	commands "pigeongo/commands"
	config "pigeongo/config"
	game "pigeongo/game"
	pigeon "pigeongo/pigeon"
	player "pigeongo/player"
	timer "pigeongo/timer"
)

func main() {
	actions.Doactions()
	commands.Docommands()
	config.LoadConfigOrPanic()
	game.NewGame()
	pigeon.NewPigeon()
	player.NewPlayer()
	timer.NewRepeatedTimer()
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
