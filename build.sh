#!/bin/bash -e

if [[ $1 = "-loc" ]]; then
    find . -name '*.go' | xargs wc -l | sort -n
    exit
fi

BuildId=$(git rev-parse --short HEAD)

BUILD_FLAGS=''
if [[ $1 = "-race" ]]; then
    BUILD_FLAGS="$BUILD_FLAGS -race"
fi
if [[ $1 = "-gc" ]]; then
    BUILD_FLAGS="$BUILD_FLAGS -gcflags '-m=1'"
fi

go build $BUILD_FLAGS -tags release -ldflags "-X BuildId $BuildId -w"
