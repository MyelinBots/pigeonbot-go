package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/jinzhu/configor"
)

type Config struct {
	AppConfig  AppConfig  `env:"APPCONFIG"`
	IRCConfig  IRCConfig  `env:"IRCCONFIG"`
	DBConfig   DBConfig   `env:"DBCONFIG"`
	GameConfig GameConfig `env:"GAMECONFIG"`
}

type AppConfig struct {
	APPName string `default:"pigeonbot"`
	Version string `default:"x.x.x" env:"VERSION"`
	Port    int    `default:"8080" env:"APP_PORT"`
}

type IRCConfig struct {
	Host             string `env:"HOST"`
	Port             int    `env:"PORT"`
	SSL              bool   `env:"SSL"`
	Nick             string `env:"NICK"`
	RealName         string `env:"REALNAME"`
	ChannelsString   string `env:"CHANNELS"`
	Channels         []string
	Network          string `env:"NETWORK"`
	NickservCommand  string `env:"NICKSERV_COMMAND" default:"PRIVMSG NickServ IDENTIFY %s"`
	NickservPassword string `env:"NICKSERV_PASSWORD" default:""`
}

type DBConfig struct {
	Host     string `default:"localhost" env:"DBHOST"`
	DataBase string `default:"pigeon" env:"DBNAME"`
	User     string `default:"postgres" env:"DBUSERNAME"`
	Password string `required:"true" env:"DBPASSWORD" default:"mysecretpassword"`
	Port     uint   `default:"5432" env:"DBPORT"`
	SSLMode  string `default:"disable" env:"DBSSL"`
}

type GameConfig struct {
	Interval int `env:"INTERVAL" default:"10"`
}

// findProjectRoot walks up from the current directory looking for go.mod
func findProjectRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

func LoadConfigOrPanic() Config {
	var config = Config{}

	configPath := "config/config.dev.json"
	if root := findProjectRoot(); root != "" {
		configPath = filepath.Join(root, configPath)
	}

	configor.Load(&config, configPath)

	config.IRCConfig.Channels = strings.Split(config.IRCConfig.ChannelsString, ",")

	return config
}
