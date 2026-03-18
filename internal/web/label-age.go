package web

import (
	"log"
	"math"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/scbrown/tapestry/internal/dolt"
)

type labelAgeEntry struct {
	Label    string
	Count    int
	MedianDays float64
	MeanDays   float64
	MaxDays    float64
	BarPct     float64 // 0-100 for max age bar
}

type labelAgeData struct {
	GeneratedAt time.Time
	Labels      []labelAgeEntry
	Total       int
	Rigs        []string
	FilterRig   string
	SortBy      string
	Err         string
}

func (s *Server) handleLabelAge(w http.ResponseWriter, r *http.Request) {
	data := labelAgeData{GeneratedAt: time.Now()}

	if s.ds == nil {
		s.render(w, r, "label-age", data)
		return
	}

	ctx := r.Context()
	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("label-age: list dbs: %v", err)
		s.render(w, r, "label-age", data)
		return
	}

	filterRig := r.URL.Query().Get("rig")
	data.FilterRig = filterRig

	var rigs []string
	for _, db := range dbs {
		rigs = append(rigs, db.Name)
	}
	sort.Strings(rigs)
	data.Rigs = rigs

	now := time.Now()

	// Collect age data per label
	type labelAges struct {
		ages []float64
	}
	labelMap := make(map[string]*labelAges)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, db := range dbs {
		if filterRig != "" && db.Name != filterRig {
			continue
		}
		wg.Add(1)
		go func(dbName string) {
			defer wg.Done()
			issues, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{Limit: 5000})
			if err != nil {
				log.Printf("label-age: issues %s: %v", dbName, err)
				return
			}

			for _, iss := range issues {
				if iss.Status == "closed" || iss.Status == "deferred" || isNoise(iss.ID, iss.Title) {
					continue
				}

				labels, err := s.ds.LabelsForIssue(ctx, dbName, iss.ID)
				if err != nil || len(labels) == 0 {
					continue
				}

				ageDays := now.Sub(iss.CreatedAt).Hours() / 24

				mu.Lock()
				for _, lbl := range labels {
					la, ok := labelMap[lbl]
					if !ok {
						la = &labelAges{}
						labelMap[lbl] = la
					}
					la.ages = append(la.ages, ageDays)
				}
				mu.Unlock()
			}
		}(db.Name)
	}
	wg.Wait()

	var entries []labelAgeEntry
	var globalMax float64
	for label, la := range labelMap {
		if len(la.ages) == 0 {
			continue
		}
		sort.Float64s(la.ages)

		var sum float64
		for _, a := range la.ages {
			sum += a
		}
		mean := sum / float64(len(la.ages))
		median := la.ages[len(la.ages)/2]
		maxAge := la.ages[len(la.ages)-1]

		if maxAge > globalMax {
			globalMax = maxAge
		}

		entries = append(entries, labelAgeEntry{
			Label:      label,
			Count:      len(la.ages),
			MedianDays: math.Round(median*10) / 10,
			MeanDays:   math.Round(mean*10) / 10,
			MaxDays:    math.Round(maxAge*10) / 10,
		})
	}

	// Compute bar widths
	if globalMax < 1 {
		globalMax = 1
	}
	for i := range entries {
		entries[i].BarPct = math.Min(100, (entries[i].MaxDays/globalMax)*100)
	}

	sortBy := r.URL.Query().Get("sort")
	if sortBy == "" {
		sortBy = "median"
	}
	data.SortBy = sortBy

	switch sortBy {
	case "label":
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].Label < entries[j].Label
		})
	case "count":
		sort.Slice(entries, func(i, j int) bool {
			if entries[i].Count != entries[j].Count {
				return entries[i].Count > entries[j].Count
			}
			return entries[i].MedianDays > entries[j].MedianDays
		})
	case "max":
		sort.Slice(entries, func(i, j int) bool {
			if entries[i].MaxDays != entries[j].MaxDays {
				return entries[i].MaxDays > entries[j].MaxDays
			}
			return entries[i].MedianDays > entries[j].MedianDays
		})
	default: // median
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].MedianDays > entries[j].MedianDays
		})
	}

	data.Labels = entries
	data.Total = len(entries)
	s.render(w, r, "label-age", data)
}
