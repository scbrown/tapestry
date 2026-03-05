[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

# tapestry

**The archivist for your agent fleet.** Drillable event log and retrospective
intelligence over Dolt-backed beads, across multiple Gas Town workspaces.

Agents create beads, close beads, sling work, spawn polecats, and hand off
context — but nobody's weaving the threads together. Tapestry does.

## What It Does

```text
February 2026 Summary
├── 47 beads closed, 23 created, 8 epics advanced
├── Reliability Foundation (aegis-gzl) ─── 85% → 100%
│   ├── Lock contention (x1e) ── CLOSED Feb 14
│   ├── Baseline integrity (gzl.2) ── CLOSED Feb 14
│   └── Canary validation (gzl.4) ── CLOSED Feb 14
├── Discovery Phase ─── 40% → 75%
│   ├── Service map (fch4) ── CLOSED Feb 13
│   └── TLS cert audit (poo0) ── IN PROGRESS
├── Agent Activity
│   ├── goldblum: 12 closed, 8 created (planner)
│   └── malcolm: 9 closed (executor)
└── Click any item to drill down →
```

Start at the month level. Drill into epics. Drill into individual beads. See
the full timeline of who did what, when, and why — all the way down to git
commits and agent handoff notes.

## Why Tapestry?

| Question | Without Tapestry | With Tapestry |
|----------|-----------------|---------------|
| "What did we accomplish this month?" | `bd list --status=closed` + manual counting | One-click monthly summary |
| "How did the cert audit go from discovery to resolution?" | Read 15 bead comment threads | Epic timeline with drill-down |
| "Which agents are most productive?" | Grep through `.events.jsonl` | Agent activity dashboard |
| "What's the arc of this epic?" | `bd show` each child bead | Visual progress tree |
| "Show me cross-rig activity" | Check each rig separately | Unified multi-gastown view |

## Quick Start

```bash
# Build
just build

# Configure Dolt connection
tapestry config set dolt.host dolt.svc
tapestry config set dolt.port 3306

# Add Gas Town workspaces to monitor
tapestry workspace add ~/gt --name homelab

# Start the dashboard
tapestry serve
# → http://localhost:8070
```

## Features

- **Monthly/weekly summaries** — auto-generated retrospectives with completion stats
- **Epic drill-down** — tree view from epic → child beads → comments → commits
- **Agent activity** — who closed what, dispatch patterns, handoff frequency
- **Multi-gastown** — monitor multiple Gas Town workspaces from one dashboard
- **Event timeline** — chronological view of slings, hooks, dones, spawns, kills
- **Dolt-native** — queries Dolt directly via MySQL protocol, no other backends
- **Dolt time-travel** — compare state at any two points in time
- **HTMX frontend** — server-rendered, fast, no JavaScript framework bloat

## Architecture

```text
┌─────────────────────────────────────────────┐
│            tapestry serve                    │
│  Go HTTP server + HTMX templates            │
├─────────────────────────────────────────────┤
│  Aggregation Layer                          │
│  - Monthly/weekly rollups                   │
│  - Epic progress calculation                │
│  - Agent activity attribution               │
│  - Cross-rig correlation                    │
├──────────┬──────────────┬───────────────────┤
│ Dolt SQL │ Events JSONL │ Git History       │
│ beads_*  │ .events.jsonl│ commit messages   │
│ dolt.svc │ per workspace│ bead ID refs      │
└──────────┴──────────────┴───────────────────┘
```

## Data Sources

| Source | What | How |
|--------|------|-----|
| **Dolt** (beads databases) | Issues, epics, comments, deps, status changes | MySQL protocol to dolt.svc |
| **Dolt diff** | State changes over time (time travel) | `dolt_diff()` table function |
| **Events JSONL** | Agent activity (sling, hook, done, spawn, kill, handoff) | Read from Gas Town workspace |
| **Git commits** | Code changes linked to beads via commit messages | Parse bead IDs from messages |

## Development

This project uses [just](https://github.com/casey/just) as a command runner.

```bash
just --list          # Show available commands
just setup           # Install pre-commit hooks + Go deps
just build           # Build the binary
just test            # Run tests
just check           # Run all quality checks
just dev             # Build + serve with auto-reload
```

## Documentation

```bash
just docs build      # Build the book
just docs serve      # Serve locally with hot reload
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup and guidelines.

## License

[MIT](LICENSE)
