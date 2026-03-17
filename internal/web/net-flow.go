package web

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/scbrown/tapestry/internal/dolt"
)

type netFlowDay struct {
	Date      string // "Mar 02"
	Created   int
	Closed    int
	Net       int // running cumulative: total open at end of day
	DailyNet  int // today's created - closed
}

type netFlowData struct {
	GeneratedAt time.Time
	Days        []netFlowDay
	CurrentOpen int
	Trend       string // "growing", "shrinking", "stable"
	DayCount    int
	MaxOpen     int
	MinOpen     int
}

func (s *Server) handleNetFlow(w http.ResponseWriter, r *http.Request) {
	data := netFlowData{GeneratedAt: time.Now(), DayCount: 30}

	if s.ds == nil {
		s.render(w, r, "net-flow", data)
		return
	}

	ctx := r.Context()
	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("net-flow: list dbs: %v", err)
		s.render(w, r, "net-flow", data)
		return
	}

	type dbResult struct {
		issues []dolt.Issue
	}
	results := make([]dbResult, len(dbs))
	var wg sync.WaitGroup
	for i, db := range dbs {
		wg.Add(1)
		go func(i int, dbName string) {
			defer wg.Done()
			issues, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{Limit: 10000})
			if err != nil {
				log.Printf("net-flow: %s: %v", dbName, err)
				return
			}
			results[i] = dbResult{issues: issues}
		}(i, db.Name)
	}
	wg.Wait()

	// Build daily buckets for last 30 days
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	startDate := today.AddDate(0, 0, -(data.DayCount - 1))

	created := make([]int, data.DayCount)
	closed := make([]int, data.DayCount)

	// Count beads that existed before the window (baseline)
	baseline := 0
	for _, r := range results {
		for _, iss := range r.issues {
			if iss.CreatedAt.Before(startDate) && (iss.Status != "closed" || !iss.UpdatedAt.Before(startDate)) {
				baseline++
			}
			// Created during window
			if !iss.CreatedAt.Before(startDate) {
				dayIdx := int(iss.CreatedAt.Sub(startDate).Hours() / 24)
				if dayIdx >= 0 && dayIdx < data.DayCount {
					created[dayIdx]++
				}
			}
			// Closed during window
			if iss.Status == "closed" && !iss.UpdatedAt.Before(startDate) {
				dayIdx := int(iss.UpdatedAt.Sub(startDate).Hours() / 24)
				if dayIdx >= 0 && dayIdx < data.DayCount {
					closed[dayIdx]++
				}
			}
		}
	}

	// Build cumulative series
	running := baseline
	for i := 0; i < data.DayCount; i++ {
		dayDate := startDate.AddDate(0, 0, i)
		running += created[i] - closed[i]
		day := netFlowDay{
			Date:     dayDate.Format("Jan 02"),
			Created:  created[i],
			Closed:   closed[i],
			Net:      running,
			DailyNet: created[i] - closed[i],
		}
		data.Days = append(data.Days, day)

		if i == 0 || running > data.MaxOpen {
			data.MaxOpen = running
		}
		if i == 0 || running < data.MinOpen {
			data.MinOpen = running
		}
	}

	data.CurrentOpen = running

	// Determine trend from first half vs second half
	if len(data.Days) >= 2 {
		firstHalf := data.Days[0].Net
		secondHalf := data.Days[len(data.Days)-1].Net
		diff := secondHalf - firstHalf
		switch {
		case diff > 5:
			data.Trend = "growing"
		case diff < -5:
			data.Trend = "shrinking"
		default:
			data.Trend = "stable"
		}
	}

	s.render(w, r, "net-flow", data)
}
