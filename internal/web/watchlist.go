package web

import (
	"log"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/scbrown/tapestry/internal/dolt"
)

type watchlistItem struct {
	Issue   dolt.Issue
	Rig     string
	AgeDays int
	IdleH   int // hours since last update
}

type watchlistData struct {
	GeneratedAt time.Time
	P0          []watchlistItem
	P1          []watchlistItem
	TotalP0     int
	TotalP1     int
	Rigs        []string
	FilterRig   string
	Assignees   []string
}

func (s *Server) handleWatchlist(w http.ResponseWriter, r *http.Request) {
	data := watchlistData{GeneratedAt: time.Now()}

	if s.ds == nil {
		s.render(w, r, "watchlist", data)
		return
	}

	ctx := r.Context()
	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("watchlist: list dbs: %v", err)
		s.render(w, r, "watchlist", data)
		return
	}

	type dbResult struct {
		issues    []dolt.Issue
		assignees []string
	}
	results := make([]dbResult, len(dbs))
	var wg sync.WaitGroup
	for i, db := range dbs {
		wg.Add(1)
		go func(i int, dbName string) {
			defer wg.Done()
			issues, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{Limit: 2000})
			if err != nil {
				log.Printf("watchlist: %s: %v", dbName, err)
				return
			}
			assignees, _ := s.ds.DistinctAssignees(ctx, dbName)
			results[i] = dbResult{issues: issues, assignees: assignees}
		}(i, db.Name)
	}
	wg.Wait()

	now := time.Now()
	for idx, r := range results {
		rig := dbs[idx].Name
		for _, iss := range r.issues {
			if iss.Status == "closed" || iss.Status == "deferred" {
				continue
			}
			if iss.Priority > 1 {
				continue
			}

			ageDays := int(now.Sub(iss.CreatedAt).Hours() / 24)
			idleH := int(now.Sub(iss.UpdatedAt).Hours())
			if ageDays < 0 {
				ageDays = 0
			}
			if idleH < 0 {
				idleH = 0
			}

			item := watchlistItem{Issue: iss, Rig: rig, AgeDays: ageDays, IdleH: idleH}
			if iss.Priority == 0 {
				data.P0 = append(data.P0, item)
			} else {
				data.P1 = append(data.P1, item)
			}
		}
	}

	// Collect distinct rigs for filter
	rigSet := make(map[string]bool)
	for _, item := range data.P0 {
		rigSet[item.Rig] = true
	}
	for _, item := range data.P1 {
		rigSet[item.Rig] = true
	}
	var rigs []string
	for rig := range rigSet {
		rigs = append(rigs, rig)
	}
	sort.Strings(rigs)
	data.Rigs = rigs

	// Apply rig filter
	filterRig := r.URL.Query().Get("rig")
	data.FilterRig = filterRig
	if filterRig != "" {
		filtered := data.P0[:0]
		for _, item := range data.P0 {
			if item.Rig == filterRig {
				filtered = append(filtered, item)
			}
		}
		data.P0 = filtered

		filteredP1 := data.P1[:0]
		for _, item := range data.P1 {
			if item.Rig == filterRig {
				filteredP1 = append(filteredP1, item)
			}
		}
		data.P1 = filteredP1
	}

	// Sort: in_progress first, then by idle time descending (most stale first)
	sortItems := func(items []watchlistItem) {
		sort.Slice(items, func(i, j int) bool {
			statusOrder := map[string]int{"in_progress": 0, "hooked": 0, "open": 1, "blocked": 2}
			si := statusOrder[items[i].Issue.Status]
			sj := statusOrder[items[j].Issue.Status]
			if si != sj {
				return si < sj
			}
			return items[i].IdleH > items[j].IdleH
		})
	}
	sortItems(data.P0)
	sortItems(data.P1)

	data.TotalP0 = len(data.P0)
	data.TotalP1 = len(data.P1)

	// Collect distinct assignees for reassign dropdown
	assigneeSet := make(map[string]bool)
	for _, r := range results {
		for _, a := range r.assignees {
			if a != "" {
				assigneeSet[a] = true
			}
		}
	}
	for a := range assigneeSet {
		data.Assignees = append(data.Assignees, a)
	}
	sort.Strings(data.Assignees)

	s.render(w, r, "watchlist", data)
}
