package game

import (
	"fmt"
	"math/rand"
	"os"
	"sync"

	"pigeongo/actions"
	"pigeongo/config"
	"pigeongo/db"
	"pigeongo/pigeon"
	"pigeongo/player"
	"pigeongo/timer"
)

// Game struct encapsulates game state and functionality
type Game struct {
	config           *config.Config
	irc              IRC
	players          []*player.Player
	actions          []actions.Action
	activePigeon     *pigeon.Pigeon
	pigeons          []*pigeon.Pigeon
	db               *db.DB
	playerRepository *db.PlayerRepository
	mu               sync.Mutex
}

// IRC interface represents an IRC client
type IRC interface {
	Privmsg(channel, message string)
	Config() config.Config
}

// NewGame initializes and returns a new Game instance
func NewGame(irc IRC) *Game {
	gameConfig := config.NewConfig(map[string]interface{}{
		"interval": os.Getenv("PIGEON_INTERVAL"),
	})

	gameDB := db.NewDB()
	repo := db.NewPlayerRepository(gameDB)

	return &Game{
		config:           gameConfig,
		irc:              irc,
		players:          []*player.Player{},
		actions:          predefinedActions(),
		activePigeon:     nil,
		pigeons:          pigeon.PredefinedPigeons(),
		db:               gameDB,
		playerRepository: repo,
	}
}

// predefinedActions returns the predefined game actions
func predefinedActions() []actions.Action {
	return []actions.Action{
		actions.Action("stole", []string{"tv 📺", "wallet 💰👛", "food 🍔 🍕 🍪 🌮"}, "❗⚠️ A %s pigeon %s your %s - - - - - 🐦", 10),
		actions.Action("pooped", []string{"car 🚗", "head 👤", "laptop 💻"}, "❗⚠️ A %s pigeon %s on your %s - - - - - 🐦", 10),
		actions.Action("landed", []string{"balcony 🏠🌿", "head 👤", "car 🚗", "house 🏠", "swimming pool 🏖️", "bed 🛏️", "couch 🛋️", "laptop 💻"}, "❗⚠️ A %s pigeon has %s on your %s - - - - - 🐦", 10),
		actions.Action("mating", []string{"balcony 🏠🌿", "car 🚗", "bed 🛏️", "swimming pool 🏖️", "couch 🛋️", "laptop 💻"}, "❗⚠️ %s pigeons are %s at your %s - - - - - 🕊️ 💕 🕊️", 10),
	}
}

// Start begins the game's timer
func (g *Game) Start() {
	timer.NewRepeatedTimer(g.config.Interval(), g.ActOnPlayer)
	fmt.Println("Game started! Press Ctrl+C to stop.")
	select {} // Keep the game running
}

// ActOnPlayer triggers an action on a player
func (g *Game) ActOnPlayer() {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.activePigeon != nil {
		g.irc.Privmsg(g.irc.Config().Channel, fmt.Sprintf("🕊️ ~ coo coo ~ the %s pigeon has made a clean escape ~ 🕊️", g.activePigeon.Type()))
		g.activePigeon = nil
		return
	}

	if len(g.players) == 0 {
		return
	}

	randomPigeon := g.pigeons[rand.Intn(len(g.pigeons))]
	randomAction := g.actions[rand.Intn(len(g.actions))]

	g.activePigeon = randomPigeon
	g.irc.Privmsg(g.irc.Config().Channel, randomAction.Act(randomPigeon.Type()))
}

// AddPlayer adds a new player to the game
func (g *Game) AddPlayer(name string) {
	g.mu.Lock()
	defer g.mu.Unlock()

	for _, p := range g.players {
		if p.Name() == name {
			return
		}
	}

	newPlayer := player.NewPlayer(name, 0, 0)
	g.players = append(g.players, newPlayer)
	g.playerRepository.Upsert(name, 0, 0)
}

// FindPlayer finds a player by name
func (g *Game) FindPlayer(name string) *player.Player {
	g.mu.Lock()
	defer g.mu.Unlock()

	for _, p := range g.players {
		if p.Name() == name {
			return p
		}
	}

	g.AddPlayer(name)
	return g.FindPlayer(name)
}
