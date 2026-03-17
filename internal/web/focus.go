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

// focusBead is a bead scored for urgency
type focusBead struct {
	ID         string
	Title      string
	Status     string
	Priority   int
	Assignee   string
	DB         string
	AgeDays    int
	Score      float64
	Reason     string // why this is high-focus
}

type focusData struct {
	GeneratedAt time.Time

	// Top focus items (sorted by composite urgency score)
	Items []focusBead

	// Summary
	TotalScored int
	AvgScore    float64
	MaxScore    float64

	// Rig filter
	FilterRig string
	Rigs      []string

	Err string
}

func (s *Server) handleFocus(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	filterRig := r.URL.Query().Get("rig")
	data := focusData{GeneratedAt: now, FilterRig: filterRig}

	if s.ds == nil {
		s.render(w, r, "focus", data)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("focus: list dbs: %v", err)
		data.Err = err.Error()
		s.render(w, r, "focus", data)
		return
	}

	for _, db := range dbs {
		data.Rigs = append(data.Rigs, db.Name)
	}

	var allItems []focusBead
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
				log.Printf("focus %s: %v", dbName, err)
				return
			}

			for _, iss := range issues {
				if isNoise(iss.ID, iss.Title) {
					continue
				}
				if iss.Status == "closed" || iss.Status == "deferred" {
					continue
				}
				if iss.Status == "blocked" {
					continue // blocked items can't be worked on
				}

				ageDays := int(now.Sub(iss.CreatedAt).Hours() / 24)

				// Compute composite score:
				// - Priority weight: P0=50, P1=30, P2=15, P3=5, P4=1
				// - Age factor: sqrt(age_days) * 2
				// - Status bonus: in_progress gets +10 (already started, keep momentum)
				// - Unassigned penalty: -5 (needs owner first)
				priorityWeight := map[int]float64{0: 50, 1: 30, 2: 15, 3: 5, 4: 1}
				pw := priorityWeight[iss.Priority]
				if pw == 0 {
					pw = 5 // default for unknown priority
				}

				ageFactor := math.Sqrt(float64(ageDays)) * 2
				score := pw + ageFactor

				var reason string
				if iss.Status == "in_progress" || iss.Status == "hooked" {
					score += 10
					reason = "in progress"
				}
				if iss.Assignee == "" {
					score -= 5
					reason = "unassigned"
				}
				if iss.Priority <= 1 {
					reason = "high priority"
				}
				if ageDays > 30 {
					reason = "aging"
				}
				if iss.Priority <= 1 && ageDays > 14 {
					reason = "stale high-pri"
					score += 20
				}

				mu.Lock()
				allItems = append(allItems, focusBead{
					ID:       iss.ID,
					Title:    iss.Title,
					Status:   iss.Status,
					Priority: iss.Priority,
					Assignee: iss.Assignee,
					DB:       dbName,
					AgeDays:  ageDays,
					Score:    score,
					Reason:   reason,
				})
				mu.Unlock()
			}
		}(db.Name)
	}
	wg.Wait()

	sort.Slice(allItems, func(i, j int) bool {
		return allItems[i].Score > allItems[j].Score
	})

	data.TotalScored = len(allItems)
	if len(allItems) > 0 {
		total := 0.0
		for _, item := range allItems {
			total += item.Score
			if item.Score > data.MaxScore {
				data.MaxScore = item.Score
			}
		}
		data.AvgScore = total / float64(len(allItems))
	}

	// Show top 30
	if len(allItems) > 30 {
		allItems = allItems[:30]
	}
	data.Items = allItems

	s.render(w, r, "focus", data)
}
