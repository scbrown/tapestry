package dolt

import (
	"strings"
	"testing"
	"time"
)

func TestBuildIssueQuery(t *testing.T) {
	tests := []struct {
		name     string
		filter   IssueFilter
		asOf     string
		wantArgs int
		contains []string
	}{
		{
			name:     "no filters",
			filter:   IssueFilter{},
			wantArgs: 0,
			contains: []string{"SELECT", "FROM issues", "issue_type IN", "ORDER BY updated_at DESC"},
		},
		{
			name:     "status filter",
			filter:   IssueFilter{Status: "closed"},
			wantArgs: 1,
			contains: []string{"status = ?"},
		},
		{
			name:     "multiple filters",
			filter:   IssueFilter{Status: "open", Priority: 1, Type: "bug"},
			wantArgs: 3,
			contains: []string{"status = ?", "priority = ?", "issue_type = ?"},
		},
		{
			name:     "with limit",
			filter:   IssueFilter{Limit: 10},
			wantArgs: 0,
			contains: []string{"LIMIT 10"},
		},
		{
			name:     "with AS OF",
			filter:   IssueFilter{Status: "closed"},
			asOf:     "2026-02-01T00:00:00",
			wantArgs: 1,
			contains: []string{"AS OF '2026-02-01T00:00:00'", "status = ?"},
		},
		{
			name: "with date range",
			filter: IssueFilter{
				Status:        "closed",
				UpdatedAfter:  time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
				UpdatedBefore: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
			},
			wantArgs: 3,
			contains: []string{"status = ?", "updated_at >= ?", "updated_at < ?"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSQL, gotArgs := buildIssueQuery(tt.filter, tt.asOf)
			for _, s := range tt.contains {
				if !strings.Contains(gotSQL, s) {
					t.Errorf("SQL %q does not contain %q", gotSQL, s)
				}
			}
			if len(gotArgs) != tt.wantArgs {
				t.Errorf("args count = %d, want %d", len(gotArgs), tt.wantArgs)
			}
		})
	}
}

func TestBuildIssueQuery_NoUsePrefix(t *testing.T) {
	sql, _ := buildIssueQuery(IssueFilter{}, "")
	if strings.HasPrefix(sql, "USE") {
		t.Error("buildIssueQuery should not include USE prefix")
	}
}

func TestBuildIssueQuery_LabelFilter(t *testing.T) {
	sql, args := buildIssueQuery(IssueFilter{Label: "urgent"}, "")
	if !strings.Contains(sql, "SELECT issue_id FROM labels WHERE label = ?") {
		t.Errorf("SQL missing label subquery, got: %s", sql)
	}
	if len(args) != 1 || args[0] != "urgent" {
		t.Errorf("args = %v, want [urgent]", args)
	}
}

func TestBuildIssueQuery_AssigneeFilter(t *testing.T) {
	sql, args := buildIssueQuery(IssueFilter{Assignee: "arnold"}, "")
	if !strings.Contains(sql, "assignee = ?") {
		t.Errorf("SQL missing assignee filter, got: %s", sql)
	}
	if len(args) != 1 || args[0] != "arnold" {
		t.Errorf("args = %v, want [arnold]", args)
	}
}

func TestBuildIssueQuery_OwnerFilter(t *testing.T) {
	sql, args := buildIssueQuery(IssueFilter{Owner: "stiwi"}, "")
	if !strings.Contains(sql, "owner = ?") {
		t.Errorf("SQL missing owner filter, got: %s", sql)
	}
	if len(args) != 1 || args[0] != "stiwi" {
		t.Errorf("args = %v, want [stiwi]", args)
	}
}

func TestBuildIssueQuery_TypeFilterOverridesDefault(t *testing.T) {
	sql, _ := buildIssueQuery(IssueFilter{Type: "epic"}, "")
	if strings.Contains(sql, "issue_type IN") {
		t.Error("explicit type filter should not include default type IN clause")
	}
	if !strings.Contains(sql, "issue_type = ?") {
		t.Errorf("SQL missing explicit type filter, got: %s", sql)
	}
}

func TestBuildIssueQuery_AllFilters(t *testing.T) {
	f := IssueFilter{
		Status:        "in_progress",
		Priority:      1,
		Type:          "bug",
		Assignee:      "arnold",
		Owner:         "stiwi",
		Label:         "critical",
		Limit:         50,
		UpdatedAfter:  time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedBefore: time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
	}
	sql, args := buildIssueQuery(f, "")
	for _, want := range []string{"status = ?", "priority = ?", "issue_type = ?", "assignee = ?", "owner = ?", "label = ?", "updated_at >= ?", "updated_at < ?", "LIMIT 50"} {
		if !strings.Contains(sql, want) {
			t.Errorf("SQL missing %q, got: %s", want, sql)
		}
	}
	if len(args) != 8 {
		t.Errorf("args count = %d, want 8", len(args))
	}
}
