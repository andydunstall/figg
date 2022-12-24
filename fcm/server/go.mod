module github.com/andydunstall/figg/fcm/server

go 1.19

require (
	github.com/google/uuid v1.3.0
	github.com/gorilla/mux v1.8.0
	go.uber.org/zap v1.23.0
)

require (
	github.com/jessevdk/go-flags v1.5.0 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	golang.org/x/sys v0.0.0-20220909162455-aba9fc2a8ff2 // indirect
)

require github.com/andydunstall/figg/server v0.0.0

replace github.com/andydunstall/figg/server v0.0.0 => ../../server

require github.com/andydunstall/figg/utils v0.0.0 // indirect

replace github.com/andydunstall/figg/utils v0.0.0 => ../../utils

require github.com/andydunstall/figg/fcm/lib v0.0.0

replace github.com/andydunstall/figg/fcm/lib v0.0.0 => ../lib
