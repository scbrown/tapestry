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
	Rigs         []string
	FilterRig    string
	SortBy       string
	Assignees    []string
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
		entries   []blockedEntry
		assignees []string
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
			assignees, _ := s.ds.DistinctAssignees(ctx, dbName)
			results[i] = dbResult{entries: entries, assignees: assignees}
		}(i, db.Name)
	}
	wg.Wait()

	var all []blockedEntry
	rigSet := make(map[string]bool)
	for _, r := range results {
		all = append(all, r.entries...)
		for _, e := range r.entries {
			rigSet[e.Rig] = true
		}
	}

	var rigs []string
	for rig := range rigSet {
		rigs = append(rigs, rig)
	}
	sort.Strings(rigs)

	filterRig := r.URL.Query().Get("rig")
	if filterRig != "" {
		filtered := all[:0]
		for _, e := range all {
			if e.Rig == filterRig {
				filtered = append(filtered, e)
			}
		}
		all = filtered
	}

	sortBy := r.URL.Query().Get("sort")
	if sortBy == "" {
		sortBy = "priority"
	}

	switch sortBy {
	case "blocker":
		// Will be re-sorted after blocker counting; for now sort by priority
		sort.Slice(all, func(i, j int) bool {
			return all[i].Issue.Priority < all[j].Issue.Priority
		})
	case "rig":
		sort.Slice(all, func(i, j int) bool {
			if all[i].Rig != all[j].Rig {
				return all[i].Rig < all[j].Rig
			}
			return all[i].Issue.Priority < all[j].Issue.Priority
		})
	case "assignee":
		sort.Slice(all, func(i, j int) bool {
			if all[i].Issue.Assignee != all[j].Issue.Assignee {
				return all[i].Issue.Assignee < all[j].Issue.Assignee
			}
			return all[i].Issue.Priority < all[j].Issue.Priority
		})
	default: // "priority"
		sort.Slice(all, func(i, j int) bool {
			if all[i].Issue.Priority != all[j].Issue.Priority {
				return all[i].Issue.Priority < all[j].Issue.Priority
			}
			return all[i].Issue.UpdatedAt.After(all[j].Issue.UpdatedAt)
		})
	}

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

	// Collect distinct assignees for reassign dropdown
	assigneeSet := make(map[string]bool)
	for _, r := range results {
		for _, a := range r.assignees {
			if a != "" {
				assigneeSet[a] = true
			}
		}
	}
	var assignees []string
	for a := range assigneeSet {
		assignees = append(assignees, a)
	}
	sort.Strings(assignees)

	s.render(w, r, "blocked", blockedData{
		Entries:   all,
		Total:     len(all),
		ByBlocker: groups,
		Rigs:      rigs,
		FilterRig: filterRig,
		SortBy:    sortBy,
		Assignees: assignees,
	})
}
