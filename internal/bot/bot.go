package bot

import (
	"context"
	"crypto/tls"
	"fmt"
	"sync"

	"github.com/MyelinBots/pigeonbot-go/config"
	"github.com/MyelinBots/pigeonbot-go/internal/db"
	"github.com/MyelinBots/pigeonbot-go/internal/db/repositories/player"
	"github.com/MyelinBots/pigeonbot-go/internal/healthcheck"
	"github.com/MyelinBots/pigeonbot-go/internal/services/commands"
	"github.com/MyelinBots/pigeonbot-go/internal/services/game"
	irc "github.com/fluffle/goirc/client"
)

type Identified struct {
	sync.Mutex
	identified bool
}

type GameInstances struct {
	sync.Mutex
	GameStarted      map[string]bool
	games            map[string]*game.Game
	commandInstances map[string]commands.CommandController
}

func StartBot() error {
	cfg := config.LoadConfigOrPanic()

	identified := &Identified{identified: false}
	fmt.Printf("Starting bot with config: %+v\n", cfg)

	database := db.NewDatabase(cfg.DBConfig)
	playerRepo := player.NewPlayerRepository(database)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	healthcheck.StartHealthcheck(ctx, cfg.AppConfig)

	ircConfig := irc.NewConfig(cfg.IRCConfig.Nick)
	ircConfig.Me.Name = cfg.IRCConfig.RealName
	ircConfig.Me.Ident = cfg.IRCConfig.Nick
	ircConfig.SSL = cfg.IRCConfig.SSL
	ircConfig.SSLConfig = &tls.Config{InsecureSkipVerify: true}
	ircConfig.Server = fmt.Sprintf("%s:%d", cfg.IRCConfig.Host, cfg.IRCConfig.Port)

	c := irc.Client(ircConfig)

	// Declare gameInstances before usage
	gameInstances := &GameInstances{
		games:            make(map[string]*game.Game),
		commandInstances: make(map[string]commands.CommandController),
		GameStarted:      make(map[string]bool),
	}

	// Register channel-specific game and command controllers
	for _, channel := range cfg.IRCConfig.Channels {
		gameInstances.Lock()
		gameInstance := game.NewGame(cfg.GameConfig, c, playerRepo, cfg.IRCConfig.Network, channel)
		commandInstance := commands.NewCommandController(gameInstance)

		commandInstance.AddCommand("!shoot", gameInstance.HandleShoot)
		commandInstance.AddCommand("!score", gameInstance.HandlePoints)
		commandInstance.AddCommand("!help", gameInstance.HandleHelp)
		commandInstance.AddCommand("!pigeons", gameInstance.HandleCount)
		commandInstance.AddCommand("!bef", gameInstance.HandleBef)
		commandInstance.AddCommand("!level", gameInstance.HandleLevel)
		commandInstance.AddCommand("!invite", commands.InviteHandler(c, startNewGameInstance(cfg, c, playerRepo, gameInstances)))

		gameInstances.games[channel] = gameInstance
		gameInstances.commandInstances[channel] = commandInstance
		gameInstances.GameStarted[channel] = false
		gameInstances.Unlock()
	}

	// Join channels on successful connection
	c.HandleFunc(irc.CONNECTED, func(conn *irc.Conn, line *irc.Line) {
		fmt.Printf("Connected to %s\n", cfg.IRCConfig.Host)
		for _, channel := range cfg.IRCConfig.Channels {
			fmt.Printf("Joining channel %s\n", channel)
			conn.Join(channel)
		}
	})

	// Join fallback handlers
	c.HandleFunc("422", func(conn *irc.Conn, line *irc.Line) {
		for _, channel := range cfg.IRCConfig.Channels {
			conn.Join(channel)
		}
	})
	c.HandleFunc("376", func(conn *irc.Conn, line *irc.Line) {
		for _, channel := range cfg.IRCConfig.Channels {
			conn.Join(channel)
		}
	})

	// Handle invite to a new channel
	c.HandleFunc(irc.INVITE, func(conn *irc.Conn, line *irc.Line) {
		channel := line.Args[1]
		fmt.Printf("Invited to %s\n", channel)
		conn.Join(channel)

		gameInstances.Lock()
		if _, exists := gameInstances.games[channel]; !exists {
			gameInstance := game.NewGame(cfg.GameConfig, conn, playerRepo, cfg.IRCConfig.Network, channel)
			commandInstance := commands.NewCommandController(gameInstance)

			commandInstance.AddCommand("!shoot", gameInstance.HandleShoot)
			commandInstance.AddCommand("!score", gameInstance.HandlePoints)
			commandInstance.AddCommand("!help", gameInstance.HandleHelp)
			commandInstance.AddCommand("!pigeons", gameInstance.HandleCount)
			commandInstance.AddCommand("!bef", gameInstance.HandleBef)
			commandInstance.AddCommand("!level", gameInstance.HandleLevel)
			commandInstance.AddCommand("!invite", commands.InviteHandler(c, startNewGameInstance(cfg, c, playerRepo, gameInstances)))

			gameInstances.games[channel] = gameInstance
			gameInstances.commandInstances[channel] = commandInstance
			gameInstances.GameStarted[channel] = false
		}
		gameInstances.Unlock()
	})

	// Game logic on JOIN
	c.HandleFunc(irc.JOIN, func(conn *irc.Conn, line *irc.Line) {
		channel := line.Args[0]
		fmt.Printf("Joined %s\n", channel)
		gameInstances.Lock()
		if gameInstance, ok := gameInstances.games[channel]; ok && !gameInstances.GameStarted[channel] {
			go gameInstance.Start(ctx)
			gameInstances.GameStarted[channel] = true
		}
		gameInstances.Unlock()

		handleNickserv(cfg.IRCConfig, identified, conn)
	})

	// Command handler
	c.HandleFunc(irc.PRIVMSG, func(conn *irc.Conn, line *irc.Line) {
		// if message is !shoot
		// if message is !start
		if line.Args[1] == "!start" {

			gameInstances.Lock()
			if gameInstance, ok := gameInstances.games[line.Args[0]]; ok {

				gameInstances.Unlock()

				if gameInstances.GameStarted[line.Args[0]] {
					fmt.Printf("Game already started for %s\n", line.Args[0])
					return
				}

				fmt.Printf("Starting gameInstance for %s\n", line.Args[0])
				gameInstance.Start(ctx)
				gameInstances.GameStarted[line.Args[0]] = true
				return
			}

		}
		gameInstances.Lock()
		commandInstance := gameInstances.commandInstances[line.Args[0]]
		gameInstances.Unlock()
		err := commandInstance.HandleCommand(ctx, line)
		if err != nil {
			fmt.Printf("Error handling command: %s\n", err.Error())
			return
		}

	})

	// Quit on disconnect
	quit := make(chan bool)
	c.HandleFunc(irc.DISCONNECTED, func(conn *irc.Conn, line *irc.Line) {
		quit <- true
	})

	// Connect
	if err := c.Connect(); err != nil {
		fmt.Printf("Connection error: %s\n", err.Error())
		return err
	}

	<-quit
	return nil
}

