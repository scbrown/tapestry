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
	labelCounts   []dolt.LabelCount
	depEdges        []dolt.DepEdge
	priorityCounts  []dolt.PriorityStatusCount
	assigneeCounts  []dolt.AssigneeStatusCount
	achievementDefs []dolt.AchievementDef
	achievementUnlocks []dolt.AchievementUnlock
	labels          []string
	err             error
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
	return m.issues, m.err
}

func (m *mockDataSource) LabelsForIssue(_ context.Context, _, _ string) ([]string, error) {
	return m.labels, m.err
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
	return m.achievementDefs, m.err
}

func (m *mockDataSource) AchievementUnlocks(_ context.Context, _ string) ([]dolt.AchievementUnlock, error) {
	return m.achievementUnlocks, m.err
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

func (m *mockDataSource) DistinctLabels(_ context.Context, _ string) ([]dolt.LabelCount, error) {
	return m.labelCounts, m.err
}

func (m *mockDataSource) IssuesByLabel(_ context.Context, _, _ string) ([]dolt.Issue, error) {
	return m.issues, m.err
}

func (m *mockDataSource) AllDependenciesWithIssues(_ context.Context, _ string) ([]dolt.DepEdge, error) {
	return m.depEdges, m.err
}

func (m *mockDataSource) CountByPriorityStatus(_ context.Context, _ string) ([]dolt.PriorityStatusCount, error) {
	return m.priorityCounts, m.err
}

func (m *mockDataSource) CountByAssigneeStatus(_ context.Context, _ string) ([]dolt.AssigneeStatusCount, error) {
	return m.assigneeCounts, m.err
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

func TestClosedPage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/closed", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /closed status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Recently Closed") {
		t.Error("expected 'Recently Closed' heading")
	}
}

func TestClosedPage_WithData(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "aegis-c1", Title: "Done task", Status: "closed", Priority: 2, Assignee: "aegis/crew/arnold", UpdatedAt: time.Now()},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/closed?days=7", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /closed status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "aegis-c1") {
		t.Error("expected closed issue ID in output")
	}
}

func TestStalePage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/stale", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /stale status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Stale Work") {
		t.Error("expected 'Stale Work' heading")
	}
}

func TestStalePage_WithData(t *testing.T) {
	staleTime := time.Now().Add(-10 * 24 * time.Hour) // 10 days ago
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "aegis-s1", Title: "Stuck task", Status: "in_progress", Priority: 1, Assignee: "aegis/crew/arnold", UpdatedAt: staleTime},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/stale?days=3", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /stale status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "aegis-s1") {
		t.Error("expected stale issue ID in output")
	}
}

func TestLabelsPage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/labels", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /labels status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Labels") {
		t.Error("expected 'Labels' heading")
	}
}

func TestLabelsPage_WithData(t *testing.T) {
	ds := &mockDataSource{
		databases:   []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		labelCounts: []dolt.LabelCount{{Label: "desire-path", Count: 5}, {Label: "bug", Count: 3}},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/labels", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /labels status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "desire-path") {
		t.Error("expected label 'desire-path' in output")
	}
	if !strings.Contains(body, "bug") {
		t.Error("expected label 'bug' in output")
	}
}

