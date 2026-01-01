package player

import (
	"context"
	"testing"

	"github.com/MyelinBots/pigeonbot-go/config"
	"github.com/MyelinBots/pigeonbot-go/internal/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) (*db.DB, func()) {
	t.Helper()

	cfg := config.LoadConfigOrPanic()
	database := db.NewDatabase(cfg.DBConfig)

	// Get underlying sql.DB to close later
	sqlDB, err := database.DB.DB()
	require.NoError(t, err)

	// Cleanup function
	cleanup := func() {
		// Truncate the player table after each test
		database.DB.Exec("TRUNCATE TABLE player RESTART IDENTITY CASCADE")
		sqlDB.Close()
	}

	// Truncate before test
	database.DB.Exec("TRUNCATE TABLE player RESTART IDENTITY CASCADE")

	return database, cleanup
}

func TestPlayerRepository_UpsertPlayer(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewPlayerRepository(database)
	ctx := context.Background()

	t.Run("create new player", func(t *testing.T) {
		player := &Player{
			Name:    "testuser",
			Network: "testnet",
			Channel: "#testchan",
			Points:  100,
			Count:   5,
		}

		err := repo.UpsertPlayer(ctx, player)
		require.NoError(t, err)
		assert.NotEmpty(t, player.ID)

		// Verify player was created
		players, err := repo.GetAllPlayers(ctx, "testnet", "#testchan")
		require.NoError(t, err)
		assert.Len(t, players, 1)
		assert.Equal(t, "testuser", players[0].Name)
		assert.Equal(t, 100, players[0].Points)
	})

	t.Run("update existing player", func(t *testing.T) {
		// First create
		player := &Player{
			Name:    "updateuser",
			Network: "testnet",
			Channel: "#testchan",
			Points:  50,
			Count:   2,
		}
		err := repo.UpsertPlayer(ctx, player)
		require.NoError(t, err)

		// Then update
		player.Points = 150
		player.Count = 10
		err = repo.UpsertPlayer(ctx, player)
		require.NoError(t, err)

		// Verify update
		players, err := repo.GetAllPlayers(ctx, "testnet", "#testchan")
		require.NoError(t, err)

		var found *Player
		for _, p := range players {
			if p.Name == "updateuser" {
				found = p
				break
			}
		}
		require.NotNil(t, found)
		assert.Equal(t, 150, found.Points)
		assert.Equal(t, 10, found.Count)
	})
}

func TestPlayerRepository_GetAllPlayers(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewPlayerRepository(database)
	ctx := context.Background()

	t.Run("empty result", func(t *testing.T) {
		players, err := repo.GetAllPlayers(ctx, "emptynet", "#emptychan")
		require.NoError(t, err)
		assert.Empty(t, players)
	})

	t.Run("returns players from correct network and channel", func(t *testing.T) {
		// Create players in different networks/channels
		players := []*Player{
			{Name: "user1", Network: "net1", Channel: "#chan1", Points: 10},
			{Name: "user2", Network: "net1", Channel: "#chan1", Points: 20},
			{Name: "user3", Network: "net1", Channel: "#chan2", Points: 30}, // different channel
			{Name: "user4", Network: "net2", Channel: "#chan1", Points: 40}, // different network
		}
		for _, p := range players {
			err := repo.UpsertPlayer(ctx, p)
			require.NoError(t, err)
		}

		// Get players from net1/#chan1
		result, err := repo.GetAllPlayers(ctx, "net1", "#chan1")
		require.NoError(t, err)
		assert.Len(t, result, 2)

		names := make([]string, len(result))
		for i, p := range result {
			names[i] = p.Name
		}
		assert.Contains(t, names, "user1")
		assert.Contains(t, names, "user2")
	})
}

