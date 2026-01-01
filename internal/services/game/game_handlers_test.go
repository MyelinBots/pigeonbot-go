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
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func setupGame(t *testing.T) (*gomock.Controller, *mocks.MockPlayerRepository, *gameMocks.MockIRCClient, *game.Game) {
	ctrl := gomock.NewController(t)
	playerRepository := mocks.NewMockPlayerRepository(ctrl)

	if playerRepository.EggsByKey == nil {
		playerRepository.EggsByKey = make(map[string]int)
	}
	if playerRepository.RareEggsByKey == nil {
		playerRepository.RareEggsByKey = make(map[string]int)
	}

	ircClient := gameMocks.NewMockIRCClient(ctrl)

	return ctrl, playerRepository, ircClient, nil
}

func TestHandleLevel(t *testing.T) {
	t.Run("HandleLevel with multiple players", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		playerRepository := mocks.NewMockPlayerRepository(ctrl)
		playerRepository.EggsByKey = make(map[string]int)
		playerRepository.RareEggsByKey = make(map[string]int)

		playerRepository.EXPECT().
			GetAllPlayers(gomock.Any(), "network", "channel").
			Return([]*player.Player{
				{ID: "1", Name: "beginner", Points: 0, Count: 5, Channel: "channel", Network: "network"},
				{ID: "2", Name: "master", Points: 1000, Count: 500, Channel: "channel", Network: "network"},
			}, nil).
			Times(1)

		ircClient := gameMocks.NewMockIRCClient(ctrl)
		// Should output levels sorted by count
		ircClient.EXPECT().Privmsg("channel", gomock.Any()).AnyTimes()

		gameInstance := game.NewGame(config.GameConfig{Interval: 3}, ircClient, playerRepository, "network", "channel")

		commandController := commands.NewCommandController(gameInstance)
		commandController.AddCommand("!level", gameInstance.HandleLevel)

		ctx := context.Background()
		line := &irc.Line{
			Args: []string{"channel", "!level"},
			Nick: "test",
		}

		go gameInstance.Start(ctx)
		time.Sleep(100 * time.Millisecond)

		err := commandController.HandleCommand(ctx, line)
		assert.Nil(t, err)
	})
}

func TestHandleTop5(t *testing.T) {
	t.Run("HandleTop5 with players", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		playerRepository := mocks.NewMockPlayerRepository(ctrl)
		playerRepository.EggsByKey = make(map[string]int)
		playerRepository.RareEggsByKey = make(map[string]int)

		playerRepository.EXPECT().
			GetAllPlayers(gomock.Any(), "network", "channel").
			Return([]*player.Player{}, nil).
			Times(1)

		playerRepository.EXPECT().
			TopByPoints(gomock.Any(), "network", "channel", 5).
			Return([]*player.Player{
				{ID: "1", Name: "player1", Points: 1000, Count: 100, Eggs: 50, RareEggs: 5},
				{ID: "2", Name: "player2", Points: 500, Count: 50, Eggs: 25, RareEggs: 2},
				{ID: "3", Name: "player3", Points: 250, Count: 25, Eggs: 10, RareEggs: 1},
			}, nil).
			Times(1)

		ircClient := gameMocks.NewMockIRCClient(ctrl)
		ircClient.EXPECT().Privmsg("channel", gomock.Any()).AnyTimes()

		gameInstance := game.NewGame(config.GameConfig{Interval: 3}, ircClient, playerRepository, "network", "channel")

		commandController := commands.NewCommandController(gameInstance)
		commandController.AddCommand("!top5", gameInstance.HandleTop5)

		ctx := context.Background()
		line := &irc.Line{
			Args: []string{"channel", "!top5"},
			Nick: "test",
		}

		go gameInstance.Start(ctx)
		time.Sleep(100 * time.Millisecond)

		err := commandController.HandleCommand(ctx, line)
		assert.Nil(t, err)
	})
}

func TestHandleTop10(t *testing.T) {
	t.Run("HandleTop10 with players", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		playerRepository := mocks.NewMockPlayerRepository(ctrl)
		playerRepository.EggsByKey = make(map[string]int)
		playerRepository.RareEggsByKey = make(map[string]int)

		playerRepository.EXPECT().
			GetAllPlayers(gomock.Any(), "network", "channel").
			Return([]*player.Player{}, nil).
			Times(1)

		playerRepository.EXPECT().
			TopByPoints(gomock.Any(), "network", "channel", 10).
			Return([]*player.Player{
				{ID: "1", Name: "player1", Points: 1000, Count: 100, Eggs: 50, RareEggs: 5},
			}, nil).
			Times(1)

		ircClient := gameMocks.NewMockIRCClient(ctrl)
		ircClient.EXPECT().Privmsg("channel", gomock.Any()).AnyTimes()

		gameInstance := game.NewGame(config.GameConfig{Interval: 3}, ircClient, playerRepository, "network", "channel")

		commandController := commands.NewCommandController(gameInstance)
		commandController.AddCommand("!top10", gameInstance.HandleTop10)

		ctx := context.Background()
		line := &irc.Line{
			Args: []string{"channel", "!top10"},
			Nick: "test",
		}

		go gameInstance.Start(ctx)
		time.Sleep(100 * time.Millisecond)

		err := commandController.HandleCommand(ctx, line)
		assert.Nil(t, err)
	})
}

