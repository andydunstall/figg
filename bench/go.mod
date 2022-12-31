module github.com/andydunstall/figg/bench

go 1.19

require github.com/jessevdk/go-flags v1.5.0

require golang.org/x/sys v0.0.0-20210320140829-1e4c9ba3b0c4 // indirect

require (
	github.com/andydunstall/figg/sdk v0.0.0
	github.com/dustin/go-humanize v1.0.0
	go.uber.org/zap v1.23.0
)

replace github.com/andydunstall/figg/sdk v0.0.0 => ../sdk

require (
	github.com/andydunstall/figg/utils v0.0.0 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
)

replace github.com/andydunstall/figg/utils v0.0.0 => ../utils
