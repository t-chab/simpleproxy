PKG:=github.com/tchabaud/simpleproxy
APP_NAME:=simpleproxy
PWD:=$(shell pwd)
UID:=$(shell id -u)
VERSION:=$(shell git describe --tags --always --dirty="-dev")
GOOS:=$(shell go env GOOS)
LDFLAGS:=-X main.Version=$(VERSION) -w -s
GOOS:=$(strip $(shell go env GOOS))
GOARCHs:=$(strip $(shell go env GOARCH))
GOVERSION:=1.16.4

# Embed app.manifest in windows executable
ifeq "$(GOOS)" "windows"
SUFFIX=.exe
endif

ifeq "$(GOOS)" "darwin"
GOARCHs=amd64 arm64
endif

# CGO must be enabled
export CGO_ENABLED:=1

build: fmt vet
	$(foreach GOARCH,$(GOARCHs),$(shell GOARCH=$(GOARCH) go build -mod=vendor -ldflags="$(LDFLAGS)" -trimpath -o bin/$(APP_NAME)_$(GOOS)_$(GOARCH)$(SUFFIX) .))

docker:
	docker pull golang:$(GOVERSION)
	docker run -ti --rm -e GOCACHE=/tmp -v $(PWD):/$(APP_NAME) -u $(UID):$(UID) --workdir /$(APP_NAME) golang:$(GOVERSION) make

fmt:
	gofmt -s -w ./

vet:
	go vet -mod=vendor ./...

static:
	staticcheck ./

mod:
	go mod vendor

test:
	go test -v ./