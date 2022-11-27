package cli

import (
	"github.com/spf13/cobra"
)

type Command interface {
	Run() error
	CobraCommand() *cobra.Command
}
