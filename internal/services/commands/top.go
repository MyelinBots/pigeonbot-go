package commands

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/MyelinBots/pigeonbot-go/internal/services/game"
)

var reTop = regexp.MustCompile(`^!top(\d+)?$`)

// TopHandler handles !top5, !top10, and !top<N> (cap 10).
func TopHandler(g *game.Game) func(ctx context.Context, args ...string) error {
	return func(ctx context.Context, args ...string) error {
		if len(args) == 0 {
			return nil
		}
		msg := strings.TrimSpace(args[0])
		if !strings.HasPrefix(msg, "!top") {
			return nil
		}

		limit := parseTopLimit(msg)

		players, err := g.TopByPoints(ctx, limit)
		if err != nil {
			g.Irc().Privmsg(g.Channel(), fmt.Sprintf("[Top] error fetching leaderboard: %v", err))
			return err
		}
		if len(players) == 0 {
			g.Irc().Privmsg(g.Channel(), "No pigeon hunters yet! Try `!shoot` to start earning points 🐦💥")
			return nil
		}

		// Header: bold + color
		header := fmt.Sprintf("%s%s%s %s",
			ircBold,
			c("🏆 Top", 8), // Yellow
			ircReset,
			c(fmt.Sprintf("%d Pigeon Hunters", limit), 12), // Blue
		)
		g.Irc().Privmsg(g.Channel(), header)

		for i, p := range players {
			// Colors per field
			rank := medal(i)
			name := c(p.Name, 11)                           // Cyan
			points := c(fmt.Sprintf("%d pts", p.Points), 9) // Lime
			level := c(g.LevelFor(p.Points, p.Count), 6)    // Purple

			line := fmt.Sprintf("%2d. %s  %s — %s  (%s)", i+1, rank, name, points, level)
			g.Irc().Privmsg(g.Channel(), strings.TrimSpace(line))
		}
		return nil
	}
}

func parseTopLimit(cmd string) int {
	switch cmd {
	case "!top5":
		return 5
	case "!top10":
		return 10
	}
	m := reTop.FindStringSubmatch(cmd)
	if len(m) == 0 || m[1] == "" {
		return 5
	}
	n, err := strconv.Atoi(m[1])
	if err != nil || n < 1 {
		return 5
	}
	if n > 10 {
		return 10
	}
	return n
}

const (
	ircBold  = "\x02"
	ircColor = "\x03"
	ircReset = "\x0F"
)

func c(s string, fg int) string { // foreground only
	return fmt.Sprintf("%s%02d%s%s", ircColor, fg, s, ircReset)
}

func cb(s string, fg, bg int) string { // with background if you ever want it
	return fmt.Sprintf("%s%02d,%02d%s%s", ircColor, fg, bg, s, ircReset)
}

func medal(i int) string {
	switch i {
	case 0:
		return "🥇"
	case 1:
		return "🥈"
	case 2:
		return "🥉"
	default:
		return "•"
	}
}
