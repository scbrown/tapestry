package events

import (
	"testing"
	"time"

	"github.com/scbrown/tapestry/internal/dolt"
)

func date(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func TestIssueEvents_Created(t *testing.T) {
	d := dolt.IssueDiffRow{
		DiffType:     "added",
		ToID:         "aegis-001",
		ToTitle:      "Fix login bug",
		ToStatus:     "open",
		ToOwner:      "goldblum",
		ToAssignee:   "nux",
		ToCommitDate: date(2026, 2, 10),
	}

	events := issueEvents(d)
	if len(events) != 1 {
		t.Fatalf("got %d events, want 1", len(events))
	}

	ev := events[0]
	if ev.Action != ActionCreated {
		t.Errorf("action = %q, want %q", ev.Action, ActionCreated)
	}
	if ev.Actor != "goldblum" {
		t.Errorf("actor = %q, want %q", ev.Actor, "goldblum")
	}
	if ev.BeadID != "aegis-001" {
		t.Errorf("bead_id = %q, want %q", ev.BeadID, "aegis-001")
	}
	if ev.Title != "Fix login bug" {
		t.Errorf("title = %q, want %q", ev.Title, "Fix login bug")
	}
}

func TestIssueEvents_Closed(t *testing.T) {
	d := dolt.IssueDiffRow{
		DiffType:     "modified",
		ToID:         "aegis-002",
		ToTitle:      "Add feature",
		ToStatus:     "closed",
		ToOwner:      "nux",
		FromStatus:   "in_progress",
		ToCommitDate: date(2026, 2, 12),
	}

	events := issueEvents(d)
	if len(events) != 1 {
		t.Fatalf("got %d events, want 1", len(events))
	}
	if events[0].Action != ActionClosed {
		t.Errorf("action = %q, want %q", events[0].Action, ActionClosed)
	}
	if events[0].Details["from_status"] != "in_progress" {
		t.Errorf("from_status = %q, want %q", events[0].Details["from_status"], "in_progress")
	}
}

func TestIssueEvents_Reopened(t *testing.T) {
	d := dolt.IssueDiffRow{
		DiffType:     "modified",
		ToID:         "aegis-003",
		ToTitle:      "Flaky test",
		ToStatus:     "open",
		ToOwner:      "goldblum",
		FromStatus:   "closed",
		ToCommitDate: date(2026, 2, 13),
	}

	events := issueEvents(d)
	if len(events) != 1 {
		t.Fatalf("got %d events, want 1", len(events))
	}
	if events[0].Action != ActionReopened {
		t.Errorf("action = %q, want %q", events[0].Action, ActionReopened)
	}
}

func TestIssueEvents_StatusChange(t *testing.T) {
	d := dolt.IssueDiffRow{
		DiffType:     "modified",
		ToID:         "aegis-004",
		ToTitle:      "Refactor auth",
		ToStatus:     "in_progress",
		ToOwner:      "nux",
		FromStatus:   "open",
		ToCommitDate: date(2026, 2, 11),
	}

	events := issueEvents(d)
	if len(events) != 1 {
		t.Fatalf("got %d events, want 1", len(events))
	}
	if events[0].Action != ActionStatus {
		t.Errorf("action = %q, want %q", events[0].Action, ActionStatus)
	}
}

func TestIssueEvents_AssigneeChange(t *testing.T) {
	d := dolt.IssueDiffRow{
		DiffType:     "modified",
		ToID:         "aegis-005",
		ToTitle:      "Database migration",
		ToStatus:     "open",
		ToOwner:      "goldblum",
		ToAssignee:   "nux",
		FromStatus:   "open",
		FromAssignee: "furiosa",
		ToCommitDate: date(2026, 2, 14),
	}

	events := issueEvents(d)
	if len(events) != 1 {
		t.Fatalf("got %d events, want 1", len(events))
	}
	if events[0].Action != ActionAssigned {
		t.Errorf("action = %q, want %q", events[0].Action, ActionAssigned)
	}
	if events[0].Details["from_assignee"] != "furiosa" {
		t.Errorf("from_assignee = %q, want %q", events[0].Details["from_assignee"], "furiosa")
	}
	if events[0].Details["to_assignee"] != "nux" {
		t.Errorf("to_assignee = %q, want %q", events[0].Details["to_assignee"], "nux")
	}
}

func TestIssueEvents_StatusAndAssigneeChange(t *testing.T) {
	// Both status and assignee change in the same diff row should produce two events.
	d := dolt.IssueDiffRow{
		DiffType:     "modified",
		ToID:         "aegis-006",
		ToTitle:      "Complex change",
		ToStatus:     "in_progress",
		ToOwner:      "goldblum",
		ToAssignee:   "nux",
		FromStatus:   "open",
		FromAssignee: "",
		ToCommitDate: date(2026, 2, 15),
	}

	events := issueEvents(d)
	if len(events) != 2 {
		t.Fatalf("got %d events, want 2", len(events))
	}

	actions := map[string]bool{}
	for _, ev := range events {
		actions[ev.Action] = true
	}
	if !actions[ActionStatus] {
		t.Error("missing status_change event")
	}
	if !actions[ActionAssigned] {
		t.Error("missing assigned event")
	}
}

func TestIssueEvents_NoChange(t *testing.T) {
	// Modified but no status or assignee change (e.g., title edit).
	d := dolt.IssueDiffRow{
		DiffType:     "modified",
		ToID:         "aegis-007",
		ToTitle:      "Updated title",
		ToStatus:     "open",
		ToOwner:      "goldblum",
		ToAssignee:   "nux",
		FromStatus:   "open",
		FromAssignee: "nux",
		ToCommitDate: date(2026, 2, 15),
	}

	events := issueEvents(d)
	if len(events) != 0 {
		t.Fatalf("got %d events, want 0", len(events))
	}
}

func TestIssueEvents_Removed(t *testing.T) {
	d := dolt.IssueDiffRow{
		DiffType:     "removed",
		ToID:         "aegis-008",
		ToTitle:      "Deleted bead",
		ToOwner:      "goldblum",
		ToCommitDate: date(2026, 2, 15),
	}

	events := issueEvents(d)
	if len(events) != 1 {
		t.Fatalf("got %d events, want 1", len(events))
	}
	if events[0].Action != ActionRemoved {
		t.Errorf("action = %q, want %q", events[0].Action, ActionRemoved)
	}
}

func TestApplyFilter(t *testing.T) {
	events := []Event{
		{Action: ActionCreated, Actor: "goldblum", BeadID: "a-1"},
		{Action: ActionClosed, Actor: "nux", BeadID: "a-2"},
		{Action: ActionCommented, Actor: "goldblum", BeadID: "a-1"},
		{Action: ActionStatus, Actor: "nux", BeadID: "a-3"},
	}

	tests := []struct {
		name   string
		filter EventFilter
		want   int
	}{
		{"no filter", EventFilter{}, 4},
		{"by bead", EventFilter{BeadID: "a-1"}, 2},
		{"by actor", EventFilter{Actor: "nux"}, 2},
		{"by action", EventFilter{Action: ActionCreated}, 1},
		{"with limit", EventFilter{Limit: 2}, 2},
		{"combined", EventFilter{Actor: "goldblum", Action: ActionCommented}, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := applyFilter(events, tt.filter)
			if len(got) != tt.want {
				t.Errorf("got %d events, want %d", len(got), tt.want)
			}
		})
	}
}

