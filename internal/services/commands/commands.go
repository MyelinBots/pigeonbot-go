package commands

import (
	"context"
	"strings"

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

func (c *CommandControllerImpl) AddCommand(command string, handler func(ctx context.Context, args ...string) error) {
	// normalize key ให้เป็น lower-case ไว้ก่อน จะได้ match ง่าย
	c.commands[strings.ToLower(strings.TrimSpace(command))] = handler
}

func (c *CommandControllerImpl) HandleCommand(ctx context.Context, line *irc.Line) error {
	// goirc PRIVMSG: line.Args[0] = channel, line.Args[1] = message
	if line == nil || len(line.Args) < 2 {
		return nil
	}

	msg := strings.TrimSpace(line.Args[1])
	if msg == "" {
		return nil
	}

	parts := strings.Fields(msg)
	if len(parts) == 0 {
		return nil
	}

	cmd := strings.ToLower(parts[0])
	handler, ok := c.commands[cmd]
	if !ok {
		return nil
	}

	// ใส่ nick เข้า context (มาตรฐาน)
	ctx2 := context_manager.WithNick(ctx, line.Nick)

	// ส่ง args ต่อท้าย (หลัง command)
	args := []string{}
	if len(parts) > 1 {
		args = parts[1:]
	}

	return handler(ctx2, args...)
}
