module github.com/andydunstall/figg/cli

go 1.19

require github.com/spf13/cobra v1.6.1

require (
	github.com/inconshreveable/mousetrap v1.0.1 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	go.uber.org/zap v1.23.0 // indirect
)

require github.com/andydunstall/figg/sdk v0.0.0

replace github.com/andydunstall/figg/sdk v0.0.0 => ../sdk

require github.com/andydunstall/figg/utils v0.0.0 // indirect

replace github.com/andydunstall/figg/utils v0.0.0 => ../utils
