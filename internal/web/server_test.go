package web

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
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
	assignees       []string
	recentComments  []dolt.Comment
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
	return m.assignees, m.err
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

func (m *mockDataSource) UpdatePriority(_ context.Context, _, _ string, _ int) error {
	return m.err
}

func (m *mockDataSource) UpdateAssignee(_ context.Context, _, _, _ string) error {
	return m.err
}
func (m *mockDataSource) UpdateTitle(_ context.Context, _, _, _ string) error {
	return m.err
}
func (m *mockDataSource) UpdateDescription(_ context.Context, _, _, _ string) error {
	return m.err
}

func (m *mockDataSource) AddLabel(_ context.Context, _, _, _ string) error {
	return m.err
}

func (m *mockDataSource) RemoveLabel(_ context.Context, _, _, _ string) error {
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

func (m *mockDataSource) DependenciesWithIssues(_ context.Context, _, _ string) ([]dolt.DepEdge, error) {
	return m.depEdges, m.err
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

func (m *mockDataSource) RecentComments(_ context.Context, _ string, _ int) ([]dolt.Comment, error) {
	return m.recentComments, m.err
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
		depEdges: []dolt.DepEdge{
			{From: dolt.Issue{ID: "aegis-001", Title: "Fix the widget"}, To: dolt.Issue{ID: "aegis-002", Title: "Related issue"}, Type: "blocks"},
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
	if !strings.Contains(body, `href="/bead/beads_aegis/aegis-010"`) {
		t.Errorf("body missing bead link")
	}
}

func TestBeadsList_BatchActions(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "aegis-b1", Title: "Batch bead 1", Status: "open", Priority: 1, Rig: "beads_aegis", UpdatedAt: time.Now()},
			{ID: "aegis-b2", Title: "Batch bead 2", Status: "open", Priority: 2, Rig: "beads_aegis", UpdatedAt: time.Now()},
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
	if !strings.Contains(body, "batch-bar-beads") {
		t.Error("expected batch-bar-beads element for batch actions")
	}
	if !strings.Contains(body, `class="batch-cb"`) {
		t.Error("expected batch checkboxes on bead rows")
	}
	if !strings.Contains(body, "close selected") {
		t.Error("expected 'close selected' batch button")
	}
	if !strings.Contains(body, "defer selected") {
		t.Error("expected 'defer selected' batch button")
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

func TestBriefingPage_QuickActions(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		counts:    map[string]int{"open": 5, "in_progress": 2, "closed": 10},
		issues: []dolt.Issue{
			{ID: "aegis-att1", Title: "Human attention task", Status: "open", Priority: 0, Owner: "stiwi", UpdatedAt: time.Now().Add(-2 * time.Hour)},
			{ID: "aegis-fly1", Title: "Active flight work", Status: "in_progress", Priority: 1, Assignee: "aegis/crew/arnold", UpdatedAt: time.Now().Add(-1 * time.Hour)},
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
	// Needs Attention items should have close/defer buttons
	if !strings.Contains(body, "attention-item") {
		t.Error("expected attention-item class on needs-attention list items")
	}
	// In Flight items should have close button
	if !strings.Contains(body, "inflight-item") {
		t.Error("expected inflight-item class on in-flight list items")
	}
	// Both sections should have briefing-actions
	if count := strings.Count(body, "briefing-actions"); count < 2 {
		t.Errorf("expected at least 2 briefing-actions spans, got %d", count)
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

func TestBeadPriorityUpdate(t *testing.T) {
	ds := &mockDataSource{}
	srv := New(ds)

	req := httptest.NewRequest("POST", "/bead/beads_aegis/aegis-200/priority", strings.NewReader("priority=1"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", w.Code, http.StatusOK, w.Body.String())
	}
	body := w.Body.String()
	if !strings.Contains(body, "P1") {
		t.Errorf("response missing 'P1' badge, got: %s", body)
	}
	if !strings.Contains(body, "priority-badge") {
		t.Errorf("response missing priority-badge class, got: %s", body)
	}
}

func TestBeadPriorityUpdate_Invalid(t *testing.T) {
	ds := &mockDataSource{}
	srv := New(ds)

	req := httptest.NewRequest("POST", "/bead/beads_aegis/aegis-200/priority", strings.NewReader("priority=9"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestBeadAssigneeUpdate(t *testing.T) {
	ds := &mockDataSource{}
	srv := New(ds)

	req := httptest.NewRequest("POST", "/bead/beads_aegis/aegis-200/assign", strings.NewReader("assignee=aegis/crew/arnold"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", w.Code, http.StatusOK, w.Body.String())
	}
	body := w.Body.String()
	if !strings.Contains(body, "arnold") {
		t.Errorf("response missing short actor name 'arnold', got: %s", body)
	}
}

func TestBeadAssigneeUpdate_Empty(t *testing.T) {
	ds := &mockDataSource{}
	srv := New(ds)

	req := httptest.NewRequest("POST", "/bead/beads_aegis/aegis-200/assign", strings.NewReader("assignee="))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", w.Code, http.StatusOK, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "unassigned") {
		t.Errorf("response missing 'unassigned', got: %s", w.Body.String())
	}
}

func TestBeadLabelAdd(t *testing.T) {
	ds := &mockDataSource{}
	srv := New(ds)

	req := httptest.NewRequest("POST", "/bead/beads_aegis/aegis-200/label", strings.NewReader("label=improvement"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", w.Code, http.StatusOK, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "improvement") {
		t.Errorf("response missing label text, got: %s", w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "/labels?label=improvement") {
		t.Errorf("response missing label link, got: %s", w.Body.String())
	}
}

func TestBeadLabelAdd_Empty(t *testing.T) {
	ds := &mockDataSource{}
	srv := New(ds)

	req := httptest.NewRequest("POST", "/bead/beads_aegis/aegis-200/label", strings.NewReader("label="))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestBeadLabelAdd_InvalidChars(t *testing.T) {
	ds := &mockDataSource{}
	srv := New(ds)

	req := httptest.NewRequest("POST", "/bead/beads_aegis/aegis-200/label", strings.NewReader("label=bad+label"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestBeadLabelRemove(t *testing.T) {
	ds := &mockDataSource{}
	srv := New(ds)

	req := httptest.NewRequest("POST", "/bead/beads_aegis/aegis-200/label/remove", strings.NewReader("label=bug"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if w.Header().Get("HX-Trigger") == "" {
		t.Error("expected HX-Trigger header")
	}
}

func TestBeadLabelRemove_Empty(t *testing.T) {
	ds := &mockDataSource{}
	srv := New(ds)

	req := httptest.NewRequest("POST", "/bead/beads_aegis/aegis-200/label/remove", strings.NewReader("label="))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestBeadDescriptionUpdate(t *testing.T) {
	ds := &mockDataSource{}
	srv := New(ds)

	req := httptest.NewRequest("POST", "/bead/beads_aegis/aegis-200/description", strings.NewReader("description=Updated+description+text"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", w.Code, http.StatusOK, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "Updated description text") {
		t.Errorf("response missing updated description, got: %s", w.Body.String())
	}
	trigger := w.Header().Get("HX-Trigger")
	if !strings.Contains(trigger, "showToast") {
		t.Errorf("missing HX-Trigger header, got: %s", trigger)
	}
}

func TestBeadDescriptionUpdate_Empty(t *testing.T) {
	ds := &mockDataSource{}
	srv := New(ds)

	// Empty description is allowed (clearing description)
	req := httptest.NewRequest("POST", "/bead/beads_aegis/aegis-200/description", strings.NewReader("description="))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", w.Code, http.StatusOK, w.Body.String())
	}
}

func TestBeadTitleUpdate(t *testing.T) {
	ds := &mockDataSource{}
	srv := New(ds)

	req := httptest.NewRequest("POST", "/bead/beads_aegis/aegis-200/title", strings.NewReader("title=New+Title"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", w.Code, http.StatusOK, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "New Title") {
		t.Errorf("response missing updated title, got: %s", w.Body.String())
	}
	trigger := w.Header().Get("HX-Trigger")
	if !strings.Contains(trigger, "showToast") {
		t.Errorf("missing HX-Trigger header, got: %s", trigger)
	}
}

func TestBeadTitleUpdate_Empty(t *testing.T) {
	ds := &mockDataSource{}
	srv := New(ds)

	req := httptest.NewRequest("POST", "/bead/beads_aegis/aegis-200/title", strings.NewReader("title="))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestBatchStatus(t *testing.T) {
	ds := &mockDataSource{}
	srv := New(ds)

	body := "status=closed&ids[]=beads_aegis/aegis-001&ids[]=beads_aegis/aegis-002"
	req := httptest.NewRequest("POST", "/batch/status", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", w.Code, http.StatusOK, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "2 beads updated") {
		t.Errorf("response should mention 2 beads updated, got: %s", w.Body.String())
	}
	trigger := w.Header().Get("HX-Trigger")
	if !strings.Contains(trigger, "showToast") {
		t.Errorf("missing HX-Trigger header, got: %s", trigger)
	}
}

func TestBatchStatus_InvalidStatus(t *testing.T) {
	ds := &mockDataSource{}
	srv := New(ds)

	body := "status=invalid&ids[]=beads_aegis/aegis-001"
	req := httptest.NewRequest("POST", "/batch/status", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestBatchStatus_NoIds(t *testing.T) {
	ds := &mockDataSource{}
	srv := New(ds)

	body := "status=closed"
	req := httptest.NewRequest("POST", "/batch/status", strings.NewReader(body))
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

func TestClosedPage_BatchReopen(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "aegis-cbr1", Title: "Closed item", Status: "closed", Priority: 1, UpdatedAt: time.Now()},
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
	if !strings.Contains(body, "batch-bar-closed") {
		t.Error("expected batch bar on closed page")
	}
	if !strings.Contains(body, "closedToggleDay") {
		t.Error("expected per-day toggle-all on closed page")
	}
	if !strings.Contains(body, "closedBatchAction") {
		t.Error("expected batch reopen script on closed page")
	}
	if !strings.Contains(body, "reopen selected") {
		t.Error("expected 'reopen selected' button text")
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

func TestActivityPage_BatchActions(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "abatch1", Title: "Open work", Status: "open", Priority: 2, UpdatedAt: time.Now()},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/activity?hours=24", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /activity status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "batch-bar-activity") {
		t.Error("expected batch bar on activity page")
	}
	if !strings.Contains(body, "activityToggleAll") {
		t.Error("expected batch toggle-all on activity page")
	}
	if !strings.Contains(body, "activityBatchAction") {
		t.Error("expected batch action script on activity page")
	}
}

func TestActivityPage_RigFilter(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}, {Name: "beads_gastown"}},
		issues: []dolt.Issue{
			{ID: "arf1", Title: "Aegis work", Status: "open", Priority: 1, Rig: "beads_aegis", UpdatedAt: time.Now()},
		},
	}

	srv := New(ds)

	// Unfiltered — should show rig filter when multiple rigs have data
	req := httptest.NewRequest("GET", "/activity?hours=24", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /activity status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	// Activity page preserves rig filter in auto-refresh URL
	if !strings.Contains(body, `hx-get="/activity?hours=24"`) {
		t.Error("expected auto-refresh URL with hours param")
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

func TestQueuePage_AssigneeDropdown(t *testing.T) {
	now := time.Now()
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "qa1", Title: "Unassigned work", Status: "open", Priority: 1, CreatedAt: now.Add(-24 * time.Hour), UpdatedAt: now},
		},
		assignees: []string{"aegis/crew/alice", "aegis/crew/bob"},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/queue", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /queue status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "triage-assign") {
		t.Error("expected assignee dropdown on queue page")
	}
	if !strings.Contains(body, "alice") {
		t.Error("expected alice in assignee dropdown")
	}
	if !strings.Contains(body, "bob") {
		t.Error("expected bob in assignee dropdown")
	}
	if !strings.Contains(body, "/assign") {
		t.Error("expected /assign endpoint in dropdown")
	}
}

func TestQueuePage_UrgencyScore(t *testing.T) {
	now := time.Now()
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "qu1", Title: "Old P0", Status: "open", Priority: 0, CreatedAt: now.Add(-72 * time.Hour), UpdatedAt: now},
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
	if !strings.Contains(body, "urgency-score") {
		t.Error("expected urgency score display on queue page")
	}
	if !strings.Contains(body, "Urgency") {
		t.Error("expected 'Urgency' column header")
	}
}

func TestDuplicatesPage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/duplicates", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /duplicates status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Duplicate Detection") {
		t.Error("expected 'Duplicate Detection' heading")
	}
}

func TestDuplicatesPage_WithDuplicates(t *testing.T) {
	now := time.Now()
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "d1", Title: "[AUTO] ServiceDown: bobbin is down", Status: "open", Priority: 0, CreatedAt: now.Add(-2 * time.Hour), UpdatedAt: now},
			{ID: "d2", Title: "[AUTO] ServiceDown: bobbin is down on luvu", Status: "open", Priority: 0, CreatedAt: now.Add(-1 * time.Hour), UpdatedAt: now},
			{ID: "d3", Title: "[AUTO] ServiceDown: bobbin crashed", Status: "open", Priority: 0, CreatedAt: now, UpdatedAt: now},
			{ID: "d4", Title: "Unique bead", Status: "open", Priority: 2, CreatedAt: now, UpdatedAt: now},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/duplicates", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /duplicates status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Duplicate groups") {
		t.Error("expected 'Duplicate groups' stat")
	}
	if !strings.Contains(body, "servicedown") {
		t.Error("expected duplicate group key 'servicedown'")
	}
}

func TestDuplicatesPage_NoDuplicates(t *testing.T) {
	now := time.Now()
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "u1", Title: "Unique one", Status: "open", Priority: 1, CreatedAt: now, UpdatedAt: now},
			{ID: "u2", Title: "Another unique", Status: "open", Priority: 2, CreatedAt: now, UpdatedAt: now},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/duplicates", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /duplicates status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "No duplicate beads") {
		t.Error("expected 'No duplicate beads' message")
	}
}

func TestWatchlistPage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/watchlist", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /watchlist status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Watchlist") {
		t.Error("expected 'Watchlist' heading")
	}
}

func TestWatchlistPage_WithP0P1(t *testing.T) {
	now := time.Now()
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "w1", Title: "Server down!", Status: "open", Priority: 0, CreatedAt: now.Add(-1 * time.Hour), UpdatedAt: now},
			{ID: "w2", Title: "High prio bug", Status: "in_progress", Priority: 1, CreatedAt: now.Add(-24 * time.Hour), UpdatedAt: now.Add(-2 * time.Hour)},
			{ID: "w3", Title: "Normal task", Status: "open", Priority: 2, CreatedAt: now, UpdatedAt: now},
			{ID: "w4", Title: "Closed P0", Status: "closed", Priority: 0, CreatedAt: now, UpdatedAt: now},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/watchlist", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /watchlist status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Server down!") {
		t.Error("expected P0 'Server down!' in watchlist")
	}
	if !strings.Contains(body, "High prio bug") {
		t.Error("expected P1 'High prio bug' in watchlist")
	}
	if strings.Contains(body, "Normal task") {
		t.Error("P2 tasks should not appear in watchlist")
	}
	if strings.Contains(body, "Closed P0") {
		t.Error("closed P0 should not appear in watchlist")
	}
}

func TestWatchlistPage_AllClear(t *testing.T) {
	now := time.Now()
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "w1", Title: "Low prio", Status: "open", Priority: 3, CreatedAt: now, UpdatedAt: now},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/watchlist", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /watchlist status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "All clear") {
		t.Error("expected 'All clear' message when no P0/P1 beads")
	}
}

func TestFlowRatePage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/flow-rate", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /flow-rate status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Flow Rate") {
		t.Error("expected 'Flow Rate' heading")
	}
}

func TestFlowRatePage_WithData(t *testing.T) {
	now := time.Now()
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "f1", Title: "Created today", Status: "open", Priority: 1, CreatedAt: now, UpdatedAt: now},
			{ID: "f2", Title: "Created yesterday", Status: "open", Priority: 2, CreatedAt: now.Add(-24 * time.Hour), UpdatedAt: now},
			{ID: "f3", Title: "Closed today", Status: "closed", Priority: 1, CreatedAt: now.Add(-48 * time.Hour), UpdatedAt: now},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/flow-rate", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /flow-rate status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Created") {
		t.Error("expected 'Created' stat")
	}
	if !strings.Contains(body, "Closed") {
		t.Error("expected 'Closed' stat")
	}
	if !strings.Contains(body, "Net change") {
		t.Error("expected 'Net change' stat")
	}
}

func TestFlowRatePage_HTMX(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/flow-rate", nil)
	req.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("HTMX GET /flow-rate status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("HTMX flow-rate should return partial, not full page")
	}
}

func TestCommentsPage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/comments", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /comments status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Recent Comments") {
		t.Error("expected 'Recent Comments' heading")
	}
}

func TestCommentsPage_WithData(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/comments", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /comments status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Comments") {
		t.Error("expected comments content")
	}
}

func TestCommentsPage_HTMX(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/comments", nil)
	req.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("HTMX GET /comments status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("HTMX comments should return partial, not full page")
	}
}

func TestCommentsPage_AuthorFilter(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		recentComments: []dolt.Comment{
			{IssueID: "c1", Author: "aegis/crew/alice", Body: "Alice comment", CreatedAt: time.Now()},
			{IssueID: "c2", Author: "aegis/crew/bob", Body: "Bob comment", CreatedAt: time.Now().Add(-time.Hour)},
		},
	}

	srv := New(ds)

	// Test unfiltered — should show both + filter bar
	req := httptest.NewRequest("GET", "/comments", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /comments status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Alice comment") {
		t.Error("expected alice comment in unfiltered view")
	}
	if !strings.Contains(body, "Bob comment") {
		t.Error("expected bob comment in unfiltered view")
	}
	if !strings.Contains(body, "dep-filters") {
		t.Error("expected author filter bar")
	}

	// Test filtered by author
	req2 := httptest.NewRequest("GET", "/comments?author=aegis/crew/alice", nil)
	w2 := httptest.NewRecorder()
	srv.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Fatalf("GET /comments?author status = %d, want %d", w2.Code, http.StatusOK)
	}
	body2 := w2.Body.String()
	if !strings.Contains(body2, "Alice comment") {
		t.Error("expected alice comment in filtered view")
	}
	if strings.Contains(body2, "Bob comment") {
		t.Error("bob comment should be hidden when filtering by alice")
	}
}

