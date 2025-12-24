package db

import (
	"testing"

	"github.com/MyelinBots/pigeonbot-go/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDatabase(t *testing.T) {
	t.Run("connects to database successfully", func(t *testing.T) {
		cfg := config.LoadConfigOrPanic()
		database := NewDatabase(cfg.DBConfig)

		require.NotNil(t, database)
		require.NotNil(t, database.DB)

		// Verify connection works
		sqlDB, err := database.DB.DB()
		require.NoError(t, err)

		err = sqlDB.Ping()
		assert.NoError(t, err)

		sqlDB.Close()
	})

	t.Run("database struct has gorm.DB", func(t *testing.T) {
		cfg := config.LoadConfigOrPanic()
		database := NewDatabase(cfg.DBConfig)

		// Verify we can use GORM methods
		var result int
		err := database.DB.Raw("SELECT 1").Scan(&result).Error
		require.NoError(t, err)
		assert.Equal(t, 1, result)

		sqlDB, _ := database.DB.DB()
		sqlDB.Close()
	})
}

func TestNewDatabase_InvalidConfig(t *testing.T) {
	t.Run("panics on invalid connection", func(t *testing.T) {
		invalidCfg := config.DBConfig{
			Host:     "invalidhost",
			Port:     9999,
			User:     "invaliduser",
			Password: "invalidpass",
			DataBase: "invaliddb",
			SSLMode:  "disable",
		}

		assert.Panics(t, func() {
			NewDatabase(invalidCfg)
		})
	})
}
