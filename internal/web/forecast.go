package web

import (
	"context"
	"log"
	"math"
	"net/http"
	"sync"
	"time"
)

type forecastWeek struct {
	Label   string
	Created int
	Closed  int
	Net     int
}

type forecastData struct {
	GeneratedAt time.Time

	// Current state
	BacklogSize int // open + in_progress + blocked
	OpenCount   int
	ActiveCount int
	BlockedCount int

	// Velocity (7-day)
	AvgCreated float64
	AvgClosed  float64
	NetRate    float64 // positive = growing, negative = shrinking

	// Forecast
	DaysToClear    int  // -1 if growing or no data
	BacklogGrowing bool
	ForecastLabel  string

	// 4-week history
	Weeks    []forecastWeek
	MaxWeek  int
}

func (s *Server) handleForecast(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	data := forecastData{GeneratedAt: now}

	if s.ds == nil {
		data.ForecastLabel = "No data source"
		s.render(w, r, "forecast", data)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 20*time.Second)
	defer cancel()

	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("forecast: list dbs: %v", err)
		s.render(w, r, "forecast", data)
		return
	}

	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	type dbResult struct {
		openCount    int
		activeCount  int
		blockedCount int
		weekData     [4]struct{ created, closed int }
	}

	results := make([]dbResult, len(dbs))
	var wg sync.WaitGroup
	for i, db := range dbs {
		wg.Add(1)
		go func(i int, dbName string) {
			defer wg.Done()
			var r dbResult

			counts, err := s.ds.CountByStatus(ctx, dbName)
			if err != nil {
				log.Printf("forecast: counts %s: %v", dbName, err)
				results[i] = r
				return
			}
			r.openCount = counts["open"]
			r.activeCount = counts["in_progress"] + counts["hooked"]
			r.blockedCount = counts["blocked"]

			// 4 weeks of data
			for w := 0; w < 4; w++ {
				weekEnd := todayStart.AddDate(0, 0, -w*7)
				weekStart := weekEnd.AddDate(0, 0, -7)
				if w == 0 {
					weekEnd = todayStart.AddDate(0, 0, 1) // include today
				}

				created, err := s.ds.CountCreatedInRange(ctx, dbName, weekStart, weekEnd)
				if err == nil {
					r.weekData[w].created = created
				}
				closed, err := s.ds.CountClosedInRange(ctx, dbName, weekStart, weekEnd)
				if err == nil {
					r.weekData[w].closed = closed
				}
			}

			results[i] = r
		}(i, db.Name)
	}
	wg.Wait()

	// Aggregate
	var weeks [4]struct{ created, closed int }
	for _, r := range results {
		data.OpenCount += r.openCount
		data.ActiveCount += r.activeCount
		data.BlockedCount += r.blockedCount
		for w := 0; w < 4; w++ {
			weeks[w].created += r.weekData[w].created
			weeks[w].closed += r.weekData[w].closed
		}
	}

	data.BacklogSize = data.OpenCount + data.ActiveCount + data.BlockedCount

	// Build weekly history (oldest first)
	weekLabels := []string{"3 weeks ago", "2 weeks ago", "Last week", "This week"}
	for w := 3; w >= 0; w-- {
		fw := forecastWeek{
			Label:   weekLabels[3-w],
			Created: weeks[w].created,
			Closed:  weeks[w].closed,
			Net:     weeks[w].created - weeks[w].closed,
		}
		data.Weeks = append(data.Weeks, fw)
		if fw.Created > data.MaxWeek {
			data.MaxWeek = fw.Created
		}
		if fw.Closed > data.MaxWeek {
			data.MaxWeek = fw.Closed
		}
	}

	// Calculate velocity from most recent 7 days
	recentCreated := weeks[0].created
	recentClosed := weeks[0].closed
	data.AvgCreated = float64(recentCreated) / 7.0
	data.AvgClosed = float64(recentClosed) / 7.0
	data.NetRate = data.AvgCreated - data.AvgClosed

	// Forecast
	if data.AvgClosed <= 0 {
		data.DaysToClear = -1
		data.ForecastLabel = "No closes recorded"
		data.BacklogGrowing = data.BacklogSize > 0
	} else if data.NetRate >= 0 {
		data.DaysToClear = -1
		data.BacklogGrowing = true
		data.ForecastLabel = "Backlog growing — close rate needs to exceed create rate"
	} else {
		// Net rate is negative (shrinking) — calculate days to clear
		netCloseRate := data.AvgClosed - data.AvgCreated
		days := math.Ceil(float64(data.BacklogSize) / netCloseRate)
		data.DaysToClear = int(days)
		data.BacklogGrowing = false

		if data.DaysToClear <= 7 {
			data.ForecastLabel = "On track — backlog clearing within a week"
		} else if data.DaysToClear <= 30 {
			data.ForecastLabel = "Steady progress — backlog clearing within a month"
		} else {
			data.ForecastLabel = "Long tail — consider prioritizing closure"
		}
	}

	s.render(w, r, "forecast", data)
}
