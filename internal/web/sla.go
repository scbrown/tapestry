package web

import (
	"log"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/scbrown/tapestry/internal/dolt"
)

// SLA targets by priority (how long a bead should stay open).
var slaTargets = map[int]time.Duration{
	0: 24 * time.Hour,        // P0: 1 day
	1: 3 * 24 * time.Hour,    // P1: 3 days
	2: 7 * 24 * time.Hour,    // P2: 7 days
	3: 14 * 24 * time.Hour,   // P3: 14 days
	4: 30 * 24 * time.Hour,   // P4: 30 days
}

type slaItem struct {
	Issue    dolt.Issue
	Target   time.Duration
	Age      time.Duration
	Overdue  bool
	OverdueH float64 // hours overdue (0 if not overdue)
	PctUsed  float64 // percentage of SLA consumed
}

type slaData struct {
	Breached  []slaItem
	AtRisk    []slaItem // >75% of SLA consumed
	OnTrack   int
	Total     int
	Rigs      []string
	FilterRig string
	Err       string
}

func (s *Server) handleSLA(w http.ResponseWriter, r *http.Request) {
	if s.ds == nil {
		s.render(w, r, "sla", slaData{})
		return
	}

	ctx := r.Context()
	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("sla: list dbs: %v", err)
		s.render(w, r, "sla", slaData{Err: err.Error()})
		return
	}

	type dbResult struct {
		rig    string
		issues []dolt.Issue
	}

	results := make([]dbResult, len(dbs))
	var wg sync.WaitGroup
	for i, db := range dbs {
		wg.Add(1)
		go func(i int, dbName string) {
			defer wg.Done()
			// Get all non-closed issues
			issues, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{})
			if err != nil {
				log.Printf("sla %s: %v", dbName, err)
				return
			}
			// Filter to open/in_progress/blocked/hooked only
			var active []dolt.Issue
			for _, iss := range issues {
				switch iss.Status {
				case "open", "in_progress", "blocked", "hooked":
					iss.Rig = dbName
					active = append(active, iss)
				}
			}
			results[i] = dbResult{rig: dbName, issues: active}
		}(i, db.Name)
	}
	wg.Wait()

	filterRig := r.URL.Query().Get("rig")
	rigSet := make(map[string]bool)
	for _, r := range results {
		if len(r.issues) > 0 {
			rigSet[r.rig] = true
		}
	}
	var rigs []string
	for rig := range rigSet {
		rigs = append(rigs, rig)
	}
	sort.Strings(rigs)

	now := time.Now()
	var breached, atRisk []slaItem
	onTrack := 0
	total := 0

	for _, r := range results {
		if filterRig != "" && r.rig != filterRig {
			continue
		}
		for _, iss := range r.issues {
			target, ok := slaTargets[iss.Priority]
			if !ok {
				target = 30 * 24 * time.Hour // default 30d
			}
			age := now.Sub(iss.CreatedAt)
			pctUsed := float64(age) / float64(target) * 100
			item := slaItem{
				Issue:   iss,
				Target:  target,
				Age:     age,
				PctUsed: pctUsed,
			}
			total++

			if age > target {
				item.Overdue = true
				item.OverdueH = (age - target).Hours()
				breached = append(breached, item)
			} else if pctUsed > 75 {
				atRisk = append(atRisk, item)
			} else {
				onTrack++
			}
		}
	}

	// Sort breached by overdue hours descending (worst first)
	sort.Slice(breached, func(i, j int) bool {
		return breached[i].OverdueH > breached[j].OverdueH
	})
	// Sort at-risk by pct used descending
	sort.Slice(atRisk, func(i, j int) bool {
		return atRisk[i].PctUsed > atRisk[j].PctUsed
	})

	s.render(w, r, "sla", slaData{
		Breached:  breached,
		AtRisk:    atRisk,
		OnTrack:   onTrack,
		Total:     total,
		Rigs:      rigs,
		FilterRig: filterRig,
	})
}
