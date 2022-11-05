module github.com/andydunstall/wombat/tests

go 1.19

require github.com/andydunstall/wombat/sdk v0.0.0

replace github.com/andydunstall/wombat/sdk => ../sdk

require (
	github.com/andydunstall/wombat/wcm/sdk v0.0.0
	github.com/stretchr/testify v1.8.1
	go.uber.org/zap v1.23.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/vmihailenco/msgpack/v5 v5.3.5 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/andydunstall/wombat/wcm/sdk => ../wcm/sdk
