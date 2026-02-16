// Package git provides commit correlation for tapestry.
// It parses git log output and extracts bead IDs from commit messages.
package git

import (
	"bufio"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// Commit represents a parsed git commit.
type Commit struct {
	SHA       string
	ShortSHA  string
	Author    string
	Timestamp time.Time
	Subject   string
	BeadIDs   []string // extracted bead IDs from message
	RepoName  string   // basename of the repo
	CommitURL string   // full URL to view commit on git host
}

// beadIDPattern matches bead IDs in commit messages.
// Beads use prefix-hash format like: aegis-9i6i, hq-c27u7, bobbin-abc1
// The prefix is lowercase alpha, the hash is lowercase alphanumeric.
var beadIDPattern = regexp.MustCompile(`\b([a-z]+-[a-z0-9]{3,8})\b`)

// knownPrefixes filters matches to known bead prefixes to avoid false positives.
// Add prefixes as new rigs are created.
var knownPrefixes = []string{
	"aegis-", "hq-", "bobbin-", "gastown-", "gt-", "beads-",
}

// ExtractBeadIDs finds bead IDs in a commit message.
func ExtractBeadIDs(message string) []string {
	matches := beadIDPattern.FindAllString(message, -1)
	var ids []string
	seen := make(map[string]bool)
	for _, m := range matches {
		if seen[m] {
			continue
		}
		for _, prefix := range knownPrefixes {
			if strings.HasPrefix(m, prefix) {
				ids = append(ids, m)
				seen[m] = true
				break
			}
		}
	}
	return ids
}

// ParseLog runs git log in the given repo directory and returns parsed commits.
// It uses a custom format for reliable parsing.
func ParseLog(repoDir string, limit int) ([]Commit, error) {
	if limit <= 0 {
		limit = 200
	}

	// Use canonical repo name from remote when available (handles worktrees)
	repoName := repoNameFromRemote(repoDir)
	baseURL := remoteBaseURL(repoDir)

	// Use NUL-delimited format for reliable parsing
	// Fields: SHA, short SHA, author, timestamp (unix), subject
	format := "%H%x00%h%x00%an%x00%at%x00%s"
	cmd := exec.Command("git", "log",
		fmt.Sprintf("--max-count=%d", limit),
		fmt.Sprintf("--format=%s", format),
	)
	cmd.Dir = repoDir

	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git log in %s: %w", repoDir, err)
	}

	var commits []Commit
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\x00", 5)
		if len(parts) < 5 {
			continue
		}

		var ts time.Time
		if unix, err := parseUnix(parts[3]); err == nil {
			ts = unix
		}

		var commitURL string
		if baseURL != "" {
			commitURL = baseURL + "/commit/" + parts[0]
		}

		c := Commit{
			SHA:       parts[0],
			ShortSHA:  parts[1],
			Author:    parts[2],
			Timestamp: ts,
			Subject:   parts[4],
			RepoName:  repoName,
			CommitURL: commitURL,
		}
		c.BeadIDs = ExtractBeadIDs(c.Subject)
		commits = append(commits, c)
	}

	return commits, nil
}

func parseUnix(s string) (time.Time, error) {
	var sec int64
	for _, c := range s {
		if c < '0' || c > '9' {
			return time.Time{}, fmt.Errorf("not a number: %s", s)
		}
		sec = sec*10 + int64(c-'0')
	}
	return time.Unix(sec, 0), nil
}

