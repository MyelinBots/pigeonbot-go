//go:generate mockgen -destination=mocks/mock_irc_Client.go -package=mocks github.com/MyelinBots/pigeonbot-go/internal/services/game IRCClient
package game

import (
	"context"
	"fmt"
	rand "math/rand/v2"
	"sort"
	"sync"
	"time"

	"github.com/MyelinBots/pigeonbot-go/config"
	player2 "github.com/MyelinBots/pigeonbot-go/internal/db/repositories/player"
	"github.com/MyelinBots/pigeonbot-go/internal/services/actions"
	"github.com/MyelinBots/pigeonbot-go/internal/services/context_manager"
	"github.com/MyelinBots/pigeonbot-go/internal/services/pigeon"
	"github.com/MyelinBots/pigeonbot-go/internal/services/player"
)

type ActivePigeon struct {
	sync.Mutex
	activePigeon *pigeon.Pigeon
}

type Players struct {
	sync.Mutex
	players []*player.Player
}

type IRCClient interface {
	Privmsg(channel, message string)
}

// Game struct encapsulates game state and functionality
type Game struct {
	config           config.GameConfig
	players          Players
	ircClient        IRCClient
	actions          []actions.Action
	activePigeon     *ActivePigeon
	pigeons          []*pigeon.Pigeon
	playerRepository player2.PlayerRepository
	channel          string
	network          string
}

// NewGame initializes and returns a new Game instance
func NewGame(cfg config.GameConfig, client IRCClient, repo player2.PlayerRepository, network string, channel string) *Game {

	return &Game{
		config:           cfg,
		ircClient:        client,
		actions:          predefinedActions(),
		activePigeon:     &ActivePigeon{},
		pigeons:          pigeon.PredefinedPigeons(),
		playerRepository: repo,
		channel:          channel,
		network:          network,
	}
}

// predefinedActions returns the predefined game actions
func predefinedActions() []actions.Action {
	return []actions.Action{
		{"stole", []string{"tv 📺", "wallet 💰👛", "food 🍔 🍕 🍪 🌮"}, "❗⚠️ A %s pigeon %s your %s - - 🐦", 10},
		{"pooped", []string{"car 🚗", "head 👤", "laptop 💻"}, "❗⚠️ A %s pigeon %s on your %s - - 🐦", 10},
		{"landed", []string{"balcony 🏠🌿", "head 👤", "car 🚗", "house 🏠", "swimming pool 🏖️", "bed 🛏️", "couch 🛋️", "laptop 💻"}, "❗⚠️ A %s pigeon has %s on your %s - - 🐦", 10},
		{"mating", []string{"balcony 🏠🌿", "car 🚗", "bed 🛏️", "swimming pool 🏖️", "couch 🛋️", "laptop 💻"}, "❗⚠️ %s pigeons are %s at your %s - - 🕊️ 💕 🕊️", 10},
	}
}

// Start begins the game's timer
func (g *Game) Start(ctx context.Context) {
	g.syncPlayers(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			g.ActOnPlayer(ctx)
			timer := g.config.Interval
			if timer == 0 {
				timer = 10
			}
			<-time.After(time.Duration(timer) * time.Second)
		}
	}
}

func (g *Game) syncPlayers(ctx context.Context) {

	players, err := g.playerRepository.GetAllPlayers(ctx, g.network, g.channel)
	if err != nil {
		return
	}
	for _, p := range players {
		g.players.players = append(g.players.players, player.NewPlayer(p.Name, p.Points, p.Count))
	}

}

// ActOnPlayer triggers an action on a player
func (g *Game) ActOnPlayer(ctx context.Context) {
	g.players.Lock()
	g.activePigeon.Lock()
	defer g.activePigeon.Unlock()
	defer g.players.Unlock()
	fmt.Println("act on player")

	if g.activePigeon.activePigeon != nil {
		g.ircClient.Privmsg(g.channel, fmt.Sprintf("🕊️ ~ coo coo ~ the %s pigeon has made a clean escape ~ 🕊️", g.activePigeon.activePigeon.Type))
		g.activePigeon.activePigeon = nil
		return
	}

	randomPigeon := g.pigeons[rand.IntN(len(g.pigeons))]
	randomAction := g.actions[rand.IntN(len(g.actions))]

	g.activePigeon.activePigeon = randomPigeon
	fmt.Println(randomAction.Act(randomPigeon.Type))
	g.ircClient.Privmsg(g.channel, randomAction.Act(randomPigeon.Type))
}

// AddPlayer adds a new player to the game
func (g *Game) addPlayer(ctx context.Context, name string) (*player.Player, error) {
	g.players.Lock()
	defer g.players.Unlock()

	for _, p := range g.players.players {
		if p.Name == name {
			return p, nil
		}
	}

	newPlayer := player.NewPlayer(name, 0, 0)
	g.players.players = append(g.players.players, newPlayer)
	playerEntity := player2.Player{
		Count:   0,
		Points:  0,
		Name:    name,
		Channel: g.channel,
		Network: g.network,
	}
	err := g.playerRepository.UpsertPlayer(ctx, &playerEntity)
	if err != nil {
		return nil, err
	}

	return newPlayer, nil
}

// FindPlayer finds a player by name
func (g *Game) FindPlayer(ctx context.Context, name string) (*player.Player, error) {
	g.players.Lock()

	for _, p := range g.players.players {
		if p.Name == name {
			g.players.Unlock()
			return p, nil
		}
	}
	g.players.Unlock()

	return g.addPlayer(ctx, name)

}

