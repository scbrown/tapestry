package web

import (
	"log"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/scbrown/tapestry/internal/dolt"
)

type unblockedItem struct {
	Rig           string
	Issue         dolt.Issue
	BlockedDays   int
	UnblockedAt   time.Time
	CurrentStatus string
}

type unblockedData struct {
	GeneratedAt time.Time
	Items       []unblockedItem
	Total       int
	Rigs        []string
	FilterRig   string
}

func (s *Server) handleUnblocked(w http.ResponseWriter, r *http.Request) {
	data := unblockedData{GeneratedAt: time.Now()}

	if s.ds == nil {
		s.render(w, r, "unblocked", data)
		return
	}

	ctx := r.Context()
	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("unblocked: list dbs: %v", err)
		s.render(w, r, "unblocked", data)
		return
	}

	filterRig := r.URL.Query().Get("rig")
	var rigs []string
	for _, db := range dbs {
		rigs = append(rigs, db.Name)
	}
	sort.Strings(rigs)
	data.Rigs = rigs
	data.FilterRig = filterRig

	// Gather open + in_progress issues from each DB
	type dbResult struct {
		rig    string
		issues []dolt.Issue
	}
	results := make([]dbResult, len(dbs))
	var wg sync.WaitGroup
	for i, db := range dbs {
		if filterRig != "" && db.Name != filterRig {
			continue
		}
		wg.Add(1)
		go func(i int, dbName string) {
			defer wg.Done()
			open, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{Status: "open", Limit: 5000})
			if err != nil {
				log.Printf("unblocked: %s open: %v", dbName, err)
				return
			}
			prog, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{Status: "in_progress", Limit: 5000})
			if err != nil {
				log.Printf("unblocked: %s in_progress: %v", dbName, err)
			}
			all := append(open, prog...)
			results[i] = dbResult{rig: dbName, issues: all}
		}(i, db.Name)
	}
	wg.Wait()

	// For each issue, check status history for blocked -> non-blocked transition
	var mu sync.Mutex
	var histWg sync.WaitGroup
	sem := make(chan struct{}, 10)

	for _, r := range results {
		for _, iss := range r.issues {
			histWg.Add(1)
			sem <- struct{}{}
			go func(rig string, issue dolt.Issue) {
				defer histWg.Done()
				defer func() { <-sem }()

				hist, err := s.ds.StatusHistory(ctx, rig, issue.ID)
				if err != nil || len(hist) < 2 {
					return
				}

				// Walk transitions to find the last blocked -> non-blocked
				var blockedStart time.Time
				var unblockedAt time.Time
				wasBlocked := false

				for _, t := range hist {
					if t.ToStatus == "blocked" {
						blockedStart = t.CommitDate
						wasBlocked = true
					} else if wasBlocked && t.FromStatus == "blocked" {
						unblockedAt = t.CommitDate
					}
				}

				if wasBlocked && !unblockedAt.IsZero() {
					days := 0
					if !blockedStart.IsZero() {
						days = int(unblockedAt.Sub(blockedStart).Hours() / 24)
						if days < 0 {
							days = 0
						}
					}
					mu.Lock()
					data.Items = append(data.Items, unblockedItem{
						Rig:           rig,
						Issue:         issue,
						BlockedDays:   days,
						UnblockedAt:   unblockedAt,
						CurrentStatus: issue.Status,
					})
					mu.Unlock()
				}
			}(r.rig, iss)
		}
	}
	histWg.Wait()

	// Sort by unblocked date descending (most recent first)
	sort.Slice(data.Items, func(i, j int) bool {
		return data.Items[i].UnblockedAt.After(data.Items[j].UnblockedAt)
	})

	data.Total = len(data.Items)
	s.render(w, r, "unblocked", data)
}
