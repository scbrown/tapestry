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

type dbFreshness struct {
	Name         string
	TotalBeads   int
	OpenBeads    int
	LastUpdate   time.Time
	StaleDays    int
	RecentCreated int // created in last 7 days
	RecentClosed  int // closed in last 7 days
}

type freshnessData struct {
	GeneratedAt time.Time

	Databases []dbFreshness

	// Aggregates
	TotalDBs     int
	StaleDBs     int // no update in 7+ days
	ActiveDBs    int
	TotalBeads   int

	Err string
}

func (s *Server) handleFreshness(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	data := freshnessData{GeneratedAt: now}

	if s.ds == nil {
		s.render(w, r, "freshness", data)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("freshness: list dbs: %v", err)
		data.Err = err.Error()
		s.render(w, r, "freshness", data)
		return
	}

	staleThreshold := now.Add(-7 * 24 * time.Hour)
	weekAgo := now.Add(-7 * 24 * time.Hour)

	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, db := range dbs {
		wg.Add(1)
		go func(dbName string) {
			defer wg.Done()

			issues, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{Limit: 2000})
			if err != nil {
				log.Printf("freshness %s: %v", dbName, err)
				return
			}

			f := dbFreshness{Name: dbName, TotalBeads: len(issues)}
			for _, iss := range issues {
				if isNoise(iss.ID, iss.Title) {
					continue
				}
				if iss.Status != "closed" && iss.Status != "deferred" {
					f.OpenBeads++
				}
				if iss.UpdatedAt.After(f.LastUpdate) {
					f.LastUpdate = iss.UpdatedAt
				}
				if iss.CreatedAt.After(weekAgo) {
					f.RecentCreated++
				}
				if iss.Status == "closed" && iss.UpdatedAt.After(weekAgo) {
					f.RecentClosed++
				}
			}

			if !f.LastUpdate.IsZero() {
				f.StaleDays = int(now.Sub(f.LastUpdate).Hours() / 24)
			}

			mu.Lock()
			data.Databases = append(data.Databases, f)
			data.TotalBeads += f.TotalBeads
			if f.LastUpdate.Before(staleThreshold) {
				data.StaleDBs++
			} else {
				data.ActiveDBs++
			}
			mu.Unlock()
		}(db.Name)
	}
	wg.Wait()

	data.TotalDBs = len(data.Databases)

	sort.Slice(data.Databases, func(i, j int) bool {
		return data.Databases[i].LastUpdate.After(data.Databases[j].LastUpdate)
	})

	s.render(w, r, "freshness", data)
}