func TestLabelsPage_FilterByLabel(t *testing.T) {
	ds := &mockDataSource{
		databases:   []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		labelCounts: []dolt.LabelCount{{Label: "bug", Count: 2}},
		issues: []dolt.Issue{
			{ID: "aegis-l1", Title: "Fix bug", Status: "open", Priority: 1, UpdatedAt: time.Now()},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/labels?label=bug", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /labels?label=bug status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "aegis-l1") {
		t.Error("expected filtered issue ID in output")
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

func TestDepsPage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/deps", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /deps status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Dependency") {
		t.Error("expected 'Dependency' heading")
	}
}

func TestDepsPage_WithData(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		depEdges: []dolt.DepEdge{
			{
				From: dolt.Issue{ID: "aegis-a1", Title: "Feature A", Status: "open", Priority: 1, UpdatedAt: time.Now()},
				To:   dolt.Issue{ID: "aegis-a2", Title: "Prereq B", Status: "in_progress", Priority: 1, UpdatedAt: time.Now()},
				Type: "depends_on",
			},
			{
				From: dolt.Issue{ID: "aegis-c1", Title: "Sub-task", Status: "open", Priority: 2, UpdatedAt: time.Now()},
				To:   dolt.Issue{ID: "aegis-e1", Title: "Epic", Status: "open", Priority: 1, UpdatedAt: time.Now()},
				Type: "child_of",
			},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/deps", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /deps status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "aegis-a1") {
		t.Error("expected from issue ID")
	}
	if !strings.Contains(body, "aegis-a2") {
		t.Error("expected to issue ID")
	}
	if !strings.Contains(body, "depends_on") {
		t.Error("expected depends_on type group")
	}
	if !strings.Contains(body, "child_of") {
		t.Error("expected child_of type group")
	}
}

func TestDepsPage_FilterByType(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		depEdges: []dolt.DepEdge{
			{
				From: dolt.Issue{ID: "aegis-d1", Title: "Dep", Status: "open", Priority: 1, UpdatedAt: time.Now()},
				To:   dolt.Issue{ID: "aegis-d2", Title: "Prereq", Status: "closed", Priority: 1, UpdatedAt: time.Now()},
				Type: "depends_on",
			},
			{
				From: dolt.Issue{ID: "aegis-ch1", Title: "Child", Status: "open", Priority: 2, UpdatedAt: time.Now()},
				To:   dolt.Issue{ID: "aegis-ep1", Title: "Parent", Status: "open", Priority: 1, UpdatedAt: time.Now()},
				Type: "child_of",
			},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/deps?type=child_of", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /deps?type=child_of status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "aegis-ch1") {
		t.Error("expected child_of entry")
	}
	if strings.Contains(body, "aegis-d1") {
		t.Error("depends_on entry should be filtered out")
	}
}

func TestPrioritiesPage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/priorities", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /priorities status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Priority") {
		t.Error("expected 'Priority' heading")
	}
}

func TestPrioritiesPage_WithData(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		priorityCounts: []dolt.PriorityStatusCount{
			{Priority: 0, Status: "open", Count: 2},
			{Priority: 1, Status: "in_progress", Count: 5},
			{Priority: 1, Status: "closed", Count: 10},
			{Priority: 2, Status: "open", Count: 20},
			{Priority: 3, Status: "deferred", Count: 3},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/priorities", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /priorities status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "P0") {
		t.Error("expected P0 row")
	}
	if !strings.Contains(body, "P1") {
		t.Error("expected P1 row")
	}
	if !strings.Contains(body, "40") {
		t.Error("expected grand total of 40")
	}
}

func TestActivityPage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/activity", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /activity status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Activity") {
		t.Error("expected 'Activity' heading")
	}
}

func TestActivityPage_WithData(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "aegis-act1", Title: "Recent work", Status: "in_progress", Priority: 1, UpdatedAt: time.Now()},
			{ID: "aegis-act2", Title: "Just closed", Status: "closed", Priority: 2, UpdatedAt: time.Now().Add(-time.Hour)},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/activity?hours=4", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /activity status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "aegis-act1") {
		t.Error("expected recent issue ID")
	}
	if !strings.Contains(body, "aegis-act2") {
		t.Error("expected second issue ID")
	}
	if !strings.Contains(body, "4h") {
		t.Error("expected 4h filter option to be highlighted")
	}
}

func TestOwnersPage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/owners", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /owners status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Workload") {
		t.Error("expected 'Workload' heading")
	}
}

func TestOwnersPage_WithData(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "a1", Title: "Task 1", Status: "open", Owner: "aegis/crew/arnold", UpdatedAt: time.Now()},
			{ID: "a2", Title: "Task 2", Status: "closed", Owner: "aegis/crew/arnold", UpdatedAt: time.Now()},
			{ID: "a3", Title: "Task 3", Status: "in_progress", Owner: "aegis/crew/grant", UpdatedAt: time.Now()},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/owners", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /owners status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "arnold") {
		t.Error("expected arnold in owners list")
	}
	if !strings.Contains(body, "grant") {
		t.Error("expected grant in owners list")
	}
}

