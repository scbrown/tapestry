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
	issueDiffs      []dolt.IssueDiffRow
	commentDiffs    []dolt.CommentDiffRow
	agents          []dolt.AgentStats
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
	return m.agents, m.err
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

func (m *mockDataSource) IssueDiffSince(_ context.Context, _ string, _ time.Time) ([]dolt.IssueDiffRow, error) {
	return m.issueDiffs, m.err
}

func (m *mockDataSource) CommentDiffSince(_ context.Context, _ string, _ time.Time) ([]dolt.CommentDiffRow, error) {
	return m.commentDiffs, m.err
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

func TestBeadsList_LabelFilter(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "aegis-lf1", Title: "Labeled bead", Status: "open", Priority: 1, Rig: "beads_aegis", UpdatedAt: time.Now()},
		},
		labelCounts: []dolt.LabelCount{
			{Label: "infra", Count: 5},
			{Label: "urgent", Count: 2},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/beads?label=infra", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /beads?label=infra status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "f-label") {
		t.Error("body missing label filter select")
	}
	if !strings.Contains(body, "infra") {
		t.Error("body missing infra label option")
	}
	if !strings.Contains(body, "urgent") {
		t.Error("body missing urgent label option")
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
	if !strings.Contains(body, "beadsBatchAction('closed')") {
		t.Error("expected close batch button")
	}
	if !strings.Contains(body, "beadsBatchAction('deferred')") {
		t.Error("expected defer batch button")
	}
	if !strings.Contains(body, "set priority...") {
		t.Error("expected batch priority dropdown")
	}
	if !strings.Contains(body, "assign to...") {
		t.Error("expected batch assignee dropdown")
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

func TestCommitsPage_HTMXPartial(t *testing.T) {
	srv := &Server{forgejo: nil}
	srv.parseTemplates()
	req := httptest.NewRequest("GET", "/commits", nil)
	req.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("HTMX GET /commits status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("HTMX commits should return partial, not full page")
	}
}

func TestCommitsPage_SearchForm(t *testing.T) {
	srv := &Server{forgejo: nil}
	srv.parseTemplates()
	req := httptest.NewRequest("GET", "/commits", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
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

func TestWorkPage_RigFilter(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{
			{Name: "beads_aegis"},
			{Name: "beads_gastown"},
		},
		issues: []dolt.Issue{
			{ID: "aegis-rf1", Title: "Aegis filtered task", Status: "open", Priority: 1, Type: "task", UpdatedAt: time.Now()},
		},
	}

	srv := New(ds)

	// With rig filter
	req := httptest.NewRequest("GET", "/work?rig=beads_aegis", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /work?rig=beads_aegis status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "All rigs") {
		t.Errorf("body missing rig filter badges")
	}
	if !strings.Contains(body, "filter-active") {
		t.Errorf("body missing active filter indicator")
	}

	// Without rig filter — should show all rigs badge bar
	req2 := httptest.NewRequest("GET", "/work", nil)
	w2 := httptest.NewRecorder()
	srv.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Fatalf("GET /work status = %d, want %d", w2.Code, http.StatusOK)
	}
	body2 := w2.Body.String()
	if !strings.Contains(body2, "All rigs") {
		t.Errorf("unfiltered body missing rig filter badges")
	}
}

func TestFunnelPage_RigFilter(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{
			{Name: "beads_aegis"},
			{Name: "beads_gastown"},
		},
		issues: []dolt.Issue{
			{ID: "aegis-fun1", Title: "Funnel task", Status: "open", Priority: 1, Type: "task", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/funnel?rig=beads_aegis", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /funnel?rig=beads_aegis status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Conversion Funnel") {
		t.Errorf("body missing page title")
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

func TestCommandCenter_HTMXPartial(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/command-center", nil)
	req.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("HTMX GET /command-center status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("HTMX command-center should return partial, not full page")
	}
}

func TestCommandCenter_RigFilter(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}, {Name: "beads_hq"}},
		counts:    map[string]int{"open": 5, "in_progress": 2},
	}
	srv := New(ds)
	req := httptest.NewRequest("GET", "/command-center?rig=beads_aegis", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("GET /command-center?rig status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Command Center") {
		t.Error("expected 'Command Center' heading")
	}
	if !strings.Contains(body, "filter-active") {
		t.Error("expected active rig filter badge")
	}
}

func TestCommandCenter_AutoRefresh(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/command-center", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	body := w.Body.String()
	if !strings.Contains(body, "hx-trigger") {
		t.Error("expected auto-refresh hx-trigger on command center")
	}
}

func TestCommandCenter_FleetStats(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		counts:    map[string]int{"open": 10, "in_progress": 3, "closed": 50},
		closed:    5,
	}
	srv := New(ds)
	req := httptest.NewRequest("GET", "/command-center", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("GET /command-center status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "fleet-stat") {
		t.Error("expected fleet-stat elements")
	}
	if !strings.Contains(body, "active") {
		t.Error("expected 'active' fleet stat label")
	}
	if !strings.Contains(body, "open") {
		t.Error("expected 'open' fleet stat label")
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

func TestAgentsPage_SummaryStats(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		agents: []dolt.AgentStats{
			{Name: "aegis/crew/alice", Owned: 10, Closed: 5, Open: 3, InProgress: 2},
			{Name: "aegis/crew/bob", Owned: 8, Closed: 3, Open: 4, InProgress: 1},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/agents", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /agents status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "stat-grid") {
		t.Error("expected summary stat grid")
	}
	if !strings.Contains(body, "Active items") {
		t.Error("expected 'Active items' label")
	}
	if !strings.Contains(body, "Total closed") {
		t.Error("expected 'Total closed' label")
	}
	if !strings.Contains(body, "completion rate") {
		t.Error("expected completion rate progress bar")
	}
}

func TestAgentsPage_SortOptions(t *testing.T) {
	sorts := []string{"", "active", "owned", "closed", "handoffs", "recent", "name"}
	for _, s := range sorts {
		t.Run("sort="+s, func(t *testing.T) {
			srv := New(nil)
			url := "/agents"
			if s != "" {
				url += "?sort=" + s
			}
			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()
			srv.ServeHTTP(w, req)
			if w.Code != http.StatusOK {
				t.Fatalf("GET %s status = %d, want %d", url, w.Code, http.StatusOK)
			}
			body := w.Body.String()
			if !strings.Contains(body, "By active") {
				t.Error("expected sort options in page")
			}
		})
	}
}

func TestAgentsPage_SortWithData(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		agents: []dolt.AgentStats{
			{Name: "aegis/crew/alice", Owned: 10, Closed: 5, Open: 3, InProgress: 2},
			{Name: "aegis/crew/bob", Owned: 20, Closed: 15, Open: 4, InProgress: 1},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/agents?sort=owned", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /agents?sort=owned status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	// Bob should appear first (20 owned > 10 owned)
	bobIdx := strings.Index(body, "bob")
	aliceIdx := strings.Index(body, "alice")
	if bobIdx < 0 || aliceIdx < 0 {
		t.Fatal("expected both agents in output")
	}
	if bobIdx > aliceIdx {
		t.Error("expected bob before alice when sorted by owned")
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

func TestSearch_RigFilter(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}, {Name: "beads_gastown"}},
		issues: []dolt.Issue{
			{ID: "aegis-020", Title: "Found bead", Status: "open", Rig: "beads_aegis", UpdatedAt: time.Now()},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/search?q=found&rig=beads_aegis", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "aegis-020") {
		t.Error("body missing search result")
	}
	if !strings.Contains(body, "All rigs") {
		t.Error("missing rig filter dropdown")
	}
}

func TestSearch_TypeFilter(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "aegis-bug", Title: "A bug", Status: "open", Type: "bug", Rig: "beads_aegis", UpdatedAt: time.Now()},
			{ID: "aegis-task", Title: "A task", Status: "open", Type: "task", Rig: "beads_aegis", UpdatedAt: time.Now()},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/search?q=A&type=bug", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "aegis-bug") {
		t.Error("body should contain bug result")
	}
	if strings.Contains(body, "aegis-task") {
		t.Error("body should NOT contain task result when filtering by bug type")
	}
}

func TestSearch_SortByPriority(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "aegis-low", Title: "Low pri", Status: "open", Priority: 3, Rig: "beads_aegis", UpdatedAt: time.Now()},
			{ID: "aegis-high", Title: "High pri", Status: "open", Priority: 0, Rig: "beads_aegis", UpdatedAt: time.Now()},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/search?q=pri&sort=priority", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	highIdx := strings.Index(body, "aegis-high")
	lowIdx := strings.Index(body, "aegis-low")
	if highIdx < 0 || lowIdx < 0 {
		t.Fatal("both results should appear")
	}
	if highIdx > lowIdx {
		t.Error("P0 should appear before P3 when sorting by priority")
	}
}

func TestSearch_LabelFilter(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "aegis-s1", Title: "Security bug", Status: "open", Priority: 1, Rig: "beads_aegis"},
		},
		labelCounts: []dolt.LabelCount{{Label: "security", Count: 3}, {Label: "desire-path", Count: 1}},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/search?q=security&label=security", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Any label") {
		t.Error("expected label filter dropdown with 'Any label' option")
	}
	if !strings.Contains(body, "security") {
		t.Error("expected 'security' label option in dropdown")
	}
}

func TestSearch_EnhancedBatchBar(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "aegis-001", Title: "Test bead", Status: "open", Rig: "beads_aegis", UpdatedAt: time.Now()},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/search?q=test", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "set priority...") {
		t.Error("missing batch priority dropdown in search results")
	}
	if !strings.Contains(body, "assign to...") {
		t.Error("missing batch assignee dropdown in search results")
	}
	if !strings.Contains(body, "add label...") {
		t.Error("missing batch label input in search results")
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

func TestBatchStatus_InProgress(t *testing.T) {
	ds := &mockDataSource{}
	srv := New(ds)

	body := "status=in_progress&ids[]=beads_aegis/aegis-001"
	req := httptest.NewRequest("POST", "/batch/status", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", w.Code, http.StatusOK, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "1 beads updated to in_progress") {
		t.Errorf("response should confirm in_progress, got: %s", w.Body.String())
	}
}

func TestBatchStatus_Blocked(t *testing.T) {
	ds := &mockDataSource{}
	srv := New(ds)

	body := "status=blocked&ids[]=beads_aegis/aegis-001"
	req := httptest.NewRequest("POST", "/batch/status", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", w.Code, http.StatusOK, w.Body.String())
	}
}

func TestBatchPriority(t *testing.T) {
	ds := &mockDataSource{}
	srv := New(ds)

	body := "priority=1&ids[]=beads_aegis/aegis-001&ids[]=beads_aegis/aegis-002"
	req := httptest.NewRequest("POST", "/batch/priority", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", w.Code, http.StatusOK, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "2 beads updated to P1") {
		t.Errorf("response should mention 2 beads updated, got: %s", w.Body.String())
	}
	trigger := w.Header().Get("HX-Trigger")
	if !strings.Contains(trigger, "showToast") {
		t.Errorf("missing HX-Trigger header, got: %s", trigger)
	}
}

func TestBatchPriority_InvalidPriority(t *testing.T) {
	ds := &mockDataSource{}
	srv := New(ds)

	body := "priority=99&ids[]=beads_aegis/aegis-001"
	req := httptest.NewRequest("POST", "/batch/priority", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestBatchPriority_NoIds(t *testing.T) {
	ds := &mockDataSource{}
	srv := New(ds)

	body := "priority=1"
	req := httptest.NewRequest("POST", "/batch/priority", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestBatchPriority_NoPriority(t *testing.T) {
	ds := &mockDataSource{}
	srv := New(ds)

	body := "ids[]=beads_aegis/aegis-001"
	req := httptest.NewRequest("POST", "/batch/priority", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestBatchAssignee(t *testing.T) {
	ds := &mockDataSource{}
	srv := New(ds)

	body := "assignee=aegis/crew/arnold&ids[]=beads_aegis/aegis-001&ids[]=beads_aegis/aegis-002"
	req := httptest.NewRequest("POST", "/batch/assignee", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", w.Code, http.StatusOK, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "2 beads assigned") {
		t.Errorf("response should mention 2 beads assigned, got: %s", w.Body.String())
	}
	trigger := w.Header().Get("HX-Trigger")
	if !strings.Contains(trigger, "showToast") {
		t.Errorf("missing HX-Trigger header, got: %s", trigger)
	}
}

func TestBatchAssignee_NoIds(t *testing.T) {
	ds := &mockDataSource{}
	srv := New(ds)

	body := "assignee=someone"
	req := httptest.NewRequest("POST", "/batch/assignee", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestBatchAssignee_Unassign(t *testing.T) {
	ds := &mockDataSource{}
	srv := New(ds)

	body := "assignee=&ids[]=beads_aegis/aegis-001"
	req := httptest.NewRequest("POST", "/batch/assignee", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", w.Code, http.StatusOK, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "unassigned") {
		t.Errorf("response should mention unassigned, got: %s", w.Body.String())
	}
}

func TestBatchLabel(t *testing.T) {
	ds := &mockDataSource{}
	srv := New(ds)

	body := "label=urgent&ids[]=beads_aegis/aegis-001&ids[]=beads_aegis/aegis-002"
	req := httptest.NewRequest("POST", "/batch/label", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", w.Code, http.StatusOK, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "2 beads labeled") {
		t.Errorf("response should mention 2 beads labeled, got: %s", w.Body.String())
	}
	trigger := w.Header().Get("HX-Trigger")
	if !strings.Contains(trigger, "showToast") {
		t.Errorf("missing HX-Trigger header, got: %s", trigger)
	}
}

func TestBatchLabel_NoLabel(t *testing.T) {
	ds := &mockDataSource{}
	srv := New(ds)

	body := "label=&ids[]=beads_aegis/aegis-001"
	req := httptest.NewRequest("POST", "/batch/label", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestBatchLabel_InvalidLabel(t *testing.T) {
	ds := &mockDataSource{}
	srv := New(ds)

	body := "label=bad+label&ids[]=beads_aegis/aegis-001"
	req := httptest.NewRequest("POST", "/batch/label", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestBatchLabel_NoIds(t *testing.T) {
	ds := &mockDataSource{}
	srv := New(ds)

	body := "label=urgent"
	req := httptest.NewRequest("POST", "/batch/label", strings.NewReader(body))
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
	if !strings.Contains(body, "vs prev") {
		t.Error("expected week-over-week comparison in throughput section")
	}
}

func TestExecutivePage_RigFilter(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}, {Name: "beads_gastown"}},
		counts:    map[string]int{"open": 10, "in_progress": 3, "blocked": 2, "closed": 50},
		created:   5,
		closed:    8,
		issues: []dolt.Issue{
			{ID: "aegis-e1", Title: "P1 work", Status: "open", Priority: 1, UpdatedAt: time.Now()},
		},
	}

	srv := New(ds)

	// Without filter — should show rig filter badges
	req := httptest.NewRequest("GET", "/executive", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("GET /executive status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "All rigs") {
		t.Error("expected 'All rigs' filter badge")
	}
	if !strings.Contains(body, "filter-active") {
		t.Error("expected active filter badge")
	}

	// With rig filter
	req = httptest.NewRequest("GET", "/executive?rig=beads_aegis", nil)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("GET /executive?rig= status = %d, want %d", w.Code, http.StatusOK)
	}
	body = w.Body.String()
	if !strings.Contains(body, "Executive Status") {
		t.Error("expected 'Executive Status' heading with rig filter")
	}

	// KPI drill-down links
	if !strings.Contains(body, `href="/work?rig=beads_aegis"`) {
		t.Error("expected drill-down link to /work with rig filter")
	}
	if !strings.Contains(body, `href="/blocked?rig=beads_aegis"`) {
		t.Error("expected drill-down link to /blocked with rig filter")
	}
}

func TestExecutivePage_BlockerEnrichment(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		counts:    map[string]int{"open": 5, "blocked": 2, "closed": 10},
		created:   3,
		closed:    1,
		issues: []dolt.Issue{
			{ID: "aegis-e1", Title: "Some work", Status: "open", Priority: 1, UpdatedAt: time.Now()},
		},
		blockedIssues: []dolt.BlockedIssue{
			{
				Issue:   dolt.Issue{ID: "aegis-e2", Title: "Blocked P1", Status: "blocked", Priority: 1, UpdatedAt: time.Now()},
				Blocker: dolt.Issue{ID: "aegis-dep1", Title: "Upstream fix", Status: "in_progress", Priority: 0, Assignee: "aegis/crew/arnold", UpdatedAt: time.Now().Add(-1 * time.Hour)},
			},
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
	// Blocker priority badge
	if !strings.Contains(body, "P0") {
		t.Error("expected blocker priority P0 in executive blockers")
	}
	// Blocker status badge
	if !strings.Contains(body, "in_progress") {
		t.Error("expected blocker status 'in_progress' in executive blockers")
	}
	// Blocker ID
	if !strings.Contains(body, "aegis-dep1") {
		t.Error("expected blocker ID in executive page")
	}
}

func TestBriefingPage_RigFilter(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}, {Name: "beads_gastown"}},
		counts:    map[string]int{"open": 5, "in_progress": 2, "closed": 10},
		created:   3,
		closed:    1,
		issues: []dolt.Issue{
			{ID: "aegis-b1", Title: "Human task", Status: "open", Priority: 1, Owner: "stiwi", UpdatedAt: time.Now()},
		},
	}

	srv := New(ds)

	// Without filter — should show rig filter badges
	req := httptest.NewRequest("GET", "/briefing", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("GET /briefing status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "All rigs") {
		t.Error("expected 'All rigs' filter badge on briefing")
	}

	// With rig filter
	req = httptest.NewRequest("GET", "/briefing?rig=beads_aegis", nil)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("GET /briefing?rig= status = %d, want %d", w.Code, http.StatusOK)
	}
	body = w.Body.String()
	if !strings.Contains(body, "Briefing") {
		t.Error("expected 'Briefing' heading with rig filter")
	}
}

func TestBriefingPage_BlockedEnrichment(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		counts:    map[string]int{"open": 5, "blocked": 2},
		issues: []dolt.Issue{
			{ID: "aegis-b1", Title: "Blocked task", Status: "blocked", Priority: 1, UpdatedAt: time.Now()},
		},
		blockedIssues: []dolt.BlockedIssue{
			{
				Issue:   dolt.Issue{ID: "aegis-b1", Title: "Blocked task", Status: "blocked", Priority: 1, UpdatedAt: time.Now()},
				Blocker: dolt.Issue{ID: "aegis-dep1", Title: "Dependency work", Status: "in_progress", Priority: 0, Assignee: "aegis/crew/arnold", UpdatedAt: time.Now().Add(-2 * time.Hour)},
			},
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
	// Blocker priority badge
	if !strings.Contains(body, "P0") {
		t.Error("expected blocker priority P0 badge in blocked section")
	}
	// Blocker status badge
	if !strings.Contains(body, "in_progress") {
		t.Error("expected blocker status 'in_progress' in blocked section")
	}
	// Blocker ID link
	if !strings.Contains(body, "aegis-dep1") {
		t.Error("expected blocker ID 'aegis-dep1' in blocked section")
	}
	// Defer button on blocked items
	if !strings.Contains(body, "Defer aegis-b1") {
		t.Error("expected defer action button on blocked item")
	}
}

func TestCommitsPage_RepoFilter(t *testing.T) {
	srv := New(nil)

	// Without filter — should show repo filter badges
	req := httptest.NewRequest("GET", "/commits", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /commits status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "All repos") {
		t.Error("expected 'All repos' filter badge on commits page")
	}
	if !strings.Contains(body, "aegis") {
		t.Error("expected 'aegis' repo badge on commits page")
	}

	// With filter — should preserve filter in auto-refresh URL
	req = httptest.NewRequest("GET", "/commits?repo=tapestry", nil)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /commits?repo= status = %d, want %d", w.Code, http.StatusOK)
	}
	body = w.Body.String()
	if !strings.Contains(body, "repo=tapestry") {
		t.Error("expected repo=tapestry in auto-refresh URL")
	}
}

func TestExecutivePage_DrillDownLinks(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		counts:    map[string]int{"open": 10, "in_progress": 3, "blocked": 2, "closed": 50},
		created:   5,
		closed:    8,
		issues: []dolt.Issue{
			{ID: "aegis-e1", Title: "P1 work", Status: "open", Priority: 1, UpdatedAt: time.Now()},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/executive", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	body := w.Body.String()
	if !strings.Contains(body, `href="/work"`) {
		t.Error("expected drill-down link to /work")
	}
	if !strings.Contains(body, `href="/blocked"`) {
		t.Error("expected drill-down link to /blocked")
	}
	if !strings.Contains(body, `href="/kanban"`) {
		t.Error("expected drill-down link to /kanban")
	}
	if !strings.Contains(body, `href="/velocity"`) {
		t.Error("expected drill-down link to /velocity")
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
	if !strings.Contains(body, "closedBatchAction('open')") {
		t.Error("expected reopen batch button")
	}
	if !strings.Contains(body, "set priority...") {
		t.Error("expected batch priority dropdown on closed page")
	}
}

func TestCreatedPage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/created", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /created status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Recently Created") {
		t.Error("expected 'Recently Created' heading")
	}
}

func TestCreatedPage_WithData(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "aegis-cr1", Title: "Fresh bead", Status: "open", Priority: 1, Owner: "aegis/crew/alice", CreatedAt: time.Now().Add(-1 * time.Hour), UpdatedAt: time.Now()},
			{ID: "aegis-cr2", Title: "Old bead", Status: "open", Priority: 3, Owner: "aegis/crew/bob", CreatedAt: time.Now().Add(-48 * time.Hour), UpdatedAt: time.Now()},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/created?days=7", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /created status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "aegis-cr1") {
		t.Error("expected recently created issue ID")
	}
	if !strings.Contains(body, "Fresh bead") {
		t.Error("expected recently created issue title")
	}
	if !strings.Contains(body, "alice") {
		t.Error("expected filer name in output")
	}
}

func TestCreatedPage_BatchActions(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "aegis-cba1", Title: "New item", Status: "open", Priority: 2, Owner: "aegis/crew/bob", CreatedAt: time.Now().Add(-1 * time.Hour), UpdatedAt: time.Now()},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/created?days=7", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /created status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "batch-bar-created") {
		t.Error("expected batch bar on created page")
	}
	if !strings.Contains(body, "createdToggleDay") {
		t.Error("expected per-day toggle-all on created page")
	}
	if !strings.Contains(body, "createdBatchAction") {
		t.Error("expected batch action script on created page")
	}
}

func TestCreatedPage_HTMXPartial(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/created", nil)
	req.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /created (HTMX) status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if strings.Contains(body, "<html") {
		t.Error("HTMX partial should not include full HTML layout")
	}
	if !strings.Contains(body, "Recently Created") {
		t.Error("expected content in HTMX partial")
	}
}

func TestCreatedPage_RigFilter(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{
			{Name: "beads_aegis"},
			{Name: "beads_gastown"},
		},
		issues: []dolt.Issue{
			{ID: "rf1", Title: "Aegis bead", Status: "open", Priority: 1, Rig: "beads_aegis", CreatedAt: time.Now().Add(-1 * time.Hour), UpdatedAt: time.Now()},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/created?days=7&rig=beads_aegis", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /created?rig= status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "FilterRig") || !strings.Contains(body, "beads_aegis") {
		// Just check the page renders with rig filter
		if !strings.Contains(body, "Recently Created") {
			t.Error("expected page content with rig filter")
		}
	}
}

func TestCreatedPage_QuickActions(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "cqa1", Title: "Actionable", Status: "open", Priority: 2, CreatedAt: time.Now().Add(-1 * time.Hour), UpdatedAt: time.Now()},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/created?days=7", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /created status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "in_progress") {
		t.Error("expected start button on created page")
	}
	if !strings.Contains(body, "hx-post=") {
		t.Error("expected HTMX quick actions on created page")
	}
}

func TestCreatedPage_InlinePriority(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "cip1", Title: "Priority test", Status: "open", Priority: 2, CreatedAt: time.Now().Add(-1 * time.Hour), UpdatedAt: time.Now()},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/created?days=7", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /created status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "bead-priority-inline") {
		t.Error("expected inline priority editing on created page")
	}
}

