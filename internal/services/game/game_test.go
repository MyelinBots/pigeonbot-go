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

// If your mock stores eggs/rare eggs in maps, initialize them here.
// NOTE: do NOT use repo.EXPECT() for eggs methods unless MockGen generated recorder methods for them.
func initEggStores(repo *mocks.MockPlayerRepository) {
	if repo.EggsByKey == nil {
		repo.EggsByKey = make(map[string]int)
	}
	// If you add RareEggsByKey to the mock (see below), init it too:
	// if repo.RareEggsByKey == nil {
	// 	repo.RareEggsByKey = make(map[string]int)
	// }
}

func TestGame(t *testing.T) {

	t.Run("TestGame", func(t *testing.T) {
		t.Parallel()
		a := assert.New(t)

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		playerRepository := mocks.NewMockPlayerRepository(ctrl)
		initEggStores(playerRepository)

		playerRepository.EXPECT().
			GetAllPlayers(gomock.Any(), "network", "channel").
			Return([]*player.Player{
				{
					ID:      "test",
					Name:    "test",
					Points:  0,
					Count:   0,
					Channel: "channel",
					Network: "network",
				},
			}, nil).
			Times(1)

		playerRepository.EXPECT().
			UpsertPlayer(gomock.Any(), gomock.Any()).
			Return(nil).
			AnyTimes()

		ircClient := gameMocks.NewMockIRCClient(ctrl)
		ircClient.EXPECT().Privmsg("channel", gomock.Any()).AnyTimes()

		gameinstance := game.NewGame(config.GameConfig{Interval: 3}, ircClient, playerRepository, "network", "channel")

		commandController := commands.NewCommandController(gameinstance)
		commandController.AddCommand("!shoot", gameinstance.HandleShoot)

		ctx := context.Background()
		line := &irc.Line{
			Args: []string{"channel", "!shoot"},
			Nick: "test",
		}

		go gameinstance.Start(ctx)

		time.Sleep(2 * time.Second)
		err := commandController.HandleCommand(ctx, line)
		a.Nil(err)
	})

	t.Run("TestGame With No Players", func(t *testing.T) {
		t.Parallel()
		a := assert.New(t)

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		playerRepository := mocks.NewMockPlayerRepository(ctrl)
		initEggStores(playerRepository)

		playerRepository.EXPECT().
			GetAllPlayers(gomock.Any(), "network", "channel").
			Return([]*player.Player{}, nil).
			Times(1)

		playerRepository.EXPECT().
			UpsertPlayer(gomock.Any(), gomock.Any()).
			Return(nil).
			AnyTimes()

		ircClient := gameMocks.NewMockIRCClient(ctrl)
		ircClient.EXPECT().Privmsg("channel", gomock.Any()).AnyTimes()

		gameinstance := game.NewGame(config.GameConfig{Interval: 3}, ircClient, playerRepository, "network", "channel")

		commandController := commands.NewCommandController(gameinstance)
		commandController.AddCommand("!shoot", gameinstance.HandleShoot)

		ctx := context.Background()
		line := &irc.Line{
			Args: []string{"channel", "!shoot"},
			Nick: "test2",
		}

		go gameinstance.Start(ctx)

		time.Sleep(2 * time.Second)
		err := commandController.HandleCommand(ctx, line)
		a.Nil(err)
	})

	t.Run("TestGame With Players points", func(t *testing.T) {
		t.Parallel()
		a := assert.New(t)

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		playerRepository := mocks.NewMockPlayerRepository(ctrl)
		initEggStores(playerRepository)

		playerRepository.EXPECT().
			GetAllPlayers(gomock.Any(), "network", "channel").
			Return([]*player.Player{
				{ID: "test", Name: "test", Points: 0, Count: 0, Channel: "channel", Network: "network"},
				{ID: "test2", Name: "test2", Points: 10, Count: 0, Channel: "channel", Network: "network"},
			}, nil).
			Times(1)

		ircClient := gameMocks.NewMockIRCClient(ctrl)
		ircClient.EXPECT().Privmsg("channel", gomock.Any()).Times(1)
		ircClient.EXPECT().Privmsg("channel", "test2: 10, test: 0, ").Times(1)

		gameinstance := game.NewGame(config.GameConfig{Interval: 3}, ircClient, playerRepository, "network", "channel")

		commandController := commands.NewCommandController(gameinstance)
		commandController.AddCommand("!score", gameinstance.HandlePoints)

		ctx := context.Background()
		line := &irc.Line{
			Args: []string{"channel", "!score"},
			Nick: "test2",
		}

		go gameinstance.Start(ctx)

		time.Sleep(2 * time.Second)
		err := commandController.HandleCommand(ctx, line)
		a.Nil(err)
	})

	t.Run("TestGame With Players count", func(t *testing.T) {
		t.Parallel()
		a := assert.New(t)

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		playerRepository := mocks.NewMockPlayerRepository(ctrl)
		initEggStores(playerRepository)

		playerRepository.EXPECT().
			GetAllPlayers(gomock.Any(), "network", "channel").
			Return([]*player.Player{
				{ID: "test", Name: "test", Points: 0, Count: 0, Channel: "channel", Network: "network"},
				{ID: "test2", Name: "test2", Points: 10, Count: 0, Channel: "channel", Network: "network"},
			}, nil).
			Times(1)

		ircClient := gameMocks.NewMockIRCClient(ctrl)
		ircClient.EXPECT().Privmsg("channel", gomock.Any()).Times(1)
		ircClient.EXPECT().Privmsg("channel", "test: 0, test2: 0, ").Times(1)

		gameinstance := game.NewGame(config.GameConfig{Interval: 3}, ircClient, playerRepository, "network", "channel")

		commandController := commands.NewCommandController(gameinstance)
		commandController.AddCommand("!pigeons", gameinstance.HandleCount)

		ctx := context.Background()
		line := &irc.Line{
			Args: []string{"channel", "!pigeons"},
			Nick: "test2",
		}

		go gameinstance.Start(ctx)

		time.Sleep(2 * time.Second)
		err := commandController.HandleCommand(ctx, line)
		a.Nil(err)
	})

	t.Run("TestGame With Players bef", func(t *testing.T) {
		t.Parallel()
		a := assert.New(t)

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		playerRepository := mocks.NewMockPlayerRepository(ctrl)
		initEggStores(playerRepository)

		playerRepository.EXPECT().
			GetAllPlayers(gomock.Any(), "network", "channel").
			Return([]*player.Player{
				{ID: "test", Name: "test", Points: 0, Count: 0, Channel: "channel", Network: "network"},
				{ID: "test2", Name: "test2", Points: 10, Count: 0, Channel: "channel", Network: "network"},
			}, nil).
			Times(1)

		ircClient := gameMocks.NewMockIRCClient(ctrl)
		ircClient.EXPECT().Privmsg("channel", gomock.Any()).Times(1)
		ircClient.EXPECT().Privmsg("channel", "🕊️ ~ coo coo ~ cannot be frens with a rat of the sky ~ 🕊️").Times(1)

		gameinstance := game.NewGame(config.GameConfig{Interval: 3}, ircClient, playerRepository, "network", "channel")

		commandController := commands.NewCommandController(gameinstance)
		commandController.AddCommand("!bef", gameinstance.HandleBef)

		ctx := context.Background()
		line := &irc.Line{
			Args: []string{"channel", "!bef"},
			Nick: "test2",
		}

		go gameinstance.Start(ctx)

		time.Sleep(2 * time.Second)
		err := commandController.HandleCommand(ctx, line)
		a.Nil(err)
	})

	t.Run("TestGame help", func(t *testing.T) {
		t.Parallel()
		a := assert.New(t)

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		playerRepository := mocks.NewMockPlayerRepository(ctrl)
		initEggStores(playerRepository)

		playerRepository.EXPECT().
			GetAllPlayers(gomock.Any(), "network", "channel").
			Return([]*player.Player{
				{ID: "test", Name: "test", Points: 0, Count: 0, Channel: "channel", Network: "network"},
				{ID: "test2", Name: "test2", Points: 10, Count: 0, Channel: "channel", Network: "network"},
			}, nil).
			Times(1)

		ircClient := gameMocks.NewMockIRCClient(ctrl)
		ircClient.EXPECT().Privmsg("channel", gomock.Any()).Times(1)
		ircClient.EXPECT().Privmsg("channel", "Commands: !shoot, !score, !pigeons, !bef, !help, !level").Times(1)

		gameinstance := game.NewGame(config.GameConfig{Interval: 3}, ircClient, playerRepository, "network", "channel")

		commandController := commands.NewCommandController(gameinstance)
		commandController.AddCommand("!help", gameinstance.HandleHelp)

		ctx := context.Background()
		line := &irc.Line{
			Args: []string{"channel", "!help"},
			Nick: "test2",
		}

		go gameinstance.Start(ctx)

		time.Sleep(2 * time.Second)
		err := commandController.HandleCommand(ctx, line)
		a.Nil(err)
	})
}
