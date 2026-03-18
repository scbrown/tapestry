package web

import (
	"log"
	"net/http"
	"sort"
	"sync"

	"github.com/scbrown/tapestry/internal/dolt"
)

type typeRow struct {
	Type     string
	Total    int
	Open     int
	Progress int
	Blocked  int
	Closed   int
	Deferred int
}

type typesData struct {
	Rows      []typeRow
	GrandTot  int
	Rigs      []string
	FilterRig string
	SortBy    string
	Err       string
}

func (s *Server) handleTypes(w http.ResponseWriter, r *http.Request) {
	if s.ds == nil {
		s.render(w, r, "types", typesData{})
		return
	}

	ctx := r.Context()
	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("types: list dbs: %v", err)
		s.render(w, r, "types", typesData{Err: err.Error()})
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
			issues, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{})
			if err != nil {
				log.Printf("types %s: %v", dbName, err)
				return
			}
			results[i] = dbResult{rig: dbName, issues: issues}
		}(i, db.Name)
	}
	wg.Wait()

	filterRig := r.URL.Query().Get("rig")
	rigSet := make(map[string]bool)
	for _, r := range results {
		if len(r.issues) > 0 {
			rigSet[r.rig] = true
		}
	}
	var rigs []string
	for rig := range rigSet {
		rigs = append(rigs, rig)
	}
	sort.Strings(rigs)

	byType := map[string]*typeRow{}
	for _, r := range results {
		if filterRig != "" && r.rig != filterRig {
			continue
		}
		for _, issue := range r.issues {
			t := issue.Type
			if t == "" {
				t = "(none)"
			}
			row, ok := byType[t]
			if !ok {
				row = &typeRow{Type: t}
				byType[t] = row
			}
			row.Total++
			switch issue.Status {
			case "open":
				row.Open++
			case "in_progress", "hooked":
				row.Progress++
			case "blocked":
				row.Blocked++
			case "closed":
				row.Closed++
			case "deferred":
				row.Deferred++
			}
		}
	}

	var rows []typeRow
	grandTotal := 0
	for _, row := range byType {
		rows = append(rows, *row)
		grandTotal += row.Total
	}

	sortBy := r.URL.Query().Get("sort")
	if sortBy == "" {
		sortBy = "type"
	}

	switch sortBy {
	case "total":
		sort.Slice(rows, func(i, j int) bool {
			return rows[i].Total > rows[j].Total
		})
	case "open":
		sort.Slice(rows, func(i, j int) bool {
			if rows[i].Open != rows[j].Open {
				return rows[i].Open > rows[j].Open
			}
			return rows[i].Total > rows[j].Total
		})
	case "name":
		sort.Slice(rows, func(i, j int) bool {
			return rows[i].Type < rows[j].Type
		})
	default: // "type" — epic, task, bug, then others alphabetically
		order := map[string]int{"epic": 0, "task": 1, "bug": 2}
		sort.Slice(rows, func(i, j int) bool {
			oi, ok1 := order[rows[i].Type]
			oj, ok2 := order[rows[j].Type]
			if ok1 && ok2 {
				return oi < oj
			}
			if ok1 {
				return true
			}
			if ok2 {
				return false
			}
			return rows[i].Type < rows[j].Type
		})
	}

	s.render(w, r, "types", typesData{
		Rows:      rows,
		GrandTot:  grandTotal,
		Rigs:      rigs,
		FilterRig: filterRig,
		SortBy:    sortBy,
	})
}