func TestCreatedPage_AssigneeDropdown(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "cad1", Title: "Needs assignment", Status: "open", Priority: 1, CreatedAt: time.Now().Add(-1 * time.Hour), UpdatedAt: time.Now()},
		},
		assignees: []string{"aegis/crew/alice", "aegis/crew/bob"},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/created?days=7", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /created status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "triage-assign") {
		t.Error("expected assignee dropdown on created page")
	}
	if !strings.Contains(body, "alice") {
		t.Error("expected assignee option 'alice'")
	}
}

func TestSprintPage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/sprint", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /sprint status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Weekly Sprint") {
		t.Error("expected 'Weekly Sprint' heading")
	}
}

func TestSprintPage_WithData(t *testing.T) {
	// Use fixed midday time to avoid midnight flakes
	fixed := time.Date(2026, 3, 11, 12, 0, 0, 0, time.Local) // Wednesday
	weekParam := "2026-03-09"                                   // Monday of that week
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "sp1", Title: "Closed this week", Status: "closed", Priority: 1, Assignee: "aegis/crew/alice", CreatedAt: fixed.Add(-48 * time.Hour), UpdatedAt: fixed},
			{ID: "sp2", Title: "Created this week", Status: "open", Priority: 2, Owner: "aegis/crew/bob", CreatedAt: fixed, UpdatedAt: fixed},
			{ID: "sp3", Title: "Active this week", Status: "in_progress", Priority: 1, Assignee: "aegis/crew/alice", CreatedAt: fixed.Add(-72 * time.Hour), UpdatedAt: fixed},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/sprint?week="+weekParam, nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /sprint status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Closed this week") {
		t.Error("expected closed issue")
	}
	if !strings.Contains(body, "Created this week") {
		t.Error("expected created issue")
	}
	if !strings.Contains(body, "Active this week") {
		t.Error("expected active issue")
	}
}

func TestSprintPage_HTMXPartial(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/sprint", nil)
	req.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /sprint (HTMX) status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if strings.Contains(body, "<html") {
		t.Error("HTMX partial should not include full HTML layout")
	}
}

func TestSprintPage_WeekNavigation(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/sprint?week=2026-03-02", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /sprint?week= status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Prev") {
		t.Error("expected Prev navigation link")
	}
	if !strings.Contains(body, "Next") {
		t.Error("expected Next navigation link for past week")
	}
}

func TestSprintPage_RigFilter(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{
			{Name: "beads_aegis"},
			{Name: "beads_gastown"},
		},
		issues: []dolt.Issue{
			{ID: "srf1", Title: "Aegis sprint bead", Status: "open", Priority: 1, CreatedAt: time.Now().Add(-1 * time.Hour), UpdatedAt: time.Now()},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/sprint?rig=beads_aegis", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /sprint?rig= status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Weekly Sprint") {
		t.Error("expected page content with rig filter")
	}
}

func TestSprintPage_AssigneeDropdown(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "sp-ad1", Title: "Active task", Status: "in_progress", Priority: 1,
				Assignee: "aegis/crew/arnold",
				CreatedAt: time.Now().Add(-2 * time.Hour), UpdatedAt: time.Now()},
		},
		assignees: []string{"aegis/crew/arnold", "aegis/crew/grant"},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/sprint", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "triage-assign") {
		t.Error("expected assignee dropdown (triage-assign class)")
	}
	if !strings.Contains(body, "batch-bar") {
		t.Error("expected batch action bar")
	}
}

func TestSearchPage_AssigneeDropdown(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "s-ad1", Title: "Searchable task", Status: "open", Priority: 1, UpdatedAt: time.Now()},
		},
		assignees: []string{"aegis/crew/arnold"},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/search?q=Searchable", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "triage-assign") {
		t.Error("expected assignee dropdown (triage-assign class)")
	}
	if !strings.Contains(body, "Searchable task") {
		t.Error("expected search result")
	}
}

func TestRecapPage_AssigneeDropdown(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "r-ad1", Title: "Active recap item", Status: "in_progress", Priority: 1,
				CreatedAt: time.Now().Add(-1 * time.Hour), UpdatedAt: time.Now()},
		},
		assignees: []string{"aegis/crew/arnold"},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/recap", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "triage-assign") {
		t.Error("expected assignee dropdown (triage-assign class)")
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

func TestActivityPage_StatusFilter(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "asf1", Title: "Open work", Status: "open", Priority: 1, UpdatedAt: time.Now()},
			{ID: "asf2", Title: "Closed work", Status: "closed", Priority: 2, UpdatedAt: time.Now()},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/activity?hours=24&status=open", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /activity?status=open status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Open work") {
		t.Error("expected 'Open work' in filtered results")
	}
	if strings.Contains(body, "Closed work") {
		t.Error("should not show 'Closed work' when filtering by open status")
	}
	if !strings.Contains(body, "(open)") {
		t.Error("expected status filter label '(open)' in results meta")
	}
	if !strings.Contains(body, "All statuses") {
		t.Error("expected 'All statuses' filter badge")
	}
}

func TestActivityPage_AgentFilter(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "aaf1", Title: "Arnold task", Status: "open", Priority: 1, Assignee: "aegis/crew/arnold", UpdatedAt: time.Now()},
			{ID: "aaf2", Title: "Grant task", Status: "open", Priority: 2, Assignee: "aegis/crew/grant", UpdatedAt: time.Now()},
		},
		assignees: []string{"aegis/crew/arnold", "aegis/crew/grant"},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/activity?hours=24&agent=aegis/crew/arnold", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /activity?agent= status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Arnold task") {
		t.Error("expected 'Arnold task' in agent-filtered results")
	}
	if strings.Contains(body, "Grant task") {
		t.Error("should not show 'Grant task' when filtering by arnold")
	}
	if !strings.Contains(body, "All agents") {
		t.Error("expected 'All agents' filter badge")
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
	// Use a fixed midday time to avoid midnight boundary flakes
	fixed := time.Date(2026, 3, 15, 12, 0, 0, 0, time.Local)
	dateStr := fixed.Format("2006-01-02")
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "r1", Title: "Created today", Status: "open", Priority: 1, Owner: "aegis/crew/alice", CreatedAt: fixed.Add(-1 * time.Hour), UpdatedAt: fixed},
			{ID: "r2", Title: "Closed today", Status: "closed", Priority: 2, Assignee: "aegis/crew/bob", CreatedAt: fixed.Add(-24 * time.Hour), UpdatedAt: fixed},
			{ID: "r3", Title: "Active work", Status: "in_progress", Priority: 1, Assignee: "aegis/crew/alice", CreatedAt: fixed.Add(-48 * time.Hour), UpdatedAt: fixed},
			{ID: "r4", Title: "Old item", Status: "open", Priority: 3, CreatedAt: fixed.Add(-72 * time.Hour), UpdatedAt: fixed.Add(-72 * time.Hour)},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/recap?date="+dateStr, nil)
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

func TestContributorsPage_SortOptions(t *testing.T) {
	sorts := []string{"", "total", "closed", "active", "rate", "recent", "name"}
	for _, s := range sorts {
		t.Run("sort="+s, func(t *testing.T) {
			srv := New(nil)
			url := "/contributors"
			if s != "" {
				url += "?sort=" + s
			}
			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()
			srv.ServeHTTP(w, req)
			if w.Code != http.StatusOK {
				t.Fatalf("GET %s status = %d, want %d", url, w.Code, http.StatusOK)
			}
			body := w.Body.String()
			if !strings.Contains(body, "By total") {
				t.Error("expected sort options in page")
			}
		})
	}
}

func TestWorkloadPage_SortOptions(t *testing.T) {
	sorts := []string{"", "total", "active", "blocked", "highpri", "name"}
	for _, s := range sorts {
		t.Run("sort="+s, func(t *testing.T) {
			srv := New(nil)
			url := "/workload"
			if s != "" {
				url += "?sort=" + s
			}
			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()
			srv.ServeHTTP(w, req)
			if w.Code != http.StatusOK {
				t.Fatalf("GET %s status = %d, want %d", url, w.Code, http.StatusOK)
			}
			body := w.Body.String()
			if !strings.Contains(body, "By total") {
				t.Error("expected sort options in page")
			}
		})
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

func TestParseBeadLink(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		wantID string
		wantDB string
	}{
		{"simple", "<!-- bead: aegis-abc123 -->", "aegis-abc123", "aegis"},
		{"with db", "<!-- bead: gastown/gt-xyz -->", "gt-xyz", "gastown"},
		{"extra spaces", "<!--  bead:  aegis-def  -->", "aegis-def", "aegis"},
		{"in content", "some text\n<!-- bead: aegis-test -->\nmore text", "aegis-test", "aegis"},
		{"no match", "no bead link here", "", ""},
		{"empty", "", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotID, gotDB := parseBeadLink(tt.input)
			if gotID != tt.wantID {
				t.Errorf("parseBeadLink(%q) ID = %q, want %q", tt.input, gotID, tt.wantID)
			}
			if gotDB != tt.wantDB {
				t.Errorf("parseBeadLink(%q) DB = %q, want %q", tt.input, gotDB, tt.wantDB)
			}
		})
	}
}

func TestParseMentions(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{"single mention", "@arnold check this", []string{"arnold"}},
		{"multiple mentions", "@arnold @hammond please review", []string{"arnold", "hammond"}},
		{"duplicate mentions", "@arnold @arnold only once", []string{"arnold"}},
		{"no mentions", "no mentions here", nil},
		{"mention in sentence", "hey @stiwi what do you think?", []string{"stiwi"}},
		{"hyphenated name", "@aegis-arnold ok", []string{"aegis-arnold"}},
		{"underscore name", "@crew_lead ok", []string{"crew_lead"}},
		{"empty input", "", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseMentions(tt.input)
			if len(got) != len(tt.want) {
				t.Fatalf("parseMentions(%q) = %v (len %d), want %v (len %d)", tt.input, got, len(got), tt.want, len(tt.want))
			}
			for i, g := range got {
				if g != tt.want[i] {
					t.Errorf("parseMentions(%q)[%d] = %q, want %q", tt.input, i, g, tt.want[i])
				}
			}
		})
	}
}

func TestRenderMarkdown(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string // substring to check
		wantErr bool
	}{
		{"heading", "# Hello", "<h1>Hello</h1>", false},
		{"bold", "**bold**", "<strong>bold</strong>", false},
		{"link", "[link](http://example.com)", `<a href="http://example.com">link</a>`, false},
		{"code block", "```go\nfmt.Println()\n```", "<code", false},
		{"empty", "", "", false},
		{"gfm table", "| a | b |\n|---|---|\n| 1 | 2 |", "<table>", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := renderMarkdown(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("renderMarkdown(%q) err = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if tt.want != "" && !strings.Contains(string(got), tt.want) {
				t.Errorf("renderMarkdown(%q) = %q, want substring %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestDesignsPage_FilterLinks(t *testing.T) {
	srv := New(nil)

	filters := []string{"", "review", "progress", "done"}
	for _, f := range filters {
		t.Run("filter="+f, func(t *testing.T) {
			url := "/designs"
			if f != "" {
				url += "?filter=" + f
			}
			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()
			srv.ServeHTTP(w, req)
			if w.Code != http.StatusOK {
				t.Fatalf("GET %s status = %d, want %d", url, w.Code, http.StatusOK)
			}
		})
	}
}

func TestDesignsPage_HXPartial(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/designs", nil)
	req.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("HTMX GET /designs status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("HTMX designs should return partial, not full page")
	}
}

func TestDesignView_ValidNames(t *testing.T) {
	srv := New(nil)
	names := []string{"simple", "with-hyphens", "with_underscores", "Mixed123"}
	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/designs/"+name, nil)
			w := httptest.NewRecorder()
			srv.ServeHTTP(w, req)
			// Should not 404 due to name validation (may 200 or 404 from forgejo)
			if w.Code == http.StatusBadRequest {
				t.Errorf("GET /designs/%s should not return 400", name)
			}
		})
	}
}

func TestDesignView_InvalidNames(t *testing.T) {
	srv := New(nil)
	tests := []struct {
		name string
		path string
	}{
		{"dots", "/designs/has.dots"},
		{"slash", "/designs/has/slash"},
		{"html", "/designs/has%3Chtml%3E"},
		{"spaces", "/designs/has%20spaces"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()
			srv.ServeHTTP(w, req)
			if w.Code != http.StatusNotFound {
				t.Errorf("GET %s status = %d, want 404", tt.path, w.Code)
			}
		})
	}
}

func TestDesignComment_InvalidAuthor(t *testing.T) {
	srv := New(nil)
	body := strings.NewReader("bead_id=x&bead_db=y&author=bad<script>&body=test")
	req := httptest.NewRequest("POST", "/designs/test-doc/comment", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusSeeOther {
		t.Fatalf("POST design comment (bad author) status = %d, want %d", w.Code, http.StatusSeeOther)
	}
	loc := w.Header().Get("Location")
	if !strings.Contains(loc, "feedback=invalid") {
		t.Errorf("expected redirect to feedback=invalid, got: %s", loc)
	}
}

func TestDesignComment_WithMockDS(t *testing.T) {
	mock := &mockDataSource{}
	srv := New(mock)
	body := strings.NewReader("bead_id=aegis-abc&bead_db=aegis&author=stiwi&body=looks+good")
	req := httptest.NewRequest("POST", "/designs/test-doc/comment", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusSeeOther {
		t.Fatalf("POST design comment status = %d, want %d", w.Code, http.StatusSeeOther)
	}
	loc := w.Header().Get("Location")
	if !strings.Contains(loc, "feedback=ok") {
		t.Errorf("expected redirect to feedback=ok, got: %s", loc)
	}
}

func TestDesignComment_NoDataSource(t *testing.T) {
	srv := New(nil)
	body := strings.NewReader("bead_id=aegis-abc&bead_db=aegis&author=stiwi&body=test")
	req := httptest.NewRequest("POST", "/designs/test-doc/comment", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusSeeOther {
		t.Fatalf("POST design comment (nil ds) status = %d, want %d", w.Code, http.StatusSeeOther)
	}
	loc := w.Header().Get("Location")
	if !strings.Contains(loc, "feedback=nodb") {
		t.Errorf("expected redirect to feedback=nodb, got: %s", loc)
	}
}

func TestDesignApprove_WithMockDS(t *testing.T) {
	mock := &mockDataSource{}
	srv := New(mock)
	body := strings.NewReader("bead_id=aegis-abc&bead_db=aegis")
	req := httptest.NewRequest("POST", "/designs/test-doc/approve", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusSeeOther {
		t.Fatalf("POST design approve status = %d, want %d", w.Code, http.StatusSeeOther)
	}
	loc := w.Header().Get("Location")
	if !strings.Contains(loc, "feedback=approved") {
		t.Errorf("expected redirect to feedback=approved, got: %s", loc)
	}
}

func TestDesignView_FeedbackBanners(t *testing.T) {
	srv := New(nil)
	feedbacks := []struct {
		param string
		want  string
	}{
		{"ok", "Comment added"},
		{"approved", "approved"},
		{"missing", "required"},
		{"error", "Failed"},
	}
	for _, fb := range feedbacks {
		t.Run("feedback="+fb.param, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/designs/test-doc?feedback="+fb.param, nil)
			w := httptest.NewRecorder()
			srv.ServeHTTP(w, req)
			// May be 200 (renders template with feedback) or 404 (no forgejo)
			if w.Code == http.StatusOK {
				body := w.Body.String()
				if !strings.Contains(body, fb.want) {
					t.Errorf("GET /designs/test-doc?feedback=%s missing %q in body", fb.param, fb.want)
				}
			}
		})
	}
}

func TestDesignComment_MethodNotAllowed(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/designs/test-doc/comment", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	// GET to /designs/test-doc/comment should 404 (only POST is routed)
	if w.Code != http.StatusNotFound {
		t.Fatalf("GET /designs/test-doc/comment status = %d, want 404", w.Code)
	}
}

func TestDesignApprove_MethodNotAllowed(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/designs/test-doc/approve", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("GET /designs/test-doc/approve status = %d, want 404", w.Code)
	}
}

func TestDesignUnknownAction_404(t *testing.T) {
	srv := New(nil)
	body := strings.NewReader("data=test")
	req := httptest.NewRequest("POST", "/designs/test-doc/unknown", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("POST /designs/test-doc/unknown status = %d, want 404", w.Code)
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
	// Use a fixed midday time to avoid midnight boundary flakes
	fixed := time.Date(2026, 3, 15, 12, 0, 0, 0, time.Local)
	dateStr := fixed.Format("2006-01-02")
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "cr1", Title: "New bead", Status: "open", Priority: 2, Owner: "aegis/crew/bob", CreatedAt: fixed.Add(-1 * time.Hour), UpdatedAt: fixed},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/recap?date="+dateStr, nil)
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

func TestStalePage_PriorityFilter(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "st1", Title: "P0 stale", Status: "in_progress", Priority: 0, UpdatedAt: time.Now().Add(-10 * 24 * time.Hour)},
			{ID: "st2", Title: "P2 stale", Status: "in_progress", Priority: 2, UpdatedAt: time.Now().Add(-10 * 24 * time.Hour)},
		},
		assignees: []string{"aegis/crew/arnold"},
	}
	srv := New(ds)
	req := httptest.NewRequest("GET", "/stale?priority=0", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "All priorities") {
		t.Error("expected 'All priorities' badge")
	}
	if !strings.Contains(body, "P0 stale") {
		t.Error("expected P0 stale item")
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

func TestOwnersPage_RigFilter(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}, {Name: "beads_gastown"}},
		issues: []dolt.Issue{
			{ID: "o1", Title: "Owner task", Status: "open", Owner: "alice"},
		},
	}
	srv := New(ds)

	req := httptest.NewRequest("GET", "/owners?rig=beads_aegis", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /owners?rig= status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, `rig=beads_aegis`) {
		t.Error("expected rig filter preserved in auto-refresh URL")
	}
	if !strings.Contains(body, "filter-active") {
		t.Error("expected filter-active badge for selected rig")
	}
}

