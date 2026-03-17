package web

import (
	"log"
	"math"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/scbrown/tapestry/internal/dolt"
)

type responseTimeEntry struct {
	IssueID      string
	Title        string
	Rig          string
	Priority     int
	CreatedAt    time.Time
	FirstPickup  time.Time
	ResponseTime time.Duration
}

type priorityResponseTime struct {
	Priority   int
	MedianMins float64
	Count      int
}

type responseTimeData struct {
	GeneratedAt    time.Time
	Entries        []responseTimeEntry
	ByPriority     []priorityResponseTime
	MedianMins     float64
	MeanMins       float64
	P90Mins        float64
	Total          int
	NeverPickedUp  int
	Rigs           []string
	FilterRig      string
}

func (s *Server) handleResponseTime(w http.ResponseWriter, r *http.Request) {
	data := responseTimeData{GeneratedAt: time.Now()}

	if s.ds == nil {
		s.render(w, r, "response-time", data)
		return
	}

	ctx := r.Context()
	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("response-time: list dbs: %v", err)
		s.render(w, r, "response-time", data)
		return
	}

	// Get recent closed/in_progress issues (last 30 days)
	now := time.Now()
	cutoff := now.AddDate(0, 0, -30)

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
			issues, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{
				UpdatedAfter: cutoff,
				Limit:        2000,
			})
			if err != nil {
				log.Printf("response-time: %s: %v", dbName, err)
				return
			}
			results[i] = dbResult{rig: dbName, issues: issues}
		}(i, db.Name)
	}
	wg.Wait()

	filterRig := r.URL.Query().Get("rig")
	rigSet := make(map[string]bool)
	for _, r := range results {
		if len(r.issues) > 0 {
			rigSet[r.rig] = true
		}
	}
	for rig := range rigSet {
		data.Rigs = append(data.Rigs, rig)
	}
	sort.Strings(data.Rigs)
	data.FilterRig = filterRig

	// For issues that have been picked up, compute response time as:
	// time from creation to when status first changed to in_progress/hooked
	var durations []float64
	priorityDurations := map[int][]float64{}

	for _, r := range results {
		if filterRig != "" && r.rig != filterRig {
			continue
		}
		for _, iss := range r.issues {
			if iss.Status == "open" {
				data.NeverPickedUp++
				continue
			}

			// For issues that are in_progress, closed, or hooked,
			// approximate response time as time between created and first activity.
			// Without full status history in this handler (too expensive for all issues),
			// we use UpdatedAt - CreatedAt for closed issues that went through quickly,
			// and for in_progress we assume pickup happened around UpdatedAt.
			if iss.Status == "in_progress" || iss.Status == "hooked" || iss.Status == "closed" {
				// Only consider issues created in our window
				if iss.CreatedAt.Before(cutoff) {
					continue
				}
				rt := iss.UpdatedAt.Sub(iss.CreatedAt)
				if rt < 0 {
					rt = 0
				}

				entry := responseTimeEntry{
					IssueID:      iss.ID,
					Title:        iss.Title,
					Rig:          r.rig,
					Priority:     iss.Priority,
					CreatedAt:    iss.CreatedAt,
					FirstPickup:  iss.UpdatedAt,
					ResponseTime: rt,
				}
				data.Entries = append(data.Entries, entry)

				mins := rt.Minutes()
				durations = append(durations, mins)
				priorityDurations[iss.Priority] = append(priorityDurations[iss.Priority], mins)
			}
		}
	}

	// Sort entries by response time ascending (fastest first)
	sort.Slice(data.Entries, func(i, j int) bool {
		return data.Entries[i].ResponseTime < data.Entries[j].ResponseTime
	})

	// Cap display at 50
	if len(data.Entries) > 50 {
		data.Entries = data.Entries[:50]
	}

	data.Total = len(durations)

	if len(durations) > 0 {
		sort.Float64s(durations)
		data.MedianMins = percentile(durations, 50)
		data.P90Mins = percentile(durations, 90)

		var sum float64
		for _, d := range durations {
			sum += d
		}
		data.MeanMins = sum / float64(len(durations))
	}

	// By priority
	for p := 0; p <= 4; p++ {
		if ds, ok := priorityDurations[p]; ok && len(ds) > 0 {
			sort.Float64s(ds)
			data.ByPriority = append(data.ByPriority, priorityResponseTime{
				Priority:   p,
				MedianMins: percentile(ds, 50),
				Count:      len(ds),
			})
		}
	}

	s.render(w, r, "response-time", data)
}

func percentile(sorted []float64, pct float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	idx := pct / 100.0 * float64(len(sorted)-1)
	lower := int(math.Floor(idx))
	upper := int(math.Ceil(idx))
	if lower == upper || upper >= len(sorted) {
		return sorted[lower]
	}
	frac := idx - float64(lower)
	return sorted[lower]*(1-frac) + sorted[upper]*frac
}
