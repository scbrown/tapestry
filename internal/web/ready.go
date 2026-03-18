package web

import (
	"context"
	"log"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/scbrown/tapestry/internal/dolt"
)

type readyItem struct {
	ID       string
	DB       string
	Title    string
	Priority int
	Type     string
	Assignee string
	AgeDays  int
	CreatedAt time.Time
}

type readyData struct {
	GeneratedAt time.Time

	Items      []readyItem
	TotalReady int

	// Priority breakdown
	P0 int
	P1 int
	P2 int
	P3 int

	Rigs      []string
	FilterRig string
	SortBy    string
	Assignees []string
	Err       string
}

func (s *Server) handleReady(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	filterRig := r.URL.Query().Get("rig")
	data := readyData{GeneratedAt: now, FilterRig: filterRig}

	if s.ds == nil {
		s.render(w, r, "ready", data)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("ready: list dbs: %v", err)
		data.Err = err.Error()
		s.render(w, r, "ready", data)
		return
	}

	for _, db := range dbs {
		data.Rigs = append(data.Rigs, db.Name)
	}
	sort.Strings(data.Rigs)

	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, db := range dbs {
		if filterRig != "" && db.Name != filterRig {
			continue
		}
		wg.Add(1)
		go func(dbName string) {
			defer wg.Done()

			assignees, _ := s.ds.DistinctAssignees(ctx, dbName)

			// Get blocked issues
			blocked, err := s.ds.BlockedIssues(ctx, dbName)
			if err != nil {
				log.Printf("ready %s: blocked: %v", dbName, err)
			}

			issues, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{Limit: 2000})
			if err != nil {
				log.Printf("ready %s: %v", dbName, err)
				return
			}

			localBlocked := make(map[string]bool)
			for _, b := range blocked {
				localBlocked[b.Issue.ID] = true
			}

			var items []readyItem
			for _, iss := range issues {
				if isNoise(iss.ID, iss.Title) {
					continue
				}
				if iss.Status != "open" {
					continue
				}
				if localBlocked[iss.ID] {
					continue
				}

				items = append(items, readyItem{
					ID:        iss.ID,
					DB:        dbName,
					Title:     iss.Title,
					Priority:  iss.Priority,
					Type:      iss.Type,
					Assignee:  iss.Assignee,
					AgeDays:   int(now.Sub(iss.CreatedAt).Hours() / 24),
					CreatedAt: iss.CreatedAt,
				})
			}

			mu.Lock()
			data.Items = append(data.Items, items...)
			for _, a := range assignees {
				data.Assignees = append(data.Assignees, a)
			}
			mu.Unlock()
		}(db.Name)
	}
	wg.Wait()

	// Deduplicate and sort assignees
	assigneeSet := make(map[string]bool)
	for _, a := range data.Assignees {
		assigneeSet[a] = true
	}
	data.Assignees = data.Assignees[:0]
	for a := range assigneeSet {
		data.Assignees = append(data.Assignees, a)
	}
	sort.Strings(data.Assignees)

	sortBy := r.URL.Query().Get("sort")
	if sortBy == "" {
		sortBy = "priority"
	}
	data.SortBy = sortBy

	switch sortBy {
	case "age":
		sort.Slice(data.Items, func(i, j int) bool {
			return data.Items[i].AgeDays > data.Items[j].AgeDays
		})
	case "type":
		sort.Slice(data.Items, func(i, j int) bool {
			if data.Items[i].Type != data.Items[j].Type {
				return data.Items[i].Type < data.Items[j].Type
			}
			return data.Items[i].Priority < data.Items[j].Priority
		})
	case "assignee":
		sort.Slice(data.Items, func(i, j int) bool {
			if data.Items[i].Assignee != data.Items[j].Assignee {
				return data.Items[i].Assignee < data.Items[j].Assignee
			}
			return data.Items[i].Priority < data.Items[j].Priority
		})
	default: // "priority"
		sort.Slice(data.Items, func(i, j int) bool {
			if data.Items[i].Priority != data.Items[j].Priority {
				return data.Items[i].Priority < data.Items[j].Priority
			}
			return data.Items[i].CreatedAt.Before(data.Items[j].CreatedAt)
		})
	}

	data.TotalReady = len(data.Items)
	for _, item := range data.Items {
		switch item.Priority {
		case 0:
			data.P0++
		case 1:
			data.P1++
		case 2:
			data.P2++
		default:
			data.P3++
		}
	}

	if len(data.Items) > 100 {
		data.Items = data.Items[:100]
	}

	s.render(w, r, "ready", data)
}
