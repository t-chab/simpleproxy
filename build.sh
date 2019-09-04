#!/usr/bin/env sh

GOPATH=$HOME/projects/go
GOROOT=$HOME/apps/sdk/go1.13

GOARCH=amd64
GOOS=linux

PATH=$GOROOT/bin:$PATH

go build simple-proxy
