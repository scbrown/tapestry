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
	databases []dolt.DatabaseInfo
	counts    map[string]int
	created   int
	closed    int
	activity  map[string]int
	issues    []dolt.Issue
	issue     *dolt.Issue
	comments  []dolt.Comment
	deps      []dolt.Dependency
	err       error
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
	return nil, m.err
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