func TestHandleEggs(t *testing.T) {
	t.Run("HandleEggs shows egg count", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		playerRepository := mocks.NewMockPlayerRepository(ctrl)
		playerRepository.EggsByKey = make(map[string]int)
		playerRepository.RareEggsByKey = make(map[string]int)
		// Set up egg counts
		playerRepository.EggsByKey["network|channel|testuser"] = 10
		playerRepository.RareEggsByKey["network|channel|testuser"] = 2

		playerRepository.EXPECT().
			GetAllPlayers(gomock.Any(), "network", "channel").
			Return([]*player.Player{}, nil).
			Times(1)

		ircClient := gameMocks.NewMockIRCClient(ctrl)
		ircClient.EXPECT().Privmsg("channel", gomock.Any()).AnyTimes()

		gameInstance := game.NewGame(config.GameConfig{Interval: 3}, ircClient, playerRepository, "network", "channel")

		commandController := commands.NewCommandController(gameInstance)
		commandController.AddCommand("!eggs", gameInstance.HandleEggs)

		ctx := context.Background()
		line := &irc.Line{
			Args: []string{"channel", "!eggs"},
			Nick: "TestUser",
		}

		go gameInstance.Start(ctx)
		time.Sleep(100 * time.Millisecond)

		err := commandController.HandleCommand(ctx, line)
		assert.Nil(t, err)
	})
}

func TestNewGame(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	playerRepository := mocks.NewMockPlayerRepository(ctrl)
	ircClient := gameMocks.NewMockIRCClient(ctrl)

	gameInstance := game.NewGame(
		config.GameConfig{Interval: 60},
		ircClient,
		playerRepository,
		"testnetwork",
		"#testchannel",
	)

	assert.NotNil(t, gameInstance)
	assert.Equal(t, "#testchannel", gameInstance.Channel())
	assert.Equal(t, "testnetwork", gameInstance.Network())
	assert.Equal(t, ircClient, gameInstance.Irc())
}

func TestTopByPoints(t *testing.T) {
	t.Run("TopByPoints with valid limit", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		playerRepository := mocks.NewMockPlayerRepository(ctrl)
		playerRepository.EXPECT().
			TopByPoints(gomock.Any(), "network", "channel", 5).
			Return([]*player.Player{
				{ID: "1", Name: "top1", Points: 100},
			}, nil).
			Times(1)

		ircClient := gameMocks.NewMockIRCClient(ctrl)

		gameInstance := game.NewGame(config.GameConfig{Interval: 3}, ircClient, playerRepository, "network", "channel")

		result, err := gameInstance.TopByPoints(context.Background(), 5)
		assert.Nil(t, err)
		assert.Len(t, result, 1)
	})

	t.Run("TopByPoints with zero limit defaults to 5", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		playerRepository := mocks.NewMockPlayerRepository(ctrl)
		playerRepository.EXPECT().
			TopByPoints(gomock.Any(), "network", "channel", 5).
			Return([]*player.Player{}, nil).
			Times(1)

		ircClient := gameMocks.NewMockIRCClient(ctrl)

		gameInstance := game.NewGame(config.GameConfig{Interval: 3}, ircClient, playerRepository, "network", "channel")

		_, err := gameInstance.TopByPoints(context.Background(), 0)
		assert.Nil(t, err)
	})

	t.Run("TopByPoints with negative limit defaults to 5", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		playerRepository := mocks.NewMockPlayerRepository(ctrl)
		playerRepository.EXPECT().
			TopByPoints(gomock.Any(), "network", "channel", 5).
			Return([]*player.Player{}, nil).
			Times(1)

		ircClient := gameMocks.NewMockIRCClient(ctrl)

		gameInstance := game.NewGame(config.GameConfig{Interval: 3}, ircClient, playerRepository, "network", "channel")

		_, err := gameInstance.TopByPoints(context.Background(), -1)
		assert.Nil(t, err)
	})

	t.Run("TopByPoints with limit over 50 caps at 50", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		playerRepository := mocks.NewMockPlayerRepository(ctrl)
		playerRepository.EXPECT().
			TopByPoints(gomock.Any(), "network", "channel", 50).
			Return([]*player.Player{}, nil).
			Times(1)

		ircClient := gameMocks.NewMockIRCClient(ctrl)

		gameInstance := game.NewGame(config.GameConfig{Interval: 3}, ircClient, playerRepository, "network", "channel")

		_, err := gameInstance.TopByPoints(context.Background(), 100)
		assert.Nil(t, err)
	})
}

