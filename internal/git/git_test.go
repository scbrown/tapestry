package git

import (
	"testing"
)

func TestExtractBeadIDs(t *testing.T) {
	tests := []struct {
		name    string
		message string
		want    []string
	}{
		{
			name:    "single bead ID",
			message: "fix auth bug in aegis-9i6i",
			want:    []string{"aegis-9i6i"},
		},
		{
			name:    "multiple bead IDs",
			message: "aegis-abc1: closes hq-c27u7 and aegis-9i6i",
			want:    []string{"aegis-abc1", "hq-c27u7", "aegis-9i6i"},
		},
		{
			name:    "no bead IDs",
			message: "refactor auth module",
			want:    nil,
		},
		{
			name:    "bead ID at start",
			message: "aegis-0a9: fix token refresh",
			want:    []string{"aegis-0a9"},
		},
		{
			name:    "unknown prefix ignored",
			message: "use some-thing in config",
			want:    nil,
		},
		{
			name:    "deduplicates",
			message: "aegis-abc1 fixes aegis-abc1",
			want:    []string{"aegis-abc1"},
		},
		{
			name:    "gastown prefix",
			message: "gastown-x1y2: add new feature",
			want:    []string{"gastown-x1y2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractBeadIDs(tt.message)
			if len(got) != len(tt.want) {
				t.Errorf("ExtractBeadIDs(%q) = %v, want %v", tt.message, got, tt.want)
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("ExtractBeadIDs(%q)[%d] = %q, want %q", tt.message, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestCommitsForBead(t *testing.T) {
	commits := []Commit{
		{SHA: "aaa", Subject: "fix aegis-abc1", BeadIDs: []string{"aegis-abc1"}},
		{SHA: "bbb", Subject: "unrelated commit", BeadIDs: nil},
		{SHA: "ccc", Subject: "aegis-abc1: followup", BeadIDs: []string{"aegis-abc1"}},
		{SHA: "ddd", Subject: "hq-xyz1 work", BeadIDs: []string{"hq-xyz1"}},
	}

	got := CommitsForBead(commits, "aegis-abc1")
	if len(got) != 2 {
		t.Fatalf("CommitsForBead got %d commits, want 2", len(got))
	}
	if got[0].SHA != "aaa" || got[1].SHA != "ccc" {
		t.Errorf("CommitsForBead returned wrong commits: %v", got)
	}
}

func TestBeadCommitMap(t *testing.T) {
	commits := []Commit{
		{SHA: "aaa", BeadIDs: []string{"aegis-abc1"}},
		{SHA: "bbb", BeadIDs: []string{"aegis-abc1", "hq-xyz1"}},
		{SHA: "ccc", BeadIDs: nil},
	}

	m := BeadCommitMap(commits)
	if len(m["aegis-abc1"]) != 2 {
		t.Errorf("aegis-abc1 got %d commits, want 2", len(m["aegis-abc1"]))
	}
	if len(m["hq-xyz1"]) != 1 {
		t.Errorf("hq-xyz1 got %d commits, want 1", len(m["hq-xyz1"]))
	}
}
