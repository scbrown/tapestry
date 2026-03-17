package web

import (
	"context"
	"log"
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/scbrown/tapestry/internal/dolt"
)

type funnelStage struct {
	Name    string
	Count   int
	Pct     int // percentage of total
	DropPct int // drop from previous stage
}

type funnelData struct {
	GeneratedAt time.Time
	Stages      []funnelStage
	Total       int

	// Median ages by current status (days)
	MedianOpenAge      int
	MedianProgressAge  int
	MedianClosedAge    int

	// Conversion rates
	TriageRate   int // % that got assigned or started (not sitting as open/unassigned)
	StartRate    int // % that moved to in_progress at some point
	CloseRate    int // % that reached closed

	Rigs      []string
	FilterRig string
	Err       string
}

func (s *Server) handleFunnel(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	data := funnelData{GeneratedAt: now}

	if s.ds == nil {
		s.render(w, r, "funnel", data)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("funnel: list dbs: %v", err)
		s.render(w, r, "funnel", funnelData{Err: err.Error(), GeneratedAt: now})
		return
	}

	type dbResult struct {
		totalFiled    int
		assigned      int // has assignee
		started       int // in_progress or was in_progress (now closed)
		closed        int
		blocked       int
		deferred      int
		openAges      []int // age in days for open items
		progressAges  []int
		closedAges    []int
	}

	results := make([]dbResult, len(dbs))
	var wg sync.WaitGroup
	for i, db := range dbs {
		wg.Add(1)
		go func(i int, dbName string) {
			defer wg.Done()
			var r dbResult

			issues, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{Limit: 1000})
			if err != nil {
				log.Printf("funnel %s: %v", dbName, err)
				results[i] = r
				return
			}

			for _, iss := range issues {
				if isNoise(iss.ID, iss.Title) {
					continue
				}
				r.totalFiled++
				age := int(now.Sub(iss.CreatedAt).Hours() / 24)

				if iss.Assignee != "" {
					r.assigned++
				}

				switch iss.Status {
				case "closed":
					r.closed++
					r.started++ // closed implies it was worked on
					r.closedAges = append(r.closedAges, age)
				case "in_progress", "hooked":
					r.started++
					r.progressAges = append(r.progressAges, age)
				case "blocked":
					r.blocked++
					r.started++ // blocked implies it was started
					r.progressAges = append(r.progressAges, age)
				case "deferred":
					r.deferred++
				case "open":
					r.openAges = append(r.openAges, age)
				}
			}

			// Also get closed issues not in the first query
			closedIssues, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{
				Status: "closed",
				Limit:  1000,
			})
			if err == nil {
				for _, iss := range closedIssues {
					if isNoise(iss.ID, iss.Title) {
						continue
					}
					// Only count those not already in the first query
					// (the first query gets all non-closed, this gets closed)
					r.totalFiled++
					r.closed++
					r.started++
					age := int(now.Sub(iss.CreatedAt).Hours() / 24)
					r.closedAges = append(r.closedAges, age)
					if iss.Assignee != "" {
						r.assigned++
					}
				}
			}

			results[i] = r
		}(i, db.Name)
	}
	wg.Wait()

	// Aggregate
	var totalFiled, assigned, started, closed, blocked, deferred int
	var allOpenAges, allProgressAges, allClosedAges []int
	rigSet := make(map[string]bool)

	for idx, r := range results {
		totalFiled += r.totalFiled
		assigned += r.assigned
		started += r.started
		closed += r.closed
		blocked += r.blocked
		deferred += r.deferred
		allOpenAges = append(allOpenAges, r.openAges...)
		allProgressAges = append(allProgressAges, r.progressAges...)
		allClosedAges = append(allClosedAges, r.closedAges...)
		if r.totalFiled > 0 {
			rigSet[dbs[idx].Name] = true
		}
	}

	data.Total = totalFiled

	// Build funnel stages
	data.Stages = []funnelStage{
		{Name: "Filed", Count: totalFiled, Pct: 100, DropPct: 0},
	}

	if totalFiled > 0 {
		assignedPct := pct(assigned, totalFiled)
		data.Stages = append(data.Stages, funnelStage{
			Name: "Assigned", Count: assigned, Pct: assignedPct, DropPct: 100 - assignedPct,
		})

		startedPct := pct(started, totalFiled)
		data.Stages = append(data.Stages, funnelStage{
			Name: "Started", Count: started, Pct: startedPct, DropPct: assignedPct - startedPct,
		})

		closedPct := pct(closed, totalFiled)
		data.Stages = append(data.Stages, funnelStage{
			Name: "Closed", Count: closed, Pct: closedPct, DropPct: startedPct - closedPct,
		})

		data.TriageRate = assignedPct
		data.StartRate = startedPct
		data.CloseRate = closedPct
	}

	data.MedianOpenAge = medianInt(allOpenAges)
	data.MedianProgressAge = medianInt(allProgressAges)
	data.MedianClosedAge = medianInt(allClosedAges)

	// Rigs
	for rig := range rigSet {
		data.Rigs = append(data.Rigs, rig)
	}

	s.render(w, r, "funnel", data)
}

func pct(num, denom int) int {
	if denom == 0 {
		return 0
	}
	return int(math.Round(float64(num) / float64(denom) * 100))
}

func medianInt(vals []int) int {
	if len(vals) == 0 {
		return 0
	}
	// Simple sort for median
	sorted := make([]int, len(vals))
	copy(sorted, vals)
	for i := 1; i < len(sorted); i++ {
		for j := i; j > 0 && sorted[j-1] > sorted[j]; j-- {
			sorted[j-1], sorted[j] = sorted[j], sorted[j-1]
		}
	}
	mid := len(sorted) / 2
	if len(sorted)%2 == 0 {
		return (sorted[mid-1] + sorted[mid]) / 2
	}
	return sorted[mid]
}