// ParseWorkspace scans a Gas Town workspace for git repos and parses their logs.
// It looks for repos in the workspace root and common subdirectories.
func ParseWorkspace(wsPath string, limit int) ([]Commit, error) {
	var allCommits []Commit

	// Check if the workspace root itself is a git repo
	if isGitRepo(wsPath) {
		commits, err := ParseLog(wsPath, limit)
		if err == nil {
			allCommits = append(allCommits, commits...)
		}
	}

	// Scan immediate subdirectories for git repos (rig directories, etc.)
	dirs := []string{"crew", "polecats", "refinery"}
	for _, sub := range dirs {
		subPath := filepath.Join(wsPath, sub)
		entries, err := filepath.Glob(filepath.Join(subPath, "*", ".git"))
		if err != nil {
			continue
		}
		for _, gitDir := range entries {
			repoDir := filepath.Dir(gitDir)
			commits, err := ParseLog(repoDir, limit)
			if err != nil {
				continue
			}
			allCommits = append(allCommits, commits...)
		}
	}

	// Sort all commits by timestamp (newest first)
	sort.Slice(allCommits, func(i, j int) bool {
		return allCommits[i].Timestamp.After(allCommits[j].Timestamp)
	})

	return allCommits, nil
}

func isGitRepo(dir string) bool {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	cmd.Dir = dir
	out, err := cmd.Output()
	return err == nil && strings.TrimSpace(string(out)) == "true"
}

// remoteBaseURL extracts a browsable base URL from git remote origin.
// Handles SSH (git@host:owner/repo.git) and HTTPS (https://host/owner/repo.git).
func remoteBaseURL(repoDir string) string {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	cmd.Dir = repoDir
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	raw := strings.TrimSpace(string(out))
	return parseRemoteURL(raw)
}

// parseRemoteURL converts a git remote URL to a browsable HTTPS base URL.
func parseRemoteURL(raw string) string {
	raw = strings.TrimSuffix(raw, ".git")

	// SSH: git@host:owner/repo
	if strings.HasPrefix(raw, "git@") {
		raw = strings.TrimPrefix(raw, "git@")
		// git@host:owner/repo -> host/owner/repo
		raw = strings.Replace(raw, ":", "/", 1)
		return "https://" + raw
	}

	// SSH: ssh://git@host/owner/repo
	if strings.HasPrefix(raw, "ssh://") {
		raw = strings.TrimPrefix(raw, "ssh://")
		raw = strings.TrimPrefix(raw, "git@")
		return "https://" + raw
	}

	// HTTPS: already a URL
	if strings.HasPrefix(raw, "https://") || strings.HasPrefix(raw, "http://") {
		return raw
	}

	return ""
}

// repoNameFromRemote extracts the repository name from a git remote URL.
func repoNameFromRemote(repoDir string) string {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	cmd.Dir = repoDir
	out, err := cmd.Output()
	if err != nil {
		return filepath.Base(repoDir)
	}
	raw := strings.TrimSpace(string(out))
	raw = strings.TrimSuffix(raw, ".git")
	parts := strings.Split(raw, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return filepath.Base(repoDir)
}

// SetCommitURLs populates CommitURL on commits using a repo name → base URL map.
// The map can come from config overrides or auto-detected remotes.
func SetCommitURLs(commits []Commit, repoURLs map[string]string) {
	for i := range commits {
		if base, ok := repoURLs[commits[i].RepoName]; ok {
			commits[i].CommitURL = base + "/commit/" + commits[i].SHA
		}
	}
}

// CommitsForBead filters commits to only those referencing a specific bead ID.
func CommitsForBead(commits []Commit, beadID string) []Commit {
	var matched []Commit
	for _, c := range commits {
		for _, id := range c.BeadIDs {
			if id == beadID {
				matched = append(matched, c)
				break
			}
		}
	}
	return matched
}

// BeadCommitMap builds a map from bead ID to commits.
func BeadCommitMap(commits []Commit) map[string][]Commit {
	m := make(map[string][]Commit)
	for _, c := range commits {
		for _, id := range c.BeadIDs {
			m[id] = append(m[id], c)
		}
	}
	return m
}

// RecentWithBeads returns only commits that reference at least one bead ID.
func RecentWithBeads(commits []Commit, limit int) []Commit {
	var result []Commit
	for _, c := range commits {
		if len(c.BeadIDs) > 0 {
			result = append(result, c)
			if limit > 0 && len(result) >= limit {
				break
			}
		}
	}
	return result
}
