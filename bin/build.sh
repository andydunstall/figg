#!/bin/bash

pushd server
	go build ./... || exit 1
	echo "server ok"
popd

pushd sdk
	go build ./... || exit 1
	echo "sdk ok"
popd

pushd cli
	go build ./... || exit 1
	echo "cli ok"
popd

pushd fcm/lib
	go build ./... || exit 1
	echo "fcm/lib ok"
popd

pushd fcm/server
	go build ./... || exit 1
	echo "fcm/server ok"
popd

pushd fcm/sdk
	go build ./... || exit 1
	echo "fcm/sdk ok"
popd

pushd fcm/cli
	go build ./... || exit 1
	echo "fcm/cli ok"
popd

pushd tests
	go build ./... || exit 1
	go test -c ./... || exit 1
	echo "tests ok"
popd
