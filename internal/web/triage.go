package web

import (
	"log"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/scbrown/tapestry/internal/dolt"
)

type triageIssue struct {
	Rig   string
	Issue dolt.Issue
	Age   int // days since created
}

type triageData struct {
	GeneratedAt  time.Time
	Unassigned   []triageIssue
	NoPriority   []triageIssue // priority < 0 or default (priority == 0 but P0 is valid, we'll use -1 sentinel)
	Untriaged    int           // total needing attention
	UnassignedN  int
	NoPriorityN  int
}

func (s *Server) handleTriage(w http.ResponseWriter, r *http.Request) {
	data := triageData{GeneratedAt: time.Now()}

	if s.ds == nil {
		s.render(w, r, "triage", data)
		return
	}

	ctx := r.Context()
	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("triage: list dbs: %v", err)
		s.render(w, r, "triage", data)
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
			// Get open issues only
			issues, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{Status: "open", Limit: 2000})
			if err != nil {
				log.Printf("triage: %s: %v", dbName, err)
				return
			}
			results[i] = dbResult{rig: dbName, issues: issues}
		}(i, db.Name)
	}
	wg.Wait()

	now := time.Now()
	seen := map[string]bool{} // track IDs to avoid duplicates across categories

	for _, r := range results {
		for _, iss := range r.issues {
			age := int(now.Sub(iss.CreatedAt).Hours() / 24)
			entry := triageIssue{Rig: r.rig, Issue: iss, Age: age}

			if iss.Assignee == "" && iss.Owner == "" {
				data.Unassigned = append(data.Unassigned, entry)
				seen[iss.ID] = true
			}
		}
	}

	// No-priority: issues with priority -1 (sentinel for "not set")
	// In beads, priority defaults to a valid int. P0=0, P1=1, etc.
	// "No priority" means priority was left at a high default (e.g., 5+ or -1).
	// For now, treat priority > 4 as "no priority set".
	for _, r := range results {
		for _, iss := range r.issues {
			if iss.Priority > 4 && !seen[iss.ID] {
				age := int(now.Sub(iss.CreatedAt).Hours() / 24)
				data.NoPriority = append(data.NoPriority, triageIssue{Rig: r.rig, Issue: iss, Age: age})
			}
		}
	}

	// Sort both lists by age descending (oldest first)
	sort.Slice(data.Unassigned, func(i, j int) bool {
		return data.Unassigned[i].Age > data.Unassigned[j].Age
	})
	sort.Slice(data.NoPriority, func(i, j int) bool {
		return data.NoPriority[i].Age > data.NoPriority[j].Age
	})

	data.UnassignedN = len(data.Unassigned)
	data.NoPriorityN = len(data.NoPriority)
	data.Untriaged = data.UnassignedN + data.NoPriorityN

	s.render(w, r, "triage", data)
}
