#!/bin/bash

pushd fcm/service
	go run cmd/fcm/main.go
popd
