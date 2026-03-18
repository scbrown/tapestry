package web

import (
	"log"
	"net/http"
	"sort"
	"sync"

	"github.com/scbrown/tapestry/internal/dolt"
)

type ownerRow struct {
	Name     string
	Total    int
	Open     int
	Progress int
	Blocked  int
	Closed   int
	Deferred int
	P0       int
	P1       int
	P2       int
}

type ownersData struct {
	Rows      []ownerRow
	Rigs      []string
	FilterRig string
	SortBy    string
	Err       string
}

func (s *Server) handleOwners(w http.ResponseWriter, r *http.Request) {
	if s.ds == nil {
		s.render(w, r, "owners", ownersData{})
		return
	}

	ctx := r.Context()
	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("owners: list dbs: %v", err)
		s.render(w, r, "owners", ownersData{Err: err.Error()})
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
				log.Printf("owners %s: %v", dbName, err)
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

	byOwner := map[string]*ownerRow{}
	for _, r := range results {
		if filterRig != "" && r.rig != filterRig {
			continue
		}
		for _, issue := range r.issues {
			owner := issue.Owner
			if owner == "" {
				owner = "(unassigned)"
			}
			row, ok := byOwner[owner]
			if !ok {
				row = &ownerRow{Name: owner}
				byOwner[owner] = row
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
			// Count high-priority open work only
			if issue.Status != "closed" && issue.Status != "deferred" {
				switch issue.Priority {
				case 0:
					row.P0++
				case 1:
					row.P1++
				case 2:
					row.P2++
				}
			}
		}
	}

	var rows []ownerRow
	for _, row := range byOwner {
		rows = append(rows, *row)
	}
	sortBy := r.URL.Query().Get("sort")
	if sortBy == "" {
		sortBy = "active"
	}

	switch sortBy {
	case "total":
		sort.Slice(rows, func(i, j int) bool {
			return rows[i].Total > rows[j].Total
		})
	case "open":
		sort.Slice(rows, func(i, j int) bool {
			return rows[i].Open > rows[j].Open
		})
	case "blocked":
		sort.Slice(rows, func(i, j int) bool {
			return rows[i].Blocked > rows[j].Blocked
		})
	case "name":
		sort.Slice(rows, func(i, j int) bool {
			return rows[i].Name < rows[j].Name
		})
	default: // "active"
		sort.Slice(rows, func(i, j int) bool {
			ai := rows[i].Open + rows[i].Progress
			aj := rows[j].Open + rows[j].Progress
			if ai != aj {
				return ai > aj
			}
			return rows[i].Total > rows[j].Total
		})
	}

	s.render(w, r, "owners", ownersData{Rows: rows, Rigs: rigs, FilterRig: filterRig, SortBy: sortBy})
}
