package web

import (
	"fmt"
	"log"
	"net/http"
	"sort"
	"sync"

	"github.com/scbrown/tapestry/internal/dolt"
)

type priorityRow struct {
	Priority int
	Label    string
	Total    int
	Open     int
	Progress int
	Blocked  int
	Closed   int
	Deferred int
	Other    int
}

type prioritiesData struct {
	Rows      []priorityRow
	GrandTot  int
	Rigs      []string
	FilterRig string
	Err       string
}

func (s *Server) handlePriorities(w http.ResponseWriter, r *http.Request) {
	if s.ds == nil {
		s.render(w, r, "priorities", prioritiesData{})
		return
	}

	ctx := r.Context()
	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("priorities: list dbs: %v", err)
		s.render(w, r, "priorities", prioritiesData{Err: err.Error()})
		return
	}

	type dbResult struct {
		rig    string
		counts []dolt.PriorityStatusCount
	}

	results := make([]dbResult, len(dbs))
	var wg sync.WaitGroup
	for i, db := range dbs {
		wg.Add(1)
		go func(i int, dbName string) {
			defer wg.Done()
			counts, err := s.ds.CountByPriorityStatus(ctx, dbName)
			if err != nil {
				log.Printf("priorities %s: %v", dbName, err)
				return
			}
			results[i] = dbResult{rig: dbName, counts: counts}
		}(i, db.Name)
	}
	wg.Wait()

	filterRig := r.URL.Query().Get("rig")
	rigSet := make(map[string]bool)
	for _, r := range results {
		if len(r.counts) > 0 {
			rigSet[r.rig] = true
		}
	}
	var rigs []string
	for rig := range rigSet {
		rigs = append(rigs, rig)
	}
	sort.Strings(rigs)

	// Aggregate across databases (filtered by rig if set)
	byPri := map[int]*priorityRow{}
	for _, r := range results {
		if filterRig != "" && r.rig != filterRig {
			continue
		}
		for _, c := range r.counts {
			row, ok := byPri[c.Priority]
			if !ok {
				row = &priorityRow{
					Priority: c.Priority,
					Label:    fmt.Sprintf("P%d", c.Priority),
				}
				byPri[c.Priority] = row
			}
			row.Total += c.Count
			switch c.Status {
			case "open":
				row.Open += c.Count
			case "in_progress", "hooked":
				row.Progress += c.Count
			case "blocked":
				row.Blocked += c.Count
			case "closed":
				row.Closed += c.Count
			case "deferred":
				row.Deferred += c.Count
			default:
				row.Other += c.Count
			}
		}
	}

	// Build sorted rows P0-P4 (plus any extras)
	var rows []priorityRow
	grandTotal := 0
	for p := 0; p <= 4; p++ {
		if row, ok := byPri[p]; ok {
			rows = append(rows, *row)
			grandTotal += row.Total
			delete(byPri, p)
		}
	}
	for _, row := range byPri {
		rows = append(rows, *row)
		grandTotal += row.Total
	}

	s.render(w, r, "priorities", prioritiesData{
		Rows:      rows,
		GrandTot:  grandTotal,
		Rigs:      rigs,
		FilterRig: filterRig,
	})
}
