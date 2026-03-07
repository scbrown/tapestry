package web

import (
	"log"
	"net/http"
	"sort"
	"sync"

	"github.com/scbrown/tapestry/internal/dolt"
)

type blockedEntry struct {
	Issue   dolt.Issue
	Blocker dolt.Issue
	Rig     string
}

type blockedData struct {
	Entries      []blockedEntry
	Total        int
	ByBlocker    []blockerGroup
	Err          string
}

type blockerGroup struct {
	Blocker dolt.Issue
	Rig     string
	Count   int
}

func (s *Server) handleBlocked(w http.ResponseWriter, r *http.Request) {
	if s.ds == nil {
		s.render(w, r, "blocked", blockedData{})
		return
	}

	ctx := r.Context()
	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("blocked: list dbs: %v", err)
		s.render(w, r, "blocked", blockedData{Err: err.Error()})
		return
	}

	type dbResult struct {
		entries []blockedEntry
	}

	results := make([]dbResult, len(dbs))
	var wg sync.WaitGroup
	for i, db := range dbs {
		wg.Add(1)
		go func(i int, dbName string) {
			defer wg.Done()
			blocked, err := s.ds.BlockedIssues(ctx, dbName)
			if err != nil {
				log.Printf("blocked %s: %v", dbName, err)
				return
			}
			var entries []blockedEntry
			for _, bi := range blocked {
				entries = append(entries, blockedEntry{
					Issue:   bi.Issue,
					Blocker: bi.Blocker,
					Rig:     dbName,
				})
			}
			results[i] = dbResult{entries: entries}
		}(i, db.Name)
	}
	wg.Wait()

	var all []blockedEntry
	for _, r := range results {
		all = append(all, r.entries...)
	}

	sort.Slice(all, func(i, j int) bool {
		if all[i].Issue.Priority != all[j].Issue.Priority {
			return all[i].Issue.Priority < all[j].Issue.Priority
		}
		return all[i].Issue.UpdatedAt.After(all[j].Issue.UpdatedAt)
	})

	// Count how many issues each blocker is blocking
	blockerCounts := map[string]blockerGroup{}
	for _, e := range all {
		key := e.Rig + "/" + e.Blocker.ID
		g := blockerCounts[key]
		g.Blocker = e.Blocker
		g.Rig = e.Rig
		g.Count++
		blockerCounts[key] = g
	}
	var groups []blockerGroup
	for _, g := range blockerCounts {
		groups = append(groups, g)
	}
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].Count > groups[j].Count
	})

	s.render(w, r, "blocked", blockedData{
		Entries:   all,
		Total:     len(all),
		ByBlocker: groups,
	})
}