func TestTriagePage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/triage", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /triage status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Triage Queue") {
		t.Error("expected 'Triage Queue' heading")
	}
}

func TestTriagePage_WithData(t *testing.T) {
	now := time.Now()
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "t1", Title: "Unassigned bug", Status: "open", Priority: 2, Type: "bug", CreatedAt: now.Add(-72 * time.Hour), UpdatedAt: now},
			{ID: "t2", Title: "Has owner", Status: "open", Priority: 1, Type: "task", Owner: "aegis/crew/arnold", CreatedAt: now, UpdatedAt: now},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/triage", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /triage status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Unassigned") {
		t.Error("expected 'Unassigned' section")
	}
	if !strings.Contains(body, "t1") {
		t.Error("expected unassigned bead t1")
	}
}

func TestTriagePage_HTMX(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/triage", nil)
	req.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("HTMX GET /triage status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("HTMX triage should return partial, not full page")
	}
}

func TestTriagePage_AllAssigned(t *testing.T) {
	now := time.Now()
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "a1", Title: "Assigned", Status: "open", Priority: 1, Owner: "someone", CreatedAt: now, UpdatedAt: now},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/triage", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /triage status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "All open beads are assigned and prioritized") {
		t.Error("expected empty state message")
	}
}

func TestInventoryPage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/inventory", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /inventory status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Bead Inventory") {
		t.Error("expected 'Bead Inventory' heading")
	}
}

