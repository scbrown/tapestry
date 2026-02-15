# Tapestry Architecture

**Status:** Design
**Date:** 2026-02-14

## Overview

Tapestry is an archivist dashboard for Gas Town agent fleets. It aggregates data
from Dolt-backed beads databases, Gas Town event logs, and git history to provide
drillable retrospective views of what happened, when, and why.

## Design Principles

1. **Dolt-native** — No SQLite, no JSONL beads backends. Dolt is the single source of truth.
2. **Multi-workspace** — Monitor multiple Gas Town installations from one dashboard.
3. **Read-only** — Tapestry never writes to beads or events. Pure observer.
4. **Server-rendered** — HTMX + Go templates. No SPA, no build step, no node_modules.
5. **Time-travel first** — Leverage Dolt's versioning for "what changed" queries.

## Data Model

### Sources

```
Dolt Server (dolt.svc:3306)
├── beads_aegis          ← issues, comments, deps for aegis rig
├── beads_gastown         ← issues for gastown rig
├── beads_bobbin          ← issues for bobbin rig
└── ... (one database per rig)

Gas Town Workspace (~gt/)
├── .events.jsonl         ← agent activity events (sling, hook, done, spawn, kill)
├── aegis/mayor/rig/.beads/  ← rig-level beads config
└── gastown/mayor/rig/.beads/

Git Repos
├── commit messages referencing bead IDs
└── author/timestamp for commit↔bead correlation
```

### Aggregation Layers

```
Raw Data (Dolt + Events + Git)
        │
        ▼
┌─────────────────────────┐
│  Snapshot Layer          │  Point-in-time state from Dolt
│  - Issue counts by status│  Uses: AS OF <timestamp>
│  - Epic completion %     │  Uses: dolt_diff()
│  - Agent assignments     │
└───────────┬─────────────┘
            │
            ▼
┌─────────────────────────┐
│  Timeline Layer          │  Chronological event stream
│  - Status transitions    │  Correlates: Dolt diff + events + git
│  - Agent activity        │
│  - Handoff chains        │
└───────────┬─────────────┘
            │
            ▼
┌─────────────────────────┐
│  Summary Layer           │  Rolled-up narratives
│  - Monthly digest        │  Groups by: epic, agent, rig, week
│  - Epic arcs             │
│  - Agent scorecards      │
└─────────────────────────┘
```

## Pages / Views

### 1. Monthly Summary (landing page)

```
/                           → current month
/2026/02                    → February 2026
/2026/02?rig=aegis          → filtered by rig
```

Shows:
- Headline stats (created, closed, in-progress, epic advancement)
- Top completions (epics that progressed most)
- Agent activity table (who did what)
- Notable events (P0 alerts, escalations, mass deaths)
- Drill-down links to each section

### 2. Epic View

```
/epic/aegis-gzl             → single epic with full tree
```

Shows:
- Epic metadata (title, priority, dates)
- Child bead tree with status badges
- Progress over time (chart or timeline)
- Activity feed filtered to this epic
- Dependency graph (which beads blocked which)

### 3. Bead Detail

```
/bead/aegis-0a9             → single bead
```

Shows:
- Full bead metadata
- Comment thread (chronological)
- Status transition history (from Dolt diff)
- Related git commits (parsed from commit messages)
- Blocking/blocked-by relationships

### 4. Agent View

```
/agent/aegis/crew/goldblum  → single agent
/agents                     → all agents
```

Shows:
- Beads created, closed, commented
- Dispatch patterns (what rigs, what priorities)
- Handoff frequency and session lengths
- Current assignments

### 5. Event Timeline

```
/events                     → full event stream
/events?type=done           → filtered by type
/events?agent=goldblum      → filtered by agent
```

Shows:
- Chronological event feed (sling, hook, done, spawn, kill, handoff, etc.)
- Filterable by type, agent, rig, time range
- Correlated with bead changes (e.g., "done" event → bead closed)

### 6. Cross-Rig Dashboard

```
/rigs                       → all rigs across all workspaces
```

Shows:
- Per-rig summary (open beads, active agents, recent closures)
- Cross-rig activity patterns
- Unified view across multiple Gas Town installations

## Tech Stack

| Component | Choice | Rationale |
|-----------|--------|-----------|
| Language | Go | Same as beads, gastown; proven ecosystem |
| HTTP | net/http | Standard library, no framework needed |
| Templates | html/template | Server-rendered, embedded in binary |
| Frontend | HTMX + minimal CSS | No build step, proven in gt dashboard |
| Database | Dolt via MySQL protocol | go-sql-driver/mysql, standard SQL |
| CLI | cobra | Same as beads, gastown |
| Config | TOML | Same as beads, gastown |
| Embedding | go:embed | Templates + static assets in binary |

## Dolt Queries

### Time Travel

```sql
-- State at a specific point in time
SELECT * FROM issues AS OF '2026-02-01T00:00:00'
WHERE status = 'closed';

-- What changed between two dates
SELECT * FROM dolt_diff('issues', '2026-02-01', '2026-02-14')
WHERE diff_type IN ('added', 'modified');

-- Status transitions for a specific bead
SELECT from_status, to_status, from_commit, to_commit
FROM dolt_diff('issues', 'main~30', 'main')
WHERE to_id = 'aegis-0a9' AND from_status != to_status;
```

### Cross-Database Queries

```sql
-- Dolt supports USE <database> for switching
-- Connect to each beads_* database and aggregate
USE beads_aegis;
SELECT 'aegis' as rig, count(*) as open FROM issues WHERE status = 'open';
```

## Configuration

```toml
# ~/.config/tapestry/config.toml

[server]
host = "localhost"
port = 8070

[dolt]
host = "dolt.svc"
port = 3306
user = "root"

# Multiple workspaces supported
[[workspace]]
name = "homelab"
path = "/home/braino/gt"
databases = ["beads_aegis", "beads_gastown", "beads_bobbin"]

[[workspace]]
name = "work"
path = "/home/user/work-gt"
databases = ["beads_projectx"]
```

## Implementation Phases

### Phase 1: Core (MVP)

- [ ] CLI scaffolding (cobra: serve, config, workspace)
- [ ] Dolt connection and query layer
- [ ] Monthly summary page (headline stats, top completions)
- [ ] Bead detail page (metadata, comments, status history)
- [ ] HTMX templates with minimal CSS

### Phase 2: Drill-Down

- [ ] Epic tree view with progress
- [ ] Agent activity view
- [ ] Event timeline (read .events.jsonl)
- [ ] Status transition history via Dolt diff
- [ ] Filtering by rig, agent, priority, date range

### Phase 3: Intelligence

- [ ] Git commit correlation (parse bead IDs from messages)
- [ ] Handoff chain reconstruction
- [ ] Cross-rig unified dashboard
- [ ] Weekly/monthly auto-generated digest (markdown export)
- [ ] Trend charts (completion velocity, backlog growth)

### Phase 4: Polish

- [ ] Responsive design (mobile-friendly)
- [ ] Keyboard navigation
- [ ] Search across all beads and events
- [ ] Bookmarkable filtered views
- [ ] Export to markdown/PDF
