package web

import (
	"log"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/scbrown/tapestry/internal/dolt"
)

type cohortWeek struct {
	WeekLabel string // "Mar 03"
	Total     int
	Closed    int
	Open      int
	CloseRate int // 0-100
}

type cohortData struct {
	GeneratedAt    time.Time
	Weeks          []cohortWeek
	WeekCount      int
	OverallClose   int // overall close rate %
	TotalCreated   int
	TotalClosed    int
	Rigs           []string
	FilterRig      string
}

func (s *Server) handleCohort(w http.ResponseWriter, r *http.Request) {
	data := cohortData{GeneratedAt: time.Now(), WeekCount: 12}

	if s.ds == nil {
		s.render(w, r, "cohort", data)
		return
	}

	ctx := r.Context()
	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("cohort: list dbs: %v", err)
		s.render(w, r, "cohort", data)
		return
	}

	// Build rig list from all DBs, then filter
	filterRig := r.URL.Query().Get("rig")
	data.FilterRig = filterRig
	rigSet := make(map[string]bool)
	for _, db := range dbs {
		rigSet[db.Name] = true
	}
	for rig := range rigSet {
		data.Rigs = append(data.Rigs, rig)
	}
	sort.Strings(data.Rigs)

	type dbResult struct {
		rig    string
		issues []dolt.Issue
	}
	results := make([]dbResult, len(dbs))
	var wg sync.WaitGroup
	for i, db := range dbs {
		if filterRig != "" && db.Name != filterRig {
			continue
		}
		wg.Add(1)
		go func(i int, dbName string) {
			defer wg.Done()
			issues, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{Limit: 10000})
			if err != nil {
				log.Printf("cohort: %s: %v", dbName, err)
				return
			}
			results[i] = dbResult{rig: dbName, issues: issues}
		}(i, db.Name)
	}
	wg.Wait()

	// Build weekly cohort buckets
	now := time.Now()
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	currentWeekStart := time.Date(now.Year(), now.Month(), now.Day()-(weekday-1), 0, 0, 0, 0, time.Local)

	type bucket struct {
		start  time.Time
		end    time.Time
		total  int
		closed int
	}
	buckets := make([]bucket, data.WeekCount)
	for i := 0; i < data.WeekCount; i++ {
		offset := data.WeekCount - 1 - i
		start := currentWeekStart.AddDate(0, 0, -7*offset)
		end := start.AddDate(0, 0, 7)
		buckets[i] = bucket{start: start, end: end}
	}

	for _, r := range results {
		for _, iss := range r.issues {
			for bi := range buckets {
				if !iss.CreatedAt.Before(buckets[bi].start) && iss.CreatedAt.Before(buckets[bi].end) {
					buckets[bi].total++
					if iss.Status == "closed" {
						buckets[bi].closed++
					}
					break
				}
			}
		}
	}

	for _, b := range buckets {
		rate := 0
		if b.total > 0 {
			rate = b.closed * 100 / b.total
		}
		data.Weeks = append(data.Weeks, cohortWeek{
			WeekLabel: b.start.Format("Jan 02"),
			Total:     b.total,
			Closed:    b.closed,
			Open:      b.total - b.closed,
			CloseRate: rate,
		})
		data.TotalCreated += b.total
		data.TotalClosed += b.closed
	}

	if data.TotalCreated > 0 {
		data.OverallClose = data.TotalClosed * 100 / data.TotalCreated
	}

	s.render(w, r, "cohort", data)
}