func TestLevelFor(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	playerRepository := mocks.NewMockPlayerRepository(ctrl)
	ircClient := gameMocks.NewMockIRCClient(ctrl)

	gameInstance := game.NewGame(config.GameConfig{Interval: 3}, ircClient, playerRepository, "network", "channel")

	tests := []struct {
		name     string
		points   int
		count    int
		expected string
	}{
		{"beginner", 0, 0, "Beginner 🐣"},
		{"beginner at 9", 100, 9, "Beginner 🐣"},
		{"initiate at 10", 0, 10, "Initiate 🐦"},
		{"initiate at 100", 500, 100, "Initiate 🐦"},
		{"adept at 101", 0, 101, "Adept 🦅"},
		{"expert at 200", 0, 200, "Expert 🕊️"},
		{"master at 500", 0, 500, "Master 🦜"},
		{"grandmaster at 800", 0, 800, "Grandmaster 🐔"},
		{"legendary at 1000", 0, 1000, "Legendary Phoenix 🐉🔥"},
		{"mythic at 3000", 0, 3000, "Mythic Dragon 🐲✨"},
		{"cosmic at 5000", 0, 5000, "Cosmic Falcon 🌌🦅"},
		{"lord at 10000", 0, 10000, "Lord of Pigeons 👑🐦"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gameInstance.LevelFor(tt.points, tt.count)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFindPlayer(t *testing.T) {
	t.Run("FindPlayer creates new player if not exists", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		playerRepository := mocks.NewMockPlayerRepository(ctrl)
		playerRepository.EggsByKey = make(map[string]int)
		playerRepository.RareEggsByKey = make(map[string]int)

		playerRepository.EXPECT().
			GetAllPlayers(gomock.Any(), "network", "channel").
			Return([]*player.Player{}, nil).
			Times(1)

		playerRepository.EXPECT().
			UpsertPlayer(gomock.Any(), gomock.Any()).
			Return(nil).
			Times(1)

		ircClient := gameMocks.NewMockIRCClient(ctrl)
		ircClient.EXPECT().Privmsg("channel", gomock.Any()).AnyTimes()

		gameInstance := game.NewGame(config.GameConfig{Interval: 3}, ircClient, playerRepository, "network", "channel")

		ctx := context.Background()
		go gameInstance.Start(ctx)
		time.Sleep(100 * time.Millisecond)

		foundPlayer, err := gameInstance.FindPlayer(ctx, "newplayer")
		assert.Nil(t, err)
		assert.NotNil(t, foundPlayer)
		assert.Equal(t, "newplayer", foundPlayer.Name)
	})

	t.Run("FindPlayer returns existing player", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		playerRepository := mocks.NewMockPlayerRepository(ctrl)
		playerRepository.EggsByKey = make(map[string]int)
		playerRepository.RareEggsByKey = make(map[string]int)

		playerRepository.EXPECT().
			GetAllPlayers(gomock.Any(), "network", "channel").
			Return([]*player.Player{
				{ID: "1", Name: "existingplayer", Points: 100, Count: 10},
			}, nil).
			Times(1)

		ircClient := gameMocks.NewMockIRCClient(ctrl)
		ircClient.EXPECT().Privmsg("channel", gomock.Any()).AnyTimes()

		gameInstance := game.NewGame(config.GameConfig{Interval: 3}, ircClient, playerRepository, "network", "channel")

		ctx := context.Background()
		go gameInstance.Start(ctx)
		time.Sleep(100 * time.Millisecond)

		foundPlayer, err := gameInstance.FindPlayer(ctx, "existingplayer")
		assert.Nil(t, err)
		assert.NotNil(t, foundPlayer)
		assert.Equal(t, "existingplayer", foundPlayer.Name)
		assert.Equal(t, 100, foundPlayer.Points)
		assert.Equal(t, 10, foundPlayer.Count)
	})
}

func TestSavePlayers(t *testing.T) {
	t.Run("SavePlayers persists all players", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		playerRepository := mocks.NewMockPlayerRepository(ctrl)
		playerRepository.EggsByKey = make(map[string]int)
		playerRepository.RareEggsByKey = make(map[string]int)

		playerRepository.EXPECT().
			GetAllPlayers(gomock.Any(), "network", "channel").
			Return([]*player.Player{
				{ID: "1", Name: "player1", Points: 100, Count: 10},
				{ID: "2", Name: "player2", Points: 200, Count: 20},
			}, nil).
			Times(1)

		// Expect UpsertPlayer to be called for each player
		playerRepository.EXPECT().
			UpsertPlayer(gomock.Any(), gomock.Any()).
			Return(nil).
			Times(2)

		ircClient := gameMocks.NewMockIRCClient(ctrl)
		ircClient.EXPECT().Privmsg("channel", gomock.Any()).AnyTimes()

		gameInstance := game.NewGame(config.GameConfig{Interval: 3}, ircClient, playerRepository, "network", "channel")

		ctx := context.Background()
		go gameInstance.Start(ctx)
		time.Sleep(100 * time.Millisecond)

		err := gameInstance.SavePlayers(ctx)
		assert.Nil(t, err)
	})
}

