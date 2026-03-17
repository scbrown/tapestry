package web

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/scbrown/tapestry/internal/dolt"
	"github.com/scbrown/tapestry/internal/events"
)

//go:embed templates/*.html
var templateFS embed.FS

//go:embed static/*
var staticFS embed.FS

// DataSource provides read access to beads data.
type DataSource interface {
	ListBeadsDatabases(ctx context.Context) ([]dolt.DatabaseInfo, error)
	CountByStatus(ctx context.Context, database string) (map[string]int, error)
	CountCreatedInRange(ctx context.Context, database string, start, end time.Time) (int, error)
	CountClosedInRange(ctx context.Context, database string, start, end time.Time) (int, error)
	AgentActivityInRange(ctx context.Context, database string, from, to time.Time) (map[string]int, error)
	Issues(ctx context.Context, database string, f dolt.IssueFilter) ([]dolt.Issue, error)
	IssueByID(ctx context.Context, database, id string) (*dolt.Issue, error)
	Comments(ctx context.Context, database, issueID string) ([]dolt.Comment, error)
	Dependencies(ctx context.Context, database, issueID string) ([]dolt.Dependency, error)
	SearchIssues(ctx context.Context, database, q string, limit int) ([]dolt.Issue, error)
	DistinctAssignees(ctx context.Context, database string) ([]string, error)
	BlockedIssues(ctx context.Context, database string) ([]dolt.BlockedIssue, error)
	AgentActivity(ctx context.Context, database string) ([]dolt.AgentStats, error)
	Decisions(ctx context.Context, database string) ([]dolt.Issue, error)
	LabelsForIssue(ctx context.Context, database, issueID string) ([]string, error)
	MetadataForIssue(ctx context.Context, database, issueID string) (*dolt.IssueMetadata, error)
	StatusHistory(ctx context.Context, database, issueID string) ([]dolt.StatusTransition, error)
	ChildIssues(ctx context.Context, database, parentID string) ([]dolt.Issue, error)
	AchievementDefs(ctx context.Context, database string) ([]dolt.AchievementDef, error)
	AchievementUnlocks(ctx context.Context, database string) ([]dolt.AchievementUnlock, error)
	Epics(ctx context.Context, database string) ([]dolt.Issue, error)
	AllChildDependencies(ctx context.Context, database string) ([]dolt.Dependency, error)
	AddComment(ctx context.Context, database, issueID, author, body string) error
	UpdateStatus(ctx context.Context, database, issueID, status string) error
	UpdatePriority(ctx context.Context, database, issueID string, priority int) error
	UpdateAssignee(ctx context.Context, database, issueID, assignee string) error
	UpdateTitle(ctx context.Context, database, issueID, title string) error
	UpdateDescription(ctx context.Context, database, issueID, description string) error
	AddLabel(ctx context.Context, database, issueID, label string) error
	RemoveLabel(ctx context.Context, database, issueID, label string) error
	ThemeParks(ctx context.Context, database string) ([]dolt.ThemePark, error)
	Rides(ctx context.Context, database, parkID string) ([]dolt.Ride, error)
	ParkVisits(ctx context.Context, database, parkID string) ([]dolt.ParkVisit, error)
	TripPlans(ctx context.Context, database string) ([]dolt.TripPlan, error)
	DistinctLabels(ctx context.Context, database string) ([]dolt.LabelCount, error)
	IssuesByLabel(ctx context.Context, database, label string) ([]dolt.Issue, error)
	DependenciesWithIssues(ctx context.Context, database, issueID string) ([]dolt.DepEdge, error)
	AllDependenciesWithIssues(ctx context.Context, database string) ([]dolt.DepEdge, error)
	CountByPriorityStatus(ctx context.Context, database string) ([]dolt.PriorityStatusCount, error)
	CountByAssigneeStatus(ctx context.Context, database string) ([]dolt.AssigneeStatusCount, error)
	RecentComments(ctx context.Context, database string, limit int) ([]dolt.Comment, error)
	IssueDiffSince(ctx context.Context, database string, since time.Time) ([]dolt.IssueDiffRow, error)
	CommentDiffSince(ctx context.Context, database string, since time.Time) ([]dolt.CommentDiffRow, error)
}