func TestPlayerRepository_TopByPoints(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewPlayerRepository(database)
	ctx := context.Background()

	// Create test players
	players := []*Player{
		{Name: "low", Network: "testnet", Channel: "#test", Points: 10, Count: 1},
		{Name: "high", Network: "testnet", Channel: "#test", Points: 100, Count: 5},
		{Name: "mid", Network: "testnet", Channel: "#test", Points: 50, Count: 3},
		{Name: "highest", Network: "testnet", Channel: "#test", Points: 200, Count: 10},
		{Name: "other", Network: "othernet", Channel: "#test", Points: 500, Count: 20}, // different network
	}
	for _, p := range players {
		err := repo.UpsertPlayer(ctx, p)
		require.NoError(t, err)
	}

	t.Run("returns top N by points", func(t *testing.T) {
		result, err := repo.TopByPoints(ctx, "testnet", "#test", 3)
		require.NoError(t, err)
		assert.Len(t, result, 3)
		assert.Equal(t, "highest", result[0].Name)
		assert.Equal(t, "high", result[1].Name)
		assert.Equal(t, "mid", result[2].Name)
	})

	t.Run("respects limit", func(t *testing.T) {
		result, err := repo.TopByPoints(ctx, "testnet", "#test", 2)
		require.NoError(t, err)
		assert.Len(t, result, 2)
	})

	t.Run("defaults to 5 when limit is 0", func(t *testing.T) {
		result, err := repo.TopByPoints(ctx, "testnet", "#test", 0)
		require.NoError(t, err)
		assert.Len(t, result, 4) // only 4 players in testnet/#test
	})

	t.Run("caps at 50", func(t *testing.T) {
		result, err := repo.TopByPoints(ctx, "testnet", "#test", 100)
		require.NoError(t, err)
		assert.Len(t, result, 4) // only 4 players, but limit would be capped at 50
	})
}

func TestPlayerRepository_Eggs(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewPlayerRepository(database)
	ctx := context.Background()

	t.Run("GetEggs returns 0 for non-existent player", func(t *testing.T) {
		eggs, err := repo.GetEggs(ctx, "testnet", "#test", "nonexistent")
		require.NoError(t, err)
		assert.Equal(t, 0, eggs)
	})

	t.Run("AddEggs creates player if not exists", func(t *testing.T) {
		eggs, err := repo.AddEggs(ctx, "testnet", "#test", "newplayer", 5)
		require.NoError(t, err)
		assert.Equal(t, 5, eggs)

		// Verify
		eggs, err = repo.GetEggs(ctx, "testnet", "#test", "newplayer")
		require.NoError(t, err)
		assert.Equal(t, 5, eggs)
	})

	t.Run("AddEggs increments existing eggs", func(t *testing.T) {
		// Create player with initial eggs
		player := &Player{
			Name:    "eggcollector",
			Network: "testnet",
			Channel: "#test",
			Eggs:    10,
		}
		err := repo.UpsertPlayer(ctx, player)
		require.NoError(t, err)

		// Add more eggs
		eggs, err := repo.AddEggs(ctx, "testnet", "#test", "eggcollector", 7)
		require.NoError(t, err)
		assert.Equal(t, 17, eggs)
	})

	t.Run("AddEggs with 0 delta returns current eggs", func(t *testing.T) {
		player := &Player{
			Name:    "zeroadd",
			Network: "testnet",
			Channel: "#test",
			Eggs:    25,
		}
		err := repo.UpsertPlayer(ctx, player)
		require.NoError(t, err)

		eggs, err := repo.AddEggs(ctx, "testnet", "#test", "zeroadd", 0)
		require.NoError(t, err)
		assert.Equal(t, 25, eggs)
	})
}

func TestPlayerRepository_RareEggs(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewPlayerRepository(database)
	ctx := context.Background()

	t.Run("GetRareEggs returns 0 for non-existent player", func(t *testing.T) {
		eggs, err := repo.GetRareEggs(ctx, "testnet", "#test", "nonexistent")
		require.NoError(t, err)
		assert.Equal(t, 0, eggs)
	})

	t.Run("AddRareEggs increments existing rare eggs", func(t *testing.T) {
		// First create a player
		player := &Player{
			Name:     "rarecollector",
			Network:  "testnet",
			Channel:  "#test",
			RareEggs: 3,
		}
		err := repo.UpsertPlayer(ctx, player)
		require.NoError(t, err)

		// Add rare eggs
		eggs, err := repo.AddRareEggs(ctx, "testnet", "#test", "rarecollector", 2)
		require.NoError(t, err)
		assert.Equal(t, 5, eggs)
	})

	t.Run("AddRareEggs with 0 delta returns current eggs", func(t *testing.T) {
		player := &Player{
			Name:     "rarezeroadd",
			Network:  "testnet",
			Channel:  "#test",
			RareEggs: 7,
		}
		err := repo.UpsertPlayer(ctx, player)
		require.NoError(t, err)

		eggs, err := repo.AddRareEggs(ctx, "testnet", "#test", "rarezeroadd", 0)
		require.NoError(t, err)
		assert.Equal(t, 7, eggs)
	})
}

