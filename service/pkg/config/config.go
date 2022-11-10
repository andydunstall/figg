package config

import (
	flags "github.com/jessevdk/go-flags"
)

type Config struct {
	Addr string `short:"a" long:"addr" description:"Listen address for pub/sub clients" default:"127.0.0.1:8119"`

	Verbose bool `short:"v" long:"verbose" description:"Show verbose debug information"`
}

func ParseConfig() (Config, error) {
	config := Config{}
	_, err := flags.Parse(&config)
	return config, err
}
