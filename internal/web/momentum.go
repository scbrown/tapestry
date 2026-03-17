package web

import (
	"context"
	"log"
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/scbrown/tapestry/internal/dolt"
)

type momentumSignal struct {
	Name   string // signal name
	Value  string // display value
	Status string // "green", "yellow", "red"
	Detail string // explanation
}

type momentumData struct {
	GeneratedAt time.Time
	Signals     []momentumSignal

	// Raw metrics for display
	ClosedThisWeek int
	ClosedLastWeek int
	CreatedThisWeek int
	CreatedLastWeek int
	VelocityDelta  int // this week - last week
	BlockedCount   int
	ActiveCount    int
	BlockerRatio   float64 // blocked / (active+blocked)
	StaleCount     int
	OpenCount      int
	StalePct       float64 // stale / open
	NetFlowWeek    int     // created - closed this week

	Err string
}

func (s *Server) handleMomentum(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	thisWeekStart := todayStart.AddDate(0, 0, -7)
	lastWeekStart := thisWeekStart.AddDate(0, 0, -7)

	data := momentumData{GeneratedAt: now}

	if s.ds == nil {
		s.render(w, r, "momentum", data)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("momentum: list dbs: %v", err)
		s.render(w, r, "momentum", momentumData{Err: err.Error(), GeneratedAt: now})
		return
	}

	type dbResult struct {
		closedThisWeek  int
		closedLastWeek  int
		createdThisWeek int
		createdLastWeek int
		blockedCount    int
		activeCount     int
		staleCount      int
		openCount       int
	}

	staleThreshold := now.AddDate(0, 0, -7)
	results := make([]dbResult, len(dbs))
	var wg sync.WaitGroup
	for i, db := range dbs {
		wg.Add(1)
		go func(i int, dbName string) {
			defer wg.Done()
			var r dbResult

			// This week
			r.closedThisWeek, _ = s.ds.CountClosedInRange(ctx, dbName, thisWeekStart, todayStart.AddDate(0, 0, 1))
			r.createdThisWeek, _ = s.ds.CountCreatedInRange(ctx, dbName, thisWeekStart, todayStart.AddDate(0, 0, 1))

			// Last week
			r.closedLastWeek, _ = s.ds.CountClosedInRange(ctx, dbName, lastWeekStart, thisWeekStart)
			r.createdLastWeek, _ = s.ds.CountCreatedInRange(ctx, dbName, lastWeekStart, thisWeekStart)

			// Current status counts
			counts, err := s.ds.CountByStatus(ctx, dbName)
			if err == nil {
				r.blockedCount = counts["blocked"]
				r.activeCount = counts["in_progress"] + counts["hooked"]
				r.openCount = counts["open"]
			}

			// Stale items
			issues, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{Limit: 500})
			if err == nil {
				for _, iss := range issues {
					if isNoise(iss.ID, iss.Title) || iss.Status == "closed" || iss.Status == "deferred" {
						continue
					}
					if (iss.Status == "open" || iss.Status == "in_progress") && iss.UpdatedAt.Before(staleThreshold) {
						r.staleCount++
					}
				}
			}

			results[i] = r
		}(i, db.Name)
	}
	wg.Wait()

	// Aggregate
	for _, r := range results {
		data.ClosedThisWeek += r.closedThisWeek
		data.ClosedLastWeek += r.closedLastWeek
		data.CreatedThisWeek += r.createdThisWeek
		data.CreatedLastWeek += r.createdLastWeek
		data.BlockedCount += r.blockedCount
		data.ActiveCount += r.activeCount
		data.StaleCount += r.staleCount
		data.OpenCount += r.openCount
	}

	data.VelocityDelta = data.ClosedThisWeek - data.ClosedLastWeek
	data.NetFlowWeek = data.CreatedThisWeek - data.ClosedThisWeek

	totalActive := data.ActiveCount + data.BlockedCount
	if totalActive > 0 {
		data.BlockerRatio = float64(data.BlockedCount) / float64(totalActive)
	}
	if data.OpenCount > 0 {
		data.StalePct = float64(data.StaleCount) / float64(data.OpenCount)
	}

	// Build signals
	data.Signals = buildSignals(data)

	s.render(w, r, "momentum", data)
}

func buildSignals(d momentumData) []momentumSignal {
	var signals []momentumSignal

	// 1. Velocity trend
	velStatus := "green"
	velDetail := "Closing rate is steady or improving"
	if d.ClosedLastWeek > 0 {
		pctChange := float64(d.VelocityDelta) / float64(d.ClosedLastWeek) * 100
		if pctChange < -30 {
			velStatus = "red"
			velDetail = "Closing rate dropped significantly vs last week"
		} else if pctChange < -10 {
			velStatus = "yellow"
			velDetail = "Closing rate slightly down vs last week"
		}
	} else if d.ClosedThisWeek == 0 {
		velStatus = "red"
		velDetail = "No closures this week or last week"
	}
	velValue := formatDelta(d.VelocityDelta)
	signals = append(signals, momentumSignal{
		Name: "Velocity", Value: velValue, Status: velStatus, Detail: velDetail,
	})

	// 2. Net flow (are we getting ahead or falling behind?)
	flowStatus := "green"
	flowDetail := "Closing more than creating — backlog shrinking"
	if d.NetFlowWeek > 5 {
		flowStatus = "red"
		flowDetail = "Creating much more than closing — backlog growing"
	} else if d.NetFlowWeek > 0 {
		flowStatus = "yellow"
		flowDetail = "Slightly more created than closed"
	}
	flowValue := formatDelta(d.NetFlowWeek)
	signals = append(signals, momentumSignal{
		Name: "Net Flow", Value: flowValue, Status: flowStatus, Detail: flowDetail,
	})

	// 3. Blocker ratio
	blkStatus := "green"
	blkDetail := "Few items are blocked"
	blkPct := int(math.Round(d.BlockerRatio * 100))
	if d.BlockerRatio > 0.3 {
		blkStatus = "red"
		blkDetail = "High proportion of work is blocked"
	} else if d.BlockerRatio > 0.15 {
		blkStatus = "yellow"
		blkDetail = "Moderate proportion of work is blocked"
	}
	signals = append(signals, momentumSignal{
		Name: "Blocker Ratio", Value: itoa(blkPct) + "%", Status: blkStatus, Detail: blkDetail,
	})

	// 4. Staleness index
	staleStatus := "green"
	staleDetail := "Most open work is fresh"
	stalePct := int(math.Round(d.StalePct * 100))
	if d.StalePct > 0.4 {
		staleStatus = "red"
		staleDetail = "Many open items are stale (7+ days idle)"
	} else if d.StalePct > 0.2 {
		staleStatus = "yellow"
		staleDetail = "Some open items going stale"
	}
	signals = append(signals, momentumSignal{
		Name: "Staleness", Value: itoa(stalePct) + "%", Status: staleStatus, Detail: staleDetail,
	})

	return signals
}

func formatDelta(n int) string {
	if n > 0 {
		return "+" + itoa(n)
	}
	return itoa(n)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	sign := ""
	if n < 0 {
		sign = "-"
		n = -n
	}
	digits := make([]byte, 0, 10)
	for n > 0 {
		digits = append(digits, byte('0'+n%10))
		n /= 10
	}
	// reverse
	for i, j := 0, len(digits)-1; i < j; i, j = i+1, j-1 {
		digits[i], digits[j] = digits[j], digits[i]
	}
	return sign + string(digits)
}
