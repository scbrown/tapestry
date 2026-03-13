package web

import (
	"log"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/scbrown/tapestry/internal/dolt"
)

type contributor struct {
	Name      string
	Owned     int
	Closed    int
	Open      int
	InProgress int
	CloseRate int // percentage
	LastActive time.Time
}

type contributorsData struct {
	GeneratedAt  time.Time
	Contributors []contributor
	Total        int
}

func (s *Server) handleContributors(w http.ResponseWriter, r *http.Request) {
	data := contributorsData{GeneratedAt: time.Now()}

	if s.ds == nil {
		s.render(w, r, "contributors", data)
		return
	}

	ctx := r.Context()
	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("contributors: list dbs: %v", err)
		s.render(w, r, "contributors", data)
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
			issues, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{Limit: 5000})
			if err != nil {
				log.Printf("contributors: %s: %v", dbName, err)
				return
			}
			results[i] = dbResult{issues: issues}
		}(i, db.Name)
	}
	wg.Wait()

	// Aggregate by owner/assignee
	contribMap := map[string]*contributor{}

	getOrCreate := func(name string) *contributor {
		if name == "" {
			return nil
		}
		c, ok := contribMap[name]
		if !ok {
			c = &contributor{Name: name}
			contribMap[name] = c
		}
		return c
	}

	for _, r := range results {
		for _, iss := range r.issues {
			// Count by owner
			if c := getOrCreate(iss.Owner); c != nil {
				c.Owned++
				switch iss.Status {
				case "closed":
					c.Closed++
				case "in_progress", "hooked":
					c.InProgress++
				default:
					c.Open++
				}
				if iss.UpdatedAt.After(c.LastActive) {
					c.LastActive = iss.UpdatedAt
				}
			}
			// If assignee differs from owner, also count assignee
			if iss.Assignee != "" && iss.Assignee != iss.Owner {
				if c := getOrCreate(iss.Assignee); c != nil {
					c.Owned++
					switch iss.Status {
					case "closed":
						c.Closed++
					case "in_progress", "hooked":
						c.InProgress++
					default:
						c.Open++
					}
					if iss.UpdatedAt.After(c.LastActive) {
						c.LastActive = iss.UpdatedAt
					}
				}
			}
		}
	}

	for _, c := range contribMap {
		if c.Owned > 0 {
			c.CloseRate = c.Closed * 100 / c.Owned
		}
		data.Contributors = append(data.Contributors, *c)
	}

	// Sort by total owned descending
	sort.Slice(data.Contributors, func(i, j int) bool {
		return data.Contributors[i].Owned > data.Contributors[j].Owned
	})

	data.Total = len(data.Contributors)
	s.render(w, r, "contributors", data)
}
