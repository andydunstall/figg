package config

import (
	flags "github.com/jessevdk/go-flags"
)

type Config struct {
	Addr string `short:"a" long:"addr" description:"Listen address for pub/sub clients" default:"127.0.0.1:8000"`

	GossipAddr   string `long:"gossip.addr" description:"Listen address for cluster gossip" default:"127.0.0.1:8001"`
	GossipPeerID string `long:"gossip.peer" description:"ID to identify this peer in the cluster" required:"true"`
	GossipSeeds  string `long:"gossip.seeds" description:"Seed addresses of other nodes in the cluster to bootstrap gossip" default:""`

	Verbose bool `short:"v" long:"verbose" description:"Show verbose debug information"`
}

func ParseConfig() (Config, error) {
	config := Config{}
	_, err := flags.Parse(&config)
	return config, err
}
