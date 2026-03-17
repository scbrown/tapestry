package web

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/scbrown/tapestry/internal/dolt"
)

type signal struct {
	Name   string
	Status string // "green", "yellow", "red"
	Value  string
	Detail string
}

type signalsData struct {
	GeneratedAt time.Time

	Signals []signal

	// Raw metrics used to compute signals
	OpenCount       int
	ClosedWeek      int
	CreatedWeek     int
	BlockedCount    int
	P0Count         int
	P1Count         int
	StaleCount      int
	UnassignedCount int
	InProgressCount int
	DeferredCount   int

	Err string
}

func (s *Server) handleSignals(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	data := signalsData{GeneratedAt: now}

	if s.ds == nil {
		s.render(w, r, "signals", data)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("signals: list dbs: %v", err)
		data.Err = err.Error()
		s.render(w, r, "signals", data)
		return
	}

	weekAgo := now.Add(-7 * 24 * time.Hour)
	staleThreshold := now.Add(-14 * 24 * time.Hour)

	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, db := range dbs {
		wg.Add(1)
		go func(dbName string) {
			defer wg.Done()

			issues, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{Limit: 2000})
			if err != nil {
				log.Printf("signals %s: %v", dbName, err)
				return
			}

			var localOpen, localBlocked, localP0, localP1, localStale int
			var localUnassigned, localInProgress, localDeferred int
			var localCreatedWeek int

			for _, iss := range issues {
				if isNoise(iss.ID, iss.Title) {
					continue
				}
				switch iss.Status {
				case "closed":
					continue
				case "in_progress", "hooked":
					localInProgress++
				case "blocked":
					localBlocked++
				case "deferred":
					localDeferred++
					continue
				default:
					localOpen++
				}

				if iss.Assignee == "" && iss.Status != "deferred" {
					localUnassigned++
				}
				if iss.Priority == 0 {
					localP0++
				} else if iss.Priority == 1 {
					localP1++
				}
				if iss.UpdatedAt.Before(staleThreshold) && iss.Status != "deferred" {
					localStale++
				}
				if iss.CreatedAt.After(weekAgo) {
					localCreatedWeek++
				}
			}

			// Count closed in last week
			closedIssues, _ := s.ds.Issues(ctx, dbName, dolt.IssueFilter{Status: "closed", Limit: 500})
			var localClosedWeek int
			for _, iss := range closedIssues {
				if !isNoise(iss.ID, iss.Title) && iss.UpdatedAt.After(weekAgo) {
					localClosedWeek++
				}
			}

			mu.Lock()
			data.OpenCount += localOpen + localInProgress + localBlocked
			data.InProgressCount += localInProgress
			data.BlockedCount += localBlocked
			data.DeferredCount += localDeferred
			data.P0Count += localP0
			data.P1Count += localP1
			data.StaleCount += localStale
			data.UnassignedCount += localUnassigned
			data.CreatedWeek += localCreatedWeek
			data.ClosedWeek += localClosedWeek
			mu.Unlock()
		}(db.Name)
	}
	wg.Wait()

	// Compute signals
	data.Signals = computeSignals(data)

	s.render(w, r, "signals", data)
}