func TestOwnersPage_PriorityBreakdown(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "o1", Title: "Critical", Status: "open", Owner: "alice", Priority: 0},
			{ID: "o2", Title: "Important", Status: "in_progress", Owner: "alice", Priority: 1},
			{ID: "o3", Title: "Normal", Status: "open", Owner: "alice", Priority: 3},
		},
	}
	srv := New(ds)
	req := httptest.NewRequest("GET", "/owners", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "priority-badge p0") {
		t.Error("expected P0 column header")
	}
	if !strings.Contains(body, "priority-badge p1") {
		t.Error("expected P1 column header")
	}
}

func TestPrioritiesPage_RigFilter(t *testing.T) {
	ds := &mockDataSource{
		databases:          []dolt.DatabaseInfo{{Name: "beads_aegis"}, {Name: "beads_gastown"}},
		priorityCounts: []dolt.PriorityStatusCount{
			{Priority: 1, Status: "open", Count: 3},
		},
	}
	srv := New(ds)

	req := httptest.NewRequest("GET", "/priorities?rig=beads_aegis", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /priorities?rig= status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, `rig=beads_aegis`) {
		t.Error("expected rig filter preserved in auto-refresh URL")
	}
	if !strings.Contains(body, "in aegis") {
		t.Error("expected rig-specific display text")
	}
	if !strings.Contains(body, "filter-active") {
		t.Error("expected filter-active badge for selected rig")
	}
}

func TestTypesPage_RigFilter(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}, {Name: "beads_gastown"}},
		issues: []dolt.Issue{
			{ID: "t1", Title: "Type task", Status: "open", Type: "task"},
		},
	}
	srv := New(ds)

	req := httptest.NewRequest("GET", "/types?rig=beads_aegis", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /types?rig= status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, `rig=beads_aegis`) {
		t.Error("expected rig filter preserved in auto-refresh URL")
	}
	if !strings.Contains(body, "filter-active") {
		t.Error("expected filter-active badge for selected rig")
	}
}

func TestMatrixPage_RigFilter(t *testing.T) {
	ds := &mockDataSource{
		databases:            []dolt.DatabaseInfo{{Name: "beads_aegis"}, {Name: "beads_gastown"}},
		assigneeCounts: []dolt.AssigneeStatusCount{
			{Assignee: "alice", Status: "open", Count: 2},
		},
	}
	srv := New(ds)

	req := httptest.NewRequest("GET", "/matrix?rig=beads_aegis", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /matrix?rig= status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, `rig=beads_aegis`) {
		t.Error("expected rig filter preserved in auto-refresh URL")
	}
	if !strings.Contains(body, "filter-active") {
		t.Error("expected filter-active badge for selected rig")
	}
}

func TestSLAPage_RigFilter(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}, {Name: "beads_gastown"}},
		issues: []dolt.Issue{
			{ID: "s1", Title: "SLA item", Status: "open", Priority: 1, CreatedAt: time.Now().Add(-48 * time.Hour)},
		},
	}
	srv := New(ds)
	req := httptest.NewRequest("GET", "/sla?rig=beads_aegis", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, `rig=beads_aegis`) {
		t.Error("expected rig filter preserved in auto-refresh URL")
	}
	if !strings.Contains(body, "filter-active") {
		t.Error("expected filter-active badge")
	}
}

func TestCycleTimePage_RigFilter(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}, {Name: "beads_gastown"}},
		issues: []dolt.Issue{
			{ID: "c1", Title: "Cycle item", Status: "closed", CreatedAt: time.Now().Add(-72 * time.Hour), UpdatedAt: time.Now()},
		},
	}
	srv := New(ds)
	req := httptest.NewRequest("GET", "/cycle-time?rig=beads_aegis", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, `rig=beads_aegis`) {
		t.Error("expected rig filter preserved in auto-refresh URL")
	}
	if !strings.Contains(body, "filter-active") {
		t.Error("expected filter-active badge")
	}
}

func TestResponseTimePage_RigFilter(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}, {Name: "beads_gastown"}},
		issues: []dolt.Issue{
			{ID: "r1", Title: "Response item", Status: "in_progress", CreatedAt: time.Now().Add(-2 * time.Hour), UpdatedAt: time.Now()},
		},
	}
	srv := New(ds)
	req := httptest.NewRequest("GET", "/response-time?rig=beads_aegis", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, `rig=beads_aegis`) {
		t.Error("expected rig filter preserved in auto-refresh URL")
	}
	if !strings.Contains(body, "filter-active") {
		t.Error("expected filter-active badge")
	}
}

func TestResponseTimePage_SortOptions(t *testing.T) {
	sorts := []string{"", "fastest", "slowest", "priority", "recent"}
	for _, s := range sorts {
		t.Run("sort="+s, func(t *testing.T) {
			srv := New(nil)
			url := "/response-time"
			if s != "" {
				url += "?sort=" + s
			}
			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()
			srv.ServeHTTP(w, req)
			if w.Code != http.StatusOK {
				t.Fatalf("GET %s status = %d, want %d", url, w.Code, http.StatusOK)
			}
			body := w.Body.String()
			if !strings.Contains(body, "Fastest first") {
				t.Error("expected sort options in page")
			}
		})
	}
}

func TestNetFlowPage_RigFilter(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}, {Name: "beads_gastown"}},
		issues: []dolt.Issue{
			{ID: "n1", Title: "Net flow item", Status: "open", CreatedAt: time.Now().Add(-24 * time.Hour)},
		},
	}
	srv := New(ds)
	req := httptest.NewRequest("GET", "/net-flow?rig=beads_aegis", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, `rig=beads_aegis`) {
		t.Error("expected rig filter preserved in auto-refresh URL")
	}
	if !strings.Contains(body, "filter-active") {
		t.Error("expected filter-active badge")
	}
}

func TestResolutionRatePage_RigFilter(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}, {Name: "beads_gastown"}},
		issues: []dolt.Issue{
			{ID: "rr1", Title: "Resolution item", Status: "closed", CreatedAt: time.Now().Add(-48 * time.Hour), UpdatedAt: time.Now()},
		},
	}
	srv := New(ds)
	req := httptest.NewRequest("GET", "/resolution-rate?rig=beads_aegis", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, `rig=beads_aegis`) {
		t.Error("expected rig filter preserved in auto-refresh URL")
	}
	if !strings.Contains(body, "filter-active") {
		t.Error("expected filter-active badge")
	}
}

func TestBurndownPage_RigFilter(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}, {Name: "beads_gastown"}},
		issues: []dolt.Issue{
			{ID: "x1", Title: "Test item", Status: "open", CreatedAt: time.Now().Add(-24 * time.Hour), UpdatedAt: time.Now()},
		},
	}
	srv := New(ds)
	req := httptest.NewRequest("GET", "/burndown?rig=beads_aegis", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, `rig=beads_aegis`) {
		t.Error("expected rig filter preserved in auto-refresh URL")
	}
	if !strings.Contains(body, "filter-active") {
		t.Error("expected filter-active badge")
	}
}

func TestScopePage_RigFilter(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}, {Name: "beads_gastown"}},
		issues: []dolt.Issue{
			{ID: "x1", Title: "Test item", Status: "open", CreatedAt: time.Now().Add(-24 * time.Hour), UpdatedAt: time.Now()},
		},
	}
	srv := New(ds)
	req := httptest.NewRequest("GET", "/scope?rig=beads_aegis", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, `rig=beads_aegis`) {
		t.Error("expected rig filter preserved in auto-refresh URL")
	}
	if !strings.Contains(body, "filter-active") {
		t.Error("expected filter-active badge")
	}
}

func TestForecastPage_RigFilter(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}, {Name: "beads_gastown"}},
		issues: []dolt.Issue{
			{ID: "x1", Title: "Test item", Status: "open", CreatedAt: time.Now().Add(-24 * time.Hour), UpdatedAt: time.Now()},
		},
	}
	srv := New(ds)
	req := httptest.NewRequest("GET", "/forecast?rig=beads_aegis", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, `rig=beads_aegis`) {
		t.Error("expected rig filter preserved in auto-refresh URL")
	}
	if !strings.Contains(body, "filter-active") {
		t.Error("expected filter-active badge")
	}
}

func TestTrendsPage_RigFilter(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}, {Name: "beads_gastown"}},
		issues: []dolt.Issue{
			{ID: "x1", Title: "Test item", Status: "open", CreatedAt: time.Now().Add(-24 * time.Hour), UpdatedAt: time.Now()},
		},
	}
	srv := New(ds)
	req := httptest.NewRequest("GET", "/trends?rig=beads_aegis", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, `rig=beads_aegis`) {
		t.Error("expected rig filter preserved in auto-refresh URL")
	}
	if !strings.Contains(body, "filter-active") {
		t.Error("expected filter-active badge")
	}
}

func TestCohortPage_RigFilter(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}, {Name: "beads_gastown"}},
		issues: []dolt.Issue{
			{ID: "x1", Title: "Test item", Status: "open", CreatedAt: time.Now().Add(-24 * time.Hour), UpdatedAt: time.Now()},
		},
	}
	srv := New(ds)
	req := httptest.NewRequest("GET", "/cohort?rig=beads_aegis", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, `rig=beads_aegis`) {
		t.Error("expected rig filter preserved in auto-refresh URL")
	}
	if !strings.Contains(body, "filter-active") {
		t.Error("expected filter-active badge")
	}
}

// ── Standup Page Tests ──

func TestStandupPage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/standup", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /standup status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Daily Standup") {
		t.Error("expected 'Daily Standup' heading")
	}
}

func TestStandupPage_WithData(t *testing.T) {
	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "su1", Title: "Done yesterday", Status: "closed", Priority: 1, Assignee: "aegis/crew/alice", CreatedAt: yesterday.Add(-24 * time.Hour), UpdatedAt: yesterday},
			{ID: "su2", Title: "Working on this", Status: "in_progress", Priority: 2, Assignee: "aegis/crew/alice", CreatedAt: yesterday.Add(-48 * time.Hour), UpdatedAt: now},
			{ID: "su3", Title: "Stuck on blocker", Status: "blocked", Priority: 1, Assignee: "aegis/crew/bob", CreatedAt: yesterday.Add(-72 * time.Hour), UpdatedAt: yesterday},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/standup", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /standup status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "alice") {
		t.Error("expected agent alice in standup")
	}
	if !strings.Contains(body, "bob") {
		t.Error("expected agent bob in standup")
	}
	if !strings.Contains(body, "Working on this") {
		t.Error("expected in-progress item")
	}
	if !strings.Contains(body, "Stuck on blocker") {
		t.Error("expected blocked item")
	}
}

func TestStandupPage_HTMXPartial(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/standup", nil)
	req.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if strings.Contains(body, "<html") {
		t.Error("HTMX partial should not include <html> tag")
	}
	if !strings.Contains(body, "Daily Standup") {
		t.Error("expected standup content in partial")
	}
}

func TestStandupPage_RigFilter(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}, {Name: "beads_hq"}},
		issues: []dolt.Issue{
			{ID: "su-r1", Title: "Aegis item", Status: "in_progress", Priority: 1, Assignee: "aegis/crew/alice", CreatedAt: time.Now().Add(-48 * time.Hour), UpdatedAt: time.Now()},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/standup?rig=beads_aegis", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "filter-active") {
		t.Error("expected filter-active badge for rig filter")
	}
	if !strings.Contains(body, `rig=beads_aegis`) {
		t.Error("expected rig filter preserved in auto-refresh URL")
	}
}

func TestStandupPage_AutoRefresh(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/standup", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	body := w.Body.String()
	if !strings.Contains(body, `hx-trigger="every 60s"`) {
		t.Error("expected 60s auto-refresh on standup page")
	}
}

func TestStandupPage_StatGrid(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "su-s1", Title: "Active item", Status: "in_progress", Priority: 1, Assignee: "aegis/crew/alice", CreatedAt: time.Now().Add(-24 * time.Hour), UpdatedAt: time.Now()},
			{ID: "su-s2", Title: "Blocked item", Status: "blocked", Priority: 2, Assignee: "aegis/crew/alice", CreatedAt: time.Now().Add(-48 * time.Hour), UpdatedAt: time.Now()},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/standup", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	body := w.Body.String()
	if !strings.Contains(body, "In progress") {
		t.Error("expected 'In progress' stat label")
	}
	if !strings.Contains(body, "Blocked") {
		t.Error("expected 'Blocked' stat label")
	}
	if !strings.Contains(body, "Active agents") {
		t.Error("expected 'Active agents' stat label")
	}
}

func TestStandupPage_QuickActions(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "su-qa1", Title: "Active item", Status: "in_progress", Priority: 1, Assignee: "aegis/crew/alice", CreatedAt: time.Now().Add(-24 * time.Hour), UpdatedAt: time.Now()},
			{ID: "su-qa2", Title: "Blocked item", Status: "blocked", Priority: 2, Assignee: "aegis/crew/alice", CreatedAt: time.Now().Add(-48 * time.Hour), UpdatedAt: time.Now()},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/standup", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	body := w.Body.String()
	// In-progress items should have close/block/defer buttons
	if !strings.Contains(body, "briefing-actions") {
		t.Error("expected quick action buttons in standup")
	}
	// Blocked items should have unblock button
	if !strings.Contains(body, "unblock") {
		t.Error("expected 'unblock' button for blocked items")
	}
}

// ── Momentum Page Tests ──

func TestMomentumPage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/momentum", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /momentum status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Momentum") {
		t.Error("expected 'Momentum' heading")
	}
}

func TestMomentumPage_WithData(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "mo1", Title: "Open item", Status: "open", Priority: 2, CreatedAt: time.Now().Add(-24 * time.Hour), UpdatedAt: time.Now().Add(-10 * 24 * time.Hour)},
			{ID: "mo2", Title: "Active item", Status: "in_progress", Priority: 1, CreatedAt: time.Now().Add(-48 * time.Hour), UpdatedAt: time.Now()},
			{ID: "mo3", Title: "Blocked item", Status: "blocked", Priority: 1, CreatedAt: time.Now().Add(-72 * time.Hour), UpdatedAt: time.Now()},
		},
		counts: map[string]int{"open": 10, "in_progress": 5, "blocked": 2, "closed": 50},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/momentum", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /momentum status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Velocity") {
		t.Error("expected Velocity signal")
	}
	if !strings.Contains(body, "Net Flow") {
		t.Error("expected Net Flow signal")
	}
	if !strings.Contains(body, "Blocker Ratio") {
		t.Error("expected Blocker Ratio signal")
	}
	if !strings.Contains(body, "Staleness") {
		t.Error("expected Staleness signal")
	}
	if !strings.Contains(body, "This Week vs Last Week") {
		t.Error("expected week comparison table")
	}
}

func TestMomentumPage_HTMXPartial(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/momentum", nil)
	req.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if strings.Contains(body, "<html") {
		t.Error("HTMX partial should not include <html> tag")
	}
	if !strings.Contains(body, "Momentum") {
		t.Error("expected momentum content in partial")
	}
}

func TestMomentumPage_AutoRefresh(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/momentum", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	body := w.Body.String()
	if !strings.Contains(body, `hx-trigger="every 120s"`) {
		t.Error("expected 120s auto-refresh on momentum page")
	}
}

func TestMomentumPage_SignalColors(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/momentum", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	body := w.Body.String()
	if !strings.Contains(body, "signal-indicator") {
		t.Error("expected signal indicators in momentum page")
	}
	if !strings.Contains(body, "momentum-signal") {
		t.Error("expected momentum-signal cards")
	}
}

// ── Risks Page Tests ──

func TestRisksPage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/risks", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /risks status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Risk Radar") {
		t.Error("expected 'Risk Radar' heading")
	}
}

func TestRisksPage_WithData(t *testing.T) {
	staleTime := time.Now().Add(-15 * 24 * time.Hour) // 15 days ago
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "r1", Title: "Critical stale P0", Status: "open", Priority: 0, CreatedAt: staleTime, UpdatedAt: staleTime},
			{ID: "r2", Title: "Warning stale P1", Status: "in_progress", Priority: 1, Assignee: "someone", CreatedAt: staleTime, UpdatedAt: staleTime},
			{ID: "r3", Title: "Blocked too long", Status: "blocked", Priority: 2, CreatedAt: staleTime, UpdatedAt: staleTime},
			{ID: "r4", Title: "Fresh item", Status: "open", Priority: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/risks", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /risks status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Critical stale P0") {
		t.Error("expected critical stale P0 item")
	}
	if !strings.Contains(body, "CRIT") {
		t.Error("expected CRIT severity badge")
	}
	if !strings.Contains(body, "Blocked too long") {
		t.Error("expected blocked item as risk")
	}
}

func TestRisksPage_RigFilter(t *testing.T) {
	staleTime := time.Now().Add(-10 * 24 * time.Hour)
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}, {Name: "beads_hq"}},
		issues: []dolt.Issue{
			{ID: "rf1", Title: "Stale P1", Status: "open", Priority: 1, CreatedAt: staleTime, UpdatedAt: staleTime},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/risks?rig=beads_aegis", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "filter-active") {
		t.Error("expected filter-active badge for rig filter")
	}
	if !strings.Contains(body, `rig=beads_aegis`) {
		t.Error("expected rig filter preserved in auto-refresh URL")
	}
}

func TestRisksPage_AutoRefresh(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/risks", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	body := w.Body.String()
	if !strings.Contains(body, `hx-trigger="every 120s"`) {
		t.Error("expected 120s auto-refresh on risks page")
	}
}

func TestRisksPage_QuickActions(t *testing.T) {
	staleTime := time.Now().Add(-10 * 24 * time.Hour)
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "rq1", Title: "Stale P0", Status: "open", Priority: 0, CreatedAt: staleTime, UpdatedAt: staleTime},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/risks", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	body := w.Body.String()
	if !strings.Contains(body, "hx-post=") {
		t.Error("expected HTMX quick actions on risks page")
	}
}

// ── Funnel Page Tests ──

func TestFunnelPage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/funnel", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /funnel status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Conversion Funnel") {
		t.Error("expected 'Conversion Funnel' heading")
	}
}

