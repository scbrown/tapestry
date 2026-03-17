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

// crossRefItem tracks beads that appear in multiple databases
type crossRefItem struct {
	ID       string
	Title    string
	Status   string
	Priority int
	DBs      []string
	DBCount  int
}

type crossRefData struct {
	GeneratedAt time.Time

	Items       []crossRefItem
	TotalCross  int // beads appearing in 2+ databases
	TotalDBs    int

	Err string
}

func (s *Server) handleCrossRef(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	data := crossRefData{GeneratedAt: now}

	if s.ds == nil {
		s.render(w, r, "crossref", data)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("crossref: list dbs: %v", err)
		data.Err = err.Error()
		s.render(w, r, "crossref", data)
		return
	}

	data.TotalDBs = len(dbs)

	// Track which databases each bead ID appears in
	beadDBs := make(map[string][]string)
	beadInfo := make(map[string]dolt.Issue)

	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, db := range dbs {
		wg.Add(1)
		go func(dbName string) {
			defer wg.Done()

			issues, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{Limit: 2000})
			if err != nil {
				log.Printf("crossref %s: %v", dbName, err)
				return
			}

			mu.Lock()
			for _, iss := range issues {
				if isNoise(iss.ID, iss.Title) {
					continue
				}
				beadDBs[iss.ID] = append(beadDBs[iss.ID], dbName)
				beadInfo[iss.ID] = iss
			}
			mu.Unlock()
		}(db.Name)
	}
	wg.Wait()

	for id, dbs := range beadDBs {
		if len(dbs) >= 2 {
			iss := beadInfo[id]
			data.Items = append(data.Items, crossRefItem{
				ID:       id,
				Title:    iss.Title,
				Status:   iss.Status,
				Priority: iss.Priority,
				DBs:      dbs,
				DBCount:  len(dbs),
			})
		}
	}

	data.TotalCross = len(data.Items)

	sort.Slice(data.Items, func(i, j int) bool {
		if data.Items[i].DBCount != data.Items[j].DBCount {
			return data.Items[i].DBCount > data.Items[j].DBCount
		}
		return data.Items[i].ID < data.Items[j].ID
	})

	if len(data.Items) > 50 {
		data.Items = data.Items[:50]
	}

	s.render(w, r, "crossref", data)
}
