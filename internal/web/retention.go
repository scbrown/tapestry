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

type retentionBucket struct {
	Status    string
	MedianH   float64
	MeanH     float64
	P90H      float64
	Count     int
}

type retentionData struct {
	GeneratedAt time.Time
	Buckets     []retentionBucket
	Rigs        []string
	FilterRig   string
}

func (s *Server) handleRetention(w http.ResponseWriter, r *http.Request) {
	data := retentionData{GeneratedAt: time.Now()}

	if s.ds == nil {
		s.render(w, r, "retention", data)
		return
	}

	ctx := r.Context()
	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("retention: list dbs: %v", err)
		s.render(w, r, "retention", data)
		return
	}

	filterRig := r.URL.Query().Get("rig")
	var rigs []string
	for _, db := range dbs {
		rigs = append(rigs, db.Name)
	}
	sort.Strings(rigs)
	data.Rigs = rigs
	data.FilterRig = filterRig

	// Gather closed issues and compute time-in-status from history
	type statusDuration struct {
		status  string
		hours   float64
	}
	var mu sync.Mutex
	var allDurations []statusDuration

	// Get closed issues from each DB
	type dbIssues struct {
		rig    string
		issues []dolt.Issue
	}
	dbResults := make([]dbIssues, len(dbs))
	var wg sync.WaitGroup
	for i, db := range dbs {
		if filterRig != "" && db.Name != filterRig {
			continue
		}
		wg.Add(1)
		go func(idx int, dbName string) {
			defer wg.Done()
			issues, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{Status: "closed", Limit: 500})
			if err != nil {
				log.Printf("retention: %s closed: %v", dbName, err)
				return
			}
			dbResults[idx] = dbIssues{rig: dbName, issues: issues}
		}(i, db.Name)
	}
	wg.Wait()

	// For each closed issue, walk status history
	sem := make(chan struct{}, 10)
	var histWg sync.WaitGroup
	for _, dr := range dbResults {
		for _, iss := range dr.issues {
			histWg.Add(1)
			sem <- struct{}{}
			go func(rig string, issue dolt.Issue) {
				defer histWg.Done()
				defer func() { <-sem }()

				hist, err := s.ds.StatusHistory(ctx, rig, issue.ID)
				if err != nil || len(hist) < 2 {
					return
				}

				for i := 0; i < len(hist)-1; i++ {
					status := hist[i].ToStatus
					dur := hist[i+1].CommitDate.Sub(hist[i].CommitDate)
					if dur > 0 {
						mu.Lock()
						allDurations = append(allDurations, statusDuration{
							status: status,
							hours:  dur.Hours(),
						})
						mu.Unlock()
					}
				}
			}(dr.rig, iss)
		}
	}
	histWg.Wait()

	// Group by status and compute statistics
	grouped := map[string][]float64{}
	for _, d := range allDurations {
		grouped[d.status] = append(grouped[d.status], d.hours)
	}

	for status, hours := range grouped {
		sort.Float64s(hours)
		n := len(hours)
		var sum float64
		for _, h := range hours {
			sum += h
		}
		median := hours[n/2]
		if n%2 == 0 && n > 1 {
			median = (hours[n/2-1] + hours[n/2]) / 2
		}
		p90idx := int(math.Ceil(float64(n)*0.9)) - 1
		if p90idx >= n {
			p90idx = n - 1
		}

		data.Buckets = append(data.Buckets, retentionBucket{
			Status:  status,
			MedianH: median,
			MeanH:   sum / float64(n),
			P90H:    hours[p90idx],
			Count:   n,
		})
	}

	sort.Slice(data.Buckets, func(i, j int) bool {
		return data.Buckets[i].MeanH > data.Buckets[j].MeanH
	})

	s.render(w, r, "retention", data)
}
