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

type createdEntry struct {
	Issue dolt.Issue
	Rig   string
}

type createdDay struct {
	Date    string
	Entries []createdEntry
	Count   int
}

type createdData struct {
	Entries   []createdEntry
	ByDay     []createdDay
	Total     int
	Days      int
	Rigs      []string
	FilterRig string
	Assignees []string
	ByFiler   map[string]int // count by owner/filer
	TopFilers []filerCount
	Err       string
}

type filerCount struct {
	Name  string
	Count int
}

func (s *Server) handleCreated(w http.ResponseWriter, r *http.Request) {
	days := 7
	if d := r.URL.Query().Get("days"); d != "" {
		if v, err := strconv.Atoi(d); err == nil && v > 0 && v <= 90 {
			days = v
		}
	}

	if s.ds == nil {
		s.render(w, r, "created", createdData{Days: days})
		return
	}

	ctx := r.Context()
	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("created: list dbs: %v", err)
		s.render(w, r, "created", createdData{Days: days, Err: err.Error()})
		return
	}

	cutoff := time.Now().Add(-time.Duration(days) * 24 * time.Hour)

	type dbResult struct {
		entries   []createdEntry
		assignees []string
	}

	results := make([]dbResult, len(dbs))
	var wg sync.WaitGroup
	for i, db := range dbs {
		wg.Add(1)
		go func(i int, dbName string) {
			defer wg.Done()
			// Fetch all recent issues and filter by creation date
			issues, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{
				Limit: 500,
			})
			if err != nil {
				log.Printf("created %s: %v", dbName, err)
				return
			}
			var entries []createdEntry
			for _, iss := range issues {
				if !iss.CreatedAt.Before(cutoff) {
					entries = append(entries, createdEntry{Issue: iss, Rig: dbName})
				}
			}
			assignees, _ := s.ds.DistinctAssignees(ctx, dbName)
			results[i] = dbResult{entries: entries, assignees: assignees}
		}(i, db.Name)
	}
	wg.Wait()

	var all []createdEntry
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

	// Sort newest first
	sort.Slice(all, func(i, j int) bool {
		return all[i].Issue.CreatedAt.After(all[j].Issue.CreatedAt)
	})

	// Group by day
	dayMap := map[string][]createdEntry{}
	var dayOrder []string
	for _, e := range all {
		key := e.Issue.CreatedAt.Format("2006-01-02")
		if _, exists := dayMap[key]; !exists {
			dayOrder = append(dayOrder, key)
		}
		dayMap[key] = append(dayMap[key], e)
	}
	var byDay []createdDay
	for _, key := range dayOrder {
		entries := dayMap[key]
		t, _ := time.Parse("2006-01-02", key)
		byDay = append(byDay, createdDay{
			Date:    t.Format("Mon, Jan 2"),
			Entries: entries,
			Count:   len(entries),
		})
	}

	// Count by filer (owner)
	byFiler := make(map[string]int)
	for _, e := range all {
		filer := e.Issue.Owner
		if filer == "" {
			filer = "(unknown)"
		}
		byFiler[filer]++
	}
	var topFilers []filerCount
	for name, count := range byFiler {
		topFilers = append(topFilers, filerCount{Name: name, Count: count})
	}
	sort.Slice(topFilers, func(i, j int) bool {
		return topFilers[i].Count > topFilers[j].Count
	})

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

	s.render(w, r, "created", createdData{
		Entries:   all,
		ByDay:     byDay,
		Total:     len(all),
		Days:      days,
		Rigs:      rigs,
		FilterRig: filterRig,
		Assignees: assignees,
		ByFiler:   byFiler,
		TopFilers: topFilers,
	})
}
