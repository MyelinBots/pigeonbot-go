package commands

import (
	"context"
	"fmt"
	"github.com/MyelinBots/pigeonbot-go/internal/services/context_manager"
	"github.com/MyelinBots/pigeonbot-go/internal/services/game"
	irc "github.com/fluffle/goirc/client"
)

type CommandController interface {
	HandleCommand(ctx context.Context, line *irc.Line) error
	AddCommand(command string, handler func(ctx context.Context, args ...string) error)
}

type CommandControllerImpl struct {
	game     *game.Game
	commands map[string]func(ctx context.Context, args ...string) error
}

func NewCommandController(gameinstance *game.Game) CommandController {
	return &CommandControllerImpl{
		game:     gameinstance,
		commands: make(map[string]func(ctx context.Context, args ...string) error),
	}
}

func (c *CommandControllerImpl) HandleCommand(ctx context.Context, line *irc.Line) error {
	command := line.Args[1]
	// args := line.Args[1:]
	fmt.Println("Handling command:", command)
	if handler, exists := c.commands[command]; exists {
		fmt.Println("Handling command:", command)
		ctx = context_manager.SetNickContext(ctx, line.Nick)
		return handler(ctx, line.Args[2:]...)
	} else {
		return nil
	}
}

func (c *CommandControllerImpl) AddCommand(command string, handler func(ctx context.Context, args ...string) error) {
	c.commands[command] = handler

	return
}
