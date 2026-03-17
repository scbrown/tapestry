package web

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/scbrown/tapestry/internal/dolt"
)

type comparePeriod struct {
	Label   string
	Start   time.Time
	End     time.Time
	Created int
	Closed  int
	Net     int
}

type compareData struct {
	GeneratedAt time.Time
	PeriodA     comparePeriod
	PeriodB     comparePeriod

	// Deltas (A - B)
	DeltaCreated int
	DeltaClosed  int
	DeltaNet     int

	// Agent breakdown per period
	AgentsA map[string]int // agent -> closed count in period A
	AgentsB map[string]int // agent -> closed count in period B

	// Top agents sorted
	TopAgentsA []agentCount
	TopAgentsB []agentCount

	Days int // period length in days
	Err  string
}

type agentCount struct {
	Agent string
	Count int
}

func (s *Server) handleCompare(w http.ResponseWriter, r *http.Request) {
	now := time.Now()

	// Default: 7-day periods (this week vs last week)
	days := 7
	if d := r.URL.Query().Get("days"); d != "" {
		if v, err := strconv.Atoi(d); err == nil && v > 0 && v <= 90 {
			days = v
		}
	}

	periodAEnd := now
	periodAStart := now.Add(-time.Duration(days) * 24 * time.Hour)
	periodBEnd := periodAStart
	periodBStart := periodBEnd.Add(-time.Duration(days) * 24 * time.Hour)

	data := compareData{
		GeneratedAt: now,
		Days:        days,
		PeriodA: comparePeriod{
			Label: "Current",
			Start: periodAStart,
			End:   periodAEnd,
		},
		PeriodB: comparePeriod{
			Label: "Previous",
			Start: periodBStart,
			End:   periodBEnd,
		},
	}

	if s.ds == nil {
		s.render(w, r, "compare", data)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("compare: list dbs: %v", err)
		data.Err = err.Error()
		s.render(w, r, "compare", data)
		return
	}

	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, db := range dbs {
		wg.Add(1)
		go func(dbName string) {
			defer wg.Done()

			// Count created/closed for period A
			createdA, _ := s.ds.CountCreatedInRange(ctx, dbName, periodAStart, periodAEnd)
			closedA, _ := s.ds.CountClosedInRange(ctx, dbName, periodAStart, periodAEnd)

			// Count created/closed for period B
			createdB, _ := s.ds.CountCreatedInRange(ctx, dbName, periodBStart, periodBEnd)
			closedB, _ := s.ds.CountClosedInRange(ctx, dbName, periodBStart, periodBEnd)

			// Agent activity for each period
			agentsA, _ := s.ds.AgentActivityInRange(ctx, dbName, periodAStart, periodAEnd)
			agentsB, _ := s.ds.AgentActivityInRange(ctx, dbName, periodBStart, periodBEnd)

			mu.Lock()
			data.PeriodA.Created += createdA
			data.PeriodA.Closed += closedA
			data.PeriodB.Created += createdB
			data.PeriodB.Closed += closedB

			if data.AgentsA == nil {
				data.AgentsA = make(map[string]int)
			}
			if data.AgentsB == nil {
				data.AgentsB = make(map[string]int)
			}
			for agent, count := range agentsA {
				data.AgentsA[agent] += count
			}
			for agent, count := range agentsB {
				data.AgentsB[agent] += count
			}
			mu.Unlock()
		}(db.Name)
	}
	wg.Wait()

	data.PeriodA.Net = data.PeriodA.Created - data.PeriodA.Closed
	data.PeriodB.Net = data.PeriodB.Created - data.PeriodB.Closed
	data.DeltaCreated = data.PeriodA.Created - data.PeriodB.Created
	data.DeltaClosed = data.PeriodA.Closed - data.PeriodB.Closed
	data.DeltaNet = data.PeriodA.Net - data.PeriodB.Net

	// Build sorted agent lists
	data.TopAgentsA = sortAgentCounts(data.AgentsA)
	data.TopAgentsB = sortAgentCounts(data.AgentsB)

	// Also count from Issues for agent breakdown of closed work
	if len(data.TopAgentsA) == 0 && len(data.TopAgentsB) == 0 {
		// AgentActivityInRange may only count open issue assignments
		// Use Issues query as fallback for agent data
		for _, db := range dbs {
			issues, err := s.ds.Issues(ctx, db.Name, dolt.IssueFilter{Status: "closed", Limit: 1000})
			if err != nil {
				continue
			}
			for _, iss := range issues {
				if isNoise(iss.ID, iss.Title) || iss.Assignee == "" {
					continue
				}
				if iss.UpdatedAt.After(periodAStart) && iss.UpdatedAt.Before(periodAEnd) {
					data.AgentsA[iss.Assignee]++
				}
				if iss.UpdatedAt.After(periodBStart) && iss.UpdatedAt.Before(periodBEnd) {
					data.AgentsB[iss.Assignee]++
				}
			}
		}
		data.TopAgentsA = sortAgentCounts(data.AgentsA)
		data.TopAgentsB = sortAgentCounts(data.AgentsB)
	}

	s.render(w, r, "compare", data)
}

func sortAgentCounts(m map[string]int) []agentCount {
	var result []agentCount
	for agent, count := range m {
		result = append(result, agentCount{Agent: agent, Count: count})
	}
	// Sort by count descending
	for i := 0; i < len(result); i++ {
		for j := i + 1; j < len(result); j++ {
			if result[j].Count > result[i].Count {
				result[i], result[j] = result[j], result[i]
			}
		}
	}
	if len(result) > 10 {
		result = result[:10]
	}
	return result
}
