package web

import (
	"log"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/scbrown/tapestry/internal/dolt"
)

type queueItem struct {
	Issue   dolt.Issue
	Rig     string
	AgeDays int
	Score   float64 // higher = more urgent
}

type queueData struct {
	GeneratedAt time.Time
	Items       []queueItem
	ByPriority  map[int]int // count by priority
	TotalReady  int
	Assignees   []string // known assignees for quick-assign dropdown
}

func (s *Server) handleQueue(w http.ResponseWriter, r *http.Request) {
	data := queueData{
		GeneratedAt: time.Now(),
		ByPriority:  map[int]int{},
	}

	if s.ds == nil {
		s.render(w, r, "queue", data)
		return
	}

	ctx := r.Context()
	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("queue: list dbs: %v", err)
		s.render(w, r, "queue", data)
		return
	}

	// Fetch open + in_progress issues and blocked set
	type dbResult struct {
		issues    []dolt.Issue
		blocked   map[string]bool
		assignees []string
	}
	results := make([]dbResult, len(dbs))
	var wg sync.WaitGroup
	for i, db := range dbs {
		wg.Add(1)
		go func(i int, dbName string) {
			defer wg.Done()
			issues, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{Limit: 2000})
			if err != nil {
				log.Printf("queue: issues %s: %v", dbName, err)
				return
			}

			blocked := map[string]bool{}
			blockedList, err := s.ds.BlockedIssues(ctx, dbName)
			if err != nil {
				log.Printf("queue: blocked %s: %v", dbName, err)
			} else {
				for _, b := range blockedList {
					blocked[b.Issue.ID] = true
				}
			}

			assignees, _ := s.ds.DistinctAssignees(ctx, dbName)
			results[i] = dbResult{issues: issues, blocked: blocked, assignees: assignees}
		}(i, db.Name)
	}
	wg.Wait()

	now := time.Now()
	for idx, r := range results {
		rig := dbs[idx].Name
		for _, iss := range r.issues {
			// Only include open beads that aren't blocked
			if iss.Status != "open" {
				continue
			}
			if r.blocked[iss.ID] {
				continue
			}

			ageDays := int(now.Sub(iss.CreatedAt).Hours() / 24)
			if ageDays < 0 {
				ageDays = 0
			}

			// Score: priority weight * age factor
			// Lower priority number = higher urgency
			priWeight := float64(5-iss.Priority) * 10
			ageFactor := float64(ageDays) + 1
			score := priWeight * ageFactor

			data.Items = append(data.Items, queueItem{
				Issue:   iss,
				Rig:     rig,
				AgeDays: ageDays,
				Score:   score,
			})
			data.ByPriority[iss.Priority]++
		}
	}

	data.TotalReady = len(data.Items)

	// Collect distinct assignees for quick-assign dropdown
	assigneeSet := make(map[string]bool)
	for _, r := range results {
		for _, a := range r.assignees {
			if a != "" {
				assigneeSet[a] = true
			}
		}
	}
	for a := range assigneeSet {
		data.Assignees = append(data.Assignees, a)
	}
	sort.Strings(data.Assignees)

	// Sort by score descending (most urgent first)
	sort.Slice(data.Items, func(i, j int) bool {
		return data.Items[i].Score > data.Items[j].Score
	})

	// Limit to top 50
	if len(data.Items) > 50 {
		data.Items = data.Items[:50]
	}

	s.render(w, r, "queue", data)
}
