#!/bin/bash

pushd fcm/cli
	go run main.go $@
popd
