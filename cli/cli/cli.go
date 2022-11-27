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
	figgCommand := NewFiggCommand()

	figgCommand.AddCommand(NewPublishCommand(figgCommand.Config))
	figgCommand.AddCommand(NewSubscribeCommand(figgCommand.Config))
	figgCommand.AddCommand(NewBenchCommand(figgCommand.Config))
	figgCommand.AddCommand(NewStreamCommand(figgCommand.Config))

	c.command = figgCommand
}
