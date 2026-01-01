package player_test

import (
	"testing"
	"time"

	"github.com/MyelinBots/pigeonbot-go/internal/db/repositories/player"
	"github.com/stretchr/testify/assert"
)

func TestPlayerEntity(t *testing.T) {
	t.Run("creates player with all fields", func(t *testing.T) {
		now := time.Now()
		p := player.Player{
			ID:        "test-uuid-123",
			Name:      "testplayer",
			Points:    100,
			Count:     50,
			Network:   "testnetwork",
			Channel:   "#testchannel",
			CreatedAt: now,
			UpdatedAt: now,
			Eggs:      10,
			RareEggs:  2,
		}

		assert.Equal(t, "test-uuid-123", p.ID)
		assert.Equal(t, "testplayer", p.Name)
		assert.Equal(t, 100, p.Points)
		assert.Equal(t, 50, p.Count)
		assert.Equal(t, "testnetwork", p.Network)
		assert.Equal(t, "#testchannel", p.Channel)
		assert.Equal(t, 10, p.Eggs)
		assert.Equal(t, 2, p.RareEggs)
	})

	t.Run("default values", func(t *testing.T) {
		p := player.Player{}

		assert.Equal(t, "", p.ID)
		assert.Equal(t, "", p.Name)
		assert.Equal(t, 0, p.Points)
		assert.Equal(t, 0, p.Count)
		assert.Equal(t, 0, p.Eggs)
		assert.Equal(t, 0, p.RareEggs)
	})
}

func TestPlayerTableName(t *testing.T) {
	p := player.Player{}

	tableName := p.TableName()

	assert.Equal(t, "player", tableName)
}

func TestPlayerEntity_FieldTypes(t *testing.T) {
	t.Run("ID is string", func(t *testing.T) {
		p := player.Player{ID: "uuid-string"}
		assert.IsType(t, "", p.ID)
	})

	t.Run("Points is int", func(t *testing.T) {
		p := player.Player{Points: 1000}
		assert.IsType(t, int(0), p.Points)
	})

	t.Run("Count is int", func(t *testing.T) {
		p := player.Player{Count: 500}
		assert.IsType(t, int(0), p.Count)
	})

	t.Run("Eggs is int", func(t *testing.T) {
		p := player.Player{Eggs: 25}
		assert.IsType(t, int(0), p.Eggs)
	})

	t.Run("RareEggs is int", func(t *testing.T) {
		p := player.Player{RareEggs: 5}
		assert.IsType(t, int(0), p.RareEggs)
	})

	t.Run("CreatedAt is time.Time", func(t *testing.T) {
		p := player.Player{CreatedAt: time.Now()}
		assert.IsType(t, time.Time{}, p.CreatedAt)
	})

	t.Run("UpdatedAt is time.Time", func(t *testing.T) {
		p := player.Player{UpdatedAt: time.Now()}
		assert.IsType(t, time.Time{}, p.UpdatedAt)
	})
}