func TestTypesPage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/types", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /types status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Type") {
		t.Error("expected 'Type' heading")
	}
}

func TestTypesPage_WithData(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "t1", Title: "A task", Type: "task", Status: "open", UpdatedAt: time.Now()},
			{ID: "t2", Title: "A bug", Type: "bug", Status: "closed", UpdatedAt: time.Now()},
			{ID: "t3", Title: "An epic", Type: "epic", Status: "in_progress", UpdatedAt: time.Now()},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/types", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /types status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "task") {
		t.Error("expected task type row")
	}
	if !strings.Contains(body, "bug") {
		t.Error("expected bug type row")
	}
	if !strings.Contains(body, "epic") {
		t.Error("expected epic type row")
	}
}

func TestMatrixPage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/matrix", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /matrix status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Matrix") {
		t.Error("expected 'Matrix' heading")
	}
}

func TestMatrixPage_WithData(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		assigneeCounts: []dolt.AssigneeStatusCount{
			{Assignee: "aegis/crew/arnold", Status: "open", Count: 5},
			{Assignee: "aegis/crew/arnold", Status: "closed", Count: 12},
			{Assignee: "aegis/crew/ellie", Status: "in_progress", Count: 3},
			{Assignee: "aegis/crew/ellie", Status: "closed", Count: 8},
			{Assignee: "(unassigned)", Status: "open", Count: 15},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/matrix", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /matrix status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "arnold") {
		t.Error("expected arnold in matrix")
	}
	if !strings.Contains(body, "ellie") {
		t.Error("expected ellie in matrix")
	}
	if !strings.Contains(body, "43") {
		t.Error("expected grand total of 43")
	}
}

func TestSLAPage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/sla", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /sla status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "SLA") {
		t.Error("expected 'SLA' heading")
	}
}

func TestSLAPage_WithBreachedBeads(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "a1", Title: "P0 old", Priority: 0, Status: "open", CreatedAt: time.Now().Add(-48 * time.Hour), UpdatedAt: time.Now()},
			{ID: "a2", Title: "P2 recent", Priority: 2, Status: "in_progress", CreatedAt: time.Now().Add(-1 * time.Hour), UpdatedAt: time.Now()},
			{ID: "a3", Title: "Already closed", Priority: 1, Status: "closed", CreatedAt: time.Now().Add(-72 * time.Hour), UpdatedAt: time.Now()},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/sla", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /sla status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "P0 old") {
		t.Error("expected breached P0 bead")
	}
	if !strings.Contains(body, "Breached") {
		t.Error("expected 'Breached' section")
	}
	if strings.Contains(body, "Already closed") {
		t.Error("closed beads should be excluded from SLA tracking")
	}
}

func TestKanbanPage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/kanban", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /kanban status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Kanban") {
		t.Error("expected 'Kanban' heading")
	}
}

func TestKanbanPage_WithData(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "k1", Title: "Open task", Status: "open", Priority: 1, UpdatedAt: time.Now()},
			{ID: "k2", Title: "WIP task", Status: "in_progress", Priority: 2, Assignee: "aegis/crew/arnold", UpdatedAt: time.Now()},
			{ID: "k3", Title: "Blocked task", Status: "blocked", Priority: 0, UpdatedAt: time.Now()},
			{ID: "k4", Title: "Done task", Status: "closed", Priority: 3, UpdatedAt: time.Now()},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/kanban", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /kanban status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Open task") {
		t.Error("expected open task on kanban")
	}
	if !strings.Contains(body, "WIP task") {
		t.Error("expected in_progress task on kanban")
	}
	if strings.Contains(body, "Done task") {
		t.Error("closed tasks should not appear on kanban")
	}
}

func TestHeatmapPage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/heatmap", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /heatmap status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Heatmap") {
		t.Error("expected 'Heatmap' heading")
	}
}

