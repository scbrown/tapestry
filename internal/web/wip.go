package web

import (
	"context"
	"log"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/scbrown/tapestry/internal/dolt"
)

type wipAgent struct {
	Name       string
	InProgress int
	Blocked    int
	Open       int
	Total      int // in_progress + blocked + open (non-closed, non-deferred)
	OverLimit  bool
}

type wipData struct {
	GeneratedAt time.Time
	FilterRig   string

	Agents    []wipAgent
	WIPLimit  int // default threshold
	OverCount int // agents over the limit
	AvgWIP    float64
	MaxWIP    int
	SortBy    string

	// Status distribution
	TotalInProgress int
	TotalBlocked    int
	TotalOpen       int

	Err string
}

const defaultWIPLimit = 8

func (s *Server) handleWIP(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	filterRig := r.URL.Query().Get("rig")
	data := wipData{
		GeneratedAt: now,
		WIPLimit:    defaultWIPLimit,
		FilterRig:   filterRig,
	}

	if s.ds == nil {
		s.render(w, r, "wip", data)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("wip: list dbs: %v", err)
		data.Err = err.Error()
		s.render(w, r, "wip", data)
		return
	}

	agentWork := make(map[string]*wipAgent)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, db := range dbs {
		if filterRig != "" && db.Name != filterRig {
			continue
		}
		wg.Add(1)
		go func(dbName string) {
			defer wg.Done()

			issues, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{Limit: 2000})
			if err != nil {
				log.Printf("wip %s: %v", dbName, err)
				return
			}

			mu.Lock()
			for _, iss := range issues {
				if isNoise(iss.ID, iss.Title) || iss.Assignee == "" {
					continue
				}
				if iss.Status == "closed" || iss.Status == "deferred" {
					continue
				}

				a, ok := agentWork[iss.Assignee]
				if !ok {
					a = &wipAgent{Name: iss.Assignee}
					agentWork[iss.Assignee] = a
				}

				switch iss.Status {
				case "in_progress", "hooked":
					a.InProgress++
				case "blocked":
					a.Blocked++
				default:
					a.Open++
				}
				a.Total++
			}
			mu.Unlock()
		}(db.Name)
	}
	wg.Wait()

	// Convert to sorted slice
	for _, a := range agentWork {
		if a.Total > data.MaxWIP {
			data.MaxWIP = a.Total
		}
		if a.Total > defaultWIPLimit {
			a.OverLimit = true
			data.OverCount++
		}
		data.TotalInProgress += a.InProgress
		data.TotalBlocked += a.Blocked
		data.TotalOpen += a.Open
		data.Agents = append(data.Agents, *a)
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
	case "blocked":
		sort.Slice(data.Agents, func(i, j int) bool {
			if data.Agents[i].Blocked != data.Agents[j].Blocked {
				return data.Agents[i].Blocked > data.Agents[j].Blocked
			}
			return data.Agents[i].Total > data.Agents[j].Total
		})
	case "progress":
		sort.Slice(data.Agents, func(i, j int) bool {
			if data.Agents[i].InProgress != data.Agents[j].InProgress {
				return data.Agents[i].InProgress > data.Agents[j].InProgress
			}
			return data.Agents[i].Total > data.Agents[j].Total
		})
	default: // total
		sort.Slice(data.Agents, func(i, j int) bool {
			return data.Agents[i].Total > data.Agents[j].Total
		})
	}

	if len(data.Agents) > 0 {
		total := 0
		for _, a := range data.Agents {
			total += a.Total
		}
		data.AvgWIP = float64(total) / float64(len(data.Agents))
	}

	s.render(w, r, "wip", data)
}
