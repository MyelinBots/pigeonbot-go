package game_test

import (
	"context"
	"testing"
	"time"

	"github.com/MyelinBots/pigeonbot-go/config"
	"github.com/MyelinBots/pigeonbot-go/internal/db/repositories/player"
	"github.com/MyelinBots/pigeonbot-go/internal/db/repositories/player/mocks"
	"github.com/MyelinBots/pigeonbot-go/internal/services/game"
	gameMocks "github.com/MyelinBots/pigeonbot-go/internal/services/game/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestHandleShoot_NoPigeon(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	playerRepository := mocks.NewMockPlayerRepository(ctrl)
	playerRepository.EggsByKey = make(map[string]int)
	playerRepository.RareEggsByKey = make(map[string]int)

	playerRepository.EXPECT().
		GetAllPlayers(gomock.Any(), "network", "channel").
		Return([]*player.Player{
			{ID: "1", Name: "testuser", Points: 0, Count: 0},
		}, nil).
		Times(1)

	playerRepository.EXPECT().
		UpsertPlayer(gomock.Any(), gomock.Any()).
		Return(nil).
		AnyTimes()

	ircClient := gameMocks.NewMockIRCClient(ctrl)
	// Expect a message about no pigeon to shoot (among other messages)
	ircClient.EXPECT().Privmsg("channel", gomock.Any()).AnyTimes()

	gameInstance := game.NewGame(config.GameConfig{Interval: 300}, ircClient, playerRepository, "network", "channel")

	// Set up context with nick
	ctx := context.WithValue(context.Background(), "nick", "testuser")

	// Start the game (but with long interval so no pigeon spawns)
	go gameInstance.Start(ctx)
	time.Sleep(50 * time.Millisecond)

	// Try to shoot when there's no pigeon
	err := gameInstance.HandleShoot(ctx)
	assert.Nil(t, err)
}

func TestHandleEggs_NoNickInContext(t *testing.T) {
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

	gameInstance := game.NewGame(config.GameConfig{Interval: 300}, ircClient, playerRepository, "network", "channel")

	// Empty context (no nick)
	ctx := context.Background()

	go gameInstance.Start(ctx)
	time.Sleep(50 * time.Millisecond)

	// Call HandleEggs with args instead of context
	err := gameInstance.HandleEggs(ctx, "testplayer")
	assert.Nil(t, err)
}

func TestHandleEggs_WithNickInContext(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	playerRepository := mocks.NewMockPlayerRepository(ctrl)
	playerRepository.EggsByKey = make(map[string]int)
	playerRepository.RareEggsByKey = make(map[string]int)
	playerRepository.EggsByKey["network|channel|testuser"] = 5
	playerRepository.RareEggsByKey["network|channel|testuser"] = 1

	playerRepository.EXPECT().
		GetAllPlayers(gomock.Any(), "network", "channel").
		Return([]*player.Player{}, nil).
		Times(1)

	ircClient := gameMocks.NewMockIRCClient(ctrl)
	ircClient.EXPECT().Privmsg("channel", gomock.Any()).AnyTimes()

	gameInstance := game.NewGame(config.GameConfig{Interval: 300}, ircClient, playerRepository, "network", "channel")

	// Context with nick
	ctx := context.WithValue(context.Background(), "nick", "TestUser")

	go gameInstance.Start(ctx)
	time.Sleep(50 * time.Millisecond)

	err := gameInstance.HandleEggs(ctx)
	assert.Nil(t, err)
}

func TestHandleTopN_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	playerRepository := mocks.NewMockPlayerRepository(ctrl)
	playerRepository.EggsByKey = make(map[string]int)
	playerRepository.RareEggsByKey = make(map[string]int)

	playerRepository.EXPECT().
		GetAllPlayers(gomock.Any(), "network", "channel").
		Return([]*player.Player{}, nil).
		Times(1)

	expectedErr := assert.AnError
	playerRepository.EXPECT().
		TopByPoints(gomock.Any(), "network", "channel", 5).
		Return(nil, expectedErr).
		Times(1)

	ircClient := gameMocks.NewMockIRCClient(ctrl)
	ircClient.EXPECT().Privmsg("channel", gomock.Any()).AnyTimes()

	gameInstance := game.NewGame(config.GameConfig{Interval: 300}, ircClient, playerRepository, "network", "channel")

	ctx := context.Background()
	go gameInstance.Start(ctx)
	time.Sleep(50 * time.Millisecond)

	err := gameInstance.HandleTop5(ctx)
	assert.Equal(t, expectedErr, err)
}

func TestSavePlayers_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	playerRepository := mocks.NewMockPlayerRepository(ctrl)
	playerRepository.EggsByKey = make(map[string]int)
	playerRepository.RareEggsByKey = make(map[string]int)

	playerRepository.EXPECT().
		GetAllPlayers(gomock.Any(), "network", "channel").
		Return([]*player.Player{
			{ID: "1", Name: "player1", Points: 100, Count: 10},
		}, nil).
		Times(1)

	expectedErr := assert.AnError
	playerRepository.EXPECT().
		UpsertPlayer(gomock.Any(), gomock.Any()).
		Return(expectedErr).
		Times(1)

	ircClient := gameMocks.NewMockIRCClient(ctrl)
	ircClient.EXPECT().Privmsg("channel", gomock.Any()).AnyTimes()

	gameInstance := game.NewGame(config.GameConfig{Interval: 300}, ircClient, playerRepository, "network", "channel")

	ctx := context.Background()
	go gameInstance.Start(ctx)
	time.Sleep(50 * time.Millisecond)

	err := gameInstance.SavePlayers(ctx)
	assert.Equal(t, expectedErr, err)
}

