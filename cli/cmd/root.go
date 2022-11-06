package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	figgAddr string
)

var rootCmd = &cobra.Command{
	Use: "figg",
	Run: func(cmd *cobra.Command, args []string) {},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&figgAddr, "addr", "127.0.0.1:8119", "figg cluster address")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
