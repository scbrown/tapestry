package web

import (
	"log"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/scbrown/tapestry/internal/dolt"
)

type agentVelocityRow struct {
	Name    string
	Weeks   [4]int // closed counts per week (index 0 = most recent)
	Total   int
	Trend   string // "up", "down", "flat"
}

type agentVelocityData struct {
	GeneratedAt time.Time
	Agents      []agentVelocityRow
	WeekLabels  [4]string
	Rigs        []string
	FilterRig   string
	SortBy      string
}

func (s *Server) handleAgentVelocity(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	data := agentVelocityData{GeneratedAt: now}

	// Compute week labels (Mon-Sun boundaries)
	for i := 0; i < 4; i++ {
		weekStart := now.AddDate(0, 0, -7*(i+1))
		data.WeekLabels[i] = weekStart.Format("Jan 2")
	}

	if s.ds == nil {
		s.render(w, r, "agent-velocity", data)
		return
	}

	ctx := r.Context()
	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("agent-velocity: list dbs: %v", err)
		s.render(w, r, "agent-velocity", data)
		return
	}

	filterRig := r.URL.Query().Get("rig")
	var rigs []string
	for _, db := range dbs {
		rigs = append(rigs, db.Name)
	}
	sort.Strings(rigs)
	data.Rigs = rigs
	data.FilterRig = filterRig

	// For each DB, fetch closed issues in the last 4 weeks
	todayEnd := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, 1)
	fourWeeksAgo := todayEnd.AddDate(0, 0, -28)

	type dbResult struct {
		issues []dolt.Issue
	}
	results := make([]dbResult, len(dbs))
	var wg sync.WaitGroup
	for i, db := range dbs {
		if filterRig != "" && db.Name != filterRig {
			continue
		}
		wg.Add(1)
		go func(i int, dbName string) {
			defer wg.Done()
			closed, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{
				Status:        "closed",
				UpdatedAfter:  fourWeeksAgo,
				UpdatedBefore: todayEnd,
				Limit:         5000,
			})
			if err != nil {
				log.Printf("agent-velocity: %s closed: %v", dbName, err)
				return
			}
			results[i] = dbResult{issues: closed}
		}(i, db.Name)
	}
	wg.Wait()

	// Aggregate per-agent weekly closed counts
	agentWeeks := make(map[string][4]int)
	for _, r := range results {
		for _, iss := range r.issues {
			agent := iss.Assignee
			if agent == "" {
				agent = iss.Owner
			}
			if agent == "" || isNoise("", agent) {
				continue
			}
			// Determine which week bucket
			for w := 0; w < 4; w++ {
				weekEnd := todayEnd.AddDate(0, 0, -7*w)
				weekStart := weekEnd.AddDate(0, 0, -7)
				if !iss.UpdatedAt.Before(weekStart) && iss.UpdatedAt.Before(weekEnd) {
					counts := agentWeeks[agent]
					counts[w]++
					agentWeeks[agent] = counts
					break
				}
			}
		}
	}

	// Build rows
	for agent, weeks := range agentWeeks {
		total := weeks[0] + weeks[1] + weeks[2] + weeks[3]
		// Trend: compare first half (weeks 2,3) to second half (weeks 0,1)
		recent := weeks[0] + weeks[1]
		older := weeks[2] + weeks[3]
		trend := "flat"
		if recent > older {
			trend = "up"
		} else if recent < older {
			trend = "down"
		}
		data.Agents = append(data.Agents, agentVelocityRow{
			Name:  agent,
			Weeks: weeks,
			Total: total,
			Trend: trend,
		})
	}

	sortBy := r.URL.Query().Get("sort")
	if sortBy == "" {
		sortBy = "total"
	}
	data.SortBy = sortBy

	switch sortBy {
	case "name":
		sort.Slice(data.Agents, func(i, j int) bool {
			return data.Agents[i].Name < data.Agents[j].Name
		})
	case "trend":
		sort.Slice(data.Agents, func(i, j int) bool {
			rank := map[string]int{"up": 0, "flat": 1, "down": 2}
			if rank[data.Agents[i].Trend] != rank[data.Agents[j].Trend] {
				return rank[data.Agents[i].Trend] < rank[data.Agents[j].Trend]
			}
			return data.Agents[i].Total > data.Agents[j].Total
		})
	case "recent":
		sort.Slice(data.Agents, func(i, j int) bool {
			ri := data.Agents[i].Weeks[0] + data.Agents[i].Weeks[1]
			rj := data.Agents[j].Weeks[0] + data.Agents[j].Weeks[1]
			if ri != rj {
				return ri > rj
			}
			return data.Agents[i].Total > data.Agents[j].Total
		})
	default: // total
		sort.Slice(data.Agents, func(i, j int) bool {
			return data.Agents[i].Total > data.Agents[j].Total
		})
	}

	s.render(w, r, "agent-velocity", data)
}
