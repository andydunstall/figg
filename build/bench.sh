#!/bin/bash

pushd server
	go test ./... -bench=. -benchtime=5s -run=$@ || exit 1
	echo "server ok"
popd
