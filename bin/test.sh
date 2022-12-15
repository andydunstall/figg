#!/bin/bash

pushd service
	go test ./... || exit 1
	echo "service ok"
popd

pushd sdk
	go test ./... || exit 1
	echo "sdk ok"
popd

pushd sdkv2
	go test ./... || exit 1
	echo "sdkv2 ok"
popd

pushd utils
	go test ./... || exit 1
	echo "utils ok"
popd
