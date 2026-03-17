package web

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/scbrown/tapestry/internal/dolt"
)

type ratio struct {
	Name       string
	Numerator  int
	Denominator int
	Value      float64
	Health     string // "good", "warning", "danger"
	Desc       string
}

type ratiosData struct {
	GeneratedAt time.Time
	Ratios      []ratio
	FilterRig   string
	Err         string
}

func (s *Server) handleRatios(w http.ResponseWriter, r *http.Request) {
	data := ratiosData{GeneratedAt: time.Now()}

	if s.ds == nil {
		s.render(w, r, "ratios", data)
		return
	}

	ctx := r.Context()
	data.FilterRig = r.URL.Query().Get("rig")

	dbs, err := s.databases(ctx)
	if err != nil {
		data.Err = err.Error()
		s.render(w, r, "ratios", data)
		return
	}

	now := time.Now()
	sevenDaysAgo := now.AddDate(0, 0, -7)
	thirtyDaysAgo := now.AddDate(0, 0, -30)
	todayEnd := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.UTC)

	var totalBugs, totalTasks, totalEpics int
	var totalOpen, totalClosed, totalBlocked, totalDeferred int
	var created7d, closed7d, created30d, closed30d int
	var totalComments int

	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, db := range dbs {
		if data.FilterRig != "" && db.Name != data.FilterRig {
			continue
		}
		wg.Add(1)
		go func(dbName string) {
			defer wg.Done()

			issues, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{Limit: 5000})
			if err != nil {
				log.Printf("ratios: issues %s: %v", dbName, err)
				return
			}

			var bugs, tasks, epics int
			for _, iss := range issues {
				switch iss.Type {
				case "bug":
					bugs++
				case "task":
					tasks++
				case "epic":
					epics++
				}
			}

			counts, err := s.ds.CountByStatus(ctx, dbName)
			if err != nil {
				log.Printf("ratios: counts %s: %v", dbName, err)
				return
			}

			c7, _ := s.ds.CountCreatedInRange(ctx, dbName, sevenDaysAgo, todayEnd)
			cl7, _ := s.ds.CountClosedInRange(ctx, dbName, sevenDaysAgo, todayEnd)
			c30, _ := s.ds.CountCreatedInRange(ctx, dbName, thirtyDaysAgo, todayEnd)
			cl30, _ := s.ds.CountClosedInRange(ctx, dbName, thirtyDaysAgo, todayEnd)

			comments, _ := s.ds.RecentComments(ctx, dbName, 5000)

			mu.Lock()
			totalBugs += bugs
			totalTasks += tasks
			totalEpics += epics
			totalOpen += counts["open"]
			totalClosed += counts["closed"] + counts["completed"]
			totalBlocked += counts["blocked"]
			totalDeferred += counts["deferred"]
			created7d += c7
			closed7d += cl7
			created30d += c30
			closed30d += cl30
			totalComments += len(comments)
			mu.Unlock()
		}(db.Name)
	}
	wg.Wait()

	totalActive := totalOpen + totalBlocked

	// Bug-to-feature ratio
	bugFeatureRatio := safeDiv(float64(totalBugs), float64(totalTasks+totalEpics))
	bugHealth := "good"
	if bugFeatureRatio > 0.5 {
		bugHealth = "danger"
	} else if bugFeatureRatio > 0.3 {
		bugHealth = "warning"
	}

	// Close rate (7d)
	closeRate7d := safeDiv(float64(closed7d), float64(created7d))
	closeHealth7d := "good"
	if closeRate7d < 0.5 {
		closeHealth7d = "danger"
	} else if closeRate7d < 0.8 {
		closeHealth7d = "warning"
	}

	// Close rate (30d)
	closeRate30d := safeDiv(float64(closed30d), float64(created30d))
	closeHealth30d := "good"
	if closeRate30d < 0.5 {
		closeHealth30d = "danger"
	} else if closeRate30d < 0.8 {
		closeHealth30d = "warning"
	}

	// Blocker ratio
	blockerRatio := safeDiv(float64(totalBlocked), float64(totalActive))
	blockerHealth := "good"
	if blockerRatio > 0.3 {
		blockerHealth = "danger"
	} else if blockerRatio > 0.15 {
		blockerHealth = "warning"
	}

	// Deferred ratio
	deferredRatio := safeDiv(float64(totalDeferred), float64(totalActive+totalDeferred))
	deferredHealth := "good"
	if deferredRatio > 0.4 {
		deferredHealth = "danger"
	} else if deferredRatio > 0.2 {
		deferredHealth = "warning"
	}

	// Completion ratio
	completionRatio := safeDiv(float64(totalClosed), float64(totalClosed+totalActive))
	completionHealth := "good"
	if completionRatio < 0.3 {
		completionHealth = "danger"
	} else if completionRatio < 0.5 {
		completionHealth = "warning"
	}

	data.Ratios = []ratio{
		{"Bug:Feature", totalBugs, totalTasks + totalEpics, bugFeatureRatio, bugHealth, "Bugs per task/epic. Lower is better."},
		{"Close Rate (7d)", closed7d, created7d, closeRate7d, closeHealth7d, "Closures per creation this week. Above 1.0 means backlog shrinks."},
		{"Close Rate (30d)", closed30d, created30d, closeRate30d, closeHealth30d, "Closures per creation this month."},
		{"Blocker Ratio", totalBlocked, totalActive, blockerRatio, blockerHealth, "Fraction of active work that's blocked."},
		{"Deferred Ratio", totalDeferred, totalActive + totalDeferred, deferredRatio, deferredHealth, "Fraction of non-closed work that's deferred."},
		{"Completion", totalClosed, totalClosed + totalActive, completionRatio, completionHealth, "Fraction of all beads that are closed."},
	}

	s.render(w, r, "ratios", data)
}

func safeDiv(a, b float64) float64 {
	if b == 0 {
		return 0
	}
	return a / b
}
