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
	"time"

	"github.com/scbrown/tapestry/internal/dolt"
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
}

// Server serves the Tapestry web dashboard.
type Server struct {
	ds        DataSource
	templates map[string]*template.Template
	static    http.Handler
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
		case "in_progress":
			return "status-progress"
		default:
			return "status-other"
		}
	},
	"priorityLabel": func(p int) string {
		switch p {
		case 1:
			return "P1"
		case 2:
			return "P2"
		case 3:
			return "P3"
		default:
			return "—"
		}
	},
	"fmtMonth": func(m time.Month) string {
		return fmt.Sprintf("%02d", int(m))
	},
}

// New creates a new Server. The DataSource may be nil, in which case pages
// will display a "no database" message instead of data.
func New(ds DataSource) *Server {
	s := &Server{ds: ds}
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
	}
}

// ServeHTTP routes requests to the appropriate handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := r.URL.Path

	if strings.HasPrefix(path, "/static/") {
		s.static.ServeHTTP(w, r)
		return
	}

	if path == "/" {
		s.handleIndex(w, r)
		return
	}

	segments := strings.Split(strings.TrimPrefix(path, "/"), "/")

	switch {
	case len(segments) == 3 && segments[0] == "bead":
		s.handleBead(w, r, segments[1], segments[2])
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
