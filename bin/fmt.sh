#!/bin/bash

pushd service
	go fmt ./... || exit 1
	echo "service ok"
popd

pushd sdk
	go fmt ./... || exit 1
	echo "sdk ok"
popd

pushd cli
	go fmt ./... || exit 1
	echo "cli ok"
popd

pushd fcm/service
	go fmt ./... || exit 1
	echo "fcm/service ok"
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