func TestInventoryPage_WithData(t *testing.T) {
	now := time.Now()
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "i1", Title: "Open bug", Status: "open", Priority: 2, Type: "bug", CreatedAt: now, UpdatedAt: now},
			{ID: "i2", Title: "Closed task", Status: "closed", Priority: 1, Type: "task", CreatedAt: now, UpdatedAt: now},
			{ID: "i3", Title: "In progress", Status: "in_progress", Priority: 1, Type: "task", CreatedAt: now, UpdatedAt: now},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/inventory", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /inventory status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "By Status") {
		t.Error("expected 'By Status' section")
	}
	if !strings.Contains(body, "By Type") {
		t.Error("expected 'By Type' section")
	}
	if !strings.Contains(body, "By Rig") {
		t.Error("expected 'By Rig' section")
	}
}

func TestInventoryPage_HTMX(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/inventory", nil)
	req.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("HTMX GET /inventory status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("HTMX inventory should return partial, not full page")
	}
}

func TestResponseTimePage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/response-time", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /response-time status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Response Time") {
		t.Error("expected 'Response Time' heading")
	}
}

func TestResponseTimePage_WithData(t *testing.T) {
	now := time.Now()
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "rt1", Title: "Quick pickup", Status: "in_progress", Priority: 1, CreatedAt: now.Add(-2 * time.Hour), UpdatedAt: now.Add(-1 * time.Hour)},
			{ID: "rt2", Title: "Slow pickup", Status: "closed", Priority: 2, CreatedAt: now.Add(-48 * time.Hour), UpdatedAt: now.Add(-24 * time.Hour)},
			{ID: "rt3", Title: "Still open", Status: "open", Priority: 3, CreatedAt: now.Add(-72 * time.Hour), UpdatedAt: now},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/response-time", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /response-time status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Median Response") {
		t.Error("expected 'Median Response' stat")
	}
	if !strings.Contains(body, "Still Open") {
		t.Error("expected 'Still Open' stat")
	}
}

func TestResponseTimePage_HTMX(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/response-time", nil)
	req.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("HTMX GET /response-time status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("HTMX response-time should return partial, not full page")
	}
}

func TestContributorsPage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/contributors", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /contributors status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Contributors") {
		t.Error("expected 'Contributors' heading")
	}
}

func TestContributorsPage_WithData(t *testing.T) {
	now := time.Now()
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "c1", Title: "Bug", Status: "closed", Priority: 1, Owner: "aegis/crew/arnold", CreatedAt: now, UpdatedAt: now},
			{ID: "c2", Title: "Task", Status: "open", Priority: 2, Owner: "aegis/crew/arnold", CreatedAt: now, UpdatedAt: now},
			{ID: "c3", Title: "Epic", Status: "closed", Priority: 1, Owner: "aegis/crew/grant", CreatedAt: now, UpdatedAt: now},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/contributors", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /contributors status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "arnold") {
		t.Error("expected 'arnold' contributor")
	}
	if !strings.Contains(body, "Close Rate") {
		t.Error("expected 'Close Rate' column")
	}
}

func TestContributorsPage_HTMX(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/contributors", nil)
	req.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("HTMX GET /contributors status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("HTMX contributors should return partial, not full page")
	}
}

// --- Deferred page tests ---

func TestDeferredPage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/deferred", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /deferred status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Deferred Items") {
		t.Error("expected 'Deferred Items' heading")
	}
}

func TestDeferredPage_WithData(t *testing.T) {
	now := time.Now()
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "d1", Title: "Deferred bug", Status: "deferred", Priority: 2, CreatedAt: now.AddDate(0, 0, -30), UpdatedAt: now.AddDate(0, 0, -10)},
			{ID: "d2", Title: "Parked feature", Status: "deferred", Priority: 1, CreatedAt: now.AddDate(0, 0, -60), UpdatedAt: now.AddDate(0, 0, -45)},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/deferred", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /deferred status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Deferred bug") {
		t.Error("expected deferred issue title")
	}
	if !strings.Contains(body, "Priority Distribution") {
		t.Error("expected priority distribution section")
	}
}

func TestDeferredPage_HTMX(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/deferred", nil)
	req.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("HTMX GET /deferred status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("HTMX deferred should return partial, not full page")
	}
}

// --- Throughput page tests ---

func TestThroughputPage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/throughput", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /throughput status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Weekly Throughput") {
		t.Error("expected 'Weekly Throughput' heading")
	}
}

func TestThroughputPage_WithData(t *testing.T) {
	now := time.Now()
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "t1", Title: "Recent", Status: "open", CreatedAt: now.AddDate(0, 0, -3), UpdatedAt: now},
			{ID: "t2", Title: "Closed this week", Status: "closed", CreatedAt: now.AddDate(0, 0, -10), UpdatedAt: now.AddDate(0, 0, -1)},
			{ID: "t3", Title: "Old", Status: "open", CreatedAt: now.AddDate(0, 0, -50), UpdatedAt: now.AddDate(0, 0, -30)},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/throughput", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /throughput status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "12-Week View") {
		t.Error("expected '12-Week View' section")
	}
	if !strings.Contains(body, "Avg Created/wk") {
		t.Error("expected average stats")
	}
}

