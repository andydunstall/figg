module github.com/andydunstall/figg/fcm/service

go 1.19

require (
	github.com/google/uuid v1.3.0
	github.com/gorilla/mux v1.8.0
	go.uber.org/zap v1.23.0
)

require (
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/jessevdk/go-flags v1.5.0 // indirect
	github.com/vmihailenco/msgpack/v5 v5.3.5 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	golang.org/x/sys v0.0.0-20220909162455-aba9fc2a8ff2 // indirect
)

require github.com/andydunstall/figg/service v0.0.0

replace github.com/andydunstall/figg/service v0.0.0 => ../../service