func TestHeatmapPage_WithData(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		created:   5,
		closed:    3,
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/heatmap", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /heatmap status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Activity") {
		t.Error("expected 'Activity' in heatmap page")
	}
}

func TestAchievementsPage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/achievements", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /achievements status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Achievements") {
		t.Error("expected 'Achievements' heading")
	}
}

func TestAchievementsPage_WithData(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "aegis"}},
		achievementDefs: []dolt.AchievementDef{
			{ID: "first-commit", Name: "First Commit", Description: "Push your first commit", Icon: "git-commit", Category: "development"},
			{ID: "infra-hero", Name: "Infrastructure Hero", Description: "Deploy 10 containers", Icon: "server", Category: "infrastructure"},
		},
		achievementUnlocks: []dolt.AchievementUnlock{
			{ID: "first-commit", UnlockedAt: time.Now().Add(-24 * time.Hour), UnlockedBy: "aegis/crew/arnold", Note: "Shipped it!"},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/achievements", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /achievements status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "First Commit") {
		t.Error("expected 'First Commit' achievement (unlocked)")
	}
	// Locked achievements show "???" not the name
	if !strings.Contains(body, "???") {
		t.Error("expected locked achievement placeholder")
	}
}

func TestAchievementsPage_CategoryFilter(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "aegis"}},
		achievementDefs: []dolt.AchievementDef{
			{ID: "a1", Name: "Dev Achievement", Category: "development"},
			{ID: "a2", Name: "Ops Achievement", Category: "operations"},
		},
		achievementUnlocks: []dolt.AchievementUnlock{
			{ID: "a1", UnlockedAt: time.Now(), UnlockedBy: "test"},
			{ID: "a2", UnlockedAt: time.Now(), UnlockedBy: "test"},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/achievements?category=development", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /achievements?category status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Dev Achievement") {
		t.Error("expected 'Dev Achievement' in filtered results")
	}
	// Ops Achievement should be filtered out
	if strings.Contains(body, "Ops Achievement") {
		t.Error("ops achievement should be filtered out by category=development")
	}
}

func TestDecisionsPage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/decisions", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /decisions status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Decisions") {
		t.Error("expected 'Decisions' heading")
	}
}

func TestDecisionsPage_WithData(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "d1", Title: "Choose database backend", Type: "decision", Status: "open", Priority: 1, UpdatedAt: time.Now(),
				Description: "Options:\nA: PostgreSQL [RECOMMENDED]\nB: SQLite\nC: MySQL"},
		},
		labels: []string{"decision:pending", "decision:requester:aegis/crew/arnold"},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/decisions", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /decisions status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Choose database backend") {
		t.Error("expected decision title in page")
	}
}

func TestDecisionsPage_FilterByState(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "d1", Title: "Pending decision", Type: "decision", Status: "open", Priority: 2, UpdatedAt: time.Now()},
		},
		labels: []string{"decision:pending"},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/decisions?filter=pending", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /decisions?filter status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestEpicsPage_WithData(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "e1", Title: "Bucket epic", Type: "epic", Status: "open", Priority: 1, UpdatedAt: time.Now()},
			{ID: "e1.1", Title: "Child task 1", Status: "closed", Priority: 2, UpdatedAt: time.Now()},
			{ID: "e1.2", Title: "Child task 2", Status: "open", Priority: 2, UpdatedAt: time.Now()},
		},
		deps: []dolt.Dependency{
			{FromID: "e1.1", ToID: "e1", Type: "child_of"},
			{FromID: "e1.2", ToID: "e1", Type: "child_of"},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/epics", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /epics status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Bucket epic") {
		t.Error("expected epic title in page")
	}
}

func TestDesignsPage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/designs", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /designs status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Design") {
		t.Error("expected 'Design' in heading")
	}
}

func TestHomelabPage_NilPrometheus(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/homelab", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /homelab status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Homelab") {
		t.Error("expected 'Homelab' heading")
	}
}

func TestBacklogPage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/backlog", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /backlog status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Backlog") {
		t.Error("expected 'Backlog' heading")
	}
}