func TestThroughputPage_HTMX(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/throughput", nil)
	req.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("HTMX GET /throughput status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("HTMX throughput should return partial, not full page")
	}
}

// --- Churn page tests ---

func TestChurnPage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/churn", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /churn status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Status Churn") {
		t.Error("expected 'Status Churn' heading")
	}
}

func TestChurnPage_WithData(t *testing.T) {
	now := time.Now()
	ds := &mockDataSource{
		databases:     []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues:        []dolt.Issue{
			{ID: "c1", Title: "Bouncy bead", Status: "open", Priority: 1, CreatedAt: now.AddDate(0, 0, -20), UpdatedAt: now},
		},
		statusHistory: []dolt.StatusTransition{
			{FromStatus: "", ToStatus: "open", CommitDate: now.AddDate(0, 0, -20)},
			{FromStatus: "open", ToStatus: "in_progress", CommitDate: now.AddDate(0, 0, -15)},
			{FromStatus: "in_progress", ToStatus: "closed", CommitDate: now.AddDate(0, 0, -10)},
			{FromStatus: "closed", ToStatus: "open", CommitDate: now.AddDate(0, 0, -5)},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/churn", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /churn status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Bouncy bead") {
		t.Error("expected churning issue title")
	}
}

func TestChurnPage_HTMX(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/churn", nil)
	req.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("HTMX GET /churn status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("HTMX churn should return partial, not full page")
	}
}

// --- Parking Lot page tests ---

func TestParkingLotPage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/parking-lot", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /parking-lot status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Parking Lot") {
		t.Error("expected 'Parking Lot' heading")
	}
}

func TestParkingLotPage_WithData(t *testing.T) {
	now := time.Now()
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "p1", Title: "Stalled task", Status: "in_progress", Priority: 1, Assignee: "agent/alpha", CreatedAt: now.AddDate(0, 0, -20), UpdatedAt: now.AddDate(0, 0, -10)},
			{ID: "p2", Title: "Active task", Status: "in_progress", Priority: 2, CreatedAt: now, UpdatedAt: now},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/parking-lot", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /parking-lot status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Stalled task") {
		t.Error("expected stalled task (idle 10d)")
	}
	if strings.Contains(body, "Active task") {
		t.Error("should NOT show active task (idle 0d, below threshold)")
	}
}

func TestParkingLotPage_HTMX(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/parking-lot", nil)
	req.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("HTMX GET /parking-lot status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("HTMX parking-lot should return partial, not full page")
	}
}

// --- Cohort page tests ---

func TestCohortPage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/cohort", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /cohort status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Cohort Analysis") {
		t.Error("expected 'Cohort Analysis' heading")
	}
}

func TestCohortPage_WithData(t *testing.T) {
	now := time.Now()
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "co1", Title: "Recent open", Status: "open", CreatedAt: now.AddDate(0, 0, -3), UpdatedAt: now},
			{ID: "co2", Title: "Recent closed", Status: "closed", CreatedAt: now.AddDate(0, 0, -5), UpdatedAt: now.AddDate(0, 0, -1)},
			{ID: "co3", Title: "Old closed", Status: "closed", CreatedAt: now.AddDate(0, 0, -20), UpdatedAt: now.AddDate(0, 0, -15)},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/cohort", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /cohort status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Weekly Cohorts") {
		t.Error("expected 'Weekly Cohorts' section")
	}
	if !strings.Contains(body, "Overall Close Rate") {
		t.Error("expected overall close rate stat")
	}
}

func TestCohortPage_HTMX(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/cohort", nil)
	req.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("HTMX GET /cohort status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("HTMX cohort should return partial, not full page")
	}
}

// --- Workload page tests ---

func TestWorkloadPage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/workload", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /workload status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Workload Balance") {
		t.Error("expected 'Workload Balance' heading")
	}
}

func TestWorkloadPage_WithData(t *testing.T) {
	now := time.Now()
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "w1", Title: "Open bug", Status: "open", Priority: 0, Assignee: "agent/alpha", CreatedAt: now, UpdatedAt: now},
			{ID: "w2", Title: "In progress", Status: "in_progress", Priority: 1, Assignee: "agent/alpha", CreatedAt: now, UpdatedAt: now},
			{ID: "w3", Title: "Blocked", Status: "blocked", Priority: 2, Assignee: "agent/beta", CreatedAt: now, UpdatedAt: now},
			{ID: "w4", Title: "Closed", Status: "closed", Priority: 1, Assignee: "agent/alpha", CreatedAt: now, UpdatedAt: now},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/workload", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /workload status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "alpha") {
		t.Error("expected agent/alpha in workload table")
	}
	if !strings.Contains(body, "beta") {
		t.Error("expected agent/beta in workload table")
	}
}

func TestWorkloadPage_HTMX(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/workload", nil)
	req.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("HTMX GET /workload status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("HTMX workload should return partial, not full page")
	}
}

// --- Age Breakdown page tests ---

func TestAgeBreakdownPage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/age-breakdown", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /age-breakdown status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Open Bead Age Breakdown") {
		t.Error("expected 'Open Bead Age Breakdown' heading")
	}
}

func TestAgeBreakdownPage_WithData(t *testing.T) {
	now := time.Now()
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "ab1", Title: "Fresh bug", Status: "open", Priority: 1, CreatedAt: now.AddDate(0, 0, -2), UpdatedAt: now},
			{ID: "ab2", Title: "Month-old task", Status: "open", Priority: 2, CreatedAt: now.AddDate(0, 0, -45), UpdatedAt: now.AddDate(0, 0, -10)},
			{ID: "ab3", Title: "Ancient issue", Status: "open", Priority: 3, CreatedAt: now.AddDate(0, -7, 0), UpdatedAt: now.AddDate(0, -6, 0)},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/age-breakdown", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /age-breakdown status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Distribution") {
		t.Error("expected 'Distribution' section")
	}
	if !strings.Contains(body, "Oldest Open Bead") {
		t.Error("expected oldest bead section")
	}
	if !strings.Contains(body, "Ancient issue") {
		t.Error("expected oldest issue title")
	}
}

func TestAgeBreakdownPage_HTMX(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/age-breakdown", nil)
	req.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("HTMX GET /age-breakdown status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("HTMX age-breakdown should return partial, not full page")
	}
}

// --- Resolution Rate page tests ---

func TestResolutionRatePage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/resolution-rate", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /resolution-rate status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Resolution Rate") {
		t.Error("expected 'Resolution Rate' heading")
	}
}

func TestResolutionRatePage_WithData(t *testing.T) {
	now := time.Now()
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "rr1", Title: "Quick fix", Status: "closed", CreatedAt: now.AddDate(0, 0, -2), UpdatedAt: now.AddDate(0, 0, -1)},
			{ID: "rr2", Title: "Slow fix", Status: "closed", CreatedAt: now.AddDate(0, -4, 0), UpdatedAt: now.AddDate(0, 0, -5)},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/resolution-rate", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /resolution-rate status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Resolution Speed Distribution") {
		t.Error("expected resolution speed distribution section")
	}
	if !strings.Contains(body, "Fastest Resolution") {
		t.Error("expected fastest resolution section")
	}
}

func TestResolutionRatePage_HTMX(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/resolution-rate", nil)
	req.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("HTMX GET /resolution-rate status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("HTMX resolution-rate should return partial, not full page")
	}
}

// --- Net Flow page tests ---

func TestNetFlowPage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/net-flow", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /net-flow status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Net Flow") {
		t.Error("expected 'Net Flow' heading")
	}
}

func TestNetFlowPage_WithData(t *testing.T) {
	now := time.Now()
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "nf1", Title: "Recent", Status: "open", CreatedAt: now.AddDate(0, 0, -3), UpdatedAt: now},
			{ID: "nf2", Title: "Old closed", Status: "closed", CreatedAt: now.AddDate(0, 0, -20), UpdatedAt: now.AddDate(0, 0, -10)},
			{ID: "nf3", Title: "Very old", Status: "open", CreatedAt: now.AddDate(0, -3, 0), UpdatedAt: now.AddDate(0, -2, 0)},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/net-flow", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /net-flow status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Daily Detail") {
		t.Error("expected 'Daily Detail' section")
	}
	if !strings.Contains(body, "Current Open") {
		t.Error("expected current open stat")
	}
}

