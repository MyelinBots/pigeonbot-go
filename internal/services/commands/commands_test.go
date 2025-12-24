package commands_test

import (
	"context"
	"errors"
	"testing"

	"github.com/MyelinBots/pigeonbot-go/config"
	"github.com/MyelinBots/pigeonbot-go/internal/db/repositories/player"
	"github.com/MyelinBots/pigeonbot-go/internal/db/repositories/player/mocks"
	"github.com/MyelinBots/pigeonbot-go/internal/services/commands"
	"github.com/MyelinBots/pigeonbot-go/internal/services/context_manager"
	"github.com/MyelinBots/pigeonbot-go/internal/services/game"
	gameMocks "github.com/MyelinBots/pigeonbot-go/internal/services/game/mocks"
	irc "github.com/fluffle/goirc/client"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewCommandController(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	playerRepository := mocks.NewMockPlayerRepository(ctrl)
	playerRepository.EXPECT().
		GetAllPlayers(gomock.Any(), "network", "channel").
		Return([]*player.Player{}, nil).
		AnyTimes()

	ircClient := gameMocks.NewMockIRCClient(ctrl)

	gameInstance := game.NewGame(
		config.GameConfig{Interval: 3},
		ircClient,
		playerRepository,
		"network",
		"channel",
	)

	controller := commands.NewCommandController(gameInstance)

	assert.NotNil(t, controller)
}

func TestCommandController_AddCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	playerRepository := mocks.NewMockPlayerRepository(ctrl)
	playerRepository.EXPECT().
		GetAllPlayers(gomock.Any(), "network", "channel").
		Return([]*player.Player{}, nil).
		AnyTimes()

	ircClient := gameMocks.NewMockIRCClient(ctrl)

	gameInstance := game.NewGame(
		config.GameConfig{Interval: 3},
		ircClient,
		playerRepository,
		"network",
		"channel",
	)

	controller := commands.NewCommandController(gameInstance)

	// Add a command
	called := false
	controller.AddCommand("!test", func(ctx context.Context, args ...string) error {
		called = true
		return nil
	})

	// Execute the command
	line := &irc.Line{
		Args: []string{"channel", "!test"},
		Nick: "testuser",
	}

	err := controller.HandleCommand(context.Background(), line)

	assert.Nil(t, err)
	assert.True(t, called)
}

func TestCommandController_HandleCommand_NilLine(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	playerRepository := mocks.NewMockPlayerRepository(ctrl)
	playerRepository.EXPECT().
		GetAllPlayers(gomock.Any(), "network", "channel").
		Return([]*player.Player{}, nil).
		AnyTimes()

	ircClient := gameMocks.NewMockIRCClient(ctrl)

	gameInstance := game.NewGame(
		config.GameConfig{Interval: 3},
		ircClient,
		playerRepository,
		"network",
		"channel",
	)

	controller := commands.NewCommandController(gameInstance)

	err := controller.HandleCommand(context.Background(), nil)

	assert.Nil(t, err)
}

func TestCommandController_HandleCommand_InsufficientArgs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	playerRepository := mocks.NewMockPlayerRepository(ctrl)
	playerRepository.EXPECT().
		GetAllPlayers(gomock.Any(), "network", "channel").
		Return([]*player.Player{}, nil).
		AnyTimes()

	ircClient := gameMocks.NewMockIRCClient(ctrl)

	gameInstance := game.NewGame(
		config.GameConfig{Interval: 3},
		ircClient,
		playerRepository,
		"network",
		"channel",
	)

	controller := commands.NewCommandController(gameInstance)

	// Line with only one arg
	line := &irc.Line{
		Args: []string{"channel"},
		Nick: "testuser",
	}

	err := controller.HandleCommand(context.Background(), line)

	assert.Nil(t, err)
}

func TestCommandController_HandleCommand_UnknownCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	playerRepository := mocks.NewMockPlayerRepository(ctrl)
	playerRepository.EXPECT().
		GetAllPlayers(gomock.Any(), "network", "channel").
		Return([]*player.Player{}, nil).
		AnyTimes()

	ircClient := gameMocks.NewMockIRCClient(ctrl)

	gameInstance := game.NewGame(
		config.GameConfig{Interval: 3},
		ircClient,
		playerRepository,
		"network",
		"channel",
	)

	controller := commands.NewCommandController(gameInstance)

	// Unknown command
	line := &irc.Line{
		Args: []string{"channel", "!unknown"},
		Nick: "testuser",
	}

	err := controller.HandleCommand(context.Background(), line)

	// Should not error on unknown command
	assert.Nil(t, err)
}

