package commands_test

import (
	"context"
	"errors"
	"testing"

	"github.com/MyelinBots/pigeonbot-go/config"
	playerRepoPkg "github.com/MyelinBots/pigeonbot-go/internal/db/repositories/player"
	playerRepoMocks "github.com/MyelinBots/pigeonbot-go/internal/db/repositories/player/mocks"
	"github.com/MyelinBots/pigeonbot-go/internal/services/commands"
	"github.com/MyelinBots/pigeonbot-go/internal/services/context_manager"
	"github.com/MyelinBots/pigeonbot-go/internal/services/game"
	gameMocks "github.com/MyelinBots/pigeonbot-go/internal/services/game/mocks"
	irc "github.com/fluffle/goirc/client"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func newController(t *testing.T) (*gomock.Controller, commands.CommandController) {
	t.Helper()

	ctrl := gomock.NewController(t)

	playerRepository := playerRepoMocks.NewMockPlayerRepository(ctrl)
	playerRepository.EXPECT().
		GetAllPlayers(gomock.Any(), "network", "channel").
		Return([]*playerRepoPkg.Player{}, nil).
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
	return ctrl, controller
}

func TestNewCommandController(t *testing.T) {
	ctrl, controller := newController(t)
	defer ctrl.Finish()

	assert.NotNil(t, controller)
}

func TestCommandController_AddCommand(t *testing.T) {
	ctrl, controller := newController(t)
	defer ctrl.Finish()

	called := false
	controller.AddCommand("!test", func(ctx context.Context, args ...string) error {
		called = true
		return nil
	})

	line := &irc.Line{
		Args: []string{"channel", "!test"},
		Nick: "testuser",
	}

	err := controller.HandleCommand(context.Background(), line)

	assert.NoError(t, err)
	assert.True(t, called)
}

func TestCommandController_HandleCommand_NilLine(t *testing.T) {
	ctrl, controller := newController(t)
	defer ctrl.Finish()

	err := controller.HandleCommand(context.Background(), nil)
	assert.NoError(t, err)
}

func TestCommandController_HandleCommand_InsufficientArgs(t *testing.T) {
	ctrl, controller := newController(t)
	defer ctrl.Finish()

	// Line มีแค่ channel ไม่มี message
	line := &irc.Line{
		Args: []string{"channel"},
		Nick: "testuser",
	}

	err := controller.HandleCommand(context.Background(), line)
	assert.NoError(t, err)
}

func TestCommandController_HandleCommand_EmptyMessage(t *testing.T) {
	ctrl, controller := newController(t)
	defer ctrl.Finish()

	line := &irc.Line{
		Args: []string{"channel", ""},
		Nick: "testuser",
	}

	err := controller.HandleCommand(context.Background(), line)
	assert.NoError(t, err)
}

func TestCommandController_HandleCommand_WhitespaceMessage(t *testing.T) {
	ctrl, controller := newController(t)
	defer ctrl.Finish()

	line := &irc.Line{
		Args: []string{"channel", "   "},
		Nick: "testuser",
	}

	err := controller.HandleCommand(context.Background(), line)
	assert.NoError(t, err)
}

func TestCommandController_HandleCommand_UnknownCommand(t *testing.T) {
	ctrl, controller := newController(t)
	defer ctrl.Finish()

	line := &irc.Line{
		Args: []string{"channel", "!unknown"},
		Nick: "testuser",
	}

	err := controller.HandleCommand(context.Background(), line)
	assert.NoError(t, err)
}

func TestCommandController_HandleCommand_WithArgs(t *testing.T) {
	ctrl, controller := newController(t)
	defer ctrl.Finish()

	var receivedArgs []string
	controller.AddCommand("!test", func(ctx context.Context, args ...string) error {
		receivedArgs = args
		return nil
	})

	// สำคัญ: goirc ไม่แยก arg ให้เรา — message อยู่ใน Args[1] เป็นทั้งบรรทัด
	line := &irc.Line{
		Args: []string{"channel", "!test arg1 arg2"},
		Nick: "testuser",
	}

	err := controller.HandleCommand(context.Background(), line)

	assert.NoError(t, err)
	// ตอนนี้ args คือ "หลังคำสั่ง" เท่านั้น (ไม่รวม nick)
	assert.Equal(t, []string{"arg1", "arg2"}, receivedArgs)
}

func TestCommandController_HandleCommand_WithExtraSpaces(t *testing.T) {
	ctrl, controller := newController(t)
	defer ctrl.Finish()

	var receivedArgs []string
	controller.AddCommand("!test", func(ctx context.Context, args ...string) error {
		receivedArgs = args
		return nil
	})

	line := &irc.Line{
		Args: []string{"channel", "   !test    arg1    arg2   "},
		Nick: "testuser",
	}

	err := controller.HandleCommand(context.Background(), line)

	assert.NoError(t, err)
	assert.Equal(t, []string{"arg1", "arg2"}, receivedArgs)
}

func TestCommandController_HandleCommand_CaseInsensitiveCommand(t *testing.T) {
	ctrl, controller := newController(t)
	defer ctrl.Finish()

	called := false
	controller.AddCommand("!test", func(ctx context.Context, args ...string) error {
		called = true
		return nil
	})

	line := &irc.Line{
		Args: []string{"channel", "!TeSt"},
		Nick: "testuser",
	}

	err := controller.HandleCommand(context.Background(), line)
	assert.NoError(t, err)
	assert.True(t, called)
}

func TestCommandController_HandleCommand_NickInContext(t *testing.T) {
	ctrl, controller := newController(t)
	defer ctrl.Finish()

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

	assert.NoError(t, err)
	// ถ้า context_manager.WithNick ของคุณ normalize เป็น lowercase อยู่แล้ว เทสนี้จะผ่าน
	assert.Equal(t, "testuser", capturedNick)
}

func TestCommandController_HandleCommand_HandlerError(t *testing.T) {
	ctrl, controller := newController(t)
	defer ctrl.Finish()

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
	ctrl, controller := newController(t)
	defer ctrl.Finish()

	called := ""

	controller.AddCommand("!cmd1", func(ctx context.Context, args ...string) error {
		called = "cmd1"
		return nil
	})

	controller.AddCommand("!cmd2", func(ctx context.Context, args ...string) error {
		called = "cmd2"
		return nil
	})

	line1 := &irc.Line{Args: []string{"channel", "!cmd1"}, Nick: "user"}
	err := controller.HandleCommand(context.Background(), line1)
	assert.NoError(t, err)
	assert.Equal(t, "cmd1", called)

	line2 := &irc.Line{Args: []string{"channel", "!cmd2"}, Nick: "user"}
	err = controller.HandleCommand(context.Background(), line2)
	assert.NoError(t, err)
	assert.Equal(t, "cmd2", called)
}

func TestCommandController_EmptyArgs(t *testing.T) {
	ctrl, controller := newController(t)
	defer ctrl.Finish()

	line := &irc.Line{
		Args: []string{},
		Nick: "testuser",
	}

	err := controller.HandleCommand(context.Background(), line)
	assert.NoError(t, err)
}