func TestNetFlowPage_HTMX(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/net-flow", nil)
	req.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("HTMX GET /net-flow status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("HTMX net-flow should return partial, not full page")
	}
}

func TestProbesPage_NoWorkspace(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/probes", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /probes status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Probe") {
		t.Error("expected 'Probe' in heading")
	}
}

func TestProbesPage_WithWorkspace(t *testing.T) {
	// Create a temp workspace with a docs/probes directory
	dir := t.TempDir()
	probesDir := dir + "/docs/probes"
	if err := os.MkdirAll(probesDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// Write a sample probe file
	if err := os.WriteFile(probesDir+"/2026-03-16-test-probe.md", []byte("# Test Probe\n\nSome findings here."), 0o644); err != nil {
		t.Fatal(err)
	}

	srv := New(nil, WithWorkspace(dir))
	req := httptest.NewRequest("GET", "/probes", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /probes status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Probe") {
		t.Error("expected 'Probe' heading")
	}
	if !strings.Contains(body, "Test Probe") {
		t.Error("expected probe entry 'Test Probe' in output")
	}
}

func TestProbesPage_CategoryFilter(t *testing.T) {
	dir := t.TempDir()
	probesDir := dir + "/docs/probes"
	subDir := probesDir + "/alerts"
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(probesDir+"/2026-03-16-general.md", []byte("# General\n\nGeneral probe."), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(subDir+"/2026-03-16-alert.md", []byte("# Alert Probe\n\nAlert finding."), 0o644); err != nil {
		t.Fatal(err)
	}

	srv := New(nil, WithWorkspace(dir))

	// Filter to alerts category only
	req := httptest.NewRequest("GET", "/probes?category=alerts", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /probes?category=alerts status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Alert Probe") {
		t.Error("expected 'Alert Probe' in filtered output")
	}
}

func TestProbesPage_HTMX(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/probes", nil)
	req.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("HTMX GET /probes status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("HTMX probes should return partial, not full page")
	}
}

func TestDesignViewPage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/designs/test-doc", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	// Should render without panic regardless of forgejo availability
	if w.Code != http.StatusOK && w.Code != http.StatusNotFound {
		t.Fatalf("GET /designs/test-doc status = %d, want 200 or 404", w.Code)
	}
}

func TestDesignViewPage_InvalidName(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/designs/bad%20name!", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("GET /designs/bad name! status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestDesignComment_MissingFields(t *testing.T) {
	srv := New(nil)
	body := strings.NewReader("bead_id=&bead_db=&author=&body=")
	req := httptest.NewRequest("POST", "/designs/test-doc/comment", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	// Should redirect with feedback=missing
	if w.Code != http.StatusSeeOther {
		t.Fatalf("POST design comment (empty) status = %d, want %d", w.Code, http.StatusSeeOther)
	}
	loc := w.Header().Get("Location")
	if !strings.Contains(loc, "feedback=missing") {
		t.Errorf("expected redirect to feedback=missing, got: %s", loc)
	}
}

func TestDesignComment_InvalidName(t *testing.T) {
	srv := New(nil)
	body := strings.NewReader("bead_id=x&bead_db=y&author=a&body=b")
	req := httptest.NewRequest("POST", "/designs/bad%20name!/comment", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("POST design comment (bad name) status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestDesignApprove_MissingFields(t *testing.T) {
	srv := New(nil)
	body := strings.NewReader("bead_id=&bead_db=")
	req := httptest.NewRequest("POST", "/designs/test-doc/approve", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	// Should redirect with feedback=error (missing bead_id/bead_db)
	if w.Code != http.StatusSeeOther {
		t.Fatalf("POST design approve (empty) status = %d, want %d", w.Code, http.StatusSeeOther)
	}
	loc := w.Header().Get("Location")
	if !strings.Contains(loc, "feedback=error") {
		t.Errorf("expected redirect to feedback=error, got: %s", loc)
	}
}

func TestDesignApprove_InvalidName(t *testing.T) {
	srv := New(nil)
	body := strings.NewReader("bead_id=x&bead_db=y")
	req := httptest.NewRequest("POST", "/designs/bad%20name!/approve", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("POST design approve (bad name) status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestRecapPage_AutoRefresh(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/recap", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /recap status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, `hx-trigger="every 60s"`) {
		t.Error("expected auto-refresh hx-trigger on recap page")
	}
	if !strings.Contains(body, `hx-get="/recap?date=`) {
		t.Error("expected hx-get with date param for auto-refresh")
	}
}

func TestRecapPage_ActiveActions(t *testing.T) {
	now := time.Now()
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "act1", Title: "Active task", Status: "in_progress", Priority: 1, Assignee: "aegis/crew/alice", CreatedAt: now.Add(-48 * time.Hour), UpdatedAt: now},
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
	if !strings.Contains(body, "Active task") {
		t.Error("expected active task in recap")
	}
	// Active items should now have close/defer buttons
	if !strings.Contains(body, `hx-vals='{"status":"closed"}'`) {
		t.Error("expected close button on active items")
	}
	if !strings.Contains(body, `hx-vals='{"status":"deferred"}'`) {
		t.Error("expected defer button on active items")
	}
}

func TestRecapPage_BatchActions(t *testing.T) {
	now := time.Now()
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "cr1", Title: "New bead", Status: "open", Priority: 2, Owner: "aegis/crew/bob", CreatedAt: now.Add(-1 * time.Hour), UpdatedAt: now},
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
	if !strings.Contains(body, "batch-bar-created") {
		t.Error("expected batch bar for created section")
	}
	if !strings.Contains(body, "recapToggleAll") {
		t.Error("expected batch toggle-all checkbox in created section")
	}
	if !strings.Contains(body, "recapBatchAction") {
		t.Error("expected batch action script")
	}
}

func TestSearchPage_BatchActions(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "s1", Title: "Search result", Status: "open", Rig: "beads_aegis", UpdatedAt: time.Now()},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/search?q=search", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /search status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "batch-bar-search") {
		t.Error("expected batch bar on search page")
	}
	if !strings.Contains(body, "searchToggleAll") {
		t.Error("expected batch toggle-all on search page")
	}
	if !strings.Contains(body, "searchBatchAction") {
		t.Error("expected batch action script on search page")
	}
}

func TestSearchPage_DeferButton(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "sd1", Title: "Open bead", Status: "open", Rig: "beads_aegis", UpdatedAt: time.Now()},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/search?q=open", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /search status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	// Open items should have defer button
	if !strings.Contains(body, `"status":"deferred"`) {
		t.Error("expected defer button for open items in search")
	}
}

func TestSearchPage_InProgressActions(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "sip1", Title: "In progress bead", Status: "in_progress", Rig: "beads_aegis", UpdatedAt: time.Now()},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/search?q=progress", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /search status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	// In-progress items should have close and defer buttons
	if !strings.Contains(body, `"status":"closed"`) {
		t.Error("expected close button for in-progress items in search")
	}
	if !strings.Contains(body, `"status":"deferred"`) {
		t.Error("expected defer button for in-progress items in search")
	}
}

func TestStalePage_RigFilter(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}, {Name: "beads_gastown"}},
		issues: []dolt.Issue{
			{ID: "st1", Title: "Stale aegis", Status: "in_progress", Priority: 1, UpdatedAt: time.Now().Add(-10 * 24 * time.Hour)},
		},
	}

	srv := New(ds)

	// Test auto-refresh URL preserves days param
	req := httptest.NewRequest("GET", "/stale?days=7", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /stale status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, `hx-get="/stale?days=7"`) {
		t.Error("expected auto-refresh URL with days param")
	}

	// Test with rig filter
	req = httptest.NewRequest("GET", "/stale?days=7&rig=beads_aegis", nil)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /stale?rig= status = %d, want %d", w.Code, http.StatusOK)
	}
	body = w.Body.String()
	if !strings.Contains(body, `&rig=beads_aegis`) {
		t.Error("expected rig filter preserved in auto-refresh URL")
	}
}

