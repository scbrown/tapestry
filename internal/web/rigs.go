package web

import (
	"log"
	"net/http"
	"sort"
	"sync"
	"time"
)

type rigRow struct {
	Name     string
	Total    int
	Open     int
	Progress int
	Blocked  int
	Closed   int
	Deferred int
}

type rigsData struct {
	Rows        []rigRow
	GrandTotal  int
	GeneratedAt time.Time
	Err         string
}

func (s *Server) handleRigs(w http.ResponseWriter, r *http.Request) {
	data := rigsData{GeneratedAt: time.Now()}

	if s.ds == nil {
		s.render(w, r, "rigs", data)
		return
	}

	ctx := r.Context()
	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("rigs: list dbs: %v", err)
		data.Err = err.Error()
		s.render(w, r, "rigs", data)
		return
	}

	type dbResult struct {
		name   string
		counts map[string]int
	}

	results := make([]dbResult, len(dbs))
	var wg sync.WaitGroup
	for i, db := range dbs {
		wg.Add(1)
		go func(i int, dbName string) {
			defer wg.Done()
			counts, err := s.ds.CountByStatus(ctx, dbName)
			if err != nil {
				log.Printf("rigs %s: %v", dbName, err)
				return
			}
			results[i] = dbResult{name: dbName, counts: counts}
		}(i, db.Name)
	}
	wg.Wait()

	var rows []rigRow
	grandTotal := 0
	for _, r := range results {
		if r.name == "" {
			continue
		}
		row := rigRow{Name: r.name}
		for status, count := range r.counts {
			switch status {
			case "open":
				row.Open += count
			case "in_progress", "hooked":
				row.Progress += count
			case "blocked":
				row.Blocked += count
			case "closed":
				row.Closed += count
			case "deferred":
				row.Deferred += count
			}
			row.Total += count
		}
		rows = append(rows, row)
		grandTotal += row.Total
	}

	// Sort by active work (open + in_progress) descending
	sort.Slice(rows, func(i, j int) bool {
		ai := rows[i].Open + rows[i].Progress + rows[i].Blocked
		aj := rows[j].Open + rows[j].Progress + rows[j].Blocked
		if ai != aj {
			return ai > aj
		}
		return rows[i].Total > rows[j].Total
	})

	data.Rows = rows
	data.GrandTotal = grandTotal
	s.render(w, r, "rigs", data)
}