func TestCommandController_HandleCommand_WithArgs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	playerRepository := mocks.NewMockPlayerRepository(ctrl)
	playerRepository.EXPECT().
		GetAllPlayers(gomock.Any(), "network", "channel").
		Return([]*player.Player{}, nil).
		AnyTimes()

	ircClient := gameMocks.NewMockIRCClient(ctrl)

	gameInstance := game.NewGame(
		config.GameConfig{Interval: 3},
		ircClient,
		playerRepository,
		"network",
		"channel",
	)

	controller := commands.NewCommandController(gameInstance)

	var receivedArgs []string
	controller.AddCommand("!test", func(ctx context.Context, args ...string) error {
		receivedArgs = args
		return nil
	})

	line := &irc.Line{
		Args: []string{"channel", "!test", "arg1", "arg2"},
		Nick: "testuser",
	}

	err := controller.HandleCommand(context.Background(), line)

	assert.Nil(t, err)
	// First arg should be the nick, then additional args
	assert.Equal(t, []string{"testuser", "arg1", "arg2"}, receivedArgs)
}

func TestCommandController_HandleCommand_NickInContext(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	playerRepository := mocks.NewMockPlayerRepository(ctrl)
	playerRepository.EXPECT().
		GetAllPlayers(gomock.Any(), "network", "channel").
		Return([]*player.Player{}, nil).
		AnyTimes()

	ircClient := gameMocks.NewMockIRCClient(ctrl)

	gameInstance := game.NewGame(
		config.GameConfig{Interval: 3},
		ircClient,
		playerRepository,
		"network",
		"channel",
	)

	controller := commands.NewCommandController(gameInstance)

	var capturedNick string
	controller.AddCommand("!test", func(ctx context.Context, args ...string) error {
		capturedNick = context_manager.GetNickContext(ctx)
		return nil
	})

	line := &irc.Line{
		Args: []string{"channel", "!test"},
		Nick: "TestUser",
	}

	err := controller.HandleCommand(context.Background(), line)

	assert.Nil(t, err)
	// Nick should be normalized to lowercase in context
	assert.Equal(t, "testuser", capturedNick)
}

func TestCommandController_HandleCommand_HandlerError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	playerRepository := mocks.NewMockPlayerRepository(ctrl)
	playerRepository.EXPECT().
		GetAllPlayers(gomock.Any(), "network", "channel").
		Return([]*player.Player{}, nil).
		AnyTimes()

	ircClient := gameMocks.NewMockIRCClient(ctrl)

	gameInstance := game.NewGame(
		config.GameConfig{Interval: 3},
		ircClient,
		playerRepository,
		"network",
		"channel",
	)

	controller := commands.NewCommandController(gameInstance)

	expectedErr := errors.New("handler error")
	controller.AddCommand("!error", func(ctx context.Context, args ...string) error {
		return expectedErr
	})

	line := &irc.Line{
		Args: []string{"channel", "!error"},
		Nick: "testuser",
	}

	err := controller.HandleCommand(context.Background(), line)

	assert.Equal(t, expectedErr, err)
}

func TestCommandController_MultipleCommands(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	playerRepository := mocks.NewMockPlayerRepository(ctrl)
	playerRepository.EXPECT().
		GetAllPlayers(gomock.Any(), "network", "channel").
		Return([]*player.Player{}, nil).
		AnyTimes()

	ircClient := gameMocks.NewMockIRCClient(ctrl)

	gameInstance := game.NewGame(
		config.GameConfig{Interval: 3},
		ircClient,
		playerRepository,
		"network",
		"channel",
	)

	controller := commands.NewCommandController(gameInstance)

	calledCommand := ""

	controller.AddCommand("!cmd1", func(ctx context.Context, args ...string) error {
		calledCommand = "cmd1"
		return nil
	})

	controller.AddCommand("!cmd2", func(ctx context.Context, args ...string) error {
		calledCommand = "cmd2"
		return nil
	})

	// Test first command
	line1 := &irc.Line{Args: []string{"channel", "!cmd1"}, Nick: "user"}
	err := controller.HandleCommand(context.Background(), line1)
	assert.Nil(t, err)
	assert.Equal(t, "cmd1", calledCommand)

	// Test second command
	line2 := &irc.Line{Args: []string{"channel", "!cmd2"}, Nick: "user"}
	err = controller.HandleCommand(context.Background(), line2)
	assert.Nil(t, err)
	assert.Equal(t, "cmd2", calledCommand)
}

func TestCommandController_EmptyArgs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	playerRepository := mocks.NewMockPlayerRepository(ctrl)
	playerRepository.EXPECT().
		GetAllPlayers(gomock.Any(), "network", "channel").
		Return([]*player.Player{}, nil).
		AnyTimes()

	ircClient := gameMocks.NewMockIRCClient(ctrl)

	gameInstance := game.NewGame(
		config.GameConfig{Interval: 3},
		ircClient,
		playerRepository,
		"network",
		"channel",
	)

	controller := commands.NewCommandController(gameInstance)

	line := &irc.Line{
		Args: []string{},
		Nick: "testuser",
	}

	err := controller.HandleCommand(context.Background(), line)

	assert.Nil(t, err)
}
