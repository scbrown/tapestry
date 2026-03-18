package web

import (
	"testing"
	"time"
)

// --- complexityScore ---

func TestComplexityScore(t *testing.T) {
	tests := []struct {
		name     string
		deps     int
		comments int
		descLen  int
		want     int
	}{
		{"zero everything", 0, 0, 0, 0},
		{"deps only", 3, 0, 0, 9},
		{"comments only", 0, 5, 0, 10},
		{"short desc", 0, 0, 30, 0},
		{"medium desc", 0, 0, 100, 1},
		{"long desc", 0, 0, 300, 2},
		{"very long desc", 0, 0, 600, 3},
		{"boundary 50", 0, 0, 50, 0},
		{"boundary 51", 0, 0, 51, 1},
		{"boundary 200", 0, 0, 200, 1},
		{"boundary 201", 0, 0, 201, 2},
		{"boundary 500", 0, 0, 500, 2},
		{"boundary 501", 0, 0, 501, 3},
		{"all factors", 2, 3, 250, 14}, // 6 + 6 + 2
		{"high complexity", 5, 10, 1000, 38}, // 15 + 20 + 3
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := complexityScore(tt.deps, tt.comments, tt.descLen)
			if got != tt.want {
				t.Errorf("complexityScore(%d, %d, %d) = %d, want %d", tt.deps, tt.comments, tt.descLen, got, tt.want)
			}
		})
	}
}

// --- formatDwell, fmtPlural, fmtDurationHelper ---

func TestFormatDwell(t *testing.T) {
	tests := []struct {
		name string
		d    time.Duration
		want string
	}{
		{"zero", 0, "<1 min"},
		{"30 seconds", 30 * time.Second, "<1 min"},
		{"1 minute", time.Minute, "1 min"},
		{"45 minutes", 45 * time.Minute, "45 mins"},
		{"1 hour", time.Hour, "1 hour"},
		{"5 hours", 5 * time.Hour, "5 hours"},
		{"23 hours", 23 * time.Hour, "23 hours"},
		{"1 day", 24 * time.Hour, "1 day"},
		{"3 days", 72 * time.Hour, "3 days"},
		{"6 days", 144 * time.Hour, "6 days"},
		{"1 week", 7 * 24 * time.Hour, "1 week"},
		{"2 weeks", 14 * 24 * time.Hour, "2 weeks"},
		{"1 week 2 days", 9 * 24 * time.Hour, "1 week 2 days"},
		{"3 weeks 4 days", 25 * 24 * time.Hour, "3 weeks 4 days"},
		{"1 week 1 day", 8 * 24 * time.Hour, "1 week 1 day"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatDwell(tt.d)
			if got != tt.want {
				t.Errorf("formatDwell(%v) = %q, want %q", tt.d, got, tt.want)
			}
		})
	}
}

func TestFmtPlural(t *testing.T) {
	tests := []struct {
		n    int
		unit string
		want string
	}{
		{1, "day", "1 day"},
		{2, "day", "2 days"},
		{0, "week", "0 weeks"},
		{1, "hour", "1 hour"},
		{10, "min", "10 mins"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := fmtPlural(tt.n, tt.unit)
			if got != tt.want {
				t.Errorf("fmtPlural(%d, %q) = %q, want %q", tt.n, tt.unit, got, tt.want)
			}
		})
	}
}

func TestFmtDurationHelper(t *testing.T) {
	tests := []struct {
		name string
		d    time.Duration
		want string
	}{
		{"zero", 0, "<1 min"},
		{"seconds only", 30 * time.Second, "<1 min"},
		{"1 min", time.Minute, "1 min"},
		{"59 mins", 59 * time.Minute, "59 mins"},
		{"1 hour", time.Hour, "1 hour"},
		{"3 hours", 3 * time.Hour, "3 hours"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fmtDurationHelper(tt.d)
			if got != tt.want {
				t.Errorf("fmtDurationHelper(%v) = %q, want %q", tt.d, got, tt.want)
			}
		})
	}
}

// --- pct, medianInt ---

