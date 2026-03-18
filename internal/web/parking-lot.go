package web

import (
	"log"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/scbrown/tapestry/internal/dolt"
)

type parkingItem struct {
	Rig      string
	Issue    dolt.Issue
	IdleDays int // days since last update
	AgeDays  int // days since created
}

type parkingData struct {
	GeneratedAt time.Time
	Items       []parkingItem
	Total       int
	MedianIdle  int
	MaxIdle     int
	ByAssignee  map[string]int
	Rigs        []string
	FilterRig   string
	SortBy      string
	Assignees   []string
}

func (s *Server) handleParkingLot(w http.ResponseWriter, r *http.Request) {
	data := parkingData{GeneratedAt: time.Now(), ByAssignee: map[string]int{}}

	if s.ds == nil {
		s.render(w, r, "parking-lot", data)
		return
	}

	ctx := r.Context()
	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("parking-lot: list dbs: %v", err)
		s.render(w, r, "parking-lot", data)
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
			issues, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{Status: "in_progress", Limit: 5000})
			if err != nil {
				log.Printf("parking-lot: %s: %v", dbName, err)
				return
			}
			assignees, _ := s.ds.DistinctAssignees(ctx, dbName)
			results[i] = dbResult{rig: dbName, issues: issues, assignees: assignees}
		}(i, db.Name)
	}
	wg.Wait()

	now := time.Now()
	threshold := 3 // days idle before considered "parked"

	rigSet := make(map[string]bool)
	for _, r := range results {
		for _, iss := range r.issues {
			idle := int(now.Sub(iss.UpdatedAt).Hours() / 24)
			if idle < threshold {
				continue
			}
			rigSet[r.rig] = true
			age := int(now.Sub(iss.CreatedAt).Hours() / 24)
			data.Items = append(data.Items, parkingItem{
				Rig: r.rig, Issue: iss, IdleDays: idle, AgeDays: age,
			})
			assignee := iss.Assignee
			if assignee == "" {
				assignee = iss.Owner
			}
			if assignee == "" {
				assignee = "(unassigned)"
			}
			data.ByAssignee[assignee]++
		}
	}

	var rigs []string
	for rig := range rigSet {
		rigs = append(rigs, rig)
	}
	sort.Strings(rigs)
	data.Rigs = rigs

	filterRig := r.URL.Query().Get("rig")
	data.FilterRig = filterRig
	if filterRig != "" {
		filtered := data.Items[:0]
		filteredAssignee := map[string]int{}
		for _, item := range data.Items {
			if item.Rig == filterRig {
				filtered = append(filtered, item)
				assignee := item.Issue.Assignee
				if assignee == "" {
					assignee = item.Issue.Owner
				}
				if assignee == "" {
					assignee = "(unassigned)"
				}
				filteredAssignee[assignee]++
			}
		}
		data.Items = filtered
		data.ByAssignee = filteredAssignee
	}

	sortBy := r.URL.Query().Get("sort")
	if sortBy == "" {
		sortBy = "idle"
	}
	data.SortBy = sortBy

	switch sortBy {
	case "priority":
		sort.Slice(data.Items, func(i, j int) bool {
			if data.Items[i].Issue.Priority != data.Items[j].Issue.Priority {
				return data.Items[i].Issue.Priority < data.Items[j].Issue.Priority
			}
			return data.Items[i].IdleDays > data.Items[j].IdleDays
		})
	case "age":
		sort.Slice(data.Items, func(i, j int) bool {
			return data.Items[i].AgeDays > data.Items[j].AgeDays
		})
	case "rig":
		sort.Slice(data.Items, func(i, j int) bool {
			if data.Items[i].Rig != data.Items[j].Rig {
				return data.Items[i].Rig < data.Items[j].Rig
			}
			return data.Items[i].IdleDays > data.Items[j].IdleDays
		})
	default: // idle
		sort.Slice(data.Items, func(i, j int) bool {
			return data.Items[i].IdleDays > data.Items[j].IdleDays
		})
	}

	data.Total = len(data.Items)
	if data.Total > 0 {
		data.MaxIdle = data.Items[0].IdleDays
		idles := make([]int, len(data.Items))
		for i, item := range data.Items {
			idles[i] = item.IdleDays
		}
		data.MedianIdle = idles[len(idles)/2]
	}

	// Collect distinct assignees for reassign dropdown
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

	s.render(w, r, "parking-lot", data)
}