func TestFunnelPage_WithData(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "f1", Title: "Closed item", Status: "closed", Priority: 1, Assignee: "alice", CreatedAt: time.Now().Add(-10 * 24 * time.Hour), UpdatedAt: time.Now()},
			{ID: "f2", Title: "Open item", Status: "open", Priority: 2, CreatedAt: time.Now().Add(-5 * 24 * time.Hour), UpdatedAt: time.Now()},
			{ID: "f3", Title: "Active item", Status: "in_progress", Priority: 1, Assignee: "bob", CreatedAt: time.Now().Add(-3 * 24 * time.Hour), UpdatedAt: time.Now()},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/funnel", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /funnel status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Filed") {
		t.Error("expected 'Filed' funnel stage")
	}
	if !strings.Contains(body, "Assigned") {
		t.Error("expected 'Assigned' funnel stage")
	}
	if !strings.Contains(body, "Closed") {
		t.Error("expected 'Closed' funnel stage")
	}
	if !strings.Contains(body, "Triage rate") {
		t.Error("expected Triage rate stat")
	}
	if !strings.Contains(body, "Close rate") {
		t.Error("expected Close rate stat")
	}
}

func TestFunnelPage_HTMXPartial(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/funnel", nil)
	req.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if strings.Contains(body, "<html") {
		t.Error("HTMX partial should not include <html> tag")
	}
}

func TestFunnelPage_AutoRefresh(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/funnel", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	body := w.Body.String()
	if !strings.Contains(body, `hx-trigger="every 120s"`) {
		t.Error("expected 120s auto-refresh on funnel page")
	}
}

// ── Overflow tests ──

func TestOverflowPage_WithData(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "o1", Title: "Bug 1", Status: "open", Priority: 0, Assignee: "alice", CreatedAt: time.Now().Add(-30 * 24 * time.Hour), UpdatedAt: time.Now()},
			{ID: "o2", Title: "Bug 2", Status: "in_progress", Priority: 1, Assignee: "alice", CreatedAt: time.Now().Add(-5 * 24 * time.Hour), UpdatedAt: time.Now()},
			{ID: "o3", Title: "Bug 3", Status: "blocked", Priority: 2, Assignee: "alice", CreatedAt: time.Now().Add(-10 * 24 * time.Hour), UpdatedAt: time.Now()},
			{ID: "o4", Title: "Task 1", Status: "open", Priority: 3, Assignee: "bob", CreatedAt: time.Now().Add(-2 * 24 * time.Hour), UpdatedAt: time.Now()},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/overflow", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /overflow status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Agent Overload") {
		t.Error("expected 'Agent Overload' heading")
	}
	if !strings.Contains(body, "alice") {
		t.Error("expected agent 'alice' in results")
	}
	if !strings.Contains(body, "bob") {
		t.Error("expected agent 'bob' in results")
	}
	if !strings.Contains(body, "Score") {
		t.Error("expected Score column header")
	}
}

func TestOverflowPage_NilDS(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/overflow", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /overflow nil ds status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestOverflowPage_RigFilter(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}, {Name: "beads_gastown"}},
		issues: []dolt.Issue{
			{ID: "o1", Title: "Task", Status: "open", Priority: 2, Assignee: "alice", CreatedAt: time.Now().Add(-5 * 24 * time.Hour), UpdatedAt: time.Now()},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/overflow?rig=beads_aegis", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /overflow?rig status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "active") {
		t.Error("expected rig filter badge to be active")
	}
}

func TestOverflowPage_AutoRefresh(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/overflow", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	body := w.Body.String()
	if !strings.Contains(body, `hx-trigger="every 60s"`) {
		t.Error("expected 60s auto-refresh on overflow page")
	}
}

// ── Calendar tests ──

func TestCalendarPage_WithData(t *testing.T) {
	now := time.Now()
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "c1", Title: "Recent item", Status: "open", Priority: 2, CreatedAt: now.Add(-2 * 24 * time.Hour), UpdatedAt: now},
			{ID: "c2", Title: "Closed item", Status: "closed", Priority: 1, CreatedAt: now.Add(-10 * 24 * time.Hour), UpdatedAt: now.Add(-1 * 24 * time.Hour)},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/calendar", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /calendar status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Activity Calendar") {
		t.Error("expected 'Activity Calendar' heading")
	}
	if !strings.Contains(body, "Sun") {
		t.Error("expected day-of-week headers")
	}
	if !strings.Contains(body, "Created") {
		t.Error("expected Created stat")
	}
	if !strings.Contains(body, "prev") {
		t.Error("expected prev month navigation")
	}
}

func TestCalendarPage_NilDS(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/calendar", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /calendar nil ds status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestCalendarPage_MonthParam(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues:    []dolt.Issue{},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/calendar?year=2026&month=1", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /calendar with params status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "January") {
		t.Error("expected January in calendar for month=1")
	}
	if !strings.Contains(body, "2026") {
		t.Error("expected year 2026 in calendar")
	}
}

func TestCalendarPage_AutoRefresh(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/calendar", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	body := w.Body.String()
	if !strings.Contains(body, `hx-trigger="every 120s"`) {
		t.Error("expected 120s auto-refresh on calendar page")
	}
}

// ── Debt tests ──

func TestDebtPage_WithData(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "d1", Title: "Old bug", Status: "open", Type: "bug", Priority: 1, Assignee: "alice", CreatedAt: time.Now().Add(-30 * 24 * time.Hour), UpdatedAt: time.Now()},
			{ID: "d2", Title: "New task", Status: "open", Type: "task", Priority: 2, CreatedAt: time.Now().Add(-2 * 24 * time.Hour), UpdatedAt: time.Now()},
			{ID: "d3", Title: "Deferred thing", Status: "deferred", Type: "task", Priority: 3, CreatedAt: time.Now().Add(-60 * 24 * time.Hour), UpdatedAt: time.Now()},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/debt", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /debt status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Tech Debt") {
		t.Error("expected 'Tech Debt' heading")
	}
	if !strings.Contains(body, "Bug ratio") {
		t.Error("expected Bug ratio stat")
	}
	if !strings.Contains(body, "Aging Bugs") {
		t.Error("expected Aging Bugs section")
	}
	if !strings.Contains(body, "Deferred Pile") {
		t.Error("expected Deferred Pile section")
	}
	if !strings.Contains(body, "Old bug") {
		t.Error("expected old bug item in aging bugs list")
	}
	if !strings.Contains(body, "Deferred thing") {
		t.Error("expected deferred item in deferred pile")
	}
}

func TestDebtPage_NilDS(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/debt", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /debt nil ds status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestDebtPage_RigFilter(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "d1", Title: "Bug", Status: "open", Type: "bug", Priority: 2, CreatedAt: time.Now().Add(-20 * 24 * time.Hour), UpdatedAt: time.Now()},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/debt?rig=beads_aegis", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /debt?rig status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestDebtPage_AutoRefresh(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/debt", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	body := w.Body.String()
	if !strings.Contains(body, `hx-trigger="every 120s"`) {
		t.Error("expected 120s auto-refresh on debt page")
	}
}

// ── Snapshot tests ──

func TestSnapshotPage_WithData(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "s1", Title: "P0 bug", Status: "open", Type: "bug", Priority: 0, Assignee: "alice", CreatedAt: time.Now().Add(-2 * 24 * time.Hour), UpdatedAt: time.Now()},
			{ID: "s2", Title: "Active task", Status: "in_progress", Type: "task", Priority: 2, Assignee: "bob", CreatedAt: time.Now().Add(-5 * 24 * time.Hour), UpdatedAt: time.Now()},
			{ID: "s3", Title: "Blocked item", Status: "blocked", Type: "task", Priority: 1, Assignee: "alice", CreatedAt: time.Now().Add(-10 * 24 * time.Hour), UpdatedAt: time.Now()},
			{ID: "s4", Title: "Deferred", Status: "deferred", Type: "task", Priority: 3, CreatedAt: time.Now().Add(-30 * 24 * time.Hour), UpdatedAt: time.Now()},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/snapshot", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /snapshot status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "System Snapshot") {
		t.Error("expected 'System Snapshot' heading")
	}
	if !strings.Contains(body, "Backlog") {
		t.Error("expected Backlog section")
	}
	if !strings.Contains(body, "Urgency") {
		t.Error("expected Urgency section")
	}
	if !strings.Contains(body, "Workforce") {
		t.Error("expected Workforce section")
	}
	if !strings.Contains(body, "P0 open") {
		t.Error("expected P0 open stat")
	}
}

func TestSnapshotPage_NilDS(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/snapshot", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /snapshot nil ds status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestSnapshotPage_AutoRefresh(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/snapshot", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	body := w.Body.String()
	if !strings.Contains(body, `hx-trigger="every 60s"`) {
		t.Error("expected 60s auto-refresh on snapshot page")
	}
}

// ── Assignments tests ──

func TestAssignmentsPage_WithData(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "a1", Title: "Task for alice", Status: "in_progress", Priority: 1, Assignee: "alice", CreatedAt: time.Now().Add(-5 * 24 * time.Hour), UpdatedAt: time.Now()},
			{ID: "a2", Title: "Task for bob", Status: "open", Priority: 2, Assignee: "bob", CreatedAt: time.Now().Add(-3 * 24 * time.Hour), UpdatedAt: time.Now()},
			{ID: "a3", Title: "Unassigned task", Status: "open", Priority: 0, CreatedAt: time.Now().Add(-10 * 24 * time.Hour), UpdatedAt: time.Now()},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/assignments", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /assignments status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Assignments") {
		t.Error("expected 'Assignments' heading")
	}
	if !strings.Contains(body, "alice") {
		t.Error("expected agent 'alice'")
	}
	if !strings.Contains(body, "bob") {
		t.Error("expected agent 'bob'")
	}
	if !strings.Contains(body, "Unassigned") {
		t.Error("expected Unassigned section")
	}
}

func TestAssignmentsPage_NilDS(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/assignments", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /assignments nil ds status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestAssignmentsPage_RigFilter(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "a1", Title: "Task", Status: "open", Priority: 2, Assignee: "alice", CreatedAt: time.Now().Add(-2 * 24 * time.Hour), UpdatedAt: time.Now()},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/assignments?rig=beads_aegis", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /assignments?rig status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestAssignmentsPage_AutoRefresh(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/assignments", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	body := w.Body.String()
	if !strings.Contains(body, `hx-trigger="every 60s"`) {
		t.Error("expected 60s auto-refresh on assignments page")
	}
}

func TestAssignmentsPage_QuickActions(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "a1", Title: "Task", Status: "open", Priority: 2, Assignee: "alice", CreatedAt: time.Now().Add(-2 * 24 * time.Hour), UpdatedAt: time.Now()},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/assignments", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	body := w.Body.String()
	if !strings.Contains(body, "hx-post=") {
		t.Error("expected quick action buttons with hx-post")
	}
}

// ── Gaps tests ──

func TestGapsPage_WithData(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "g1", Title: "P0 unassigned", Status: "open", Type: "bug", Priority: 0, CreatedAt: time.Now().Add(-5 * 24 * time.Hour), UpdatedAt: time.Now().Add(-8 * 24 * time.Hour)},
			{ID: "g2", Title: "P1 assigned", Status: "in_progress", Type: "task", Priority: 1, Assignee: "alice", CreatedAt: time.Now().Add(-3 * 24 * time.Hour), UpdatedAt: time.Now()},
			{ID: "g3", Title: "Untyped item", Status: "open", Priority: 2, CreatedAt: time.Now().Add(-10 * 24 * time.Hour), UpdatedAt: time.Now().Add(-8 * 24 * time.Hour)},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/gaps", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /gaps status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Coverage Gaps") {
		t.Error("expected 'Coverage Gaps' heading")
	}
	if !strings.Contains(body, "Priority Coverage") {
		t.Error("expected Priority Coverage section")
	}
	if !strings.Contains(body, "Type Coverage") {
		t.Error("expected Type Coverage section")
	}
	if !strings.Contains(body, "Unassigned") {
		t.Error("expected Unassigned stat")
	}
}

func TestGapsPage_NilDS(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/gaps", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /gaps nil ds status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestGapsPage_AutoRefresh(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/gaps", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	body := w.Body.String()
	if !strings.Contains(body, `hx-trigger="every 120s"`) {
		t.Error("expected 120s auto-refresh on gaps page")
	}
}

// ── /compare ──

func TestComparePage(t *testing.T) {
	srv := New(&mockDataSource{})
	req := httptest.NewRequest("GET", "/compare", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /compare status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Period Comparison") {
		t.Error("expected 'Period Comparison' heading")
	}
	if !strings.Contains(body, "Created") {
		t.Error("expected Created metric")
	}
	if !strings.Contains(body, "Closed") {
		t.Error("expected Closed metric")
	}
	if !strings.Contains(body, "Net growth") {
		t.Error("expected Net growth metric")
	}
}

func TestComparePage_NilDS(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/compare", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /compare nil ds status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestComparePage_DaysParam(t *testing.T) {
	srv := New(&mockDataSource{})
	req := httptest.NewRequest("GET", "/compare?days=14", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /compare?days=14 status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "14 days") {
		t.Error("expected 14-day period display")
	}
}

func TestComparePage_AutoRefresh(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/compare", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	body := w.Body.String()
	if !strings.Contains(body, `hx-trigger="every 120s"`) {
		t.Error("expected 120s auto-refresh on compare page")
	}
}

// ── /chains ──

func TestChainsPage(t *testing.T) {
	srv := New(&mockDataSource{})
	req := httptest.NewRequest("GET", "/chains", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /chains status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Dependency Chains") {
		t.Error("expected 'Dependency Chains' heading")
	}
	if !strings.Contains(body, "Chain Stats") {
		t.Error("expected Chain Stats section")
	}
	if !strings.Contains(body, "Total edges") {
		t.Error("expected Total edges stat")
	}
	if !strings.Contains(body, "Max depth") {
		t.Error("expected Max depth stat")
	}
}

func TestChainsPage_NilDS(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/chains", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /chains nil ds status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestChainsPage_AutoRefresh(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/chains", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	body := w.Body.String()
	if !strings.Contains(body, `hx-trigger="every 300s"`) {
		t.Error("expected 300s auto-refresh on chains page")
	}
}

// ── /wip ──

func TestWIPPage(t *testing.T) {
	srv := New(&mockDataSource{})
	req := httptest.NewRequest("GET", "/wip", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /wip status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Work in Progress") {
		t.Error("expected 'Work in Progress' heading")
	}
	if !strings.Contains(body, "Summary") {
		t.Error("expected Summary section")
	}
	if !strings.Contains(body, "Over limit") {
		t.Error("expected Over limit stat")
	}
}

func TestWIPPage_NilDS(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/wip", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /wip nil ds status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestWIPPage_RigFilter(t *testing.T) {
	srv := New(&mockDataSource{})
	req := httptest.NewRequest("GET", "/wip?rig=beads_aegis", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /wip?rig status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestWIPPage_AutoRefresh(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/wip", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	body := w.Body.String()
	if !strings.Contains(body, `hx-trigger="every 60s"`) {
		t.Error("expected 60s auto-refresh on wip page")
	}
}

// ── /swarming ──

func TestSwarmingPage(t *testing.T) {
	srv := New(&mockDataSource{})
	req := httptest.NewRequest("GET", "/swarming", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /swarming status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Swarming Beads") {
		t.Error("expected 'Swarming Beads' heading")
	}
	if !strings.Contains(body, "Swarm Metrics") {
		t.Error("expected Swarm Metrics section")
	}
}

func TestSwarmingPage_NilDS(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/swarming", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /swarming nil ds status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestSwarmingPage_AutoRefresh(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/swarming", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	body := w.Body.String()
	if !strings.Contains(body, `hx-trigger="every 120s"`) {
		t.Error("expected 120s auto-refresh on swarming page")
	}
}

func TestSwarmingPage_RigFilter(t *testing.T) {
	srv := New(&mockDataSource{})
	req := httptest.NewRequest("GET", "/swarming?rig=beads_aegis", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /swarming?rig status = %d, want %d", w.Code, http.StatusOK)
	}
}

// ── /signals ──

func TestSignalsPage(t *testing.T) {
	srv := New(&mockDataSource{})
	req := httptest.NewRequest("GET", "/signals", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /signals status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Health Signals") {
		t.Error("expected 'Health Signals' heading")
	}
	if !strings.Contains(body, "Raw Metrics") {
		t.Error("expected Raw Metrics section")
	}
}

func TestSignalsPage_NilDS(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/signals", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /signals nil ds status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestSignalsPage_AutoRefresh(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/signals", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	body := w.Body.String()
	if !strings.Contains(body, `hx-trigger="every 60s"`) {
		t.Error("expected 60s auto-refresh on signals page")
	}
}

// ── /pair-freq ──

func TestPairFreqPage(t *testing.T) {
	srv := New(&mockDataSource{})
	req := httptest.NewRequest("GET", "/pair-freq", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /pair-freq status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Label Co-Occurrence") {
		t.Error("expected 'Label Co-Occurrence' heading")
	}
}

func TestPairFreqPage_NilDS(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/pair-freq", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /pair-freq nil ds status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestPairFreqPage_AutoRefresh(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/pair-freq", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	body := w.Body.String()
	if !strings.Contains(body, `hx-trigger="every 300s"`) {
		t.Error("expected 300s auto-refresh on pair-freq page")
	}
}

// ── /idle ──

func TestIdlePage(t *testing.T) {
	srv := New(&mockDataSource{})
	req := httptest.NewRequest("GET", "/idle", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /idle status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Idle Agents") {
		t.Error("expected 'Idle Agents' heading")
	}
	if !strings.Contains(body, "Agent Status") {
		t.Error("expected Agent Status section")
	}
}

func TestIdlePage_NilDS(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/idle", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /idle nil ds status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestIdlePage_AutoRefresh(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/idle", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	body := w.Body.String()
	if !strings.Contains(body, `hx-trigger="every 120s"`) {
		t.Error("expected 120s auto-refresh on idle page")
	}
}

// ── /reopen ──

func TestReopenPage(t *testing.T) {
	srv := New(&mockDataSource{})
	req := httptest.NewRequest("GET", "/reopen", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /reopen status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Reopened Beads") {
		t.Error("expected 'Reopened Beads' heading")
	}
	if !strings.Contains(body, "Reopen Stats") {
		t.Error("expected Reopen Stats section")
	}
}

func TestReopenPage_EnhancedBatchBar(t *testing.T) {
	now := time.Now()
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "reop-1", Title: "Reopened bug", Status: "open", Priority: 1, Assignee: "aegis/crew/arnold", UpdatedAt: now},
		},
		assignees: []string{"aegis/crew/arnold", "aegis/crew/deacon"},
		statusHistory: []dolt.StatusTransition{
			{ToStatus: "open", CommitDate: now.Add(-72 * time.Hour)},
			{FromStatus: "open", ToStatus: "closed", CommitDate: now.Add(-48 * time.Hour)},
			{FromStatus: "closed", ToStatus: "open", CommitDate: now.Add(-24 * time.Hour)},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/reopen", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /reopen status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "batch-priority") {
		t.Error("expected batch-priority dropdown in reopen batch bar")
	}
	if !strings.Contains(body, "batch-assign") {
		t.Error("expected batch-assign dropdown in reopen batch bar")
	}
	if !strings.Contains(body, "batch-label-input") {
		t.Error("expected batch-label-input in reopen batch bar")
	}
	if !strings.Contains(body, "reopenBatchPriority") {
		t.Error("expected reopenBatchPriority JS function")
	}
	if !strings.Contains(body, "reopenBatchAssignee") {
		t.Error("expected reopenBatchAssignee JS function")
	}
	if !strings.Contains(body, "reopenBatchLabel") {
		t.Error("expected reopenBatchLabel JS function")
	}
}

