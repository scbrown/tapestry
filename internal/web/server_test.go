package web

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/scbrown/tapestry/internal/dolt"
)

type mockDataSource struct {
	databases     []dolt.DatabaseInfo
	counts        map[string]int
	created       int
	closed        int
	activity      map[string]int
	issues        []dolt.Issue
	issue         *dolt.Issue
	comments      []dolt.Comment
	deps          []dolt.Dependency
	metadata      *dolt.IssueMetadata
	statusHistory []dolt.StatusTransition
	children      []dolt.Issue
	blockedIssues []dolt.BlockedIssue
	err           error
}

func (m *mockDataSource) ListBeadsDatabases(_ context.Context) ([]dolt.DatabaseInfo, error) {
	return m.databases, m.err
}

func (m *mockDataSource) CountByStatus(_ context.Context, _ string) (map[string]int, error) {
	return m.counts, m.err
}

func (m *mockDataSource) CountCreatedInRange(_ context.Context, _ string, _, _ time.Time) (int, error) {
	return m.created, m.err
}

func (m *mockDataSource) CountClosedInRange(_ context.Context, _ string, _, _ time.Time) (int, error) {
	return m.closed, m.err
}

func (m *mockDataSource) AgentActivityInRange(_ context.Context, _ string, _, _ time.Time) (map[string]int, error) {
	return m.activity, m.err
}

func (m *mockDataSource) Issues(_ context.Context, _ string, _ dolt.IssueFilter) ([]dolt.Issue, error) {
	return m.issues, m.err
}

func (m *mockDataSource) IssueByID(_ context.Context, _, _ string) (*dolt.Issue, error) {
	return m.issue, m.err
}

func (m *mockDataSource) Comments(_ context.Context, _, _ string) ([]dolt.Comment, error) {
	return m.comments, m.err
}

func (m *mockDataSource) Dependencies(_ context.Context, _, _ string) ([]dolt.Dependency, error) {
	return m.deps, m.err
}

func (m *mockDataSource) SearchIssues(_ context.Context, _, _ string, _ int) ([]dolt.Issue, error) {
	return m.issues, m.err
}

func (m *mockDataSource) DistinctAssignees(_ context.Context, _ string) ([]string, error) {
	return nil, m.err
}

func (m *mockDataSource) BlockedIssues(_ context.Context, _ string) ([]dolt.BlockedIssue, error) {
	return m.blockedIssues, m.err
}

func (m *mockDataSource) AgentActivity(_ context.Context, _ string) ([]dolt.AgentStats, error) {
	return nil, m.err
}

func (m *mockDataSource) Decisions(_ context.Context, _ string) ([]dolt.Issue, error) {
	return nil, m.err
}

func (m *mockDataSource) LabelsForIssue(_ context.Context, _, _ string) ([]string, error) {
	return nil, m.err
}

func (m *mockDataSource) MetadataForIssue(_ context.Context, _, _ string) (*dolt.IssueMetadata, error) {
	if m.metadata != nil {
		return m.metadata, m.err
	}
	return &dolt.IssueMetadata{}, m.err
}

func (m *mockDataSource) StatusHistory(_ context.Context, _, _ string) ([]dolt.StatusTransition, error) {
	return m.statusHistory, m.err
}

func (m *mockDataSource) ChildIssues(_ context.Context, _, _ string) ([]dolt.Issue, error) {
	return m.children, m.err
}

func (m *mockDataSource) AchievementDefs(_ context.Context, _ string) ([]dolt.AchievementDef, error) {
	return nil, m.err
}

func (m *mockDataSource) AchievementUnlocks(_ context.Context, _ string) ([]dolt.AchievementUnlock, error) {
	return nil, m.err
}

func (m *mockDataSource) Epics(_ context.Context, _ string) ([]dolt.Issue, error) {
	return m.issues, m.err
}

func (m *mockDataSource) AllChildDependencies(_ context.Context, _ string) ([]dolt.Dependency, error) {
	return m.deps, m.err
}

func (m *mockDataSource) AddComment(_ context.Context, _, _, _, _ string) error {
	return m.err
}

func (m *mockDataSource) UpdateStatus(_ context.Context, _, _, _ string) error {
	return m.err
}

func (m *mockDataSource) AddLabel(_ context.Context, _, _, _ string) error {
	return m.err
}

func (m *mockDataSource) ThemeParks(_ context.Context, _ string) ([]dolt.ThemePark, error) {
	return nil, m.err
}

