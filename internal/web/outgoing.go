package web

import (
	"log"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/scbrown/tapestry/internal/dolt"
)

type outgoingItem struct {
	Issue dolt.Issue
	Rig   string
	Age   int // days since creation
}

type outgoingData struct {
	GeneratedAt time.Time
	Oldest      []outgoingItem
	Newest      []outgoingItem
	TotalOpen   int
	MedianAge   int
	FilterRig   string
	Err         string
}

func (s *Server) handleOutgoing(w http.ResponseWriter, r *http.Request) {
	data := outgoingData{GeneratedAt: time.Now()}

	if s.ds == nil {
		s.render(w, r, "outgoing", data)
		return
	}

	ctx := r.Context()
	data.FilterRig = r.URL.Query().Get("rig")

	dbs, err := s.databases(ctx)
	if err != nil {
		data.Err = err.Error()
		s.render(w, r, "outgoing", data)
		return
	}

	now := time.Now()
	var allOpen []outgoingItem
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, db := range dbs {
		if data.FilterRig != "" && db.Name != data.FilterRig {
			continue
		}
		wg.Add(1)
		go func(dbName string) {
			defer wg.Done()
			issues, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{Limit: 5000})
			if err != nil {
				log.Printf("outgoing: %s: %v", dbName, err)
				return
			}

			var items []outgoingItem
			for _, iss := range issues {
				if iss.Status == "closed" || iss.Status == "deferred" {
					continue
				}
				age := int(now.Sub(iss.CreatedAt).Hours() / 24)
				items = append(items, outgoingItem{Issue: iss, Rig: dbName, Age: age})
			}

			mu.Lock()
			allOpen = append(allOpen, items...)
			mu.Unlock()
		}(db.Name)
	}
	wg.Wait()

	data.TotalOpen = len(allOpen)

	// Sort by age descending for oldest
	sort.Slice(allOpen, func(i, j int) bool {
		return allOpen[i].Age > allOpen[j].Age
	})

	// Median age
	if len(allOpen) > 0 {
		data.MedianAge = allOpen[len(allOpen)/2].Age
	}

	// Top 25 oldest
	if len(allOpen) > 25 {
		data.Oldest = allOpen[:25]
	} else {
		data.Oldest = allOpen
	}

	// Top 25 newest (reverse)
	sort.Slice(allOpen, func(i, j int) bool {
		return allOpen[i].Age < allOpen[j].Age
	})
	if len(allOpen) > 25 {
		data.Newest = allOpen[:25]
	} else {
		data.Newest = allOpen
	}

	s.render(w, r, "outgoing", data)
}
