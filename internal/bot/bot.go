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

	// for each channel make a new game instance
	gameInstances := &GameInstances{
		games:            make(map[string]*game.Game),
		commandInstances: make(map[string]commands.CommandController),
		GameStarted:      make(map[string]bool),
	}

	for _, channel := range cfg.IRCConfig.Channels {
		gameInstances.Lock()

		// ✅ IMPORTANT: pass wrapper (Privmsg/Notice/Raw) not raw conn
		gameInstance := game.NewGame(cfg.GameConfig, IRCWrapper{c}, playerRepo, cfg.IRCConfig.Network, channel)
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

		// ✅ ping command (CTCP PING -> NOTICE reply -> Pong)
		commandInstance.AddCommand("!ping", gameInstance.HandlePingCommand)

		gameInstances.games[channel] = gameInstance
		gameInstances.commandInstances[channel] = commandInstance
		gameInstances.GameStarted[channel] = false

		gameInstances.Unlock()
	}

	c.HandleFunc(irc.CONNECTED, func(conn *irc.Conn, _ *irc.Line) {
		fmt.Printf("Connected to %s\n", cfg.IRCConfig.Host)
		for _, channel := range cfg.IRCConfig.Channels {
			fmt.Printf("Joining channel %s\n", channel)
			conn.Join(channel)
		}
	})

	// Also join on MOTD end / no MOTD
	c.HandleFunc("422", func(conn *irc.Conn, _ *irc.Line) {
		for _, channel := range cfg.IRCConfig.Channels {
			conn.Join(channel)
		}
	})
	c.HandleFunc("376", func(conn *irc.Conn, _ *irc.Line) {
		for _, channel := range cfg.IRCConfig.Channels {
			conn.Join(channel)
		}
	})

	c.HandleFunc(irc.JOIN, func(conn *irc.Conn, line *irc.Line) {
		channel := line.Args[0]
		fmt.Printf("Joined %s\n", channel)

		handleNickserv(cfg.IRCConfig, identified, conn)

		gameInstances.Lock()
		defer gameInstances.Unlock()

		if g, ok := gameInstances.games[channel]; ok {
			if !gameInstances.GameStarted[channel] {
				go g.Start(ctx)
				gameInstances.GameStarted[channel] = true
			}
		}
	})

	c.HandleFunc(irc.PRIVMSG, func(_ *irc.Conn, line *irc.Line) {
		channel := line.Args[0]
		msg := line.Args[1]

		// manual start
		if msg == "!start" {
			gameInstances.Lock()
			g, ok := gameInstances.games[channel]
			if !ok {
				gameInstances.Unlock()
				return
			}
			if gameInstances.GameStarted[channel] {
				gameInstances.Unlock()
				fmt.Printf("Game already started for %s\n", channel)
				return
			}
			gameInstances.GameStarted[channel] = true
			gameInstances.Unlock()

			fmt.Printf("Starting gameInstance for %s\n", channel)
			go g.Start(ctx)
			return
		}

		gameInstances.Lock()
		commandInstance := gameInstances.commandInstances[channel]
		gameInstances.Unlock()

		ctxWithNick := context.WithValue(ctx, "nick", line.Nick)
		if err := commandInstance.HandleCommand(ctxWithNick, line); err != nil {
			fmt.Printf("Error handling command: %s\n", err.Error())
			return
		}
	})

	// CTCPREPLY handler - goirc parses CTCP and dispatches to this event
	c.HandleFunc(irc.CTCPREPLY, func(_ *irc.Conn, line *irc.Line) {
		if len(line.Args) < 1 {
			return
		}

		gameInstances.Lock()
		defer gameInstances.Unlock()

		for _, g := range gameInstances.games {
			g.HandleCTCPReply(line.Nick, line.Args)
		}
	})

	quit := make(chan bool, 1)
	c.HandleFunc(irc.DISCONNECTED, func(_ *irc.Conn, _ *irc.Line) { quit <- true })

	if err := c.Connect(); err != nil {
		fmt.Printf("Connection error: %s\n", err.Error())
		return err
	}

	<-quit
	return nil
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

type IRCWrapper struct {
	*irc.Conn
}

func (w IRCWrapper) Privmsg(channel, message string) { w.Conn.Privmsg(channel, message) }
func (w IRCWrapper) Kick(channel, nick, reason string) {
	w.Conn.Kick(channel, nick, reason)
}
func (w IRCWrapper) Notice(target, message string) { w.Conn.Notice(target, message) }
func (w IRCWrapper) Raw(message string)            { w.Conn.Raw(message) }