func (m *mockDataSource) Rides(_ context.Context, _, _ string) ([]dolt.Ride, error) {
	return nil, m.err
}

func (m *mockDataSource) ParkVisits(_ context.Context, _, _ string) ([]dolt.ParkVisit, error) {
	return nil, m.err
}

func (m *mockDataSource) TripPlans(_ context.Context, _ string) ([]dolt.TripPlan, error) {
	return nil, m.err
}

func TestIndexRendersMonthly(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET / status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(strings.ToLower(body), "tapestry") {
		t.Error("expected 'tapestry' in home page")
	}
}

func TestMonthlyPage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/2026/02", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "No database connection configured") {
		t.Error("expected error message for nil data source")
	}
	if !strings.Contains(body, "February 2026") {
		t.Error("expected month/year in output")
	}
}

func TestMonthlyPage_WithData(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		counts:    map[string]int{"open": 5, "closed": 3},
		created:   2,
		closed:    1,
		activity:  map[string]int{"goldblum": 4},
		issues: []dolt.Issue{
			{ID: "aegis-001", Title: "Test issue", Status: "open", UpdatedAt: time.Now()},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/2026/02", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}

	body := w.Body.String()
	checks := []string{
		"February 2026",
		"beads_aegis",
		"Browse beads",
		"goldblum",
	}
	for _, check := range checks {
		if !strings.Contains(body, check) {
			t.Errorf("body missing %q", check)
		}
	}
}

func TestMonthlyPage_HTMXPartial(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/2026/02", nil)
	req.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}

	body := w.Body.String()
	// HTMX partial should NOT include the full HTML shell
	if strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("HTMX partial should not include DOCTYPE")
	}
	if !strings.Contains(body, "February 2026") {
		t.Error("partial should include page content")
	}
}

func TestMonthlyPage_InvalidMonth(t *testing.T) {
	srv := New(nil)

	tests := []struct {
		path string
	}{
		{"/2026/00"},
		{"/2026/13"},
		{"/2026/ab"},
	}

	for _, tt := range tests {
		req := httptest.NewRequest("GET", tt.path, nil)
		w := httptest.NewRecorder()
		srv.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("GET %s status = %d, want %d", tt.path, w.Code, http.StatusBadRequest)
		}
	}
}

func TestBeadPage_Found(t *testing.T) {
	ds := &mockDataSource{
		issue: &dolt.Issue{
			ID:          "aegis-001",
			Title:       "Fix the widget",
			Status:      "open",
			Priority:    1,
			Type:        "bug",
			Description: "The widget is broken.",
			CreatedAt:   time.Date(2026, 2, 10, 0, 0, 0, 0, time.UTC),
			UpdatedAt:   time.Date(2026, 2, 14, 0, 0, 0, 0, time.UTC),
		},
		comments: []dolt.Comment{
			{ID: 1, IssueID: "aegis-001", Author: "nux", Body: "Working on it."},
		},
		deps: []dolt.Dependency{
			{FromID: "aegis-001", ToID: "aegis-002", Type: "blocks"},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/bead/beads_aegis/aegis-001", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}

	body := w.Body.String()
	checks := []string{
		"Fix the widget",
		"aegis-001",
		"The widget is broken.",
		"Working on it.",
		"nux",
		"aegis-002",
	}
	for _, check := range checks {
		if !strings.Contains(body, check) {
			t.Errorf("body missing %q", check)
		}
	}
}

func TestBeadPage_NotFound(t *testing.T) {
	ds := &mockDataSource{
		issue: nil, // not found
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/bead/beads_aegis/nonexistent", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestBeadPage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/bead/beads_aegis/aegis-001", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}

	body := w.Body.String()
	if !strings.Contains(body, "No database connection configured") {
		t.Error("expected error message for nil data source")
	}
}

func TestStaticAssets(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/static/style.css", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if !strings.Contains(w.Body.String(), ":root") {
		t.Error("expected CSS content")
	}
}

func TestMonthNavigation(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/2026/01", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	body := w.Body.String()
	// Should have link to previous month (December 2025)
	if !strings.Contains(body, "/2025/12") {
		t.Error("expected link to December 2025")
	}
	// Should have link to next month (February 2026)
	if !strings.Contains(body, "/2026/02") {
		t.Error("expected link to February 2026")
	}
}

func TestBeadsList_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/beads", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /beads status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "No database connection configured") {
		t.Error("expected error message for nil data source")
	}
	if !strings.Contains(body, "Beads") {
		t.Error("expected 'Beads' heading")
	}
}

