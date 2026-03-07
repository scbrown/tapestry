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
	AddLabel(ctx context.Context, database, issueID, label string) error
	ThemeParks(ctx context.Context, database string) ([]dolt.ThemePark, error)
	Rides(ctx context.Context, database, parkID string) ([]dolt.Ride, error)
	ParkVisits(ctx context.Context, database, parkID string) ([]dolt.ParkVisit, error)
	TripPlans(ctx context.Context, database string) ([]dolt.TripPlan, error)
	DistinctLabels(ctx context.Context, database string) ([]dolt.LabelCount, error)
	IssuesByLabel(ctx context.Context, database, label string) ([]dolt.Issue, error)
	AllDependenciesWithIssues(ctx context.Context, database string) ([]dolt.DepEdge, error)
	CountByPriorityStatus(ctx context.Context, database string) ([]dolt.PriorityStatusCount, error)
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
	case len(segments) == 1 && segments[0] == "homelab":
		s.handleHomelab(w, r)
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
