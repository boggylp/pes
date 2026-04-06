.PHONY: all build lint test

all: lint test build

build:
	go build .

lint:
	golangci-lint run ./...

test:
	go test ./...
