package web

import (
	"log"
	"net/http"
	"sort"
	"sync"

	"github.com/scbrown/tapestry/internal/dolt"
)

type matrixCell struct {
	Count int
	Pct   float64 // percentage of row total for heat intensity
}

type matrixRow struct {
	Assignee string
	Total    int
	Cells    map[string]matrixCell
}

type matrixData struct {
	Rows      []matrixRow
	Statuses  []string
	GrandTot  int
	MaxCell   int // largest single cell value (for heat scaling)
	Rigs      []string
	FilterRig string
	Err       string
}

func (s *Server) handleMatrix(w http.ResponseWriter, r *http.Request) {
	if s.ds == nil {
		s.render(w, r, "matrix", matrixData{})
		return
	}

	ctx := r.Context()
	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("matrix: list dbs: %v", err)
		s.render(w, r, "matrix", matrixData{Err: err.Error()})
		return
	}

	type dbResult struct {
		rig    string
		counts []dolt.AssigneeStatusCount
	}

	results := make([]dbResult, len(dbs))
	var wg sync.WaitGroup
	for i, db := range dbs {
		wg.Add(1)
		go func(i int, dbName string) {
			defer wg.Done()
			counts, err := s.ds.CountByAssigneeStatus(ctx, dbName)
			if err != nil {
				log.Printf("matrix %s: %v", dbName, err)
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
	byAssignee := map[string]*matrixRow{}
	statusSet := map[string]bool{}
	for _, r := range results {
		if filterRig != "" && r.rig != filterRig {
			continue
		}
		for _, c := range r.counts {
			row, ok := byAssignee[c.Assignee]
			if !ok {
				row = &matrixRow{
					Assignee: c.Assignee,
					Cells:    map[string]matrixCell{},
				}
				byAssignee[c.Assignee] = row
			}
			cell := row.Cells[c.Status]
			cell.Count += c.Count
			row.Cells[c.Status] = cell
			row.Total += c.Count
			statusSet[c.Status] = true
		}
	}

	// Ordered statuses
	statusOrder := []string{"open", "in_progress", "hooked", "blocked", "deferred", "closed"}
	var statuses []string
	for _, s := range statusOrder {
		if statusSet[s] {
			statuses = append(statuses, s)
			delete(statusSet, s)
		}
	}
	// Append any unexpected statuses
	for s := range statusSet {
		statuses = append(statuses, s)
	}

	// Build sorted rows by total (descending)
	var rows []matrixRow
	grandTotal := 0
	maxCell := 0
	for _, row := range byAssignee {
		rows = append(rows, *row)
		grandTotal += row.Total
		for _, cell := range row.Cells {
			if cell.Count > maxCell {
				maxCell = cell.Count
			}
		}
	}
	sort.Slice(rows, func(i, j int) bool {
		return rows[i].Total > rows[j].Total
	})

	s.render(w, r, "matrix", matrixData{
		Rows:      rows,
		Statuses:  statuses,
		GrandTot:  grandTotal,
		MaxCell:   maxCell,
		Rigs:      rigs,
		FilterRig: filterRig,
	})
}
