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
	if line == nil || len(line.Args) < 2 {
		return nil
	}

	command := line.Args[1]
	fmt.Println("Handling command:", command)

	handler, exists := c.commands[command]
	if !exists {
		return nil
	}

	// Put nick into context (optional, but good)
	ctx2 := context_manager.WithNick(ctx, line.Nick)

	// Your game handlers expect args[0] to ALWAYS be the caller nick.
	// So we inject it as the first arg, then append any extra args after the command.
	args := []string{line.Nick}
	if len(line.Args) > 2 {
		args = append(args, line.Args[2:]...)
	}

	return handler(ctx2, args...)
}

func (c *CommandControllerImpl) AddCommand(command string, handler func(ctx context.Context, args ...string) error) {
	c.commands[command] = handler
}
