module github.com/andydunstall/figg/fcm/cli

go 1.19

require github.com/spf13/cobra v1.6.1

require (
	github.com/inconshreveable/mousetrap v1.0.1 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
)

require github.com/andydunstall/figg/fcm/sdk v0.0.0

replace github.com/andydunstall/figg/fcm/sdk v0.0.0 => ../sdk
