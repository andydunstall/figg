#!/bin/bash

if ! pgrep "toxiproxy-ser" > /dev/null
then
    echo "toxiproxy-server not running"
	exit 1
fi

pushd fcm/service
	go run cmd/fcm/main.go
popd
