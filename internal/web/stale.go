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

type staleEntry struct {
	Issue    dolt.Issue
	Rig      string
	DaysStale int
}

type staleData struct {
	Entries []staleEntry
	Total   int
	Days    int
	Err     string
}

func (s *Server) handleStale(w http.ResponseWriter, r *http.Request) {
	days := 3
	if d := r.URL.Query().Get("days"); d != "" {
		if v, err := strconv.Atoi(d); err == nil && v > 0 && v <= 90 {
			days = v
		}
	}

	if s.ds == nil {
		s.render(w, r, "stale", staleData{Days: days})
		return
	}

	ctx := r.Context()
	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("stale: list dbs: %v", err)
		s.render(w, r, "stale", staleData{Days: days, Err: err.Error()})
		return
	}

	cutoff := time.Now().Add(-time.Duration(days) * 24 * time.Hour)

	type dbResult struct {
		entries []staleEntry
	}

	results := make([]dbResult, len(dbs))
	var wg sync.WaitGroup
	for i, db := range dbs {
		wg.Add(1)
		go func(i int, dbName string) {
			defer wg.Done()
			var res dbResult

			for _, status := range []string{"in_progress", "hooked"} {
				issues, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{
					Status:        status,
					UpdatedBefore: cutoff,
					Limit:         100,
				})
				if err != nil {
					log.Printf("stale %s/%s: %v", dbName, status, err)
					continue
				}
				for _, iss := range issues {
					daysSince := int(time.Since(iss.UpdatedAt).Hours() / 24)
					res.entries = append(res.entries, staleEntry{
						Issue:     iss,
						Rig:       dbName,
						DaysStale: daysSince,
					})
				}
			}
			results[i] = res
		}(i, db.Name)
	}
	wg.Wait()

	var all []staleEntry
	for _, r := range results {
		all = append(all, r.entries...)
	}

	sort.Slice(all, func(i, j int) bool {
		return all[i].DaysStale > all[j].DaysStale
	})

	s.render(w, r, "stale", staleData{
		Entries: all,
		Total:   len(all),
		Days:    days,
	})
}
