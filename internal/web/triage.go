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
	Assignees    []string // known assignees for quick-assign dropdown
	Rigs         []string
	FilterRig    string
	SortBy       string
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
		rig       string
		issues    []dolt.Issue
		assignees []string
	}
	results := make([]dbResult, len(dbs))
	var wg sync.WaitGroup
	for i, db := range dbs {
		wg.Add(1)
		go func(i int, dbName string) {
			defer wg.Done()
			issues, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{Status: "open", Limit: 2000})
			if err != nil {
				log.Printf("triage: %s: %v", dbName, err)
				return
			}
			assignees, _ := s.ds.DistinctAssignees(ctx, dbName)
			results[i] = dbResult{rig: dbName, issues: issues, assignees: assignees}
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

	// Collect distinct rigs for filter
	rigSet := make(map[string]bool)
	for _, e := range data.Unassigned {
		rigSet[e.Rig] = true
	}
	for _, e := range data.NoPriority {
		rigSet[e.Rig] = true
	}
	var rigs []string
	for rig := range rigSet {
		rigs = append(rigs, rig)
	}
	sort.Strings(rigs)
	data.Rigs = rigs

	// Apply rig filter
	filterRig := r.URL.Query().Get("rig")
	data.FilterRig = filterRig
	if filterRig != "" {
		filtered := data.Unassigned[:0]
		for _, e := range data.Unassigned {
			if e.Rig == filterRig {
				filtered = append(filtered, e)
			}
		}
		data.Unassigned = filtered

		filteredNP := data.NoPriority[:0]
		for _, e := range data.NoPriority {
			if e.Rig == filterRig {
				filteredNP = append(filteredNP, e)
			}
		}
		data.NoPriority = filteredNP
	}

	sortBy := r.URL.Query().Get("sort")
	if sortBy == "" {
		sortBy = "age"
	}
	data.SortBy = sortBy

	// Sort both lists based on selected sort
	triageSort := func(items []triageIssue) {
		switch sortBy {
		case "priority":
			sort.Slice(items, func(i, j int) bool {
				if items[i].Issue.Priority != items[j].Issue.Priority {
					return items[i].Issue.Priority < items[j].Issue.Priority
				}
				return items[i].Age > items[j].Age
			})
		case "type":
			sort.Slice(items, func(i, j int) bool {
				if items[i].Issue.Type != items[j].Issue.Type {
					return items[i].Issue.Type < items[j].Issue.Type
				}
				return items[i].Age > items[j].Age
			})
		case "rig":
			sort.Slice(items, func(i, j int) bool {
				if items[i].Rig != items[j].Rig {
					return items[i].Rig < items[j].Rig
				}
				return items[i].Age > items[j].Age
			})
		default: // "age"
			sort.Slice(items, func(i, j int) bool {
				return items[i].Age > items[j].Age
			})
		}
	}
	triageSort(data.Unassigned)
	triageSort(data.NoPriority)

	data.UnassignedN = len(data.Unassigned)
	data.NoPriorityN = len(data.NoPriority)
	data.Untriaged = data.UnassignedN + data.NoPriorityN

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

	s.render(w, r, "triage", data)
}
