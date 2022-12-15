package cli

import (
	"github.com/spf13/cobra"
)

var (
	figgAddr string
)

type FiggConfig struct {
	Addr string
	Verbose bool
}

type FiggCommand struct {
	Config   *FiggConfig
	cobraCmd *cobra.Command
}

func NewFiggCommand() *FiggCommand {
	var cobraCmd = &cobra.Command{
		Use: "figg",
		Run: func(cmd *cobra.Command, args []string) {},
	}
	config := &FiggConfig{}
	cobraCmd.PersistentFlags().StringVar(&config.Addr, "addr", "127.0.0.1:8119", "figg cluster address")
	cobraCmd.PersistentFlags().BoolVar(&config.Verbose, "verbose", false, "show verbose debug output")
	return &FiggCommand{
		Config:   config,
		cobraCmd: cobraCmd,
	}
}

func (c *FiggCommand) Run() error {
	return c.cobraCmd.Execute()
}

func (c *FiggCommand) CobraCommand() *cobra.Command {
	return c.cobraCmd
}

func (c *FiggCommand) AddCommand(command Command) {
	c.cobraCmd.AddCommand(command.CobraCommand())
}
