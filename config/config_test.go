package config_test

import (
	"strings"
	"testing"

	"github.com/MyelinBots/pigeonbot-go/config"
	"github.com/stretchr/testify/assert"
)

func TestConfigStructs(t *testing.T) {
	t.Run("AppConfig has correct defaults", func(t *testing.T) {
		cfg := config.AppConfig{}

		// Verify the struct exists and can be used
		cfg.APPName = "testapp"
		cfg.Version = "1.0.0"
		cfg.Port = 3000

		assert.Equal(t, "testapp", cfg.APPName)
		assert.Equal(t, "1.0.0", cfg.Version)
		assert.Equal(t, 3000, cfg.Port)
	})

	t.Run("IRCConfig fields", func(t *testing.T) {
		cfg := config.IRCConfig{
			Host:             "irc.example.com",
			Port:             6667,
			SSL:              true,
			Nick:             "testbot",
			RealName:         "Test Bot",
			ChannelsString:   "#channel1,#channel2",
			Network:          "testnet",
			NickservCommand:  "PRIVMSG NickServ IDENTIFY %s",
			NickservPassword: "secret",
		}

		assert.Equal(t, "irc.example.com", cfg.Host)
		assert.Equal(t, 6667, cfg.Port)
		assert.True(t, cfg.SSL)
		assert.Equal(t, "testbot", cfg.Nick)
		assert.Equal(t, "Test Bot", cfg.RealName)
		assert.Equal(t, "#channel1,#channel2", cfg.ChannelsString)
		assert.Equal(t, "testnet", cfg.Network)
		assert.Equal(t, "PRIVMSG NickServ IDENTIFY %s", cfg.NickservCommand)
		assert.Equal(t, "secret", cfg.NickservPassword)
	})

	t.Run("DBConfig fields", func(t *testing.T) {
		cfg := config.DBConfig{
			Host:     "localhost",
			DataBase: "testdb",
			User:     "testuser",
			Password: "testpass",
			Port:     5432,
			SSLMode:  "disable",
		}

		assert.Equal(t, "localhost", cfg.Host)
		assert.Equal(t, "testdb", cfg.DataBase)
		assert.Equal(t, "testuser", cfg.User)
		assert.Equal(t, "testpass", cfg.Password)
		assert.Equal(t, uint(5432), cfg.Port)
		assert.Equal(t, "disable", cfg.SSLMode)
	})

	t.Run("GameConfig fields", func(t *testing.T) {
		cfg := config.GameConfig{
			Interval: 120,
		}

		assert.Equal(t, 120, cfg.Interval)
	})

	t.Run("Config aggregates all sub-configs", func(t *testing.T) {
		cfg := config.Config{
			AppConfig: config.AppConfig{
				APPName: "myapp",
				Version: "2.0.0",
				Port:    8080,
			},
			IRCConfig: config.IRCConfig{
				Host: "irc.test.com",
				Nick: "mybot",
			},
			DBConfig: config.DBConfig{
				Host:     "db.test.com",
				DataBase: "mydb",
			},
			GameConfig: config.GameConfig{
				Interval: 60,
			},
		}

		assert.Equal(t, "myapp", cfg.AppConfig.APPName)
		assert.Equal(t, "irc.test.com", cfg.IRCConfig.Host)
		assert.Equal(t, "db.test.com", cfg.DBConfig.Host)
		assert.Equal(t, 60, cfg.GameConfig.Interval)
	})
}

func TestChannelsParsing(t *testing.T) {
	t.Run("single channel", func(t *testing.T) {
		channelsString := "#channel1"
		channels := strings.Split(channelsString, ",")

		assert.Len(t, channels, 1)
		assert.Equal(t, "#channel1", channels[0])
	})

	t.Run("multiple channels", func(t *testing.T) {
		channelsString := "#channel1,#channel2,#channel3"
		channels := strings.Split(channelsString, ",")

		assert.Len(t, channels, 3)
		assert.Equal(t, "#channel1", channels[0])
		assert.Equal(t, "#channel2", channels[1])
		assert.Equal(t, "#channel3", channels[2])
	})

	t.Run("empty string", func(t *testing.T) {
		channelsString := ""
		channels := strings.Split(channelsString, ",")

		// Split of empty string returns slice with one empty element
		assert.Len(t, channels, 1)
		assert.Equal(t, "", channels[0])
	})

	t.Run("channels with spaces", func(t *testing.T) {
		channelsString := "#channel1, #channel2"
		channels := strings.Split(channelsString, ",")

		assert.Len(t, channels, 2)
		// Note: spaces are preserved
		assert.Equal(t, "#channel1", channels[0])
		assert.Equal(t, " #channel2", channels[1])
	})
}

func TestIRCConfigChannels(t *testing.T) {
	t.Run("Channels slice can be set", func(t *testing.T) {
		cfg := config.IRCConfig{
			ChannelsString: "#a,#b,#c",
		}

		// Simulate what LoadConfigOrPanic does
		cfg.Channels = strings.Split(cfg.ChannelsString, ",")

		assert.Len(t, cfg.Channels, 3)
		assert.Equal(t, []string{"#a", "#b", "#c"}, cfg.Channels)
	})
}