func TestBacklogPage_WithData(t *testing.T) {
	now := time.Now()
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "b1", Title: "Fresh task", Status: "open", Priority: 1, CreatedAt: now.Add(-2 * time.Hour), UpdatedAt: now},
			{ID: "b2", Title: "Week old task", Status: "in_progress", Priority: 2, CreatedAt: now.Add(-7 * 24 * time.Hour), UpdatedAt: now},
			{ID: "b3", Title: "Old task", Status: "open", Priority: 3, CreatedAt: now.Add(-45 * 24 * time.Hour), UpdatedAt: now},
			{ID: "b4", Title: "Closed task", Status: "closed", Priority: 1, CreatedAt: now.Add(-10 * 24 * time.Hour), UpdatedAt: now},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/backlog", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /backlog status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Fresh task") {
		t.Error("expected 'Fresh task' in backlog")
	}
	if !strings.Contains(body, "Old task") {
		t.Error("expected 'Old task' in backlog")
	}
	// Closed tasks should not appear
	if strings.Contains(body, "Closed task") {
		t.Error("closed tasks should not appear in backlog")
	}
}

func TestThemeParksPage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/theme-parks", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /theme-parks status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Theme Park") {
		t.Error("expected 'Theme Park' in heading")
	}
}

func TestHTMXPartialResponse(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/briefing", nil)
	req.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("HTMX GET /briefing status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	// HTMX partial should NOT include full HTML wrapper
	if strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("HTMX response should not contain DOCTYPE (should be partial)")
	}
	if !strings.Contains(body, "Briefing") {
		t.Error("HTMX response should contain page content")
	}
}

func TestFullPageResponse(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/briefing", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("full GET /briefing status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	// Full page should include layout wrapper
	if !strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("full page response should contain DOCTYPE")
	}
	if !strings.Contains(body, "Tapestry") {
		t.Error("full page should contain 'Tapestry' in layout")
	}
}

func TestNotFound(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/nonexistent-page", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("GET /nonexistent-page status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestBeadStatusUpdate_MethodNotAllowed(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/bead/aegis/test-123/status", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	// GET on a POST-only endpoint should 404 or method not allowed
	if w.Code == http.StatusOK {
		t.Error("GET on POST endpoint should not return 200")
	}
}

func TestSearchPage_HTMXPartial(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "s1", Title: "Search result", Status: "open", Priority: 2, UpdatedAt: time.Now()},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/search?q=search", nil)
	req.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("HTMX GET /search status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("HTMX search should return partial, not full page")
	}
	if !strings.Contains(body, "Search result") {
		t.Error("expected search result in response")
	}
}

func TestRecapPage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/recap", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /recap status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Daily Recap") {
		t.Error("expected 'Daily Recap' heading")
	}
}

func TestRecapPage_WithData(t *testing.T) {
	now := time.Now()
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "r1", Title: "Created today", Status: "open", Priority: 1, Owner: "aegis/crew/alice", CreatedAt: now.Add(-1 * time.Hour), UpdatedAt: now},
			{ID: "r2", Title: "Closed today", Status: "closed", Priority: 2, Assignee: "aegis/crew/bob", CreatedAt: now.Add(-24 * time.Hour), UpdatedAt: now},
			{ID: "r3", Title: "Active work", Status: "in_progress", Priority: 1, Assignee: "aegis/crew/alice", CreatedAt: now.Add(-48 * time.Hour), UpdatedAt: now},
			{ID: "r4", Title: "Old item", Status: "open", Priority: 3, CreatedAt: now.Add(-72 * time.Hour), UpdatedAt: now.Add(-72 * time.Hour)},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/recap", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /recap status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Created today") {
		t.Error("expected 'Created today' issue")
	}
	if !strings.Contains(body, "Closed today") {
		t.Error("expected 'Closed today' issue")
	}
	if !strings.Contains(body, "Active work") {
		t.Error("expected 'Active work' issue")
	}
}

