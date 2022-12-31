package bench

import (
	"fmt"
	"math/rand"
	"strings"

	humanize "github.com/dustin/go-humanize"
	flags "github.com/jessevdk/go-flags"
)

type Config struct {
	Messages    int    `long:"msgs" description:"Number of messages to publish" default:"1000000"`
	MessageSize int    `long:"msg-size" description:"Size of the test messages" default:"128"`
	Topic       string `long:"topic" description:"Topic name to use for benchmark (if empty generates random topic)" default:""`
	Publishers  int    `long:"pubs" description:"Number of concurrent publishers" default:"1"`
	Subscribers int    `long:"subs" description:"Number of concurrent subscribers" default:"1"`
	Resumers    int    `long:"resumers" description:"Number of concurrent resumers (subscribers that start subscribing after all messages published)" default:"1"`

	Addr string `long:"addr" description:"Address of the Figg server" default:"127.0.0.1:8119"`

	Verbose bool `short:"v" long:"verbose" description:"Add verbose logging to SDK clients"`
}

func (c Config) String() string {
	return fmt.Sprintf(
		"msgs=%s msg-size=%s topic=%s addr=%s publishers=%d subscribers=%d resumers=%d",
		humanize.Comma(int64(c.Messages)),
		strings.ReplaceAll(humanize.Bytes(uint64(c.MessageSize)), " ", ""),
		c.Topic,
		c.Addr,
		c.Publishers,
		c.Subscribers,
		c.Resumers,
	)
}

func ParseConfig() (*Config, error) {
	config := &Config{}
	_, err := flags.Parse(config)
	if err != nil {
		return nil, err
	}

	if config.Messages < 0 {
		return nil, fmt.Errorf("msgs cannot be negative")
	}
	if config.MessageSize < 0 {
		return nil, fmt.Errorf("msg-size cannot be negative")
	}
	if config.Publishers <= 0 {
		return nil, fmt.Errorf("publishers must be positive")
	}

	if config.Topic == "" {
		config.Topic = fmt.Sprintf("bench-%d", rand.Int()%0xffff)
	}

	return config, nil
}