func TestClosedPage_RigFilter(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}, {Name: "beads_gastown"}},
		issues: []dolt.Issue{
			{ID: "cl1", Title: "Closed aegis", Status: "closed", Priority: 1, UpdatedAt: time.Now()},
		},
	}

	srv := New(ds)

	// Test auto-refresh URL preserves days param
	req := httptest.NewRequest("GET", "/closed?days=7", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /closed status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, `hx-get="/closed?days=7"`) {
		t.Error("expected auto-refresh URL with days param")
	}

	// Test with rig filter
	req = httptest.NewRequest("GET", "/closed?days=7&rig=beads_aegis", nil)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /closed?rig= status = %d, want %d", w.Code, http.StatusOK)
	}
	body = w.Body.String()
	if !strings.Contains(body, `&rig=beads_aegis`) {
		t.Error("expected rig filter preserved in auto-refresh URL")
	}
}

func TestParkingLotPage_RigFilter(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}, {Name: "beads_gastown"}},
		issues: []dolt.Issue{
			{ID: "pk1", Title: "Parked item", Status: "in_progress", Priority: 1,
				UpdatedAt: time.Now().Add(-10 * 24 * time.Hour),
				CreatedAt: time.Now().Add(-20 * 24 * time.Hour)},
		},
	}

	srv := New(ds)

	// Unfiltered
	req := httptest.NewRequest("GET", "/parking-lot", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /parking-lot status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Parking Lot") {
		t.Error("expected Parking Lot heading")
	}

	// With rig filter
	req = httptest.NewRequest("GET", "/parking-lot?rig=beads_aegis", nil)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /parking-lot?rig= status = %d, want %d", w.Code, http.StatusOK)
	}
	body = w.Body.String()
	if !strings.Contains(body, `?rig=beads_aegis`) {
		t.Error("expected rig filter preserved in auto-refresh URL")
	}
}

func TestDeferredPage_RigFilter(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}, {Name: "beads_gastown"}},
		issues: []dolt.Issue{
			{ID: "df1", Title: "Deferred item", Status: "deferred", Priority: 2,
				UpdatedAt: time.Now().Add(-5 * 24 * time.Hour),
				CreatedAt: time.Now().Add(-10 * 24 * time.Hour)},
		},
	}

	srv := New(ds)

	// Unfiltered
	req := httptest.NewRequest("GET", "/deferred", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /deferred status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "All rigs") {
		t.Error("expected rig filter badges when multiple rigs have data")
	}

	// With rig filter
	req = httptest.NewRequest("GET", "/deferred?rig=beads_aegis", nil)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /deferred?rig= status = %d, want %d", w.Code, http.StatusOK)
	}
	body = w.Body.String()
	if !strings.Contains(body, `?rig=beads_aegis`) {
		t.Error("expected rig filter preserved in auto-refresh URL")
	}
}

func TestTriagePage_RigFilter(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}, {Name: "beads_gastown"}},
		issues: []dolt.Issue{
			{ID: "tr1", Title: "Unassigned triage", Status: "open", Priority: 1,
				CreatedAt: time.Now().Add(-3 * 24 * time.Hour)},
		},
		assignees: []string{"alice", "bob"},
	}

	srv := New(ds)

	// Unfiltered
	req := httptest.NewRequest("GET", "/triage", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /triage status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "All rigs") {
		t.Error("expected rig filter badges when multiple rigs have data")
	}

	// With rig filter
	req = httptest.NewRequest("GET", "/triage?rig=beads_aegis", nil)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /triage?rig= status = %d, want %d", w.Code, http.StatusOK)
	}
	body = w.Body.String()
	if !strings.Contains(body, `?rig=beads_aegis`) {
		t.Error("expected rig filter preserved in triage auto-refresh URL")
	}
}

func TestQueuePage_RigFilter(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}, {Name: "beads_gastown"}},
		issues: []dolt.Issue{
			{ID: "q1", Title: "Queue item", Status: "open", Priority: 1,
				CreatedAt: time.Now().Add(-5 * 24 * time.Hour)},
		},
		assignees: []string{"alice"},
	}

	srv := New(ds)

	// Unfiltered
	req := httptest.NewRequest("GET", "/queue", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /queue status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "All rigs") {
		t.Error("expected rig filter badges when multiple rigs have data")
	}

	// With rig filter
	req = httptest.NewRequest("GET", "/queue?rig=beads_aegis", nil)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /queue?rig= status = %d, want %d", w.Code, http.StatusOK)
	}
	body = w.Body.String()
	if !strings.Contains(body, `?rig=beads_aegis`) {
		t.Error("expected rig filter preserved in queue auto-refresh URL")
	}
}

func TestWatchlistPage_RigFilter(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}, {Name: "beads_gastown"}},
		issues: []dolt.Issue{
			{ID: "wl1", Title: "Critical item", Status: "open", Priority: 0,
				CreatedAt: time.Now().Add(-2 * 24 * time.Hour),
				UpdatedAt: time.Now().Add(-1 * time.Hour)},
		},
	}

	srv := New(ds)

	// Unfiltered — should show rig filter when multiple rigs
	req := httptest.NewRequest("GET", "/watchlist", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /watchlist status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "All rigs") {
		t.Error("expected rig filter badges when multiple rigs have data")
	}

	// With rig filter
	req = httptest.NewRequest("GET", "/watchlist?rig=beads_aegis", nil)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /watchlist?rig= status = %d, want %d", w.Code, http.StatusOK)
	}
	body = w.Body.String()
	if !strings.Contains(body, `?rig=beads_aegis`) {
		t.Error("expected rig filter preserved in watchlist auto-refresh URL")
	}
}

func TestStalePage_AssigneeDropdown(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "st1", Title: "Stale work", Status: "in_progress", Priority: 1,
				Assignee:  "alice",
				UpdatedAt: time.Now().Add(-10 * 24 * time.Hour)},
		},
		assignees: []string{"alice", "bob", "charlie"},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/stale?days=3", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /stale status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "triage-assign") {
		t.Error("expected assignee dropdown with triage-assign class")
	}
	if !strings.Contains(body, "/assign") {
		t.Error("expected assign POST endpoint in dropdown")
	}
}

func TestParkingLotPage_AssigneeDropdown(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "pk1", Title: "Parked", Status: "in_progress", Priority: 1,
				Assignee:  "alice",
				UpdatedAt: time.Now().Add(-10 * 24 * time.Hour),
				CreatedAt: time.Now().Add(-20 * 24 * time.Hour)},
		},
		assignees: []string{"alice", "bob"},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/parking-lot", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "triage-assign") {
		t.Error("expected assignee dropdown on parking-lot page")
	}
}

func TestDeferredPage_AssigneeDropdown(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "df1", Title: "Deferred", Status: "deferred", Priority: 2,
				Assignee:  "bob",
				UpdatedAt: time.Now().Add(-5 * 24 * time.Hour),
				CreatedAt: time.Now().Add(-10 * 24 * time.Hour)},
		},
		assignees: []string{"alice", "bob"},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/deferred", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "triage-assign") {
		t.Error("expected assignee dropdown on deferred page")
	}
}

func TestBlockedPage_RigFilter(t *testing.T) {
	ds := &mockDataSource{
		databases:     []dolt.DatabaseInfo{{Name: "beads_aegis"}, {Name: "beads_gastown"}},
		blockedIssues: []dolt.BlockedIssue{
			{
				Issue:   dolt.Issue{ID: "aegis-b1", Title: "Needs auth", Status: "open", Priority: 1, Assignee: "aegis/crew/arnold", UpdatedAt: time.Now()},
				Blocker: dolt.Issue{ID: "aegis-b2", Title: "Auth module", Status: "in_progress", Owner: "aegis/crew/grant"},
			},
		},
	}

	srv := New(ds)

	// Test without filter — should show rig filter badges
	req := httptest.NewRequest("GET", "/blocked", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /blocked status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "All rigs") {
		t.Error("expected 'All rigs' filter badge")
	}

	// Test with rig filter
	req = httptest.NewRequest("GET", "/blocked?rig=beads_aegis", nil)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /blocked?rig= status = %d, want %d", w.Code, http.StatusOK)
	}
	body = w.Body.String()
	if !strings.Contains(body, `rig=beads_aegis`) {
		t.Error("expected rig filter preserved in auto-refresh URL")
	}
}

