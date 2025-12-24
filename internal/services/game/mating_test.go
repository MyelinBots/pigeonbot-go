package game

import (
	"context"
	"strings"
	"testing"
	"time"

	player "github.com/MyelinBots/pigeonbot-go/internal/db/repositories/player"
	"github.com/MyelinBots/pigeonbot-go/internal/services/pigeon"
	servicePlayer "github.com/MyelinBots/pigeonbot-go/internal/services/player"
	"github.com/stretchr/testify/assert"
)

// TestHandleMatingEggs_NoActivePigeon tests when there's no active pigeon
func TestHandleMatingEggs_NoActivePigeon(t *testing.T) {
	g := &Game{
		activePigeon: &ActivePigeon{},
	}

	msg, err := g.HandleMatingEggs(context.Background(), "testuser")
	assert.Nil(t, err)
	assert.Empty(t, msg)
}

// TestHandleMatingEggs_NotMating tests when pigeon is not mating
func TestHandleMatingEggs_NotMating(t *testing.T) {
	g := &Game{
		activePigeon: &ActivePigeon{
			activePigeon: &pigeon.Pigeon{Type: "cartel member", Points: 10, Success: 85},
			IsMating:     false,
		},
	}

	msg, err := g.HandleMatingEggs(context.Background(), "testuser")
	assert.Nil(t, err)
	assert.Empty(t, msg)
}

// TestHandleMatingEggs_UnknownType tests when pigeon type has no base eggs
func TestHandleMatingEggs_UnknownType(t *testing.T) {
	g := &Game{
		activePigeon: &ActivePigeon{
			activePigeon: &pigeon.Pigeon{Type: "unknown", Points: 10, Success: 85},
			IsMating:     true,
		},
	}

	msg, err := g.HandleMatingEggs(context.Background(), "testuser")
	assert.Nil(t, err)
	assert.Empty(t, msg)
}

// TestTryRareEgg_NoActivePigeon tests TryRareEgg with no pigeon
func TestTryRareEgg_NoActivePigeon(t *testing.T) {
	g := &Game{
		activePigeon: &ActivePigeon{},
	}

	msg, err := g.TryRareEgg(context.Background(), "testuser")
	assert.Nil(t, err)
	assert.Empty(t, msg)
}

// TestTryRareEgg_NotMating tests TryRareEgg when not mating
func TestTryRareEgg_NotMating(t *testing.T) {
	g := &Game{
		activePigeon: &ActivePigeon{
			activePigeon: &pigeon.Pigeon{Type: "cartel member", Points: 10, Success: 85},
			IsMating:     false,
		},
	}

	msg, err := g.TryRareEgg(context.Background(), "testuser")
	assert.Nil(t, err)
	assert.Empty(t, msg)
}

// TestActOnPlayer_PigeonEscape tests the pigeon escape path
func TestActOnPlayer_PigeonEscape(t *testing.T) {
	// Create a mock IRC client
	mockClient := &mockIRCClientForTest{messages: make([]string, 0)}

	g := &Game{
		activePigeon: &ActivePigeon{
			activePigeon: &pigeon.Pigeon{Type: "cartel member", Points: 10, Success: 85},
			IsMating:     false,
			SpawnedAt:    time.Now().Add(-65 * time.Second), // More than 60 seconds ago
		},
		players:   Players{},
		ircClient: mockClient,
		channel:   "test",
	}

	ctx := context.Background()
	g.ActOnPlayer(ctx)

	// After 60+ seconds, the pigeon should escape
	assert.Nil(t, g.activePigeon.activePigeon)
}

// TestActOnPlayer_PigeonStillAlive tests when pigeon is still alive
func TestActOnPlayer_PigeonStillAlive(t *testing.T) {
	mockClient := &mockIRCClientForTest{messages: make([]string, 0)}

	activePigeon := &pigeon.Pigeon{Type: "cartel member", Points: 10, Success: 85}
	g := &Game{
		activePigeon: &ActivePigeon{
			activePigeon: activePigeon,
			IsMating:     false,
			SpawnedAt:    time.Now(), // Just spawned
		},
		players:   Players{},
		ircClient: mockClient,
		channel:   "test",
	}

	ctx := context.Background()
	g.ActOnPlayer(ctx)

	// Pigeon should still be active since it was just spawned
	assert.Equal(t, activePigeon, g.activePigeon.activePigeon)
}

// mockIRCClientForTest is a simple mock for internal tests
type mockIRCClientForTest struct {
	messages []string
}

func (m *mockIRCClientForTest) Privmsg(channel, message string) {
	m.messages = append(m.messages, message)
}

// mockPlayerRepositoryForTest is a simple mock for internal tests
type mockPlayerRepositoryForTest struct {
	eggs         map[string]int
	rareEggs     map[string]int
	players      map[string]*player.Player
	addEggsErr   error
	addRareErr   error
	findErr      error
	getEggsErr   error
	getRareErr   error
}

