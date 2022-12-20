#!/bin/bash

pushd service
	go build ./... || exit 1
	echo "service ok"
popd

pushd sdk
	go build ./... || exit 1
	echo "sdk ok"
popd

pushd cli
	go build ./... || exit 1
	echo "cli ok"
popd

pushd fcm/service
	go build ./... || exit 1
	echo "fcm/service ok"
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
