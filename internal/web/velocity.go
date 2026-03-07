package web

import (
	"context"
	"log"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/scbrown/tapestry/internal/dolt"
)

type velocityDay struct {
	Date    time.Time
	Created int
	Closed  int
}

type velocityAgent struct {
	Name    string
	Created int
	Closed  int
	Net     int
}

type velocityData struct {
	GeneratedAt time.Time
	Days        []velocityDay
	MaxCount    int
	TotalCreate int
	TotalClose  int
	NetChange   int
	AvgCreate   float64
	AvgClose    float64
	Agents      []velocityAgent
}

func (s *Server) handleVelocity(w http.ResponseWriter, r *http.Request) {
	data := velocityData{GeneratedAt: time.Now()}

	if s.ds == nil {
		s.render(w, r, "velocity", data)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 20*time.Second)
	defer cancel()

	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("velocity: list dbs: %v", err)
		s.render(w, r, "velocity", data)
		return
	}

	const numDays = 14
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	type dayResult struct {
		created int
		closed  int
	}

	type dbResult struct {
		days   [numDays]dayResult
		agents map[string][2]int // [created, closed]
	}

	results := make([]dbResult, len(dbs))
	var wg sync.WaitGroup
	for i, db := range dbs {
		wg.Add(1)
		go func(i int, dbName string) {
			defer wg.Done()
			var r dbResult
			r.agents = make(map[string][2]int)

			for d := 0; d < numDays; d++ {
				dayStart := todayStart.AddDate(0, 0, -d)
				dayEnd := dayStart.AddDate(0, 0, 1)

				created, err := s.ds.CountCreatedInRange(ctx, dbName, dayStart, dayEnd)
				if err != nil {
					log.Printf("velocity: created %s day %d: %v", dbName, d, err)
				}
				closed, err := s.ds.CountClosedInRange(ctx, dbName, dayStart, dayEnd)
				if err != nil {
					log.Printf("velocity: closed %s day %d: %v", dbName, d, err)
				}
				r.days[d] = dayResult{created: created, closed: closed}
			}

			// Agent activity over the full period
			periodStart := todayStart.AddDate(0, 0, -(numDays - 1))
			periodEnd := todayStart.AddDate(0, 0, 1)

			// Count closed per agent
			closedIssues, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{
				Status:       "closed",
				UpdatedAfter: periodStart,
				UpdatedBefore: periodEnd,
				Limit:        500,
			})
			if err == nil {
				for _, iss := range closedIssues {
					agent := iss.Assignee
					if agent == "" {
						agent = iss.Owner
					}
					if agent == "" {
						continue
					}
					v := r.agents[agent]
					v[1]++
					r.agents[agent] = v
				}
			}

			// Count created per agent (owner)
			allIssues, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{
				Limit: 500,
			})
			if err == nil {
				for _, iss := range allIssues {
					if iss.CreatedAt.Before(periodStart) || !iss.CreatedAt.Before(periodEnd) {
						continue
					}
					agent := iss.Owner
					if agent == "" {
						continue
					}
					v := r.agents[agent]
					v[0]++
					r.agents[agent] = v
				}
			}

			results[i] = r
		}(i, db.Name)
	}
	wg.Wait()

	// Aggregate across databases
	var days [numDays]dayResult
	agentTotals := make(map[string][2]int)
	for _, r := range results {
		for d := 0; d < numDays; d++ {
			days[d].created += r.days[d].created
			days[d].closed += r.days[d].closed
		}
		for agent, counts := range r.agents {
			v := agentTotals[agent]
			v[0] += counts[0]
			v[1] += counts[1]
			agentTotals[agent] = v
		}
	}

	// Build day list (reverse so oldest is first)
	for d := numDays - 1; d >= 0; d-- {
		day := velocityDay{
			Date:    todayStart.AddDate(0, 0, -d),
			Created: days[d].created,
			Closed:  days[d].closed,
		}
		data.Days = append(data.Days, day)
		data.TotalCreate += day.Created
		data.TotalClose += day.Closed
		if day.Created > data.MaxCount {
			data.MaxCount = day.Created
		}
		if day.Closed > data.MaxCount {
			data.MaxCount = day.Closed
		}
	}

	data.NetChange = data.TotalCreate - data.TotalClose
	if numDays > 0 {
		data.AvgCreate = float64(data.TotalCreate) / float64(numDays)
		data.AvgClose = float64(data.TotalClose) / float64(numDays)
	}

	// Build agent list
	for agent, counts := range agentTotals {
		if isNoise("", agent) {
			continue
		}
		data.Agents = append(data.Agents, velocityAgent{
			Name:    agent,
			Created: counts[0],
			Closed:  counts[1],
			Net:     counts[0] - counts[1],
		})
	}
	sort.Slice(data.Agents, func(i, j int) bool {
		return data.Agents[i].Closed > data.Agents[j].Closed
	})

	s.render(w, r, "velocity", data)
}
