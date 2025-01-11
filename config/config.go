package config

import "github.com/jinzhu/configor"

type Config struct {
	AppConfig AppConfig `env:"APPCONFIG"`
	IRCConfig IRCConfig `env:"IRCCONFIG"`
}

type AppConfig struct {
	APPName string `default:"pigeonbot"`
	Version string `default:"x.x.x" env:"VERSION"`
}

type IRCConfig struct {
	Host string `env:"HOST"`
	Port int    `env:"PORT"`
	SSL  bool   `env:"SSL"`
	Nick string `env:"NICK"`
}

func LoadConfigOrPanic() Config {
	var config = Config{}
	configor.Load(&config, "config/config.dev.json")

	return config
}