func TestMonthly(t *testing.T) {
	events := []Event{
		{Timestamp: date(2026, 1, 5), Action: ActionCreated, Actor: "goldblum"},
		{Timestamp: date(2026, 1, 10), Action: ActionClosed, Actor: "nux"},
		{Timestamp: date(2026, 1, 15), Action: ActionCommented, Actor: "goldblum"},
		{Timestamp: date(2026, 2, 1), Action: ActionCreated, Actor: "nux"},
		{Timestamp: date(2026, 2, 5), Action: ActionCreated, Actor: "goldblum"},
		{Timestamp: date(2026, 2, 10), Action: ActionClosed, Actor: "goldblum"},
	}

	stats := Monthly(events)
	if len(stats) != 2 {
		t.Fatalf("got %d months, want 2", len(stats))
	}

	// January
	jan := stats[0]
	if jan.Month != date(2026, 1, 1) {
		t.Errorf("month[0] = %v, want 2026-01-01", jan.Month)
	}
	if jan.Created != 1 {
		t.Errorf("jan.Created = %d, want 1", jan.Created)
	}
	if jan.Closed != 1 {
		t.Errorf("jan.Closed = %d, want 1", jan.Closed)
	}
	if jan.Comments != 1 {
		t.Errorf("jan.Comments = %d, want 1", jan.Comments)
	}
	if jan.ByActor["goldblum"] != 2 {
		t.Errorf("jan.ByActor[goldblum] = %d, want 2", jan.ByActor["goldblum"])
	}

	// February
	feb := stats[1]
	if feb.Month != date(2026, 2, 1) {
		t.Errorf("month[1] = %v, want 2026-02-01", feb.Month)
	}
	if feb.Created != 2 {
		t.Errorf("feb.Created = %d, want 2", feb.Created)
	}
	if feb.Closed != 1 {
		t.Errorf("feb.Closed = %d, want 1", feb.Closed)
	}
}

