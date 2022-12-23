#!/bin/bash

pushd service
	go test ./... $@ || exit 1
	echo "service ok"
popd

pushd sdk
	go test ./... $@ || exit 1
	echo "sdk ok"
popd

pushd utils
	go test ./... $@ || exit 1
	echo "utils ok"
popd

pushd tests
	go test ./... $@ || exit 1
	echo "system tests ok"
popd
