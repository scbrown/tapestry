package web

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

type commitsData struct {
	Commits      []commitRow
	RecentLinked []commitRow
	TotalCommits int
	LinkedCount  int
}

type commitRow struct {
	beadCommit
	BeadIDs []string
}

// beadIDPattern matches bead IDs like aegis-abc123, hq-xyz, tp-8px etc.
var beadIDPattern = regexp.MustCompile(`\b([a-z]+-[a-z0-9]{2,8})\b`)

func extractBeadIDs(msg string) []string {
	matches := beadIDPattern.FindAllString(msg, -1)
	// Deduplicate
	seen := make(map[string]bool, len(matches))
	var result []string
	for _, m := range matches {
		// Skip common false positives
		if m == "co-authored" || m == "pre-commit" || m == "no-verify" || m == "non-nil" {
			continue
		}
		if !seen[m] {
			seen[m] = true
			result = append(result, m)
		}
	}
	return result
}

func (s *Server) handleCommits(w http.ResponseWriter, r *http.Request) {
	data := commitsData{}

	if s.forgejo == nil {
		s.render(w, r, "commits", data)
		return
	}

	ctx := r.Context()

	var allCommits []commitRow
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, repo := range searchRepos {
		wg.Add(1)
		go func(repo string) {
			defer wg.Done()
			url := fmt.Sprintf("%s/api/v1/repos/%s/git/commits?sha=main&limit=50",
				s.forgejo.baseURL, repo)
			req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
			if err != nil {
				return
			}
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				log.Printf("commits: fetch %s: %v", repo, err)
				return
			}
			defer resp.Body.Close()
			if resp.StatusCode != 200 {
				return
			}
			var commits []forgejoCommit
			if err := json.NewDecoder(resp.Body).Decode(&commits); err != nil {
				return
			}

			parts := strings.Split(repo, "/")
			repoName := parts[len(parts)-1]

			mu.Lock()
			defer mu.Unlock()
			for _, c := range commits {
				subject := c.Commit.Message
				if idx := strings.IndexByte(subject, '\n'); idx > 0 {
					subject = subject[:idx]
				}
				shortSHA := c.SHA
				if len(shortSHA) > 7 {
					shortSHA = shortSHA[:7]
				}
				ts, _ := time.Parse(time.RFC3339, c.Commit.Author.Date)
				cr := commitRow{
					beadCommit: beadCommit{
						SHA:       c.SHA,
						ShortSHA:  shortSHA,
						Subject:   subject,
						Author:    c.Commit.Author.Name,
						Timestamp: ts,
						CommitURL: c.HTMLURL,
						RepoName:  repoName,
					},
					BeadIDs: extractBeadIDs(c.Commit.Message),
				}
				allCommits = append(allCommits, cr)
			}
		}(repo)
	}
	wg.Wait()

	sort.Slice(allCommits, func(i, j int) bool {
		return allCommits[i].Timestamp.After(allCommits[j].Timestamp)
	})

	data.TotalCommits = len(allCommits)
	for _, c := range allCommits {
		if len(c.BeadIDs) > 0 {
			data.RecentLinked = append(data.RecentLinked, c)
		}
	}
	data.LinkedCount = len(data.RecentLinked)

	// Cap lists
	if len(data.RecentLinked) > 30 {
		data.RecentLinked = data.RecentLinked[:30]
	}
	if len(allCommits) > 50 {
		allCommits = allCommits[:50]
	}
	data.Commits = allCommits

	s.render(w, r, "commits", data)
}
