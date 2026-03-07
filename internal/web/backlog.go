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

type ageBucket struct {
	Label string
	Count int
	Pct   float64
}

type backlogItem struct {
	Issue  dolt.Issue
	AgeDays int
}

type backlogData struct {
	GeneratedAt time.Time
	Buckets     []ageBucket
	MaxBucket   int
	TotalOpen   int
	MedianAge   int
	AvgAge      float64
	P90Age      int
	Oldest      []backlogItem
	ByPriority  []priorityAge
}

type priorityAge struct {
	Priority int
	Count    int
	AvgAge   float64
}

func (s *Server) handleBacklog(w http.ResponseWriter, r *http.Request) {
	data := backlogData{GeneratedAt: time.Now()}

	if s.ds == nil {
		s.render(w, r, "backlog", data)
		return
	}

	ctx := r.Context()
	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("backlog: list dbs: %v", err)
		s.render(w, r, "backlog", data)
		return
	}

	type dbResult struct {
		issues []dolt.Issue
	}
	results := make([]dbResult, len(dbs))
	var wg sync.WaitGroup
	for i, db := range dbs {
		wg.Add(1)
		go func(i int, dbName string) {
			defer wg.Done()
			issues, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{Limit: 2000})
			if err != nil {
				log.Printf("backlog: %s: %v", dbName, err)
				return
			}
			results[i] = dbResult{issues: issues}
		}(i, db.Name)
	}
	wg.Wait()

	now := time.Now()
	var openIssues []backlogItem
	for _, r := range results {
		for _, iss := range r.issues {
			if iss.Status == "closed" || iss.Status == "deferred" {
				continue
			}
			age := int(now.Sub(iss.CreatedAt).Hours() / 24)
			if age < 0 {
				age = 0
			}
			openIssues = append(openIssues, backlogItem{Issue: iss, AgeDays: age})
		}
	}

	data.TotalOpen = len(openIssues)
	if data.TotalOpen == 0 {
		s.render(w, r, "backlog", data)
		return
	}

	// Sort by age descending for oldest items
	sort.Slice(openIssues, func(i, j int) bool {
		return openIssues[i].AgeDays > openIssues[j].AgeDays
	})

	// Age buckets
	bucketDefs := []struct {
		label string
		max   int // exclusive upper bound in days
	}{
		{"< 1d", 1},
		{"1-3d", 4},
		{"3-7d", 8},
		{"1-2w", 15},
		{"2-4w", 29},
		{"1-2m", 61},
		{"2m+", math.MaxInt32},
	}

	bucketCounts := make([]int, len(bucketDefs))
	for _, item := range openIssues {
		for bi, bd := range bucketDefs {
			if item.AgeDays < bd.max {
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
		pct := 0.0
		if data.TotalOpen > 0 {
			pct = float64(bucketCounts[i]) * 100.0 / float64(data.TotalOpen)
		}
		data.Buckets = append(data.Buckets, ageBucket{
			Label: bd.label,
			Count: bucketCounts[i],
			Pct:   pct,
		})
	}

	// Statistics
	ages := make([]int, len(openIssues))
	totalAge := 0
	for i, item := range openIssues {
		ages[i] = item.AgeDays
		totalAge += item.AgeDays
	}
	sort.Ints(ages)
	data.MedianAge = ages[len(ages)/2]
	data.AvgAge = float64(totalAge) / float64(len(ages))
	data.P90Age = ages[int(float64(len(ages))*0.9)]

	// Top 10 oldest
	limit := 10
	if len(openIssues) < limit {
		limit = len(openIssues)
	}
	data.Oldest = openIssues[:limit]

	// Age by priority
	priMap := map[int]*priorityAge{}
	for _, item := range openIssues {
		pa, ok := priMap[item.Issue.Priority]
		if !ok {
			pa = &priorityAge{Priority: item.Issue.Priority}
			priMap[item.Issue.Priority] = pa
		}
		pa.Count++
		pa.AvgAge += float64(item.AgeDays)
	}
	for _, pa := range priMap {
		if pa.Count > 0 {
			pa.AvgAge /= float64(pa.Count)
		}
		data.ByPriority = append(data.ByPriority, *pa)
	}
	sort.Slice(data.ByPriority, func(i, j int) bool {
		return data.ByPriority[i].Priority < data.ByPriority[j].Priority
	})

	s.render(w, r, "backlog", data)
}
