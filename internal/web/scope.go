package web

import (
	"context"
	"log"
	"net/http"
	"sort"
	"sync"
	"time"
)

type scopeDay struct {
	Date           time.Time
	Created        int
	Closed         int
	CumCreated     int
	CumClosed      int
	BacklogSize    int // cumCreated - cumClosed
}

type scopeData struct {
	GeneratedAt time.Time
	Days        []scopeDay
	MaxBacklog  int
	MinBacklog  int
	Period      int // number of days shown

	// Summary
	TotalCreated int
	TotalClosed  int
	NetChange    int
	StartBacklog int
	EndBacklog   int
	BacklogDelta int

	Rigs      []string
	FilterRig string
}

func (s *Server) handleScope(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	data := scopeData{GeneratedAt: now, Period: 30}

	if s.ds == nil {
		s.render(w, r, "scope", data)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("scope: list dbs: %v", err)
		s.render(w, r, "scope", data)
		return
	}

	// Build rig list from all DBs, then filter
	filterRig := r.URL.Query().Get("rig")
	data.FilterRig = filterRig
	rigSet := make(map[string]bool)
	for _, db := range dbs {
		rigSet[db.Name] = true
	}
	for rig := range rigSet {
		data.Rigs = append(data.Rigs, rig)
	}
	sort.Strings(data.Rigs)

	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	numDays := data.Period

	type dbResult struct {
		days [30]struct{ created, closed int }
	}

	results := make([]dbResult, len(dbs))
	var wg sync.WaitGroup
	for i, db := range dbs {
		if filterRig != "" && db.Name != filterRig {
			continue
		}
		wg.Add(1)
		go func(i int, dbName string) {
			defer wg.Done()
			var r dbResult
			for d := 0; d < numDays; d++ {
				dayStart := todayStart.AddDate(0, 0, -d)
				dayEnd := dayStart.AddDate(0, 0, 1)
				created, err := s.ds.CountCreatedInRange(ctx, dbName, dayStart, dayEnd)
				if err == nil {
					r.days[d].created = created
				}
				closed, err := s.ds.CountClosedInRange(ctx, dbName, dayStart, dayEnd)
				if err == nil {
					r.days[d].closed = closed
				}
			}
			results[i] = r
		}(i, db.Name)
	}
	wg.Wait()

	// Aggregate (oldest first)
	var cumCreated, cumClosed int
	for d := numDays - 1; d >= 0; d-- {
		day := scopeDay{Date: todayStart.AddDate(0, 0, -d)}
		for _, r := range results {
			day.Created += r.days[d].created
			day.Closed += r.days[d].closed
		}
		cumCreated += day.Created
		cumClosed += day.Closed
		day.CumCreated = cumCreated
		day.CumClosed = cumClosed
		day.BacklogSize = cumCreated - cumClosed

		data.Days = append(data.Days, day)
		data.TotalCreated += day.Created
		data.TotalClosed += day.Closed

		if day.BacklogSize > data.MaxBacklog || len(data.Days) == 1 {
			data.MaxBacklog = day.BacklogSize
		}
		if day.BacklogSize < data.MinBacklog || len(data.Days) == 1 {
			data.MinBacklog = day.BacklogSize
		}
	}

	data.NetChange = data.TotalCreated - data.TotalClosed
	if len(data.Days) > 0 {
		data.StartBacklog = data.Days[0].BacklogSize
		data.EndBacklog = data.Days[len(data.Days)-1].BacklogSize
		data.BacklogDelta = data.EndBacklog - data.StartBacklog
	}

	s.render(w, r, "scope", data)
}
