package web

import (
	"log"
	"net/http"
	"sync"
	"time"
)

type heatmapDay struct {
	Date    time.Time
	Created int
	Closed  int
	Total   int
	Weekday int // 0=Sun
}

type heatmapWeek struct {
	Days [7]*heatmapDay // indexed by weekday
}

type heatmapData struct {
	Weeks    []heatmapWeek
	MaxDay   int
	Total    int
	NumDays  int
	AvgDay   float64
	Err      string
}

func (s *Server) handleHeatmap(w http.ResponseWriter, r *http.Request) {
	if s.ds == nil {
		s.render(w, r, "heatmap", heatmapData{})
		return
	}

	ctx := r.Context()
	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("heatmap: list dbs: %v", err)
		s.render(w, r, "heatmap", heatmapData{Err: err.Error()})
		return
	}

	const numDays = 91 // ~13 weeks
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	type dayResult struct {
		created int
		closed  int
	}

	// Fetch per-day counts across all databases in parallel
	dayResults := make([]dayResult, numDays)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, db := range dbs {
		wg.Add(1)
		go func(dbName string) {
			defer wg.Done()
			for d := 0; d < numDays; d++ {
				dayStart := todayStart.AddDate(0, 0, -d)
				dayEnd := dayStart.AddDate(0, 0, 1)

				created, err := s.ds.CountCreatedInRange(ctx, dbName, dayStart, dayEnd)
				if err != nil {
					continue
				}
				closed, err := s.ds.CountClosedInRange(ctx, dbName, dayStart, dayEnd)
				if err != nil {
					continue
				}

				mu.Lock()
				dayResults[d].created += created
				dayResults[d].closed += closed
				mu.Unlock()
			}
		}(db.Name)
	}
	wg.Wait()

	// Build week/day grid
	maxDay := 0
	total := 0
	var allDays []heatmapDay
	for d := numDays - 1; d >= 0; d-- {
		date := todayStart.AddDate(0, 0, -d)
		day := heatmapDay{
			Date:    date,
			Created: dayResults[d].created,
			Closed:  dayResults[d].closed,
			Total:   dayResults[d].created + dayResults[d].closed,
			Weekday: int(date.Weekday()),
		}
		if day.Total > maxDay {
			maxDay = day.Total
		}
		total += day.Total
		allDays = append(allDays, day)
	}

	// Group into weeks
	var weeks []heatmapWeek
	var currentWeek heatmapWeek
	for i := range allDays {
		wd := allDays[i].Weekday
		currentWeek.Days[wd] = &allDays[i]
		if wd == 6 || i == len(allDays)-1 {
			weeks = append(weeks, currentWeek)
			currentWeek = heatmapWeek{}
		}
	}

	avgDay := 0.0
	if numDays > 0 {
		avgDay = float64(total) / float64(numDays)
	}

	s.render(w, r, "heatmap", heatmapData{
		Weeks:  weeks,
		MaxDay: maxDay,
		Total:  total,
		NumDays: numDays,
		AvgDay: avgDay,
	})
}