// Server serves the Tapestry web dashboard.
type Server struct {
	ds            DataSource
	prom          *promClient
	forgejo       *forgejoClient
	templates     map[string]*template.Template
	static        http.Handler
	workspacePath string // path to Gas Town workspace root (for events)

	dbMu    sync.Mutex
	dbCache []dolt.DatabaseInfo
	dbExp   time.Time
}

const dbCacheTTL = 5 * time.Minute

// databases returns all known beads databases, caching the discovery result
// for dbCacheTTL to avoid repeated SHOW DATABASES + probe queries per request.
func (s *Server) databases(ctx context.Context) ([]dolt.DatabaseInfo, error) {
	s.dbMu.Lock()
	if s.dbCache != nil && time.Now().Before(s.dbExp) {
		result := s.dbCache
		s.dbMu.Unlock()
		return result, nil
	}
	s.dbMu.Unlock()

	dbs, err := s.ds.ListBeadsDatabases(ctx)
	if err != nil {
		return nil, err
	}

	s.dbMu.Lock()
	s.dbCache = dbs
	s.dbExp = time.Now().Add(dbCacheTTL)
	s.dbMu.Unlock()

	return dbs, nil
}

var funcMap = template.FuncMap{
	"formatDate": func(t time.Time) string {
		if t.IsZero() {
			return "—"
		}
		return t.Format("Jan 2, 2006")
	},
	"formatDateTime": func(t time.Time) string {
		if t.IsZero() {
			return "—"
		}
		return t.Format("Jan 2, 2006 15:04")
	},
	"statusClass": func(s string) string {
		switch s {
		case "open":
			return "status-open"
		case "closed", "completed":
			return "status-closed"
		case "in_progress", "hooked":
			return "status-progress"
		case "blocked":
			return "status-blocked"
		case "deferred":
			return "status-deferred"
		default:
			return "status-other"
		}
	},
	"priorityLabel": func(p int) string {
		if p >= 0 && p <= 4 {
			return fmt.Sprintf("P%d", p)
		}
		return "—"
	},
	"statusBadge": func(s string) template.HTML {
		cls := "status-other"
		switch s {
		case "open":
			cls = "status-open"
		case "closed", "completed":
			cls = "status-closed"
		case "in_progress", "hooked":
			cls = "status-progress"
		case "blocked":
			cls = "status-blocked"
		case "deferred":
			cls = "status-deferred"
		}
		return template.HTML(fmt.Sprintf(`<span class="badge %s">%s</span>`, cls, template.HTMLEscapeString(s)))
	},
	"timeAgo": func(t time.Time) string {
		if t.IsZero() {
			return "—"
		}
		d := time.Since(t)
		switch {
		case d < time.Minute:
			return "just now"
		case d < time.Hour:
			return fmt.Sprintf("%dm ago", int(d.Minutes()))
		case d < 24*time.Hour:
			return fmt.Sprintf("%dh ago", int(d.Hours()))
		default:
			return fmt.Sprintf("%dd ago", int(d.Hours()/24))
		}
	},
	"shortActor": func(s string) string {
		if s == "" {
			return "—"
		}
		// Handle email addresses: extract local part
		if idx := strings.Index(s, "@"); idx > 0 {
			return s[:idx]
		}
		// Handle path-style names: aegis/crew/ellie → ellie
		parts := strings.Split(s, "/")
		return parts[len(parts)-1]
	},
	"rigName": func(s string) string {
		return strings.TrimPrefix(s, "beads_")
	},
	"fmtMonth": func(m time.Month) string {
		return fmt.Sprintf("%02d", int(m))
	},
	"priorityClass": func(p int) string {
		return fmt.Sprintf("p%d", p)
	},
	"daysAgo": func(t time.Time) int {
		return int(time.Since(t).Hours() / 24)
	},
	"progressPct": func(p dolt.EpicProgress) int {
		if p.Total == 0 {
			return 0
		}
		return p.Closed * 100 / p.Total
	},
	"fmtDuration": func(d time.Duration) string {
		if d < time.Minute {
			return "< 1m"
		}
		if d < time.Hour {
			return fmt.Sprintf("%dm", int(d.Minutes()))
		}
		h := int(d.Hours())
		m := int(d.Minutes()) % 60
		if m == 0 {
			return fmt.Sprintf("%dh", h)
		}
		return fmt.Sprintf("%dh%dm", h, m)
	},
	"fmtHours": func(h float64) string {
		if h < 1 {
			return fmt.Sprintf("%dm", int(h*60))
		}
		days := int(h / 24)
		hours := int(h) % 24
		if days > 0 {
			if hours == 0 {
				return fmt.Sprintf("%dd", days)
			}
			return fmt.Sprintf("%dd%dh", days, hours)
		}
		return fmt.Sprintf("%dh", int(h))
	},
	"payloadString": func(e events.Event, key string) string {
		return events.PayloadString(e, key)
	},
	"itof": func(i int) float64 {
		return float64(i)
	},
	"mulf": func(a, b float64) float64 {
		return a * b
	},
	"divf": func(a, b float64) float64 {
		if b == 0 {
			return 0
		}
		return a / b
	},
	"sub": func(a, b int) int {
		return a - b
	},
	"barHeight": func(val, max int) int {
		if max == 0 {
			return 0
		}
		h := val * 100 / max
		if h == 0 && val > 0 {
			return 2
		}
		return h
	},
	"lower": func(s string) string {
		return strings.ToLower(s)
	},
	"add1": func(i int) int {
		return i + 1
	},
	"calIntensity": func(count, max int) string {
		if count == 0 || max == 0 {
			return ""
		}
		ratio := float64(count) / float64(max)
		switch {
		case ratio > 0.75:
			return "cal-hot"
		case ratio > 0.5:
			return "cal-warm"
		case ratio > 0.25:
			return "cal-mild"
		default:
			return "cal-cool"
		}
	},
	"calQueryParam": func(year, month int, rig string) string {
		q := fmt.Sprintf("?year=%d&month=%d", year, month)
		if rig != "" {
			q += "&rig=" + rig
		}
		return q
	},
	"fmtMonthInt": func(m time.Month) int {
		return int(m)
	},
}

