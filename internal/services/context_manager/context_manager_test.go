package context_manager_test

import (
	"context"
	"testing"

	"github.com/MyelinBots/pigeonbot-go/internal/services/context_manager"
	"github.com/stretchr/testify/assert"
)

func TestWithNick(t *testing.T) {
	tests := []struct {
		name     string
		nick     string
		expected string
	}{
		{
			name:     "lowercase nick",
			nick:     "testuser",
			expected: "testuser",
		},
		{
			name:     "uppercase nick",
			nick:     "TESTUSER",
			expected: "testuser",
		},
		{
			name:     "mixed case nick",
			nick:     "TestUser",
			expected: "testuser",
		},
		{
			name:     "empty nick",
			nick:     "",
			expected: "",
		},
		{
			name:     "nick with numbers",
			nick:     "User123",
			expected: "user123",
		},
		{
			name:     "nick with special characters",
			nick:     "Test_User-123",
			expected: "test_user-123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			newCtx := context_manager.WithNick(ctx, tt.nick)

			// Verify context is not nil
			assert.NotNil(t, newCtx)

			// Verify we can retrieve the nick
			result := context_manager.GetNickContext(newCtx)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetNickContext(t *testing.T) {
	t.Run("returns nick when set", func(t *testing.T) {
		ctx := context.Background()
		ctx = context_manager.WithNick(ctx, "TestPlayer")

		result := context_manager.GetNickContext(ctx)

		assert.Equal(t, "testplayer", result)
	})

	t.Run("returns empty string when not set", func(t *testing.T) {
		ctx := context.Background()

		result := context_manager.GetNickContext(ctx)

		assert.Equal(t, "", result)
	})

	t.Run("returns empty string for nil context value", func(t *testing.T) {
		ctx := context.Background()
		// Context without nick set
		result := context_manager.GetNickContext(ctx)

		assert.Equal(t, "", result)
	})
}

func TestWithNick_ChainedContexts(t *testing.T) {
	ctx := context.Background()

	// Add first nick
	ctx1 := context_manager.WithNick(ctx, "FirstUser")
	assert.Equal(t, "firstuser", context_manager.GetNickContext(ctx1))

	// Add second nick (should override in new context)
	ctx2 := context_manager.WithNick(ctx1, "SecondUser")
	assert.Equal(t, "seconduser", context_manager.GetNickContext(ctx2))

	// Original context should still have first nick
	assert.Equal(t, "firstuser", context_manager.GetNickContext(ctx1))

	// Background context should have no nick
	assert.Equal(t, "", context_manager.GetNickContext(ctx))
}

func TestWithNick_ContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	nickCtx := context_manager.WithNick(ctx, "TestUser")

	// Nick should be retrievable before cancel
	assert.Equal(t, "testuser", context_manager.GetNickContext(nickCtx))

	// Cancel the context
	cancel()

	// Nick should still be retrievable after cancel
	assert.Equal(t, "testuser", context_manager.GetNickContext(nickCtx))
}

func TestNickNormalization(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"ABC", "abc"},
		{"abc", "abc"},
		{"AbC", "abc"},
		{"ABC123", "abc123"},
		{"User_Name", "user_name"},
		{"USER-NAME", "user-name"},
		{"  ", "  "}, // Note: spaces are preserved, only case is changed
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			ctx := context_manager.WithNick(context.Background(), tt.input)
			result := context_manager.GetNickContext(ctx)
			assert.Equal(t, tt.expected, result)
		})
	}
}
