package config

import (
	flags "github.com/jessevdk/go-flags"
)

type Config struct {
	Verbose bool `short:"v" long:"verbose" description:"Show verbose debug information"`
}

func ParseConfig() (Config, error) {
	config := Config{}
	_, err := flags.Parse(&config)
	return config, err
}
