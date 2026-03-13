package web

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/scbrown/tapestry/internal/dolt"
)

type flowDay struct {
	Date    string
	Created int
	Closed  int
	Net     int // created - closed (positive = growing backlog)
}

type flowRateData struct {
	GeneratedAt    time.Time
	Days           []flowDay
	MaxDaily       int
	TotalCreated   int
	TotalClosed    int
	AvgDailyCreate float64
	AvgDailyClose  float64
	NetChange      int
}

func (s *Server) handleFlowRate(w http.ResponseWriter, r *http.Request) {
	data := flowRateData{GeneratedAt: time.Now()}

	if s.ds == nil {
		s.render(w, r, "flow-rate", data)
		return
	}

	ctx := r.Context()
	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("flow-rate: list dbs: %v", err)
		s.render(w, r, "flow-rate", data)
		return
	}

	// Fetch all issues from last 30 days (created or updated)
	now := time.Now()
	cutoff := now.AddDate(0, 0, -30)

	type dbResult struct {
		issues []dolt.Issue
	}
	results := make([]dbResult, len(dbs))
	var wg sync.WaitGroup
	for i, db := range dbs {
		wg.Add(1)
		go func(i int, dbName string) {
			defer wg.Done()
			issues, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{Limit: 3000})
			if err != nil {
				log.Printf("flow-rate: %s: %v", dbName, err)
				return
			}
			results[i] = dbResult{issues: issues}
		}(i, db.Name)
	}
	wg.Wait()

	// Count created and closed per day
	createdByDay := map[string]int{}
	closedByDay := map[string]int{}

	for _, r := range results {
		for _, iss := range r.issues {
			if iss.CreatedAt.After(cutoff) {
				day := iss.CreatedAt.Format("2006-01-02")
				createdByDay[day]++
			}
			if iss.Status == "closed" && iss.UpdatedAt.After(cutoff) {
				day := iss.UpdatedAt.Format("2006-01-02")
				closedByDay[day]++
			}
		}
	}

	// Build day-by-day series for last 30 days
	maxDaily := 0
	for d := 29; d >= 0; d-- {
		day := now.AddDate(0, 0, -d)
		dayStr := day.Format("2006-01-02")
		label := day.Format("Jan 2")

		created := createdByDay[dayStr]
		closed := closedByDay[dayStr]

		data.Days = append(data.Days, flowDay{
			Date:    label,
			Created: created,
			Closed:  closed,
			Net:     created - closed,
		})

		data.TotalCreated += created
		data.TotalClosed += closed

		if created > maxDaily {
			maxDaily = created
		}
		if closed > maxDaily {
			maxDaily = closed
		}
	}

	data.MaxDaily = maxDaily
	data.NetChange = data.TotalCreated - data.TotalClosed
	if len(data.Days) > 0 {
		data.AvgDailyCreate = float64(data.TotalCreated) / float64(len(data.Days))
		data.AvgDailyClose = float64(data.TotalClosed) / float64(len(data.Days))
	}

	s.render(w, r, "flow-rate", data)
}
