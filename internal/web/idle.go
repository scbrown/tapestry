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

type idleAgent struct {
	Name       string
	LastUpdate time.Time
	IdleDays   int
	OpenBeads  int
	InProgress int
}

type idleData struct {
	GeneratedAt time.Time
	FilterRig   string
	SortBy      string
	Rigs        []string

	// Agents idle for 3+ days
	IdleAgents []idleAgent

	// Active agents (for comparison)
	ActiveCount int
	IdleCount   int
	TotalAgents int

	ThresholdDays int

	Err string
}

func (s *Server) handleIdle(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	filterRig := r.URL.Query().Get("rig")
	sortBy := r.URL.Query().Get("sort")
	if sortBy == "" {
		sortBy = "idle"
	}
	data := idleData{GeneratedAt: now, FilterRig: filterRig, SortBy: sortBy, ThresholdDays: 3}

	if s.ds == nil {
		s.render(w, r, "idle", data)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("idle: list dbs: %v", err)
		data.Err = err.Error()
		s.render(w, r, "idle", data)
		return
	}

	for _, db := range dbs {
		data.Rigs = append(data.Rigs, db.Name)
	}
	sort.Strings(data.Rigs)

	threshold := now.Add(-time.Duration(data.ThresholdDays) * 24 * time.Hour)

	agentData := make(map[string]*idleAgent)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, db := range dbs {
		if filterRig != "" && db.Name != filterRig {
			continue
		}
		wg.Add(1)
		go func(dbName string) {
			defer wg.Done()

			issues, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{Limit: 2000})
			if err != nil {
				log.Printf("idle %s: %v", dbName, err)
				return
			}

			mu.Lock()
			for _, iss := range issues {
				if isNoise(iss.ID, iss.Title) || iss.Assignee == "" {
					continue
				}
				if iss.Status == "closed" || iss.Status == "deferred" {
					// Track last update from closed items too
					a, ok := agentData[iss.Assignee]
					if !ok {
						a = &idleAgent{Name: iss.Assignee}
						agentData[iss.Assignee] = a
					}
					if iss.UpdatedAt.After(a.LastUpdate) {
						a.LastUpdate = iss.UpdatedAt
					}
					continue
				}

				a, ok := agentData[iss.Assignee]
				if !ok {
					a = &idleAgent{Name: iss.Assignee}
					agentData[iss.Assignee] = a
				}
				a.OpenBeads++
				if iss.Status == "in_progress" || iss.Status == "hooked" {
					a.InProgress++
				}
				if iss.UpdatedAt.After(a.LastUpdate) {
					a.LastUpdate = iss.UpdatedAt
				}
			}
			mu.Unlock()
		}(db.Name)
	}
	wg.Wait()

	data.TotalAgents = len(agentData)

	for _, a := range agentData {
		if !a.LastUpdate.IsZero() {
			a.IdleDays = int(now.Sub(a.LastUpdate).Hours() / 24)
		}
		if a.LastUpdate.Before(threshold) {
			data.IdleAgents = append(data.IdleAgents, *a)
			data.IdleCount++
		} else {
			data.ActiveCount++
		}
	}

	switch sortBy {
	case "name":
		sort.Slice(data.IdleAgents, func(i, j int) bool {
			return data.IdleAgents[i].Name < data.IdleAgents[j].Name
		})
	case "open":
		sort.Slice(data.IdleAgents, func(i, j int) bool {
			return data.IdleAgents[i].OpenBeads > data.IdleAgents[j].OpenBeads
		})
	case "activity":
		sort.Slice(data.IdleAgents, func(i, j int) bool {
			return data.IdleAgents[i].LastUpdate.After(data.IdleAgents[j].LastUpdate)
		})
	default: // "idle"
		sort.Slice(data.IdleAgents, func(i, j int) bool {
			return data.IdleAgents[i].IdleDays > data.IdleAgents[j].IdleDays
		})
	}

	s.render(w, r, "idle", data)
}
