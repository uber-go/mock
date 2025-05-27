BIN = bin

export GO111MODULE=on
export GOBIN ?= $(shell pwd)/$(BIN)

GO_FILES = $(shell find . \
	   -path '*/.*' -prune -o \
	   '(' -type f -a -name '*.go' ')' -print)

EXTRACT_CHANGELOG = $(BIN)/extract-changelog

.PHONY: all
all: build test

.PHONY: build
build:
	go build ./...

.PHONY: test
test:
	go test -v -race ./...

.PHONY: cover
cover:
	go test -race -coverprofile=cover.out -coverpkg=./... ./...
	go tool cover -html=cover.out -o cover.html

$(EXTRACT_CHANGELOG): tools/cmd/extract-changelog/main.go
	cd tools && go install github.com/uber-go/mock/tools/cmd/extract-changelog