func TestBeadsList_WithData(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "aegis-010", Title: "Test bead", Status: "open", Priority: 1, Rig: "beads_aegis", UpdatedAt: time.Now()},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/beads", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /beads status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "aegis-010") {
		t.Errorf("body missing bead ID")
	}
	if !strings.Contains(body, "Test bead") {
		t.Errorf("body missing bead title")
	}
	if !strings.Contains(body, `href="/bead/aegis-010"`) {
		t.Errorf("body missing bead link")
	}
}

func TestCommitsPage_NoForgejo(t *testing.T) {
	srv := &Server{forgejo: nil}
	srv.parseTemplates()
	req := httptest.NewRequest("GET", "/commits", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /commits status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Commits") {
		t.Error("expected 'Commits' heading")
	}
}

func TestExtractBeadIDs(t *testing.T) {
	tests := []struct {
		msg  string
		want []string
	}{
		{"fix(aegis): resolve aegis-abc123 issue", []string{"aegis-abc123"}},
		{"no bead refs here", nil},
		{"multiple aegis-aaa hq-bbb tp-cc", []string{"aegis-aaa", "hq-bbb", "tp-cc"}},
		{"dupes aegis-xx aegis-xx", []string{"aegis-xx"}},
	}
	for _, tt := range tests {
		got := extractBeadIDs(tt.msg)
		if len(got) != len(tt.want) {
			t.Errorf("extractBeadIDs(%q) = %v, want %v", tt.msg, got, tt.want)
			continue
		}
		for i := range got {
			if got[i] != tt.want[i] {
				t.Errorf("extractBeadIDs(%q)[%d] = %q, want %q", tt.msg, i, got[i], tt.want[i])
			}
		}
	}
}

func TestSearch_EmptyQuery(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/search", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /search status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Search") {
		t.Error("expected 'Search' heading")
	}
}

func TestWorkPage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/work", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /work status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Work") {
		t.Error("expected 'Work' heading")
	}
}

func TestWorkPage_WithData(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "aegis-epic1", Title: "Big Epic", Status: "open", Priority: 1, Type: "epic", UpdatedAt: time.Now()},
			{ID: "aegis-task1", Title: "Standalone task", Status: "in_progress", Priority: 2, Type: "task", UpdatedAt: time.Now()},
		},
		deps: []dolt.Dependency{
			{FromID: "aegis-child1", ToID: "aegis-epic1", Type: "child_of"},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/work", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /work status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Big Epic") {
		t.Errorf("body missing epic title")
	}
	if !strings.Contains(body, "Standalone task") {
		t.Errorf("body missing standalone task")
	}
}

func TestWorkPage_PriorityMode(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "aegis-t1", Title: "P1 task", Status: "open", Priority: 1, Type: "task", UpdatedAt: time.Now()},
			{ID: "aegis-t2", Title: "P2 task", Status: "open", Priority: 2, Type: "task", UpdatedAt: time.Now()},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/work?mode=priority", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /work?mode=priority status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "P1 task") {
		t.Errorf("body missing P1 task")
	}
	if !strings.Contains(body, "P2 task") {
		t.Errorf("body missing P2 task")
	}
}

func TestWorkPage_AgentMode(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "aegis-a1", Title: "Arnold task", Status: "in_progress", Priority: 1, Type: "task", Assignee: "aegis/crew/arnold", UpdatedAt: time.Now()},
			{ID: "aegis-a2", Title: "Grant task", Status: "open", Priority: 2, Type: "task", Assignee: "aegis/crew/grant", UpdatedAt: time.Now()},
			{ID: "aegis-a3", Title: "Unowned task", Status: "open", Priority: 2, Type: "task", UpdatedAt: time.Now()},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/work?mode=agent", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /work?mode=agent status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "arnold") {
		t.Errorf("body missing agent 'arnold'")
	}
	if !strings.Contains(body, "Arnold task") {
		t.Errorf("body missing Arnold task")
	}
	if !strings.Contains(body, "By Agent") {
		t.Errorf("body missing agent mode toggle")
	}
}

func TestCommandCenter_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/command-center", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /command-center status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Command Center") {
		t.Error("expected 'Command Center' heading")
	}
}

