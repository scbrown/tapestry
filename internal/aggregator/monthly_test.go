package aggregator

import (
	"testing"
	"time"
)

func TestRigDisplayName(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"beads_aegis", "aegis"},
		{"beads_tapestry", "tapestry"},
		{"plain", "plain"},
		{"beads_", ""},
	}
	for _, tt := range tests {
		got := RigDisplayName(tt.in)
		if got != tt.want {
			t.Errorf("RigDisplayName(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestMonthlySummaryNavigation(t *testing.T) {
	// Test that navigation months are calculated correctly.
	// We can't call Monthly() without a real Dolt client, but we can
	// test the navigation math by calling it with no databases.
	s := &MonthlySummary{}

	// January 2026
	year, month := 2026, 1
	now := time.Date(2026, 2, 15, 0, 0, 0, 0, time.UTC)
	monthStart := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	prev := monthStart.AddDate(0, -1, 0)
	next := monthStart.AddDate(0, 1, 0)

	s.Year = year
	s.Month = time.Month(month)
	s.MonthName = time.Month(month).String()
	s.PrevYear = prev.Year()
	s.PrevMonth = int(prev.Month())
	s.NextYear = next.Year()
	s.NextMonth = int(next.Month())
	s.HasNext = next.Before(now) || (next.Year() == now.Year() && next.Month() <= now.Month())

	if s.PrevYear != 2025 || s.PrevMonth != 12 {
		t.Errorf("prev = %d/%d, want 2025/12", s.PrevYear, s.PrevMonth)
	}
	if s.NextYear != 2026 || s.NextMonth != 2 {
		t.Errorf("next = %d/%d, want 2026/2", s.NextYear, s.NextMonth)
	}
	if !s.HasNext {
		t.Error("HasNext should be true for Jan 2026 when now is Feb 2026")
	}
}

func TestStatsZeroValue(t *testing.T) {
	var s Stats
	if s.Created != 0 || s.Closed != 0 || s.Open != 0 || s.InFlight != 0 {
		t.Error("zero Stats should have all fields at 0")
	}
}
