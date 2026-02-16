// Package web provides the HTTP server and HTMX-powered frontend.
package web

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/scbrown/tapestry/internal/config"
	"github.com/scbrown/tapestry/internal/dolt"
	"github.com/scbrown/tapestry/internal/events"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
)

//go:embed templates/*.html
var templateFS embed.FS

//go:embed static/*
var staticFS embed.FS

// Server is the tapestry HTTP server.
type Server struct {
	cfg    config.Config
	client *dolt.Client
	pages  map[string]*template.Template
	mux    *http.ServeMux
}

// New creates a server and connects to Dolt.
func New(cfg config.Config) (*Server, error) {
	funcMap := template.FuncMap{
		"priorityLabel": priorityLabel,
		"statusBadge":   statusBadge,
		"progressPct":   progressPct,
		"payloadString": events.PayloadString,
		"timeAgo":       timeAgo,
		"shortActor":    shortActor,
		"fmtDuration":   fmtDuration,
		"rigName":       func(s string) string { return strings.TrimPrefix(s, "beads_") },
		"nl":            func(s string) string { return strings.ReplaceAll(s, `\n`, "\n") },
		"markdown":      renderMarkdown,
	}

	pages := make(map[string]*template.Template)
	for _, name := range []string{
		"monthly.html", "bead.html", "beads.html",
		"epic.html", "epics.html", "agents.html", "agent.html", "events.html",
		"handoffs.html", "commits.html", "search.html",
	} {
		t, err := template.New(name).Funcs(funcMap).ParseFS(templateFS,
			"templates/layout.html", "templates/"+name)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", name, err)
		}
		pages[name] = t
	}

	// Connect to Dolt
	doltCfg := dolt.Config{
		Host:     cfg.Dolt.Host,
		Port:     cfg.Dolt.Port,
		User:     cfg.Dolt.User,
		Password: cfg.Dolt.Password,
	}
	client, err := dolt.New(doltCfg)
	if err != nil {
		return nil, fmt.Errorf("connect dolt: %w", err)
	}
	if err := client.Ping(context.Background()); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("ping dolt: %w", err)
	}

	// Discover beads databases if none configured
	if len(cfg.Workspace) == 0 || allDBsEmpty(cfg.Workspace) {
		dbs, err := client.ListBeadsDatabases(context.Background())
		if err != nil {
			log.Printf("warning: cannot list databases: %v", err)
		} else {
			names := make([]string, len(dbs))
			for i, db := range dbs {
				names[i] = db.Name
			}
			cfg.Workspace = []config.WorkspaceConfig{
				{Name: "auto", Databases: names},
			}
			log.Printf("auto-discovered %d beads databases", len(names))
		}
	}

	s := &Server{
		cfg:    cfg,
		client: client,
		pages:  pages,
		mux:    http.NewServeMux(),
	}
	s.routes()
	log.Printf("dolt connected at %s:%d", cfg.Dolt.Host, cfg.Dolt.Port)
	return s, nil
}

func allDBsEmpty(ws []config.WorkspaceConfig) bool {
	for _, w := range ws {
		if len(w.Databases) > 0 {
			return false
		}
	}
	return true
}

// Close shuts down database connections.
func (s *Server) Close() {
	if err := s.client.Close(); err != nil {
		log.Printf("close dolt: %v", err)
	}
}

// ListenAndServe starts the HTTP server.
func (s *Server) ListenAndServe() error {
	addr := fmt.Sprintf("%s:%d", s.cfg.Server.Host, s.cfg.Server.Port)
	log.Printf("tapestry serving at http://%s", addr)
	return http.ListenAndServe(addr, s.mux)
}

// databases returns all configured database names.
func (s *Server) databases() []string {
	var dbs []string
	for _, ws := range s.cfg.Workspace {
		dbs = append(dbs, ws.Databases...)
	}
	return dbs
}

func (s *Server) routes() {
	staticSub, _ := fs.Sub(staticFS, "static")
	s.mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticSub))))
	s.mux.HandleFunc("GET /{$}", s.handleMonthly)
	s.mux.HandleFunc("GET /month/{year}/{month}", s.handleMonthly)
	s.mux.HandleFunc("GET /bead/{id}", s.handleBead)
	s.mux.HandleFunc("GET /beads", s.handleBeadList)
	s.mux.HandleFunc("GET /epic/{id}", s.handleEpic)
	s.mux.HandleFunc("GET /epics", s.handleEpicsList)
	s.mux.HandleFunc("GET /agents", s.handleAgents)
	s.mux.HandleFunc("GET /agent/{name...}", s.handleAgent)
	s.mux.HandleFunc("GET /events", s.handleEvents)
	s.mux.HandleFunc("GET /handoffs", s.handleHandoffs)
	s.mux.HandleFunc("GET /commits", s.handleCommits)
	s.mux.HandleFunc("GET /search", s.handleSearch)
	s.mux.HandleFunc("GET /digest/{year}/{month}", s.handleDigest)
}

func (s *Server) render(w http.ResponseWriter, page string, data any) {
	t, ok := s.pages[page]
	if !ok {
		http.Error(w, "unknown page", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := t.ExecuteTemplate(w, page, data); err != nil {
		log.Printf("render %s: %v", page, err)
	}
}

func priorityLabel(p int) string {
	switch p {
	case 0:
		return "P0"
	case 1:
		return "P1"
	case 2:
		return "P2"
	case 3:
		return "P3"
	default:
		return fmt.Sprintf("P%d", p)
	}
}

func statusBadge(s string) template.HTML {
	color := "gray"
	switch s {
	case "open":
		color = "#3b82f6" // blue
	case "in_progress", "hooked":
		color = "#f59e0b" // amber
	case "closed":
		color = "#22c55e" // green
	}
	return template.HTML(fmt.Sprintf(
		`<span class="badge" style="background:%s">%s</span>`, color, s))
}

func progressPct(p dolt.EpicProgress) int {
	if p.Total == 0 {
		return 0
	}
	return p.Closed * 100 / p.Total
}

func timeAgo(t time.Time) string {
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
}

func shortActor(name string) string {
	parts := strings.Split(name, "/")
	return parts[len(parts)-1]
}

func fmtDuration(d time.Duration) string {
	if d == 0 {
		return "-"
	}
	d = d.Round(time.Minute)
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	if h > 0 {
		return fmt.Sprintf("%dh %dm", h, m)
	}
	return fmt.Sprintf("%dm", m)
}

// mdRenderer is a goldmark instance configured for safe HTML output.
var mdRenderer = goldmark.New(
	goldmark.WithExtensions(extension.GFM),
	goldmark.WithRendererOptions(html.WithHardWraps()),
)

// renderMarkdown converts a markdown string to safe HTML.
// It first normalises escaped newlines (\n literals) from the database.
func renderMarkdown(s string) template.HTML {
	// Dolt stores literal \n — expand to real newlines before parsing.
	s = strings.ReplaceAll(s, `\n`, "\n")
	var buf bytes.Buffer
	if err := mdRenderer.Convert([]byte(s), &buf); err != nil {
		return template.HTML(template.HTMLEscapeString(s))
	}
	return template.HTML(buf.String()) //nolint:gosec // goldmark output is safe
}
