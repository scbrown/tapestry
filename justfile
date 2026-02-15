# tapestry — the archivist for your agent fleet
# Run `just --list` to see available recipes

# Quiet by default to save context; use verbose=true for full output
verbose := "false"

# Default recipe - show available commands
default:
    @just --list

# === Setup ===

# Install pre-commit hooks and verify dependencies
setup:
    pre-commit install
    go mod tidy
    @echo "Setup complete."

# === Build ===

# Build the tapestry binary
build:
    go build -o tapestry ./cmd/tapestry

# Build and install to ~/.local/bin
install: build
    install -m 755 tapestry ~/.local/bin/tapestry

# === Quality ===

# Run all tests
test:
    go test ./...

# Run linter
lint:
    golangci-lint run ./...

# Run all quality checks (pre-push gate)
check: lint test
    pre-commit run --all-files

# === Development ===

# Build and run the server
dev: build
    ./tapestry serve

# === Documentation ===

# Documentation management: just docs <cmd>
# Commands: build, serve, lint, fix, fmt, vale, check

docs cmd="build":
    #!/usr/bin/env bash
    set -euo pipefail
    case "{{cmd}}" in
        build)    mdbook build docs/book ;;
        serve)    mdbook serve docs/book --open ;;
        lint)     npx markdownlint-cli2 "docs/book/src/**/*.md" "README.md" "CONTRIBUTING.md" ;;
        fix)      npx markdownlint-cli2 --fix "docs/book/src/**/*.md" "README.md" "CONTRIBUTING.md" ;;
        fmt)      npx prettier --write "docs/book/src/**/*.md" --prose-wrap preserve ;;
        vale)     vale docs/book/src/ ;;
        check)    just docs lint && just docs build ;;
        *)        echo "Unknown: {{cmd}}. Try: build serve lint fix fmt vale check" ;;
    esac