// Option configures the server.
type Option func(*Server)

// WithWorkspace sets the Gas Town workspace path for reading events.
func WithWorkspace(path string) Option {
	return func(s *Server) { s.workspacePath = path }
}

// New creates a new Server. The DataSource may be nil, in which case pages
// will display a "no database" message instead of data.
func New(ds DataSource, opts ...Option) *Server {
	s := &Server{ds: ds, prom: newPromClient(), forgejo: newForgejoClient()}
	for _, o := range opts {
		o(s)
	}
	s.parseTemplates()

	staticSub, _ := fs.Sub(staticFS, "static")
	s.static = http.StripPrefix("/static/", http.FileServerFS(staticSub))

	return s
}

func (s *Server) parseTemplates() {
	s.templates = map[string]*template.Template{
		"monthly": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/monthly.html"),
		),
		"bead": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/bead.html"),
		),
		"beads": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/beads.html"),
		),
		"search": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/search.html"),
		),
		"status": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/status.html"),
		),
		"briefing": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/briefing.html"),
		),
		"agents": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/agents.html"),
		),
		"decisions": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/decisions.html"),
		),
		"achievements": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/achievements.html"),
		),
		"homelab": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/homelab.html"),
		),
		"designs": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/designs.html"),
		),
		"design": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/design.html"),
		),
		"theme-parks": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/theme-parks.html"),
		),
		"work": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/work.html"),
		),
		"commits": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/commits.html"),
		),
		"epics": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/epics.html"),
		),
		"agent": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/agent.html"),
		),
		"epic": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/epic.html"),
		),
		"command-center": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/command-center.html"),
		),
		"events": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/events.html"),
		),
		"handoffs": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/handoffs.html"),
		),
		"velocity": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/velocity.html"),
		),
		"executive": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/executive.html"),
		),
		"blocked": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/blocked.html"),
		),
		"labels": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/labels.html"),
		),
		"stale": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/stale.html"),
		),
		"closed": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/closed.html"),
		),
		"deps": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/deps.html"),
		),
		"priorities": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/priorities.html"),
		),
		"activity": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/activity.html"),
		),
		"owners": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/owners.html"),
		),
		"types": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/types.html"),
		),
		"matrix": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/matrix.html"),
		),
		"sla": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/sla.html"),
		),
		"kanban": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/kanban.html"),
		),
		"heatmap": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/heatmap.html"),
		),
		"backlog": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/backlog.html"),
		),
		"recap": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/recap.html"),
		),
		"forecast": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/forecast.html"),
		),
		"scope": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/scope.html"),
		),
		"rigs": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/rigs.html"),
		),
		"burndown": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/burndown.html"),
		),
		"trends": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/trends.html"),
		),
		"cycle-time": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/cycle-time.html"),
		),
		"queue": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/queue.html"),
		),
		"duplicates": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/duplicates.html"),
		),
		"watchlist": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/watchlist.html"),
		),
		"flow-rate": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/flow-rate.html"),
		),
		"comments": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/comments.html"),
		),
		"triage": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/triage.html"),
		),
		"inventory": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/inventory.html"),
		),
		"response-time": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/response-time.html"),
		),
		"contributors": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/contributors.html"),
		),
		"deferred": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/deferred.html"),
		),
		"throughput": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/throughput.html"),
		),
		"churn": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/churn.html"),
		),
		"parking-lot": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/parking-lot.html"),
		),
		"net-flow": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/net-flow.html"),
		),
		"resolution-rate": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/resolution-rate.html"),
		),
		"age-breakdown": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/age-breakdown.html"),
		),
		"cohort": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/cohort.html"),
		),
		"workload": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/workload.html"),
		),
		"probes": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/probes.html"),
		),
		"created": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/created.html"),
		),
		"sprint": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/sprint.html"),
		),
		"standup": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/standup.html"),
		),
		"momentum": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/momentum.html"),
		),
		"risks": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/risks.html"),
		),
		"funnel": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/funnel.html"),
		),
		"overflow": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/overflow.html"),
		),
		"calendar": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/calendar.html"),
		),
		"debt": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/debt.html"),
		),
		"snapshot": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/snapshot.html"),
		),
		"assignments": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/assignments.html"),
		),
		"gaps": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/gaps.html"),
		),
		"compare": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/compare.html"),
		),
		"chains": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/chains.html"),
		),
		"wip": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/wip.html"),
		),
		"swarming": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/swarming.html"),
		),
		"signals": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/signals.html"),
		),
		"pair-freq": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/pair-freq.html"),
		),
		"idle": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/idle.html"),
		),
		"reopen": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/reopen.html"),
		),
		"escalations": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/escalations.html"),
		),
		"focus": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/focus.html"),
		),
		"crossref": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/crossref.html"),
		),
		"freshness": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/freshness.html"),
		),
		"complexity": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/complexity.html"),
		),
		"label-matrix": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/label-matrix.html"),
		),
		"label-trends": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/label-trends.html"),
		),
		"ready": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/ready.html"),
		),
		"burnup": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/burnup.html"),
		),
		"disposition": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/disposition.html"),
		),
		"phase": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/phase.html"),
		),
		"tag-velocity": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/tag-velocity.html"),
		),
		"pacing": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/pacing.html"),
		),
		"agent-velocity": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/agent-velocity.html"),
		),
		"unblocked": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/unblocked.html"),
		),
		"audit-log": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/audit-log.html"),
		),
		"label-detail": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/label-detail.html"),
		),
		"reschedules": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/reschedules.html"),
		),
		"retention": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/retention.html"),
		),
		"dog-pile": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/dog-pile.html"),
		),
		"quick-wins": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/quick-wins.html"),
		),
		"orphans": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/orphans.html"),
		),
		"dwell": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/dwell.html"),
		),
		"transfers": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/transfers.html"),
		),
		"load-balance": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/load-balance.html"),
		),
		"stats": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/stats.html"),
		),
		"pending": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/pending.html"),
		),
		"label-age": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/label-age.html"),
		),
		"status-flow": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/status-flow.html"),
		),
		"priority-drift": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/priority-drift.html"),
		),
		"sitemap": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/sitemap.html"),
		),
		"pulse": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/pulse.html"),
		),
		"timeline": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/timeline.html"),
		),
		"changelog": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/changelog.html"),
		),
		"impact": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/impact.html"),
		),
		"streaks": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/streaks.html"),
		),
		"ratios": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/ratios.html"),
		),
		"outgoing": template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS,
				"templates/layout.html", "templates/outgoing.html"),
		),
	}
}

