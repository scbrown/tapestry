package web

import (
	"log"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/scbrown/tapestry/internal/dolt"
)

type ageBand struct {
	Label   string
	Count   int
	Percent int // 0-100
	Color   string
}

type ageBreakdownData struct {
	GeneratedAt time.Time
	Bands       []ageBand
	TotalOpen   int
	MedianAge   int
	MeanAge     int
	MaxAge      int
	OldestID    string
	OldestTitle string
	OldestRig   string
	Rigs        []string // available rigs for filter
	FilterRig   string   // current rig filter
}

func (s *Server) handleAgeBreakdown(w http.ResponseWriter, r *http.Request) {
	data := ageBreakdownData{GeneratedAt: time.Now()}

	if s.ds == nil {
		s.render(w, r, "age-breakdown", data)
		return
	}

	ctx := r.Context()
	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("age-breakdown: list dbs: %v", err)
		s.render(w, r, "age-breakdown", data)
		return
	}

	type issueAge struct {
		rig   string
		issue dolt.Issue
		age   int
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
			issues, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{Status: "open", Limit: 5000})
			if err != nil {
				log.Printf("age-breakdown: %s: %v", dbName, err)
				return
			}
			prog, _ := s.ds.Issues(ctx, dbName, dolt.IssueFilter{Status: "in_progress", Limit: 5000})
			issues = append(issues, prog...)
			results[i] = dbResult{rig: dbName, issues: issues}
		}(i, db.Name)
	}
	wg.Wait()

	// Collect distinct rigs for filter
	rigSet := make(map[string]bool)
	for _, r := range results {
		if len(r.issues) > 0 {
			rigSet[r.rig] = true
		}
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

	now := time.Now()
	var ages []issueAge
	for _, r := range results {
		if filterRig != "" && r.rig != filterRig {
			continue
		}
		for _, iss := range r.issues {
			age := int(now.Sub(iss.CreatedAt).Hours() / 24)
			ages = append(ages, issueAge{rig: r.rig, issue: iss, age: age})
		}
	}

	data.TotalOpen = len(ages)
	if data.TotalOpen == 0 {
		data.Bands = defaultBands(0)
		s.render(w, r, "age-breakdown", data)
		return
	}

	// Compute stats
	totalAge := 0
	for _, a := range ages {
		totalAge += a.age
		if a.age > data.MaxAge {
			data.MaxAge = a.age
			data.OldestID = a.issue.ID
			data.OldestTitle = a.issue.Title
			data.OldestRig = a.rig
		}
	}
	data.MeanAge = totalAge / data.TotalOpen

	// Sort for median
	sorted := make([]int, len(ages))
	for i, a := range ages {
		sorted[i] = a.age
	}
	// Simple insertion sort (good enough for small N)
	for i := 1; i < len(sorted); i++ {
		key := sorted[i]
		j := i - 1
		for j >= 0 && sorted[j] > key {
			sorted[j+1] = sorted[j]
			j--
		}
		sorted[j+1] = key
	}
	data.MedianAge = sorted[len(sorted)/2]

	// Count by band
	counts := [5]int{} // 0-7, 7-30, 30-90, 90-180, 180+
	for _, a := range ages {
		switch {
		case a.age < 7:
			counts[0]++
		case a.age < 30:
			counts[1]++
		case a.age < 90:
			counts[2]++
		case a.age < 180:
			counts[3]++
		default:
			counts[4]++
		}
	}

	labels := [5]string{"< 1 week", "1-4 weeks", "1-3 months", "3-6 months", "6+ months"}
	colors := [5]string{"var(--green)", "var(--cyan)", "var(--yellow)", "var(--orange)", "var(--red)"}
	for i := 0; i < 5; i++ {
		pct := 0
		if data.TotalOpen > 0 {
			pct = counts[i] * 100 / data.TotalOpen
		}
		data.Bands = append(data.Bands, ageBand{
			Label: labels[i], Count: counts[i], Percent: pct, Color: colors[i],
		})
	}

	s.render(w, r, "age-breakdown", data)
}

func defaultBands(total int) []ageBand {
	labels := [5]string{"< 1 week", "1-4 weeks", "1-3 months", "3-6 months", "6+ months"}
	colors := [5]string{"var(--green)", "var(--cyan)", "var(--yellow)", "var(--orange)", "var(--red)"}
	bands := make([]ageBand, 5)
	for i := 0; i < 5; i++ {
		bands[i] = ageBand{Label: labels[i], Color: colors[i]}
	}
	return bands
}
