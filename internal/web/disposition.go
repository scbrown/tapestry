package web

import (
	"log"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/scbrown/tapestry/internal/dolt"
)

type dispositionWeek struct {
	WeekStart string // "Mar 03"
	Closed    int
	Deferred  int
	Blocked   int
	Open      int // still open (created this week but not closed/deferred)
	Total     int
}

type dispositionData struct {
	GeneratedAt    time.Time
	Weeks          []dispositionWeek
	TotalClosed    int
	TotalDeferred  int
	TotalBlocked   int
	TotalOpen      int
	CloseRate      float64 // pct of resolved items that were closed (vs deferred)
	MaxWeekTotal   int
	WeekCount      int
	Rigs           []string
	FilterRig      string
}

func (s *Server) handleDisposition(w http.ResponseWriter, r *http.Request) {
	data := dispositionData{GeneratedAt: time.Now(), WeekCount: 8}

	if s.ds == nil {
		s.render(w, r, "disposition", data)
		return
	}

	ctx := r.Context()
	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("disposition: list dbs: %v", err)
		s.render(w, r, "disposition", data)
		return
	}

	now := time.Now()
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	currentWeekStart := time.Date(now.Year(), now.Month(), now.Day()-(weekday-1), 0, 0, 0, 0, time.Local)

	type weekBucket struct {
		start, end time.Time
	}
	buckets := make([]weekBucket, data.WeekCount)
	for i := 0; i < data.WeekCount; i++ {
		offset := data.WeekCount - 1 - i
		start := currentWeekStart.AddDate(0, 0, -7*offset)
		buckets[i] = weekBucket{start: start, end: start.AddDate(0, 0, 7)}
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
			issues, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{Limit: 10000})
			if err != nil {
				log.Printf("disposition: %s: %v", dbName, err)
				return
			}
			results[i] = dbResult{rig: dbName, issues: issues}
		}(i, db.Name)
	}
	wg.Wait()

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

	filterRig := r.URL.Query().Get("rig")
	data.FilterRig = filterRig

	// For each issue updated in range, bucket by its current status
	closed := make([]int, data.WeekCount)
	deferred := make([]int, data.WeekCount)
	blocked := make([]int, data.WeekCount)
	open := make([]int, data.WeekCount)

	rangeStart := buckets[0].start

	for _, r := range results {
		if filterRig != "" && r.rig != filterRig {
			continue
		}
		for _, iss := range r.issues {
			if iss.UpdatedAt.Before(rangeStart) {
				continue
			}
			for wi, b := range buckets {
				if !iss.UpdatedAt.Before(b.start) && iss.UpdatedAt.Before(b.end) {
					switch iss.Status {
					case "closed":
						closed[wi]++
					case "deferred":
						deferred[wi]++
					case "blocked":
						blocked[wi]++
					default:
						open[wi]++
					}
					break
				}
			}
		}
	}

	for i, b := range buckets {
		total := closed[i] + deferred[i] + blocked[i] + open[i]
		week := dispositionWeek{
			WeekStart: b.start.Format("Jan 02"),
			Closed:    closed[i],
			Deferred:  deferred[i],
			Blocked:   blocked[i],
			Open:      open[i],
			Total:     total,
		}
		data.Weeks = append(data.Weeks, week)
		data.TotalClosed += closed[i]
		data.TotalDeferred += deferred[i]
		data.TotalBlocked += blocked[i]
		data.TotalOpen += open[i]
		if total > data.MaxWeekTotal {
			data.MaxWeekTotal = total
		}
	}

	resolved := data.TotalClosed + data.TotalDeferred
	if resolved > 0 {
		data.CloseRate = float64(data.TotalClosed) / float64(resolved) * 100
	}

	s.render(w, r, "disposition", data)
}
