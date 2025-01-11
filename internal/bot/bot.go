package bot

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/MyelinBots/pigeonbot-go/config"
	"github.com/MyelinBots/pigeonbot-go/internal/db"
	"github.com/MyelinBots/pigeonbot-go/internal/db/repositories/player"
	"github.com/MyelinBots/pigeonbot-go/internal/services/commands"
	"github.com/MyelinBots/pigeonbot-go/internal/services/game"
	irc "github.com/fluffle/goirc/client"
)

func StartBot() error {
	cfg := config.LoadConfigOrPanic()

	database := db.NewDatabase(cfg.DBConfig)
	playerRepo := player.NewPlayerRepository(database)

	ctx, cancel := context.WithCancel(context.Background())

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
	commandInstance.AddCommand("!points", gameInstance.HandlePoints)

	c.HandleFunc(irc.CONNECTED, func(conn *irc.Conn, line *irc.Line) {
		for _, channel := range cfg.IRCConfig.Channels {
			conn.Join(channel)
		}

		go gameInstance.Start(ctx)
	})

	c.HandleFunc(irc.PRIVMSG, func(conn *irc.Conn, line *irc.Line) {
		// if message is !shoot
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