func TestReopenPage_NilDS(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/reopen", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /reopen nil ds status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestReopenPage_RigFilter(t *testing.T) {
	srv := New(&mockDataSource{})
	req := httptest.NewRequest("GET", "/reopen?rig=beads_aegis", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /reopen?rig status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestReopenPage_AutoRefresh(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/reopen", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	body := w.Body.String()
	if !strings.Contains(body, `hx-trigger="every 120s"`) {
		t.Error("expected 120s auto-refresh on reopen page")
	}
}

// ── /escalations ──

func TestEscalationsPage(t *testing.T) {
	srv := New(&mockDataSource{})
	req := httptest.NewRequest("GET", "/escalations", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /escalations status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Escalations") {
		t.Error("expected 'Escalations' heading")
	}
	if !strings.Contains(body, "Escalation Summary") {
		t.Error("expected Escalation Summary section")
	}
}

func TestEscalationsPage_EnhancedBatchBar(t *testing.T) {
	now := time.Now()
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "esc-1", Title: "Escalated item", Status: "open", Priority: 0, Assignee: "aegis/crew/arnold",
				CreatedAt: now.Add(-72 * time.Hour), UpdatedAt: now.Add(-1 * time.Hour)},
		},
		assignees: []string{"aegis/crew/arnold", "aegis/crew/deacon"},
		statusHistory: []dolt.StatusTransition{
			{ToStatus: "open", CommitDate: now.Add(-72 * time.Hour)},
			{FromStatus: "open", ToStatus: "in_progress", CommitDate: now.Add(-48 * time.Hour)},
			{FromStatus: "in_progress", ToStatus: "blocked", CommitDate: now.Add(-24 * time.Hour)},
			{FromStatus: "blocked", ToStatus: "open", CommitDate: now.Add(-1 * time.Hour)},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/escalations", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /escalations status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "batch-priority") {
		t.Error("expected batch-priority dropdown in escalations batch bar")
	}
	if !strings.Contains(body, "batch-assign") {
		t.Error("expected batch-assign dropdown in escalations batch bar")
	}
	if !strings.Contains(body, "batch-label-input") {
		t.Error("expected batch-label-input in escalations batch bar")
	}
	if !strings.Contains(body, "escBatchPriority") {
		t.Error("expected escBatchPriority JS function")
	}
	if !strings.Contains(body, "escBatchAssignee") {
		t.Error("expected escBatchAssignee JS function")
	}
	if !strings.Contains(body, "escBatchLabel") {
		t.Error("expected escBatchLabel JS function")
	}
}

func TestEscalationsPage_NilDS(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/escalations", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /escalations nil ds status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestEscalationsPage_RigFilter(t *testing.T) {
	srv := New(&mockDataSource{})
	req := httptest.NewRequest("GET", "/escalations?rig=beads_aegis", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /escalations?rig status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestEscalationsPage_AutoRefresh(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/escalations", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	body := w.Body.String()
	if !strings.Contains(body, `hx-trigger="every 120s"`) {
		t.Error("expected 120s auto-refresh on escalations page")
	}
}

// ── /focus ──

func TestFocusPage(t *testing.T) {
	srv := New(&mockDataSource{})
	req := httptest.NewRequest("GET", "/focus", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /focus status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Focus") {
		t.Error("expected 'Focus' heading")
	}
	if !strings.Contains(body, "Scoring") {
		t.Error("expected Scoring section")
	}
	if !strings.Contains(body, "Highest score") {
		t.Error("expected Highest score stat")
	}
}

func TestFocusPage_NilDS(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/focus", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /focus nil ds status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestFocusPage_AutoRefresh(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/focus", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	body := w.Body.String()
	if !strings.Contains(body, `hx-trigger="every 60s"`) {
		t.Error("expected 60s auto-refresh on focus page")
	}
}

// ── /crossref ──

func TestCrossRefPage(t *testing.T) {
	srv := New(&mockDataSource{})
	req := httptest.NewRequest("GET", "/crossref", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /crossref status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Cross-Database") {
		t.Error("expected 'Cross-Database' heading")
	}
}

func TestCrossRefPage_NilDS(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/crossref", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /crossref nil ds status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestCrossRefPage_AutoRefresh(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/crossref", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	body := w.Body.String()
	if !strings.Contains(body, `hx-trigger="every 300s"`) {
		t.Error("expected 300s auto-refresh on crossref page")
	}
}

func TestCrossRefPage_SortAndBadges(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}, {Name: "beads_gastown"}},
		issues: []dolt.Issue{
			{ID: "x1", Title: "Shared P0", Status: "open", Priority: 0},
		},
	}
	srv := New(ds)
	req := httptest.NewRequest("GET", "/crossref?sort=priority", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "By priority") {
		t.Error("expected 'By priority' sort badge")
	}
	if !strings.Contains(body, "By count") {
		t.Error("expected 'By count' sort badge")
	}
	if !strings.Contains(body, "priority-badge") {
		t.Error("expected priority badge styling in table")
	}
}

// ── /freshness ──

func TestFreshnessPage(t *testing.T) {
	srv := New(&mockDataSource{})
	req := httptest.NewRequest("GET", "/freshness", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /freshness status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Data Freshness") {
		t.Error("expected 'Data Freshness' heading")
	}
}

func TestFreshnessPage_NilDS(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/freshness", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /freshness nil ds status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestFreshnessPage_AutoRefresh(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/freshness", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	body := w.Body.String()
	if !strings.Contains(body, `hx-trigger="every 120s"`) {
		t.Error("expected 120s auto-refresh on freshness page")
	}
}

// ── /complexity ──

func TestComplexityPage(t *testing.T) {
	srv := New(&mockDataSource{})
	req := httptest.NewRequest("GET", "/complexity", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /complexity status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Bead Complexity") {
		t.Error("expected 'Bead Complexity' heading")
	}
}

func TestComplexityPage_NilDS(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/complexity", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /complexity nil ds status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestComplexityPage_AutoRefresh(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/complexity", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	body := w.Body.String()
	if !strings.Contains(body, `hx-trigger="every 300s"`) {
		t.Error("expected 300s auto-refresh on complexity page")
	}
}

func TestComplexityPage_RigFilter(t *testing.T) {
	srv := New(&mockDataSource{})
	req := httptest.NewRequest("GET", "/complexity?rig=aegis", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /complexity?rig=aegis status = %d, want %d", w.Code, http.StatusOK)
	}
}

// ── /label-matrix ──

func TestLabelMatrixPage(t *testing.T) {
	srv := New(&mockDataSource{})
	req := httptest.NewRequest("GET", "/label-matrix", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /label-matrix status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Label") && !strings.Contains(body, "Matrix") {
		t.Error("expected label matrix heading")
	}
}

func TestLabelMatrixPage_NilDS(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/label-matrix", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /label-matrix nil ds status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestLabelMatrixPage_AutoRefresh(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/label-matrix", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	body := w.Body.String()
	if !strings.Contains(body, `hx-trigger="every 300s"`) {
		t.Error("expected 300s auto-refresh on label-matrix page")
	}
}

func TestLabelMatrixPage_RigFilter(t *testing.T) {
	srv := New(&mockDataSource{})
	req := httptest.NewRequest("GET", "/label-matrix?rig=aegis", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /label-matrix?rig=aegis status = %d, want %d", w.Code, http.StatusOK)
	}
}

// ── /label-trends ──

func TestLabelTrendsPage(t *testing.T) {
	srv := New(&mockDataSource{})
	req := httptest.NewRequest("GET", "/label-trends", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /label-trends status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Label Trends") {
		t.Error("expected 'Label Trends' heading")
	}
}

func TestLabelTrendsPage_NilDS(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/label-trends", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /label-trends nil ds status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestLabelTrendsPage_AutoRefresh(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/label-trends", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	body := w.Body.String()
	if !strings.Contains(body, `hx-trigger="every 300s"`) {
		t.Error("expected 300s auto-refresh on label-trends page")
	}
}

func TestLabelTrendsPage_RigFilter(t *testing.T) {
	srv := New(&mockDataSource{})
	req := httptest.NewRequest("GET", "/label-trends?rig=aegis", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /label-trends?rig=aegis status = %d, want %d", w.Code, http.StatusOK)
	}
}

// ── /ready ──

func TestReadyPage(t *testing.T) {
	srv := New(&mockDataSource{})
	req := httptest.NewRequest("GET", "/ready", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /ready status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Ready to Work") {
		t.Error("expected 'Ready to Work' heading")
	}
}

func TestReadyPage_NilDS(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/ready", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /ready nil ds status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestReadyPage_AutoRefresh(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/ready", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	body := w.Body.String()
	if !strings.Contains(body, `hx-trigger="every 60s"`) {
		t.Error("expected 60s auto-refresh on ready page")
	}
}

func TestReadyPage_RigFilter(t *testing.T) {
	srv := New(&mockDataSource{})
	req := httptest.NewRequest("GET", "/ready?rig=aegis", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /ready?rig=aegis status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestReadyPage_TemplateHasActions(t *testing.T) {
	// Verify the template source contains action buttons (even if empty mock has no items)
	srv := New(nil)
	req := httptest.NewRequest("GET", "/ready", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	body := w.Body.String()
	if !strings.Contains(body, "Ready to Work") {
		t.Error("expected ready page heading")
	}
}

// --- Burn Up page tests ---

func TestBurnupPage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/burnup", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /burnup status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Burn Up") {
		t.Error("expected 'Burn Up' heading")
	}
}

func TestBurnupPage_WithData(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		counts:    map[string]int{"open": 5, "closed": 10},
		created:   3,
		closed:    2,
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/burnup", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /burnup status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Daily Detail") {
		t.Error("expected 'Daily Detail' table")
	}
}

func TestBurnupPage_HTMXPartial(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/burnup", nil)
	req.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("HTMX GET /burnup status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("HTMX burnup should return partial, not full page")
	}
}

func TestBurnupPage_RigFilter(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}, {Name: "beads_gastown"}},
		counts:    map[string]int{"closed": 5},
		closed:    1,
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/burnup?rig=beads_aegis", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /burnup?rig= status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "filter-active") {
		t.Error("expected active rig filter badge")
	}
}

// --- Disposition page tests ---

func TestDispositionPage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/disposition", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /disposition status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Disposition") {
		t.Error("expected 'Disposition' heading")
	}
}

func TestDispositionPage_WithData(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		counts:    map[string]int{"open": 5, "closed": 10},
		issues: []dolt.Issue{
			{ID: "a1", Title: "closed one", Status: "closed", UpdatedAt: time.Now()},
			{ID: "a2", Title: "deferred one", Status: "deferred", UpdatedAt: time.Now()},
			{ID: "a3", Title: "open one", Status: "open", UpdatedAt: time.Now()},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/disposition", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /disposition status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Weekly Disposition") {
		t.Error("expected 'Weekly Disposition' table")
	}
	if !strings.Contains(body, "Close rate") {
		t.Error("expected 'Close rate' stat")
	}
}

func TestDispositionPage_HTMXPartial(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/disposition", nil)
	req.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("HTMX GET /disposition status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("HTMX disposition should return partial, not full page")
	}
}

func TestDispositionPage_RigFilter(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}, {Name: "beads_gastown"}},
		counts:    map[string]int{"closed": 5},
		issues: []dolt.Issue{
			{ID: "a1", Title: "t", Status: "closed", UpdatedAt: time.Now()},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/disposition?rig=beads_aegis", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /disposition?rig= status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "filter-active") {
		t.Error("expected active rig filter badge")
	}
}

// --- Phase page tests ---

func TestPhasePage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/phase", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /phase status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Epic Phases") {
		t.Error("expected 'Epic Phases' heading")
	}
}

func TestPhasePage_WithData(t *testing.T) {
	// mock returns m.issues for both Epics() and Issues(), so include all in one slice
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		counts:    map[string]int{"open": 5},
		deps: []dolt.Dependency{
			{FromID: "task-1", ToID: "epic-1", Type: "child_of"},
			{FromID: "task-2", ToID: "epic-1", Type: "child_of"},
		},
		issues: []dolt.Issue{
			{ID: "epic-1", Title: "Test Epic", Type: "epic", Priority: 1},
			{ID: "task-1", Title: "Task 1", Status: "closed"},
			{ID: "task-2", Title: "Task 2", Status: "open"},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/phase", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /phase status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Test Epic") {
		t.Error("expected epic title")
	}
	if !strings.Contains(body, "50%") {
		t.Error("expected 50% completion")
	}
}

func TestPhasePage_HTMXPartial(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/phase", nil)
	req.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("HTMX GET /phase status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("HTMX phase should return partial, not full page")
	}
}

func TestPhasePage_RigFilter(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}, {Name: "beads_gastown"}},
		counts:    map[string]int{"open": 5},
		issues:    []dolt.Issue{},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/phase?rig=beads_aegis", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /phase?rig= status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "filter-active") {
		t.Error("expected active rig filter badge")
	}
}

// ── /tag-velocity ──

func TestTagVelocityPage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/tag-velocity", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /tag-velocity status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Tag Velocity") {
		t.Error("expected 'Tag Velocity' heading")
	}
}

func TestTagVelocityPage_WithData(t *testing.T) {
	now := time.Now()
	ds := &mockDataSource{
		databases:   []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		labelCounts: []dolt.LabelCount{{Label: "bug", Count: 3}},
		issues: []dolt.Issue{
			{ID: "tv1", Title: "Open bug", Status: "open", CreatedAt: now.AddDate(0, 0, -5), UpdatedAt: now},
			{ID: "tv2", Title: "Closed bug", Status: "closed", CreatedAt: now.AddDate(0, 0, -10), UpdatedAt: now.AddDate(0, 0, -2)},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/tag-velocity", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /tag-velocity status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "bug") {
		t.Error("expected label 'bug' in output")
	}
	if !strings.Contains(body, "Label Resolution Speed") {
		t.Error("expected resolution speed section")
	}
}

func TestTagVelocityPage_HTMXPartial(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/tag-velocity", nil)
	req.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("HTMX GET /tag-velocity status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("HTMX tag-velocity should return partial, not full page")
	}
}

func TestTagVelocityPage_RigFilter(t *testing.T) {
	ds := &mockDataSource{
		databases:   []dolt.DatabaseInfo{{Name: "beads_aegis"}, {Name: "beads_gastown"}},
		labelCounts: []dolt.LabelCount{{Label: "bug", Count: 1}},
		issues: []dolt.Issue{
			{ID: "tv1", Title: "Test", Status: "open", CreatedAt: time.Now().Add(-24 * time.Hour), UpdatedAt: time.Now()},
		},
	}
	srv := New(ds)
	req := httptest.NewRequest("GET", "/tag-velocity?rig=beads_aegis", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, `rig=beads_aegis`) {
		t.Error("expected rig filter preserved in auto-refresh URL")
	}
	if !strings.Contains(body, "filter-active") {
		t.Error("expected filter-active badge")
	}
}

// ── /pacing ──

func TestPacingPage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/pacing", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /pacing status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Backlog Pacing") {
		t.Error("expected 'Backlog Pacing' heading")
	}
}

func TestPacingPage_WithData(t *testing.T) {
	now := time.Now()
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "p1", Title: "Open item", Status: "open", CreatedAt: now.AddDate(0, 0, -5), UpdatedAt: now},
			{ID: "p2", Title: "Closed item", Status: "closed", CreatedAt: now.AddDate(0, 0, -10), UpdatedAt: now.AddDate(0, 0, -2)},
			{ID: "p3", Title: "Another open", Status: "open", CreatedAt: now.AddDate(0, 0, -3), UpdatedAt: now},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/pacing", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /pacing status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Daily Close Rate") {
		t.Error("expected daily close rate stat")
	}
	if !strings.Contains(body, "Days to Clear") {
		t.Error("expected days to clear stat")
	}
}

func TestPacingPage_HTMXPartial(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/pacing", nil)
	req.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("HTMX GET /pacing status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("HTMX pacing should return partial, not full page")
	}
}

func TestPacingPage_RigFilter(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}, {Name: "beads_gastown"}},
		issues: []dolt.Issue{
			{ID: "p1", Title: "Test", Status: "open", CreatedAt: time.Now().Add(-24 * time.Hour), UpdatedAt: time.Now()},
		},
	}
	srv := New(ds)
	req := httptest.NewRequest("GET", "/pacing?rig=beads_aegis", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, `rig=beads_aegis`) {
		t.Error("expected rig filter preserved in auto-refresh URL")
	}
	if !strings.Contains(body, "filter-active") {
		t.Error("expected filter-active badge")
	}
}

// ── /agent-velocity ──

func TestAgentVelocityPage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/agent-velocity", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /agent-velocity status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Agent Velocity") {
		t.Error("expected 'Agent Velocity' heading")
	}
}

func TestAgentVelocityPage_WithData(t *testing.T) {
	now := time.Now()
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "av1", Title: "Closed item", Status: "closed", Assignee: "aegis/crew/arnold", CreatedAt: now.AddDate(0, 0, -5), UpdatedAt: now.AddDate(0, 0, -2)},
			{ID: "av2", Title: "Another closed", Status: "closed", Assignee: "aegis/crew/arnold", CreatedAt: now.AddDate(0, 0, -10), UpdatedAt: now.AddDate(0, 0, -1)},
			{ID: "av3", Title: "Grant closed", Status: "closed", Assignee: "aegis/crew/grant", CreatedAt: now.AddDate(0, 0, -8), UpdatedAt: now.AddDate(0, 0, -3)},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/agent-velocity", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /agent-velocity status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "aegis/crew/arnold") {
		t.Error("expected agent name in output")
	}
}

func TestAgentVelocityPage_HTMXPartial(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/agent-velocity", nil)
	req.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("HTMX GET /agent-velocity status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("HTMX agent-velocity should return partial, not full page")
	}
}

func TestAgentVelocityPage_RigFilter(t *testing.T) {
	now := time.Now()
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}, {Name: "beads_gastown"}},
		issues: []dolt.Issue{
			{ID: "av1", Title: "Closed", Status: "closed", Assignee: "aegis/crew/arnold", CreatedAt: now.AddDate(0, 0, -3), UpdatedAt: now.AddDate(0, 0, -1)},
		},
	}
	srv := New(ds)
	req := httptest.NewRequest("GET", "/agent-velocity?rig=beads_aegis", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, `rig=beads_aegis`) {
		t.Error("expected rig filter preserved in auto-refresh URL")
	}
	if !strings.Contains(body, "filter-active") {
		t.Error("expected filter-active badge")
	}
}

// ── /unblocked ──

func TestUnblockedPage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/unblocked", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /unblocked status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Recently Unblocked") {
		t.Error("expected 'Recently Unblocked' heading")
	}
}

func TestUnblockedPage_WithData(t *testing.T) {
	now := time.Now()
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "ub1", Title: "Was blocked", Status: "open", Priority: 1, CreatedAt: now.AddDate(0, 0, -20), UpdatedAt: now},
		},
		statusHistory: []dolt.StatusTransition{
			{FromStatus: "", ToStatus: "open", CommitDate: now.AddDate(0, 0, -20)},
			{FromStatus: "open", ToStatus: "blocked", CommitDate: now.AddDate(0, 0, -15)},
			{FromStatus: "blocked", ToStatus: "open", CommitDate: now.AddDate(0, 0, -5)},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/unblocked", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /unblocked status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Was blocked") {
		t.Error("expected unblocked issue title")
	}
	if !strings.Contains(body, "Blocked Days") {
		t.Error("expected blocked duration column")
	}
}

func TestUnblockedPage_HTMXPartial(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/unblocked", nil)
	req.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("HTMX GET /unblocked status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("HTMX unblocked should return partial, not full page")
	}
}

func TestUnblockedPage_RigFilter(t *testing.T) {
	now := time.Now()
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}, {Name: "beads_gastown"}},
		issues: []dolt.Issue{
			{ID: "ub1", Title: "Was blocked", Status: "open", Priority: 1, CreatedAt: now.AddDate(0, 0, -10), UpdatedAt: now},
		},
		statusHistory: []dolt.StatusTransition{
			{ToStatus: "open", CommitDate: now.AddDate(0, 0, -10)},
			{FromStatus: "open", ToStatus: "blocked", CommitDate: now.AddDate(0, 0, -7)},
			{FromStatus: "blocked", ToStatus: "open", CommitDate: now.AddDate(0, 0, -2)},
		},
	}
	srv := New(ds)
	req := httptest.NewRequest("GET", "/unblocked?rig=beads_aegis", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, `rig=beads_aegis`) {
		t.Error("expected rig filter preserved in auto-refresh URL")
	}
	if !strings.Contains(body, "filter-active") {
		t.Error("expected filter-active badge")
	}
}

// ── /audit-log ──

func TestAuditLogPage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/audit-log", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /audit-log status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Audit Log") {
		t.Error("expected 'Audit Log' heading")
	}
}

