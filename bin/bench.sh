#!/bin/bash

pushd service
	go test ./... -bench=. -benchtime=5s -run=$@ || exit 1
	echo "service ok"
popd
