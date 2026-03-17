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

type gapRig struct {
	Name       string
	Open       int
	LastUpdate time.Time
	StaleDays  int
}

type gapPriority struct {
	Priority   int
	Open       int
	Unassigned int
}

type gapType struct {
	Type       string
	Open       int
	Unassigned int
}

type gapsData struct {
	GeneratedAt time.Time

	// Rigs with no recent updates
	StaleRigs []gapRig
	// Priorities with unassigned work
	PriorityGaps []gapPriority
	// Types with unassigned work
	TypeGaps []gapType

	// Overall stats
	TotalUnassigned   int
	TotalOpen         int
	UnassignedHighPri int // P0+P1 with no assignee

	Err string
}

func (s *Server) handleGaps(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	data := gapsData{GeneratedAt: now}

	if s.ds == nil {
		s.render(w, r, "gaps", data)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("gaps: list dbs: %v", err)
		s.render(w, r, "gaps", gapsData{Err: err.Error(), GeneratedAt: now})
		return
	}

	type rigStats struct {
		name       string
		open       int
		lastUpdate time.Time
	}

	priMap := make(map[int][2]int) // priority -> [open, unassigned]
	typeMap := make(map[string][2]int)
	rigResults := make([]rigStats, len(dbs))

	var mu sync.Mutex
	var wg sync.WaitGroup

	for i, db := range dbs {
		wg.Add(1)
		go func(idx int, dbName string) {
			defer wg.Done()
			issues, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{Limit: 2000})
			if err != nil {
				log.Printf("gaps %s: %v", dbName, err)
				return
			}

			var localOpen int
			var lastUpdate time.Time
			localPri := make(map[int][2]int)
			localType := make(map[string][2]int)

			for _, iss := range issues {
				if isNoise(iss.ID, iss.Title) || iss.Status == "closed" || iss.Status == "deferred" {
					continue
				}
				localOpen++

				if iss.UpdatedAt.After(lastUpdate) {
					lastUpdate = iss.UpdatedAt
				}

				// Priority gaps
				p := localPri[iss.Priority]
				p[0]++
				if iss.Assignee == "" {
					p[1]++
				}
				localPri[iss.Priority] = p

				// Type gaps
				typ := iss.Type
				if typ == "" {
					typ = "(untyped)"
				}
				t := localType[typ]
				t[0]++
				if iss.Assignee == "" {
					t[1]++
				}
				localType[typ] = t
			}

			mu.Lock()
			defer mu.Unlock()
			rigResults[idx] = rigStats{name: dbName, open: localOpen, lastUpdate: lastUpdate}
			for k, v := range localPri {
				p := priMap[k]
				p[0] += v[0]
				p[1] += v[1]
				priMap[k] = p
			}
			for k, v := range localType {
				t := typeMap[k]
				t[0] += v[0]
				t[1] += v[1]
				typeMap[k] = t
			}
		}(i, db.Name)
	}
	wg.Wait()

	// Build stale rigs (no update in 7+ days)
	for _, rs := range rigResults {
		if rs.open == 0 {
			continue
		}
		staleDays := int(now.Sub(rs.lastUpdate).Hours() / 24)
		if staleDays >= 7 {
			data.StaleRigs = append(data.StaleRigs, gapRig{
				Name:       rs.name,
				Open:       rs.open,
				LastUpdate: rs.lastUpdate,
				StaleDays:  staleDays,
			})
		}
	}
	sort.Slice(data.StaleRigs, func(i, j int) bool {
		return data.StaleRigs[i].StaleDays > data.StaleRigs[j].StaleDays
	})

	// Build priority gaps
	for pri := 0; pri <= 4; pri++ {
		v := priMap[pri]
		if v[0] > 0 {
			data.PriorityGaps = append(data.PriorityGaps, gapPriority{
				Priority:   pri,
				Open:       v[0],
				Unassigned: v[1],
			})
			data.TotalOpen += v[0]
			data.TotalUnassigned += v[1]
			if pri <= 1 {
				data.UnassignedHighPri += v[1]
			}
		}
	}

	// Build type gaps
	for typ, v := range typeMap {
		if v[0] > 0 {
			data.TypeGaps = append(data.TypeGaps, gapType{
				Type:       typ,
				Open:       v[0],
				Unassigned: v[1],
			})
		}
	}
	sort.Slice(data.TypeGaps, func(i, j int) bool {
		return data.TypeGaps[i].Unassigned > data.TypeGaps[j].Unassigned
	})

	s.render(w, r, "gaps", data)
}
