package web

import (
	"log"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/scbrown/tapestry/internal/dolt"
)

type rigStats struct {
	Name     string
	Open     int
	Active   int
	Blocked  int
	Closed   int
	Deferred int
	Total    int
}

type statsData struct {
	GeneratedAt time.Time
	FilterRig   string
	AllRigs     []string
	Rigs        []rigStats
	TotalOpen   int
	TotalActive int
	TotalBlocked int
	TotalClosed int
	TotalDeferred int
	TotalBeads  int
	Created7d   int
	Closed7d    int
	Created30d  int
	Closed30d   int
	NetFlow7d   int
	NetFlow30d  int
	AgentCount  int
	SortBy      string
	Err         string
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	filterRig := r.URL.Query().Get("rig")
	data := statsData{GeneratedAt: time.Now(), FilterRig: filterRig}

	if s.ds == nil {
		s.render(w, r, "stats", data)
		return
	}

	ctx := r.Context()
	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("stats: list dbs: %v", err)
		s.render(w, r, "stats", data)
		return
	}

	for _, db := range dbs {
		data.AllRigs = append(data.AllRigs, db.Name)
	}
	sort.Strings(data.AllRigs)

	now := time.Now()
	sevenDaysAgo := now.AddDate(0, 0, -7)
	thirtyDaysAgo := now.AddDate(0, 0, -30)
	todayEnd := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.UTC)

	type dbResult struct {
		name        string
		counts      map[string]int
		created7d   int
		closed7d    int
		created30d  int
		closed30d   int
		agents      map[string]bool
	}
	results := make([]dbResult, len(dbs))
	var wg sync.WaitGroup

	for i, db := range dbs {
		if filterRig != "" && db.Name != filterRig {
			continue
		}
		wg.Add(1)
		go func(idx int, dbName string) {
			defer wg.Done()
			r := dbResult{name: dbName, agents: map[string]bool{}}

			counts, err := s.ds.CountByStatus(ctx, dbName)
			if err != nil {
				log.Printf("stats: counts %s: %v", dbName, err)
				return
			}
			r.counts = counts

			if c, err := s.ds.CountCreatedInRange(ctx, dbName, sevenDaysAgo, todayEnd); err == nil {
				r.created7d = c
			}
			if c, err := s.ds.CountClosedInRange(ctx, dbName, sevenDaysAgo, todayEnd); err == nil {
				r.closed7d = c
			}
			if c, err := s.ds.CountCreatedInRange(ctx, dbName, thirtyDaysAgo, todayEnd); err == nil {
				r.created30d = c
			}
			if c, err := s.ds.CountClosedInRange(ctx, dbName, thirtyDaysAgo, todayEnd); err == nil {
				r.closed30d = c
			}

			// Count unique agents
			issues, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{Limit: 5000})
			if err == nil {
				for _, iss := range issues {
					if iss.Assignee != "" {
						r.agents[iss.Assignee] = true
					}
					if iss.Owner != "" {
						r.agents[iss.Owner] = true
					}
				}
			}

			results[idx] = r
		}(i, db.Name)
	}
	wg.Wait()

	allAgents := map[string]bool{}
	for _, r := range results {
		if r.counts == nil {
			continue
		}

		open := r.counts["open"]
		active := r.counts["in_progress"] + r.counts["hooked"]
		blocked := r.counts["blocked"]
		closed := r.counts["closed"] + r.counts["completed"]
		deferred := r.counts["deferred"]
		total := 0
		for _, v := range r.counts {
			total += v
		}

		data.Rigs = append(data.Rigs, rigStats{
			Name:     r.name,
			Open:     open,
			Active:   active,
			Blocked:  blocked,
			Closed:   closed,
			Deferred: deferred,
			Total:    total,
		})

		data.TotalOpen += open
		data.TotalActive += active
		data.TotalBlocked += blocked
		data.TotalClosed += closed
		data.TotalDeferred += deferred
		data.TotalBeads += total
		data.Created7d += r.created7d
		data.Closed7d += r.closed7d
		data.Created30d += r.created30d
		data.Closed30d += r.closed30d

		for a := range r.agents {
			allAgents[a] = true
		}
	}

	data.NetFlow7d = data.Created7d - data.Closed7d
	data.NetFlow30d = data.Created30d - data.Closed30d
	data.AgentCount = len(allAgents)

	sortBy := r.URL.Query().Get("sort")
	if sortBy == "" {
		sortBy = "total"
	}
	data.SortBy = sortBy

	switch sortBy {
	case "name":
		sort.Slice(data.Rigs, func(i, j int) bool {
			return data.Rigs[i].Name < data.Rigs[j].Name
		})
	case "open":
		sort.Slice(data.Rigs, func(i, j int) bool {
			if data.Rigs[i].Open != data.Rigs[j].Open {
				return data.Rigs[i].Open > data.Rigs[j].Open
			}
			return data.Rigs[i].Total > data.Rigs[j].Total
		})
	case "active":
		sort.Slice(data.Rigs, func(i, j int) bool {
			if data.Rigs[i].Active != data.Rigs[j].Active {
				return data.Rigs[i].Active > data.Rigs[j].Active
			}
			return data.Rigs[i].Total > data.Rigs[j].Total
		})
	default: // total
		sort.Slice(data.Rigs, func(i, j int) bool {
			return data.Rigs[i].Total > data.Rigs[j].Total
		})
	}

	s.render(w, r, "stats", data)
}
