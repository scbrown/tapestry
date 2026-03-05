# Contributing to tapestry

## Prerequisites

- Go 1.24+
- [just](https://github.com/casey/just) command runner
- [golangci-lint](https://golangci-lint.run/) for linting
- [pre-commit](https://pre-commit.com/) for git hooks
- Access to a Dolt server (for integration tests)

## Setup

```bash
git clone git@github.com:scbrown/tapestry.git
cd tapestry
just setup           # Install hooks, tidy deps
just build           # Build the binary
just test            # Run tests
```

## Using Just

This project uses [just](https://github.com/casey/just) as a command runner.
**Always prefer `just` commands over raw tool commands** — they're configured
with quiet output by default to save context.

```bash
just --list          # Show available commands
just build           # Build the binary
just test            # Run tests
just check           # Run all quality checks (pre-push gate)
```

## Project Structure

```text
cmd/tapestry/        # CLI entry point (cobra)
internal/
  dolt/              # Dolt connection and queries
  events/            # Gas Town event log parser
  git/               # Git history correlation
  aggregator/        # Summary and rollup logic
  config/            # Configuration management
  web/               # HTTP server and templates
docs/
  plans/             # Planning documents
  design/            # Architecture decisions
  book/              # mdbook user documentation
```

## Development Workflow

1. Check for available work: `bd ready`
2. Claim an issue: `bd update <id> --status in_progress`
3. Implement the change
4. Run quality gates: `just check`
5. Commit with a descriptive message referencing the bead ID
6. Push: `git push`
7. Close the issue: `bd close <id> -m "reason"`

## Quality Gates

All checks must pass before pushing:

```bash
just check           # Runs: lint, test, pre-commit hooks
just docs build      # If you changed documentation
```

## Commit Messages

Follow conventional commits:

```text
feat: add monthly summary aggregation
fix: handle missing Dolt database gracefully
docs: add configuration guide
refactor: extract timeline builder from aggregator
```

## Documentation

Documentation must build cleanly and pass linting:

```bash
just docs build      # Build the book
just docs serve      # Serve locally with hot reload
just docs lint       # Lint markdown files
just docs check      # Full docs quality gate (lint + build)
```

When making user-facing changes, update the relevant documentation under
`docs/book/src/`.