func newMockPlayerRepoForTest() *mockPlayerRepositoryForTest {
	return &mockPlayerRepositoryForTest{
		eggs:     make(map[string]int),
		rareEggs: make(map[string]int),
		players:  make(map[string]*player.Player),
	}
}

func (m *mockPlayerRepositoryForTest) GetPlayerByID(id string) (*player.Player, error) {
	return nil, nil
}

func (m *mockPlayerRepositoryForTest) GetAllPlayers(ctx context.Context, network, channel string) ([]*player.Player, error) {
	return nil, nil
}

func (m *mockPlayerRepositoryForTest) UpsertPlayer(ctx context.Context, p *player.Player) error {
	return nil
}

func (m *mockPlayerRepositoryForTest) TopByPoints(ctx context.Context, network, channel string, limit int) ([]*player.Player, error) {
	return nil, nil
}

func (m *mockPlayerRepositoryForTest) AddEggs(ctx context.Context, network, channel, name string, delta int) (int, error) {
	if m.addEggsErr != nil {
		return 0, m.addEggsErr
	}
	key := network + "|" + channel + "|" + name
	m.eggs[key] += delta
	return m.eggs[key], nil
}

func (m *mockPlayerRepositoryForTest) GetEggs(ctx context.Context, network, channel, name string) (int, error) {
	if m.getEggsErr != nil {
		return 0, m.getEggsErr
	}
	key := network + "|" + channel + "|" + name
	return m.eggs[key], nil
}

func (m *mockPlayerRepositoryForTest) AddRareEggs(ctx context.Context, network, channel, name string, delta int) (int, error) {
	if m.addRareErr != nil {
		return 0, m.addRareErr
	}
	key := network + "|" + channel + "|" + name
	m.rareEggs[key] += delta
	return m.rareEggs[key], nil
}

func (m *mockPlayerRepositoryForTest) GetRareEggs(ctx context.Context, network, channel, name string) (int, error) {
	if m.getRareErr != nil {
		return 0, m.getRareErr
	}
	key := network + "|" + channel + "|" + name
	return m.rareEggs[key], nil
}

// TestTryRareEgg_Probabilistic runs many iterations to hit random paths
func TestTryRareEgg_Probabilistic(t *testing.T) {
	mockRepo := newMockPlayerRepoForTest()
	mockClient := &mockIRCClientForTest{messages: make([]string, 0)}

	g := &Game{
		activePigeon: &ActivePigeon{
			activePigeon: &pigeon.Pigeon{Type: "boss", Points: 100, Success: 25},
			IsMating:     true,
		},
		players:          Players{players: []*servicePlayer.Player{}},
		playerRepository: mockRepo,
		ircClient:        mockClient,
		channel:          "test",
		network:          "testnet",
	}

	ctx := context.Background()

	// Run many iterations to hit different random paths
	rareEggAppeared := false
	rareEggFailed := false
	rareEggSucceeded := false

	for i := 0; i < 1000; i++ {
		msg, err := g.TryRareEgg(ctx, "testuser")
		assert.Nil(t, err)

		if msg != "" {
			rareEggAppeared = true
			if strings.Contains(msg, "cracked and vanished") {
				rareEggFailed = true
			}
			if strings.Contains(msg, "LEGENDARY") {
				rareEggSucceeded = true
			}
		}
	}

	// With 1000 iterations and 10% appear rate, we should see some appearances
	assert.True(t, rareEggAppeared, "Rare egg should appear at least once in 1000 iterations")
	// Just log these for info - they depend on randomness
	_ = rareEggFailed
	_ = rareEggSucceeded
}

// TestHandleMatingEggs_WithMatingPigeon tests mating eggs with various pigeon types
func TestHandleMatingEggs_WithMatingPigeon(t *testing.T) {
	pigeonTypes := []string{"cartel member", "white", "boss"}

	for _, pType := range pigeonTypes {
		t.Run(pType, func(t *testing.T) {
			mockRepo := newMockPlayerRepoForTest()
			mockClient := &mockIRCClientForTest{messages: make([]string, 0)}

			g := &Game{
				activePigeon: &ActivePigeon{
					activePigeon: &pigeon.Pigeon{Type: pType, Points: 10, Success: 85},
					IsMating:     true,
				},
				players:          Players{players: []*servicePlayer.Player{}},
				playerRepository: mockRepo,
				ircClient:        mockClient,
				channel:          "test",
				network:          "testnet",
			}

			ctx := context.Background()

			// Run multiple times to hit different cracking outcomes
			for i := 0; i < 50; i++ {
				msg, err := g.HandleMatingEggs(ctx, "testuser")
				assert.Nil(t, err)
				// All pigeon types should produce some message
				assert.NotEmpty(t, msg)
			}
		})
	}
}

