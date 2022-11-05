module github.com/andydunstall/wombat/wcm/service

go 1.19

require (
	github.com/Shopify/toxiproxy/v2 v2.5.0
	github.com/google/uuid v1.3.0
	github.com/gorilla/mux v1.8.0
	go.uber.org/zap v1.23.0
)

require (
	github.com/andydunstall/scuttlebutt v0.0.0-20221028180859-55a7dfdb0cfd // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/hashicorp/errwrap v1.0.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/jessevdk/go-flags v1.5.0 // indirect
	github.com/spaolacci/murmur3 v1.1.0 // indirect
	github.com/vmihailenco/msgpack/v5 v5.3.5 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	golang.org/x/sys v0.0.0-20220909162455-aba9fc2a8ff2 // indirect
)

require github.com/andydunstall/wombat/service v0.0.0

replace github.com/andydunstall/wombat/service v0.0.0 => ../../service