func TestPct(t *testing.T) {
	tests := []struct {
		num, denom, want int
	}{
		{0, 0, 0},    // division by zero
		{0, 100, 0},  // zero numerator
		{50, 100, 50},
		{1, 3, 33},   // rounds down
		{2, 3, 67},   // rounds up
		{100, 100, 100},
		{7, 10, 70},
		{1, 1000, 0},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := pct(tt.num, tt.denom)
			if got != tt.want {
				t.Errorf("pct(%d, %d) = %d, want %d", tt.num, tt.denom, got, tt.want)
			}
		})
	}
}

func TestMedianInt(t *testing.T) {
	tests := []struct {
		name string
		vals []int
		want int
	}{
		{"empty", nil, 0},
		{"single", []int{5}, 5},
		{"two values", []int{3, 7}, 5},
		{"odd count", []int{1, 3, 5}, 3},
		{"even count", []int{1, 2, 3, 4}, 2}, // (2+3)/2 = 2 (int division)
		{"unsorted", []int{9, 1, 5, 3, 7}, 5},
		{"duplicates", []int{3, 3, 3}, 3},
		{"negative", []int{-5, -1, 0, 3, 10}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := medianInt(tt.vals)
			if got != tt.want {
				t.Errorf("medianInt(%v) = %d, want %d", tt.vals, got, tt.want)
			}
		})
	}
}

func TestMedianInt_DoesNotMutate(t *testing.T) {
	vals := []int{5, 3, 1}
	medianInt(vals)
	if vals[0] != 5 || vals[1] != 3 || vals[2] != 1 {
		t.Errorf("medianInt mutated input: %v", vals)
	}
}

// --- normalizeTitle ---

func TestNormalizeTitle(t *testing.T) {
	tests := []struct {
		name  string
		title string
		want  string
	}{
		{"plain", "fix login bug", "fix login bug"},
		{"auto prefix", "[AUTO] ScrapeTargetDown: aegis-dolt:9100", "scrapetargetdown"},
		{"alert colon", "HighLatency: web service slow", "highlatency"},
		{"no colon", "simple task title", "simple task title"},
		{"colon far away", "a very long title that happens to have a colon: somewhere after 60 chars", "a very long title that happens to have a colon"},
		{"uppercase", "FIX THE THING", "fix the thing"},
		{"auto only", "[AUTO] just a title", "just a title"},
		{"colon at position 0", ": starts with colon", ": starts with colon"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeTitle(tt.title)
			if got != tt.want {
				t.Errorf("normalizeTitle(%q) = %q, want %q", tt.title, got, tt.want)
			}
		})
	}
}

// --- extractBeadIDs ---

func TestExtractBeadIDs_EdgeCases(t *testing.T) {
	tests := []struct {
		name string
		msg  string
		want []string
	}{
		{"empty", "", nil},
		{"no match", "just a regular commit message", nil},
		{"single bead", "fix aegis-abc123 regression", []string{"aegis-abc123"}},
		{"multiple beads", "close aegis-abc and hq-xyz", []string{"aegis-abc", "hq-xyz"}},
		{"dedup", "aegis-abc aegis-abc aegis-abc", []string{"aegis-abc"}},
		{"false positive co-authored", "Co-Authored-By: co-authored someone", nil},
		{"false positive pre-commit", "pre-commit hook failed", nil},
		{"false positive no-verify", "skip no-verify flag", nil},
		{"false positive non-nil", "check non-nil pointer", nil},
		{"mixed real and false", "fix aegis-abc pre-commit issue", []string{"aegis-abc"}},
		{"short id", "tp-ab is valid", []string{"tp-ab"}},
		{"too short", "x-a is not valid", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractBeadIDs(tt.msg)
			if len(got) != len(tt.want) {
				t.Fatalf("extractBeadIDs(%q) = %v, want %v", tt.msg, got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("extractBeadIDs(%q)[%d] = %q, want %q", tt.msg, i, got[i], tt.want[i])
				}
			}
		})
	}
}

// --- changelogQuery ---

