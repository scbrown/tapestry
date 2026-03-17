package web

import (
	"log"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/scbrown/tapestry/internal/dolt"
)

type churnItem struct {
	Rig         string
	Issue       dolt.Issue
	Transitions int // number of status changes
}

type churnData struct {
	GeneratedAt time.Time
	Items       []churnItem
	Total       int
	AvgChurn    float64
	MaxChurn    int
	Rigs        []string
	FilterRig   string
	Assignees   []string
}

func (s *Server) handleChurn(w http.ResponseWriter, r *http.Request) {
	data := churnData{GeneratedAt: time.Now()}

	if s.ds == nil {
		s.render(w, r, "churn", data)
		return
	}

	ctx := r.Context()
	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("churn: list dbs: %v", err)
		s.render(w, r, "churn", data)
		return
	}

	type dbResult struct {
		rig    string
		issues []dolt.Issue
	}
	results := make([]dbResult, len(dbs))
	var wg sync.WaitGroup
	for i, db := range dbs {
		wg.Add(1)
		go func(i int, dbName string) {
			defer wg.Done()
			assignees, _ := s.ds.DistinctAssignees(ctx, dbName)
			if len(assignees) > 0 {
				data.Assignees = append(data.Assignees, assignees...)
			}
			// Get open and in_progress issues
			open, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{Status: "open", Limit: 5000})
			if err != nil {
				log.Printf("churn: %s open: %v", dbName, err)
				return
			}
			prog, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{Status: "in_progress", Limit: 5000})
			if err != nil {
				log.Printf("churn: %s in_progress: %v", dbName, err)
			}
			all := append(open, prog...)
			results[i] = dbResult{rig: dbName, issues: all}
		}(i, db.Name)
	}
	wg.Wait()

	sort.Strings(data.Assignees)
	// For each open/in_progress issue, check status history for churn
	type histResult struct {
		rig         string
		issue       dolt.Issue
		transitions int
	}

	var histMu sync.Mutex
	var histResults []histResult

	var histWg sync.WaitGroup
	// Use semaphore to limit concurrent history queries
	sem := make(chan struct{}, 10)

	for _, r := range results {
		for _, iss := range r.issues {
			histWg.Add(1)
			sem <- struct{}{}
			go func(rig string, issue dolt.Issue) {
				defer histWg.Done()
				defer func() { <-sem }()

				hist, err := s.ds.StatusHistory(ctx, rig, issue.ID)
				if err != nil {
					return
				}
				if len(hist) >= 3 { // 3+ transitions = churn
					histMu.Lock()
					histResults = append(histResults, histResult{
						rig: rig, issue: issue, transitions: len(hist),
					})
					histMu.Unlock()
				}
			}(r.rig, iss)
		}
	}
	histWg.Wait()

	rigSet := make(map[string]bool)
	for _, hr := range histResults {
		data.Items = append(data.Items, churnItem{
			Rig: hr.rig, Issue: hr.issue, Transitions: hr.transitions,
		})
		rigSet[hr.rig] = true
		if hr.transitions > data.MaxChurn {
			data.MaxChurn = hr.transitions
		}
	}

	for rig := range rigSet {
		data.Rigs = append(data.Rigs, rig)
	}
	sort.Strings(data.Rigs)

	data.FilterRig = r.URL.Query().Get("rig")
	if data.FilterRig != "" {
		filtered := data.Items[:0]
		for _, item := range data.Items {
			if item.Rig == data.FilterRig {
				filtered = append(filtered, item)
			}
		}
		data.Items = filtered
		data.MaxChurn = 0
		for _, item := range data.Items {
			if item.Transitions > data.MaxChurn {
				data.MaxChurn = item.Transitions
			}
		}
	}

	// Sort by transitions descending
	sort.Slice(data.Items, func(i, j int) bool {
		return data.Items[i].Transitions > data.Items[j].Transitions
	})

	data.Total = len(data.Items)
	if data.Total > 0 {
		sum := 0
		for _, item := range data.Items {
			sum += item.Transitions
		}
		data.AvgChurn = float64(sum) / float64(data.Total)
	}

	s.render(w, r, "churn", data)
}
