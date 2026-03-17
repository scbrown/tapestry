package web

import (
	"log"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/scbrown/tapestry/internal/dolt"
)

type standupAgent struct {
	Name         string
	ClosedYest   []standupEntry // closed yesterday
	StartedYest  []standupEntry // started/progressed yesterday
	InProgress   []standupEntry // currently in_progress/hooked
	Blocked      []standupEntry // currently blocked
	ClosedCount  int
	ActiveCount  int
	BlockedCount int
}

type standupEntry struct {
	Issue dolt.Issue
	Rig   string
}

type standupData struct {
	GeneratedAt time.Time
	Agents      []standupAgent
	Unassigned  []standupEntry // in_progress/blocked with no assignee
	TotalClosed int
	TotalActive int
	TotalBlocked int
	YestLabel   string
	Rigs        []string
	FilterRig   string
	Err         string
}

func (s *Server) handleStandup(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	yesterdayStart := todayStart.AddDate(0, 0, -1)

	data := standupData{
		GeneratedAt: now,
		YestLabel:   yesterdayStart.Format("Mon, Jan 2"),
	}

	if s.ds == nil {
		s.render(w, r, "standup", data)
		return
	}

	ctx := r.Context()
	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("standup: list dbs: %v", err)
		s.render(w, r, "standup", standupData{Err: err.Error(), GeneratedAt: now, YestLabel: data.YestLabel})
		return
	}

	type dbResult struct {
		closedYest  []standupEntry
		startedYest []standupEntry
		inProgress  []standupEntry
		blocked     []standupEntry
	}

	results := make([]dbResult, len(dbs))
	var wg sync.WaitGroup
	for i, db := range dbs {
		wg.Add(1)
		go func(i int, dbName string) {
			defer wg.Done()
			var r dbResult

			// Fetch recently closed (yesterday)
			closedIssues, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{
				Status:       "closed",
				UpdatedAfter: yesterdayStart,
				Limit:        200,
			})
			if err == nil {
				for _, iss := range closedIssues {
					if isNoise(iss.ID, iss.Title) {
						continue
					}
					if !iss.UpdatedAt.Before(yesterdayStart) && iss.UpdatedAt.Before(todayStart) {
						r.closedYest = append(r.closedYest, standupEntry{Issue: iss, Rig: dbName})
					}
				}
			}

			// Fetch in-progress items (current)
			inprogIssues, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{Limit: 300})
			if err == nil {
				for _, iss := range inprogIssues {
					if isNoise(iss.ID, iss.Title) {
						continue
					}
					switch iss.Status {
					case "in_progress", "hooked":
						r.inProgress = append(r.inProgress, standupEntry{Issue: iss, Rig: dbName})
						// If started yesterday (updated yesterday, wasn't closed)
						if !iss.UpdatedAt.Before(yesterdayStart) && iss.UpdatedAt.Before(todayStart) {
							r.startedYest = append(r.startedYest, standupEntry{Issue: iss, Rig: dbName})
						}
					case "blocked":
						r.blocked = append(r.blocked, standupEntry{Issue: iss, Rig: dbName})
					}
				}
			}

			results[i] = r
		}(i, db.Name)
	}
	wg.Wait()

	// Aggregate by agent
	agentMap := make(map[string]*standupAgent)
	rigSet := make(map[string]bool)

	getAgent := func(name string) *standupAgent {
		if name == "" {
			return nil
		}
		if a, ok := agentMap[name]; ok {
			return a
		}
		a := &standupAgent{Name: name}
		agentMap[name] = a
		return a
	}

	var unassigned []standupEntry

	for _, r := range results {
		for _, e := range r.closedYest {
			rigSet[e.Rig] = true
			agent := e.Issue.Assignee
			if agent == "" {
				agent = e.Issue.Owner
			}
			if a := getAgent(agent); a != nil {
				a.ClosedYest = append(a.ClosedYest, e)
				a.ClosedCount++
			}
			data.TotalClosed++
		}
		for _, e := range r.startedYest {
			rigSet[e.Rig] = true
			agent := e.Issue.Assignee
			if agent == "" {
				agent = e.Issue.Owner
			}
			if a := getAgent(agent); a != nil {
				a.StartedYest = append(a.StartedYest, e)
			}
		}
		for _, e := range r.inProgress {
			rigSet[e.Rig] = true
			agent := e.Issue.Assignee
			if agent == "" {
				agent = e.Issue.Owner
			}
			if a := getAgent(agent); a != nil {
				a.InProgress = append(a.InProgress, e)
				a.ActiveCount++
			} else {
				unassigned = append(unassigned, e)
			}
			data.TotalActive++
		}
		for _, e := range r.blocked {
			rigSet[e.Rig] = true
			agent := e.Issue.Assignee
			if agent == "" {
				agent = e.Issue.Owner
			}
			if a := getAgent(agent); a != nil {
				a.Blocked = append(a.Blocked, e)
				a.BlockedCount++
			} else {
				unassigned = append(unassigned, e)
			}
			data.TotalBlocked++
		}
	}

	// Build sorted rig list
	for rig := range rigSet {
		data.Rigs = append(data.Rigs, rig)
	}
	sort.Strings(data.Rigs)

	// Apply rig filter
	filterRig := r.URL.Query().Get("rig")
	data.FilterRig = filterRig

	// Convert agent map to sorted slice
	for _, a := range agentMap {
		// Sort each agent's lists by priority
		sort.Slice(a.ClosedYest, func(i, j int) bool {
			return a.ClosedYest[i].Issue.Priority < a.ClosedYest[j].Issue.Priority
		})
		sort.Slice(a.InProgress, func(i, j int) bool {
			return a.InProgress[i].Issue.Priority < a.InProgress[j].Issue.Priority
		})
		sort.Slice(a.Blocked, func(i, j int) bool {
			return a.Blocked[i].Issue.Priority < a.Blocked[j].Issue.Priority
		})

		if filterRig != "" {
			a.ClosedYest = filterByRig(a.ClosedYest, filterRig)
			a.StartedYest = filterByRig(a.StartedYest, filterRig)
			a.InProgress = filterByRig(a.InProgress, filterRig)
			a.Blocked = filterByRig(a.Blocked, filterRig)
			a.ClosedCount = len(a.ClosedYest)
			a.ActiveCount = len(a.InProgress)
			a.BlockedCount = len(a.Blocked)
		}

		// Only include agents that have some activity
		if a.ClosedCount > 0 || a.ActiveCount > 0 || a.BlockedCount > 0 || len(a.StartedYest) > 0 {
			data.Agents = append(data.Agents, *a)
		}
	}

	// Sort agents: most active first (closed + active + blocked)
	sort.Slice(data.Agents, func(i, j int) bool {
		si := data.Agents[i].ClosedCount + data.Agents[i].ActiveCount
		sj := data.Agents[j].ClosedCount + data.Agents[j].ActiveCount
		return si > sj
	})

	if filterRig != "" {
		unassigned = filterByRig(unassigned, filterRig)
	}
	data.Unassigned = unassigned

	s.render(w, r, "standup", data)
}

func filterByRig(entries []standupEntry, rig string) []standupEntry {
	var out []standupEntry
	for _, e := range entries {
		if e.Rig == rig {
			out = append(out, e)
		}
	}
	return out
}
