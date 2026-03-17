package web

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/scbrown/tapestry/internal/dolt"
)

type resolutionBucket struct {
	Label string
	Count int
	Pct   int
}

type resolutionData struct {
	GeneratedAt  time.Time
	TotalClosed  int
	Within7d     resolutionBucket
	Within30d    resolutionBucket
	Within90d    resolutionBucket
	Over90d      resolutionBucket
	MedianHours  int
	FastestID    string
	FastestTitle string
	FastestHours int
	FastestRig   string
}

func (s *Server) handleResolutionRate(w http.ResponseWriter, r *http.Request) {
	data := resolutionData{GeneratedAt: time.Now()}

	if s.ds == nil {
		s.render(w, r, "resolution-rate", data)
		return
	}

	ctx := r.Context()
	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("resolution-rate: list dbs: %v", err)
		s.render(w, r, "resolution-rate", data)
		return
	}

	type closedIssue struct {
		rig   string
		issue dolt.Issue
		hours int
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
			issues, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{Status: "closed", Limit: 10000})
			if err != nil {
				log.Printf("resolution-rate: %s: %v", dbName, err)
				return
			}
			results[i] = dbResult{rig: dbName, issues: issues}
		}(i, db.Name)
	}
	wg.Wait()

	var items []closedIssue
	for _, r := range results {
		for _, iss := range r.issues {
			hours := int(iss.UpdatedAt.Sub(iss.CreatedAt).Hours())
			if hours < 0 {
				hours = 0
			}
			items = append(items, closedIssue{rig: r.rig, issue: iss, hours: hours})
		}
	}

	data.TotalClosed = len(items)
	if data.TotalClosed == 0 {
		s.render(w, r, "resolution-rate", data)
		return
	}

	// Count by resolution speed
	var w7, w30, w90, over int
	fastest := items[0]
	totalHours := 0

	for _, it := range items {
		totalHours += it.hours
		switch {
		case it.hours < 7*24:
			w7++
		case it.hours < 30*24:
			w30++
		case it.hours < 90*24:
			w90++
		default:
			over++
		}
		if it.hours < fastest.hours {
			fastest = it
		}
	}

	pct := func(n int) int { return n * 100 / data.TotalClosed }

	data.Within7d = resolutionBucket{Label: "< 7 days", Count: w7, Pct: pct(w7)}
	data.Within30d = resolutionBucket{Label: "7-30 days", Count: w30, Pct: pct(w30)}
	data.Within90d = resolutionBucket{Label: "30-90 days", Count: w90, Pct: pct(w90)}
	data.Over90d = resolutionBucket{Label: "> 90 days", Count: over, Pct: pct(over)}
	data.MedianHours = totalHours / data.TotalClosed
	data.FastestID = fastest.issue.ID
	data.FastestTitle = fastest.issue.Title
	data.FastestHours = fastest.hours
	data.FastestRig = fastest.rig

	s.render(w, r, "resolution-rate", data)
}