func TestPlayerRepository_GetPlayerByID(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewPlayerRepository(database)

	// Current implementation returns nil, nil
	player, err := repo.GetPlayerByID("someid")
	assert.NoError(t, err)
	assert.Nil(t, player)
}

func TestPlayerRepository_TransactionRollback(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("transaction rollback undoes changes", func(t *testing.T) {
		// Start a transaction
		tx := database.DB.Begin()
		require.NoError(t, tx.Error)

		txRepo := NewPlayerRepository(&db.DB{DB: tx})

		// Create a player in transaction
		player := &Player{
			Name:    "txuser",
			Network: "testnet",
			Channel: "#test",
			Points:  100,
		}
		err := txRepo.UpsertPlayer(ctx, player)
		require.NoError(t, err)

		// Verify player exists in transaction
		players, err := txRepo.GetAllPlayers(ctx, "testnet", "#test")
		require.NoError(t, err)
		assert.Len(t, players, 1)

		// Rollback
		tx.Rollback()

		// Create new repo with original DB
		repo := NewPlayerRepository(database)

		// Verify player does not exist after rollback
		players, err = repo.GetAllPlayers(ctx, "testnet", "#test")
		require.NoError(t, err)
		assert.Empty(t, players)
	})

	t.Run("transaction commit persists changes", func(t *testing.T) {
		// Start a transaction
		tx := database.DB.Begin()
		require.NoError(t, tx.Error)

		txRepo := NewPlayerRepository(&db.DB{DB: tx})

		// Create a player in transaction
		player := &Player{
			Name:    "commituser",
			Network: "testnet",
			Channel: "#test",
			Points:  200,
		}
		err := txRepo.UpsertPlayer(ctx, player)
		require.NoError(t, err)

		// Commit
		tx.Commit()

		// Create new repo with original DB
		repo := NewPlayerRepository(database)

		// Verify player exists after commit
		players, err := repo.GetAllPlayers(ctx, "testnet", "#test")
		require.NoError(t, err)
		assert.Len(t, players, 1)
		assert.Equal(t, "commituser", players[0].Name)
	})
}

// TestWithTransaction demonstrates running tests in a transaction that gets rolled back
func TestWithTransaction(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	runInTransaction := func(t *testing.T, fn func(repo PlayerRepository)) {
		tx := database.DB.Begin()
		require.NoError(t, tx.Error)
		defer tx.Rollback()

		repo := NewPlayerRepository(&db.DB{DB: tx})
		fn(repo)
	}

	t.Run("test 1 - isolated", func(t *testing.T) {
		runInTransaction(t, func(repo PlayerRepository) {
			ctx := context.Background()
			player := &Player{
				Name:    "isolated1",
				Network: "testnet",
				Channel: "#test",
				Points:  100,
			}
			err := repo.UpsertPlayer(ctx, player)
			require.NoError(t, err)

			players, err := repo.GetAllPlayers(ctx, "testnet", "#test")
			require.NoError(t, err)
			assert.Len(t, players, 1)
		})
	})

	t.Run("test 2 - isolated", func(t *testing.T) {
		runInTransaction(t, func(repo PlayerRepository) {
			ctx := context.Background()
			// This should start with empty table due to rollback
			players, err := repo.GetAllPlayers(ctx, "testnet", "#test")
			require.NoError(t, err)
			assert.Empty(t, players)

			player := &Player{
				Name:    "isolated2",
				Network: "testnet",
				Channel: "#test",
				Points:  200,
			}
			err = repo.UpsertPlayer(ctx, player)
			require.NoError(t, err)
		})
	})
}

// Helper to run a function within a transaction for test isolation
func withTx(t *testing.T, gormDB *gorm.DB, fn func(tx *gorm.DB)) {
	tx := gormDB.Begin()
	require.NoError(t, tx.Error)
	defer tx.Rollback()
	fn(tx)
}
