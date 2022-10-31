module github.com/andydunstall/wombat/sdk

go 1.19

require (
	github.com/gorilla/websocket v1.5.0
	github.com/vmihailenco/msgpack/v5 v5.3.5
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	golang.org/x/net v0.0.0-20201021035429-f5854403a974 // indirect
	golang.org/x/sys v0.0.0-20210119212857-b64e53b001e4 // indirect
	golang.org/x/text v0.3.3 // indirect
	google.golang.org/genproto v0.0.0-20200526211855-cb27e3aa2013 // indirect
	google.golang.org/grpc v1.50.1 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

require (
	github.com/andydunstall/wombat/wcm/sdk v0.0.0
	github.com/stretchr/testify v1.8.1
	go.uber.org/zap v1.23.0
)

replace github.com/andydunstall/wombat/wcm/sdk v0.0.0 => ../wcm/sdk