func TestCommandCenter_WithData(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		counts:    map[string]int{"open": 10, "in_progress": 3, "closed": 50},
		closed:    2,
		issues: []dolt.Issue{
			{ID: "aegis-crit1", Title: "Critical fix", Status: "in_progress", Priority: 1, Type: "task", Assignee: "aegis/crew/arnold", UpdatedAt: time.Now()},
			{ID: "aegis-epic1", Title: "Big Epic", Status: "open", Priority: 1, Type: "epic", UpdatedAt: time.Now()},
		},
		deps: []dolt.Dependency{
			{FromID: "aegis-child1", ToID: "aegis-epic1", Type: "child_of"},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/command-center", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /command-center status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	checks := []string{
		"Command Center",
		"Critical fix",
		"arnold",
	}
	for _, check := range checks {
		if !strings.Contains(body, check) {
			t.Errorf("body missing %q", check)
		}
	}
}

func TestStatusPage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/status", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /status status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Executive Status") {
		t.Error("expected 'Executive Status' heading")
	}
}

func TestBriefingPage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/briefing", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /briefing status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Briefing") {
		t.Error("expected 'Briefing' heading")
	}
}

func TestBriefingPage_WithData(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		counts:    map[string]int{"open": 5, "in_progress": 2, "closed": 10},
		created:   3,
		closed:    1,
		issues: []dolt.Issue{
			{ID: "aegis-b1", Title: "Human task", Status: "open", Priority: 1, Owner: "stiwi", UpdatedAt: time.Now()},
			{ID: "aegis-b2", Title: "Active work", Status: "in_progress", Priority: 1, Assignee: "aegis/crew/arnold", UpdatedAt: time.Now()},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/briefing", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /briefing status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Human task") {
		t.Errorf("body missing needs-attention item")
	}
	if !strings.Contains(body, "Active work") {
		t.Errorf("body missing in-flight item")
	}
}

func TestBriefingPage_StaleWork(t *testing.T) {
	staleDate := time.Now().AddDate(0, 0, -10) // 10 days old
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		counts:    map[string]int{"open": 5},
		issues: []dolt.Issue{
			{ID: "aegis-stale1", Title: "Forgotten task", Status: "open", Priority: 1, UpdatedAt: staleDate},
			{ID: "aegis-fresh1", Title: "Fresh task", Status: "open", Priority: 1, UpdatedAt: time.Now()},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/briefing", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /briefing status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Forgotten task") {
		t.Error("expected stale 'Forgotten task' in briefing")
	}
	if !strings.Contains(body, "idle") {
		t.Error("expected 'idle' indicator for stale work")
	}
}

func TestAgentDetailPage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/agent/arnold", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /agent/arnold status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "arnold") {
		t.Error("expected agent name in page")
	}
}

func TestAgentDetailPage_WithData(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "aegis-a1", Title: "Agent task", Status: "open", Priority: 1, Assignee: "aegis/crew/arnold", UpdatedAt: time.Now()},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/agent/aegis/crew/arnold", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /agent/aegis/crew/arnold status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Agent task") {
		t.Errorf("body missing agent's issue")
	}
}

func TestAgentsPage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/agents", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /agents status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Agents") {
		t.Error("expected 'Agents' heading")
	}
}

func TestEpicsPage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/epics", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /epics status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Epics") {
		t.Error("expected 'Epics' heading")
	}
}

func TestHealthz(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /healthz status = %d, want %d", w.Code, http.StatusOK)
	}
	if w.Body.String() != "ok" {
		t.Errorf("expected body 'ok', got %q", w.Body.String())
	}
}

func TestMethodNotAllowed(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("DELETE", "/status", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("DELETE /status status = %d, want %d", w.Code, http.StatusMethodNotAllowed)
	}
}

func TestBeadLookup_CrossDB(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issue: &dolt.Issue{
			ID:     "aegis-lookup1",
			Title:  "Found via lookup",
			Status: "open",
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/bead/aegis-lookup1", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /bead/aegis-lookup1 status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Found via lookup") {
		t.Error("expected bead title in lookup result")
	}
}

func TestEventsPage_NoWorkspace(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/events", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /events status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Events") {
		t.Error("expected 'Events' heading")
	}
}

func TestHandoffsPage_NoWorkspace(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/handoffs", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /handoffs status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Handoff") {
		t.Error("expected 'Handoff' heading")
	}
}

