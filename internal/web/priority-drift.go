package web

import (
	"log"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/scbrown/tapestry/internal/dolt"
)

type priorityBucket struct {
	Priority int
	Open     int
	Active   int
	Blocked  int
	Stale    int // open/active but not updated in 7+ days
	Total    int
}

type priorityDriftData struct {
	GeneratedAt time.Time
	Buckets     []priorityBucket
	Rigs        []string
	FilterRig   string
	Err         string
}

func (s *Server) handlePriorityDrift(w http.ResponseWriter, r *http.Request) {
	data := priorityDriftData{GeneratedAt: time.Now()}

	if s.ds == nil {
		s.render(w, r, "priority-drift", data)
		return
	}

	ctx := r.Context()
	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("priority-drift: list dbs: %v", err)
		s.render(w, r, "priority-drift", data)
		return
	}

	filterRig := r.URL.Query().Get("rig")
	data.FilterRig = filterRig

	var rigs []string
	for _, db := range dbs {
		rigs = append(rigs, db.Name)
	}
	sort.Strings(rigs)
	data.Rigs = rigs

	now := time.Now()
	staleThreshold := now.AddDate(0, 0, -7)

	bucketMap := make(map[int]*priorityBucket)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, db := range dbs {
		if filterRig != "" && db.Name != filterRig {
			continue
		}
		wg.Add(1)
		go func(dbName string) {
			defer wg.Done()
			issues, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{Limit: 5000})
			if err != nil {
				log.Printf("priority-drift: issues %s: %v", dbName, err)
				return
			}

			mu.Lock()
			defer mu.Unlock()
			for _, iss := range issues {
				if iss.Status == "closed" || iss.Status == "deferred" || isNoise(iss.ID, iss.Title) {
					continue
				}

				b, ok := bucketMap[iss.Priority]
				if !ok {
					b = &priorityBucket{Priority: iss.Priority}
					bucketMap[iss.Priority] = b
				}

				b.Total++
				switch iss.Status {
				case "open":
					b.Open++
				case "in_progress", "hooked":
					b.Active++
				case "blocked":
					b.Blocked++
				}

				if iss.UpdatedAt.Before(staleThreshold) {
					b.Stale++
				}
			}
		}(db.Name)
	}
	wg.Wait()

	var buckets []priorityBucket
	for _, b := range bucketMap {
		buckets = append(buckets, *b)
	}
	sort.Slice(buckets, func(i, j int) bool {
		return buckets[i].Priority < buckets[j].Priority
	})

	data.Buckets = buckets
	s.render(w, r, "priority-drift", data)
}