func computeSignals(d signalsData) []signal {
	var signals []signal

	// 1. Velocity: are we closing more than creating?
	netFlow := d.CreatedWeek - d.ClosedWeek
	if netFlow <= 0 {
		signals = append(signals, signal{
			Name: "Velocity", Status: "green",
			Value:  fmt.Sprintf("+%d net closed", -netFlow),
			Detail: fmt.Sprintf("Created %d, closed %d this week", d.CreatedWeek, d.ClosedWeek),
		})
	} else if netFlow <= 5 {
		signals = append(signals, signal{
			Name: "Velocity", Status: "yellow",
			Value:  fmt.Sprintf("+%d net growth", netFlow),
			Detail: fmt.Sprintf("Created %d, closed %d — slight backlog growth", d.CreatedWeek, d.ClosedWeek),
		})
	} else {
		signals = append(signals, signal{
			Name: "Velocity", Status: "red",
			Value:  fmt.Sprintf("+%d net growth", netFlow),
			Detail: fmt.Sprintf("Created %d, closed %d — backlog growing fast", d.CreatedWeek, d.ClosedWeek),
		})
	}

	// 2. Blockers: blocked items as % of open
	if d.OpenCount > 0 {
		blockedPct := float64(d.BlockedCount) / float64(d.OpenCount) * 100
		if blockedPct < 10 {
			signals = append(signals, signal{
				Name: "Blockers", Status: "green",
				Value:  fmt.Sprintf("%.0f%%", blockedPct),
				Detail: fmt.Sprintf("%d blocked of %d open", d.BlockedCount, d.OpenCount),
			})
		} else if blockedPct < 25 {
			signals = append(signals, signal{
				Name: "Blockers", Status: "yellow",
				Value:  fmt.Sprintf("%.0f%%", blockedPct),
				Detail: fmt.Sprintf("%d blocked of %d open — many items stuck", d.BlockedCount, d.OpenCount),
			})
		} else {
			signals = append(signals, signal{
				Name: "Blockers", Status: "red",
				Value:  fmt.Sprintf("%.0f%%", blockedPct),
				Detail: fmt.Sprintf("%d blocked of %d open — system clogged", d.BlockedCount, d.OpenCount),
			})
		}
	}

	// 3. Staleness
	if d.StaleCount == 0 {
		signals = append(signals, signal{
			Name: "Staleness", Status: "green",
			Value: "0 stale", Detail: "No beads idle for 14+ days",
		})
	} else if d.StaleCount <= 10 {
		signals = append(signals, signal{
			Name: "Staleness", Status: "yellow",
			Value:  fmt.Sprintf("%d stale", d.StaleCount),
			Detail: "Some beads haven't been updated in 14+ days",
		})
	} else {
		signals = append(signals, signal{
			Name: "Staleness", Status: "red",
			Value:  fmt.Sprintf("%d stale", d.StaleCount),
			Detail: "Many beads going stale — work may be abandoned",
		})
	}

	// 4. Urgency: P0/P1 open items
	highPri := d.P0Count + d.P1Count
	if highPri == 0 {
		signals = append(signals, signal{
			Name: "Urgency", Status: "green",
			Value: "0 high-pri", Detail: "No P0/P1 items open",
		})
	} else if highPri <= 3 {
		signals = append(signals, signal{
			Name: "Urgency", Status: "yellow",
			Value:  fmt.Sprintf("%d high-pri", highPri),
			Detail: fmt.Sprintf("P0: %d, P1: %d open", d.P0Count, d.P1Count),
		})
	} else {
		signals = append(signals, signal{
			Name: "Urgency", Status: "red",
			Value:  fmt.Sprintf("%d high-pri", highPri),
			Detail: fmt.Sprintf("P0: %d, P1: %d open — too many fires", d.P0Count, d.P1Count),
		})
	}

	// 5. Coverage: unassigned items
	if d.UnassignedCount == 0 {
		signals = append(signals, signal{
			Name: "Coverage", Status: "green",
			Value: "All assigned", Detail: "Every open bead has an owner",
		})
	} else if d.UnassignedCount <= 10 {
		signals = append(signals, signal{
			Name: "Coverage", Status: "yellow",
			Value:  fmt.Sprintf("%d unassigned", d.UnassignedCount),
			Detail: "Some beads need an owner",
		})
	} else {
		signals = append(signals, signal{
			Name: "Coverage", Status: "red",
			Value:  fmt.Sprintf("%d unassigned", d.UnassignedCount),
			Detail: "Many beads without owners — work falling through cracks",
		})
	}

	// 6. Active work ratio
	if d.OpenCount > 0 {
		activePct := float64(d.InProgressCount) / float64(d.OpenCount) * 100
		if activePct >= 20 {
			signals = append(signals, signal{
				Name: "Engagement", Status: "green",
				Value:  fmt.Sprintf("%.0f%% active", activePct),
				Detail: fmt.Sprintf("%d in-progress of %d open", d.InProgressCount, d.OpenCount),
			})
		} else if activePct >= 10 {
			signals = append(signals, signal{
				Name: "Engagement", Status: "yellow",
				Value:  fmt.Sprintf("%.0f%% active", activePct),
				Detail: fmt.Sprintf("%d in-progress of %d open — could be more active", d.InProgressCount, d.OpenCount),
			})
		} else {
			signals = append(signals, signal{
				Name: "Engagement", Status: "red",
				Value:  fmt.Sprintf("%.0f%% active", activePct),
				Detail: fmt.Sprintf("%d in-progress of %d open — most work sitting idle", d.InProgressCount, d.OpenCount),
			})
		}
	}

	return signals
}
