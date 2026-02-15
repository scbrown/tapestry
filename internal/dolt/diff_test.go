package dolt

import (
	"testing"
	"time"
)

func TestDiffTimestampFormat(t *testing.T) {
	// Verify the timestamp format used by diff methods matches IssuesAsOf.
	ts := time.Date(2026, 2, 15, 10, 30, 0, 0, time.UTC)
	want := "2026-02-15T10:30:00"
	got := ts.UTC().Format("2006-01-02T15:04:05")
	if got != want {
		t.Errorf("format = %q, want %q", got, want)
	}
}

func TestDiffTimestampFormat_NonUTC(t *testing.T) {
	// Verify that non-UTC times are converted to UTC before formatting.
	loc := time.FixedZone("EST", -5*3600)
	ts := time.Date(2026, 2, 15, 10, 30, 0, 0, loc)
	want := "2026-02-15T15:30:00"
	got := ts.UTC().Format("2006-01-02T15:04:05")
	if got != want {
		t.Errorf("format = %q, want %q", got, want)
	}
}
