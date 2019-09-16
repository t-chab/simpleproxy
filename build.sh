#!/usr/bin/env bash

# Project is expected to be in $GOPATH/src/project-name
CURRENT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
GOPATH="$(readlink -f "${CURRENT_DIR}"/../../)"
export GOPATH

GOROOT=$HOME/apps/sdk/go1.13

PATH=$GOPATH/bin:$GOROOT/bin:$PATH

WIN_FLAGS="-ldflags -H=windowsgui"
#WIN_FLAGS=""

go get -u github.com/konsorten/go-windows-terminal-sequences
go get -u github.com/gobuffalo/packr/v2/packr2

packr2 clean
packr2

# For linux be sure to have libgtk-3-dev and libappindicator3-dev installed
env GOOS=linux GOARCH=amd64 go build simple-proxy
env GOOS=linux GOARCH=arm64 go build simple-proxy
env GOOS=windows GOARCH=amd64 go build "${WIN_FLAGS}" simple-proxy

# Mac Os executable can't be cross compiled. We need to be on Mac OS to do this ...
if [[ "$OSTYPE" == "darwin"* ]]; then
    DARWIN_FLAGS="-gccgoflags=\"-DDARWIN -x objective-c -fobjc-arc\" -ldflags=\"framework Cocoa\""
    go get github.com/machinebox/appify
    env CC=gcc CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build ${DARWIN_FLAGS} simple-proxy
    appify -name "Simple Proxy" -icon ./assets/simple-proxy.png simple-proxy
fi