func TestChangelogQuery(t *testing.T) {
	tests := []struct {
		name     string
		rig      string
		typ      string
		priority string
		want     string
	}{
		{"empty all", "", "", "", ""},
		{"rig only", "aegis", "", "", "?rig=aegis"},
		{"type only", "", "bug", "", "?type=bug"},
		{"priority only", "", "", "1", "?priority=1"},
		{"rig and type", "aegis", "task", "", "?rig=aegis&type=task"},
		{"all three", "aegis", "bug", "0", "?priority=0&rig=aegis&type=bug"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := changelogQuery(tt.rig, tt.typ, tt.priority)
			if got != tt.want {
				t.Errorf("changelogQuery(%q, %q, %q) = %q, want %q", tt.rig, tt.typ, tt.priority, got, tt.want)
			}
		})
	}
}

// --- weekStartDate ---

func TestWeekStartDate(t *testing.T) {
	tests := []struct {
		name string
		t    time.Time
		want time.Time
	}{
		{"monday", time.Date(2026, 3, 16, 14, 30, 0, 0, time.UTC), time.Date(2026, 3, 16, 0, 0, 0, 0, time.UTC)},
		{"tuesday", time.Date(2026, 3, 17, 10, 0, 0, 0, time.UTC), time.Date(2026, 3, 16, 0, 0, 0, 0, time.UTC)},
		{"wednesday", time.Date(2026, 3, 18, 0, 0, 0, 0, time.UTC), time.Date(2026, 3, 16, 0, 0, 0, 0, time.UTC)},
		{"sunday", time.Date(2026, 3, 22, 23, 59, 0, 0, time.UTC), time.Date(2026, 3, 16, 0, 0, 0, 0, time.UTC)},
		{"saturday", time.Date(2026, 3, 21, 12, 0, 0, 0, time.UTC), time.Date(2026, 3, 16, 0, 0, 0, 0, time.UTC)},
		{"friday", time.Date(2026, 3, 20, 8, 0, 0, 0, time.UTC), time.Date(2026, 3, 16, 0, 0, 0, 0, time.UTC)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := weekStartDate(tt.t)
			if !got.Equal(tt.want) {
				t.Errorf("weekStartDate(%v) = %v, want %v", tt.t, got, tt.want)
			}
		})
	}
}

// --- buildSignals ---

func TestBuildSignals(t *testing.T) {
	tests := []struct {
		name       string
		data       momentumData
		wantCount  int
		wantStatus map[string]string // signal name → expected status
	}{
		{
			name:       "all green",
			data:       momentumData{ClosedThisWeek: 10, ClosedLastWeek: 10, VelocityDelta: 0, NetFlowWeek: -2, BlockerRatio: 0.05, StalePct: 0.1, OpenCount: 20},
			wantCount:  4,
			wantStatus: map[string]string{"Velocity": "green", "Net Flow": "green", "Blocker Ratio": "green", "Staleness": "green"},
		},
		{
			name:       "all red",
			data:       momentumData{ClosedThisWeek: 2, ClosedLastWeek: 10, VelocityDelta: -8, NetFlowWeek: 10, BlockerRatio: 0.5, StalePct: 0.6, OpenCount: 20},
			wantCount:  4,
			wantStatus: map[string]string{"Velocity": "red", "Net Flow": "red", "Blocker Ratio": "red", "Staleness": "red"},
		},
		{
			name:       "yellow thresholds",
			data:       momentumData{ClosedThisWeek: 8, ClosedLastWeek: 10, VelocityDelta: -2, NetFlowWeek: 3, BlockerRatio: 0.2, StalePct: 0.3, OpenCount: 20},
			wantCount:  4,
			wantStatus: map[string]string{"Velocity": "yellow", "Net Flow": "yellow", "Blocker Ratio": "yellow", "Staleness": "yellow"},
		},
		{
			name:       "zero closures both weeks",
			data:       momentumData{ClosedThisWeek: 0, ClosedLastWeek: 0},
			wantCount:  4,
			wantStatus: map[string]string{"Velocity": "red"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signals := buildSignals(tt.data)
			if len(signals) != tt.wantCount {
				t.Fatalf("got %d signals, want %d", len(signals), tt.wantCount)
			}
			for _, s := range signals {
				if want, ok := tt.wantStatus[s.Name]; ok {
					if s.Status != want {
						t.Errorf("signal %q: status = %q, want %q", s.Name, s.Status, want)
					}
				}
			}
		})
	}
}

