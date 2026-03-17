package web

import (
	"log"
	"net/http"
	"sort"
	"sync"

	"github.com/scbrown/tapestry/internal/dolt"
)

type depEntry struct {
	From dolt.Issue
	To   dolt.Issue
	Type string
	Rig  string
}

type depTypeGroup struct {
	Type    string
	Entries []depEntry
}

type depsData struct {
	ByType    []depTypeGroup
	Total     int
	Stats     depStats
	Filter    string
	Rigs      []string
	FilterRig string
	Err       string
}

type depStats struct {
	DependsOn int
	ChildOf   int
	Other     int
	UniqueIDs int
}

func (s *Server) handleDeps(w http.ResponseWriter, r *http.Request) {
	if s.ds == nil {
		s.render(w, r, "deps", depsData{})
		return
	}

	ctx := r.Context()
	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("deps: list dbs: %v", err)
		s.render(w, r, "deps", depsData{Err: err.Error()})
		return
	}

	filterType := r.URL.Query().Get("type")
	filterRig := r.URL.Query().Get("rig")

	type dbResult struct {
		entries []depEntry
	}

	results := make([]dbResult, len(dbs))
	var wg sync.WaitGroup
	for i, db := range dbs {
		wg.Add(1)
		go func(i int, dbName string) {
			defer wg.Done()
			edges, err := s.ds.AllDependenciesWithIssues(ctx, dbName)
			if err != nil {
				log.Printf("deps %s: %v", dbName, err)
				return
			}
			var entries []depEntry
			for _, e := range edges {
				if filterType != "" && e.Type != filterType {
					continue
				}
				entries = append(entries, depEntry{
					From: e.From,
					To:   e.To,
					Type: e.Type,
					Rig:  dbName,
				})
			}
			results[i] = dbResult{entries: entries}
		}(i, db.Name)
	}
	wg.Wait()

	var all []depEntry
	rigSet := make(map[string]bool)
	for _, db := range dbs {
		rigSet[db.Name] = true
	}
	for _, r := range results {
		all = append(all, r.entries...)
	}
	var rigs []string
	for rig := range rigSet {
		rigs = append(rigs, rig)
	}
	sort.Strings(rigs)

	// Apply rig filter
	if filterRig != "" {
		filtered := all[:0]
		for _, e := range all {
			if e.Rig == filterRig {
				filtered = append(filtered, e)
			}
		}
		all = filtered
	}

	sort.Slice(all, func(i, j int) bool {
		if all[i].Type != all[j].Type {
			return all[i].Type < all[j].Type
		}
		if all[i].From.Priority != all[j].From.Priority {
			return all[i].From.Priority < all[j].From.Priority
		}
		return all[i].From.UpdatedAt.After(all[j].From.UpdatedAt)
	})

	// Group by type
	grouped := map[string][]depEntry{}
	for _, e := range all {
		grouped[e.Type] = append(grouped[e.Type], e)
	}
	var byType []depTypeGroup
	for _, t := range []string{"depends_on", "child_of"} {
		if entries, ok := grouped[t]; ok {
			byType = append(byType, depTypeGroup{Type: t, Entries: entries})
			delete(grouped, t)
		}
	}
	// Any other types
	for t, entries := range grouped {
		byType = append(byType, depTypeGroup{Type: t, Entries: entries})
	}

	// Stats
	uniqueIDs := map[string]struct{}{}
	var stats depStats
	for _, e := range all {
		uniqueIDs[e.Rig+"/"+e.From.ID] = struct{}{}
		uniqueIDs[e.Rig+"/"+e.To.ID] = struct{}{}
		switch e.Type {
		case "depends_on":
			stats.DependsOn++
		case "child_of":
			stats.ChildOf++
		default:
			stats.Other++
		}
	}
	stats.UniqueIDs = len(uniqueIDs)

	s.render(w, r, "deps", depsData{
		ByType:    byType,
		Total:     len(all),
		Stats:     stats,
		Filter:    filterType,
		Rigs:      rigs,
		FilterRig: filterRig,
	})
}
