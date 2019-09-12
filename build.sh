#!/usr/bin/env bash

# Project is expected to be in $GOPATH/src/project-name
CURRENT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
GOPATH="$(readlink -f "${CURRENT_DIR}"/../../)"

GOROOT=$HOME/apps/sdk/go1.13

PATH=$GOROOT/bin:$PATH

# For linux be sure to have libgtk-3-dev and libappindicator3-dev installed
env GOOS=linux GOARCH=amd64 go build simple-proxy
env GOOS=linux GOARCH=arm64 go build simple-proxy
env GOOS=darwin GOARCH=amd64 go build simple-proxy
env GOOS=windows GOARCH=amd64 go build -ldflags -H=windowsgui simple-proxy
