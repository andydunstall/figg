module github.com/andydunstall/figg/server

go 1.19

require (
	github.com/google/uuid v1.3.0
	github.com/jessevdk/go-flags v1.5.0
	github.com/stretchr/testify v1.8.1
	go.uber.org/zap v1.23.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	golang.org/x/sys v0.0.0-20220111092808-5a964db01320 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

require github.com/andydunstall/figg/utils v0.0.0

replace github.com/andydunstall/figg/utils v0.0.0 => ../utils