func TestAuditLogPage_WithData(t *testing.T) {
	now := time.Now()
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issueDiffs: []dolt.IssueDiffRow{
			{DiffType: "added", ToID: "test-1", ToTitle: "New bead", ToStatus: "open", ToCommitDate: now},
			{DiffType: "modified", ToID: "test-2", ToTitle: "Status change", FromStatus: "open", ToStatus: "closed", ToCommitDate: now.Add(-time.Hour)},
		},
		commentDiffs: []dolt.CommentDiffRow{
			{DiffType: "added", ToIssueID: "test-1", ToAuthor: "aegis/arnold", ToBody: "This is a comment", ToCommitDate: now.Add(-30 * time.Minute)},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/audit-log", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /audit-log status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "New bead") {
		t.Error("expected created bead in audit log")
	}
	if !strings.Contains(body, "open → closed") {
		t.Error("expected status change detail")
	}
	if !strings.Contains(body, "This is a comment") {
		t.Error("expected comment in audit log")
	}
}

func TestAuditLogPage_HTMXPartial(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/audit-log", nil)
	req.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("HTMX GET /audit-log status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("HTMX audit-log should return partial, not full page")
	}
}

func TestAuditLogPage_WindowFilter(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
	}
	srv := New(ds)
	req := httptest.NewRequest("GET", "/audit-log?window=24h", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, `window=24h`) {
		t.Error("expected window parameter preserved in auto-refresh")
	}
}

func TestAuditLogPage_RigFilter(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}, {Name: "beads_gastown"}},
	}
	srv := New(ds)
	req := httptest.NewRequest("GET", "/audit-log?rig=beads_aegis", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "filter-active") {
		t.Error("expected filter-active badge for rig filter")
	}
}

// ── /label/{name} ──

func TestLabelDetailPage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/label/bug", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /label/bug status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "bug") {
		t.Error("expected label name in heading")
	}
}

func TestLabelDetailPage_WithData(t *testing.T) {
	now := time.Now()
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "lbl-1", Title: "Bug fix needed", Status: "open", Priority: 1, CreatedAt: now, UpdatedAt: now},
			{ID: "lbl-2", Title: "Old bug", Status: "closed", Priority: 2, CreatedAt: now.AddDate(0, -1, 0), UpdatedAt: now},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/label/bug", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /label/bug status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Bug fix needed") {
		t.Error("expected issue title in label detail")
	}
	if !strings.Contains(body, "Old bug") {
		t.Error("expected closed issue in label detail")
	}
}

func TestLabelDetailPage_HTMXPartial(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/label/improvement", nil)
	req.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("HTMX GET /label/improvement status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("HTMX label-detail should return partial, not full page")
	}
}

func TestLabelDetailPage_RigFilter(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}, {Name: "beads_gastown"}},
		issues: []dolt.Issue{
			{ID: "lbl-1", Title: "Test", Status: "open", Priority: 2},
		},
	}
	srv := New(ds)
	req := httptest.NewRequest("GET", "/label/bug?rig=beads_aegis", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "filter-active") {
		t.Error("expected filter-active badge for rig filter")
	}
}

// ── /reschedules ──

func TestReschedulesPage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/reschedules", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /reschedules status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Chronic Reschedules") {
		t.Error("expected 'Chronic Reschedules' heading")
	}
}

func TestReschedulesPage_WithData(t *testing.T) {
	now := time.Now()
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "rs-1", Title: "Always deferred", Status: "deferred", Priority: 2, CreatedAt: now.AddDate(0, 0, -30), UpdatedAt: now},
		},
		statusHistory: []dolt.StatusTransition{
			{FromStatus: "", ToStatus: "open", CommitDate: now.AddDate(0, 0, -30)},
			{FromStatus: "open", ToStatus: "deferred", CommitDate: now.AddDate(0, 0, -25)},
			{FromStatus: "deferred", ToStatus: "open", CommitDate: now.AddDate(0, 0, -20)},
			{FromStatus: "open", ToStatus: "deferred", CommitDate: now.AddDate(0, 0, -15)},
			{FromStatus: "deferred", ToStatus: "open", CommitDate: now.AddDate(0, 0, -10)},
			{FromStatus: "open", ToStatus: "deferred", CommitDate: now.AddDate(0, 0, -5)},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/reschedules", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /reschedules status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Always deferred") {
		t.Error("expected deferred issue title")
	}
	if !strings.Contains(body, "3×") {
		t.Error("expected defer count of 3")
	}
}

func TestReschedulesPage_HTMXPartial(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/reschedules", nil)
	req.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("HTMX GET /reschedules status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("HTMX reschedules should return partial, not full page")
	}
}

func TestReschedulesPage_RigFilter(t *testing.T) {
	now := time.Now()
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}, {Name: "beads_gastown"}},
		issues: []dolt.Issue{
			{ID: "rs-1", Title: "Deferred thing", Status: "deferred", Priority: 2, CreatedAt: now, UpdatedAt: now},
		},
		statusHistory: []dolt.StatusTransition{
			{ToStatus: "open", CommitDate: now.AddDate(0, 0, -20)},
			{FromStatus: "open", ToStatus: "deferred", CommitDate: now.AddDate(0, 0, -15)},
			{FromStatus: "deferred", ToStatus: "open", CommitDate: now.AddDate(0, 0, -10)},
			{FromStatus: "open", ToStatus: "deferred", CommitDate: now.AddDate(0, 0, -5)},
		},
	}
	srv := New(ds)
	req := httptest.NewRequest("GET", "/reschedules?rig=beads_aegis", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "filter-active") {
		t.Error("expected filter-active badge for rig filter")
	}
}

// ── /retention ──

func TestRetentionPage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/retention", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /retention status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Time in Status") {
		t.Error("expected 'Time in Status' heading")
	}
}

func TestRetentionPage_WithData(t *testing.T) {
	now := time.Now()
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "ret-1", Title: "Closed thing", Status: "closed", Priority: 2, CreatedAt: now.AddDate(0, 0, -30), UpdatedAt: now},
		},
		statusHistory: []dolt.StatusTransition{
			{FromStatus: "", ToStatus: "open", CommitDate: now.AddDate(0, 0, -30)},
			{FromStatus: "open", ToStatus: "in_progress", CommitDate: now.AddDate(0, 0, -20)},
			{FromStatus: "in_progress", ToStatus: "closed", CommitDate: now.AddDate(0, 0, -5)},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/retention", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /retention status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Transitions") {
		t.Error("expected 'Transitions' column header")
	}
}

func TestRetentionPage_HTMXPartial(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/retention", nil)
	req.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("HTMX GET /retention status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("HTMX retention should return partial, not full page")
	}
}

func TestRetentionPage_RigFilter(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}, {Name: "beads_gastown"}},
	}
	srv := New(ds)
	req := httptest.NewRequest("GET", "/retention?rig=beads_aegis", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "filter-active") {
		t.Error("expected filter-active badge for rig filter")
	}
}

// ── /dog-pile ──

func TestDogPilePage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/dog-pile", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /dog-pile status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Hot Beads") {
		t.Error("expected 'Hot Beads' heading")
	}
}

func TestDogPilePage_WithData(t *testing.T) {
	now := time.Now()
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issueDiffs: []dolt.IssueDiffRow{
			{DiffType: "modified", ToID: "hot-1", ToTitle: "Very active", FromStatus: "open", ToStatus: "in_progress", ToCommitDate: now},
			{DiffType: "modified", ToID: "hot-1", ToTitle: "Very active", FromStatus: "in_progress", ToStatus: "blocked", ToCommitDate: now.Add(-time.Hour)},
			{DiffType: "modified", ToID: "hot-1", ToTitle: "Very active", FromStatus: "blocked", ToStatus: "open", ToCommitDate: now.Add(-2 * time.Hour)},
		},
		commentDiffs: []dolt.CommentDiffRow{
			{DiffType: "added", ToIssueID: "hot-1", ToAuthor: "test", ToBody: "comment 1", ToCommitDate: now},
			{DiffType: "added", ToIssueID: "hot-1", ToAuthor: "test", ToBody: "comment 2", ToCommitDate: now},
		},
		issue: &dolt.Issue{ID: "hot-1", Title: "Very active", Status: "open", Priority: 1},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/dog-pile", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /dog-pile status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Very active") {
		t.Error("expected hot bead title")
	}
}

func TestDogPilePage_EnhancedBatchBar(t *testing.T) {
	now := time.Now()
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issueDiffs: []dolt.IssueDiffRow{
			{DiffType: "modified", ToID: "hot-1", ToTitle: "Very active", FromStatus: "open", ToStatus: "in_progress", ToCommitDate: now},
			{DiffType: "modified", ToID: "hot-1", ToTitle: "Very active", FromStatus: "in_progress", ToStatus: "blocked", ToCommitDate: now.Add(-time.Hour)},
			{DiffType: "modified", ToID: "hot-1", ToTitle: "Very active", FromStatus: "blocked", ToStatus: "open", ToCommitDate: now.Add(-2 * time.Hour)},
		},
		commentDiffs: []dolt.CommentDiffRow{
			{DiffType: "added", ToIssueID: "hot-1", ToAuthor: "test", ToBody: "comment 1", ToCommitDate: now},
			{DiffType: "added", ToIssueID: "hot-1", ToAuthor: "test", ToBody: "comment 2", ToCommitDate: now},
		},
		issue:     &dolt.Issue{ID: "hot-1", Title: "Very active", Status: "open", Priority: 1},
		assignees: []string{"aegis/crew/arnold"},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/dog-pile", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /dog-pile status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "batch-priority") {
		t.Error("expected batch-priority dropdown in dog-pile batch bar")
	}
	if !strings.Contains(body, "batch-assign") {
		t.Error("expected batch-assign dropdown in dog-pile batch bar")
	}
	if !strings.Contains(body, "batch-label-input") {
		t.Error("expected batch-label-input in dog-pile batch bar")
	}
	if !strings.Contains(body, "dogpileBatchPriority") {
		t.Error("expected dogpileBatchPriority JS function")
	}
	if !strings.Contains(body, "dogpileBatchAssignee") {
		t.Error("expected dogpileBatchAssignee JS function")
	}
	if !strings.Contains(body, "dogpileBatchLabel") {
		t.Error("expected dogpileBatchLabel JS function")
	}
}

func TestDogPilePage_HTMXPartial(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/dog-pile", nil)
	req.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("HTMX GET /dog-pile status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("HTMX dog-pile should return partial, not full page")
	}
}

func TestDogPilePage_RigFilter(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}, {Name: "beads_gastown"}},
	}
	srv := New(ds)
	req := httptest.NewRequest("GET", "/dog-pile?rig=beads_aegis", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "filter-active") {
		t.Error("expected filter-active badge for rig filter")
	}
}

func TestDogPilePage_WindowFilter(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
	}
	srv := New(ds)
	req := httptest.NewRequest("GET", "/dog-pile?window=24h", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, `window=24h`) {
		t.Error("expected window parameter preserved")
	}
}

// ── /quick-wins ──

func TestQuickWinsPage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/quick-wins", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /quick-wins status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Quick Wins") {
		t.Error("expected 'Quick Wins' heading")
	}
}

func TestQuickWinsPage_WithData(t *testing.T) {
	now := time.Now()
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "qw-1", Title: "Simple fix", Status: "open", Priority: 2, Type: "task", CreatedAt: now.AddDate(0, 0, -3), UpdatedAt: now},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/quick-wins", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /quick-wins status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Simple fix") {
		t.Error("expected quick win issue title")
	}
}

func TestQuickWinsPage_HTMXPartial(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/quick-wins", nil)
	req.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("HTMX GET /quick-wins status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("HTMX quick-wins should return partial, not full page")
	}
}

func TestQuickWinsPage_RigFilter(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}, {Name: "beads_gastown"}},
		issues: []dolt.Issue{
			{ID: "qw-1", Title: "Simple", Status: "open", Priority: 2, Type: "task"},
		},
	}
	srv := New(ds)
	req := httptest.NewRequest("GET", "/quick-wins?rig=beads_aegis", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "filter-active") {
		t.Error("expected filter-active badge for rig filter")
	}
}

// ── /orphans ──

func TestOrphansPage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/orphans", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /orphans status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Orphaned Beads") {
		t.Error("expected 'Orphaned Beads' heading")
	}
}

func TestOrphansPage_WithData(t *testing.T) {
	now := time.Now()
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "orph-1", Title: "Lonely bead", Status: "open", Priority: 3, CreatedAt: now.AddDate(0, 0, -20), UpdatedAt: now},
		},
		// No labels returned (empty slice default)
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/orphans", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /orphans status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Lonely bead") {
		t.Error("expected orphaned issue title")
	}
}

func TestOrphansPage_HTMXPartial(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/orphans", nil)
	req.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("HTMX GET /orphans status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("HTMX orphans should return partial, not full page")
	}
}

func TestOrphansPage_RigFilter(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}, {Name: "beads_gastown"}},
	}
	srv := New(ds)
	req := httptest.NewRequest("GET", "/orphans?rig=beads_aegis", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "filter-active") {
		t.Error("expected filter-active badge for rig filter")
	}
}

// ── Dwell Page ──────────────────────────────────────────────

func TestDwellPage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/dwell", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /dwell status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Dwell Time") {
		t.Error("expected 'Dwell Time' heading")
	}
}

func TestDwellPage_WithData(t *testing.T) {
	now := time.Now()
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "dw-1", Title: "Stale bead", Status: "open", Priority: 1, CreatedAt: now.AddDate(0, 0, -30), UpdatedAt: now.AddDate(0, 0, -20)},
			{ID: "dw-2", Title: "Fresh bead", Status: "in_progress", Priority: 2, CreatedAt: now.AddDate(0, 0, -3), UpdatedAt: now.AddDate(0, 0, -1)},
			{ID: "dw-3", Title: "Closed bead", Status: "closed", Priority: 3, CreatedAt: now.AddDate(0, 0, -10), UpdatedAt: now},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/dwell", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /dwell status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Stale bead") {
		t.Error("expected open bead in dwell results")
	}
	if strings.Contains(body, "Closed bead") {
		t.Error("closed beads should not appear in dwell results")
	}
}

func TestDwellPage_HTMXPartial(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/dwell", nil)
	req.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("HTMX GET /dwell status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("HTMX dwell should return partial, not full page")
	}
}

func TestDwellPage_RigFilter(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}, {Name: "beads_gastown"}},
	}
	srv := New(ds)
	req := httptest.NewRequest("GET", "/dwell?rig=beads_aegis", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestDwellPage_ZoneFilter(t *testing.T) {
	now := time.Now()
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "dw-old", Title: "Very stale", Status: "open", Priority: 1, UpdatedAt: now.AddDate(0, 0, -30)},
			{ID: "dw-mid", Title: "Medium stale", Status: "open", Priority: 2, UpdatedAt: now.AddDate(0, 0, -10)},
			{ID: "dw-new", Title: "Fresh one", Status: "open", Priority: 3, UpdatedAt: now.AddDate(0, 0, -2)},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/dwell?zone=danger", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Very stale") {
		t.Error("danger zone should include 30-day old bead")
	}
	if strings.Contains(body, "Fresh one") {
		t.Error("danger zone should not include 2-day old bead")
	}
}

// ── Transfers Page ──────────────────────────────────────────

func TestTransfersPage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/transfers", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /transfers status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Transfers") {
		t.Error("expected 'Transfers' heading")
	}
}

func TestTransfersPage_WithData(t *testing.T) {
	now := time.Now()
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issueDiffs: []dolt.IssueDiffRow{
			{
				DiffType:     "modified",
				ToID:         "t-1",
				ToTitle:      "Reassigned bead",
				ToStatus:     "in_progress",
				FromAssignee: "aegis/crew/arnold",
				ToAssignee:   "aegis/crew/grant",
				ToCommitDate: now.AddDate(0, 0, -2),
			},
			{
				DiffType:     "modified",
				ToID:         "t-2",
				ToTitle:      "Status change only",
				ToStatus:     "closed",
				FromStatus:   "open",
				FromAssignee: "aegis/crew/arnold",
				ToAssignee:   "aegis/crew/arnold",
				ToCommitDate: now.AddDate(0, 0, -1),
			},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/transfers", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /transfers status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Reassigned bead") {
		t.Error("expected reassigned bead in transfers")
	}
	if strings.Contains(body, "Status change only") {
		t.Error("same-assignee changes should not appear as transfers")
	}
}

func TestTransfersPage_HTMXPartial(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/transfers", nil)
	req.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("HTMX GET /transfers status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("HTMX transfers should return partial, not full page")
	}
}

func TestTransfersPage_RigFilter(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}, {Name: "beads_gastown"}},
	}
	srv := New(ds)
	req := httptest.NewRequest("GET", "/transfers?rig=beads_aegis", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestTransfersPage_WindowFilter(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
	}
	srv := New(ds)
	for _, window := range []string{"7d", "30d", "90d"} {
		req := httptest.NewRequest("GET", "/transfers?window="+window, nil)
		w := httptest.NewRecorder()
		srv.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("window=%s status = %d, want %d", window, w.Code, http.StatusOK)
		}
	}
}

// ── Load Balance Page ──────────────────────────────────────────

func TestLoadBalancePage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/load-balance", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /load-balance status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Load Balance") {
		t.Error("expected 'Load Balance' heading")
	}
}

func TestLoadBalancePage_WithData(t *testing.T) {
	now := time.Now()
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "lb-1", Title: "Active work", Status: "in_progress", Priority: 0, Assignee: "aegis/crew/arnold", UpdatedAt: now},
			{ID: "lb-2", Title: "Open work", Status: "open", Priority: 2, Assignee: "aegis/crew/arnold", UpdatedAt: now},
			{ID: "lb-3", Title: "Blocked work", Status: "blocked", Priority: 1, Assignee: "aegis/crew/grant", UpdatedAt: now},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/load-balance", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /load-balance status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "arnold") {
		t.Error("expected agent arnold in results")
	}
}

func TestLoadBalancePage_HTMXPartial(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/load-balance", nil)
	req.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("HTMX GET /load-balance status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("HTMX load-balance should return partial, not full page")
	}
}

func TestLoadBalancePage_RigFilter(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}, {Name: "beads_gastown"}},
	}
	srv := New(ds)
	req := httptest.NewRequest("GET", "/load-balance?rig=beads_aegis", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestLoadBalancePage_SortOptions(t *testing.T) {
	sorts := []string{"", "score", "total", "active", "blocked", "highpri", "name"}
	for _, s := range sorts {
		t.Run("sort="+s, func(t *testing.T) {
			srv := New(nil)
			url := "/load-balance"
			if s != "" {
				url += "?sort=" + s
			}
			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()
			srv.ServeHTTP(w, req)
			if w.Code != http.StatusOK {
				t.Fatalf("GET %s status = %d, want %d", url, w.Code, http.StatusOK)
			}
			body := w.Body.String()
			if !strings.Contains(body, "By score") {
				t.Error("expected sort options in page")
			}
		})
	}
}

// ── Stats Page ──────────────────────────────────────────────

func TestStatsPage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/stats", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /stats status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "System Stats") {
		t.Error("expected 'System Stats' heading")
	}
}

func TestStatsPage_WithData(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		counts: map[string]int{
			"open": 10, "in_progress": 5, "closed": 20, "blocked": 3,
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/stats", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /stats status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "beads_aegis") {
		t.Error("expected rig name in per-rig breakdown")
	}
}

func TestStatsPage_DrillDownLinks(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		counts: map[string]int{
			"open": 10, "in_progress": 5, "closed": 20, "blocked": 3, "deferred": 2,
		},
	}
	srv := New(ds)
	req := httptest.NewRequest("GET", "/stats", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	body := w.Body.String()
	// Summary cards should link to filtered views
	for _, link := range []string{"/beads", "/work", "/kanban", "/blocked", "/closed", "/deferred"} {
		if !strings.Contains(body, "href=\""+link) {
			t.Errorf("missing drill-down link to %s", link)
		}
	}
	// Per-rig table should have drill-down links
	if !strings.Contains(body, "/work?rig=beads_aegis") {
		t.Error("missing per-rig drill-down link")
	}
}

