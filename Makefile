BINARY := tapestry
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X main.version=$(VERSION)"
INSTALL_DIR := $(HOME)/.local/bin

.PHONY: build test lint check install clean

build:
	go build $(LDFLAGS) -o $(BINARY) ./cmd/tapestry

test:
	go test ./...

lint:
	golangci-lint run ./...

check: lint test

install: build
	install -m 755 $(BINARY) $(INSTALL_DIR)/$(BINARY)

clean:
	rm -f $(BINARY)

dev: build
	./$(BINARY) serve
