# tapestry - Agent Instructions

## Project Overview

Tapestry is an archivist dashboard for Gas Town agent fleets. It reads from
Dolt-backed beads databases and Gas Town event logs to produce drillable
retrospective views: monthly summaries, epic timelines, agent activity, and
event streams.

**Language:** Go
**Backend:** Dolt (MySQL protocol) — no other beads backends
**Frontend:** HTMX + Go html/template — no SPA frameworks

## Conventions

- **Always use `just` instead of raw commands.** The justfile is configured with
  quiet output by default to save context — you only see errors and warnings.
- **Prefer subcommands over separate recipes.** Group related operations under a
  single recipe with a subcommand argument (e.g., `just docs build`).
- **Standard Go project layout.** `cmd/tapestry/` for the binary entry point,
  `internal/` for all packages. Nothing in `internal/` is exported.
- **Dolt-only.** Never add SQLite, JSONL, or other storage backends. Dolt is
  the single source of truth for beads data.
- **Read-only.** Tapestry never writes to beads databases or event logs.

## Build Commands

```bash
just --list          # Show available commands
just setup           # Install pre-commit hooks + Go deps
just build           # Build the binary
just test            # Run tests
just lint            # Run golangci-lint
just check           # Full quality gate (lint + test)
just install         # Build and install to ~/.local/bin
just dev             # Build + serve (development)
```

For verbose output when debugging:

```bash
just check verbose=true
```

## Documentation Commands

```bash
just docs build      # Build the book
just docs serve      # Serve locally with hot reload
just docs lint       # Lint markdown files
just docs check      # Full docs quality gate
```

## Project Structure

```
cmd/tapestry/        # CLI entry point (cobra)
internal/
  dolt/              # Dolt connection, queries, time-travel
  events/            # .events.jsonl parser
  git/               # Git log parser, bead ID extraction
  aggregator/        # Summary/rollup logic
  config/            # TOML config management
  web/               # HTTP handlers, HTMX templates
    templates/       # Go html/template files
    static/          # CSS, JS (minimal)
docs/
  plans/             # Planning documents
  design/            # Architecture decisions
  book/              # mdbook documentation
```

## Quality Requirements

### Before Every Push

You MUST run and pass the full quality gate before pushing:

```bash
just check
```

This runs:

- `golangci-lint` (Go linting)
- `go test ./...` (all tests)
- Pre-commit hooks (whitespace, YAML, merge conflicts, markdown lint)

**Do NOT push if any check fails.** Fix the issues and re-run.

### Test Requirements

- All existing tests must pass before pushing
- New functionality must include corresponding tests
- Tests are part of the `just check` quality gate

### Documentation Requirements

- User-facing changes MUST include documentation updates
- Run `just docs build` to verify the book builds cleanly
- Update README.md if the change affects quick-start or usage

## Issue Tracking

This project uses beads for issue tracking.

```bash
bd ready              # Find available work
bd show <id>          # View issue details
bd update <id> --status in_progress  # Claim work
bd close <id> -m "reason"            # Complete work
```

## Landing the Plane (Session Completion)

**When ending a work session**, complete ALL steps below. Work is NOT complete
until `git push` succeeds.

1. **Run quality gates** — `just check` must pass
2. **Build docs** — `just docs build` must succeed (if docs changed)
3. **Commit and push**:
   ```bash
   git add <files>
   git commit -m "<type>: <description>"
   git push
   ```
4. **Verify** — All changes committed AND pushed

**CRITICAL RULES:**

- Work is NOT complete until `git push` succeeds
- NEVER stop before pushing — that leaves work stranded locally
- NEVER say "ready to push when you are" — YOU must push
- If push fails, resolve and retry until it succeeds
