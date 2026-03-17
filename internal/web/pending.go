package web

import (
	"log"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/scbrown/tapestry/internal/dolt"
)

type pendingItem struct {
	dolt.Issue
	Rig    string
	Reason string // "unblocked", "assigned-not-started", "high-pri-open"
	Age    string // human-readable time since assignment/unblock
}

type pendingData struct {
	GeneratedAt time.Time
	Items       []pendingItem
	Total       int
	Rigs        []string
	FilterRig   string
	Err         string
}

func (s *Server) handlePending(w http.ResponseWriter, r *http.Request) {
	data := pendingData{GeneratedAt: time.Now()}

	if s.ds == nil {
		s.render(w, r, "pending", data)
		return
	}

	ctx := r.Context()
	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("pending: list dbs: %v", err)
		s.render(w, r, "pending", data)
		return
	}

	filterRig := r.URL.Query().Get("rig")
	data.FilterRig = filterRig

	var rigs []string
	for _, db := range dbs {
		rigs = append(rigs, db.Name)
	}
	sort.Strings(rigs)
	data.Rigs = rigs

	now := time.Now()

	type dbResult struct {
		items []pendingItem
	}
	results := make([]dbResult, len(dbs))
	var wg sync.WaitGroup

	for i, db := range dbs {
		if filterRig != "" && db.Name != filterRig {
			continue
		}
		wg.Add(1)
		go func(idx int, dbName string) {
			defer wg.Done()
			issues, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{Limit: 5000})
			if err != nil {
				log.Printf("pending: issues %s: %v", dbName, err)
				return
			}

			var items []pendingItem
			for _, iss := range issues {
				if iss.Status == "closed" || iss.Status == "deferred" || isNoise(iss.ID, iss.Title) {
					continue
				}

				var reason string

				// High-priority open beads with assignee but not started
				if iss.Status == "open" && iss.Assignee != "" && iss.Priority <= 2 {
					reason = "assigned-not-started"
				}

				// High-priority unassigned open beads
				if iss.Status == "open" && iss.Assignee == "" && iss.Priority <= 1 {
					reason = "high-pri-open"
				}

				if reason == "" {
					continue
				}

				items = append(items, pendingItem{
					Issue:  iss,
					Rig:    dbName,
					Reason: reason,
					Age:    formatDwell(now.Sub(iss.UpdatedAt)),
				})
			}
			results[idx] = dbResult{items: items}
		}(i, db.Name)
	}
	wg.Wait()

	var all []pendingItem
	for _, r := range results {
		all = append(all, r.items...)
	}

	// Sort by priority then age
	sort.Slice(all, func(i, j int) bool {
		if all[i].Priority != all[j].Priority {
			return all[i].Priority < all[j].Priority
		}
		return all[i].UpdatedAt.Before(all[j].UpdatedAt)
	})

	if len(all) > 100 {
		all = all[:100]
	}

	data.Items = all
	data.Total = len(all)
	s.render(w, r, "pending", data)
}
