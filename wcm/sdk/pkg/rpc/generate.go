package rpc

//go:generate protoc -I ../../../rpc --go_out=plugins=grpc,paths=source_relative:. rpc.proto --go_opt=Mrpc.proto=github.com/andydunstall/wombat/wcm/sdk/pkg/rpc
