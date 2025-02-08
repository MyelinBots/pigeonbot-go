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

type GameStarted struct {
	sync.Mutex
	started bool
}

type Identified struct {
	sync.Mutex
	identified bool
}

func StartBot() error {
	cfg := config.LoadConfigOrPanic()

	started := &GameStarted{
		started: false,
	}

	identified := &Identified{
		identified: false,
	}

	fmt.Printf("Starting bot with config: %+v\n", cfg)
	database := db.NewDatabase(cfg.DBConfig)
	playerRepo := player.NewPlayerRepository(database)

	ctx, cancel := context.WithCancel(context.Background())
	healthcheck.StartHealthcheck(ctx, cfg.AppConfig)

	ircConfig := irc.NewConfig(cfg.IRCConfig.Nick)
	ircConfig.SSL = cfg.IRCConfig.SSL
	ircConfig.SSLConfig = &tls.Config{
		InsecureSkipVerify: true,
	}
	ircConfig.Server = fmt.Sprintf("%s:%d", cfg.IRCConfig.Host, cfg.IRCConfig.Port)

	c := irc.Client(ircConfig)

	gameInstance := game.NewGame(cfg.GameConfig, c, playerRepo, cfg.IRCConfig.Network, cfg.IRCConfig.Channels[0])

	commandInstance := commands.NewCommandController(gameInstance)

	commandInstance.AddCommand("!shoot", gameInstance.HandleShoot)
	commandInstance.AddCommand("!score", gameInstance.HandlePoints)
	commandInstance.AddCommand("!help", gameInstance.HandleHelp)
	commandInstance.AddCommand("!pigeons", gameInstance.HandleCount)
	commandInstance.AddCommand("!bef", gameInstance.HandleBef)
	//commandInstance.AddCommand("!level", gameInstance.HandleLevel)

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
		started.Lock()
		defer started.Unlock()
		// if channel is first channel in config
		if line.Args[0] == cfg.IRCConfig.Channels[0] && !started.started {
			go gameInstance.Start(ctx)
			started.started = true
		}

		handleNickserv(cfg.IRCConfig, identified, conn)
		return

	})

	c.HandleFunc(irc.INVITE, func(conn *irc.Conn, line *irc.Line) {

		fmt.Printf("Invited to %s\n", line.Args[1])
		conn.Join(line.Args[1])

	})

	c.HandleFunc(irc.PRIVMSG, func(conn *irc.Conn, line *irc.Line) {
		// if message is !shoot
		// if message is !start
		if line.Args[1] == "!start" {
			started.Lock()
			defer started.Unlock()
			if !started.started {
				go gameInstance.Start(ctx)
				started.started = true
				return
			}

		}
		err := commandInstance.HandleCommand(ctx, line)
		if err != nil {
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