func TestBlockedPage_AssigneeDropdown(t *testing.T) {
	ds := &mockDataSource{
		databases:     []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		blockedIssues: []dolt.BlockedIssue{
			{
				Issue:   dolt.Issue{ID: "aegis-b1", Title: "Needs auth", Status: "open", Priority: 1, Assignee: "aegis/crew/arnold", UpdatedAt: time.Now()},
				Blocker: dolt.Issue{ID: "aegis-b2", Title: "Auth module", Status: "in_progress", Owner: "aegis/crew/grant"},
			},
		},
		assignees: []string{"aegis/crew/arnold", "aegis/crew/grant"},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/blocked", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "triage-assign") {
		t.Error("expected assignee dropdown on blocked page")
	}
}

func TestChurnPage_RigFilter(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}, {Name: "beads_gastown"}},
		issues: []dolt.Issue{
			{ID: "ch1", Title: "Churning", Status: "open", Priority: 2, UpdatedAt: time.Now()},
		},
		statusHistory: []dolt.StatusTransition{
			{ToStatus: "open", CommitDate: time.Now().Add(-3 * 24 * time.Hour)},
			{FromStatus: "open", ToStatus: "in_progress", CommitDate: time.Now().Add(-2 * 24 * time.Hour)},
			{FromStatus: "in_progress", ToStatus: "open", CommitDate: time.Now().Add(-1 * 24 * time.Hour)},
		},
	}

	srv := New(ds)

	// Test without filter
	req := httptest.NewRequest("GET", "/churn", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /churn status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "All rigs") {
		t.Error("expected 'All rigs' filter badge")
	}

	// Test with rig filter
	req = httptest.NewRequest("GET", "/churn?rig=beads_aegis", nil)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /churn?rig= status = %d, want %d", w.Code, http.StatusOK)
	}
	body = w.Body.String()
	if !strings.Contains(body, `rig=beads_aegis`) {
		t.Error("expected rig filter preserved in auto-refresh URL")
	}
}

func TestWorkloadPage_RigFilter(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}, {Name: "beads_gastown"}},
		issues: []dolt.Issue{
			{ID: "wl1", Title: "Work item", Status: "open", Priority: 1, Assignee: "aegis/crew/arnold", UpdatedAt: time.Now()},
		},
	}

	srv := New(ds)

	// Test without filter
	req := httptest.NewRequest("GET", "/workload", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /workload status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "All rigs") {
		t.Error("expected 'All rigs' filter badge")
	}

	// Test with rig filter
	req = httptest.NewRequest("GET", "/workload?rig=beads_aegis", nil)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /workload?rig= status = %d, want %d", w.Code, http.StatusOK)
	}
	body = w.Body.String()
	if !strings.Contains(body, `rig=beads_aegis`) {
		t.Error("expected rig filter preserved in auto-refresh URL")
	}
}

func TestEpicsPage_RigFilter(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}, {Name: "beads_gastown"}},
		issues: []dolt.Issue{
			{ID: "ep1", Title: "Epic one", Status: "open", Type: "epic", Priority: 1, UpdatedAt: time.Now()},
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
	if !strings.Contains(body, "All rigs") {
		t.Error("expected 'All rigs' filter badge")
	}

	req = httptest.NewRequest("GET", "/epics?rig=beads_aegis", nil)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /epics?rig= status = %d, want %d", w.Code, http.StatusOK)
	}
	body = w.Body.String()
	if !strings.Contains(body, `rig=beads_aegis`) {
		t.Error("expected rig filter preserved in auto-refresh URL")
	}
}

func TestDuplicatesPage_RigFilter(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}, {Name: "beads_gastown"}},
		issues: []dolt.Issue{
			{ID: "d1", Title: "Dup title", Status: "open", Priority: 2, CreatedAt: time.Now()},
			{ID: "d2", Title: "Dup title", Status: "open", Priority: 2, CreatedAt: time.Now()},
		},
	}

	srv := New(ds)

	req := httptest.NewRequest("GET", "/duplicates", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /duplicates status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "All rigs") {
		t.Error("expected 'All rigs' filter badge")
	}

	req = httptest.NewRequest("GET", "/duplicates?rig=beads_aegis", nil)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /duplicates?rig= status = %d, want %d", w.Code, http.StatusOK)
	}
	body = w.Body.String()
	if !strings.Contains(body, `rig=beads_aegis`) {
		t.Error("expected rig filter preserved in auto-refresh URL")
	}
}

func TestActivityPage_AssigneeDropdown(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "act1", Title: "Recent work", Status: "open", Priority: 1, Assignee: "aegis/crew/arnold", UpdatedAt: time.Now()},
		},
		assignees: []string{"aegis/crew/arnold", "aegis/crew/grant"},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/activity", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "triage-assign") {
		t.Error("expected assignee dropdown on activity page")
	}
}

func TestWatchlistPage_AssigneeDropdown(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "wl1", Title: "P0 item", Status: "open", Priority: 0, Assignee: "aegis/crew/arnold", UpdatedAt: time.Now(), CreatedAt: time.Now()},
		},
		assignees: []string{"aegis/crew/arnold", "aegis/crew/grant"},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/watchlist", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "triage-assign") {
		t.Error("expected assignee dropdown on watchlist page")
	}
}

func TestKanbanPage_RigFilter(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}, {Name: "beads_gastown"}},
		issues: []dolt.Issue{
			{ID: "k1", Title: "Card one", Status: "open", Priority: 1, UpdatedAt: time.Now()},
		},
	}

	srv := New(ds)

	// Test with rig filter preserves URL param
	req := httptest.NewRequest("GET", "/kanban?rig=beads_aegis", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /kanban?rig= status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, `rig=beads_aegis`) {
		t.Error("expected rig filter preserved in auto-refresh URL")
	}
	if !strings.Contains(body, "Kanban Board") {
		t.Error("expected kanban board heading")
	}
}

func TestActivityPage_InlinePriority(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "act1", Title: "Recent work", Status: "open", Priority: 2, UpdatedAt: time.Now()},
		},
		assignees: []string{"aegis/crew/arnold"},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/activity", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "bead-priority-inline") {
		t.Error("expected inline priority editing on activity page")
	}
	if !strings.Contains(body, "/priority") {
		t.Error("expected priority POST endpoint in activity template")
	}
}

func TestStalePage_InlinePriority(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "st1", Title: "Stale item", Status: "in_progress", Priority: 1, UpdatedAt: time.Now().Add(-10 * 24 * time.Hour)},
		},
		assignees: []string{"aegis/crew/arnold"},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/stale", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "bead-priority-inline") {
		t.Error("expected inline priority editing on stale page")
	}
}

func TestBlockedPage_InlinePriority(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		blockedIssues: []dolt.BlockedIssue{
			{
				Issue:   dolt.Issue{ID: "b1", Title: "Blocked", Status: "open", Priority: 1, Assignee: "aegis/crew/arnold", UpdatedAt: time.Now()},
				Blocker: dolt.Issue{ID: "b2", Title: "Blocker", Status: "in_progress", Owner: "aegis/crew/grant"},
			},
		},
		assignees: []string{"aegis/crew/arnold"},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/blocked", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "bead-priority-inline") {
		t.Error("expected inline priority editing on blocked page")
	}
}

func TestClosedPage_AssigneeDropdown(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "c1", Title: "Done item", Status: "closed", Priority: 2, Assignee: "aegis/crew/arnold", UpdatedAt: time.Now()},
		},
		assignees: []string{"aegis/crew/arnold", "aegis/crew/grant"},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/closed", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "triage-assign") {
		t.Error("expected assignee dropdown on closed page")
	}
}

func TestDeferredPage_InlinePriority(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "df1", Title: "Deferred item", Status: "deferred", Priority: 3, UpdatedAt: time.Now(), CreatedAt: time.Now()},
		},
		assignees: []string{"aegis/crew/arnold"},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/deferred", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "bead-priority-inline") {
		t.Error("expected inline priority editing on deferred page")
	}
}

func TestParkingLotPage_InlinePriority(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "pl1", Title: "Stalled item", Status: "in_progress", Priority: 2, UpdatedAt: time.Now().Add(-5 * 24 * time.Hour), CreatedAt: time.Now().Add(-10 * 24 * time.Hour)},
		},
		assignees: []string{"aegis/crew/arnold"},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/parking-lot", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "bead-priority-inline") {
		t.Error("expected inline priority editing on parking-lot page")
	}
}

