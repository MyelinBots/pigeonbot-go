package game_test

import (
	"context"
	"github.com/MyelinBots/pigeonbot-go/config"
	"github.com/MyelinBots/pigeonbot-go/internal/db/repositories/player"
	"github.com/MyelinBots/pigeonbot-go/internal/db/repositories/player/mocks"
	"github.com/MyelinBots/pigeonbot-go/internal/services/commands"
	"github.com/MyelinBots/pigeonbot-go/internal/services/game"
	mocks2 "github.com/MyelinBots/pigeonbot-go/internal/services/game/mocks"
	irc "github.com/fluffle/goirc/client"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
)

func TestGame(t *testing.T) {
	t.Run("TestGame", func(t *testing.T) {
		t.Parallel()
		a := assert.New(t)

		ctrl := gomock.NewController(t)

		playerRepository := mocks.NewMockPlayerRepository(ctrl)
		playerRepository.EXPECT().GetAllPlayers(gomock.Any()).Return([]*player.Player{
			{
				ID:      "test",
				Name:    "test",
				Points:  0,
				Count:   0,
				Channel: "channel",
				Network: "network",
			},
		}, nil).Times(1)

		playerRepository.EXPECT().UpsertPlayer(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

		ircClient := mocks2.NewMockIRCClient(ctrl)
		ircClient.EXPECT().Privmsg("channel", gomock.Any()).Times(2)

		// Create a new game
		gameinstance := game.NewGame(config.GameConfig{
			Interval: 3,
		}, ircClient, playerRepository, "network", "channel")

		// Create a new command controller
		commandController := commands.NewCommandController(gameinstance)

		// Add a command to the command controller
		commandController.AddCommand("!shoot", gameinstance.HandleShoot)

		// Create a new context
		ctx := context.Background()

		// Create a new irc line
		line := &irc.Line{
			Args: []string{"channel", "!shoot", "test"},
			Nick: "test",
		}
		go func() {
			gameinstance.Start(ctx)
		}()

		time.Sleep(2 * time.Second)
		// Handle the command
		err := commandController.HandleCommand(ctx, line)
		a.Nil(err)
	})
	t.Run("TestGame With No Players", func(t *testing.T) {
		t.Parallel()
		a := assert.New(t)

		ctrl := gomock.NewController(t)

		playerRepository := mocks.NewMockPlayerRepository(ctrl)
		playerRepository.EXPECT().GetAllPlayers(gomock.Any()).Return([]*player.Player{}, nil).Times(1)
		playerRepository.EXPECT().UpsertPlayer(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

		ircClient := mocks2.NewMockIRCClient(ctrl)
		ircClient.EXPECT().Privmsg("channel", gomock.Any()).Times(2)

		// Create a new game
		gameinstance := game.NewGame(config.GameConfig{
			Interval: 3,
		}, ircClient, playerRepository, "network", "channel")

		// Create a new command controller
		commandController := commands.NewCommandController(gameinstance)

		// Add a command to the command controller
		commandController.AddCommand("!shoot", gameinstance.HandleShoot)

		// Create a new context
		ctx := context.Background()

		// Create a new irc line
		line := &irc.Line{
			Args: []string{"channel", "!shoot", "test"},
			Nick: "test2",
		}
		go func() {
			gameinstance.Start(ctx)
		}()

		time.Sleep(2 * time.Second)
		// Handle the command
		err := commandController.HandleCommand(ctx, line)
		a.Nil(err)

	})

	t.Run("TestGame With Players points", func(t *testing.T) {
		t.Parallel()
		a := assert.New(t)

		ctrl := gomock.NewController(t)

		playerRepository := mocks.NewMockPlayerRepository(ctrl)
		playerRepository.EXPECT().GetAllPlayers(gomock.Any()).Return([]*player.Player{
			{
				ID:      "test",
				Name:    "test",
				Points:  0,
				Count:   0,
				Channel: "channel",
				Network: "network",
			},
			{
				ID:      "test2",
				Name:    "test2",
				Points:  10,
				Count:   0,
				Channel: "channel",
				Network: "network",
			},
		}, nil).Times(1)

		ircClient := mocks2.NewMockIRCClient(ctrl)
		ircClient.EXPECT().Privmsg("channel", gomock.Any()).Times(1)
		ircClient.EXPECT().Privmsg("channel", "test: 0, test2: 10, ").Times(1)

		// Create a new game
		gameinstance := game.NewGame(config.GameConfig{
			Interval: 3,
		}, ircClient, playerRepository, "network", "channel")

		// Create a new command controller
		commandController := commands.NewCommandController(gameinstance)

		// Add a command to the command controller
		commandController.AddCommand("!score", gameinstance.HandlePoints)

		// Create a new context
		ctx := context.Background()

		// Create a new irc line
		line := &irc.Line{
			Args: []string{"channel", "!score", "test"},
			Nick: "test2",
		}
		go func() {
			gameinstance.Start(ctx)
		}()

		time.Sleep(2 * time.Second)
		// Handle the command
		err := commandController.HandleCommand(ctx, line)
		a.Nil(err)

	})

	t.Run("TestGame With Players count", func(t *testing.T) {
		t.Parallel()
		a := assert.New(t)

		ctrl := gomock.NewController(t)

		playerRepository := mocks.NewMockPlayerRepository(ctrl)
		playerRepository.EXPECT().GetAllPlayers(gomock.Any()).Return([]*player.Player{
			{
				ID:      "test",
				Name:    "test",
				Points:  0,
				Count:   0,
				Channel: "channel",
				Network: "network",
			},
			{
				ID:      "test2",
				Name:    "test2",
				Points:  10,
				Count:   0,
				Channel: "channel",
				Network: "network",
			},
		}, nil).Times(1)

		ircClient := mocks2.NewMockIRCClient(ctrl)
		ircClient.EXPECT().Privmsg("channel", gomock.Any()).Times(1)
		ircClient.EXPECT().Privmsg("channel", "test: 0, test2: 0, ").Times(1)

		// Create a new game
		gameinstance := game.NewGame(config.GameConfig{
			Interval: 3,
		}, ircClient, playerRepository, "network", "channel")

		// Create a new command controller
		commandController := commands.NewCommandController(gameinstance)

		// Add a command to the command controller
		commandController.AddCommand("!pigeons", gameinstance.HandleCount)

		// Create a new context
		ctx := context.Background()

		// Create a new irc line
		line := &irc.Line{
			Args: []string{"channel", "!pigeons", "test"},
			Nick: "test2",
		}
		go func() {
			gameinstance.Start(ctx)
		}()

		time.Sleep(2 * time.Second)
		// Handle the command
		err := commandController.HandleCommand(ctx, line)
		a.Nil(err)
	})

	t.Run("TestGame With Players bef", func(t *testing.T) {
		t.Parallel()
		a := assert.New(t)

		ctrl := gomock.NewController(t)

		playerRepository := mocks.NewMockPlayerRepository(ctrl)
		playerRepository.EXPECT().GetAllPlayers(gomock.Any()).Return([]*player.Player{
			{
				ID:      "test",
				Name:    "test",
				Points:  0,
				Count:   0,
				Channel: "channel",
				Network: "network",
			},
			{
				ID:      "test2",
				Name:    "test2",
				Points:  10,
				Count:   0,
				Channel: "channel",
				Network: "network",
			},
		}, nil).Times(1)

		ircClient := mocks2.NewMockIRCClient(ctrl)
		ircClient.EXPECT().Privmsg("channel", gomock.Any()).Times(1)
		ircClient.EXPECT().Privmsg("channel", "🕊️ ~ coo coo ~ cannot be frens with a rat of the sky ~ 🕊️").Times(1)

		// Create a new game
		gameinstance := game.NewGame(config.GameConfig{
			Interval: 3,
		}, ircClient, playerRepository, "network", "channel")

		// Create a new command controller
		commandController := commands.NewCommandController(gameinstance)

		// Add a command to the command controller
		commandController.AddCommand("!bef", gameinstance.HandleBef)

		// Create a new context
		ctx := context.Background()

		// Create a new irc line
		line := &irc.Line{
			Args: []string{"channel", "!bef", "test"},
			Nick: "test2",
		}
		go func() {
			gameinstance.Start(ctx)
		}()

		time.Sleep(2 * time.Second)
		// Handle the command
		err := commandController.HandleCommand(ctx, line)
		a.Nil(err)

	})

	t.Run("TestGame help", func(t *testing.T) {
		t.Parallel()
		a := assert.New(t)

		ctrl := gomock.NewController(t)

		playerRepository := mocks.NewMockPlayerRepository(ctrl)
		playerRepository.EXPECT().GetAllPlayers(gomock.Any()).Return([]*player.Player{
			{
				ID:      "test",
				Name:    "test",
				Points:  0,
				Count:   0,
				Channel: "channel",
				Network: "network",
			},
			{
				ID:      "test2",
				Name:    "test2",
				Points:  10,
				Count:   0,
				Channel: "channel",
				Network: "network",
			},
		}, nil).Times(1)

		ircClient := mocks2.NewMockIRCClient(ctrl)
		ircClient.EXPECT().Privmsg("channel", gomock.Any()).Times(1)
		ircClient.EXPECT().Privmsg("channel", "Commands: !shoot, !score, !pigeons, !bef, !help").Times(1)

		// Create a new game
		gameinstance := game.NewGame(config.GameConfig{
			Interval: 3,
		}, ircClient, playerRepository, "network", "channel")

		// Create a new command controller
		commandController := commands.NewCommandController(gameinstance)

		// Add a command to the command controller
		commandController.AddCommand("!help", gameinstance.HandleHelp)

		// Create a new context
		ctx := context.Background()

		// Create a new irc line
		line := &irc.Line{
			Args: []string{"channel", "!help", "test"},
			Nick: "test2",
		}
		go func() {
			gameinstance.Start(ctx)
		}()

		time.Sleep(2 * time.Second)
		// Handle the command
		err := commandController.HandleCommand(ctx, line)
		a.Nil(err)
	})
}