// ServeHTTP routes requests to the appropriate handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if strings.HasPrefix(path, "/static/") {
		s.static.ServeHTTP(w, r)
		return
	}

	// Allow POST for design comments and bead actions
	segments := strings.Split(strings.TrimPrefix(path, "/"), "/")
	if r.Method == http.MethodPost && len(segments) == 3 && segments[0] == "designs" {
		switch segments[2] {
		case "comment":
			s.handleDesignComment(w, r, segments[1])
		case "approve":
			s.handleDesignApprove(w, r, segments[1])
		default:
			http.NotFound(w, r)
		}
		return
	}

	// POST /bead/{db}/{id}/status — update bead status via HTMX
	if r.Method == http.MethodPost && len(segments) == 4 && segments[0] == "bead" && segments[3] == "status" {
		s.handleBeadStatusUpdate(w, r, segments[1], segments[2])
		return
	}

	// POST /bead/{db}/{id}/comment — add comment via HTMX
	if r.Method == http.MethodPost && len(segments) == 4 && segments[0] == "bead" && segments[3] == "comment" {
		s.handleBeadComment(w, r, segments[1], segments[2])
		return
	}

	// POST /bead/{db}/{id}/priority — set priority via HTMX
	if r.Method == http.MethodPost && len(segments) == 4 && segments[0] == "bead" && segments[3] == "priority" {
		s.handleBeadPriorityUpdate(w, r, segments[1], segments[2])
		return
	}

	// POST /bead/{db}/{id}/assign — set assignee via HTMX
	if r.Method == http.MethodPost && len(segments) == 4 && segments[0] == "bead" && segments[3] == "assign" {
		s.handleBeadAssigneeUpdate(w, r, segments[1], segments[2])
		return
	}

	// POST /bead/{db}/{id}/label — add label via HTMX
	if r.Method == http.MethodPost && len(segments) == 4 && segments[0] == "bead" && segments[3] == "label" {
		s.handleBeadLabelAdd(w, r, segments[1], segments[2])
		return
	}

	// POST /bead/{db}/{id}/label/remove — remove label via HTMX
	if r.Method == http.MethodPost && len(segments) == 5 && segments[0] == "bead" && segments[3] == "label" && segments[4] == "remove" {
		s.handleBeadLabelRemove(w, r, segments[1], segments[2])
		return
	}

	// POST /bead/{db}/{id}/description — update description via HTMX
	if r.Method == http.MethodPost && len(segments) == 4 && segments[0] == "bead" && segments[3] == "description" {
		s.handleBeadDescriptionUpdate(w, r, segments[1], segments[2])
		return
	}

	// POST /bead/{db}/{id}/title — update title via HTMX
	if r.Method == http.MethodPost && len(segments) == 4 && segments[0] == "bead" && segments[3] == "title" {
		s.handleBeadTitleUpdate(w, r, segments[1], segments[2])
		return
	}

	// POST /batch/status — bulk status update via HTMX
	if r.Method == http.MethodPost && len(segments) == 2 && segments[0] == "batch" && segments[1] == "status" {
		s.handleBatchStatus(w, r)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if path == "/healthz" {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "ok")
		return
	}

	if path == "/" {
		s.handleIndex(w, r)
		return
	}

	switch {
	case len(segments) == 1 && segments[0] == "beads":
		s.handleBeadList(w, r)
	case len(segments) == 1 && segments[0] == "search":
		s.handleSearch(w, r)
	case len(segments) == 1 && segments[0] == "status":
		s.handleStatus(w, r)
	case len(segments) == 1 && segments[0] == "briefing":
		s.handleBriefing(w, r)
	case len(segments) == 1 && segments[0] == "agents":
		s.handleAgents(w, r)
	case len(segments) == 1 && segments[0] == "decisions":
		s.handleDecisions(w, r)
	case len(segments) == 1 && segments[0] == "achievements":
		s.handleAchievements(w, r)
	case len(segments) == 1 && segments[0] == "theme-parks":
		s.handleThemeParks(w, r)
	case len(segments) == 1 && segments[0] == "work":
		s.handleWork(w, r)
	case len(segments) == 1 && segments[0] == "commits":
		s.handleCommits(w, r)
	case len(segments) == 1 && segments[0] == "epics":
		s.handleEpics(w, r)
	case len(segments) == 1 && segments[0] == "command-center":
		s.handleCommandCenter(w, r)
	case len(segments) == 1 && segments[0] == "events":
		s.handleEvents(w, r)
	case len(segments) == 1 && segments[0] == "handoffs":
		s.handleHandoffs(w, r)
	case len(segments) == 1 && segments[0] == "velocity":
		s.handleVelocity(w, r)
	case len(segments) == 1 && segments[0] == "executive":
		s.handleExecutive(w, r)
	case len(segments) == 1 && segments[0] == "blocked":
		s.handleBlocked(w, r)
	case len(segments) == 1 && segments[0] == "labels":
		s.handleLabels(w, r)
	case len(segments) == 1 && segments[0] == "stale":
		s.handleStale(w, r)
	case len(segments) == 1 && segments[0] == "closed":
		s.handleClosed(w, r)
	case len(segments) == 1 && segments[0] == "deps":
		s.handleDeps(w, r)
	case len(segments) == 1 && segments[0] == "priorities":
		s.handlePriorities(w, r)
	case len(segments) == 1 && segments[0] == "activity":
		s.handleActivity(w, r)
	case len(segments) == 1 && segments[0] == "owners":
		s.handleOwners(w, r)
	case len(segments) == 1 && segments[0] == "types":
		s.handleTypes(w, r)
	case len(segments) == 1 && segments[0] == "matrix":
		s.handleMatrix(w, r)
	case len(segments) == 1 && segments[0] == "sla":
		s.handleSLA(w, r)
	case len(segments) == 1 && segments[0] == "kanban":
		s.handleKanban(w, r)
	case len(segments) == 1 && segments[0] == "backlog":
		s.handleBacklog(w, r)
	case len(segments) == 1 && segments[0] == "recap":
		s.handleRecap(w, r)
	case len(segments) == 1 && segments[0] == "forecast":
		s.handleForecast(w, r)
	case len(segments) == 1 && segments[0] == "scope":
		s.handleScope(w, r)
	case len(segments) == 1 && segments[0] == "rigs":
		s.handleRigs(w, r)
	case len(segments) == 1 && segments[0] == "burndown":
		s.handleBurndown(w, r)
	case len(segments) == 1 && segments[0] == "trends":
		s.handleTrends(w, r)
	case len(segments) == 1 && segments[0] == "cycle-time":
		s.handleCycleTime(w, r)
	case len(segments) == 1 && segments[0] == "queue":
		s.handleQueue(w, r)
	case len(segments) == 1 && segments[0] == "duplicates":
		s.handleDuplicates(w, r)
	case len(segments) == 1 && segments[0] == "watchlist":
		s.handleWatchlist(w, r)
	case len(segments) == 1 && segments[0] == "flow-rate":
		s.handleFlowRate(w, r)
	case len(segments) == 1 && segments[0] == "comments":
		s.handleComments(w, r)
	case len(segments) == 1 && segments[0] == "triage":
		s.handleTriage(w, r)
	case len(segments) == 1 && segments[0] == "inventory":
		s.handleInventory(w, r)
	case len(segments) == 1 && segments[0] == "response-time":
		s.handleResponseTime(w, r)
	case len(segments) == 1 && segments[0] == "contributors":
		s.handleContributors(w, r)
	case len(segments) == 1 && segments[0] == "deferred":
		s.handleDeferred(w, r)
	case len(segments) == 1 && segments[0] == "throughput":
		s.handleThroughput(w, r)
	case len(segments) == 1 && segments[0] == "churn":
		s.handleChurn(w, r)
	case len(segments) == 1 && segments[0] == "parking-lot":
		s.handleParkingLot(w, r)
	case len(segments) == 1 && segments[0] == "net-flow":
		s.handleNetFlow(w, r)
	case len(segments) == 1 && segments[0] == "resolution-rate":
		s.handleResolutionRate(w, r)
	case len(segments) == 1 && segments[0] == "age-breakdown":
		s.handleAgeBreakdown(w, r)
	case len(segments) == 1 && segments[0] == "cohort":
		s.handleCohort(w, r)
	case len(segments) == 1 && segments[0] == "workload":
		s.handleWorkload(w, r)
	case len(segments) == 1 && segments[0] == "heatmap":
		s.handleHeatmap(w, r)
	case len(segments) == 1 && segments[0] == "homelab":
		s.handleHomelab(w, r)
	case len(segments) == 1 && segments[0] == "probes":
		s.handleProbes(w, r)
	case len(segments) == 1 && segments[0] == "created":
		s.handleCreated(w, r)
	case len(segments) == 1 && segments[0] == "sprint":
		s.handleSprint(w, r)
	case len(segments) == 1 && segments[0] == "standup":
		s.handleStandup(w, r)
	case len(segments) == 1 && segments[0] == "momentum":
		s.handleMomentum(w, r)
	case len(segments) == 1 && segments[0] == "risks":
		s.handleRisks(w, r)
	case len(segments) == 1 && segments[0] == "funnel":
		s.handleFunnel(w, r)
	case len(segments) == 1 && segments[0] == "overflow":
		s.handleOverflow(w, r)
	case len(segments) == 1 && segments[0] == "calendar":
		s.handleCalendar(w, r)
	case len(segments) == 1 && segments[0] == "debt":
		s.handleDebt(w, r)
	case len(segments) == 1 && segments[0] == "snapshot":
		s.handleSnapshot(w, r)
	case len(segments) == 1 && segments[0] == "assignments":
		s.handleAssignments(w, r)
	case len(segments) == 1 && segments[0] == "gaps":
		s.handleGaps(w, r)
	case len(segments) == 1 && segments[0] == "compare":
		s.handleCompare(w, r)
	case len(segments) == 1 && segments[0] == "chains":
		s.handleChains(w, r)
	case len(segments) == 1 && segments[0] == "wip":
		s.handleWIP(w, r)
	case len(segments) == 1 && segments[0] == "swarming":
		s.handleSwarming(w, r)
	case len(segments) == 1 && segments[0] == "signals":
		s.handleSignals(w, r)
	case len(segments) == 1 && segments[0] == "pair-freq":
		s.handlePairFreq(w, r)
	case len(segments) == 1 && segments[0] == "idle":
		s.handleIdle(w, r)
	case len(segments) == 1 && segments[0] == "reopen":
		s.handleReopen(w, r)
	case len(segments) == 1 && segments[0] == "escalations":
		s.handleEscalations(w, r)
	case len(segments) == 1 && segments[0] == "focus":
		s.handleFocus(w, r)
	case len(segments) == 1 && segments[0] == "crossref":
		s.handleCrossRef(w, r)
	case len(segments) == 1 && segments[0] == "freshness":
		s.handleFreshness(w, r)
	case len(segments) == 1 && segments[0] == "complexity":
		s.handleComplexity(w, r)
	case len(segments) == 1 && segments[0] == "label-matrix":
		s.handleLabelMatrix(w, r)
	case len(segments) == 1 && segments[0] == "label-trends":
		s.handleLabelTrends(w, r)
	case len(segments) == 1 && segments[0] == "ready":
		s.handleReady(w, r)
	case len(segments) == 1 && segments[0] == "burnup":
		s.handleBurnup(w, r)
	case len(segments) == 1 && segments[0] == "disposition":
		s.handleDisposition(w, r)
	case len(segments) == 1 && segments[0] == "phase":
		s.handlePhase(w, r)
	case len(segments) == 1 && segments[0] == "tag-velocity":
		s.handleTagVelocity(w, r)
	case len(segments) == 1 && segments[0] == "pacing":
		s.handlePacing(w, r)
	case len(segments) == 1 && segments[0] == "agent-velocity":
		s.handleAgentVelocity(w, r)
	case len(segments) == 1 && segments[0] == "unblocked":
		s.handleUnblocked(w, r)
	case len(segments) == 1 && segments[0] == "audit-log":
		s.handleAuditLog(w, r)
	case len(segments) == 2 && segments[0] == "label":
		s.handleLabelDetail(w, r, segments[1])
	case len(segments) == 1 && segments[0] == "reschedules":
		s.handleReschedules(w, r)
	case len(segments) == 1 && segments[0] == "retention":
		s.handleRetention(w, r)
	case len(segments) == 1 && segments[0] == "dog-pile":
		s.handleDogPile(w, r)
	case len(segments) == 1 && segments[0] == "quick-wins":
		s.handleQuickWins(w, r)
	case len(segments) == 1 && segments[0] == "orphans":
		s.handleOrphans(w, r)
	case len(segments) == 1 && segments[0] == "dwell":
		s.handleDwell(w, r)
	case len(segments) == 1 && segments[0] == "transfers":
		s.handleTransfers(w, r)
	case len(segments) == 1 && segments[0] == "load-balance":
		s.handleLoadBalance(w, r)
	case len(segments) == 1 && segments[0] == "stats":
		s.handleStats(w, r)
	case len(segments) == 1 && segments[0] == "pending":
		s.handlePending(w, r)
	case len(segments) == 1 && segments[0] == "label-age":
		s.handleLabelAge(w, r)
	case len(segments) == 1 && segments[0] == "status-flow":
		s.handleStatusFlow(w, r)
	case len(segments) == 1 && segments[0] == "priority-drift":
		s.handlePriorityDrift(w, r)
	case len(segments) == 1 && segments[0] == "sitemap":
		s.handleSitemap(w, r)
	case len(segments) == 1 && segments[0] == "pulse":
		s.handlePulse(w, r)
	case len(segments) == 1 && segments[0] == "timeline":
		s.handleTimeline(w, r)
	case len(segments) == 1 && segments[0] == "changelog":
		s.handleChangelog(w, r)
	case len(segments) == 1 && segments[0] == "impact":
		s.handleImpact(w, r)
	case len(segments) == 1 && segments[0] == "streaks":
		s.handleStreaks(w, r)
	case len(segments) == 1 && segments[0] == "ratios":
		s.handleRatios(w, r)
	case len(segments) == 1 && segments[0] == "outgoing":
		s.handleOutgoing(w, r)
	case len(segments) == 1 && segments[0] == "designs":
		s.handleDesignsList(w, r)
	case len(segments) == 2 && segments[0] == "designs":
		s.handleDesignView(w, r, segments[1])
	case len(segments) == 2 && segments[0] == "epic":
		s.handleEpicDetail(w, r, segments[1])
	case len(segments) >= 2 && segments[0] == "agent":
		s.handleAgentDetail(w, r, strings.Join(segments[1:], "/"))
	case len(segments) == 3 && segments[0] == "bead":
		s.handleBead(w, r, segments[1], segments[2])
	case len(segments) == 2 && segments[0] == "bead":
		s.handleBeadLookup(w, r, segments[1])
	case len(segments) == 2:
		s.handleMonthly(w, r, segments[0], segments[1])
	default:
		http.NotFound(w, r)
	}
}

func (s *Server) render(w http.ResponseWriter, r *http.Request, name string, data any) {
	t, ok := s.templates[name]
	if !ok {
		http.Error(w, "template not found", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	tmplName := "layout"
	if r.Header.Get("HX-Request") != "" {
		tmplName = "content"
	}

	if err := t.ExecuteTemplate(w, tmplName, data); err != nil {
		log.Printf("template error: %v", err)
	}
}
