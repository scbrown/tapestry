package dolt

import "testing"

func TestBuildIssueQuery(t *testing.T) {
	tests := []struct {
		name     string
		database string
		filter   IssueFilter
		asOf     string
		wantSQL  string
		wantArgs int
	}{
		{
			name:     "no filters",
			database: "beads_aegis",
			filter:   IssueFilter{},
			wantSQL:  "USE `beads_aegis`; SELECT id, title, description, status, priority, type, owner, assignee, labels, created_at, updated_at FROM issues ORDER BY updated_at DESC",
			wantArgs: 0,
		},
		{
			name:     "status filter",
			database: "beads_aegis",
			filter:   IssueFilter{Status: "closed"},
			wantSQL:  "USE `beads_aegis`; SELECT id, title, description, status, priority, type, owner, assignee, labels, created_at, updated_at FROM issues WHERE status = ? ORDER BY updated_at DESC",
			wantArgs: 1,
		},
		{
			name:     "multiple filters",
			database: "beads_gastown",
			filter:   IssueFilter{Status: "open", Priority: 1, Type: "bug"},
			wantSQL:  "USE `beads_gastown`; SELECT id, title, description, status, priority, type, owner, assignee, labels, created_at, updated_at FROM issues WHERE status = ? AND priority = ? AND type = ? ORDER BY updated_at DESC",
			wantArgs: 3,
		},
		{
			name:     "with limit",
			database: "beads_aegis",
			filter:   IssueFilter{Limit: 10},
			wantSQL:  "USE `beads_aegis`; SELECT id, title, description, status, priority, type, owner, assignee, labels, created_at, updated_at FROM issues ORDER BY updated_at DESC LIMIT 10",
			wantArgs: 0,
		},
		{
			name:     "with AS OF",
			database: "beads_aegis",
			filter:   IssueFilter{Status: "closed"},
			asOf:     "2026-02-01T00:00:00",
			wantSQL:  "USE `beads_aegis`; SELECT id, title, description, status, priority, type, owner, assignee, labels, created_at, updated_at FROM issues AS OF '2026-02-01T00:00:00' WHERE status = ? ORDER BY updated_at DESC",
			wantArgs: 1,
		},
		{
			name:     "assignee filter",
			database: "beads_aegis",
			filter:   IssueFilter{Assignee: "nux"},
			wantSQL:  "USE `beads_aegis`; SELECT id, title, description, status, priority, type, owner, assignee, labels, created_at, updated_at FROM issues WHERE assignee = ? ORDER BY updated_at DESC",
			wantArgs: 1,
		},
		{
			name:     "all filters with limit and AS OF",
			database: "beads_work",
			filter:   IssueFilter{Status: "open", Priority: 2, Type: "task", Assignee: "goldblum", Limit: 25},
			asOf:     "2026-01-15T12:00:00",
			wantSQL:  "USE `beads_work`; SELECT id, title, description, status, priority, type, owner, assignee, labels, created_at, updated_at FROM issues AS OF '2026-01-15T12:00:00' WHERE status = ? AND priority = ? AND type = ? AND assignee = ? ORDER BY updated_at DESC LIMIT 25",
			wantArgs: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSQL, gotArgs := buildIssueQuery(tt.database, tt.filter, tt.asOf)
			if gotSQL != tt.wantSQL {
				t.Errorf("SQL =\n  %q\nwant\n  %q", gotSQL, tt.wantSQL)
			}
			if len(gotArgs) != tt.wantArgs {
				t.Errorf("args count = %d, want %d", len(gotArgs), tt.wantArgs)
			}
		})
	}
}

func TestUseDB(t *testing.T) {
	got := useDB("beads_aegis")
	want := "USE `beads_aegis`; "
	if got != want {
		t.Errorf("useDB() = %q, want %q", got, want)
	}
}
