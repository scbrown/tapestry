package web

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/scbrown/tapestry/internal/dolt"
)

type dwellItem struct {
	dolt.Issue
	Rig      string
	DwellStr string  // human-readable dwell time
	DwellPct float64 // 0-100 for bar width
}

type dwellData struct {
	GeneratedAt time.Time
	Items       []dwellItem
	Total       int
	Rigs        []string
	FilterRig   string
	Zone        string // "all", "danger" (14d+), "warning" (7-14d), "ok" (<7d)
	Err         string
}

func (s *Server) handleDwell(w http.ResponseWriter, r *http.Request) {
	data := dwellData{GeneratedAt: time.Now()}

	if s.ds == nil {
		s.render(w, r, "dwell", data)
		return
	}

	ctx := r.Context()
	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("dwell: list dbs: %v", err)
		s.render(w, r, "dwell", data)
		return
	}

	filterRig := r.URL.Query().Get("rig")
	zone := r.URL.Query().Get("zone")
	if zone == "" {
		zone = "all"
	}
	data.FilterRig = filterRig
	data.Zone = zone

	var rigs []string
	for _, db := range dbs {
		rigs = append(rigs, db.Name)
	}
	sort.Strings(rigs)
	data.Rigs = rigs

	now := time.Now()

	type dbResult struct {
		items []dwellItem
	}
	results := make([]dbResult, len(dbs))
	var wg sync.WaitGroup

	for i, db := range dbs {
		if filterRig != "" && db.Name != filterRig {
			continue
		}
		wg.Add(1)
		go func(idx int, dbName string) {
			defer wg.Done()
			issues, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{Limit: 5000})
			if err != nil {
				log.Printf("dwell: issues %s: %v", dbName, err)
				return
			}

			var items []dwellItem
			for _, iss := range issues {
				if iss.Status == "closed" || iss.Status == "deferred" || isNoise(iss.ID, iss.Title) {
					continue
				}

				dwell := now.Sub(iss.UpdatedAt)
				days := dwell.Hours() / 24

				// Zone filter
				switch zone {
				case "danger":
					if days < 14 {
						continue
					}
				case "warning":
					if days < 7 || days >= 14 {
						continue
					}
				case "ok":
					if days >= 7 {
						continue
					}
				}

				items = append(items, dwellItem{
					Issue:    iss,
					Rig:      dbName,
					DwellStr: formatDwell(dwell),
				})
			}
			results[idx] = dbResult{items: items}
		}(i, db.Name)
	}
	wg.Wait()

	var all []dwellItem
	for _, r := range results {
		all = append(all, r.items...)
	}

	// Sort by dwell time descending (longest first)
	sort.Slice(all, func(i, j int) bool {
		return all[i].UpdatedAt.Before(all[j].UpdatedAt)
	})

	// Compute bar widths relative to max dwell
	if len(all) > 0 {
		maxDwell := now.Sub(all[0].UpdatedAt).Hours()
		if maxDwell < 1 {
			maxDwell = 1
		}
		for i := range all {
			d := now.Sub(all[i].UpdatedAt).Hours()
			all[i].DwellPct = math.Min(100, (d/maxDwell)*100)
		}
	}

	if len(all) > 100 {
		all = all[:100]
	}

	data.Items = all
	data.Total = len(all)
	s.render(w, r, "dwell", data)
}

func formatDwell(d time.Duration) string {
	hours := d.Hours()
	if hours < 24 {
		return fmtDurationHelper(d)
	}
	days := int(hours / 24)
	if days < 7 {
		return fmtPlural(days, "day")
	}
	weeks := days / 7
	remDays := days % 7
	if remDays == 0 {
		return fmtPlural(weeks, "week")
	}
	return fmtPlural(weeks, "week") + " " + fmtPlural(remDays, "day")
}

func fmtPlural(n int, unit string) string {
	if n == 1 {
		return "1 " + unit
	}
	return fmt.Sprintf("%d %ss", n, unit)
}

func fmtDurationHelper(d time.Duration) string {
	h := int(d.Hours())
	if h == 0 {
		m := int(d.Minutes())
		if m == 0 {
			return "<1 min"
		}
		return fmtPlural(m, "min")
	}
	return fmtPlural(h, "hour")
}
