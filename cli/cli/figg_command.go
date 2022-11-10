package cli

import (
	// "fmt"
	// "os"

	"github.com/spf13/cobra"
)

var (
	figgAddr string
)

// var rootCmd = &cobra.Command{
// 	Use: "figg",
// 	Run: func(cmd *cobra.Command, args []string) {},
// }

// func init() {
// 	rootCmd.PersistentFlags().StringVar(&figgAddr, "addr", "127.0.0.1:8119", "figg cluster address")
// }

type FiggConfig struct {
	Addr string
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
