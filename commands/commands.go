package commands

import "strings"

// IRC interface for interacting with the IRC client
type IRC interface {
	Privmsg(channel, message string)
	Config() IRCConfig
}

// IRCConfig holds configuration data for the IRC client
type IRCConfig struct {
	Channel string
}

// Game interface for game-related actions
type Game interface {
	AttemptShoot(player string) string
	ScoreBoard() string
	PigeonsShot(player string) string
}

// Commands struct encapsulates the IRC and Game dependencies
type Commands struct {
	IRC     IRC
	Fantasy string
	Game    Game
}

// NewCommands initializes the Commands struct
func Docommands(irc IRC, game Game) *Commands {
	return &Commands{
		IRC:     irc,
		Fantasy: "!",
		Game:    game,
	}
}

// Shoot processes the "shoot" command
func (c *Commands) Shoot(message IRCMessage, command IRCCommand) {
	commandName := "shoot"
	if message.Command == "PRIVMSG" && command.Command == c.Fantasy+commandName {
		player := strings.ToLower(message.MessageFrom)
		shootMessage := c.Game.AttemptShoot(player)
		c.IRC.Privmsg(c.IRC.Config().Channel, shootMessage)
	}
}

// ScoreBoard processes the "score" command
func (c *Commands) ScoreBoard(message IRCMessage, command IRCCommand) {
	commandName := "score"
	if message.Command == "PRIVMSG" && command.Command == c.Fantasy+commandName {
		score := c.Game.ScoreBoard()
		c.IRC.Privmsg(c.IRC.Config().Channel, "Scoreboard: "+score)
	}
}

// ScorePigeon processes the "pigeons" command
func (c *Commands) ScorePigeon(message IRCMessage, command IRCCommand) {
	commandName := "pigeons"
	if message.Command == "PRIVMSG" && command.Command == c.Fantasy+commandName {
		player := strings.ToLower(message.MessageFrom)
		pigeonsShot := c.Game.PigeonsShot(player)
		c.IRC.Privmsg(c.IRC.Config().Channel, "Pigeons shot: "+pigeonsShot)
	}
}

// Bef processes the "bef" command
func (c *Commands) Bef(message IRCMessage, command IRCCommand) {
	commandName := "bef"
	if message.Command == "PRIVMSG" && command.Command == c.Fantasy+commandName {
		c.IRC.Privmsg(c.IRC.Config().Channel, "You cannot be friends with *rat of the sky*")
	}
}

// IRCMessage represents a message received from IRC
type IRCMessage struct {
	Command     string
	MessageFrom string
}

// IRCCommand represents a command parsed from a message
type IRCCommand struct {
	Command string
}
