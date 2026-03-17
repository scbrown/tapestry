package web

import (
	"log"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/scbrown/tapestry/internal/dolt"
)

type throughputWeek struct {
	WeekStart string // "Mar 03"
	Created   int
	Closed    int
	Net       int // Created - Closed (positive = growing backlog)
}

type throughputData struct {
	GeneratedAt  time.Time
	Weeks        []throughputWeek
	AvgCreated   int
	AvgClosed    int
	TotalCreated int
	TotalClosed  int
	MaxCount     int // for bar scaling
	WeekCount    int
	Rigs         []string // available rigs for filter
	FilterRig    string   // current rig filter
}

func (s *Server) handleThroughput(w http.ResponseWriter, r *http.Request) {
	data := throughputData{GeneratedAt: time.Now(), WeekCount: 12}

	if s.ds == nil {
		s.render(w, r, "throughput", data)
		return
	}

	ctx := r.Context()
	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("throughput: list dbs: %v", err)
		s.render(w, r, "throughput", data)
		return
	}

	// Build 12 weekly buckets ending at current week
	now := time.Now()
	// Find start of current week (Monday)
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	currentWeekStart := time.Date(now.Year(), now.Month(), now.Day()-(weekday-1), 0, 0, 0, 0, time.Local)

	type weekBucket struct {
		start time.Time
		end   time.Time
	}
	buckets := make([]weekBucket, data.WeekCount)
	for i := 0; i < data.WeekCount; i++ {
		offset := data.WeekCount - 1 - i
		start := currentWeekStart.AddDate(0, 0, -7*offset)
		end := start.AddDate(0, 0, 7)
		buckets[i] = weekBucket{start: start, end: end}
	}

	rangeStart := buckets[0].start
	rangeEnd := buckets[len(buckets)-1].end

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
				log.Printf("throughput: %s: %v", dbName, err)
				return
			}
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

	// Count created/closed per week
	created := make([]int, data.WeekCount)
	closed := make([]int, data.WeekCount)

	for _, r := range results {
		if filterRig != "" && r.rig != filterRig {
			continue
		}
		for _, iss := range r.issues {
			if !iss.CreatedAt.Before(rangeStart) && iss.CreatedAt.Before(rangeEnd) {
				for wi, b := range buckets {
					if !iss.CreatedAt.Before(b.start) && iss.CreatedAt.Before(b.end) {
						created[wi]++
						break
					}
				}
			}
			if iss.Status == "closed" && !iss.UpdatedAt.Before(rangeStart) && iss.UpdatedAt.Before(rangeEnd) {
				for wi, b := range buckets {
					if !iss.UpdatedAt.Before(b.start) && iss.UpdatedAt.Before(b.end) {
						closed[wi]++
						break
					}
				}
			}
		}
	}

	for i, b := range buckets {
		week := throughputWeek{
			WeekStart: b.start.Format("Jan 02"),
			Created:   created[i],
			Closed:    closed[i],
			Net:       created[i] - closed[i],
		}
		data.Weeks = append(data.Weeks, week)
		data.TotalCreated += created[i]
		data.TotalClosed += closed[i]
		if created[i] > data.MaxCount {
			data.MaxCount = created[i]
		}
		if closed[i] > data.MaxCount {
			data.MaxCount = closed[i]
		}
	}

	if data.WeekCount > 0 {
		data.AvgCreated = data.TotalCreated / data.WeekCount
		data.AvgClosed = data.TotalClosed / data.WeekCount
	}

	s.render(w, r, "throughput", data)
}
