package web

import (
	"log"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/scbrown/tapestry/internal/dolt"
)

type agentLoad struct {
	Name       string
	Open       int
	InProgress int
	Blocked    int
	Deferred   int
	Total      int // open + in_progress + blocked
	HighPri    int // P0 + P1
}

type workloadData struct {
	GeneratedAt time.Time
	Agents      []agentLoad
	TotalWork   int
	AvgLoad     float64
	MaxLoad     int
	MinLoad     int
}

func (s *Server) handleWorkload(w http.ResponseWriter, r *http.Request) {
	data := workloadData{GeneratedAt: time.Now()}

	if s.ds == nil {
		s.render(w, r, "workload", data)
		return
	}

	ctx := r.Context()
	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("workload: list dbs: %v", err)
		s.render(w, r, "workload", data)
		return
	}

	type dbResult struct {
		issues []dolt.Issue
	}
	results := make([]dbResult, len(dbs))
	var wg sync.WaitGroup
	for i, db := range dbs {
		wg.Add(1)
		go func(i int, dbName string) {
			defer wg.Done()
			issues, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{Limit: 10000})
			if err != nil {
				log.Printf("workload: %s: %v", dbName, err)
				return
			}
			results[i] = dbResult{issues: issues}
		}(i, db.Name)
	}
	wg.Wait()

	agentMap := map[string]*agentLoad{}

	for _, r := range results {
		for _, iss := range r.issues {
			// Skip closed items for workload
			if iss.Status == "closed" {
				continue
			}

			assignee := iss.Assignee
			if assignee == "" {
				assignee = iss.Owner
			}
			if assignee == "" {
				continue // unassigned items don't count toward anyone's load
			}

			al, ok := agentMap[assignee]
			if !ok {
				al = &agentLoad{Name: assignee}
				agentMap[assignee] = al
			}

			switch iss.Status {
			case "open":
				al.Open++
			case "in_progress", "hooked":
				al.InProgress++
			case "blocked":
				al.Blocked++
			case "deferred":
				al.Deferred++
			}

			if iss.Status != "deferred" {
				al.Total++
				if iss.Priority <= 1 {
					al.HighPri++
				}
			}
		}
	}

	for _, al := range agentMap {
		data.Agents = append(data.Agents, *al)
		data.TotalWork += al.Total
	}

	// Sort by total workload descending
	sort.Slice(data.Agents, func(i, j int) bool {
		return data.Agents[i].Total > data.Agents[j].Total
	})

	if len(data.Agents) > 0 {
		data.AvgLoad = float64(data.TotalWork) / float64(len(data.Agents))
		data.MaxLoad = data.Agents[0].Total
		data.MinLoad = data.Agents[len(data.Agents)-1].Total
	}

	s.render(w, r, "workload", data)
}
