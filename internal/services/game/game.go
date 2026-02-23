//go:generate mockgen -destination=mocks/mock_irc_Client.go -package=mocks github.com/MyelinBots/pigeonbot-go/internal/services/game IRCClient
package game

import (
	"context"
	"fmt"
	rand "math/rand/v2"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/MyelinBots/pigeonbot-go/config"
	player2 "github.com/MyelinBots/pigeonbot-go/internal/db/repositories/player"
	"github.com/MyelinBots/pigeonbot-go/internal/services/actions"
	"github.com/MyelinBots/pigeonbot-go/internal/services/context_manager"
	"github.com/MyelinBots/pigeonbot-go/internal/services/pigeon"
	"github.com/MyelinBots/pigeonbot-go/internal/services/player"
)

const (
	ircBold  = "\x02"
	ircColor = "\x03"
	ircReset = "\x0F"
)

type ActivePigeon struct {
	sync.Mutex
	activePigeon   *pigeon.Pigeon
	IsMating       bool
	SpawnedAt      time.Time
	SpawnID        int64
	CurrentSpawnID int
	IsAlive        bool
}

type Players struct {
	sync.Mutex
	players []*player.Player
}

type PlayerShootState struct {
	LastShotSpawnID int
}

type IRCClient interface {
	Privmsg(channel, message string)
	Notice(target, message string)
	Raw(message string)
}

type pendingPing struct {
	nick    string
	channel string
	start   time.Time
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

	spawnMu        sync.RWMutex
	currentSpawnID int64

	lastShot map[string]*shotState
	shotMu   sync.Mutex

	// --- ping state ---
	pingMu  sync.Mutex
	pending map[string]pendingPing
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
		lastShot:         make(map[string]*shotState),

		// ping state
		pending: make(map[string]pendingPing),
	}
}

// predefinedActions returns the predefined game actions
func predefinedActions() []actions.Action {
	return []actions.Action{
		{
			Action:      "stole",
			Items:       []string{"tv 📺", "wallet 💰👛", "food 🍔 🍕 🍪 🌮"},
			Format:      "❗⚠️ A %s pigeon %s your %s - - 🐦",
			ActionPoint: 10,
		},
		{
			Action:      "pooped",
			Items:       []string{"car 🚗", "head 👤", "laptop 💻"},
			Format:      "❗⚠️ A %s pigeon %s on your %s - - 🐦",
			ActionPoint: 10,
		},
		{
			Action: "landed",
			Items: []string{
				"balcony 🏠🌿", "head 👤", "car 🚗", "house 🏠",
				"swimming pool 🏖️", "bed 🛏️", "couch 🛋️", "laptop 💻",
			},
			Format:      "❗⚠️ A %s pigeon has %s on your %s - - 🐦",
			ActionPoint: 10,
		},
		{
			Action: "mating",
			Items: []string{
				"balcony 🏠🌿", "car 🚗", "bed 🛏️",
				"swimming pool 🏖️", "couch 🛋️", "laptop 💻",
			},
			Format:      "❗⚠️ %s pigeons are %s at your %s - - 🕊️ 💕 🕊️",
			ActionPoint: 10,
		},
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
				timer = 120 // default to 2 minutes
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
		// Use canonical name for consistency (DB should already be lowercase after migration)
		canonicalName := canonicalPlayerName(p.Name)
		g.players.players = append(g.players.players, player.NewPlayer(canonicalName, p.Points, p.Count))
	}

}

func (g *Game) NewPigeonSpawn() int64 {
	g.spawnMu.Lock()
	defer g.spawnMu.Unlock()
	g.currentSpawnID++
	return g.currentSpawnID
}

// ActOnPlayer triggers an action on a player
func (g *Game) ActOnPlayer(ctx context.Context) {
	g.players.Lock()
	g.activePigeon.Lock()
	defer g.activePigeon.Unlock()
	defer g.players.Unlock()

	if g.activePigeon.activePigeon != nil {
		aliveFor := time.Since(g.activePigeon.SpawnedAt)

		// pigeon must live at least 60 seconds (adjust as you like)
		if aliveFor < 60*time.Second {
			return
		}

		g.ircClient.Privmsg(g.channel, fmt.Sprintf(
			"🕊️ ~ coo coo ~ the %s pigeon has made a clean escape ~ 🕊️",
			g.activePigeon.activePigeon.Type,
		))

		g.activePigeon.activePigeon = nil
		g.activePigeon.IsMating = false
		return
	}

	randomPigeon := g.pigeons[rand.IntN(len(g.pigeons))]
	randomAction := g.actions[rand.IntN(len(g.actions))]

	// ✅ นกเกิดใหม่จริง ๆ → เพิ่ม spawnID
	newSpawnID := g.NewPigeonSpawn()

	g.activePigeon.activePigeon = randomPigeon
	g.activePigeon.IsMating = (randomAction.Action == "mating")
	g.activePigeon.SpawnedAt = time.Now()

	g.ircClient.Privmsg(g.channel, randomAction.Act(randomPigeon.Type))

	// (optional debug)
	fmt.Printf("[dbg] NEW PIGEON spawnID=%d type=%s\n", newSpawnID, randomPigeon.Type)

}

