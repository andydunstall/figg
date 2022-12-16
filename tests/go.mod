module github.com/andydunstall/figg/tests

go 1.19

require github.com/andydunstall/figg/sdk v0.0.0

replace github.com/andydunstall/figg/sdk => ../sdk

require (
	github.com/andydunstall/figg/fcm/sdk v0.0.0
	github.com/stretchr/testify v1.8.1
	go.uber.org/zap v1.24.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/andydunstall/figg/fcm/sdk => ../fcm/sdk

require github.com/andydunstall/figg/utils v0.0.0 // indirect

replace github.com/andydunstall/figg/utils v0.0.0 => ../utils

require github.com/andydunstall/figg/sdkv2 v0.0.0

replace github.com/andydunstall/figg/sdkv2 => ../sdkv2
