package dolt

import (
	"strings"
	"testing"
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
			contains: []string{"SELECT", "FROM issues", "deleted_at IS NULL", "ORDER BY updated_at DESC"},
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
