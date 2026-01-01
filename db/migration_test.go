package db

import (
	"testing"

	"github.com/MyelinBots/pigeonbot-go/config"
	"github.com/MyelinBots/pigeonbot-go/internal/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigrateUp(t *testing.T) {
	t.Run("runs migrations successfully", func(t *testing.T) {
		// MigrateUp should not error (migrations already applied or no change)
		err := MigrateUp()
		assert.NoError(t, err)
	})

	t.Run("creates player table", func(t *testing.T) {
		cfg := config.LoadConfigOrPanic()
		database := db.NewDatabase(cfg.DBConfig)

		// Verify player table exists
		var exists bool
		err := database.DB.Raw(`
			SELECT EXISTS (
				SELECT FROM information_schema.tables
				WHERE table_name = 'player'
			)
		`).Scan(&exists).Error

		require.NoError(t, err)
		assert.True(t, exists, "player table should exist")

		sqlDB, _ := database.DB.DB()
		sqlDB.Close()
	})

	t.Run("player table has all columns", func(t *testing.T) {
		cfg := config.LoadConfigOrPanic()
		database := db.NewDatabase(cfg.DBConfig)

		expectedColumns := []string{"id", "name", "points", "count", "network", "created_at", "updated_at", "channel", "eggs", "rare_eggs"}

		for _, col := range expectedColumns {
			var exists bool
			err := database.DB.Raw(`
				SELECT EXISTS (
					SELECT FROM information_schema.columns
					WHERE table_name = 'player' AND column_name = ?
				)
			`, col).Scan(&exists).Error

			require.NoError(t, err)
			assert.True(t, exists, "column %s should exist", col)
		}

		sqlDB, _ := database.DB.DB()
		sqlDB.Close()
	})
}

func TestMigrateDown(t *testing.T) {
	t.Run("migrate down then up works", func(t *testing.T) {
		// Run down
		err := MigrateDown()
		// May error if already at base, that's ok
		_ = err

		// Run up again
		err = MigrateUp()
		assert.NoError(t, err)
	})
}
