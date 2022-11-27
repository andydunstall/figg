#!/bin/bash

pushd service
	go build ./... || exit 1
	echo "service ok"
popd

pushd sdk
	go build ./... || exit 1
	echo "sdk ok"
popd

pushd fcm/sdk
	go build ./... || exit 1
	echo "fcm/sdk ok"
popd

pushd fcm/service
	go build ./... || exit 1
	echo "fcm/service ok"
popd

