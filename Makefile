# Makefile for Veda Engine

VERSION ?= $(shell git describe --tags --always --dirty --first-parent 2>/dev/null || echo "dev")

.PHONY: all build fmt lint clean

all: build
build:
	@echo "Building Veda Engine for windows..."
	CGO_ENABLED=1 CC="zig cc -target x86_64-windows-gnu -Wl,--subsystem,windows" GOOS=windows GOARCH=amd64 go build -ldflags="-w -s -X main.Version=$(VERSION)" -o ./bin/veda-engine.exe ./src/

fmt:
	@echo "Formatting code..."
	go fmt ./...

lint:
	CGO_ENABLED=1 CC="zig cc -target x86_64-windows-gnu" GOOS=windows golangci-lint run

clean:
	@echo "Cleaning..."
	rm -rf ./bin/veda-engine.exe

