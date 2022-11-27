package cli

type CLI struct {
	command Command
}

func NewCLI() *CLI {
	cli := &CLI{}
	cli.addCommands()
	return cli
}

func (c *CLI) Run() error {
	return c.command.Run()
}

func (c *CLI) addCommands() {
	fcmCommand := NewFCMCommand()

	fcmCommand.AddCommand(NewClusterCommand())
	fcmCommand.AddCommand(NewChaosCommand())

	c.command = fcmCommand
}
