#!/bin/bash

pushd service
	go run cmd/figg/main.go $@
popd
