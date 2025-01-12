package config

import (
	"github.com/jinzhu/configor"
	"strings"
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
}

type IRCConfig struct {
	Host           string `env:"HOST"`
	Port           int    `env:"PORT"`
	SSL            bool   `env:"SSL"`
	Nick           string `env:"NICK"`
	ChannelsString string `env:"CHANNELS"`
	Channels       []string
	Network        string `env:"NETWORK"`
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

func LoadConfigOrPanic() Config {
	var config = Config{}
	configor.Load(&config, "config/config.dev.json")

	config.IRCConfig.Channels = strings.Split(config.IRCConfig.ChannelsString, ",")

	return config
}
