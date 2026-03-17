package web

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/scbrown/tapestry/internal/dolt"
)

type agentWorkGroup struct {
	Agent  string
	Issues []dolt.Issue
}

type commandCenterData struct {
	GeneratedAt      time.Time
	FilterRig        string
	Rigs             []string
	InProgressCount  int
	OpenCount        int
	ClosedToday      int
	TotalBeads       int
	PendingDecisions int
	CriticalWork     []dolt.Issue
	AgentWork        []agentWorkGroup
	EpicProgress     []epicTree
	RecentClosed     []dolt.Issue
	RecentCommits    []beadCommit
}

func (s *Server) handleCommandCenter(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	todayEnd := todayStart.AddDate(0, 0, 1)
	yesterday := now.Add(-24 * time.Hour)

	filterRig := r.URL.Query().Get("rig")
	data := commandCenterData{GeneratedAt: now, FilterRig: filterRig}

	if s.ds == nil {
		s.render(w, r, "command-center", data)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 20*time.Second)
	defer cancel()

	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("command-center: list dbs: %v", err)
		s.render(w, r, "command-center", data)
		return
	}

	for _, db := range dbs {
		data.Rigs = append(data.Rigs, db.Name)
	}
	sort.Strings(data.Rigs)

	type ccDBResult struct {
		openCount       int
		inProgressCount int
		totalBeads      int
		closedToday     int
		decisions       int
		criticalWork    []dolt.Issue
		agentIssues     map[string][]dolt.Issue
		epics           []epicTree
		recentClosed    []dolt.Issue
	}

	results := make([]ccDBResult, len(dbs))
	var wg sync.WaitGroup
	for i, db := range dbs {
		if filterRig != "" && db.Name != filterRig {
			continue
		}
		wg.Add(1)
		go func(i int, dbName string) {
			defer wg.Done()
			r := ccDBResult{agentIssues: make(map[string][]dolt.Issue)}

			// Counts
			counts, err := s.ds.CountByStatus(ctx, dbName)
			if err != nil {
				log.Printf("command-center: counts %s: %v", dbName, err)
				results[i] = r
				return
			}
			r.openCount = counts["open"]
			r.inProgressCount = counts["in_progress"] + counts["hooked"]
			for _, v := range counts {
				r.totalBeads += v
			}

			closed, err := s.ds.CountClosedInRange(ctx, dbName, todayStart, todayEnd)
			if err == nil {
				r.closedToday = closed
			}

			// Decisions
			decisions, err := s.ds.Decisions(ctx, dbName)
			if err == nil {
				for _, d := range decisions {
					if d.Status != "closed" {
						r.decisions++
					}
				}
			}

			// Issues for critical work + agent grouping
			issues, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{Limit: 200})
			if err != nil {
				log.Printf("command-center: issues %s: %v", dbName, err)
				results[i] = r
				return
			}

			for _, iss := range issues {
				if iss.Status == "closed" || isNoise(iss.ID, iss.Title) {
					continue
				}

				// Critical work: P1 in_progress or open
				if iss.Priority <= 1 && (iss.Status == "in_progress" || iss.Status == "hooked" || iss.Status == "open") {
					iss.Rig = dbName
					r.criticalWork = append(r.criticalWork, iss)
				}

				// Agent grouping: in_progress/hooked issues
				if iss.Status == "in_progress" || iss.Status == "hooked" {
					agent := iss.Assignee
					if agent == "" {
						agent = iss.Owner
					}
					if agent != "" {
						iss.Rig = dbName
						r.agentIssues[agent] = append(r.agentIssues[agent], iss)
					}
				}
			}

			// Recently closed
			recent, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{
				Status:       "closed",
				UpdatedAfter: yesterday,
				Limit:        10,
			})
			if err == nil {
				for _, iss := range recent {
					if !isNoise(iss.ID, iss.Title) {
						iss.Rig = dbName
						r.recentClosed = append(r.recentClosed, iss)
					}
				}
			}

			// Epics with progress
			epics, err := s.ds.Epics(ctx, dbName)
			if err != nil {
				results[i] = r
				return
			}

			childDeps, _ := s.ds.AllChildDependencies(ctx, dbName)
			parentChildren := make(map[string][]string)
			for _, dep := range childDeps {
				parentChildren[dep.ToID] = append(parentChildren[dep.ToID], dep.FromID)
			}

			issueMap := make(map[string]dolt.Issue, len(issues))
			for _, iss := range issues {
				issueMap[iss.ID] = iss
			}

			for _, epic := range epics {
				if epic.Status == "closed" || isNoise(epic.ID, epic.Title) {
					continue
				}
				et := epicTree{Epic: epic, Rig: dbName}
				for _, childID := range parentChildren[epic.ID] {
					if child, ok := issueMap[childID]; ok {
						et.Progress.Total++
						if child.Status == "closed" {
							et.Progress.Closed++
						}
					}
				}
				if et.Progress.Total > 0 {
					r.epics = append(r.epics, et)
				}
			}

			results[i] = r
		}(i, db.Name)
	}
	wg.Wait()

	// Merge results
	agentMap := make(map[string][]dolt.Issue)
	for _, r := range results {
		data.OpenCount += r.openCount
		data.InProgressCount += r.inProgressCount
		data.TotalBeads += r.totalBeads
		data.ClosedToday += r.closedToday
		data.PendingDecisions += r.decisions
		data.CriticalWork = append(data.CriticalWork, r.criticalWork...)
		data.EpicProgress = append(data.EpicProgress, r.epics...)
		data.RecentClosed = append(data.RecentClosed, r.recentClosed...)
		for agent, issues := range r.agentIssues {
			key := shortActorName(agent)
			agentMap[key] = append(agentMap[key], issues...)
		}
	}

	// Sort critical work by priority then status
	sort.Slice(data.CriticalWork, func(i, j int) bool {
		statusOrder := map[string]int{"in_progress": 0, "hooked": 0, "open": 1}
		if statusOrder[data.CriticalWork[i].Status] != statusOrder[data.CriticalWork[j].Status] {
			return statusOrder[data.CriticalWork[i].Status] < statusOrder[data.CriticalWork[j].Status]
		}
		return data.CriticalWork[i].Priority < data.CriticalWork[j].Priority
	})
	if len(data.CriticalWork) > 15 {
		data.CriticalWork = data.CriticalWork[:15]
	}

	// Build agent work groups
	for agent, issues := range agentMap {
		sort.Slice(issues, func(i, j int) bool {
			return issues[i].Priority < issues[j].Priority
		})
		data.AgentWork = append(data.AgentWork, agentWorkGroup{Agent: agent, Issues: issues})
	}
	sort.Slice(data.AgentWork, func(i, j int) bool {
		return len(data.AgentWork[i].Issues) > len(data.AgentWork[j].Issues)
	})

	// Sort epics by priority
	sort.Slice(data.EpicProgress, func(i, j int) bool {
		if data.EpicProgress[i].Epic.Priority != data.EpicProgress[j].Epic.Priority {
			return data.EpicProgress[i].Epic.Priority < data.EpicProgress[j].Epic.Priority
		}
		return data.EpicProgress[i].Epic.UpdatedAt.After(data.EpicProgress[j].Epic.UpdatedAt)
	})
	if len(data.EpicProgress) > 10 {
		data.EpicProgress = data.EpicProgress[:10]
	}

	// Sort recent closed by time
	sort.Slice(data.RecentClosed, func(i, j int) bool {
		return data.RecentClosed[i].UpdatedAt.After(data.RecentClosed[j].UpdatedAt)
	})
	if len(data.RecentClosed) > 10 {
		data.RecentClosed = data.RecentClosed[:10]
	}

	// Fetch recent commits from Forgejo
	if s.forgejo != nil {
		data.RecentCommits = s.fetchRecentCommits(ctx, 10)
	}

	s.render(w, r, "command-center", data)
}

// fetchRecentCommits gets the most recent commits across all repos.
func (s *Server) fetchRecentCommits(ctx context.Context, limit int) []beadCommit {
	var allCommits []beadCommit
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, repo := range searchRepos {
		wg.Add(1)
		go func(repo string) {
			defer wg.Done()
			url := fmt.Sprintf("%s/api/v1/repos/%s/git/commits?sha=main&limit=10",
				s.forgejo.baseURL, repo)
			req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
			if err != nil {
				return
			}
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
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
				allCommits = append(allCommits, beadCommit{
					SHA:       c.SHA,
					ShortSHA:  shortSHA,
					Subject:   subject,
					Author:    c.Commit.Author.Name,
					Timestamp: ts,
					CommitURL: c.HTMLURL,
					RepoName:  repoName,
				})
			}
		}(repo)
	}
	wg.Wait()

	sort.Slice(allCommits, func(i, j int) bool {
		return allCommits[i].Timestamp.After(allCommits[j].Timestamp)
	})

	if len(allCommits) > limit {
		allCommits = allCommits[:limit]
	}
	return allCommits
}
