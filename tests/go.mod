module github.com/andydunstall/figg/tests

go 1.19

require github.com/andydunstall/figg/sdk/go v0.0.0

replace github.com/andydunstall/figg/sdk/go => ../sdk/go

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

require (
	github.com/andydunstall/figg/utils v0.0.0 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/jessevdk/go-flags v1.5.0 // indirect
	golang.org/x/sys v0.0.0-20220909162455-aba9fc2a8ff2 // indirect
)

replace github.com/andydunstall/figg/utils v0.0.0 => ../utils

require github.com/andydunstall/figg/server v0.0.0 // indirect

replace github.com/andydunstall/figg/server v0.0.0 => ../server

require github.com/andydunstall/figg/fcm/lib v0.0.0

replace github.com/andydunstall/figg/fcm/lib v0.0.0 => ../fcm/lib
