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

func TestParseRemoteURL(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{"ssh colon", "git@git.svc:stiwi/aegis.git", "https://git.svc/stiwi/aegis"},
		{"ssh colon no .git", "git@git.svc:stiwi/aegis", "https://git.svc/stiwi/aegis"},
		{"github ssh", "git@github.com:scbrown/bobbin.git", "https://github.com/scbrown/bobbin"},
		{"https", "https://git.svc/stiwi/goldblum.git", "https://git.svc/stiwi/goldblum"},
		{"https no .git", "https://github.com/scbrown/tapestry", "https://github.com/scbrown/tapestry"},
		{"ssh proto", "ssh://git@git.svc/stiwi/aegis.git", "https://git.svc/stiwi/aegis"},
		{"http", "http://git.svc/stiwi/aegis.git", "http://git.svc/stiwi/aegis"},
		{"empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseRemoteURL(tt.raw)
			if got != tt.want {
				t.Errorf("parseRemoteURL(%q) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}

func TestSetCommitURLs(t *testing.T) {
	commits := []Commit{
		{SHA: "abc123", RepoName: "aegis"},
		{SHA: "def456", RepoName: "bobbin"},
		{SHA: "ghi789", RepoName: "unknown"},
	}
	urls := map[string]string{
		"aegis":  "https://git.svc/stiwi/aegis",
		"bobbin": "https://github.com/scbrown/bobbin",
	}
	SetCommitURLs(commits, urls)
	if commits[0].CommitURL != "https://git.svc/stiwi/aegis/commit/abc123" {
		t.Errorf("aegis commit URL = %q", commits[0].CommitURL)
	}
	if commits[1].CommitURL != "https://github.com/scbrown/bobbin/commit/def456" {
		t.Errorf("bobbin commit URL = %q", commits[1].CommitURL)
	}
	if commits[2].CommitURL != "" {
		t.Errorf("unknown commit URL should be empty, got %q", commits[2].CommitURL)
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
