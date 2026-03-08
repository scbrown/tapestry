package web

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"
)

type trendWeek struct {
	Start   time.Time
	End     time.Time
	Created int
	Closed  int
	Net     int // Created - Closed
}

type trendsData struct {
	GeneratedAt  time.Time
	Weeks        []trendWeek
	MaxCreated   int
	MaxClosed    int
	TotalCreated int
	TotalClosed  int
	AvgCreated   float64
	AvgClosed    float64
	Trend        string // "improving", "stable", "growing"
}

func (s *Server) handleTrends(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	data := trendsData{GeneratedAt: now}

	if s.ds == nil {
		s.render(w, r, "trends", data)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("trends: list dbs: %v", err)
		s.render(w, r, "trends", data)
		return
	}

	numWeeks := 8
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	// Find the start of this week (Monday)
	weekday := int(todayStart.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	thisMonday := todayStart.AddDate(0, 0, -(weekday - 1))

	type weekCounts struct {
		created, closed int
	}

	type dbResult struct {
		weeks [8]weekCounts
	}

	results := make([]dbResult, len(dbs))
	var wg sync.WaitGroup
	for i, db := range dbs {
		wg.Add(1)
		go func(i int, dbName string) {
			defer wg.Done()
			var r dbResult
			for w := 0; w < numWeeks; w++ {
				weekStart := thisMonday.AddDate(0, 0, -w*7)
				weekEnd := weekStart.AddDate(0, 0, 7)
				if weekEnd.After(now) {
					weekEnd = now
				}
				created, err := s.ds.CountCreatedInRange(ctx, dbName, weekStart, weekEnd)
				if err == nil {
					r.weeks[w].created = created
				}
				closed, err := s.ds.CountClosedInRange(ctx, dbName, weekStart, weekEnd)
				if err == nil {
					r.weeks[w].closed = closed
				}
			}
			results[i] = r
		}(i, db.Name)
	}
	wg.Wait()

	// Aggregate (oldest first)
	for w := numWeeks - 1; w >= 0; w-- {
		weekStart := thisMonday.AddDate(0, 0, -w*7)
		weekEnd := weekStart.AddDate(0, 0, 7)
		tw := trendWeek{Start: weekStart, End: weekEnd}
		for _, r := range results {
			tw.Created += r.weeks[w].created
			tw.Closed += r.weeks[w].closed
		}
		tw.Net = tw.Created - tw.Closed
		data.Weeks = append(data.Weeks, tw)

		data.TotalCreated += tw.Created
		data.TotalClosed += tw.Closed
		if tw.Created > data.MaxCreated {
			data.MaxCreated = tw.Created
		}
		if tw.Closed > data.MaxClosed {
			data.MaxClosed = tw.Closed
		}
	}

	if numWeeks > 0 {
		data.AvgCreated = float64(data.TotalCreated) / float64(numWeeks)
		data.AvgClosed = float64(data.TotalClosed) / float64(numWeeks)
	}

	// Determine trend: compare last 4 weeks net vs first 4 weeks net
	if len(data.Weeks) >= 8 {
		var earlyNet, lateNet int
		for i := 0; i < 4; i++ {
			earlyNet += data.Weeks[i].Net
		}
		for i := 4; i < 8; i++ {
			lateNet += data.Weeks[i].Net
		}
		if lateNet < earlyNet-5 {
			data.Trend = "improving"
		} else if lateNet > earlyNet+5 {
			data.Trend = "growing"
		} else {
			data.Trend = "stable"
		}
	}

	s.render(w, r, "trends", data)
}