func TestMonthly_Empty(t *testing.T) {
	stats := Monthly(nil)
	if len(stats) != 0 {
		t.Fatalf("got %d months, want 0", len(stats))
	}
}

func TestAgentActivity(t *testing.T) {
	events := []Event{
		{Action: ActionCreated, Actor: "goldblum"},
		{Action: ActionClosed, Actor: "goldblum"},
		{Action: ActionCommented, Actor: "goldblum"},
		{Action: ActionCreated, Actor: "nux"},
		{Action: ActionCommented, Actor: "nux"},
		{Action: ActionStatus, Actor: ""}, // no actor
	}

	stats := AgentActivity(events)
	if len(stats) != 2 {
		t.Fatalf("got %d agents, want 2", len(stats))
	}

	// goldblum should be first (3 total > 2 total).
	if stats[0].Agent != "goldblum" {
		t.Errorf("stats[0].Agent = %q, want %q", stats[0].Agent, "goldblum")
	}
	if stats[0].Total != 3 {
		t.Errorf("goldblum.Total = %d, want 3", stats[0].Total)
	}
	if stats[0].Created != 1 {
		t.Errorf("goldblum.Created = %d, want 1", stats[0].Created)
	}
	if stats[0].Closed != 1 {
		t.Errorf("goldblum.Closed = %d, want 1", stats[0].Closed)
	}
	if stats[0].Comments != 1 {
		t.Errorf("goldblum.Comments = %d, want 1", stats[0].Comments)
	}

	if stats[1].Agent != "nux" {
		t.Errorf("stats[1].Agent = %q, want %q", stats[1].Agent, "nux")
	}
	if stats[1].Total != 2 {
		t.Errorf("nux.Total = %d, want 2", stats[1].Total)
	}
}

func TestAgentActivity_Empty(t *testing.T) {
	stats := AgentActivity(nil)
	if len(stats) != 0 {
		t.Fatalf("got %d agents, want 0", len(stats))
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input string
		n     int
		want  string
	}{
		{"short", 10, "short"},
		{"exactly10!", 10, "exactly10!"},
		{"this is a longer string", 10, "this is a ..."},
		{"", 5, ""},
	}

	for _, tt := range tests {
		got := truncate(tt.input, tt.n)
		if got != tt.want {
			t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.n, got, tt.want)
		}
	}
}
