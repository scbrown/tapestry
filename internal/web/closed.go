package web

import (
	"log"
	"net/http"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/scbrown/tapestry/internal/dolt"
)

type closedEntry struct {
	Issue dolt.Issue
	Rig   string
}

type closedDay struct {
	Date    string
	Entries []closedEntry
	Count   int
}

type closedData struct {
	Entries   []closedEntry
	ByDay     []closedDay
	Total     int
	Days      int
	Rigs      []string
	FilterRig string
	SortBy    string
	Assignees []string
	Err       string
}

func (s *Server) handleClosed(w http.ResponseWriter, r *http.Request) {
	days := 7
	if d := r.URL.Query().Get("days"); d != "" {
		if v, err := strconv.Atoi(d); err == nil && v > 0 && v <= 90 {
			days = v
		}
	}

	if s.ds == nil {
		s.render(w, r, "closed", closedData{Days: days})
		return
	}

	ctx := r.Context()
	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("closed: list dbs: %v", err)
		s.render(w, r, "closed", closedData{Days: days, Err: err.Error()})
		return
	}

	cutoff := time.Now().Add(-time.Duration(days) * 24 * time.Hour)

	type dbResult struct {
		entries   []closedEntry
		assignees []string
	}

	results := make([]dbResult, len(dbs))
	var wg sync.WaitGroup
	for i, db := range dbs {
		wg.Add(1)
		go func(i int, dbName string) {
			defer wg.Done()
			issues, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{
				Status:       "closed",
				UpdatedAfter: cutoff,
				Limit:        200,
			})
			if err != nil {
				log.Printf("closed %s: %v", dbName, err)
				return
			}
			var entries []closedEntry
			for _, iss := range issues {
				entries = append(entries, closedEntry{Issue: iss, Rig: dbName})
			}
			assignees, _ := s.ds.DistinctAssignees(ctx, dbName)
			results[i] = dbResult{entries: entries, assignees: assignees}
		}(i, db.Name)
	}
	wg.Wait()

	var all []closedEntry
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
		sortBy = "closed"
	}

	switch sortBy {
	case "priority":
		sort.Slice(all, func(i, j int) bool {
			if all[i].Issue.Priority != all[j].Issue.Priority {
				return all[i].Issue.Priority < all[j].Issue.Priority
			}
			return all[i].Issue.UpdatedAt.After(all[j].Issue.UpdatedAt)
		})
	case "assignee":
		sort.Slice(all, func(i, j int) bool {
			if all[i].Issue.Assignee != all[j].Issue.Assignee {
				return all[i].Issue.Assignee < all[j].Issue.Assignee
			}
			return all[i].Issue.UpdatedAt.After(all[j].Issue.UpdatedAt)
		})
	case "rig":
		sort.Slice(all, func(i, j int) bool {
			if all[i].Rig != all[j].Rig {
				return all[i].Rig < all[j].Rig
			}
			return all[i].Issue.UpdatedAt.After(all[j].Issue.UpdatedAt)
		})
	default: // "closed"
		sort.Slice(all, func(i, j int) bool {
			return all[i].Issue.UpdatedAt.After(all[j].Issue.UpdatedAt)
		})
	}

	// Group by day
	dayMap := map[string][]closedEntry{}
	var dayOrder []string
	for _, e := range all {
		key := e.Issue.UpdatedAt.Format("2006-01-02")
		if _, exists := dayMap[key]; !exists {
			dayOrder = append(dayOrder, key)
		}
		dayMap[key] = append(dayMap[key], e)
	}
	var byDay []closedDay
	for _, key := range dayOrder {
		entries := dayMap[key]
		t, _ := time.Parse("2006-01-02", key)
		byDay = append(byDay, closedDay{
			Date:    t.Format("Mon, Jan 2"),
			Entries: entries,
			Count:   len(entries),
		})
	}

	// Collect distinct assignees
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

	s.render(w, r, "closed", closedData{
		Entries:   all,
		ByDay:     byDay,
		Total:     len(all),
		Days:      days,
		Rigs:      rigs,
		FilterRig: filterRig,
		SortBy:    sortBy,
		Assignees: assignees,
	})
}
