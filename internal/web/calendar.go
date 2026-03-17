package web

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/scbrown/tapestry/internal/dolt"
)

type calDay struct {
	Date    time.Time
	Created int
	Closed  int
	InMonth bool // whether this day is in the displayed month
}

type calWeek struct {
	Days [7]calDay
}

type calendarData struct {
	GeneratedAt time.Time
	Year        int
	Month       time.Month
	MonthName   string
	Weeks       []calWeek
	TotalCreate int
	TotalClose  int
	PrevYear    int
	PrevMonth   int
	NextYear    int
	NextMonth   int
	HasNext     bool
	MaxDay      int // max daily activity for scaling
	Rigs        []string
	FilterRig   string
	Err         string
}

func (s *Server) handleCalendar(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	data := calendarData{GeneratedAt: now}

	// Parse year/month from query, default to current
	year := now.Year()
	month := now.Month()
	if y := r.URL.Query().Get("year"); y != "" {
		if v, err := strconv.Atoi(y); err == nil && v >= 2020 && v <= 2030 {
			year = v
		}
	}
	if m := r.URL.Query().Get("month"); m != "" {
		if v, err := strconv.Atoi(m); err == nil && v >= 1 && v <= 12 {
			month = time.Month(v)
		}
	}

	data.Year = year
	data.Month = month
	data.MonthName = month.String()

	// Prev/next month
	prev := time.Date(year, month-1, 1, 0, 0, 0, 0, time.UTC)
	data.PrevYear = prev.Year()
	data.PrevMonth = int(prev.Month())
	next := time.Date(year, month+1, 1, 0, 0, 0, 0, time.UTC)
	data.NextYear = next.Year()
	data.NextMonth = int(next.Month())
	data.HasNext = next.Before(now) || (next.Year() == now.Year() && next.Month() <= now.Month())

	filterRig := r.URL.Query().Get("rig")
	data.FilterRig = filterRig

	if s.ds == nil {
		s.render(w, r, "calendar", data)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("calendar: list dbs: %v", err)
		s.render(w, r, "calendar", calendarData{Err: err.Error(), GeneratedAt: now})
		return
	}

	// Get all issues to count created/closed per day in this month
	monthStart := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	monthEnd := time.Date(year, month+1, 1, 0, 0, 0, 0, time.UTC)

	created := make(map[string]int) // date string -> count
	closed := make(map[string]int)
	rigSet := make(map[string]bool)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, db := range dbs {
		if filterRig != "" && db.Name != filterRig {
			continue
		}
		wg.Add(1)
		go func(dbName string) {
			defer wg.Done()

			// Get all issues and filter by date
			issues, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{Limit: 2000})
			if err != nil {
				log.Printf("calendar %s: %v", dbName, err)
				return
			}

			localCreated := make(map[string]int)
			localClosed := make(map[string]int)
			hasData := false

			for _, iss := range issues {
				if isNoise(iss.ID, iss.Title) {
					continue
				}
				if !iss.CreatedAt.Before(monthStart) && iss.CreatedAt.Before(monthEnd) {
					key := iss.CreatedAt.Format("2006-01-02")
					localCreated[key]++
					hasData = true
				}
			}

			// Also check closed issues
			closedIssues, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{
				Status: "closed",
				Limit:  2000,
			})
			if err == nil {
				for _, iss := range closedIssues {
					if isNoise(iss.ID, iss.Title) {
						continue
					}
					if !iss.UpdatedAt.Before(monthStart) && iss.UpdatedAt.Before(monthEnd) {
						key := iss.UpdatedAt.Format("2006-01-02")
						localClosed[key]++
						hasData = true
					}
				}
			}

			mu.Lock()
			defer mu.Unlock()
			if hasData {
				rigSet[dbName] = true
			}
			for k, v := range localCreated {
				created[k] += v
			}
			for k, v := range localClosed {
				closed[k] += v
			}
		}(db.Name)
	}
	wg.Wait()

	// Build calendar grid
	firstDay := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	lastDay := time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC) // last day of month

	// Start grid on the Sunday before or on the 1st
	gridStart := firstDay
	for gridStart.Weekday() != time.Sunday {
		gridStart = gridStart.AddDate(0, 0, -1)
	}

	// End grid on the Saturday after or on the last day
	gridEnd := lastDay
	for gridEnd.Weekday() != time.Saturday {
		gridEnd = gridEnd.AddDate(0, 0, 1)
	}

	maxDay := 0
	var weeks []calWeek
	cursor := gridStart
	for cursor.Before(gridEnd.AddDate(0, 0, 1)) {
		var week calWeek
		for d := 0; d < 7; d++ {
			key := cursor.Format("2006-01-02")
			c := created[key]
			cl := closed[key]
			inMonth := cursor.Month() == month
			week.Days[d] = calDay{
				Date:    cursor,
				Created: c,
				Closed:  cl,
				InMonth: inMonth,
			}
			if inMonth {
				data.TotalCreate += c
				data.TotalClose += cl
				if c+cl > maxDay {
					maxDay = c + cl
				}
			}
			cursor = cursor.AddDate(0, 0, 1)
		}
		weeks = append(weeks, week)
	}

	data.Weeks = weeks
	data.MaxDay = maxDay

	for rig := range rigSet {
		data.Rigs = append(data.Rigs, rig)
	}

	s.render(w, r, "calendar", data)
}
