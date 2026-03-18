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

type labelMatrixRow struct {
	Label   string
	Open    int
	InProg  int
	Blocked int
	Closed  int
	Deferred int
	Total   int
}

type labelMatrixData struct {
	GeneratedAt time.Time

	Rows      []labelMatrixRow
	Statuses  []string
	MaxTotal  int

	Rigs      []string
	FilterRig string
	SortBy    string
	Err       string
}

func (s *Server) handleLabelMatrix(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	filterRig := r.URL.Query().Get("rig")
	data := labelMatrixData{
		GeneratedAt: now,
		FilterRig:   filterRig,
		Statuses:    []string{"open", "in_progress", "blocked", "closed", "deferred"},
	}

	if s.ds == nil {
		s.render(w, r, "label-matrix", data)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("label-matrix: list dbs: %v", err)
		data.Err = err.Error()
		s.render(w, r, "label-matrix", data)
		return
	}

	for _, db := range dbs {
		data.Rigs = append(data.Rigs, db.Name)
	}
	sort.Strings(data.Rigs)

	// Collect label → status → count across all DBs
	type key struct{ label, status string }
	counts := make(map[key]int)
	labelTotals := make(map[string]int)

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
				log.Printf("label-matrix %s: %v", dbName, err)
				return
			}

			for _, iss := range issues {
				if isNoise(iss.ID, iss.Title) {
					continue
				}

				labels, err := s.ds.LabelsForIssue(ctx, dbName, iss.ID)
				if err != nil {
					continue
				}

				mu.Lock()
				for _, label := range labels {
					counts[key{label, iss.Status}]++
					labelTotals[label]++
				}
				mu.Unlock()
			}
		}(db.Name)
	}
	wg.Wait()

	for label, total := range labelTotals {
		row := labelMatrixRow{
			Label:    label,
			Open:     counts[key{label, "open"}],
			InProg:   counts[key{label, "in_progress"}],
			Blocked:  counts[key{label, "blocked"}],
			Closed:   counts[key{label, "closed"}],
			Deferred: counts[key{label, "deferred"}],
			Total:    total,
		}
		data.Rows = append(data.Rows, row)
		if total > data.MaxTotal {
			data.MaxTotal = total
		}
	}

	sortBy := r.URL.Query().Get("sort")
	if sortBy == "" {
		sortBy = "total"
	}
	data.SortBy = sortBy

	switch sortBy {
	case "name":
		sort.Slice(data.Rows, func(i, j int) bool {
			return data.Rows[i].Label < data.Rows[j].Label
		})
	case "open":
		sort.Slice(data.Rows, func(i, j int) bool {
			if data.Rows[i].Open != data.Rows[j].Open {
				return data.Rows[i].Open > data.Rows[j].Open
			}
			return data.Rows[i].Total > data.Rows[j].Total
		})
	case "closed":
		sort.Slice(data.Rows, func(i, j int) bool {
			if data.Rows[i].Closed != data.Rows[j].Closed {
				return data.Rows[i].Closed > data.Rows[j].Closed
			}
			return data.Rows[i].Total > data.Rows[j].Total
		})
	default: // total
		sort.Slice(data.Rows, func(i, j int) bool {
			return data.Rows[i].Total > data.Rows[j].Total
		})
	}

	if len(data.Rows) > 50 {
		data.Rows = data.Rows[:50]
	}

	s.render(w, r, "label-matrix", data)
}
