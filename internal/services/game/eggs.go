package game

import (
	"context"
	"fmt"
	rand "math/rand/v2"
	"strings"
)

// Base eggs by pigeon type
// cartel member = 1, white = 2, boss = 5
func baseEggsByType(pigeonType string) int {
	switch strings.ToLower(strings.TrimSpace(pigeonType)) {
	case "cartel member":
		return 1
	case "white":
		return 2
	case "boss":
		return 5
	default:
		return 0
	}
}

// Cracking rules (exactly as requested)
// cartel member: 1 -> 0
// white:         2 -> 0 or 1
// boss:          5 -> 0..4
func eggsAfterCrack(pigeonType string) (final int, cracked int) {
	base := baseEggsByType(pigeonType)
	if base == 0 {
		return 0, 0
	}

	switch strings.ToLower(strings.TrimSpace(pigeonType)) {
	case "cartel member":
		return 0, 1

	case "white":
		final = rand.IntN(2) // 0..1
		return final, base - final

	case "boss":
		final = rand.IntN(5) // 0..4
		return final, base - final

	default:
		return base, 0
	}
}

// HandleMatingEggs is called AFTER a successful shot
// g.activePigeon must already be locked by the caller
func (g *Game) HandleMatingEggs(ctx context.Context, shooterName string) (string, error) {

	// Must have an active mating pigeon
	if g.activePigeon == nil ||
		g.activePigeon.activePigeon == nil ||
		!g.activePigeon.IsMating {
		return "", nil
	}

	pType := g.activePigeon.activePigeon.Type
	if baseEggsByType(pType) == 0 {
		return "", nil
	}

	final, cracked := eggsAfterCrack(pType)

	// ✅ canonical name for DB read/write (works for ALL users)
	dbName := canonicalPlayerName(shooterName)

	// All eggs cracked
	if final <= 0 {
		total, err := g.playerRepository.GetEggs(ctx, g.network, g.channel, dbName)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf(
			"🥚💥 Oh no... :( the eggs cracked during the chaos! - no eggs collected ... You now have %s egg(s) in total.",
			fmtNum(total),
		), nil
	}

	// Add eggs
	total, err := g.playerRepository.AddEggs(
		ctx,
		g.network,
		g.channel,
		dbName, // ✅ store canonical
		final,
	)
	if err != nil {
		return "", err
	}

	if cracked > 0 {
		return fmt.Sprintf(
			"🥚🐣 Yay!! %s has collected %s egg(s) ... Unfortunately, %s cracked!... You now have %s egg(s) in total.",
			shooterName, // ✅ display original nick
			fmtNum(final),
			fmtNum(cracked),
			fmtNum(total),
		), nil
	}

	rareEggs, err := g.playerRepository.GetRareEggs(ctx, g.network, g.channel, dbName)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(
		"%s collected %s egg(s)! Total eggs: %s (Rare egg(s): %s 🌟🥚)",
		shooterName,
		fmtNum(final),
		fmtNum(total),
		fmtNum(rareEggs),
	), nil
}

func (g *Game) EggsAfterShot(ctx context.Context, shooterName string) (string, error) {
	return g.HandleMatingEggs(ctx, shooterName)
}
