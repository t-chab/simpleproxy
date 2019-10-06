#!/usr/bin/env bash

# Debug
#set -x

# Fail fast
set -euo pipefail

GOPATH=$(go env GOPATH)

# Build deps
sudo apt -y install build-essential vagrant virtualbox \
  mesa-common-dev libglu1-mesa-dev libpulse-dev libglib2.0-dev libfontconfig1

# Qt dev tools
sudo apt -y --no-install-recommends install "libqt*5-dev" "qt*5-dev" "qml-module-qtquick-*" "qt*5-doc-html"

# Install Qt bindings
go get -u -v -tags=no_env github.com/therecipe/qt/cmd/... &&
  QT_PKG_CONFIG=true "${GOPATH}"/bin/qtsetup

# Fetch / Prepare Docker / VM images for cross compile
docker pull therecipe/qt:linux
docker pull therecipe/qt:windows_64_static
cd "${GOPATH}"/src/github.com/therecipe/qt/internal/vagrant/darwin && vagrant up darwin

echo "Done. You should be able to build Go/Qt apps."