func TestWatchlistPage_InlinePriority(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "wlp1", Title: "P0 critical", Status: "open", Priority: 0, Assignee: "aegis/crew/arnold", UpdatedAt: time.Now(), CreatedAt: time.Now()},
		},
		assignees: []string{"aegis/crew/arnold"},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/watchlist", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "bead-priority-inline") {
		t.Error("expected inline priority editing on watchlist page")
	}
	if !strings.Contains(body, "/priority") {
		t.Error("expected priority POST endpoint in watchlist template")
	}
}

func TestWatchlistPage_BatchActions(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "wlb1", Title: "P1 item", Status: "open", Priority: 1, UpdatedAt: time.Now(), CreatedAt: time.Now()},
		},
		assignees: []string{"aegis/crew/arnold"},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/watchlist", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "batch-cb") {
		t.Error("expected batch checkboxes on watchlist page")
	}
	if !strings.Contains(body, "batch-bar") {
		t.Error("expected batch action bar on watchlist page")
	}
	if !strings.Contains(body, "batchAction") {
		t.Error("expected batch action JS on watchlist page")
	}
}

func TestEpicsPage_InlinePriority(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "ep1", Title: "Epic one", Status: "open", Type: "epic", Priority: 1, UpdatedAt: time.Now(), CreatedAt: time.Now()},
		},
		assignees: []string{"aegis/crew/arnold"},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/epics", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "bead-priority-inline") {
		t.Error("expected inline priority editing on epics page")
	}
	if !strings.Contains(body, "/priority") {
		t.Error("expected priority POST endpoint in epics template")
	}
}

func TestEpicsPage_AssigneeDropdown(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "ep2", Title: "Epic two", Status: "open", Type: "epic", Priority: 1, Assignee: "aegis/crew/arnold", UpdatedAt: time.Now(), CreatedAt: time.Now()},
		},
		assignees: []string{"aegis/crew/arnold", "aegis/crew/grant"},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/epics", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "triage-assign") {
		t.Error("expected assignee dropdown on epics page")
	}
}

func TestEpicsPage_BatchActions(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "ep3", Title: "Epic three", Status: "open", Type: "epic", Priority: 2, UpdatedAt: time.Now(), CreatedAt: time.Now()},
		},
		assignees: []string{"aegis/crew/arnold"},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/epics", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "batch-cb") {
		t.Error("expected batch checkboxes on epics page")
	}
	if !strings.Contains(body, "batch-bar") {
		t.Error("expected batch action bar on epics page")
	}
}

func TestDuplicatesPage_InlinePriority(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "dup1", Title: "[AUTO] test alert", Status: "open", Priority: 2, UpdatedAt: time.Now(), CreatedAt: time.Now()},
			{ID: "dup2", Title: "[AUTO] test alert", Status: "open", Priority: 2, UpdatedAt: time.Now().Add(-time.Hour), CreatedAt: time.Now().Add(-time.Hour)},
		},
		assignees: []string{"aegis/crew/arnold"},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/duplicates", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "bead-priority-inline") {
		t.Error("expected inline priority editing on duplicates page")
	}
}

func TestCommentsPage_FilterPreservation(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		comments: []dolt.Comment{
			{IssueID: "c1", Author: "aegis/crew/arnold", Body: "test", CreatedAt: time.Now()},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/comments?author=aegis/crew/arnold", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, `hx-get="/comments?author=aegis/crew/arnold"`) {
		t.Error("expected auto-refresh to preserve author filter")
	}
}

func TestLabelsPage_FilterPreservation(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		labelCounts: []dolt.LabelCount{{Label: "bug", Count: 3}},
		issues: []dolt.Issue{
			{ID: "lb1", Title: "Bug one", Status: "open", Priority: 1, UpdatedAt: time.Now(), CreatedAt: time.Now()},
		},
		assignees: []string{"aegis/crew/arnold"},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/labels?label=bug", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, `hx-get="/labels?label=bug"`) {
		t.Error("expected auto-refresh to preserve label filter")
	}
}

func TestLabelsPage_InlinePriority(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		labelCounts: []dolt.LabelCount{{Label: "bug", Count: 1}},
		issues: []dolt.Issue{
			{ID: "lbp1", Title: "Bug with label", Status: "open", Priority: 2, UpdatedAt: time.Now(), CreatedAt: time.Now()},
		},
		assignees: []string{"aegis/crew/arnold"},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/labels?label=bug", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "bead-priority-inline") {
		t.Error("expected inline priority editing on labels page")
	}
	if !strings.Contains(body, "triage-assign") {
		t.Error("expected assignee dropdown on labels page")
	}
	if !strings.Contains(body, "batch-cb") {
		t.Error("expected batch checkboxes on labels page")
	}
}

func TestWatchlistPage_StartButton(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "wls1", Title: "Open P0", Status: "open", Priority: 0, UpdatedAt: time.Now(), CreatedAt: time.Now()},
		},
		assignees: []string{"aegis/crew/arnold"},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/watchlist", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, `"status":"in_progress"`) {
		t.Error("expected start button on watchlist page")
	}
}

func TestTriagePage_StartButton(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "trs1", Title: "Untriaged", Status: "open", Priority: 3, UpdatedAt: time.Now(), CreatedAt: time.Now()},
		},
		assignees: []string{"aegis/crew/arnold"},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/triage", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, `"status":"in_progress"`) {
		t.Error("expected start button on triage page")
	}
}

func TestLabelsPage_StartAndReopen(t *testing.T) {
	ds := &mockDataSource{
		databases:   []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		labelCounts: []dolt.LabelCount{{Label: "bug", Count: 2}},
		issues: []dolt.Issue{
			{ID: "lbs1", Title: "Open bug", Status: "open", Priority: 1, UpdatedAt: time.Now(), CreatedAt: time.Now()},
			{ID: "lbs2", Title: "Closed bug", Status: "closed", Priority: 2, UpdatedAt: time.Now(), CreatedAt: time.Now()},
		},
		assignees: []string{"aegis/crew/arnold"},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/labels?label=bug", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, `"status":"in_progress"`) {
		t.Error("expected start button on labels page for open beads")
	}
	if !strings.Contains(body, ">reopen<") {
		t.Error("expected reopen button on labels page for closed beads")
	}
}

func TestDepsPage_FilterPreservation(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/deps?type=child_of", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, `hx-get="/deps?type=child_of"`) {
		t.Error("expected auto-refresh to preserve type filter on deps page")
	}
}

func TestHandoffsPage_FilterPreservation(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/handoffs?actor=aegis/crew/arnold", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, `hx-get="/handoffs?actor=aegis/crew/arnold"`) {
		t.Error("expected auto-refresh to preserve actor filter on handoffs page")
	}
}

func TestCommentsPage_RigFilter(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}, {Name: "beads_gastown"}},
		comments: []dolt.Comment{
			{IssueID: "c1", Author: "aegis/crew/arnold", Body: "test", CreatedAt: time.Now()},
		},
	}

	srv := New(ds)

	// Without filter
	req := httptest.NewRequest("GET", "/comments", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "All rigs") {
		t.Error("expected rig filter bar on comments page with multiple rigs")
	}

	// With rig filter
	req = httptest.NewRequest("GET", "/comments?rig=beads_aegis", nil)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body = w.Body.String()
	if !strings.Contains(body, "rig=beads_aegis") {
		t.Error("expected rig filter preserved in auto-refresh URL")
	}
}

func TestDepsPage_RigFilter(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}, {Name: "beads_gastown"}},
	}

	srv := New(ds)

	// Without filter — should show rig filter bar
	req := httptest.NewRequest("GET", "/deps", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "All rigs") {
		t.Error("expected rig filter bar on deps page with multiple rigs")
	}

	// With rig filter — should preserve in URL
	req = httptest.NewRequest("GET", "/deps?rig=beads_aegis&type=child_of", nil)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body = w.Body.String()
	if !strings.Contains(body, "rig=beads_aegis") {
		t.Error("expected rig filter preserved in auto-refresh URL")
	}
	if !strings.Contains(body, "type=child_of") {
		t.Error("expected type filter preserved alongside rig filter")
	}
}
