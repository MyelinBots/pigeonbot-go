package game

import (
	"context"
	"fmt"
	"sync"
	"time"

	player2 "github.com/MyelinBots/pigeonbot-go/internal/db/repositories/player"
)

const (
	debounceDelay = 2 * time.Second // wait 2s after last change before saving
)

type saveDebouncer struct {
	mu           sync.Mutex
	dirtyPlayers map[string]bool
	timer        *time.Timer
	game         *Game
	ctx          context.Context
}

func newSaveDebouncer(g *Game, ctx context.Context) *saveDebouncer {
	return &saveDebouncer{
		dirtyPlayers: make(map[string]bool),
		game:         g,
		ctx:          ctx,
	}
}

// MarkDirty marks a player as needing to be saved and starts/resets the debounce timer
func (s *saveDebouncer) MarkDirty(playerName string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.dirtyPlayers[playerName] = true

	// Reset timer on each mark
	if s.timer != nil {
		s.timer.Stop()
	}
	s.timer = time.AfterFunc(debounceDelay, s.flush)
}

// flush saves all dirty players to the database
func (s *saveDebouncer) flush() {
	s.mu.Lock()
	if len(s.dirtyPlayers) == 0 {
		s.mu.Unlock()
		return
	}

	// Copy dirty set and clear
	toSave := make([]string, 0, len(s.dirtyPlayers))
	for name := range s.dirtyPlayers {
		toSave = append(toSave, name)
	}
	s.dirtyPlayers = make(map[string]bool)
	s.mu.Unlock()

	// Save only dirty players
	s.game.savePlayers(s.ctx, toSave)
}

// FlushNow forces an immediate save (useful for shutdown)
func (s *saveDebouncer) FlushNow() {
	s.mu.Lock()
	if s.timer != nil {
		s.timer.Stop()
		s.timer = nil
	}
	s.mu.Unlock()

	s.flush()
}

// savePlayers saves only the specified players (called by debouncer)
func (g *Game) savePlayers(ctx context.Context, playerNames []string) {
	g.players.Lock()
	defer g.players.Unlock()

	nameSet := make(map[string]bool, len(playerNames))
	for _, n := range playerNames {
		nameSet[n] = true
	}

	for _, p := range g.players.players {
		if !nameSet[p.Name] {
			continue
		}

		playerEntity := player2.Player{
			Count:   p.Count,
			Points:  p.Points,
			Name:    p.Name,
			Channel: g.channel,
			Network: g.network,
		}
		if err := g.playerRepository.UpsertPlayer(ctx, &playerEntity); err != nil {
			fmt.Printf("Error saving player %s: %v\n", p.Name, err)
		}
	}

	fmt.Printf("[debouncer] saved %d players\n", len(playerNames))
}