// AddPlayer adds a new player to the game
func (g *Game) addPlayer(ctx context.Context, name string) (*player.Player, error) {
	g.players.Lock()
	defer g.players.Unlock()

	// Canonicalize name for consistent storage
	canonicalName := canonicalPlayerName(name)

	for _, p := range g.players.players {
		if p.Name == canonicalName {
			return p, nil
		}
	}

	newPlayer := player.NewPlayer(canonicalName, 0, 0)
	g.players.players = append(g.players.players, newPlayer)
	playerEntity := player2.Player{
		Count:   0,
		Points:  0,
		Name:    canonicalName,
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

	// Canonicalize name for consistent lookup
	canonicalName := canonicalPlayerName(name)

	for _, p := range g.players.players {
		if p.Name == canonicalName {
			g.players.Unlock()
			return p, nil
		}
	}
	g.players.Unlock()

	return g.addPlayer(ctx, name)

}

func (g *Game) CurrentSpawnID() int64 {
	g.spawnMu.RLock()
	defer g.spawnMu.RUnlock()
	return g.currentSpawnID
}

func (g *Game) HandleShoot(ctx context.Context, args ...string) error {
	name := context_manager.GetNickContext(ctx)

	// 🔒 PER-USER COOLDOWN CHECK (5 shots before cooldown)
	spawnID := g.CurrentSpawnID()
	ok, wait := g.canShoot(name, spawnID)
	if !ok {
		g.ircClient.Privmsg(g.channel,
			fmt.Sprintf("...%s slow down... you can shoot again in %.1f seconds ⏳🕊️", name, wait.Seconds()),
		)
		return nil
	}

	// 🔐 Lock pigeon AFTER cooldown
	g.activePigeon.Lock()
	defer g.activePigeon.Unlock()

	if g.activePigeon.activePigeon == nil {
		g.ircClient.Privmsg(
			g.channel,
			fmt.Sprintf("❗⚠️ %s has shot a pigeon!, but there are no pigeons to shoot! - - 🐦", name),
		)
		return nil
	}

	foundPlayer, err := g.FindPlayer(ctx, name)
	if err != nil {
		g.ircClient.Privmsg(
			g.channel,
			fmt.Sprintf("❗⚠️ %s has shot a pigeon, but there was an error finding the player! - - 🐦", name),
		)
		return err
	}

	randomValue := rand.IntN(100)
	success := randomValue < g.activePigeon.activePigeon.Success

	if success {
		foundPlayer.Points += g.activePigeon.activePigeon.Points
		foundPlayer.Count++

		level := foundPlayer.GetPlayerLevel()

		g.ircClient.Privmsg(
			g.channel,
			fmt.Sprintf(
				"❗⚠️ %s has shot a pigeon! - - 🐦 🔫 You are a murderer! . .  You have shot a total of %s pigeon(s)! . . 🐦 🕊️ . . You now have a total of %s points and reached the level: %s",
				name,
				fmtNum(foundPlayer.Count),
				fmtNum(foundPlayer.Points),
				level,
			),
		)

		if eggMsg, err := g.EggsAfterShot(ctx, name); err == nil && eggMsg != "" {
			g.ircClient.Privmsg(g.channel, eggMsg)
		}

		if rareMsg, err := g.TryRareEgg(ctx, name); err == nil && rareMsg != "" {
			g.ircClient.Privmsg(g.channel, rareMsg)
		}

		g.activePigeon.activePigeon = nil
		g.activePigeon.IsMating = false

	} else {
		g.ircClient.Privmsg(
			g.channel,
			fmt.Sprintf("❗⚠️ %s has shot a pigeon, but it got away! - - 🐦", name),
		)
	}

	return g.SavePlayers(ctx)
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

	text := "Commands: !shoot, !score, !pigeons, !bef, !help, !level, !top5, !top10, !eggs"
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

func (g *Game) handleTopN(ctx context.Context, n int) error {
	// 🏆 Header (gold)
	g.ircClient.Privmsg(
		g.channel,
		fmt.Sprintf(
			"%s%s%s",
			ircBold,
			c(fmt.Sprintf("🏆 Top %d Pigeon Hunters", n), 8),
			ircReset,
		),
	)

	topPlayers, err := g.TopByPoints(ctx, n)
	if err != nil {
		g.ircClient.Privmsg(g.channel, "Error fetching top players")
		return err
	}

	for i, p := range topPlayers {
		rank := medal(i)

		pointsText := fmt.Sprintf("%s points", fmtNum(p.Points))
		pigeonsText := fmt.Sprintf("%s pigeons", fmtNum(p.Count))
		levelText := fmt.Sprintf("Level: %s ", g.LevelFor(p.Points, p.Count))
		eggsText := fmt.Sprintf("Eggs: %s", fmtNum(p.Eggs))
		rareText := fmt.Sprintf("Rare: %s 🌟", fmtNum(p.RareEggs))

		g.ircClient.Privmsg(
			g.channel,
			fmt.Sprintf(
				"%s %s :::::: %s | %s | %s | %s (%s)",
				rank,
				p.Name,
				c(pointsText, 7),  // orange points
				c(pigeonsText, 4), // 🔴 pigeons
				c(levelText, 13),  // pink level
				c(eggsText, 8),    // 🟡 eggs
				c(rareText, 8),    // 🌟 rare eggs (gold)
			),
		)
	}

	return nil
}

func (g *Game) HandleTop5(ctx context.Context, args ...string) error {
	return g.handleTopN(ctx, 5)
}

func (g *Game) HandleTop10(ctx context.Context, args ...string) error {
	return g.handleTopN(ctx, 10)
}

func c(s string, fg int) string { // foreground only
	return fmt.Sprintf("%s%02d%s%s", ircColor, fg, s, ircReset)
}

func medal(i int) string {
	switch i {
	case 0:
		return "🥇"
	case 1:
		return "🥈"
	case 2:
		return "🥉"
	default:
		return "•"
	}
}

// --- Helpers for commands.TopHandler ---

// Irc exposes the IRC client (read-only)
func (g *Game) Irc() IRCClient { return g.ircClient }

// Channel returns the current channel (read-only)
func (g *Game) Channel() string { return g.channel }

// Network returns the current network (read-only)
func (g *Game) Network() string { return g.network }

// TopByPoints fetches the top-N players for this game's scope.
func (g *Game) TopByPoints(ctx context.Context, limit int) ([]*player2.Player, error) {
	if limit <= 0 {
		limit = 5
	}
	if limit > 50 {
		limit = 50
	}
	return g.playerRepository.TopByPoints(ctx, g.network, g.channel, limit)
}

// LevelFor maps (points,count) to the player's level using your services/player logic.
func (g *Game) LevelFor(points, count int) string {
	tmp := &player.Player{Name: "", Points: points, Count: count}
	return tmp.GetPlayerLevel()
}

func (g *Game) HandlePingCommand(ctx context.Context, args ...string) error {
	nick := context_manager.GetNickContext(ctx)

	token := fmt.Sprintf("%d", time.Now().UnixNano())

	g.pingMu.Lock()
	g.pending[token] = pendingPing{
		nick:    nick,
		channel: g.channel,
		start:   time.Now(),
	}
	g.pingMu.Unlock()

	// Send CTCP PING to the user (their client will auto-reply)
	g.ircClient.Privmsg(nick, "\x01PING "+token+"\x01")

	// Timeout fallback
	time.AfterFunc(10*time.Second, func() {
		g.pingMu.Lock()
		p, ok := g.pending[token]
		if ok {
			delete(g.pending, token)
		}
		g.pingMu.Unlock()

		if ok {
			g.ircClient.Privmsg(p.channel, fmt.Sprintf("%s: Pong (timeout)", p.nick))
		}
	})

	return nil
}

// HandleCTCPReply is called when we receive a CTCPREPLY event
// goirc parses CTCP: Args[0] = verb (PING), Args[1] = target, Args[2+] = payload
func (g *Game) HandleCTCPReply(from string, args []string) {
	if len(args) < 3 || strings.ToUpper(args[0]) != "PING" {
		return
	}

	token := args[2]
	g.pingMu.Lock()
	p, ok := g.pending[token]
	if ok {
		delete(g.pending, token)
	}
	g.pingMu.Unlock()

	if !ok {
		return
	}

	secs := time.Since(p.start).Seconds()
	g.ircClient.Privmsg(p.channel, fmt.Sprintf("%s: Pong (%.3fs)", p.nick, secs))
}
