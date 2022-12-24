#!/bin/bash

pushd server
	go run cmd/figg/main.go $@
popd
