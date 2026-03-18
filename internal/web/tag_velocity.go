package web

import (
	"context"
	"log"
	"net/http"
	"sort"
	"sync"
	"time"
)

type tagVelocityRow struct {
	Label     string
	Open      int
	Closed30  int // closed in last 30 days
	Created30 int // created in last 30 days
	Net       int // Created30 - Closed30 (negative = shrinking)
}

type tagVelocityData struct {
	GeneratedAt time.Time
	Labels      []tagVelocityRow
	TotalOpen   int
	FilterRig   string
	Rigs        []string
	SortBy      string
}

func (s *Server) handleTagVelocity(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	data := tagVelocityData{GeneratedAt: now}

	if s.ds == nil {
		s.render(w, r, "tag-velocity", data)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("tag-velocity: list dbs: %v", err)
		s.render(w, r, "tag-velocity", data)
		return
	}

	filterRig := r.URL.Query().Get("rig")
	data.FilterRig = filterRig

	for _, db := range dbs {
		data.Rigs = append(data.Rigs, db.Name)
	}
	sort.Strings(data.Rigs)

	cutoff := now.AddDate(0, 0, -30)

	// Per-label aggregation across all DBs
	type labelStats struct {
		open      int
		closed30  int
		created30 int
	}
	agg := make(map[string]*labelStats)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, db := range dbs {
		if filterRig != "" && db.Name != filterRig {
			continue
		}
		wg.Add(1)
		go func(dbName string) {
			defer wg.Done()

			labels, err := s.ds.DistinctLabels(ctx, dbName)
			if err != nil {
				log.Printf("tag-velocity %s: labels: %v", dbName, err)
				return
			}

			for _, lc := range labels {
				issues, err := s.ds.IssuesByLabel(ctx, dbName, lc.Label)
				if err != nil {
					continue
				}

				var open, closed30, created30 int
				for _, iss := range issues {
					if isNoise(iss.ID, iss.Title) {
						continue
					}
					if iss.Status != "closed" && iss.Status != "deferred" {
						open++
					}
					if iss.Status == "closed" && !iss.UpdatedAt.Before(cutoff) {
						closed30++
					}
					if !iss.CreatedAt.Before(cutoff) {
						created30++
					}
				}

				mu.Lock()
				s := agg[lc.Label]
				if s == nil {
					s = &labelStats{}
					agg[lc.Label] = s
				}
				s.open += open
				s.closed30 += closed30
				s.created30 += created30
				mu.Unlock()
			}
		}(db.Name)
	}
	wg.Wait()

	for label, stats := range agg {
		row := tagVelocityRow{
			Label:     label,
			Open:      stats.open,
			Closed30:  stats.closed30,
			Created30: stats.created30,
			Net:       stats.created30 - stats.closed30,
		}
		data.Labels = append(data.Labels, row)
		data.TotalOpen += stats.open
	}

	sortBy := r.URL.Query().Get("sort")
	if sortBy == "" {
		sortBy = "net"
	}
	data.SortBy = sortBy

	switch sortBy {
	case "name":
		sort.Slice(data.Labels, func(i, j int) bool {
			return data.Labels[i].Label < data.Labels[j].Label
		})
	case "open":
		sort.Slice(data.Labels, func(i, j int) bool {
			if data.Labels[i].Open != data.Labels[j].Open {
				return data.Labels[i].Open > data.Labels[j].Open
			}
			return data.Labels[i].Net < data.Labels[j].Net
		})
	case "closed":
		sort.Slice(data.Labels, func(i, j int) bool {
			if data.Labels[i].Closed30 != data.Labels[j].Closed30 {
				return data.Labels[i].Closed30 > data.Labels[j].Closed30
			}
			return data.Labels[i].Net < data.Labels[j].Net
		})
	default: // net
		sort.Slice(data.Labels, func(i, j int) bool {
			return data.Labels[i].Net < data.Labels[j].Net
		})
	}

	s.render(w, r, "tag-velocity", data)
}