func TestVelocityPage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/velocity", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /velocity status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Velocity") {
		t.Error("expected 'Velocity' heading")
	}
}

func TestVelocityPage_WithData(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		created:   5,
		closed:    3,
		issues: []dolt.Issue{
			{ID: "aegis-v1", Title: "Closed work", Status: "closed", Assignee: "aegis/crew/arnold", Owner: "aegis/crew/arnold", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/velocity", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /velocity status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	checks := []string{"Velocity", "created/day", "closed/day", "Daily Throughput"}
	for _, check := range checks {
		if !strings.Contains(body, check) {
			t.Errorf("body missing %q", check)
		}
	}
}

func TestSearch_WithResults(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "aegis-020", Title: "Found bead", Status: "open", Rig: "beads_aegis", UpdatedAt: time.Now()},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/search?q=found", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /search?q=found status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "aegis-020") {
		t.Errorf("body missing search result ID")
	}
	if !strings.Contains(body, "Found bead") {
		t.Errorf("body missing search result title")
	}
}

func TestBeadPage_WithLineage(t *testing.T) {
	ds := &mockDataSource{
		issue: &dolt.Issue{
			ID:        "aegis-100",
			Title:     "Bead with lineage",
			Status:    "closed",
			Priority:  1,
			Type:      "task",
			CreatedAt: time.Date(2026, 2, 10, 0, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2026, 2, 14, 0, 0, 0, 0, time.UTC),
		},
		metadata: &dolt.IssueMetadata{
			Lineage: &dolt.Lineage{
				Origin:      "human",
				OriginHuman: "stiwi",
				ExecutedBy:  "aegis/crew/goldblum",
			},
		},
		statusHistory: []dolt.StatusTransition{
			{ToStatus: "open", CommitDate: time.Date(2026, 2, 10, 0, 0, 0, 0, time.UTC)},
			{FromStatus: "open", ToStatus: "in_progress", CommitDate: time.Date(2026, 2, 12, 0, 0, 0, 0, time.UTC)},
			{FromStatus: "in_progress", ToStatus: "closed", CommitDate: time.Date(2026, 2, 14, 0, 0, 0, 0, time.UTC)},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/bead/beads_aegis/aegis-100", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}

	body := w.Body.String()
	checks := []string{
		"Lineage",
		"human",
		"stiwi",
		"aegis/crew/goldblum",
		"Status Timeline",
		"open",
		"in_progress",
		"closed",
	}
	for _, check := range checks {
		if !strings.Contains(body, check) {
			t.Errorf("body missing %q", check)
		}
	}
}

func TestBeadPage_EpicWithChildren(t *testing.T) {
	ds := &mockDataSource{
		issue: &dolt.Issue{
			ID:        "aegis-epic-1",
			Title:     "Parent epic",
			Status:    "open",
			Priority:  1,
			Type:      "epic",
			CreatedAt: time.Date(2026, 2, 10, 0, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2026, 2, 14, 0, 0, 0, 0, time.UTC),
		},
		children: []dolt.Issue{
			{ID: "aegis-epic-1.1", Title: "Child task one", Status: "closed", Priority: 1},
			{ID: "aegis-epic-1.2", Title: "Child task two", Status: "open", Priority: 2},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/bead/beads_aegis/aegis-epic-1", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}

	body := w.Body.String()
	checks := []string{
		"Children (2)",
		"aegis-epic-1.1",
		"Child task one",
		"aegis-epic-1.2",
		"Child task two",
	}
	for _, check := range checks {
		if !strings.Contains(body, check) {
			t.Errorf("body missing %q", check)
		}
	}
}

func TestBeadStatusUpdate(t *testing.T) {
	ds := &mockDataSource{
		issue: &dolt.Issue{
			ID:     "aegis-200",
			Title:  "Test bead",
			Status: "open",
		},
	}

	srv := New(ds)

	// POST to update status
	req := httptest.NewRequest("POST", "/bead/beads_aegis/aegis-200/status", strings.NewReader("status=closed"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	body := w.Body.String()
	if !strings.Contains(body, "closed") {
		t.Errorf("response missing 'closed' status")
	}
	if !strings.Contains(body, "Reopen") {
		t.Errorf("response missing 'Reopen' button for closed bead")
	}
}

func TestBeadStatusUpdate_InvalidStatus(t *testing.T) {
	ds := &mockDataSource{}
	srv := New(ds)

	req := httptest.NewRequest("POST", "/bead/beads_aegis/aegis-200/status", strings.NewReader("status=invalid"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestParseDescriptionMetadata(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantKeys  []string
		wantClean string
	}{
		{
			name:      "no metadata",
			input:     "Just a plain description.",
			wantKeys:  nil,
			wantClean: "",
		},
		{
			name:      "metadata prefix",
			input:     "attached_molecule: aegis-wisp-123\ndispatched_by: aegis/crew/maldoon\n\nThe actual description here.",
			wantKeys:  []string{"attached_molecule", "dispatched_by"},
			wantClean: "The actual description here.",
		},
		{
			name:      "metadata only",
			input:     "attached_at: 2026-03-07T13:37:50Z",
			wantKeys:  []string{"attached_at"},
			wantClean: "",
		},
		{
			name:      "empty",
			input:     "",
			wantKeys:  nil,
			wantClean: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, clean := parseDescriptionMetadata(tt.input)
			if tt.wantKeys == nil && info != nil {
				t.Errorf("expected nil info, got %v", info)
			}
			for _, key := range tt.wantKeys {
				if _, ok := info[key]; !ok {
					t.Errorf("missing key %q in info", key)
				}
			}
			if tt.wantClean != "" && clean != tt.wantClean {
				t.Errorf("clean = %q, want %q", clean, tt.wantClean)
			}
		})
	}
}

func TestExecutivePage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/executive", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /executive status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Executive Status") {
		t.Error("expected 'Executive Status' heading")
	}
}

func TestExecutivePage_WithData(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		counts:    map[string]int{"open": 10, "in_progress": 3, "blocked": 2, "closed": 50},
		created:   5,
		closed:    8,
		issues: []dolt.Issue{
			{ID: "aegis-e1", Title: "P0 critical", Status: "open", Priority: 0, Owner: "stiwi", UpdatedAt: time.Now()},
			{ID: "aegis-e2", Title: "P1 work", Status: "in_progress", Priority: 1, Assignee: "aegis/crew/arnold", UpdatedAt: time.Now()},
			{ID: "aegis-e3", Title: "P2 task", Status: "open", Priority: 2, UpdatedAt: time.Now()},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/executive", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /executive status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Executive Status") {
		t.Error("expected 'Executive Status' heading")
	}
	if !strings.Contains(body, "7-Day Throughput") {
		t.Error("expected throughput chart section")
	}
	if !strings.Contains(body, "Priority Distribution") {
		t.Error("expected priority distribution section")
	}
	if !strings.Contains(body, "Agent Leaderboard") {
		t.Error("expected agent leaderboard section")
	}
}

func TestBeadComment_Post(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issue:     &dolt.Issue{ID: "aegis-c1", Title: "Test bead", Status: "open"},
	}

	srv := New(ds)
	body := strings.NewReader("body=Hello+comment")
	req := httptest.NewRequest("POST", "/bead/beads_aegis/aegis-c1/comment", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("POST comment status = %d, want %d", w.Code, http.StatusOK)
	}
	result := w.Body.String()
	if !strings.Contains(result, "Hello comment") {
		t.Error("response missing comment body")
	}
	if !strings.Contains(result, "tapestry-web") {
		t.Error("response missing author")
	}
}

func TestBeadComment_EmptyBody(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
	}

	srv := New(ds)
	body := strings.NewReader("body=")
	req := httptest.NewRequest("POST", "/bead/beads_aegis/aegis-c1/comment", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("POST empty comment status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestBlockedPage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/blocked", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /blocked status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Blocked") {
		t.Error("expected 'Blocked' heading")
	}
}

func TestBlockedPage_WithData(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		blockedIssues: []dolt.BlockedIssue{
			{
				Issue:   dolt.Issue{ID: "aegis-b1", Title: "Needs auth", Status: "open", Priority: 1, Assignee: "aegis/crew/arnold", UpdatedAt: time.Now()},
				Blocker: dolt.Issue{ID: "aegis-b2", Title: "Auth module", Status: "in_progress", Owner: "aegis/crew/grant"},
			},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/blocked", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /blocked status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "aegis-b1") {
		t.Error("expected blocked issue ID in output")
	}
	if !strings.Contains(body, "aegis-b2") {
		t.Error("expected blocker ID in output")
	}
	if !strings.Contains(body, "Top Blockers") {
		t.Error("expected 'Top Blockers' section")
	}
}
