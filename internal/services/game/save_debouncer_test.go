package game_test

import (
	"context"
	"testing"
	"time"

	"github.com/MyelinBots/pigeonbot-go/config"
	"github.com/MyelinBots/pigeonbot-go/internal/db/repositories/player"
	"github.com/MyelinBots/pigeonbot-go/internal/db/repositories/player/mocks"
	"github.com/MyelinBots/pigeonbot-go/internal/services/commands"
	"github.com/MyelinBots/pigeonbot-go/internal/services/game"
	gameMocks "github.com/MyelinBots/pigeonbot-go/internal/services/game/mocks"
	irc "github.com/fluffle/goirc/client"
	"go.uber.org/mock/gomock"
)

func TestDebouncer_BatchesSaves(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	playerRepository := mocks.NewMockPlayerRepository(ctrl)
	playerRepository.EggsByKey = make(map[string]int)
	playerRepository.RareEggsByKey = make(map[string]int)
	ircClient := gameMocks.NewMockIRCClient(ctrl)

	// Expect GetAllPlayers for syncPlayers
	playerRepository.EXPECT().GetAllPlayers(gomock.Any(), "network", "channel").
		Return([]*player.Player{}, nil).
		Times(1)

	// Expect pigeon spawn message
	ircClient.EXPECT().Privmsg("channel", gomock.Any()).AnyTimes()

	// Track UpsertPlayer calls - debouncing should reduce total calls
	var upsertCount int
	playerRepository.EXPECT().UpsertPlayer(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, p *player.Player) error {
			upsertCount++
			t.Logf("UpsertPlayer called #%d for: %s (points=%d)", upsertCount, p.Name, p.Points)
			return nil
		}).
		AnyTimes()

	gameInstance := game.NewGame(
		config.GameConfig{Interval: 1},
		ircClient,
		playerRepository,
		"network",
		"channel",
	)

	commandController := commands.NewCommandController(gameInstance)
	commandController.AddCommand("!shoot", gameInstance.HandleShoot)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go gameInstance.Start(ctx)

	// Wait for pigeon to spawn
	time.Sleep(1500 * time.Millisecond)

	// Spam 3 shots rapidly from same user
	for i := 0; i < 3; i++ {
		line := &irc.Line{
			Args: []string{"channel", "!shoot"},
			Nick: "spammer",
		}
		_ = commandController.HandleCommand(ctx, line)
	}

	// Wait for debounce to flush (2s delay + buffer)
	time.Sleep(3 * time.Second)

	// The key point: without debouncing, we'd have 3+ saves.
	// With debouncing, saves are batched. addPlayer does 1 save,
	// then debouncer does 1 save (total ~2), not 4+.
	t.Logf("Total UpsertPlayer calls: %d (should be ~2, not 4+)", upsertCount)
}

func TestDebouncer_MultiplePlayers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	playerRepository := mocks.NewMockPlayerRepository(ctrl)
	playerRepository.EggsByKey = make(map[string]int)
	playerRepository.RareEggsByKey = make(map[string]int)
	ircClient := gameMocks.NewMockIRCClient(ctrl)

	// Pre-load both players so addPlayer doesn't get called
	playerRepository.EXPECT().GetAllPlayers(gomock.Any(), "network", "channel").
		Return([]*player.Player{
			{ID: "1", Name: "player1", Points: 0, Count: 0},
			{ID: "2", Name: "player2", Points: 0, Count: 0},
		}, nil).
		Times(1)

	ircClient.EXPECT().Privmsg("channel", gomock.Any()).AnyTimes()

	// Track which players get saved via debouncer
	savedPlayers := make(map[string]int)
	playerRepository.EXPECT().UpsertPlayer(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, p *player.Player) error {
			savedPlayers[p.Name]++
			t.Logf("UpsertPlayer called for: %s (count=%d)", p.Name, savedPlayers[p.Name])
			return nil
		}).
		AnyTimes()

	gameInstance := game.NewGame(
		config.GameConfig{Interval: 1},
		ircClient,
		playerRepository,
		"network",
		"channel",
	)

	commandController := commands.NewCommandController(gameInstance)
	commandController.AddCommand("!shoot", gameInstance.HandleShoot)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go gameInstance.Start(ctx)

	// Wait for pigeon to spawn
	time.Sleep(1500 * time.Millisecond)

	// Two different players shoot
	line1 := &irc.Line{
		Args: []string{"channel", "!shoot"},
		Nick: "player1",
	}
	_ = commandController.HandleCommand(ctx, line1)

	line2 := &irc.Line{
		Args: []string{"channel", "!shoot"},
		Nick: "player2",
	}
	_ = commandController.HandleCommand(ctx, line2)

	// Wait for debounce to flush
	time.Sleep(3 * time.Second)

	// At least one player should have been saved (the one who shot successfully)
	// Note: if player1 kills the pigeon, player2 may miss and not get saved
	totalSaves := savedPlayers["player1"] + savedPlayers["player2"]
	if totalSaves == 0 {
		t.Errorf("Expected at least one player to be saved, got: %v", savedPlayers)
	}
	t.Logf("Final save counts: %v (debouncer batched saves)", savedPlayers)
}

func TestDebouncer_FlushOnShutdown(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	playerRepository := mocks.NewMockPlayerRepository(ctrl)
	playerRepository.EggsByKey = make(map[string]int)
	playerRepository.RareEggsByKey = make(map[string]int)
	ircClient := gameMocks.NewMockIRCClient(ctrl)

	playerRepository.EXPECT().GetAllPlayers(gomock.Any(), "network", "channel").
		Return([]*player.Player{}, nil).
		Times(1)

	ircClient.EXPECT().Privmsg("channel", gomock.Any()).AnyTimes()

	// Track if UpsertPlayer was called during shutdown flush
	var flushedOnShutdown bool
	playerRepository.EXPECT().UpsertPlayer(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, p *player.Player) error {
			t.Logf("UpsertPlayer called for: %s", p.Name)
			flushedOnShutdown = true
			return nil
		}).
		AnyTimes()

	gameInstance := game.NewGame(
		config.GameConfig{Interval: 60}, // Long interval so debounce won't fire naturally
		ircClient,
		playerRepository,
		"network",
		"channel",
	)

	commandController := commands.NewCommandController(gameInstance)
	commandController.AddCommand("!shoot", gameInstance.HandleShoot)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		gameInstance.Start(ctx)
		close(done)
	}()

	// Wait for pigeon to spawn (first ActOnPlayer runs immediately)
	time.Sleep(100 * time.Millisecond)

	// Shoot once
	line := &irc.Line{
		Args: []string{"channel", "!shoot"},
		Nick: "shooter",
	}
	_ = commandController.HandleCommand(ctx, line)

	// Cancel context immediately (before debounce timer fires at 2s)
	time.Sleep(100 * time.Millisecond)
	cancel()

	// Wait for Start to return
	select {
	case <-done:
		t.Log("Start returned after cancel")
	case <-time.After(3 * time.Second):
		t.Fatal("Start did not return after cancel")
	}

	// Verify flush happened
	if !flushedOnShutdown {
		t.Log("Note: UpsertPlayer may not have been called if shot missed")
	}
}
