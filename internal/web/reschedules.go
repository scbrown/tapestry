package web

import (
	"log"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/scbrown/tapestry/internal/dolt"
)

type rescheduleItem struct {
	Rig        string
	Issue      dolt.Issue
	DeferCount int
	LastDefer  time.Time
}

type rescheduleData struct {
	GeneratedAt time.Time
	Items       []rescheduleItem
	Total       int
	Rigs        []string
	FilterRig   string
	Assignees   []string
}

func (s *Server) handleReschedules(w http.ResponseWriter, r *http.Request) {
	data := rescheduleData{GeneratedAt: time.Now()}

	if s.ds == nil {
		s.render(w, r, "reschedules", data)
		return
	}

	ctx := r.Context()
	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("reschedules: list dbs: %v", err)
		s.render(w, r, "reschedules", data)
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

	// Fetch all non-closed issues from each DB
	type dbResult struct {
		rig       string
		issues    []dolt.Issue
		assignees []string
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
			assignees, _ := s.ds.DistinctAssignees(ctx, dbName)
			// Get deferred issues — these are the most likely reschedule candidates
			deferred, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{Status: "deferred", Limit: 5000})
			if err != nil {
				log.Printf("reschedules: %s deferred: %v", dbName, err)
				return
			}
			open, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{Status: "open", Limit: 5000})
			if err != nil {
				log.Printf("reschedules: %s open: %v", dbName, err)
			}
			prog, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{Status: "in_progress", Limit: 5000})
			if err != nil {
				log.Printf("reschedules: %s in_progress: %v", dbName, err)
			}
			all := append(deferred, open...)
			all = append(all, prog...)
			results[idx] = dbResult{rig: dbName, issues: all, assignees: assignees}
		}(i, db.Name)
	}
	wg.Wait()

	// Collect assignees from all DBs
	assigneeSet := make(map[string]bool)
	for _, r := range results {
		for _, a := range r.assignees {
			assigneeSet[a] = true
		}
	}
	for a := range assigneeSet {
		data.Assignees = append(data.Assignees, a)
	}
	sort.Strings(data.Assignees)

	// Check status history for each issue — count times it went to "deferred"
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

				deferCount := 0
				var lastDefer time.Time
				for _, t := range hist {
					if t.ToStatus == "deferred" {
						deferCount++
						lastDefer = t.CommitDate
					}
				}

				if deferCount >= 2 {
					mu.Lock()
					data.Items = append(data.Items, rescheduleItem{
						Rig:        rig,
						Issue:      issue,
						DeferCount: deferCount,
						LastDefer:  lastDefer,
					})
					mu.Unlock()
				}
			}(r.rig, iss)
		}
	}
	histWg.Wait()

	// Sort by defer count descending, then by last defer time
	sort.Slice(data.Items, func(i, j int) bool {
		if data.Items[i].DeferCount != data.Items[j].DeferCount {
			return data.Items[i].DeferCount > data.Items[j].DeferCount
		}
		return data.Items[i].LastDefer.After(data.Items[j].LastDefer)
	})

	data.Total = len(data.Items)
	s.render(w, r, "reschedules", data)
}
