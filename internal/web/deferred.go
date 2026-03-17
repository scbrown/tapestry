package web

import (
	"log"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/scbrown/tapestry/internal/dolt"
)

type deferredItem struct {
	Rig      string
	Issue    dolt.Issue
	AgeDays  int // days since created
	IdleDays int // days since last updated
}

type deferredData struct {
	GeneratedAt time.Time
	Items       []deferredItem
	Total       int
	ByRig       map[string]int
	ByPriority  [5]int // P0-P4
	MedianAge   int
	OldestAge   int
	Rigs        []string
	FilterRig   string
}

func (s *Server) handleDeferred(w http.ResponseWriter, r *http.Request) {
	data := deferredData{GeneratedAt: time.Now(), ByRig: map[string]int{}}

	if s.ds == nil {
		s.render(w, r, "deferred", data)
		return
	}

	ctx := r.Context()
	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("deferred: list dbs: %v", err)
		s.render(w, r, "deferred", data)
		return
	}

	type dbResult struct {
		rig    string
		issues []dolt.Issue
	}
	results := make([]dbResult, len(dbs))
	var wg sync.WaitGroup
	for i, db := range dbs {
		wg.Add(1)
		go func(i int, dbName string) {
			defer wg.Done()
			issues, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{Status: "deferred", Limit: 5000})
			if err != nil {
				log.Printf("deferred: %s: %v", dbName, err)
				return
			}
			results[i] = dbResult{rig: dbName, issues: issues}
		}(i, db.Name)
	}
	wg.Wait()

	now := time.Now()
	for _, r := range results {
		for _, iss := range r.issues {
			age := int(now.Sub(iss.CreatedAt).Hours() / 24)
			idle := int(now.Sub(iss.UpdatedAt).Hours() / 24)
			data.Items = append(data.Items, deferredItem{
				Rig: r.rig, Issue: iss, AgeDays: age, IdleDays: idle,
			})
			data.ByRig[r.rig]++
			if iss.Priority >= 0 && iss.Priority <= 4 {
				data.ByPriority[iss.Priority]++
			}
		}
	}

	// Collect distinct rigs for filter
	var rigs []string
	for rig := range data.ByRig {
		rigs = append(rigs, rig)
	}
	sort.Strings(rigs)
	data.Rigs = rigs

	// Apply rig filter
	filterRig := r.URL.Query().Get("rig")
	data.FilterRig = filterRig
	if filterRig != "" {
		filtered := data.Items[:0]
		filteredByRig := map[string]int{}
		var filteredPriority [5]int
		for _, item := range data.Items {
			if item.Rig == filterRig {
				filtered = append(filtered, item)
				filteredByRig[item.Rig]++
				if item.Issue.Priority >= 0 && item.Issue.Priority <= 4 {
					filteredPriority[item.Issue.Priority]++
				}
			}
		}
		data.Items = filtered
		data.ByRig = filteredByRig
		data.ByPriority = filteredPriority
	}

	// Sort by idle days descending (longest-parked first)
	sort.Slice(data.Items, func(i, j int) bool {
		return data.Items[i].IdleDays > data.Items[j].IdleDays
	})

	data.Total = len(data.Items)
	if data.Total > 0 {
		data.OldestAge = data.Items[0].IdleDays
		ages := make([]int, len(data.Items))
		for i, item := range data.Items {
			ages[i] = item.IdleDays
		}
		data.MedianAge = ages[len(ages)/2]
	}

	s.render(w, r, "deferred", data)
}