// --- formatDelta ---

func TestFormatDelta(t *testing.T) {
	tests := []struct {
		n    int
		want string
	}{
		{0, "0"},
		{5, "+5"},
		{-3, "-3"},
		{100, "+100"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := formatDelta(tt.n)
			if got != tt.want {
				t.Errorf("formatDelta(%d) = %q, want %q", tt.n, got, tt.want)
			}
		})
	}
}

// --- firstNonEmpty ---

func TestFirstNonEmpty(t *testing.T) {
	tests := []struct {
		a, b, want string
	}{
		{"hello", "world", "hello"},
		{"", "fallback", "fallback"},
		{"", "", ""},
		{"x", "", "x"},
	}
	for _, tt := range tests {
		t.Run(tt.a+"|"+tt.b, func(t *testing.T) {
			got := firstNonEmpty(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("firstNonEmpty(%q, %q) = %q, want %q", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

// --- loadScore, fmtScore ---

func TestLoadScore(t *testing.T) {
	tests := []struct {
		name                           string
		open, inProgress, blocked, hp int
		want                          float64
	}{
		{"zero", 0, 0, 0, 0, 0},
		{"open only", 10, 0, 0, 0, 5},        // 10*0.5
		{"in-progress only", 0, 5, 0, 0, 15},  // 5*3
		{"blocked only", 0, 0, 3, 0, 6},       // 3*2
		{"high-pri only", 0, 0, 0, 4, 8},      // 4*2
		{"mixed", 10, 3, 2, 5, 28},            // 10*0.5 + 3*3 + 2*2 + 5*2 = 5+9+4+10
		{"heavy load", 20, 10, 5, 8, 66},      // 20*0.5 + 10*3 + 5*2 + 8*2 = 10+30+10+16
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := loadScore(tt.open, tt.inProgress, tt.blocked, tt.hp)
			if got != tt.want {
				t.Errorf("loadScore(%d,%d,%d,%d) = %.1f, want %.1f", tt.open, tt.inProgress, tt.blocked, tt.hp, got, tt.want)
			}
		})
	}
}

func TestFmtScore(t *testing.T) {
	tests := []struct {
		score float64
		want  string
	}{
		{0, "0"},
		{5.5, "6"},
		{10.0, "10"},
		{99.4, "99"},
		{99.5, "100"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := fmtScore(tt.score)
			if got != tt.want {
				t.Errorf("fmtScore(%.1f) = %q, want %q", tt.score, got, tt.want)
			}
		})
	}
}

// --- percentile ---

func TestPercentile(t *testing.T) {
	tests := []struct {
		name   string
		sorted []float64
		pct    float64
		want   float64
	}{
		{"empty", nil, 50, 0},
		{"single", []float64{5}, 50, 5},
		{"single p0", []float64{5}, 0, 5},
		{"single p100", []float64{5}, 100, 5},
		{"median odd", []float64{1, 3, 5}, 50, 3},
		{"median even", []float64{1, 2, 3, 4}, 50, 2.5},
		{"p90", []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, 90, 9.1},
		{"p0", []float64{1, 2, 3}, 0, 1},
		{"p100", []float64{1, 2, 3}, 100, 3},
		{"p25", []float64{10, 20, 30, 40}, 25, 17.5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := percentile(tt.sorted, tt.pct)
			diff := got - tt.want
			if diff < 0 {
				diff = -diff
			}
			if diff > 0.01 {
				t.Errorf("percentile(%v, %.0f) = %.2f, want %.2f", tt.sorted, tt.pct, got, tt.want)
			}
		})
	}
}
