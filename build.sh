#!/usr/bin/env bash

# Debug
set -x

# Fails fast
set -eEuo pipefail

# Project is expected to be in $GO_PATH/src/project-name
CURRENT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
GO_PATH="$(go env GOPATH)"
export GO_PATH

GO_ROOT="$(go env GOROOT)"
PATH=${GO_PATH}/bin:${GO_ROOT}/bin:${PATH}

BUILD_DIR="${CURRENT_DIR}/dist"
if [[ -d "${BUILD_DIR}" ]]; then
  rm -rf "${BUILD_DIR}"
fi
mkdir -p "${BUILD_DIR}"

#go get -u github.com/konsorten/go-windows-terminal-sequences

go mod tidy
go mod verify
go generate

TARGET_ARCH="amd64"
for arch in ${TARGET_ARCH}; do
  # For linux be sure to have libgtk-3-dev,
  # libappindicator3-dev, libwebkit2gtk-4.0-dev and libxapp-dev installed
  env GOOS=linux GOARCH="${arch}" go build -o "${BUILD_DIR}"/simple-proxy.linux

  # Embed app.manifest in windows executable
  "${GO_PATH}"/bin/rsrc -arch "${arch}" -manifest app.manifest -o rsrc.syso

  # Windows standard build
  env GOOS=windows GOARCH="${arch}" go build \
    -ldflags -H=windowsgui \
    -o "${BUILD_DIR}"/simple-proxy.exe

  # Windows CLI build
  env GOOS=windows GOARCH="${arch}" go build \
    -o "${BUILD_DIR}"/simple-proxy-cli.exe

  # Mac Os executable can't be cross compiled. We need to be on Mac OS to do this ...
  if [[ "${OSTYPE}" == "darwin"* ]]; then
    go get github.com/machinebox/appify
    env CC=gcc CGO_ENABLED=1 GOOS=darwin GOARCH="${arch}" go build \
      -gccgoflags=-DDARWIN -x objective-c -fobjc-arc -ldflags=framework\ Cocoa \
      -o "${BUILD_DIR}"/simple-proxy
    appify -name "Simple Proxy" -icon ./assets/simple-proxy.png simple-proxy
  fi
done
