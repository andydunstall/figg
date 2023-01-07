#!/bin/bash

pushd server
	go test ./... $@ || exit 1
	echo "server ok"
popd

pushd sdk/go
	go test ./... $@ || exit 1
	echo "sdk/go ok"
popd

pushd utils
	go test ./... $@ || exit 1
	echo "utils ok"
popd

pushd tests
	go test ./... $@ || exit 1
	echo "system tests ok"
popd
