package config

import (
	flags "github.com/jessevdk/go-flags"
	"go.uber.org/zap/zapcore"
)

type Config struct {
	Addr string `short:"a" long:"addr" description:"Listen address for pub/sub clients" default:"127.0.0.1:8119"`

	AdminAddr string `long:"admin-addr" description:"Listen address for admin endpoints" default:"127.0.0.1:8229"`

	CommitLogInMemory    bool   `long:"commitlog.inmemory" description:"Whether the commit log should be in-memory only"`
	CommitLogDir         string `long:"commitlog.dir" description:"The directory to store the commit log segments if persisted" default:"./data"`
	CommitLogSegmentSize uint64 `long:"commitlog.segment-size" description:"The size of the commit log segments to use" default:"4194304"`

	Verbose bool `short:"v" long:"verbose" description:"Show verbose debug information"`
}

func (c Config) MarshalLogObject(e zapcore.ObjectEncoder) error {
	e.AddString("addr", c.Addr)

	e.AddBool("commitlog.inmemory", c.CommitLogInMemory)
	e.AddString("commitlog.dir", c.CommitLogDir)
	e.AddUint64("commitlog.segment-size", c.CommitLogSegmentSize)

	e.AddBool("verbose", c.Verbose)
	return nil
}

func ParseConfig() (Config, error) {
	config := Config{}
	_, err := flags.Parse(&config)
	return config, err
}