func TestRecapPage_DateParam(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/recap?date=2026-03-01", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /recap?date status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "March 1, 2026") {
		t.Error("expected date label for March 1, 2026")
	}
	if !strings.Contains(body, "Next") {
		t.Error("expected Next link for non-today date")
	}
}

func TestForecastPage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/forecast", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /forecast status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Forecast") {
		t.Error("expected 'Forecast' heading")
	}
	if !strings.Contains(body, "No data source") {
		t.Error("expected 'No data source' label")
	}
}

func TestForecastPage_WithData(t *testing.T) {
	now := time.Now()
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "f1", Title: "Open task", Status: "open", Priority: 1, CreatedAt: now.Add(-2 * time.Hour), UpdatedAt: now},
			{ID: "f2", Title: "Active task", Status: "in_progress", Priority: 2, CreatedAt: now.Add(-24 * time.Hour), UpdatedAt: now},
			{ID: "f3", Title: "Blocked task", Status: "blocked", Priority: 1, CreatedAt: now.Add(-48 * time.Hour), UpdatedAt: now},
		},
		counts: map[string]int{"open": 5, "in_progress": 3, "blocked": 2, "closed": 20},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/forecast", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /forecast status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Backlog") {
		t.Error("expected 'Backlog' in forecast")
	}
	if !strings.Contains(body, "Weekly Trend") {
		t.Error("expected 'Weekly Trend' section")
	}
}

func TestRecapPage_HTMXPartial(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/recap", nil)
	req.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("HTMX GET /recap status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("HTMX recap should return partial, not full page")
	}
	if !strings.Contains(body, "Daily Recap") {
		t.Error("expected recap content in partial response")
	}
}

func TestScopePage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/scope", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /scope status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Scope Tracker") {
		t.Error("expected 'Scope Tracker' heading")
	}
}

func TestScopePage_WithData(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		counts:    map[string]int{"open": 5, "closed": 10},
		created:   3,
		closed:    2,
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/scope", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /scope status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Cumulative Flow") {
		t.Error("expected 'Cumulative Flow' section")
	}
	if !strings.Contains(body, "Daily Detail") {
		t.Error("expected 'Daily Detail' table")
	}
}

func TestRigsPage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/rigs", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /rigs status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Rigs") {
		t.Error("expected 'Rigs' heading")
	}
}

func TestRigsPage_WithData(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		counts:    map[string]int{"open": 10, "in_progress": 5, "blocked": 3, "closed": 50, "deferred": 2},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/rigs", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /rigs status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "beads_aegis") {
		t.Error("expected rig name 'beads_aegis' in output")
	}
	if !strings.Contains(body, "70") {
		t.Error("expected total count 70 in output")
	}
}

func TestRigsPage_HTMXPartial(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/rigs", nil)
	req.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("HTMX GET /rigs status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("HTMX rigs should return partial, not full page")
	}
	if !strings.Contains(body, "Rigs") {
		t.Error("expected rigs content in partial response")
	}
}

func TestBurndownPage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/burndown", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /burndown status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Burndown") {
		t.Error("expected 'Burndown' heading")
	}
}

func TestBurndownPage_WithData(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		counts:    map[string]int{"open": 5, "closed": 10},
		created:   3,
		closed:    2,
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/burndown", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /burndown status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Daily Detail") {
		t.Error("expected 'Daily Detail' table")
	}
}

func TestBurndownPage_HTMXPartial(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/burndown", nil)
	req.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("HTMX GET /burndown status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("HTMX burndown should return partial, not full page")
	}
}

func TestTrendsPage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/trends", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /trends status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Trends") {
		t.Error("expected 'Trends' heading")
	}
}

func TestTrendsPage_WithData(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		counts:    map[string]int{"open": 5, "closed": 10},
		created:   3,
		closed:    2,
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/trends", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /trends status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Weekly Detail") {
		t.Error("expected 'Weekly Detail' table")
	}
}

func TestTrendsPage_HTMXPartial(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/trends", nil)
	req.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("HTMX GET /trends status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("HTMX trends should return partial, not full page")
	}
}

