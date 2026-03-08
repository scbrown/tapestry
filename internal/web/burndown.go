package web

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"
)

type burndownDay struct {
	Date    time.Time
	Open    int // total non-closed, non-deferred at end of day
	Created int
	Closed  int
}

type burndownData struct {
	GeneratedAt time.Time
	Days        []burndownDay
	Period      int
	MaxOpen     int
	StartOpen   int
	EndOpen     int
	Delta       int // EndOpen - StartOpen (negative = progress)
}

func (s *Server) handleBurndown(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	data := burndownData{GeneratedAt: now, Period: 30}

	if s.ds == nil {
		s.render(w, r, "burndown", data)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("burndown: list dbs: %v", err)
		s.render(w, r, "burndown", data)
		return
	}

	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	numDays := data.Period

	type dbResult struct {
		days [30]struct{ created, closed int }
	}

	results := make([]dbResult, len(dbs))
	var wg sync.WaitGroup
	for i, db := range dbs {
		wg.Add(1)
		go func(i int, dbName string) {
			defer wg.Done()
			var r dbResult
			for d := 0; d < numDays; d++ {
				dayStart := todayStart.AddDate(0, 0, -d)
				dayEnd := dayStart.AddDate(0, 0, 1)
				created, err := s.ds.CountCreatedInRange(ctx, dbName, dayStart, dayEnd)
				if err == nil {
					r.days[d].created = created
				}
				closed, err := s.ds.CountClosedInRange(ctx, dbName, dayStart, dayEnd)
				if err == nil {
					r.days[d].closed = closed
				}
			}
			results[i] = r
		}(i, db.Name)
	}
	wg.Wait()

	// Build the burndown: start from the oldest day and track cumulative open count.
	// We approximate: open(t) = open(t-1) + created(t) - closed(t)
	// Seed with 0, so the chart shows relative change.
	var cumOpen int
	for d := numDays - 1; d >= 0; d-- {
		day := burndownDay{Date: todayStart.AddDate(0, 0, -d)}
		for _, r := range results {
			day.Created += r.days[d].created
			day.Closed += r.days[d].closed
		}
		cumOpen += day.Created - day.Closed
		day.Open = cumOpen

		data.Days = append(data.Days, day)

		if day.Open > data.MaxOpen || len(data.Days) == 1 {
			data.MaxOpen = day.Open
		}
	}

	if len(data.Days) > 0 {
		data.StartOpen = data.Days[0].Open
		data.EndOpen = data.Days[len(data.Days)-1].Open
		data.Delta = data.EndOpen - data.StartOpen
	}

	s.render(w, r, "burndown", data)
}
