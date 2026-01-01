package game

import (
	"context"
	"fmt"
)

func (g *Game) HandleEggs(ctx context.Context, args ...string) error {
	// ✅ prefer ctx nick (works even when args=[])
	nickAny := ctx.Value("nick")
	nick, _ := nickAny.(string)

	// fallback if context not set
	if nick == "" && len(args) > 0 {
		nick = args[0]
	}

	if nick == "" {
		g.ircClient.Privmsg(g.channel, "DEBUG eggs: no nick in ctx and no args")
		return nil
	}

	dbName := canonicalPlayerName(nick)

	totalEggs, err := g.playerRepository.GetEggs(ctx, g.network, g.channel, dbName)
	if err != nil {
		g.ircClient.Privmsg(g.channel, fmt.Sprintf("DEBUG eggs: GetEggs err=%v", err))
		return err
	}

	totalRare, err := g.playerRepository.GetRareEggs(ctx, g.network, g.channel, dbName)
	if err != nil {
		g.ircClient.Privmsg(g.channel, fmt.Sprintf("DEBUG eggs: GetRareEggs err=%v", err))
		return err
	}

	g.ircClient.Privmsg(
		g.channel,
		fmt.Sprintf("🥚 %s has %d egg(s) total — including %d rare egg(s) 🌟🥚", nick, totalEggs, totalRare),
	)
	return nil
}