func TestGameStart_ContextCancellation(t *testing.T) {
	t.Run("Start stops when context is cancelled", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		playerRepository := mocks.NewMockPlayerRepository(ctrl)
		playerRepository.EggsByKey = make(map[string]int)
		playerRepository.RareEggsByKey = make(map[string]int)

		playerRepository.EXPECT().
			GetAllPlayers(gomock.Any(), "network", "channel").
			Return([]*player.Player{}, nil).
			Times(1)

		ircClient := gameMocks.NewMockIRCClient(ctrl)
		ircClient.EXPECT().Privmsg("channel", gomock.Any()).AnyTimes()

		gameInstance := game.NewGame(config.GameConfig{Interval: 1}, ircClient, playerRepository, "network", "channel")

		ctx, cancel := context.WithCancel(context.Background())

		done := make(chan struct{})
		go func() {
			gameInstance.Start(ctx)
			close(done)
		}()

		// Let it run briefly
		time.Sleep(100 * time.Millisecond)

		// Cancel the context
		cancel()

		// Wait for Start to return
		select {
		case <-done:
			// Success
		case <-time.After(2 * time.Second):
			t.Fatal("Start did not stop after context cancellation")
		}
	})
}

func TestHandlePoints_EmptyPlayers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	playerRepository := mocks.NewMockPlayerRepository(ctrl)
	playerRepository.EggsByKey = make(map[string]int)
	playerRepository.RareEggsByKey = make(map[string]int)

	playerRepository.EXPECT().
		GetAllPlayers(gomock.Any(), "network", "channel").
		Return([]*player.Player{}, nil).
		Times(1)

	ircClient := gameMocks.NewMockIRCClient(ctrl)
	ircClient.EXPECT().Privmsg("channel", gomock.Any()).AnyTimes()

	gameInstance := game.NewGame(config.GameConfig{Interval: 3}, ircClient, playerRepository, "network", "channel")

	ctx := context.Background()
	go gameInstance.Start(ctx)
	time.Sleep(100 * time.Millisecond)

	err := gameInstance.HandlePoints(ctx)
	assert.Nil(t, err)
}

func TestHandleCount_EmptyPlayers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	playerRepository := mocks.NewMockPlayerRepository(ctrl)
	playerRepository.EggsByKey = make(map[string]int)
	playerRepository.RareEggsByKey = make(map[string]int)

	playerRepository.EXPECT().
		GetAllPlayers(gomock.Any(), "network", "channel").
		Return([]*player.Player{}, nil).
		Times(1)

	ircClient := gameMocks.NewMockIRCClient(ctrl)
	ircClient.EXPECT().Privmsg("channel", gomock.Any()).AnyTimes()

	gameInstance := game.NewGame(config.GameConfig{Interval: 3}, ircClient, playerRepository, "network", "channel")

	ctx := context.Background()
	go gameInstance.Start(ctx)
	time.Sleep(100 * time.Millisecond)

	err := gameInstance.HandleCount(ctx)
	assert.Nil(t, err)
}

func TestHandleLevel_EmptyPlayers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	playerRepository := mocks.NewMockPlayerRepository(ctrl)
	playerRepository.EggsByKey = make(map[string]int)
	playerRepository.RareEggsByKey = make(map[string]int)

	playerRepository.EXPECT().
		GetAllPlayers(gomock.Any(), "network", "channel").
		Return([]*player.Player{}, nil).
		Times(1)

	ircClient := gameMocks.NewMockIRCClient(ctrl)
	ircClient.EXPECT().Privmsg("channel", gomock.Any()).AnyTimes()

	gameInstance := game.NewGame(config.GameConfig{Interval: 3}, ircClient, playerRepository, "network", "channel")

	ctx := context.Background()
	go gameInstance.Start(ctx)
	time.Sleep(100 * time.Millisecond)

	err := gameInstance.HandleLevel(ctx)
	assert.Nil(t, err)
}
