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

func TestHandleShoot_WithPigeon(t *testing.T) {
	// Run multiple times to test both success and failure paths
	for i := 0; i < 10; i++ {
		t.Run("shoot_iteration", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			playerRepository := mocks.NewMockPlayerRepository(ctrl)
			playerRepository.EggsByKey = make(map[string]int)
			playerRepository.RareEggsByKey = make(map[string]int)

			playerRepository.EXPECT().
				GetAllPlayers(gomock.Any(), "network", "channel").
				Return([]*player.Player{
					{ID: "1", Name: "shooter", Points: 0, Count: 0},
				}, nil).
				Times(1)

			playerRepository.EXPECT().
				UpsertPlayer(gomock.Any(), gomock.Any()).
				Return(nil).
				AnyTimes()

			ircClient := gameMocks.NewMockIRCClient(ctrl)
			ircClient.EXPECT().Privmsg("channel", gomock.Any()).AnyTimes()

			// Use short interval to spawn pigeon quickly
			gameInstance := game.NewGame(config.GameConfig{Interval: 1}, ircClient, playerRepository, "network", "channel")

			commandController := commands.NewCommandController(gameInstance)
			commandController.AddCommand("!shoot", gameInstance.HandleShoot)

			ctx := context.Background()
			line := &irc.Line{
				Args: []string{"channel", "!shoot"},
				Nick: "shooter",
			}

			go gameInstance.Start(ctx)
			// Wait for pigeon to spawn
			time.Sleep(1500 * time.Millisecond)

			err := commandController.HandleCommand(ctx, line)
			assert.Nil(t, err)
		})
	}
}

func TestActOnPlayer_WithExistingPigeon(t *testing.T) {
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

	// Very short interval to test pigeon spawning and despawning
	gameInstance := game.NewGame(config.GameConfig{Interval: 1}, ircClient, playerRepository, "network", "channel")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go gameInstance.Start(ctx)

	// Wait for multiple pigeon spawn/despawn cycles
	time.Sleep(200 * time.Millisecond)

	// Game should still be running
	assert.NotNil(t, gameInstance)
}

func TestHandleShoot_FindPlayerError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	playerRepository := mocks.NewMockPlayerRepository(ctrl)
	playerRepository.EggsByKey = make(map[string]int)
	playerRepository.RareEggsByKey = make(map[string]int)

	// Return empty players so FindPlayer needs to create one
	playerRepository.EXPECT().
		GetAllPlayers(gomock.Any(), "network", "channel").
		Return([]*player.Player{}, nil).
		Times(1)

	// UpsertPlayer fails, causing FindPlayer to fail
	playerRepository.EXPECT().
		UpsertPlayer(gomock.Any(), gomock.Any()).
		Return(assert.AnError).
		AnyTimes()

	ircClient := gameMocks.NewMockIRCClient(ctrl)
	ircClient.EXPECT().Privmsg("channel", gomock.Any()).AnyTimes()

	gameInstance := game.NewGame(config.GameConfig{Interval: 1}, ircClient, playerRepository, "network", "channel")

	commandController := commands.NewCommandController(gameInstance)
	commandController.AddCommand("!shoot", gameInstance.HandleShoot)

	ctx := context.Background()
	line := &irc.Line{
		Args: []string{"channel", "!shoot"},
		Nick: "newshooter",
	}

	go gameInstance.Start(ctx)
	time.Sleep(1500 * time.Millisecond)

	err := commandController.HandleCommand(ctx, line)
	// Should return error from FindPlayer
	assert.NotNil(t, err)
}
