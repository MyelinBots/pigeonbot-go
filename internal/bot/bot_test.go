package bot

import (
	"testing"

	"github.com/MyelinBots/pigeonbot-go/config"
	"github.com/MyelinBots/pigeonbot-go/internal/services/commands"
	"github.com/MyelinBots/pigeonbot-go/internal/services/game"
	"github.com/stretchr/testify/assert"
)

func TestIdentified(t *testing.T) {
	t.Run("initial state is not identified", func(t *testing.T) {
		id := &Identified{
			identified: false,
		}
		assert.False(t, id.identified)
	})

	t.Run("can be set to identified", func(t *testing.T) {
		id := &Identified{
			identified: false,
		}
		id.Lock()
		id.identified = true
		id.Unlock()

		assert.True(t, id.identified)
	})
}

func TestGameInstances(t *testing.T) {
	t.Run("initializes empty maps", func(t *testing.T) {
		gi := &GameInstances{
			games:            make(map[string]*game.Game),
			commandInstances: make(map[string]commands.CommandController),
			GameStarted:      make(map[string]bool),
		}

		assert.NotNil(t, gi.games)
		assert.NotNil(t, gi.commandInstances)
		assert.NotNil(t, gi.GameStarted)
		assert.Empty(t, gi.games)
	})

	t.Run("can track game started state", func(t *testing.T) {
		gi := &GameInstances{
			GameStarted: make(map[string]bool),
		}

		gi.Lock()
		gi.GameStarted["#test"] = true
		gi.Unlock()

		assert.True(t, gi.GameStarted["#test"])
		assert.False(t, gi.GameStarted["#other"])
	})
}

func TestHandleNickserv(t *testing.T) {
	t.Run("does nothing when already identified", func(t *testing.T) {
		cfg := config.IRCConfig{
			NickservPassword: "secret",
			NickservCommand:  "PRIVMSG NickServ IDENTIFY %s",
		}
		id := &Identified{
			identified: true,
		}

		// Should not panic with nil connection since it won't try to send
		handleNickserv(cfg, id, nil)

		// Still identified
		assert.True(t, id.identified)
	})

	t.Run("does nothing when no password set", func(t *testing.T) {
		cfg := config.IRCConfig{
			NickservPassword: "",
			NickservCommand:  "PRIVMSG NickServ IDENTIFY %s",
		}
		id := &Identified{
			identified: false,
		}

		// Should not panic with nil connection since it won't try to send
		handleNickserv(cfg, id, nil)
	})
}
