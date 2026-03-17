package web

import (
	"context"
	"log"
	"math"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/scbrown/tapestry/internal/dolt"
)

type pacingData struct {
	GeneratedAt    time.Time
	TotalOpen      int
	DailyCloseRate float64 // 30-day average closes per day
	DaysToClear    int     // at current rate, -1 if rate is 0
	Indicator      string  // "ahead", "behind", "on-pace"
	Closed30       int     // total closed in last 30 days
	Created30      int     // total created in last 30 days
	NetRate        float64 // daily net change (created - closed) / 30
	FilterRig      string
	Rigs           []string
}

func (s *Server) handlePacing(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	data := pacingData{GeneratedAt: now, DaysToClear: -1}

	if s.ds == nil {
		s.render(w, r, "pacing", data)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("pacing: list dbs: %v", err)
		s.render(w, r, "pacing", data)
		return
	}

	filterRig := r.URL.Query().Get("rig")
	data.FilterRig = filterRig

	for _, db := range dbs {
		data.Rigs = append(data.Rigs, db.Name)
	}
	sort.Strings(data.Rigs)

	type dbResult struct {
		open      int
		closed30  int
		created30 int
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
			issues, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{Limit: 10000})
			if err != nil {
				log.Printf("pacing %s: %v", dbName, err)
				return
			}
			cutoff := now.AddDate(0, 0, -30)
			var r dbResult
			for _, iss := range issues {
				if isNoise(iss.ID, iss.Title) {
					continue
				}
				if iss.Status != "closed" && iss.Status != "deferred" {
					r.open++
				}
				if iss.Status == "closed" && !iss.UpdatedAt.Before(cutoff) {
					r.closed30++
				}
				if !iss.CreatedAt.Before(cutoff) {
					r.created30++
				}
			}
			results[idx] = r
		}(i, db.Name)
	}
	wg.Wait()

	for _, r := range results {
		data.TotalOpen += r.open
		data.Closed30 += r.closed30
		data.Created30 += r.created30
	}

	data.DailyCloseRate = float64(data.Closed30) / 30.0
	data.NetRate = float64(data.Created30-data.Closed30) / 30.0

	if data.DailyCloseRate > 0 {
		days := float64(data.TotalOpen) / data.DailyCloseRate
		data.DaysToClear = int(math.Ceil(days))
	}

	// Determine pacing indicator
	// "ahead" = closing faster than creating (net rate < 0)
	// "behind" = creating faster than closing (net rate > 0)
	// "on-pace" = roughly balanced (within 0.1/day)
	if data.NetRate < -0.1 {
		data.Indicator = "ahead"
	} else if data.NetRate > 0.1 {
		data.Indicator = "behind"
	} else {
		data.Indicator = "on-pace"
	}

	s.render(w, r, "pacing", data)
}
