package web

import (
	"context"
	"log"
	"net/http"
	"sort"
	"sync"
	"time"
)

type burnupDay struct {
	Date       time.Time
	Closed     int // closed on this day
	Cumulative int // running total of closures
}

type burnupData struct {
	GeneratedAt time.Time
	Days        []burnupDay
	Period      int
	TotalClosed int
	MaxCum      int
	AvgPerDay   float64
	Rigs        []string
	FilterRig   string
}

func (s *Server) handleBurnup(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	data := burnupData{GeneratedAt: now, Period: 30}

	if s.ds == nil {
		s.render(w, r, "burnup", data)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("burnup: list dbs: %v", err)
		s.render(w, r, "burnup", data)
		return
	}

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

	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	numDays := data.Period

	type dbResult struct {
		days [30]int // closed per day
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
			var r dbResult
			for d := 0; d < numDays; d++ {
				dayStart := todayStart.AddDate(0, 0, -d)
				dayEnd := dayStart.AddDate(0, 0, 1)
				closed, err := s.ds.CountClosedInRange(ctx, dbName, dayStart, dayEnd)
				if err == nil {
					r.days[d] = closed
				}
			}
			results[i] = r
		}(i, db.Name)
	}
	wg.Wait()

	var cumulative int
	for d := numDays - 1; d >= 0; d-- {
		day := burnupDay{Date: todayStart.AddDate(0, 0, -d)}
		for _, r := range results {
			day.Closed += r.days[d]
		}
		cumulative += day.Closed
		day.Cumulative = cumulative
		data.Days = append(data.Days, day)
		data.TotalClosed += day.Closed
		if cumulative > data.MaxCum {
			data.MaxCum = cumulative
		}
	}

	if numDays > 0 {
		data.AvgPerDay = float64(data.TotalClosed) / float64(numDays)
	}

	s.render(w, r, "burnup", data)
}