func TestCycleTimePage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/cycle-time", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /cycle-time status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Cycle Time") {
		t.Error("expected 'Cycle Time' heading")
	}
}

func TestCycleTimePage_WithData(t *testing.T) {
	now := time.Now()
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "ct1", Title: "Quick fix", Status: "closed", Priority: 0, Type: "bug", CreatedAt: now.Add(-2 * time.Hour), UpdatedAt: now},
			{ID: "ct2", Title: "Medium task", Status: "closed", Priority: 1, Type: "task", CreatedAt: now.Add(-3 * 24 * time.Hour), UpdatedAt: now},
			{ID: "ct3", Title: "Long epic", Status: "closed", Priority: 2, Type: "epic", CreatedAt: now.Add(-14 * 24 * time.Hour), UpdatedAt: now},
			{ID: "ct4", Title: "Still open", Status: "open", Priority: 1, Type: "task", CreatedAt: now.Add(-5 * 24 * time.Hour), UpdatedAt: now},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/cycle-time", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /cycle-time status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Cycle Time Analytics") {
		t.Error("expected 'Cycle Time Analytics' heading")
	}
	if !strings.Contains(body, "Closed beads") {
		t.Error("expected 'Closed beads' stat")
	}
	if !strings.Contains(body, "Median cycle") {
		t.Error("expected 'Median cycle' stat")
	}
	if !strings.Contains(body, "By Priority") {
		t.Error("expected 'By Priority' section")
	}
	if !strings.Contains(body, "By Type") {
		t.Error("expected 'By Type' section")
	}
	if !strings.Contains(body, "Quick fix") {
		t.Error("expected fastest completion 'Quick fix'")
	}
	if !strings.Contains(body, "Long epic") {
		t.Error("expected slowest completion 'Long epic'")
	}
}

func TestCycleTimePage_HTMX(t *testing.T) {
	now := time.Now()
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "ct1", Title: "Done", Status: "closed", Priority: 1, CreatedAt: now.Add(-24 * time.Hour), UpdatedAt: now},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/cycle-time", nil)
	req.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("HTMX GET /cycle-time status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("HTMX cycle-time should return partial, not full page")
	}
}

func TestCycleTimePage_NoClosedBeads(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues:    []dolt.Issue{},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/cycle-time", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /cycle-time status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "No closed beads") {
		t.Error("expected 'No closed beads' message when no closed issues")
	}
}

func TestQueuePage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/queue", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /queue status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Work Queue") {
		t.Error("expected 'Work Queue' heading")
	}
}

func TestQueuePage_WithData(t *testing.T) {
	now := time.Now()
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "q1", Title: "Urgent P0", Status: "open", Priority: 0, Type: "bug", CreatedAt: now.Add(-48 * time.Hour), UpdatedAt: now},
			{ID: "q2", Title: "Normal task", Status: "open", Priority: 2, Type: "task", CreatedAt: now.Add(-24 * time.Hour), UpdatedAt: now},
			{ID: "q3", Title: "In progress", Status: "in_progress", Priority: 1, CreatedAt: now, UpdatedAt: now},
			{ID: "q4", Title: "Closed one", Status: "closed", Priority: 1, CreatedAt: now, UpdatedAt: now},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/queue", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /queue status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Urgent P0") {
		t.Error("expected open P0 'Urgent P0' in queue")
	}
	if !strings.Contains(body, "Normal task") {
		t.Error("expected open P2 'Normal task' in queue")
	}
	if strings.Contains(body, "In progress") {
		t.Error("in_progress items should not appear in queue")
	}
	if strings.Contains(body, "Closed one") {
		t.Error("closed items should not appear in queue")
	}
}

func TestQueuePage_Empty(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues:    []dolt.Issue{},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/queue", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /queue status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "No unblocked work") {
		t.Error("expected empty queue message")
	}
}

func TestQueuePage_HTMX(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "q1", Title: "Task", Status: "open", Priority: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/queue", nil)
	req.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("HTMX GET /queue status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("HTMX queue should return partial, not full page")
	}
}
