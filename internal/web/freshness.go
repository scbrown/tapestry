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
	FilterRig   string
	SortBy      string
	Rigs        []string

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
	filterRig := r.URL.Query().Get("rig")
	sortBy := r.URL.Query().Get("sort")
	if sortBy == "" {
		sortBy = "updated"
	}
	data := freshnessData{GeneratedAt: now, FilterRig: filterRig, SortBy: sortBy}

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

	for _, db := range dbs {
		data.Rigs = append(data.Rigs, db.Name)
	}
	sort.Strings(data.Rigs)

	staleThreshold := now.Add(-7 * 24 * time.Hour)
	weekAgo := now.Add(-7 * 24 * time.Hour)

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

	switch sortBy {
	case "name":
		sort.Slice(data.Databases, func(i, j int) bool {
			return data.Databases[i].Name < data.Databases[j].Name
		})
	case "total":
		sort.Slice(data.Databases, func(i, j int) bool {
			return data.Databases[i].TotalBeads > data.Databases[j].TotalBeads
		})
	case "stale":
		sort.Slice(data.Databases, func(i, j int) bool {
			return data.Databases[i].StaleDays > data.Databases[j].StaleDays
		})
	default: // "updated"
		sort.Slice(data.Databases, func(i, j int) bool {
			return data.Databases[i].LastUpdate.After(data.Databases[j].LastUpdate)
		})
	}

	s.render(w, r, "freshness", data)
}
