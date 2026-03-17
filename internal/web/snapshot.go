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

type snapshotData struct {
	GeneratedAt time.Time
	FilterRig   string
	Rigs        []string

	// Counts
	TotalOpen   int
	TotalClosed int
	InProgress  int
	Blocked     int
	Deferred    int

	// Activity (7 days)
	CreatedWeek int
	ClosedWeek  int
	NetWeek     int

	// Health signals
	P0Open        int
	P1Open        int
	StaleCount    int // open > 14 days with no update
	UnassignedCnt int

	// Agent summary
	ActiveAgents int // agents with in_progress work
	TotalAgents  int // agents with any open work

	// Type breakdown
	BugCount  int
	TaskCount int
	EpicCount int

	Err string
}

func (s *Server) handleSnapshot(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	filterRig := r.URL.Query().Get("rig")
	data := snapshotData{GeneratedAt: now, FilterRig: filterRig}

	if s.ds == nil {
		s.render(w, r, "snapshot", data)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("snapshot: list dbs: %v", err)
		s.render(w, r, "snapshot", snapshotData{Err: err.Error(), GeneratedAt: now})
		return
	}

	for _, db := range dbs {
		data.Rigs = append(data.Rigs, db.Name)
	}
	sort.Strings(data.Rigs)

	weekAgo := now.Add(-7 * 24 * time.Hour)
	staleThreshold := now.Add(-14 * 24 * time.Hour)
	activeAgents := make(map[string]bool)
	allAgents := make(map[string]bool)

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
				log.Printf("snapshot %s: %v", dbName, err)
				return
			}

			var localOpen, localProgress, localBlocked, localDeferred int
			var localP0, localP1, localStale, localUnassigned int
			var localBug, localTask, localEpic int
			var localCreatedWeek int
			localActive := make(map[string]bool)
			localAll := make(map[string]bool)

			for _, iss := range issues {
				if isNoise(iss.ID, iss.Title) {
					continue
				}

				switch iss.Status {
				case "closed":
					continue
				case "in_progress", "hooked":
					localProgress++
					if iss.Assignee != "" {
						localActive[iss.Assignee] = true
					}
				case "blocked":
					localBlocked++
				case "deferred":
					localDeferred++
				default:
					localOpen++
				}

				if iss.Assignee != "" {
					localAll[iss.Assignee] = true
				} else {
					localUnassigned++
				}

				if iss.Priority == 0 {
					localP0++
				} else if iss.Priority == 1 {
					localP1++
				}

				if iss.UpdatedAt.Before(staleThreshold) && iss.Status != "deferred" {
					localStale++
				}

				if iss.CreatedAt.After(weekAgo) {
					localCreatedWeek++
				}

				switch iss.Type {
				case "bug":
					localBug++
				case "task":
					localTask++
				case "epic":
					localEpic++
				}
			}

			// Count closed in last week
			closedIssues, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{
				Status: "closed",
				Limit:  500,
			})
			var localClosedWeek int
			if err == nil {
				for _, iss := range closedIssues {
					if !isNoise(iss.ID, iss.Title) && iss.UpdatedAt.After(weekAgo) {
						localClosedWeek++
					}
				}
			}

			mu.Lock()
			data.TotalOpen += localOpen + localProgress + localBlocked
			data.InProgress += localProgress
			data.Blocked += localBlocked
			data.Deferred += localDeferred
			data.P0Open += localP0
			data.P1Open += localP1
			data.StaleCount += localStale
			data.UnassignedCnt += localUnassigned
			data.CreatedWeek += localCreatedWeek
			data.ClosedWeek += localClosedWeek
			data.BugCount += localBug
			data.TaskCount += localTask
			data.EpicCount += localEpic
			for a := range localActive {
				activeAgents[a] = true
			}
			for a := range localAll {
				allAgents[a] = true
			}
			mu.Unlock()
		}(db.Name)
	}
	wg.Wait()

	data.ActiveAgents = len(activeAgents)
	data.TotalAgents = len(allAgents)
	data.NetWeek = data.CreatedWeek - data.ClosedWeek

	s.render(w, r, "snapshot", data)
}
