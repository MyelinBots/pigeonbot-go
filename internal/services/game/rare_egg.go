package game

import (
	"context"
	"fmt"
	rand "math/rand/v2"
)

const (
	rareEggAppearPercent  = 80 // 80% chance to appear
	rareEggSuccessPercent = 90 // 90% chance to successfully collect if it appears
	rareEggPointBoost     = 500
)

// Called after successful shot (and while activePigeon is locked by caller)
func (g *Game) TryRareEgg(ctx context.Context, shooterName string) (string, error) {
	// Must be mating and must have an active pigeon object
	if g.activePigeon == nil || g.activePigeon.activePigeon == nil || !g.activePigeon.IsMating {
		return "", nil
	}

	// Step 1: does it appear?
	if rand.IntN(100) >= rareEggAppearPercent {
		return "", nil
	}

	// Step 2: fail (no odds mentioned)
	if rand.IntN(100) >= rareEggSuccessPercent {
		return fmt.Sprintf(
			"✨ A mysterious rare egg appeared for %s ... but it cracked and vanished! 💥",
			shooterName,
		), nil
	}

	// Step 3: success → DB updates (eggs includes rare eggs) // go run ./cmd serve to test
	dbName := canonicalPlayerName(shooterName)

	totalEggs, err := g.playerRepository.AddEggs(ctx, g.network, g.channel, dbName, 1)
	if err != nil {
		return "", err
	}

	totalRare, err := g.playerRepository.AddRareEggs(ctx, g.network, g.channel, dbName, 1)
	if err != nil {
		return "", err
	}

	// Step 4: points boost (in-memory player)
	foundPlayer, err := g.FindPlayer(ctx, shooterName)
	if err != nil {
		return "", err
	}
	foundPlayer.Points += rareEggPointBoost

	return fmt.Sprintf(
		"🌟 WOW! %s collected a LEGENDARY rare egg! 🚀🚀🚀🚀🚀 +%d points with +1 egg 🥚 | Eggs: %d (Rare: %d) | Points: %d",
		shooterName,
		rareEggPointBoost,
		totalEggs,
		totalRare,
		foundPlayer.Points,
	), nil
}