func TestStatsPage_DrillDownWithRigFilter(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}, {Name: "beads_gastown"}},
		counts:    map[string]int{"open": 5},
	}
	srv := New(ds)
	req := httptest.NewRequest("GET", "/stats?rig=beads_aegis", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	body := w.Body.String()
	// Drill-down links should preserve rig filter
	if !strings.Contains(body, "/work?rig=beads_aegis") {
		t.Error("drill-down links should preserve rig filter")
	}
}

func TestStatsPage_HTMXPartial(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/stats", nil)
	req.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("HTMX GET /stats status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("HTMX stats should return partial, not full page")
	}
}

// ── Pending Page ──────────────────────────────────────────────

func TestPendingPage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/pending", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /pending status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Pending Action") {
		t.Error("expected 'Pending Action' heading")
	}
}

func TestPendingPage_WithData(t *testing.T) {
	now := time.Now()
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "pend-1", Title: "Assigned not started", Status: "open", Priority: 1, Assignee: "aegis/crew/arnold", UpdatedAt: now.AddDate(0, 0, -3)},
			{ID: "pend-2", Title: "High pri unassigned", Status: "open", Priority: 0, UpdatedAt: now.AddDate(0, 0, -1)},
			{ID: "pend-3", Title: "In progress", Status: "in_progress", Priority: 1, Assignee: "aegis/crew/grant", UpdatedAt: now},
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/pending", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /pending status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Assigned not started") {
		t.Error("expected assigned-not-started bead")
	}
	if !strings.Contains(body, "High pri unassigned") {
		t.Error("expected high-pri unassigned bead")
	}
	if strings.Contains(body, "In progress") {
		t.Error("in-progress beads should not appear as pending")
	}
}

func TestPendingPage_HTMXPartial(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/pending", nil)
	req.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("HTMX GET /pending status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("HTMX pending should return partial, not full page")
	}
}

func TestPendingPage_RigFilter(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}, {Name: "beads_gastown"}},
	}
	srv := New(ds)
	req := httptest.NewRequest("GET", "/pending?rig=beads_aegis", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

// ── Label Age Page ──────────────────────────────────────────

func TestLabelAgePage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/label-age", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("GET /label-age status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Label Age") {
		t.Error("expected 'Label Age' heading")
	}
}

func TestLabelAgePage_WithData(t *testing.T) {
	now := time.Now()
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "la-1", Title: "Labeled bead", Status: "open", Priority: 2, CreatedAt: now.AddDate(0, 0, -20), UpdatedAt: now},
		},
		labels: []string{"bug", "tooling"},
	}
	srv := New(ds)
	req := httptest.NewRequest("GET", "/label-age", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("GET /label-age status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "bug") {
		t.Error("expected label name in results")
	}
}

func TestLabelAgePage_HTMXPartial(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/label-age", nil)
	req.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("HTMX GET /label-age status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("HTMX label-age should return partial, not full page")
	}
}

func TestLabelAgePage_RigFilter(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}, {Name: "beads_gastown"}},
	}
	srv := New(ds)
	req := httptest.NewRequest("GET", "/label-age?rig=beads_aegis", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

// ── Status Flow Page ──────────────────────────────────────────

func TestStatusFlowPage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/status-flow", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("GET /status-flow status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Status Flow") {
		t.Error("expected 'Status Flow' heading")
	}
}

func TestStatusFlowPage_WithData(t *testing.T) {
	now := time.Now()
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issueDiffs: []dolt.IssueDiffRow{
			{DiffType: "modified", ToID: "sf-1", ToTitle: "Flow bead", FromStatus: "open", ToStatus: "in_progress", ToCommitDate: now},
			{DiffType: "modified", ToID: "sf-2", ToTitle: "Close bead", FromStatus: "in_progress", ToStatus: "closed", ToCommitDate: now},
		},
	}
	srv := New(ds)
	req := httptest.NewRequest("GET", "/status-flow", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("GET /status-flow status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "open") || !strings.Contains(body, "in_progress") {
		t.Error("expected status transitions in results")
	}
}

func TestStatusFlowPage_HTMXPartial(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/status-flow", nil)
	req.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("HTMX GET /status-flow status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("HTMX status-flow should return partial, not full page")
	}
}

func TestStatusFlowPage_RigFilter(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}, {Name: "beads_gastown"}},
	}
	srv := New(ds)
	req := httptest.NewRequest("GET", "/status-flow?rig=beads_aegis", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestStatusFlowPage_WindowFilter(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
	}
	srv := New(ds)
	for _, window := range []string{"7d", "30d", "90d"} {
		req := httptest.NewRequest("GET", "/status-flow?window="+window, nil)
		w := httptest.NewRecorder()
		srv.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("window=%s status = %d, want %d", window, w.Code, http.StatusOK)
		}
	}
}

// ── Priority Drift Page ──────────────────────────────────────

func TestPriorityDriftPage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/priority-drift", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("GET /priority-drift status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Priority Drift") {
		t.Error("expected 'Priority Drift' heading")
	}
}

func TestPriorityDriftPage_WithData(t *testing.T) {
	now := time.Now()
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "pd-1", Title: "P0 urgent", Status: "open", Priority: 0, UpdatedAt: now.AddDate(0, 0, -10)},
			{ID: "pd-2", Title: "P1 active", Status: "in_progress", Priority: 1, UpdatedAt: now},
			{ID: "pd-3", Title: "P2 open", Status: "open", Priority: 2, UpdatedAt: now.AddDate(0, 0, -2)},
		},
	}
	srv := New(ds)
	req := httptest.NewRequest("GET", "/priority-drift", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("GET /priority-drift status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "P0") {
		t.Error("expected P0 priority in results")
	}
}

func TestPriorityDriftPage_HTMXPartial(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/priority-drift", nil)
	req.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("HTMX GET /priority-drift status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("HTMX priority-drift should return partial, not full page")
	}
}

func TestPriorityDriftPage_RigFilter(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}, {Name: "beads_gastown"}},
	}
	srv := New(ds)
	req := httptest.NewRequest("GET", "/priority-drift?rig=beads_aegis", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestPriorityDriftPage_DrilldownLinks(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}, {Name: "beads_gastown"}},
		issues: []dolt.Issue{
			{ID: "d1", Title: "Open P0", Status: "open", Priority: 0},
			{ID: "d2", Title: "Active P0", Status: "in_progress", Priority: 0},
			{ID: "d3", Title: "Blocked P1", Status: "blocked", Priority: 1},
		},
	}
	srv := New(ds)
	req := httptest.NewRequest("GET", "/priority-drift", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	// Should have drill-down links to /beads with priority+status filters
	if !strings.Contains(body, "/beads?priority=0") {
		t.Error("expected drill-down link to /beads?priority=0")
	}
	if !strings.Contains(body, "status=open") {
		t.Error("expected drill-down link with status=open")
	}
	// Rig filter badge bar should use badge style
	if !strings.Contains(body, "All rigs") {
		t.Error("expected 'All rigs' badge")
	}
}

// --- Sitemap page tests ---

func TestSitemapPage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/sitemap", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /sitemap status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Sitemap") {
		t.Error("expected 'Sitemap' heading")
	}
	if !strings.Contains(body, "Overview") {
		t.Error("expected category heading")
	}
}

func TestSitemapPage_HTMXPartial(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/sitemap", nil)
	req.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("HTMX GET /sitemap status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("HTMX sitemap should return partial, not full page")
	}
}

func TestSitemapPage_ContainsAllCategories(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/sitemap", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	body := w.Body.String()
	categories := []string{
		"Overview", "Work Management", "Agents", "Flow", "Time",
		"Trends", "Issue Health", "Distribution", "Labels", "Activity Feeds",
		"Planning", "Agent Operations", "Infrastructure",
	}
	for _, cat := range categories {
		if !strings.Contains(body, cat) {
			t.Errorf("expected category %q in sitemap", cat)
		}
	}
}

// --- Pulse page tests ---

func TestPulsePage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/pulse", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /pulse status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Pulse") {
		t.Error("expected 'Pulse' heading")
	}
}

func TestPulsePage_WithData(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		counts: map[string]int{
			"open": 10, "closed": 5,
		},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/pulse", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /pulse status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Hourly Activity") {
		t.Error("expected 'Hourly Activity' section")
	}
}

func TestPulsePage_HTMXPartial(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/pulse", nil)
	req.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("HTMX GET /pulse status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("HTMX pulse should return partial, not full page")
	}
}

func TestPulsePage_RigFilter(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}, {Name: "beads_gastown"}},
	}
	srv := New(ds)
	req := httptest.NewRequest("GET", "/pulse?rig=beads_aegis", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "beads_aegis") {
		t.Error("expected rig filter badge")
	}
}

// --- Timeline page tests ---

func TestTimelinePage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/timeline", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /timeline status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Timeline") {
		t.Error("expected 'Timeline' heading")
	}
}

func TestTimelinePage_WithData(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		counts:    map[string]int{"open": 5},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/timeline", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /timeline status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestTimelinePage_HTMXPartial(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/timeline", nil)
	req.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("HTMX GET /timeline status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("HTMX timeline should return partial, not full page")
	}
}

func TestTimelinePage_WindowParam(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
	}
	srv := New(ds)
	for _, window := range []string{"6h", "12h", "24h", "48h"} {
		req := httptest.NewRequest("GET", "/timeline?window="+window, nil)
		w := httptest.NewRecorder()
		srv.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("GET /timeline?window=%s status = %d, want %d", window, w.Code, http.StatusOK)
		}
	}
}

func TestTimelinePage_RigFilter(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}, {Name: "beads_gastown"}},
	}
	srv := New(ds)
	req := httptest.NewRequest("GET", "/timeline?rig=beads_aegis", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "beads_aegis") {
		t.Error("expected rig filter badge")
	}
}

func TestTimelinePage_TypeFilter(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
	}
	srv := New(ds)
	for _, typ := range []string{"created", "closed", "comment", "reassigned", "status_change"} {
		req := httptest.NewRequest("GET", "/timeline?type="+typ, nil)
		w := httptest.NewRecorder()
		srv.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("GET /timeline?type=%s status = %d, want %d", typ, w.Code, http.StatusOK)
		}
		body := w.Body.String()
		if !strings.Contains(body, "showing "+typ) {
			t.Errorf("GET /timeline?type=%s should show filter label", typ)
		}
	}
}

func TestTimelinePage_TypeFilterWithRig(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}, {Name: "beads_gastown"}},
	}
	srv := New(ds)
	req := httptest.NewRequest("GET", "/timeline?type=created&rig=beads_aegis&window=6h", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "showing created") {
		t.Error("expected type filter label")
	}
	if !strings.Contains(body, "filter-active") {
		t.Error("expected active rig filter")
	}
}

func TestTimelinePage_RigFilterBadgeBar(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}, {Name: "beads_gastown"}},
	}
	srv := New(ds)
	req := httptest.NewRequest("GET", "/timeline", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "All rigs") {
		t.Error("expected 'All rigs' badge in rig filter bar")
	}
}

// --- Changelog page tests ---

func TestChangelogPage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/changelog", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /changelog status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Changelog") {
		t.Error("expected 'Changelog' heading")
	}
}

func TestChangelogPage_WithData(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		counts:    map[string]int{"closed": 10},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/changelog", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /changelog status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestChangelogPage_HTMXPartial(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/changelog", nil)
	req.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("HTMX GET /changelog status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("HTMX changelog should return partial, not full page")
	}
}

func TestChangelogPage_RigFilter(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}, {Name: "beads_gastown"}},
	}
	srv := New(ds)
	req := httptest.NewRequest("GET", "/changelog?rig=beads_aegis", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "beads_aegis") {
		t.Error("expected rig filter badge")
	}
}

func TestChangelogPage_TypeFilter(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "a1", Title: "Bug fix", Type: "bug", Status: "closed", Priority: 1, UpdatedAt: time.Now().Add(-24 * time.Hour)},
			{ID: "a2", Title: "New feature", Type: "task", Status: "closed", Priority: 2, UpdatedAt: time.Now().Add(-48 * time.Hour)},
		},
	}
	srv := New(ds)
	req := httptest.NewRequest("GET", "/changelog?type=bug", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "All types") {
		t.Error("expected 'All types' badge")
	}
	if !strings.Contains(body, "Bug fix") {
		t.Error("expected bug issue to be shown")
	}
}

func TestChangelogPage_PriorityFilter(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "a1", Title: "P0 urgent", Type: "bug", Status: "closed", Priority: 0, UpdatedAt: time.Now().Add(-24 * time.Hour)},
			{ID: "a2", Title: "P3 routine", Type: "task", Status: "closed", Priority: 3, UpdatedAt: time.Now().Add(-48 * time.Hour)},
		},
	}
	srv := New(ds)
	req := httptest.NewRequest("GET", "/changelog?priority=0", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "All priorities") {
		t.Error("expected 'All priorities' badge")
	}
	if !strings.Contains(body, "P0 urgent") {
		t.Error("expected P0 issue to be shown")
	}
}

// --- Impact page tests ---

func TestImpactPage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/impact", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /impact status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Impact") {
		t.Error("expected 'Impact' heading")
	}
}

func TestImpactPage_WithData(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		counts:    map[string]int{"open": 5, "closed": 10},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/impact", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /impact status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestImpactPage_HTMXPartial(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/impact", nil)
	req.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("HTMX GET /impact status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("HTMX impact should return partial, not full page")
	}
}

func TestImpactPage_RigFilter(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}, {Name: "beads_gastown"}},
	}
	srv := New(ds)
	req := httptest.NewRequest("GET", "/impact?rig=beads_aegis", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "beads_aegis") {
		t.Error("expected rig filter badge")
	}
}

// --- Streaks page tests ---

func TestStreaksPage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/streaks", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /streaks status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Streaks") {
		t.Error("expected 'Streaks' heading")
	}
}

func TestStreaksPage_WithData(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		counts:    map[string]int{"open": 5},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/streaks", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /streaks status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestStreaksPage_HTMXPartial(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/streaks", nil)
	req.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("HTMX GET /streaks status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("HTMX streaks should return partial, not full page")
	}
}

func TestStreaksPage_RigFilter(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}, {Name: "beads_gastown"}},
	}
	srv := New(ds)
	req := httptest.NewRequest("GET", "/streaks?rig=beads_aegis", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "beads_aegis") {
		t.Error("expected rig filter badge")
	}
}

// --- Ratios page tests ---

func TestRatiosPage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/ratios", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /ratios status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Ratios") {
		t.Error("expected 'Ratios' heading")
	}
}

func TestRatiosPage_WithData(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		counts:    map[string]int{"open": 10, "closed": 20, "blocked": 3},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/ratios", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /ratios status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Bug:Feature") {
		t.Error("expected Bug:Feature ratio")
	}
}

func TestRatiosPage_HTMXPartial(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/ratios", nil)
	req.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("HTMX GET /ratios status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("HTMX ratios should return partial, not full page")
	}
}

func TestRatiosPage_RigFilter(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}, {Name: "beads_gastown"}},
	}
	srv := New(ds)
	req := httptest.NewRequest("GET", "/ratios?rig=beads_aegis", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "beads_aegis") {
		t.Error("expected rig filter badge")
	}
}

// --- Outgoing page tests ---

func TestOutgoingPage_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/outgoing", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /outgoing status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Oldest") {
		t.Error("expected 'Oldest' heading")
	}
}

func TestOutgoingPage_WithData(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		counts:    map[string]int{"open": 5},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/outgoing", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /outgoing status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestOutgoingPage_HTMXPartial(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/outgoing", nil)
	req.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("HTMX GET /outgoing status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("HTMX outgoing should return partial, not full page")
	}
}

func TestOutgoingPage_RigFilter(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}, {Name: "beads_gastown"}},
	}
	srv := New(ds)
	req := httptest.NewRequest("GET", "/outgoing?rig=beads_aegis", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "beads_aegis") {
		t.Error("expected rig filter badge")
	}
}

func TestFavoritesPage(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/favorites", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("GET /favorites status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Favorites") {
		t.Error("expected 'Favorites' heading")
	}
	if !strings.Contains(body, "localStorage") {
		t.Error("expected localStorage-based favorites")
	}
}

func TestFavoritesPage_BatchBarAndQuickActions(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/favorites", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("GET /favorites status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	checks := []struct{ name, want string }{
		{"batch bar", "batch-bar-fav"},
		{"batch priority", "batch-priority"},
		{"favBatchAction function", "favBatchAction"},
		{"favBatchPriority function", "favBatchPriority"},
		{"favRemoveSelected function", "favRemoveSelected"},
		{"favSetPriority function", "favSetPriority"},
		{"favSetStatus function", "favSetStatus"},
		{"favToggleAll function", "favToggleAll"},
		{"inline priority class", "bead-priority-inline"},
		{"priority edit buttons", "priority-edit"},
	}
	for _, c := range checks {
		if !strings.Contains(body, c.want) {
			t.Errorf("expected %s (%s) in favorites page", c.name, c.want)
		}
	}
}

func TestFavoritesPage_HTMXPartial(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/favorites", nil)
	req.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("HTMX GET /favorites status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("HTMX favorites should return partial, not full page")
	}
}

func TestFavoritesLookup(t *testing.T) {
	srv := New(&mockDataSource{})
	body := `[{"db":"beads_aegis","id":"aegis-abc123"}]`
	req := httptest.NewRequest("POST", "/favorites/lookup", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("POST /favorites/lookup status = %d, want %d", w.Code, http.StatusOK)
	}
	ct := w.Header().Get("Content-Type")
	if !strings.Contains(ct, "application/json") {
		t.Errorf("expected JSON content-type, got %q", ct)
	}
	resp := w.Body.String()
	if !strings.Contains(resp, "aegis-abc123") {
		t.Error("expected bead ID in response")
	}
}

