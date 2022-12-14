#!/bin/bash

pushd server
	go fmt ./... || exit 1
	echo "server ok"
popd

pushd sdk/go
	go fmt ./... || exit 1
	echo "sdk/go ok"
popd

pushd utils
	go fmt ./... || exit 1
	echo "utils ok"
popd

pushd cli
	go fmt ./... || exit 1
	echo "cli ok"
popd

pushd bench
	go fmt ./... || exit 1
	echo "bench ok"
popd

pushd fcm/server
	go fmt ./... || exit 1
	echo "fcm/server ok"
popd

pushd fcm/sdk
	go fmt ./... || exit 1
	echo "fcm/sdk ok"
popd

pushd fcm/cli
	go fmt ./... || exit 1
	echo "fcm/cli ok"
popd

pushd tests
	go fmt ./... || exit 1
	echo "tests ok"
popd
