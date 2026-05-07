.PHONY: all build clean install run linux darwin windows

BINARY_NAME := project-context
VERSION     ?= v0.3.0

build:
	go build -ldflags "-s -w -X main.Version=$(VERSION)" -o $(BINARY_NAME) ./cmd/project-context

install:
	go install -ldflags "-s -w -X main.Version=$(VERSION)" ./cmd/project-context

run:
	go run ./cmd/project-context

all: linux darwin windows

linux:
	mkdir -p bin
	GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o bin/$(BINARY_NAME)-linux-amd64 ./cmd/project-context
	GOOS=linux GOARCH=arm64 go build -ldflags "-s -w" -o bin/$(BINARY_NAME)-linux-arm64 ./cmd/project-context

darwin:
	mkdir -p bin
	GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w" -o bin/$(BINARY_NAME)-darwin-amd64 ./cmd/project-context
	GOOS=darwin GOARCH=arm64 go build -ldflags "-s -w" -o bin/$(BINARY_NAME)-darwin-arm64 ./cmd/project-context

windows:
	mkdir -p bin
	GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o bin/$(BINARY_NAME)-windows-amd64.exe ./cmd/project-context

clean:
	rm -rf bin/ $(BINARY_NAME) $(BINARY_NAME).exe