func startNewGameInstance(cfg config.Config, conn *irc.Conn, playerRepo player.PlayerRepository, gameInstances *GameInstances) func(ctx context.Context, channel string) {
	return func(ctx context.Context, channel string) {
		gameInstances.Lock()
		gameInstance := game.NewGame(cfg.GameConfig, conn, playerRepo, cfg.IRCConfig.Network, channel)
		commandInstance := commands.NewCommandController(gameInstance)

		commandInstance.AddCommand("!shoot", gameInstance.HandleShoot)
		commandInstance.AddCommand("!score", gameInstance.HandlePoints)
		commandInstance.AddCommand("!help", gameInstance.HandleHelp)
		commandInstance.AddCommand("!pigeons", gameInstance.HandleCount)
		commandInstance.AddCommand("!bef", gameInstance.HandleBef)
		commandInstance.AddCommand("!level", gameInstance.HandleLevel)
		commandInstance.AddCommand("!invite", commands.InviteHandler(conn, startNewGameInstance(cfg, conn, playerRepo, gameInstances)))

		gameInstances.games[channel] = gameInstance
		gameInstances.commandInstances[channel] = commandInstance
		gameInstances.GameStarted[channel] = false
		gameInstances.Unlock()
	}
}

func handleNickserv(cfg config.IRCConfig, identified *Identified, c *irc.Conn) {
	identified.Lock()
	defer identified.Unlock()

	if !identified.identified && cfg.NickservPassword != "" {
		command := fmt.Sprintf(cfg.NickservCommand, cfg.NickservPassword)
		c.Raw(command)
		identified.identified = true
	}
}
