package web

import (
	"context"
	"log"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/scbrown/tapestry/internal/dolt"
	"github.com/scbrown/tapestry/internal/events"
)

type agentDetailData struct {
	Name           string
	Stats          dolt.AgentStats
	Issues         []dolt.Issue
	Assignees      []string
	HandoffStats   *events.ChainStats
	RecentHandoffs []events.HandoffEvent
}

func (s *Server) handleAgentDetail(w http.ResponseWriter, r *http.Request, name string) {
	data := agentDetailData{Name: name}

	if s.ds == nil {
		s.render(w, r, "agent", data)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("agent: list dbs: %v", err)
		s.render(w, r, "agent", data)
		return
	}

	type dbResult struct {
		issues    []dolt.Issue
		stats     dolt.AgentStats
		assignees []string
	}

	results := make([]dbResult, len(dbs))
	var wg sync.WaitGroup
	for i, db := range dbs {
		wg.Add(1)
		go func(i int, dbName string) {
			defer wg.Done()
			var r dbResult
			r.assignees, _ = s.ds.DistinctAssignees(ctx, dbName)

			// Search by assignee (primary) and owner (fallback)
			for _, field := range []string{"assignee", "owner"} {
				var filter dolt.IssueFilter
				if field == "assignee" {
					filter.Assignee = name
				} else {
					filter.Owner = name
				}
				filter.Limit = 200
				issues, err := s.ds.Issues(ctx, dbName, filter)
				if err != nil {
					continue
				}
				seen := make(map[string]bool)
				for _, iss := range r.issues {
					seen[iss.ID] = true
				}
				for _, iss := range issues {
					if seen[iss.ID] {
						continue
					}
					if isNoise(iss.ID, iss.Title) {
						continue
					}
					iss.Rig = dbName
					r.issues = append(r.issues, iss)
				}
			}

			// Compute stats
			for _, iss := range r.issues {
				r.stats.Owned++
				switch iss.Status {
				case "open":
					r.stats.Open++
				case "in_progress", "hooked":
					r.stats.InProgress++
				case "closed":
					r.stats.Closed++
				}
			}

			results[i] = r
		}(i, db.Name)
	}
	wg.Wait()

	assigneeSet := make(map[string]bool)
	for _, r := range results {
		data.Issues = append(data.Issues, r.issues...)
		data.Stats.Owned += r.stats.Owned
		data.Stats.Open += r.stats.Open
		data.Stats.InProgress += r.stats.InProgress
		data.Stats.Closed += r.stats.Closed
		for _, a := range r.assignees {
			assigneeSet[a] = true
		}
	}
	for a := range assigneeSet {
		data.Assignees = append(data.Assignees, a)
	}
	sort.Strings(data.Assignees)

	// Also try matching by short name (e.g., "arnold" matches "aegis/crew/arnold")
	if !strings.Contains(name, "/") && len(data.Issues) == 0 {
		for i, db := range dbs {
			wg.Add(1)
			go func(i int, dbName string) {
				defer wg.Done()
				allIssues, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{Limit: 500})
				if err != nil {
					return
				}
				var matched []dolt.Issue
				for _, iss := range allIssues {
					if isNoise(iss.ID, iss.Title) {
						continue
					}
					assigneeShort := shortActorName(iss.Assignee)
					ownerShort := shortActorName(iss.Owner)
					if assigneeShort == name || ownerShort == name {
						iss.Rig = dbName
						matched = append(matched, iss)
					}
				}
				results[i] = dbResult{issues: matched}
			}(i, db.Name)
		}
		wg.Wait()
		for _, r := range results {
			data.Issues = append(data.Issues, r.issues...)
		}
		for _, iss := range data.Issues {
			data.Stats.Owned++
			switch iss.Status {
			case "open":
				data.Stats.Open++
			case "in_progress", "hooked":
				data.Stats.InProgress++
			case "closed":
				data.Stats.Closed++
			}
		}
	}

	sort.Slice(data.Issues, func(i, j int) bool {
		if data.Issues[i].Status != data.Issues[j].Status {
			statusOrder := map[string]int{"in_progress": 0, "hooked": 0, "open": 1, "blocked": 2, "closed": 3}
			return statusOrder[data.Issues[i].Status] < statusOrder[data.Issues[j].Status]
		}
		if data.Issues[i].Priority != data.Issues[j].Priority {
			return data.Issues[i].Priority < data.Issues[j].Priority
		}
		return data.Issues[i].UpdatedAt.After(data.Issues[j].UpdatedAt)
	})

	data.Stats.Name = name

	// Enrich with handoff data from events
	if s.workspacePath != "" {
		allEvents, err := events.ReadWorkspace(s.workspacePath)
		if err == nil {
			chains := events.BuildHandoffChains(allEvents)
			for _, chain := range chains {
				if strings.Contains(chain.Actor, name) || (len(name) > 0 && strings.HasSuffix(chain.Actor, "/"+name)) {
					summary := events.ChainSummary([]events.HandoffChain{chain})
					if len(summary) > 0 {
						data.HandoffStats = &summary[0]
					}
					// Show last 10 handoffs
					end := len(chain.Handoffs)
					start := end - 10
					if start < 0 {
						start = 0
					}
					data.RecentHandoffs = chain.Handoffs[start:end]
					break
				}
			}
		}
	}

	s.render(w, r, "agent", data)
}
