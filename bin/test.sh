#!/bin/bash

pushd service
	go test ./... || exit 1
	echo "service ok"
popd

pushd sdk
	go test ./... || exit 1
	echo "sdk ok"
popd