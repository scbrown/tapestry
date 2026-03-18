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

type escalatedBead struct {
	ID          string
	Title       string
	Status      string
	Priority    int
	Assignee    string
	DB          string
	PriorityOld int
	PriorityNew int
	EscalatedAt time.Time
}

type escalationsData struct {
	GeneratedAt time.Time
	FilterRig   string
	SortBy      string

	Escalations []escalatedBead
	TotalEsc    int

	// Summary counts
	ToP0 int
	ToP1 int

	Assignees []string
	Err       string
}

func (s *Server) handleEscalations(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	filterRig := r.URL.Query().Get("rig")
	sortBy := r.URL.Query().Get("sort")
	if sortBy == "" {
		sortBy = "priority"
	}
	data := escalationsData{GeneratedAt: now, FilterRig: filterRig, SortBy: sortBy}

	if s.ds == nil {
		s.render(w, r, "escalations", data)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("escalations: list dbs: %v", err)
		data.Err = err.Error()
		s.render(w, r, "escalations", data)
		return
	}

	var mu sync.Mutex
	var wg sync.WaitGroup

	thirtyDaysAgo := now.Add(-30 * 24 * time.Hour)

	for _, db := range dbs {
		if filterRig != "" && db.Name != filterRig {
			continue
		}
		wg.Add(1)
		go func(dbName string) {
			defer wg.Done()

			assignees, _ := s.ds.DistinctAssignees(ctx, dbName)
			if len(assignees) > 0 {
				mu.Lock()
				data.Assignees = append(data.Assignees, assignees...)
				mu.Unlock()
			}

			issues, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{Limit: 2000})
			if err != nil {
				log.Printf("escalations %s: %v", dbName, err)
				return
			}

			for _, iss := range issues {
				if isNoise(iss.ID, iss.Title) {
					continue
				}
				if iss.Status == "closed" || iss.Status == "deferred" {
					continue
				}

				// Check status history for priority changes
				// StatusHistory tracks status, but we need to detect priority increases
				// We can infer from metadata or check the current state
				// For now: look at beads created as lower priority that are now higher
				history, err := s.ds.StatusHistory(ctx, dbName, iss.ID)
				if err != nil || len(history) < 2 {
					continue
				}

				// StatusHistory shows status changes, not priority changes
				// However, we can detect escalation by looking at issues
				// that have been updated recently and have high priority
				// This is an approximation — look for recent high-pri updates
				if iss.Priority <= 1 && iss.UpdatedAt.After(thirtyDaysAgo) && iss.CreatedAt.Before(iss.UpdatedAt.Add(-24*time.Hour)) {
					// Check if this has had multiple status transitions (sign of activity/escalation)
					if len(history) >= 3 {
						mu.Lock()
						data.Escalations = append(data.Escalations, escalatedBead{
							ID:          iss.ID,
							Title:       iss.Title,
							Status:      iss.Status,
							Priority:    iss.Priority,
							Assignee:    iss.Assignee,
							DB:          dbName,
							PriorityNew: iss.Priority,
							EscalatedAt: iss.UpdatedAt,
						})
						if iss.Priority == 0 {
							data.ToP0++
						} else if iss.Priority == 1 {
							data.ToP1++
						}
						mu.Unlock()
					}
				}
			}
		}(db.Name)
	}
	wg.Wait()

	sort.Strings(data.Assignees)
	data.TotalEsc = len(data.Escalations)

	switch sortBy {
	case "date":
		sort.Slice(data.Escalations, func(i, j int) bool {
			return data.Escalations[i].EscalatedAt.After(data.Escalations[j].EscalatedAt)
		})
	case "assignee":
		sort.Slice(data.Escalations, func(i, j int) bool {
			if data.Escalations[i].Assignee != data.Escalations[j].Assignee {
				return data.Escalations[i].Assignee < data.Escalations[j].Assignee
			}
			return data.Escalations[i].Priority < data.Escalations[j].Priority
		})
	case "status":
		sort.Slice(data.Escalations, func(i, j int) bool {
			if data.Escalations[i].Status != data.Escalations[j].Status {
				return data.Escalations[i].Status < data.Escalations[j].Status
			}
			return data.Escalations[i].Priority < data.Escalations[j].Priority
		})
	default: // "priority"
		sort.Slice(data.Escalations, func(i, j int) bool {
			if data.Escalations[i].Priority != data.Escalations[j].Priority {
				return data.Escalations[i].Priority < data.Escalations[j].Priority
			}
			return data.Escalations[i].EscalatedAt.After(data.Escalations[j].EscalatedAt)
		})
	}

	if len(data.Escalations) > 50 {
		data.Escalations = data.Escalations[:50]
	}

	s.render(w, r, "escalations", data)
}