func TestAddPlayer_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	playerRepository := mocks.NewMockPlayerRepository(ctrl)
	playerRepository.EggsByKey = make(map[string]int)
	playerRepository.RareEggsByKey = make(map[string]int)

	playerRepository.EXPECT().
		GetAllPlayers(gomock.Any(), "network", "channel").
		Return([]*player.Player{}, nil).
		Times(1)

	expectedErr := assert.AnError
	playerRepository.EXPECT().
		UpsertPlayer(gomock.Any(), gomock.Any()).
		Return(expectedErr).
		Times(1)

	ircClient := gameMocks.NewMockIRCClient(ctrl)
	ircClient.EXPECT().Privmsg("channel", gomock.Any()).AnyTimes()

	gameInstance := game.NewGame(config.GameConfig{Interval: 300}, ircClient, playerRepository, "network", "channel")

	ctx := context.Background()
	go gameInstance.Start(ctx)
	time.Sleep(50 * time.Millisecond)

	// FindPlayer should try to add a new player and fail
	foundPlayer, err := gameInstance.FindPlayer(ctx, "newplayer")
	assert.Equal(t, expectedErr, err)
	assert.Nil(t, foundPlayer)
}

func TestSyncPlayers_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	playerRepository := mocks.NewMockPlayerRepository(ctrl)
	playerRepository.EggsByKey = make(map[string]int)
	playerRepository.RareEggsByKey = make(map[string]int)

	// Return error when getting all players
	playerRepository.EXPECT().
		GetAllPlayers(gomock.Any(), "network", "channel").
		Return(nil, assert.AnError).
		Times(1)

	ircClient := gameMocks.NewMockIRCClient(ctrl)
	ircClient.EXPECT().Privmsg("channel", gomock.Any()).AnyTimes()

	gameInstance := game.NewGame(config.GameConfig{Interval: 300}, ircClient, playerRepository, "network", "channel")

	ctx := context.Background()
	go gameInstance.Start(ctx)
	time.Sleep(50 * time.Millisecond)

	// Game should still work, just without synced players
	err := gameInstance.HandlePoints(ctx)
	assert.Nil(t, err)
}

func TestGameStart_DefaultInterval(t *testing.T) {
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

	// Game with 0 interval should default to 120 seconds
	gameInstance := game.NewGame(config.GameConfig{Interval: 0}, ircClient, playerRepository, "network", "channel")

	ctx, cancel := context.WithCancel(context.Background())

	go gameInstance.Start(ctx)
	time.Sleep(50 * time.Millisecond)
	cancel()

	// Just verify it doesn't crash
	assert.NotNil(t, gameInstance)
}

func TestHandleEggs_GetEggsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	playerRepository := mocks.NewMockPlayerRepository(ctrl)
	playerRepository.EggsByKey = make(map[string]int)
	playerRepository.RareEggsByKey = make(map[string]int)
	playerRepository.GetEggsErr = assert.AnError

	playerRepository.EXPECT().
		GetAllPlayers(gomock.Any(), "network", "channel").
		Return([]*player.Player{}, nil).
		Times(1)

	ircClient := gameMocks.NewMockIRCClient(ctrl)
	ircClient.EXPECT().Privmsg("channel", gomock.Any()).AnyTimes()

	gameInstance := game.NewGame(config.GameConfig{Interval: 300}, ircClient, playerRepository, "network", "channel")

	ctx := context.WithValue(context.Background(), "nick", "testuser")
	go gameInstance.Start(ctx)
	time.Sleep(50 * time.Millisecond)

	err := gameInstance.HandleEggs(ctx)
	assert.Equal(t, assert.AnError, err)
}

func TestHandleEggs_GetRareEggsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	playerRepository := mocks.NewMockPlayerRepository(ctrl)
	playerRepository.EggsByKey = make(map[string]int)
	playerRepository.RareEggsByKey = make(map[string]int)
	playerRepository.GetRareEggsErr = assert.AnError

	playerRepository.EXPECT().
		GetAllPlayers(gomock.Any(), "network", "channel").
		Return([]*player.Player{}, nil).
		Times(1)

	ircClient := gameMocks.NewMockIRCClient(ctrl)
	ircClient.EXPECT().Privmsg("channel", gomock.Any()).AnyTimes()

	gameInstance := game.NewGame(config.GameConfig{Interval: 300}, ircClient, playerRepository, "network", "channel")

	ctx := context.WithValue(context.Background(), "nick", "testuser")
	go gameInstance.Start(ctx)
	time.Sleep(50 * time.Millisecond)

	err := gameInstance.HandleEggs(ctx)
	assert.Equal(t, assert.AnError, err)
}
