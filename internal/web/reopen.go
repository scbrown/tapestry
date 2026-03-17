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

type reopenedBead struct {
	ID         string
	Title      string
	Status     string
	Priority   int
	Assignee   string
	DB         string
	Reopens    int // number of times reopened
	LastReopen time.Time
}

type reopenData struct {
	GeneratedAt time.Time
	FilterRig   string

	Beads        []reopenedBead
	TotalReopens int
	UniqueBeads  int

	Err string
}

func (s *Server) handleReopen(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	filterRig := r.URL.Query().Get("rig")
	data := reopenData{GeneratedAt: now, FilterRig: filterRig}

	if s.ds == nil {
		s.render(w, r, "reopen", data)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("reopen: list dbs: %v", err)
		data.Err = err.Error()
		s.render(w, r, "reopen", data)
		return
	}

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
				log.Printf("reopen %s: %v", dbName, err)
				return
			}

			for _, iss := range issues {
				if isNoise(iss.ID, iss.Title) {
					continue
				}
				if iss.Status == "closed" || iss.Status == "deferred" {
					continue
				}

				// Check status history for reopen patterns (closed → non-closed)
				history, err := s.ds.StatusHistory(ctx, dbName, iss.ID)
				if err != nil || len(history) < 2 {
					continue
				}

				reopens := 0
				var lastReopen time.Time
				for _, tr := range history {
					if tr.FromStatus == "closed" && tr.ToStatus != "closed" {
						reopens++
						if tr.CommitDate.After(lastReopen) {
							lastReopen = tr.CommitDate
						}
					}
				}

				if reopens > 0 {
					mu.Lock()
					data.Beads = append(data.Beads, reopenedBead{
						ID:         iss.ID,
						Title:      iss.Title,
						Status:     iss.Status,
						Priority:   iss.Priority,
						Assignee:   iss.Assignee,
						DB:         dbName,
						Reopens:    reopens,
						LastReopen: lastReopen,
					})
					data.TotalReopens += reopens
					mu.Unlock()
				}
			}
		}(db.Name)
	}
	wg.Wait()

	data.UniqueBeads = len(data.Beads)

	sort.Slice(data.Beads, func(i, j int) bool {
		if data.Beads[i].Reopens != data.Beads[j].Reopens {
			return data.Beads[i].Reopens > data.Beads[j].Reopens
		}
		return data.Beads[i].LastReopen.After(data.Beads[j].LastReopen)
	})

	if len(data.Beads) > 50 {
		data.Beads = data.Beads[:50]
	}

	s.render(w, r, "reopen", data)
}
