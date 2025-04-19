package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/MyelinBots/pigeonbot-go/internal/services/context_manager"
	irc "github.com/fluffle/goirc/client"
)

// InviteHandler allows users to invite purrito to their own channels
func InviteHandler(ircClient *irc.Conn,
	startNewGame func(ctx context.Context, channel string),
) func(ctx context.Context, args ...string) error {
	return func(ctx context.Context, args ...string) error {
		nick := context_manager.GetNickContext(ctx)

		commandArgs := strings.Split(args[0], " ")[1:]

		if len(commandArgs) < 1 || strings.ToLower(commandArgs[0]) != "pigeonbot" {
			return fmt.Errorf("usage: !invite pigeonbot")
		}

		// get line from args
		channel := commandArgs[1]

		ircClient.Join(channel)
		ircClient.Privmsg(channel, fmt.Sprintf("pigeonbot joins %s's channel. 🐾", nick))

		startNewGame(ctx, channel)

		fmt.Println("Invite command received from", nick)
		return nil
	}
}