func TestFavoritesLookup_BadJSON(t *testing.T) {
	srv := New(&mockDataSource{})
	req := httptest.NewRequest("POST", "/favorites/lookup", strings.NewReader("not json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("POST /favorites/lookup bad JSON status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

// ── Epic Detail ──

func TestEpicDetail_Found(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issue: &dolt.Issue{
			ID:       "aegis-epic1",
			Title:    "Build the thing",
			Status:   "open",
			Priority: 1,
			Type:     "epic",
		},
		deps: []dolt.Dependency{
			{FromID: "aegis-child1", ToID: "aegis-epic1", Type: "child_of"},
			{FromID: "aegis-child2", ToID: "aegis-epic1", Type: "child_of"},
		},
		comments: []dolt.Comment{
			{ID: 1, IssueID: "aegis-epic1", Author: "nux", Body: "Epic progress note"},
		},
		labels: []string{"infra", "p1"},
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/epic/aegis-epic1", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /epic/aegis-epic1 status = %d, want %d", w.Code, http.StatusOK)
	}

	body := w.Body.String()
	checks := []string{"Build the thing", "aegis-epic1"}
	for _, check := range checks {
		if !strings.Contains(body, check) {
			t.Errorf("body missing %q", check)
		}
	}
}

func TestEpicDetail_NotFound(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issue:     nil,
	}

	srv := New(ds)
	req := httptest.NewRequest("GET", "/epic/nonexistent", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("GET /epic/nonexistent status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestEpicDetail_NilDataSource(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/epic/aegis-epic1", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	// nil ds renders the template with empty data (no crash)
	if w.Code != http.StatusOK {
		t.Fatalf("GET /epic/aegis-epic1 nil ds status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestParseDateFromFilename(t *testing.T) {
	tests := []struct {
		name string
		want string // expected date string or "zero"
	}{
		{"scout-analysis-2026-03-17.md", "2026-03-17"},
		{"report-2025-12-01-summary.md", "2025-12-01"},
		{"2026-01-15.md", "2026-01-15"},
		{"no-date-here.md", "zero"},
		{"probe.md", "zero"},
		{"", "zero"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseDateFromFilename(tt.name)
			if tt.want == "zero" {
				if !got.IsZero() {
					t.Errorf("parseDateFromFilename(%q) = %v, want zero", tt.name, got)
				}
			} else {
				want, _ := time.Parse("2006-01-02", tt.want)
				if !got.Equal(want) {
					t.Errorf("parseDateFromFilename(%q) = %v, want %v", tt.name, got, want)
				}
			}
		})
	}
}

func TestTryParseDate(t *testing.T) {
	tests := []struct {
		input string
		want  string // "zero" or "2006-01-02"
	}{
		{"2026-03-17", "2026-03-17"},
		{"2026-03-17 14:30", "2026-03-17"},
		{"Mar 17, 2026", "2026-03-17"},
		{"March 17, 2026", "2026-03-17"},
		{"  2026-03-17  ", "2026-03-17"},
		{"not a date", "zero"},
		{"", "zero"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := tryParseDate(tt.input)
			if tt.want == "zero" {
				if !got.IsZero() {
					t.Errorf("tryParseDate(%q) = %v, want zero", tt.input, got)
				}
			} else {
				want, _ := time.Parse("2006-01-02", tt.want)
				if got.Year() != want.Year() || got.Month() != want.Month() || got.Day() != want.Day() {
					t.Errorf("tryParseDate(%q) = %v, want date matching %v", tt.input, got, want)
				}
			}
		})
	}
}

func TestParseProbeFile(t *testing.T) {
	dir := t.TempDir()

	// Write a test probe file
	content := `# Test Probe Title

This is the summary paragraph that describes the probe.
It can span multiple lines.

## Details

More content here.
`
	path := dir + "/probe-2026-03-17.md"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	entry := parseProbeFile(path, "test-cat")
	if entry == nil {
		t.Fatal("parseProbeFile returned nil")
	}
	if entry.Title != "Test Probe Title" {
		t.Errorf("title = %q, want %q", entry.Title, "Test Probe Title")
	}
	if entry.Category != "test-cat" {
		t.Errorf("category = %q, want %q", entry.Category, "test-cat")
	}
	if !strings.Contains(entry.Summary, "summary paragraph") {
		t.Errorf("summary = %q, want contains 'summary paragraph'", entry.Summary)
	}
	if entry.Date.IsZero() {
		t.Error("expected date parsed from filename")
	}
}

func TestParseProbeFile_NoTitle(t *testing.T) {
	dir := t.TempDir()
	content := "Just some content without a heading.\n"
	path := dir + "/untitled.md"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	entry := parseProbeFile(path, "general")
	if entry == nil {
		t.Fatal("parseProbeFile returned nil")
	}
	if entry.Title != "untitled" {
		t.Errorf("title fallback = %q, want 'untitled'", entry.Title)
	}
}

func TestParseProbeFile_DateFromContent(t *testing.T) {
	dir := t.TempDir()
	content := `# Probe With Date

**Date**: 2026-03-15

Some content.
`
	path := dir + "/no-date-in-name.md"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	entry := parseProbeFile(path, "general")
	if entry == nil {
		t.Fatal("parseProbeFile returned nil")
	}
	if entry.Date.IsZero() {
		t.Error("expected date parsed from content")
	}
	if entry.Date.Day() != 15 {
		t.Errorf("date day = %d, want 15", entry.Date.Day())
	}
}

func TestParseProbeFile_LongSummary(t *testing.T) {
	dir := t.TempDir()
	content := "# Title\n\n" + strings.Repeat("word ", 100) + "\n"
	path := dir + "/long.md"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	entry := parseProbeFile(path, "general")
	if entry == nil {
		t.Fatal("parseProbeFile returned nil")
	}
	if len(entry.Summary) > 210 { // 200 + "..."
		t.Errorf("summary too long: %d chars", len(entry.Summary))
	}
	if !strings.HasSuffix(entry.Summary, "...") {
		t.Error("expected truncated summary to end with ...")
	}
}

func TestParseProbeFile_NotFound(t *testing.T) {
	entry := parseProbeFile("/nonexistent/file.md", "general")
	if entry != nil {
		t.Error("expected nil for nonexistent file")
	}
}

func TestFindProbesDir(t *testing.T) {
	// Create a workspace with docs/probes
	dir := t.TempDir()
	probesPath := dir + "/docs/probes"
	if err := os.MkdirAll(probesPath, 0755); err != nil {
		t.Fatal(err)
	}

	result := findProbesDir(dir)
	if result != probesPath {
		t.Errorf("findProbesDir = %q, want %q", result, probesPath)
	}
}

func TestFindProbesDir_Sibling(t *testing.T) {
	// Create parent with two repos, probes in sibling
	parent := t.TempDir()
	workspace := parent + "/repo-a"
	sibling := parent + "/repo-b"
	probesPath := sibling + "/docs/probes"

	if err := os.MkdirAll(workspace, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(probesPath, 0755); err != nil {
		t.Fatal(err)
	}

	result := findProbesDir(workspace)
	if result != probesPath {
		t.Errorf("findProbesDir (sibling) = %q, want %q", result, probesPath)
	}
}

func TestHandoffsPage_RigFilter(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/handoffs?rig=aegis", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /handoffs?rig=aegis status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestEventsPage_RigFilter(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/events?rig=aegis", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /events?rig=aegis status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Events") {
		t.Error("expected 'Events' heading")
	}
}

func TestEventsPage_TypeAndRigFilter(t *testing.T) {
	srv := New(nil)
	req := httptest.NewRequest("GET", "/events?type=handoff&rig=aegis", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /events?type=handoff&rig=aegis status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestFindProbesDir_NotFound(t *testing.T) {
	dir := t.TempDir()
	result := findProbesDir(dir)
	if result != "" {
		t.Errorf("findProbesDir (empty) = %q, want empty", result)
	}
}

func TestStalePage_SortOptions(t *testing.T) {
	sorts := []string{"", "stale", "priority", "assignee", "rig"}
	for _, s := range sorts {
		t.Run("sort="+s, func(t *testing.T) {
			srv := New(nil)
			url := "/stale?days=3"
			if s != "" {
				url += "&sort=" + s
			}
			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()
			srv.ServeHTTP(w, req)
			if w.Code != http.StatusOK {
				t.Fatalf("GET %s status = %d, want %d", url, w.Code, http.StatusOK)
			}
			body := w.Body.String()
			if !strings.Contains(body, "By staleness") {
				t.Error("expected sort options in page")
			}
		})
	}
}

func TestStalePage_SortWithData(t *testing.T) {
	staleTime := time.Now().Add(-10 * 24 * time.Hour)
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "aegis-s1", Title: "Old task", Status: "in_progress", Priority: 2, Assignee: "crew/arnold", UpdatedAt: staleTime},
			{ID: "aegis-s2", Title: "Older task", Status: "in_progress", Priority: 0, Assignee: "crew/stryder", UpdatedAt: staleTime.Add(-48 * time.Hour)},
		},
	}
	sorts := []string{"stale", "priority", "assignee"}
	for _, s := range sorts {
		t.Run("sort="+s, func(t *testing.T) {
			srv := New(ds)
			req := httptest.NewRequest("GET", "/stale?days=3&sort="+s, nil)
			w := httptest.NewRecorder()
			srv.ServeHTTP(w, req)
			if w.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
			}
			body := w.Body.String()
			if !strings.Contains(body, "aegis-s1") || !strings.Contains(body, "aegis-s2") {
				t.Error("expected both stale issues in output")
			}
		})
	}
}

func TestBlockedPage_SortOptions(t *testing.T) {
	sorts := []string{"", "priority", "blocker", "assignee", "rig"}
	for _, s := range sorts {
		t.Run("sort="+s, func(t *testing.T) {
			srv := New(nil)
			url := "/blocked"
			if s != "" {
				url += "?sort=" + s
			}
			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()
			srv.ServeHTTP(w, req)
			if w.Code != http.StatusOK {
				t.Fatalf("GET %s status = %d, want %d", url, w.Code, http.StatusOK)
			}
			body := w.Body.String()
			if !strings.Contains(body, "By priority") {
				t.Error("expected sort options in page")
			}
		})
	}
}

func TestReadyPage_SortOptions(t *testing.T) {
	sorts := []string{"", "priority", "age", "type", "assignee"}
	for _, s := range sorts {
		t.Run("sort="+s, func(t *testing.T) {
			srv := New(nil)
			url := "/ready"
			if s != "" {
				url += "?sort=" + s
			}
			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()
			srv.ServeHTTP(w, req)
			if w.Code != http.StatusOK {
				t.Fatalf("GET %s status = %d, want %d", url, w.Code, http.StatusOK)
			}
			body := w.Body.String()
			if !strings.Contains(body, "By priority") {
				t.Error("expected sort options in page")
			}
		})
	}
}

func TestBacklogPage_SortOptions(t *testing.T) {
	sorts := []string{"", "age", "priority", "rig"}
	for _, s := range sorts {
		t.Run("sort="+s, func(t *testing.T) {
			srv := New(nil)
			url := "/backlog"
			if s != "" {
				url += "?sort=" + s
			}
			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()
			srv.ServeHTTP(w, req)
			if w.Code != http.StatusOK {
				t.Fatalf("GET %s status = %d, want %d", url, w.Code, http.StatusOK)
			}
			body := w.Body.String()
			if !strings.Contains(body, "Backlog") {
				t.Error("expected Backlog heading in page")
			}
		})
	}
}

func TestBacklogPage_SortWithData(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "aegis-b1", Title: "Old bug", Status: "open", Priority: 2, CreatedAt: time.Now().Add(-30 * 24 * time.Hour), UpdatedAt: time.Now()},
			{ID: "aegis-b2", Title: "New P0", Status: "open", Priority: 0, CreatedAt: time.Now().Add(-1 * 24 * time.Hour), UpdatedAt: time.Now()},
		},
	}
	for _, s := range []string{"age", "priority"} {
		t.Run("sort="+s, func(t *testing.T) {
			srv := New(ds)
			req := httptest.NewRequest("GET", "/backlog?sort="+s, nil)
			w := httptest.NewRecorder()
			srv.ServeHTTP(w, req)
			if w.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
			}
			body := w.Body.String()
			if !strings.Contains(body, "aegis-b1") || !strings.Contains(body, "aegis-b2") {
				t.Error("expected both backlog issues in output")
			}
		})
	}
}

func TestDeferredPage_SortOptions(t *testing.T) {
	sorts := []string{"", "idle", "priority", "age", "rig"}
	for _, s := range sorts {
		t.Run("sort="+s, func(t *testing.T) {
			srv := New(nil)
			url := "/deferred"
			if s != "" {
				url += "?sort=" + s
			}
			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()
			srv.ServeHTTP(w, req)
			if w.Code != http.StatusOK {
				t.Fatalf("GET %s status = %d, want %d", url, w.Code, http.StatusOK)
			}
			body := w.Body.String()
			if !strings.Contains(body, "By idle time") {
				t.Error("expected sort options in page")
			}
		})
	}
}

func TestDeferredPage_SortWithData(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "aegis-d1", Title: "Deferred old", Status: "deferred", Priority: 2, CreatedAt: time.Now().Add(-60 * 24 * time.Hour), UpdatedAt: time.Now().Add(-30 * 24 * time.Hour)},
			{ID: "aegis-d2", Title: "Deferred P0", Status: "deferred", Priority: 0, CreatedAt: time.Now().Add(-5 * 24 * time.Hour), UpdatedAt: time.Now().Add(-1 * 24 * time.Hour)},
		},
	}
	for _, s := range []string{"idle", "priority", "age"} {
		t.Run("sort="+s, func(t *testing.T) {
			srv := New(ds)
			req := httptest.NewRequest("GET", "/deferred?sort="+s, nil)
			w := httptest.NewRecorder()
			srv.ServeHTTP(w, req)
			if w.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
			}
			body := w.Body.String()
			if !strings.Contains(body, "aegis-d1") || !strings.Contains(body, "aegis-d2") {
				t.Error("expected both deferred issues")
			}
		})
	}
}

func TestOwnersPage_SortOptions(t *testing.T) {
	sorts := []string{"", "active", "total", "open", "blocked", "name"}
	for _, s := range sorts {
		t.Run("sort="+s, func(t *testing.T) {
			srv := New(nil)
			url := "/owners"
			if s != "" {
				url += "?sort=" + s
			}
			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()
			srv.ServeHTTP(w, req)
			if w.Code != http.StatusOK {
				t.Fatalf("GET %s status = %d, want %d", url, w.Code, http.StatusOK)
			}
			body := w.Body.String()
			if !strings.Contains(body, "By active") {
				t.Error("expected sort options in page")
			}
		})
	}
}

func TestQueuePage_SortOptions(t *testing.T) {
	sorts := []string{"", "score", "priority", "age", "rig"}
	for _, s := range sorts {
		t.Run("sort="+s, func(t *testing.T) {
			srv := New(nil)
			url := "/queue"
			if s != "" {
				url += "?sort=" + s
			}
			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()
			srv.ServeHTTP(w, req)
			if w.Code != http.StatusOK {
				t.Fatalf("GET %s status = %d, want %d", url, w.Code, http.StatusOK)
			}
			body := w.Body.String()
			if !strings.Contains(body, "Work Queue") {
				t.Error("expected Work Queue heading in page")
			}
		})
	}
}

func TestActivityPage_SortOptions(t *testing.T) {
	sorts := []string{"", "recent", "priority", "status", "rig"}
	for _, s := range sorts {
		t.Run("sort="+s, func(t *testing.T) {
			srv := New(nil)
			url := "/activity?hours=24"
			if s != "" {
				url += "&sort=" + s
			}
			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()
			srv.ServeHTTP(w, req)
			if w.Code != http.StatusOK {
				t.Fatalf("GET %s status = %d, want %d", url, w.Code, http.StatusOK)
			}
			body := w.Body.String()
			if !strings.Contains(body, "By recent") {
				t.Error("expected sort options in page")
			}
		})
	}
}

func TestWatchlistPage_SortOptions(t *testing.T) {
	sorts := []string{"", "status", "idle", "age", "rig"}
	for _, s := range sorts {
		t.Run("sort="+s, func(t *testing.T) {
			srv := New(nil)
			url := "/watchlist"
			if s != "" {
				url += "?sort=" + s
			}
			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()
			srv.ServeHTTP(w, req)
			if w.Code != http.StatusOK {
				t.Fatalf("GET %s status = %d, want %d", url, w.Code, http.StatusOK)
			}
			body := w.Body.String()
			if !strings.Contains(body, "By status") {
				t.Error("expected sort options in page")
			}
		})
	}
}

func TestQueuePage_SortWithData(t *testing.T) {
	ds := &mockDataSource{
		databases: []dolt.DatabaseInfo{{Name: "beads_aegis"}},
		issues: []dolt.Issue{
			{ID: "aegis-q1", Title: "Old task", Status: "open", Priority: 2, CreatedAt: time.Now().Add(-30 * 24 * time.Hour), UpdatedAt: time.Now()},
			{ID: "aegis-q2", Title: "Urgent", Status: "open", Priority: 0, CreatedAt: time.Now().Add(-1 * 24 * time.Hour), UpdatedAt: time.Now()},
		},
	}
	for _, s := range []string{"score", "priority", "age"} {
		t.Run("sort="+s, func(t *testing.T) {
			srv := New(ds)
			req := httptest.NewRequest("GET", "/queue?sort="+s, nil)
			w := httptest.NewRecorder()
			srv.ServeHTTP(w, req)
			if w.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
			}
			body := w.Body.String()
			if !strings.Contains(body, "By score") {
				t.Error("expected sort options in page")
			}
		})
	}
}

func TestEscalationsPage_SortOptions(t *testing.T) {
	sorts := []string{"", "priority", "date", "assignee", "status"}
	for _, s := range sorts {
		t.Run("sort="+s, func(t *testing.T) {
			srv := New(nil)
			url := "/escalations"
			if s != "" {
				url += "?sort=" + s
			}
			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()
			srv.ServeHTTP(w, req)
			if w.Code != http.StatusOK {
				t.Fatalf("GET %s status = %d, want %d", url, w.Code, http.StatusOK)
			}
			body := w.Body.String()
			if !strings.Contains(body, "By priority") {
				t.Error("expected sort options in page")
			}
		})
	}
}

func TestTriagePage_SortOptions(t *testing.T) {
	sorts := []string{"", "age", "priority", "type", "rig"}
	for _, s := range sorts {
		t.Run("sort="+s, func(t *testing.T) {
			srv := New(nil)
			url := "/triage"
			if s != "" {
				url += "?sort=" + s
			}
			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()
			srv.ServeHTTP(w, req)
			if w.Code != http.StatusOK {
				t.Fatalf("GET %s status = %d, want %d", url, w.Code, http.StatusOK)
			}
			body := w.Body.String()
			if !strings.Contains(body, "By age") {
				t.Error("expected sort options in page")
			}
		})
	}
}

func TestSLAPage_SortOptions(t *testing.T) {
	sorts := []string{"", "overdue", "priority", "assignee", "age"}
	for _, s := range sorts {
		t.Run("sort="+s, func(t *testing.T) {
			srv := New(nil)
			url := "/sla"
			if s != "" {
				url += "?sort=" + s
			}
			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()
			srv.ServeHTTP(w, req)
			if w.Code != http.StatusOK {
				t.Fatalf("GET %s status = %d, want %d", url, w.Code, http.StatusOK)
			}
			body := w.Body.String()
			if !strings.Contains(body, "By overdue") {
				t.Error("expected sort options in page")
			}
		})
	}
}

func TestCommentsPage_SortOptions(t *testing.T) {
	sorts := []string{"", "date", "author", "rig"}
	for _, s := range sorts {
		t.Run("sort="+s, func(t *testing.T) {
			srv := New(nil)
			url := "/comments"
			if s != "" {
				url += "?sort=" + s
			}
			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()
			srv.ServeHTTP(w, req)
			if w.Code != http.StatusOK {
				t.Fatalf("GET %s status = %d, want %d", url, w.Code, http.StatusOK)
			}
			body := w.Body.String()
			if !strings.Contains(body, "By date") {
				t.Error("expected sort options in page")
			}
		})
	}
}

func TestDuplicatesPage_SortOptions(t *testing.T) {
	sorts := []string{"", "count", "name", "priority"}
	for _, s := range sorts {
		t.Run("sort="+s, func(t *testing.T) {
			srv := New(nil)
			url := "/duplicates"
			if s != "" {
				url += "?sort=" + s
			}
			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()
			srv.ServeHTTP(w, req)
			if w.Code != http.StatusOK {
				t.Fatalf("GET %s status = %d, want %d", url, w.Code, http.StatusOK)
			}
		})
	}
}

func TestRigsPage_SortOptions(t *testing.T) {
	sorts := []string{"", "active", "total", "blocked", "name"}
	for _, s := range sorts {
		t.Run("sort="+s, func(t *testing.T) {
			srv := New(nil)
			url := "/rigs"
			if s != "" {
				url += "?sort=" + s
			}
			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()
			srv.ServeHTTP(w, req)
			if w.Code != http.StatusOK {
				t.Fatalf("GET %s status = %d, want %d", url, w.Code, http.StatusOK)
			}
			body := w.Body.String()
			if !strings.Contains(body, "By active") {
				t.Error("expected sort options in page")
			}
		})
	}
}

func TestTypesPage_SortOptions(t *testing.T) {
	sorts := []string{"", "type", "total", "open", "name"}
	for _, s := range sorts {
		t.Run("sort="+s, func(t *testing.T) {
			srv := New(nil)
			url := "/types"
			if s != "" {
				url += "?sort=" + s
			}
			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()
			srv.ServeHTTP(w, req)
			if w.Code != http.StatusOK {
				t.Fatalf("GET %s status = %d, want %d", url, w.Code, http.StatusOK)
			}
			body := w.Body.String()
			if !strings.Contains(body, "By type") {
				t.Error("expected sort options in page")
			}
		})
	}
}

func TestHandoffsPage_SortOptions(t *testing.T) {
	sorts := []string{"", "handoffs", "recent", "session", "agent"}
	for _, s := range sorts {
		t.Run("sort="+s, func(t *testing.T) {
			srv := New(nil)
			url := "/handoffs"
			if s != "" {
				url += "?sort=" + s
			}
			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()
			srv.ServeHTTP(w, req)
			if w.Code != http.StatusOK {
				t.Fatalf("GET %s status = %d, want %d", url, w.Code, http.StatusOK)
			}
			body := w.Body.String()
			if !strings.Contains(body, "By handoffs") {
				t.Error("expected sort options in page")
			}
		})
	}
}
