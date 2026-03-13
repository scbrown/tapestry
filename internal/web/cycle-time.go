package web

import (
	"log"
	"math"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/scbrown/tapestry/internal/dolt"
)

type cycleTimeBucket struct {
	Label string
	Count int
	Pct   float64
}

type cycleTimeItem struct {
	Issue     dolt.Issue
	CycleDays float64
}

type weeklyCycleTime struct {
	WeekLabel  string
	Count      int
	MedianDays float64
}

type priorityCycleTime struct {
	Priority   int
	Count      int
	MedianDays float64
	AvgDays    float64
	P90Days    float64
}

type typeCycleTime struct {
	Type       string
	Count      int
	MedianDays float64
	AvgDays    float64
}

type cycleTimeData struct {
	GeneratedAt time.Time
	TotalClosed int
	MedianDays  float64
	AvgDays     float64
	P90Days     float64
	P50Days     float64
	Buckets     []cycleTimeBucket
	MaxBucket   int
	ByPriority  []priorityCycleTime
	ByType      []typeCycleTime
	Weekly      []weeklyCycleTime
	Fastest     []cycleTimeItem
	Slowest     []cycleTimeItem
}

func (s *Server) handleCycleTime(w http.ResponseWriter, r *http.Request) {
	data := cycleTimeData{GeneratedAt: time.Now()}

	if s.ds == nil {
		s.render(w, r, "cycle-time", data)
		return
	}

	ctx := r.Context()
	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("cycle-time: list dbs: %v", err)
		s.render(w, r, "cycle-time", data)
		return
	}

	// Fetch closed issues from all databases
	type dbResult struct {
		issues []dolt.Issue
	}
	results := make([]dbResult, len(dbs))
	var wg sync.WaitGroup
	for i, db := range dbs {
		wg.Add(1)
		go func(i int, dbName string) {
			defer wg.Done()
			issues, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{
				Status: "closed",
				Limit:  3000,
			})
			if err != nil {
				log.Printf("cycle-time: %s: %v", dbName, err)
				return
			}
			results[i] = dbResult{issues: issues}
		}(i, db.Name)
	}
	wg.Wait()

	// Compute cycle times
	var items []cycleTimeItem
	for _, r := range results {
		for _, iss := range r.issues {
			if iss.UpdatedAt.IsZero() || iss.CreatedAt.IsZero() {
				continue
			}
			days := iss.UpdatedAt.Sub(iss.CreatedAt).Hours() / 24
			if days < 0 {
				days = 0
			}
			items = append(items, cycleTimeItem{Issue: iss, CycleDays: days})
		}
	}

	data.TotalClosed = len(items)
	if data.TotalClosed == 0 {
		s.render(w, r, "cycle-time", data)
		return
	}

	// Sort by cycle time for statistics
	sort.Slice(items, func(i, j int) bool {
		return items[i].CycleDays < items[j].CycleDays
	})

	// Overall statistics
	cycleTimes := make([]float64, len(items))
	totalDays := 0.0
	for i, item := range items {
		cycleTimes[i] = item.CycleDays
		totalDays += item.CycleDays
	}
	data.MedianDays = cycleTimes[len(cycleTimes)/2]
	data.AvgDays = totalDays / float64(len(cycleTimes))
	data.P50Days = cycleTimes[len(cycleTimes)/2]
	data.P90Days = cycleTimes[int(float64(len(cycleTimes))*0.9)]

	// Distribution buckets
	bucketDefs := []struct {
		label string
		max   float64
	}{
		{"< 1h", 1.0 / 24},
		{"1-6h", 0.25},
		{"6-24h", 1.0},
		{"1-3d", 3.0},
		{"3-7d", 7.0},
		{"1-2w", 14.0},
		{"2-4w", 28.0},
		{"1-2m", 60.0},
		{"2m+", math.MaxFloat64},
	}
	bucketCounts := make([]int, len(bucketDefs))
	for _, item := range items {
		for bi, bd := range bucketDefs {
			if item.CycleDays < bd.max {
				bucketCounts[bi]++
				break
			}
		}
	}
	maxBucket := 0
	for _, c := range bucketCounts {
		if c > maxBucket {
			maxBucket = c
		}
	}
	data.MaxBucket = maxBucket
	for i, bd := range bucketDefs {
		pct := float64(bucketCounts[i]) * 100.0 / float64(data.TotalClosed)
		data.Buckets = append(data.Buckets, cycleTimeBucket{
			Label: bd.label,
			Count: bucketCounts[i],
			Pct:   pct,
		})
	}

	// By priority
	priMap := map[int][]float64{}
	for _, item := range items {
		priMap[item.Issue.Priority] = append(priMap[item.Issue.Priority], item.CycleDays)
	}
	for p, times := range priMap {
		sort.Float64s(times)
		total := 0.0
		for _, t := range times {
			total += t
		}
		data.ByPriority = append(data.ByPriority, priorityCycleTime{
			Priority:   p,
			Count:      len(times),
			MedianDays: times[len(times)/2],
			AvgDays:    total / float64(len(times)),
			P90Days:    times[int(float64(len(times))*0.9)],
		})
	}
	sort.Slice(data.ByPriority, func(i, j int) bool {
		return data.ByPriority[i].Priority < data.ByPriority[j].Priority
	})

	// By type
	typeMap := map[string][]float64{}
	for _, item := range items {
		t := item.Issue.Type
		if t == "" {
			t = "untyped"
		}
		typeMap[t] = append(typeMap[t], item.CycleDays)
	}
	for t, times := range typeMap {
		sort.Float64s(times)
		total := 0.0
		for _, ti := range times {
			total += ti
		}
		data.ByType = append(data.ByType, typeCycleTime{
			Type:       t,
			Count:      len(times),
			MedianDays: times[len(times)/2],
			AvgDays:    total / float64(len(times)),
		})
	}
	sort.Slice(data.ByType, func(i, j int) bool {
		return data.ByType[i].MedianDays < data.ByType[j].MedianDays
	})

	// Weekly trend (last 8 weeks)
	now := time.Now()
	for w := 7; w >= 0; w-- {
		weekEnd := now.AddDate(0, 0, -7*w)
		weekStart := weekEnd.AddDate(0, 0, -7)
		label := weekStart.Format("Jan 2")

		var weekTimes []float64
		for _, item := range items {
			if item.Issue.UpdatedAt.After(weekStart) && item.Issue.UpdatedAt.Before(weekEnd) {
				weekTimes = append(weekTimes, item.CycleDays)
			}
		}
		median := 0.0
		if len(weekTimes) > 0 {
			sort.Float64s(weekTimes)
			median = weekTimes[len(weekTimes)/2]
		}
		data.Weekly = append(data.Weekly, weeklyCycleTime{
			WeekLabel:  label,
			Count:      len(weekTimes),
			MedianDays: median,
		})
	}

	// Fastest 5 (non-zero)
	for _, item := range items {
		if item.CycleDays > 0.001 && len(data.Fastest) < 5 {
			data.Fastest = append(data.Fastest, item)
		}
	}

	// Slowest 5
	for i := len(items) - 1; i >= 0 && len(data.Slowest) < 5; i-- {
		data.Slowest = append(data.Slowest, items[i])
	}

	s.render(w, r, "cycle-time", data)
}
