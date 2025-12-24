package player_test

import (
	"testing"

	"github.com/MyelinBots/pigeonbot-go/internal/services/player"
	"github.com/stretchr/testify/assert"
)

func TestNewPlayer(t *testing.T) {
	tests := []struct {
		name           string
		playerName     string
		startingPoints int
		startingCount  int
	}{
		{
			name:           "create player with zero values",
			playerName:     "testplayer",
			startingPoints: 0,
			startingCount:  0,
		},
		{
			name:           "create player with positive values",
			playerName:     "veteran",
			startingPoints: 100,
			startingCount:  50,
		},
		{
			name:           "create player with high values",
			playerName:     "highlevel",
			startingPoints: 10000,
			startingCount:  5000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := player.NewPlayer(tt.playerName, tt.startingPoints, tt.startingCount)

			assert.NotNil(t, p)
			assert.Equal(t, tt.playerName, p.Name)
			assert.Equal(t, tt.startingPoints, p.Points)
			assert.Equal(t, tt.startingCount, p.Count)
		})
	}
}

func TestPlayer_String(t *testing.T) {
	tests := []struct {
		name     string
		player   *player.Player
		expected string
	}{
		{
			name:     "player with zero points",
			player:   player.NewPlayer("alice", 0, 0),
			expected: "alice has 0 points.",
		},
		{
			name:     "player with some points",
			player:   player.NewPlayer("bob", 150, 10),
			expected: "bob has 150 points.",
		},
		{
			name:     "player with many points",
			player:   player.NewPlayer("charlie", 9999, 500),
			expected: "charlie has 9999 points.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.player.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPlayer_GetPlayerLevel(t *testing.T) {
	tests := []struct {
		name     string
		count    int
		expected string
	}{
		// Beginner level (0-9)
		{name: "beginner at 0", count: 0, expected: "Beginner 🐣"},
		{name: "beginner at 5", count: 5, expected: "Beginner 🐣"},
		{name: "beginner at 9", count: 9, expected: "Beginner 🐣"},

		// Initiate level (10-100)
		{name: "initiate at 10", count: 10, expected: "Initiate 🐦"},
		{name: "initiate at 50", count: 50, expected: "Initiate 🐦"},
		{name: "initiate at 100", count: 100, expected: "Initiate 🐦"},

		// Adept level (101-199)
		{name: "adept at 101", count: 101, expected: "Adept 🦅"},
		{name: "adept at 150", count: 150, expected: "Adept 🦅"},
		{name: "adept at 199", count: 199, expected: "Adept 🦅"},

		// Expert level (200-499)
		{name: "expert at 200", count: 200, expected: "Expert 🕊️"},
		{name: "expert at 350", count: 350, expected: "Expert 🕊️"},
		{name: "expert at 499", count: 499, expected: "Expert 🕊️"},

		// Master level (500-799)
		{name: "master at 500", count: 500, expected: "Master 🦜"},
		{name: "master at 650", count: 650, expected: "Master 🦜"},
		{name: "master at 799", count: 799, expected: "Master 🦜"},

		// Grandmaster level (800-999)
		{name: "grandmaster at 800", count: 800, expected: "Grandmaster 🐔"},
		{name: "grandmaster at 900", count: 900, expected: "Grandmaster 🐔"},
		{name: "grandmaster at 999", count: 999, expected: "Grandmaster 🐔"},

		// Legendary Phoenix level (1000-2999)
		{name: "legendary phoenix at 1000", count: 1000, expected: "Legendary Phoenix 🐉🔥"},
		{name: "legendary phoenix at 2000", count: 2000, expected: "Legendary Phoenix 🐉🔥"},
		{name: "legendary phoenix at 2999", count: 2999, expected: "Legendary Phoenix 🐉🔥"},

		// Mythic Dragon level (3000-4999)
		{name: "mythic dragon at 3000", count: 3000, expected: "Mythic Dragon 🐲✨"},
		{name: "mythic dragon at 4000", count: 4000, expected: "Mythic Dragon 🐲✨"},
		{name: "mythic dragon at 4999", count: 4999, expected: "Mythic Dragon 🐲✨"},

		// Cosmic Falcon level (5000-9999)
		{name: "cosmic falcon at 5000", count: 5000, expected: "Cosmic Falcon 🌌🦅"},
		{name: "cosmic falcon at 7500", count: 7500, expected: "Cosmic Falcon 🌌🦅"},
		{name: "cosmic falcon at 9999", count: 9999, expected: "Cosmic Falcon 🌌🦅"},

		// Lord of Pigeons level (10000+)
		{name: "lord of pigeons at 10000", count: 10000, expected: "Lord of Pigeons 👑🐦"},
		{name: "lord of pigeons at 50000", count: 50000, expected: "Lord of Pigeons 👑🐦"},
		{name: "lord of pigeons at 100000", count: 100000, expected: "Lord of Pigeons 👑🐦"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := player.NewPlayer("testplayer", 0, tt.count)
			result := p.GetPlayerLevel()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStartingPoints(t *testing.T) {
	assert.Equal(t, 0, player.StartingPoints)
}
