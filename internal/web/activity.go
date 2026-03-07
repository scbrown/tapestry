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

type activityEntry struct {
	Issue dolt.Issue
	Rig   string
}

type activityData struct {
	Entries []activityEntry
	Total   int
	Hours   int
	Err     string
}

func (s *Server) handleActivity(w http.ResponseWriter, r *http.Request) {
	if s.ds == nil {
		s.render(w, r, "activity", activityData{Hours: 24})
		return
	}

	ctx := r.Context()
	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("activity: list dbs: %v", err)
		s.render(w, r, "activity", activityData{Err: err.Error(), Hours: 24})
		return
	}

	hours := 24
	if h := r.URL.Query().Get("hours"); h != "" {
		if v, err := strconv.Atoi(h); err == nil && v > 0 && v <= 168 {
			hours = v
		}
	}

	cutoff := time.Now().Add(-time.Duration(hours) * time.Hour)

	type dbResult struct {
		entries []activityEntry
	}

	results := make([]dbResult, len(dbs))
	var wg sync.WaitGroup
	for i, db := range dbs {
		wg.Add(1)
		go func(i int, dbName string) {
			defer wg.Done()
			issues, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{
				UpdatedAfter: cutoff,
				Limit:        200,
			})
			if err != nil {
				log.Printf("activity %s: %v", dbName, err)
				return
			}
			var entries []activityEntry
			for _, issue := range issues {
				issue.Rig = dbName
				entries = append(entries, activityEntry{
					Issue: issue,
					Rig:   dbName,
				})
			}
			results[i] = dbResult{entries: entries}
		}(i, db.Name)
	}
	wg.Wait()

	var all []activityEntry
	for _, r := range results {
		all = append(all, r.entries...)
	}

	// Sort by most recently updated first
	sort.Slice(all, func(i, j int) bool {
		return all[i].Issue.UpdatedAt.After(all[j].Issue.UpdatedAt)
	})

	// Cap to 200
	if len(all) > 200 {
		all = all[:200]
	}

	s.render(w, r, "activity", activityData{
		Entries: all,
		Total:   len(all),
		Hours:   hours,
	})
}
