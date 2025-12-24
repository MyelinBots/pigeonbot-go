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

	identified := &Identified{
		identified: false,
	}

	fmt.Printf("Starting bot with config: %+v\n", cfg)
	database := db.NewDatabase(cfg.DBConfig)
	playerRepo := player.NewPlayerRepository(database)

	ctx, cancel := context.WithCancel(context.Background())
	healthcheck.StartHealthcheck(ctx, cfg.AppConfig)

	ircConfig := irc.NewConfig(cfg.IRCConfig.Nick)
	ircConfig.Me.Name = cfg.IRCConfig.RealName
	ircConfig.Me.Ident = cfg.IRCConfig.Nick
	ircConfig.SSL = cfg.IRCConfig.SSL
	ircConfig.SSLConfig = &tls.Config{
		InsecureSkipVerify: true,
	}

	ircConfig.Server = fmt.Sprintf("%s:%d", cfg.IRCConfig.Host, cfg.IRCConfig.Port)

	c := irc.Client(ircConfig)

	// for each channel make a new game instance
	gameInstances := &GameInstances{
		games:            make(map[string]*game.Game),
		commandInstances: make(map[string]commands.CommandController),
		GameStarted:      make(map[string]bool),
	}
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
		commandInstance.AddCommand("!top5", gameInstance.HandleTop5)
		commandInstance.AddCommand("!top10", gameInstance.HandleTop10)
		commandInstance.AddCommand("!eggs", gameInstance.HandleEggs)
		gameInstances.games[channel] = gameInstance
		gameInstances.commandInstances[channel] = commandInstance
		gameInstances.GameStarted[channel] = false
		gameInstances.Unlock()
	}

	c.HandleFunc(irc.CONNECTED, func(conn *irc.Conn, line *irc.Line) {
		fmt.Printf("Connected to %s\n", cfg.IRCConfig.Host)
		// list channels from config
		fmt.Printf("Joining channel %v\n", cfg.IRCConfig)

		for _, channel := range cfg.IRCConfig.Channels {
			fmt.Printf("Joining channel %s\n", channel)
			conn.Join(channel)
		}

	})

	c.HandleFunc("422", func(conn *irc.Conn, line *irc.Line) {
		for _, channel := range cfg.IRCConfig.Channels {
			fmt.Printf("Joining channel %s\n", channel)
			conn.Join(channel)
		}
	})

	c.HandleFunc("376", func(conn *irc.Conn, line *irc.Line) {
		for _, channel := range cfg.IRCConfig.Channels {
			fmt.Printf("Joining channel %s\n", channel)
			conn.Join(channel)
		}
	})

	c.HandleFunc(irc.JOIN, func(conn *irc.Conn, line *irc.Line) {
		fmt.Printf("Joined %s\n", line.Args[0])
		gameInstances.Lock()

		if gameInstance, ok := gameInstances.games[line.Args[0]]; ok {
			gameInstance := gameInstance

			if !gameInstances.GameStarted[line.Args[0]] {
				go gameInstance.Start(ctx)
				gameInstances.GameStarted[line.Args[0]] = true
			}

		}
		gameInstances.Unlock()
		//// if channel is first channel in config
		//if line.Args[0] == cfg.IRCConfig.Channels[0] && !started.started {
		//	go gameInstance.Start(ctx)
		//	started.started = true
		//}

		handleNickserv(cfg.IRCConfig, identified, conn)

	})
	// disable invites
	//c.HandleFunc(irc.INVITE, func(conn *irc.Conn, line *irc.Line) {
	//
	//	fmt.Printf("Invited to %s\n", line.Args[1])
	//	conn.Join(line.Args[1])
	//
	//})

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
		ctxWithNick := context.WithValue(ctx, "nick", line.Nick)
		if err := commandInstance.HandleCommand(ctxWithNick, line); err != nil {
			fmt.Printf("Error handling command: %s\n", err.Error())
			return
		}

	})

	quit := make(chan bool)
	c.HandleFunc(irc.DISCONNECTED, func(conn *irc.Conn, line *irc.Line) {
		quit <- true
	})

	if err := c.Connect(); err != nil {
		fmt.Printf("Connection error: %s\n", err.Error())
	}

	<-quit

	cancel()
	return nil
}

func handleNickserv(cfg config.IRCConfig, identified *Identified, c *irc.Conn) {
	identified.Lock()
	defer identified.Unlock()

	if !identified.identified && cfg.NickservPassword != "" {
		// use nickserv command
		command := fmt.Sprintf(cfg.NickservCommand, cfg.NickservPassword)
		c.Raw(command)
	}
}
