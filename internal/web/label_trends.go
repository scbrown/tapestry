package web

import (
	"context"
	"log"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/scbrown/tapestry/internal/dolt"
)

type labelWeek struct {
	WeekStart string // "2026-03-10"
	Count     int
}

type labelTrend struct {
	Label     string
	Weeks     []labelWeek
	Total     int
	Delta     int    // change from prev week to latest week
	Direction string // "up", "down", "flat"
}

type labelTrendsData struct {
	GeneratedAt time.Time

	Trends   []labelTrend
	WeekHeaders []string // column headers for the table

	FilterRig string
	Rigs      []string
	SortBy    string
	Err       string
}

func (s *Server) handleLabelTrends(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	filterRig := r.URL.Query().Get("rig")
	data := labelTrendsData{GeneratedAt: now, FilterRig: filterRig}

	if s.ds == nil {
		s.render(w, r, "label-trends", data)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("label-trends: list dbs: %v", err)
		data.Err = err.Error()
		s.render(w, r, "label-trends", data)
		return
	}

	for _, db := range dbs {
		data.Rigs = append(data.Rigs, db.Name)
	}
	sort.Strings(data.Rigs)

	// Build 8-week buckets
	numWeeks := 8
	weekStarts := make([]time.Time, numWeeks)
	for i := 0; i < numWeeks; i++ {
		weekStarts[numWeeks-1-i] = now.AddDate(0, 0, -7*i).Truncate(24*time.Hour)
		// Align to Monday
		for weekStarts[numWeeks-1-i].Weekday() != time.Monday {
			weekStarts[numWeeks-1-i] = weekStarts[numWeeks-1-i].AddDate(0, 0, -1)
		}
	}
	for _, ws := range weekStarts {
		data.WeekHeaders = append(data.WeekHeaders, ws.Format("Jan 2"))
	}

	// Collect label → week → count across all DBs
	type key struct{ label string; week int }
	counts := make(map[key]int)
	labelTotal := make(map[string]int)

	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, db := range dbs {
		if filterRig != "" && db.Name != filterRig {
			continue
		}
		wg.Add(1)
		go func(dbName string) {
			defer wg.Done()

			issues, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{Limit: 2000})
			if err != nil {
				log.Printf("label-trends %s: %v", dbName, err)
				return
			}

			for _, iss := range issues {
				if isNoise(iss.ID, iss.Title) {
					continue
				}

				labels, err := s.ds.LabelsForIssue(ctx, dbName, iss.ID)
				if err != nil || len(labels) == 0 {
					continue
				}

				// Bucket by creation week
				weekIdx := -1
				for i := numWeeks - 1; i >= 0; i-- {
					end := weekStarts[i].AddDate(0, 0, 7)
					if !iss.CreatedAt.Before(weekStarts[i]) && iss.CreatedAt.Before(end) {
						weekIdx = i
						break
					}
				}
				if weekIdx < 0 {
					continue
				}

				mu.Lock()
				for _, label := range labels {
					counts[key{label, weekIdx}]++
					labelTotal[label]++
				}
				mu.Unlock()
			}
		}(db.Name)
	}
	wg.Wait()

	for label, total := range labelTotal {
		trend := labelTrend{
			Label: label,
			Total: total,
			Weeks: make([]labelWeek, numWeeks),
		}
		for i := 0; i < numWeeks; i++ {
			trend.Weeks[i] = labelWeek{
				WeekStart: weekStarts[i].Format("2006-01-02"),
				Count:     counts[key{label, i}],
			}
		}
		// Calculate delta between last two weeks
		if numWeeks >= 2 {
			latest := trend.Weeks[numWeeks-1].Count
			prev := trend.Weeks[numWeeks-2].Count
			trend.Delta = latest - prev
			if trend.Delta > 0 {
				trend.Direction = "up"
			} else if trend.Delta < 0 {
				trend.Direction = "down"
			} else {
				trend.Direction = "flat"
			}
		}
		data.Trends = append(data.Trends, trend)
	}

	sortBy := r.URL.Query().Get("sort")
	if sortBy == "" {
		sortBy = "total"
	}
	data.SortBy = sortBy

	switch sortBy {
	case "name":
		sort.Slice(data.Trends, func(i, j int) bool {
			return data.Trends[i].Label < data.Trends[j].Label
		})
	case "delta":
		sort.Slice(data.Trends, func(i, j int) bool {
			if data.Trends[i].Delta != data.Trends[j].Delta {
				return data.Trends[i].Delta > data.Trends[j].Delta
			}
			return data.Trends[i].Total > data.Trends[j].Total
		})
	case "direction":
		sort.Slice(data.Trends, func(i, j int) bool {
			rank := map[string]int{"up": 0, "flat": 1, "down": 2}
			if rank[data.Trends[i].Direction] != rank[data.Trends[j].Direction] {
				return rank[data.Trends[i].Direction] < rank[data.Trends[j].Direction]
			}
			return data.Trends[i].Total > data.Trends[j].Total
		})
	default: // total
		sort.Slice(data.Trends, func(i, j int) bool {
			return data.Trends[i].Total > data.Trends[j].Total
		})
	}

	if len(data.Trends) > 30 {
		data.Trends = data.Trends[:30]
	}

	s.render(w, r, "label-trends", data)
}