func (g *Game) HandleShoot(ctx context.Context, args ...string) error {
	name := context_manager.GetNickContext(ctx)
	fmt.Printf("Handling shoot for player: %s\n", name)
	g.activePigeon.Lock()
	defer g.activePigeon.Unlock()

	if g.activePigeon.activePigeon == nil {
		g.ircClient.Privmsg(g.channel, fmt.Sprintf("❗⚠️ %s has shot a pigeon, but there are no pigeons to shoot! - - 🐦", name))
		return nil
	}

	foundPlayer, err := g.FindPlayer(ctx, name)
	if err != nil {
		g.ircClient.Privmsg(g.channel, fmt.Sprintf("❗⚠️ %s has shot a pigeon, but there was an error finding the player! - - 🐦", name))
		return err
	}
	//print("Random result: %s, success rate: %s" % (str(randomResult), str(self.active.success() / 100)))
	// calculate success using rand and active pigeon success rate, must return a 0 or 1, if 0, player loses points, if 1, player gains points

	// Generate a random number between 0 and 99
	randomValue := rand.IntN(100)

	// Generate 1 or 0 based on the success rate
	result := 0
	if randomValue < g.activePigeon.activePigeon.Success {
		result = 1
	}

	if result == 1 {
		// Success: Update player's count and points
		foundPlayer.Points += g.activePigeon.activePigeon.Points
		foundPlayer.Count += 1

		// Determine player's level
		level := foundPlayer.GetPlayerLevel()

		// Inform the player of their success and current level
		g.ircClient.Privmsg(g.channel, fmt.Sprintf("❗⚠️ %s has shot a pigeon! - -  🐦 🔫 You are a murderer! . .  You have shot a total of %d pigeon(s)! . . 🐦 🕊️ . . You now have a total of %d points and reached the level: %s ", name, foundPlayer.Count, foundPlayer.Points, level))

		// Remove the pigeon from activePigeon
		g.activePigeon.activePigeon = nil
	} else {
		// Failure: Inform the player
		g.ircClient.Privmsg(g.channel, fmt.Sprintf("❗⚠️ %s has shot a pigeon, but it got away! - - - - - 🐦", name))
	}

	err = g.SavePlayers(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (g *Game) SavePlayers(ctx context.Context) error {
	g.players.Lock()
	defer g.players.Unlock()
	for _, p := range g.players.players {
		playerEntity := player2.Player{
			Count:   p.Count,
			Points:  p.Points,
			Name:    p.Name,
			Channel: g.channel,
			Network: g.network,
		}
		err := g.playerRepository.UpsertPlayer(ctx, &playerEntity)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}
	return nil
}

func (g *Game) HandlePoints(ctx context.Context, args ...string) error {
	// list player points in one line
	// format: <player name>: <points>, <player name>: <points>, ...
	g.players.Lock()
	defer g.players.Unlock()

	sortedPlayers := make([]*player.Player, len(g.players.players))
	copy(sortedPlayers, g.players.players)
	sort.Slice(sortedPlayers, func(i, j int) bool {
		return sortedPlayers[i].Points > sortedPlayers[j].Points
	})

	text := ""
	for _, p := range sortedPlayers {
		text += fmt.Sprintf("%s: %d, ", p.Name, p.Points)

	}

	g.ircClient.Privmsg(g.channel, text)
	return nil

}

func (g *Game) HandleHelp(ctx context.Context, args ...string) error {
	// list player points in one line
	// format: <player name>: <points>, <player name>: <points>, ...
	g.players.Lock()
	defer g.players.Unlock()

	text := "Commands: !shoot, !score, !pigeons, !bef, !help, !level"
	g.ircClient.Privmsg(g.channel, text)
	return nil

}

func (g *Game) HandleCount(ctx context.Context, args ...string) error {
	// list player points in one line
	// format: <player name>: <points>, <player name>: <points>, ...
	g.players.Lock()
	defer g.players.Unlock()

	text := ""
	// sort players by count
	sortedPlayers := make([]*player.Player, len(g.players.players))
	copy(sortedPlayers, g.players.players)
	sort.Slice(sortedPlayers, func(i, j int) bool {
		return sortedPlayers[i].Count > sortedPlayers[j].Count
	})
	for _, p := range sortedPlayers {
		text += fmt.Sprintf("%s: %d, ", p.Name, p.Count)

	}

	g.ircClient.Privmsg(g.channel, text)
	return nil

}

func (g *Game) HandleBef(ctx context.Context, args ...string) error {
	g.ircClient.Privmsg(g.channel, "🕊️ ~ coo coo ~ cannot be frens with a rat of the sky ~ 🕊️")

	return nil
}

func (g *Game) HandleLevel(ctx context.Context, args ...string) error {
	// list player points in one line
	// format: <player name>: <points>, <player name>: <points>, ...
	g.players.Lock()
	defer g.players.Unlock()

	text := ""
	// sort players by count
	sortedPlayers := make([]*player.Player, len(g.players.players))
	copy(sortedPlayers, g.players.players)
	sort.Slice(sortedPlayers, func(i, j int) bool {
		return sortedPlayers[i].Count > sortedPlayers[j].Count
	})
	for _, p := range sortedPlayers {
		level := p.GetPlayerLevel()
		text += fmt.Sprintf("%s: %s, ", p.Name, level)

	}

	g.ircClient.Privmsg(g.channel, text)
	return nil

}
