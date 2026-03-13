package web

import (
	"log"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/scbrown/tapestry/internal/dolt"
)

type parkingItem struct {
	Rig      string
	Issue    dolt.Issue
	IdleDays int // days since last update
	AgeDays  int // days since created
}

type parkingData struct {
	GeneratedAt time.Time
	Items       []parkingItem
	Total       int
	MedianIdle  int
	MaxIdle     int
	ByAssignee  map[string]int
}

func (s *Server) handleParkingLot(w http.ResponseWriter, r *http.Request) {
	data := parkingData{GeneratedAt: time.Now(), ByAssignee: map[string]int{}}

	if s.ds == nil {
		s.render(w, r, "parking-lot", data)
		return
	}

	ctx := r.Context()
	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("parking-lot: list dbs: %v", err)
		s.render(w, r, "parking-lot", data)
		return
	}

	type dbResult struct {
		rig    string
		issues []dolt.Issue
	}
	results := make([]dbResult, len(dbs))
	var wg sync.WaitGroup
	for i, db := range dbs {
		wg.Add(1)
		go func(i int, dbName string) {
			defer wg.Done()
			issues, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{Status: "in_progress", Limit: 5000})
			if err != nil {
				log.Printf("parking-lot: %s: %v", dbName, err)
				return
			}
			results[i] = dbResult{rig: dbName, issues: issues}
		}(i, db.Name)
	}
	wg.Wait()

	now := time.Now()
	threshold := 3 // days idle before considered "parked"

	for _, r := range results {
		for _, iss := range r.issues {
			idle := int(now.Sub(iss.UpdatedAt).Hours() / 24)
			if idle < threshold {
				continue
			}
			age := int(now.Sub(iss.CreatedAt).Hours() / 24)
			data.Items = append(data.Items, parkingItem{
				Rig: r.rig, Issue: iss, IdleDays: idle, AgeDays: age,
			})
			assignee := iss.Assignee
			if assignee == "" {
				assignee = iss.Owner
			}
			if assignee == "" {
				assignee = "(unassigned)"
			}
			data.ByAssignee[assignee]++
		}
	}

	// Sort by idle days descending
	sort.Slice(data.Items, func(i, j int) bool {
		return data.Items[i].IdleDays > data.Items[j].IdleDays
	})

	data.Total = len(data.Items)
	if data.Total > 0 {
		data.MaxIdle = data.Items[0].IdleDays
		idles := make([]int, len(data.Items))
		for i, item := range data.Items {
			idles[i] = item.IdleDays
		}
		data.MedianIdle = idles[len(idles)/2]
	}

	s.render(w, r, "parking-lot", data)
}
