package web

import (
	"net/http"
	"time"
)

type sitemapPage struct {
	Name string
	Path string
	Desc string
}

type sitemapCategory struct {
	Name  string
	Pages []sitemapPage
}

type sitemapData struct {
	GeneratedAt time.Time
	Categories  []sitemapCategory
	TotalPages  int
}

func (s *Server) handleSitemap(w http.ResponseWriter, r *http.Request) {
	categories := []sitemapCategory{
		{
			Name: "Overview & Leadership",
			Pages: []sitemapPage{
				{"/", "Home", "Monthly overview dashboard"},
				{"/executive", "Executive", "KPIs, 7-day throughput, agent leaderboard"},
				{"/command-center", "Command Center", "Fleet dashboard"},
				{"/briefing", "Briefing", "Status overview with stale work + unclaimed queue"},
				{"/standup", "Standup", "Daily per-agent activity: done, working, blocked"},
				{"/snapshot", "Snapshot", "Single-screen system health summary"},
				{"/momentum", "Momentum", "Health traffic lights: velocity, flow, blockers, staleness"},
				{"/signals", "Signals", "6 traffic-light health signals"},
				{"/pulse", "Pulse", "24-hour system pulse with hourly breakdown"},
				{"/timeline", "Timeline", "Unified chronological activity feed"},
			},
		},
		{
			Name: "Work Management",
			Pages: []sitemapPage{
				{"/beads", "Beads", "Full bead list with filter bar: status, rig, type, priority, assignee, sort"},
				{"/work", "Work", "Work by repo, priority, or agent mode"},
				{"/kanban", "Kanban", "Column board view"},
				{"/epics", "Epics", "Epic progress overview"},
				{"/queue", "Queue", "Priority-ranked unblocked work with urgency scores"},
				{"/ready", "Ready", "Unblocked work queue with start/close/defer actions"},
				{"/triage", "Triage", "Unassigned + unprioritized beads needing attention"},
				{"/pending", "Pending", "Beads awaiting action: assigned-not-started, unassigned high-pri"},
				{"/focus", "Focus", "Composite urgency scoring: top 30 ranked items"},
				{"/quick-wins", "Quick Wins", "Low-complexity open beads: no deps, easy pickups"},
				{"/blocked", "Blocked", "Dependency bottlenecks + top blockers"},
				{"/search", "Search", "Search all beads"},
			},
		},
		{
			Name: "Agents & Workload",
			Pages: []sitemapPage{
				{"/agents", "Agents", "Agent activity stats with last-active times"},
				{"/owners", "Owners", "Workload by owner with completion bars"},
				{"/assignments", "Assignments", "Per-agent bead list with close/defer/start actions"},
				{"/workload", "Workload", "Agent workload balance: open/blocked/high-pri breakdown"},
				{"/overflow", "Overflow", "Agent overload detector: composite score ranking"},
				{"/wip", "WIP", "Work-in-progress limits per agent"},
				{"/load-balance", "Load Balance", "Composite agent load scoring with status indicators"},
				{"/idle", "Idle", "Agents with no bead updates in 3+ days"},
				{"/contributors", "Contributors", "Agent contribution leaderboard: total/closed/open"},
				{"/swarming", "Swarming", "Multi-agent beads with 2+ agents involved"},
				{"/impact", "Impact", "Agent impact scores: closures, creations, comments weighted"},
				{"/streaks", "Streaks", "Agent activity streaks: consecutive active days"},
			},
		},
		{
			Name: "Flow & Velocity",
			Pages: []sitemapPage{
				{"/velocity", "Velocity", "Throughput dashboard"},
				{"/throughput", "Throughput", "12-week created vs closed with bar charts"},
				{"/flow-rate", "Flow Rate", "30-day daily created vs closed"},
				{"/net-flow", "Net Flow", "30-day cumulative open count trend"},
				{"/scope", "Scope", "30-day cumulative flow diagram"},
				{"/forecast", "Forecast", "Backlog clearance prediction and 4-week trend"},
				{"/pacing", "Pacing", "Backlog clearance tracker: daily close rate needed"},
				{"/agent-velocity", "Agent Velocity", "Per-agent weekly close rate over 4 weeks"},
				{"/funnel", "Funnel", "Conversion funnel: filed to assigned to started to closed"},
				{"/tag-velocity", "Tag Velocity", "Label resolution speed: closed/created per label"},
			},
		},
		{
			Name: "Time & Cycle Analytics",
			Pages: []sitemapPage{
				{"/cycle-time", "Cycle Time", "Open-to-close analytics: median/mean/P90"},
				{"/response-time", "Response Time", "Bead pickup speed: median/mean/P90 by priority"},
				{"/resolution-rate", "Resolution Rate", "Time-to-close distribution bands"},
				{"/retention", "Retention", "Time-in-status analytics for closed beads"},
				{"/sla", "SLA", "Priority-based SLA tracker"},
				{"/dwell", "Dwell", "Per-bead dwell time with zone filters: danger/warning/ok"},
			},
		},
		{
			Name: "Trends & History",
			Pages: []sitemapPage{
				{"/trends", "Trends", "8-week weekly trend analysis"},
				{"/burndown", "Burndown", "30-day open bead trend"},
				{"/burnup", "Burn Up", "30-day cumulative closed beads chart"},
				{"/heatmap", "Heatmap", "91-day activity calendar"},
				{"/calendar", "Calendar", "Monthly activity calendar with intensity shading"},
				{"/cohort", "Cohort", "Weekly cohort close-rate analysis"},
				{"/disposition", "Disposition", "8-week resolution breakdown: closed/deferred/blocked/open"},
				{"/compare", "Compare", "Side-by-side period comparison: 7/14/30 day deltas"},
			},
		},
		{
			Name: "Issue Health & Risks",
			Pages: []sitemapPage{
				{"/risks", "Risks", "Risk radar: stale P0/P1s, unassigned high-pri, long-blocked"},
				{"/watchlist", "Watchlist", "Active P0/P1 ops dashboard with idle time"},
				{"/stale", "Stale", "Stale work detector"},
				{"/parking-lot", "Parking Lot", "Stalled in-progress beads idle 3+ days"},
				{"/debt", "Debt", "Tech debt signals: bug ratio, deferred pile, aging bugs"},
				{"/churn", "Churn", "Status churn detector: beads with 3+ transitions"},
				{"/escalations", "Escalations", "High-priority beads with significant activity"},
				{"/reopen", "Reopen", "Regression tracker: beads closed then reopened"},
				{"/orphans", "Orphans", "Beads missing owner, assignee, labels, or description"},
				{"/duplicates", "Duplicates", "Duplicate bead detection by normalized title"},
			},
		},
		{
			Name: "Distribution & Breakdown",
			Pages: []sitemapPage{
				{"/status", "Status", "System status overview"},
				{"/priorities", "Priorities", "P0-P4 stacked bars + table"},
				{"/types", "Types", "Epic/task/bug breakdown"},
				{"/matrix", "Matrix", "Assignee x status heatmap"},
				{"/rigs", "Rigs", "Per-rig status breakdown"},
				{"/inventory", "Inventory", "Bead counts by status/type/rig"},
				{"/backlog", "Backlog", "Age distribution histogram with median/mean/P90"},
				{"/age-breakdown", "Age Breakdown", "Open bead age distribution by band"},
				{"/priority-drift", "Priority Drift", "Priority health: open/active/blocked/stale per priority"},
				{"/stats", "Stats", "System-wide stats: totals, 7/30-day flow, per-rig breakdown"},
				{"/ratios", "Ratios", "Operational health ratios: bug:feature, close rate, blocker ratio"},
			},
		},
		{
			Name: "Labels & Dependencies",
			Pages: []sitemapPage{
				{"/labels", "Labels", "Label cloud and filter"},
				{"/label-matrix", "Label Matrix", "Labels x status heatmap"},
				{"/label-trends", "Label Trends", "8-week label usage sparklines"},
				{"/label-age", "Label Age", "Median/mean/max age of open beads per label"},
				{"/pair-freq", "Label Pairs", "Label co-occurrence analysis"},
				{"/deps", "Dependencies", "Full dependency graph with type filter"},
				{"/chains", "Chains", "Dependency chain depth analysis: longest chains, top blockers"},
				{"/phase", "Epic Phases", "Epic completion progress with phase bars"},
			},
		},
		{
			Name: "Activity Feeds",
			Pages: []sitemapPage{
				{"/recap", "Recap", "Daily activity digest with date navigation and agent attribution"},
				{"/activity", "Activity", "Recent updates with time window filter"},
				{"/comments", "Comments", "Recent comments across all beads"},
				{"/created", "Created", "Recently filed beads intake feed"},
				{"/closed", "Closed", "Recently completed beads feed"},
				{"/sprint", "Sprint", "Weekly summary with agent breakdown and week nav"},
				{"/changelog", "Changelog", "What shipped — closed beads organized by week"},
				{"/audit-log", "Audit Log", "Global change feed: status changes, creates, comments"},
				{"/status-flow", "Status Flow", "Status transition frequency analysis"},
				{"/dog-pile", "Hot Beads", "High-activity bead tracker by heat score"},
				{"/crossref", "Crossref", "Cross-database bead references"},
				{"/freshness", "Freshness", "Per-database update recency with stale detection"},
				{"/unblocked", "Unblocked", "Recently unblocked beads"},
				{"/transfers", "Transfers", "Bead reassignment log"},
			},
		},
		{
			Name: "Planning & Scheduling",
			Pages: []sitemapPage{
				{"/deferred", "Deferred", "Deferred beads dashboard with aging and reopen/close actions"},
				{"/reschedules", "Reschedules", "Chronic deferral detector: beads deferred 2+ times"},
				{"/gaps", "Gaps", "Coverage gap detector: stale rigs, unassigned priorities"},
				{"/complexity", "Complexity", "Bead complexity ranking by deps, comments, size"},
			},
		},
		{
			Name: "Agent Operations",
			Pages: []sitemapPage{
				{"/handoffs", "Handoffs", "Session handoff history"},
				{"/commits", "Commits", "Git commit feed"},
				{"/events", "Events", "Gas Town event feed"},
			},
		},
		{
			Name: "Infrastructure & Domain",
			Pages: []sitemapPage{
				{"/homelab", "Homelab", "Prometheus + Alertmanager dashboard"},
				{"/probes", "Probes", "Probe findings from docs/probes"},
				{"/designs", "Designs", "Design documents"},
				{"/decisions", "Decisions", "Decision log"},
				{"/achievements", "Achievements", "Unlocked achievements"},
				{"/theme-parks", "Theme Parks", "Trip planning"},
			},
		},
	}

	total := 0
	for _, cat := range categories {
		total += len(cat.Pages)
	}

	data := sitemapData{
		GeneratedAt: time.Now(),
		Categories:  categories,
		TotalPages:  total,
	}

	s.render(w, r, "sitemap", data)
}
