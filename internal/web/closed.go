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

type closedData struct {
	Entries []closedEntry
	Total   int
	Days    int
	Err     string
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
		entries []closedEntry
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
			results[i] = dbResult{entries: entries}
		}(i, db.Name)
	}
	wg.Wait()

	var all []closedEntry
	for _, r := range results {
		all = append(all, r.entries...)
	}

	sort.Slice(all, func(i, j int) bool {
		return all[i].Issue.UpdatedAt.After(all[j].Issue.UpdatedAt)
	})

	s.render(w, r, "closed", closedData{
		Entries: all,
		Total:   len(all),
		Days:    days,
	})
}