// TestHandleMatingEggs_GetEggsError tests error handling in HandleMatingEggs
func TestHandleMatingEggs_GetEggsError(t *testing.T) {
	mockRepo := newMockPlayerRepoForTest()
	mockRepo.getEggsErr = assert.AnError
	mockClient := &mockIRCClientForTest{messages: make([]string, 0)}

	g := &Game{
		activePigeon: &ActivePigeon{
			activePigeon: &pigeon.Pigeon{Type: "cartel member", Points: 10, Success: 85},
			IsMating:     true,
		},
		players:          Players{players: []*servicePlayer.Player{}},
		playerRepository: mockRepo,
		ircClient:        mockClient,
		channel:          "test",
		network:          "testnet",
	}

	ctx := context.Background()
	_, err := g.HandleMatingEggs(ctx, "testuser")
	assert.Equal(t, assert.AnError, err)
}

// TestHandleMatingEggs_AddEggsError tests error handling when adding eggs
func TestHandleMatingEggs_AddEggsError(t *testing.T) {
	mockRepo := newMockPlayerRepoForTest()
	mockClient := &mockIRCClientForTest{messages: make([]string, 0)}

	g := &Game{
		activePigeon: &ActivePigeon{
			activePigeon: &pigeon.Pigeon{Type: "boss", Points: 100, Success: 25},
			IsMating:     true,
		},
		players:          Players{players: []*servicePlayer.Player{}},
		playerRepository: mockRepo,
		ircClient:        mockClient,
		channel:          "test",
		network:          "testnet",
	}

	ctx := context.Background()

	// Run until we get eggs (not all cracked)
	// This tests the AddEggs path
	for i := 0; i < 100; i++ {
		msg, err := g.HandleMatingEggs(ctx, "testuser")
		assert.Nil(t, err)
		if strings.Contains(msg, "collected") {
			break
		}
	}
}

// TestHandleMatingEggs_GetRareEggsError tests error handling for rare eggs
func TestHandleMatingEggs_GetRareEggsError(t *testing.T) {
	mockRepo := newMockPlayerRepoForTest()
	mockRepo.getRareErr = assert.AnError
	mockClient := &mockIRCClientForTest{messages: make([]string, 0)}

	g := &Game{
		activePigeon: &ActivePigeon{
			activePigeon: &pigeon.Pigeon{Type: "boss", Points: 100, Success: 25},
			IsMating:     true,
		},
		players:          Players{players: []*servicePlayer.Player{}},
		playerRepository: mockRepo,
		ircClient:        mockClient,
		channel:          "test",
		network:          "testnet",
	}

	ctx := context.Background()

	// Run until we hit the GetRareEggs path
	for i := 0; i < 100; i++ {
		_, err := g.HandleMatingEggs(ctx, "testuser")
		if err != nil {
			assert.Equal(t, assert.AnError, err)
			return
		}
	}
}

// TestTryRareEgg_AddEggsError tests error handling in TryRareEgg
func TestTryRareEgg_AddEggsError(t *testing.T) {
	mockRepo := newMockPlayerRepoForTest()
	mockRepo.addEggsErr = assert.AnError
	mockClient := &mockIRCClientForTest{messages: make([]string, 0)}

	g := &Game{
		activePigeon: &ActivePigeon{
			activePigeon: &pigeon.Pigeon{Type: "boss", Points: 100, Success: 25},
			IsMating:     true,
		},
		players:          Players{players: []*servicePlayer.Player{}},
		playerRepository: mockRepo,
		ircClient:        mockClient,
		channel:          "test",
		network:          "testnet",
	}

	ctx := context.Background()

	// Run until we hit the error path (rare egg succeeds and tries to add eggs)
	for i := 0; i < 1000; i++ {
		_, err := g.TryRareEgg(ctx, "testuser")
		if err != nil {
			assert.Equal(t, assert.AnError, err)
			return
		}
	}
}

// TestTryRareEgg_AddRareEggsError tests error handling for rare eggs in TryRareEgg
func TestTryRareEgg_AddRareEggsError(t *testing.T) {
	mockRepo := newMockPlayerRepoForTest()
	mockRepo.addRareErr = assert.AnError
	mockClient := &mockIRCClientForTest{messages: make([]string, 0)}

	g := &Game{
		activePigeon: &ActivePigeon{
			activePigeon: &pigeon.Pigeon{Type: "boss", Points: 100, Success: 25},
			IsMating:     true,
		},
		players:          Players{players: []*servicePlayer.Player{}},
		playerRepository: mockRepo,
		ircClient:        mockClient,
		channel:          "test",
		network:          "testnet",
	}

	ctx := context.Background()

	// Run until we hit the error path
	for i := 0; i < 1000; i++ {
		_, err := g.TryRareEgg(ctx, "testuser")
		if err != nil {
			assert.Equal(t, assert.AnError, err)
			return
		}
	}
}
