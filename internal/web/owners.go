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
}

type ownersData struct {
	Rows []ownerRow
	Err  string
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
			results[i] = dbResult{issues: issues}
		}(i, db.Name)
	}
	wg.Wait()

	byOwner := map[string]*ownerRow{}
	for _, r := range results {
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
		}
	}

	var rows []ownerRow
	for _, row := range byOwner {
		rows = append(rows, *row)
	}
	sort.Slice(rows, func(i, j int) bool {
		// Sort by active work (open + in_progress) descending
		ai := rows[i].Open + rows[i].Progress
		aj := rows[j].Open + rows[j].Progress
		if ai != aj {
			return ai > aj
		}
		return rows[i].Total > rows[j].Total
	})

	s.render(w, r, "owners", ownersData{Rows: rows})
}
