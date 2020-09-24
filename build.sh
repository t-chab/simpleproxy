#!/usr/bin/env bash

# Debug
#set -x

# Fail fast
set -euo pipefail

GOPATH=$(go env GOPATH)

QTTOOLDIR="/usr/lib/qt5/bin"
QTLIBDIR="/usr/lib/x86_64-linux-gnu"
QT_PKG_CONFIG=true

# For linux be sure to have libgtk-3-dev and libappindicator3-dev installed
#env GOOS=linux GOARCH=amd64 go build simple-proxy
qtdeploy -docker build linux

#env GOOS=linux GOARCH=arm64 go build simple-proxy

#env GOOS=darwin GOARCH=amd64 go build simple-proxy
#cd $(go env GOPATH)/src/github.com/therecipe/qt/internal/vagrant/darwin && vagrant up darwin
#qtdeploy -vagrant build darwin/darwin

#env GOOS=windows GOARCH=amd64 go build -ldflags -H=windowsgui simple-proxy
#qtdeploy -docker build windows_64_static
