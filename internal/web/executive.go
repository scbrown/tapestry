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

type execDay struct {
	Date    time.Time
	Created int
	Closed  int
}

type execPriorityBucket struct {
	Label string
	Count int
}

type execBlocker struct {
	Issue       dolt.Issue
	BlockerID   string
	BlockerDesc string
	BlockerRig  string
	Owner       string
}

type executiveData struct {
	GeneratedAt time.Time

	// Top-line metrics
	OpenCount       int
	InProgressCount int
	BlockedCount    int
	ClosedCount     int
	TotalBeads      int
	ClosedToday     int
	CreatedToday    int
	ClosedWeek      int
	CreatedWeek     int

	// 7-day throughput
	Days     []execDay
	MaxCount int

	// Priority breakdown (open + in_progress only)
	Priorities []execPriorityBucket

	// Top blockers needing human action
	Blockers []execBlocker

	// Epic progress
	Epics []epicTree

	// Agent leaderboard (closed in last 7 days)
	Agents []velocityAgent
}

func (s *Server) handleExecutive(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	todayEnd := todayStart.AddDate(0, 0, 1)
	weekStart := todayStart.AddDate(0, 0, -7)

	data := executiveData{GeneratedAt: now}

	if s.ds == nil {
		s.render(w, r, "executive", data)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 20*time.Second)
	defer cancel()

	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("executive: list dbs: %v", err)
		s.render(w, r, "executive", data)
		return
	}

	type execDBResult struct {
		openCount       int
		inProgressCount int
		blockedCount    int
		closedCount     int
		totalBeads      int
		createdToday    int
		closedToday     int
		createdWeek     int
		closedWeek      int
		days            [7]struct{ created, closed int }
		priorities      [5]int // P0-P4 counts for open/in_progress
		blockers        []execBlocker
		epics           []epicTree
		agentClosed     map[string]int
	}

	results := make([]execDBResult, len(dbs))
	var wg sync.WaitGroup
	for i, db := range dbs {
		wg.Add(1)
		go func(i int, dbName string) {
			defer wg.Done()
			r := execDBResult{agentClosed: make(map[string]int)}

			// Status counts
			counts, err := s.ds.CountByStatus(ctx, dbName)
			if err != nil {
				log.Printf("executive: counts %s: %v", dbName, err)
				results[i] = r
				return
			}
			r.openCount = counts["open"]
			r.inProgressCount = counts["in_progress"] + counts["hooked"]
			r.blockedCount = counts["blocked"]
			r.closedCount = counts["closed"] + counts["completed"]
			for _, v := range counts {
				r.totalBeads += v
			}

			// Today counts
			created, err := s.ds.CountCreatedInRange(ctx, dbName, todayStart, todayEnd)
			if err == nil {
				r.createdToday = created
			}
			closed, err := s.ds.CountClosedInRange(ctx, dbName, todayStart, todayEnd)
			if err == nil {
				r.closedToday = closed
			}

			// Week counts
			createdW, err := s.ds.CountCreatedInRange(ctx, dbName, weekStart, todayEnd)
			if err == nil {
				r.createdWeek = createdW
			}
			closedW, err := s.ds.CountClosedInRange(ctx, dbName, weekStart, todayEnd)
			if err == nil {
				r.closedWeek = closedW
			}

			// 7-day throughput
			for d := 0; d < 7; d++ {
				dayStart := todayStart.AddDate(0, 0, -d)
				dayEnd := dayStart.AddDate(0, 0, 1)
				c, _ := s.ds.CountCreatedInRange(ctx, dbName, dayStart, dayEnd)
				cl, _ := s.ds.CountClosedInRange(ctx, dbName, dayStart, dayEnd)
				r.days[d] = struct{ created, closed int }{c, cl}
			}

			// Issues for priority breakdown
			issues, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{Limit: 500})
			if err != nil {
				results[i] = r
				return
			}
			for _, iss := range issues {
				if iss.Status == "closed" || iss.Status == "deferred" || isNoise(iss.ID, iss.Title) {
					continue
				}
				if iss.Priority >= 0 && iss.Priority <= 4 {
					r.priorities[iss.Priority]++
				}
			}

			// Blocked items needing human action
			blocked, err := s.ds.BlockedIssues(ctx, dbName)
			if err == nil {
				for _, bi := range blocked {
					if isNoise(bi.Issue.ID, bi.Issue.Title) || bi.Issue.Priority > 2 {
						continue
					}
					owner := bi.Blocker.Assignee
					if owner == "" {
						owner = bi.Blocker.Owner
					}
					bi.Issue.Rig = dbName
					r.blockers = append(r.blockers, execBlocker{
						Issue:       bi.Issue,
						BlockerID:   bi.Blocker.ID,
						BlockerDesc: bi.Blocker.Title,
						BlockerRig:  dbName,
						Owner:       owner,
					})
				}
			}

			// Epics with progress
			epics, err := s.ds.Epics(ctx, dbName)
			if err == nil {
				childDeps, _ := s.ds.AllChildDependencies(ctx, dbName)
				parentChildren := make(map[string][]string)
				for _, dep := range childDeps {
					parentChildren[dep.ToID] = append(parentChildren[dep.ToID], dep.FromID)
				}
				issueMap := make(map[string]dolt.Issue, len(issues))
				for _, iss := range issues {
					issueMap[iss.ID] = iss
				}
				// Also include closed issues in the map for progress calc
				closedIssues, _ := s.ds.Issues(ctx, dbName, dolt.IssueFilter{Status: "closed", Limit: 500})
				for _, iss := range closedIssues {
					issueMap[iss.ID] = iss
				}
				for _, epic := range epics {
					if epic.Status == "closed" || isNoise(epic.ID, epic.Title) {
						continue
					}
					et := epicTree{Epic: epic, Rig: dbName}
					for _, childID := range parentChildren[epic.ID] {
						if child, ok := issueMap[childID]; ok {
							et.Progress.Total++
							if child.Status == "closed" {
								et.Progress.Closed++
							}
						}
					}
					if et.Progress.Total > 0 {
						r.epics = append(r.epics, et)
					}
				}
			}

			// Agent closed counts (last 7 days)
			closedIssues, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{
				Status:       "closed",
				UpdatedAfter: weekStart,
				Limit:        500,
			})
			if err == nil {
				for _, iss := range closedIssues {
					agent := iss.Assignee
					if agent == "" {
						agent = iss.Owner
					}
					if agent != "" && !isNoise("", agent) {
						r.agentClosed[agent]++
					}
				}
			}

			results[i] = r
		}(i, db.Name)
	}
	wg.Wait()

	// Aggregate
	var priorities [5]int
	agentTotals := make(map[string]int)
	for _, r := range results {
		data.OpenCount += r.openCount
		data.InProgressCount += r.inProgressCount
		data.BlockedCount += r.blockedCount
		data.ClosedCount += r.closedCount
		data.TotalBeads += r.totalBeads
		data.CreatedToday += r.createdToday
		data.ClosedToday += r.closedToday
		data.CreatedWeek += r.createdWeek
		data.ClosedWeek += r.closedWeek
		data.Blockers = append(data.Blockers, r.blockers...)
		data.Epics = append(data.Epics, r.epics...)
		for p := 0; p < 5; p++ {
			priorities[p] += r.priorities[p]
		}
		for agent, count := range r.agentClosed {
			agentTotals[agent] += count
		}
	}

	// Build 7-day chart (oldest first)
	for d := 6; d >= 0; d-- {
		day := execDay{Date: todayStart.AddDate(0, 0, -d)}
		for _, r := range results {
			day.Created += r.days[d].created
			day.Closed += r.days[d].closed
		}
		data.Days = append(data.Days, day)
		if day.Created > data.MaxCount {
			data.MaxCount = day.Created
		}
		if day.Closed > data.MaxCount {
			data.MaxCount = day.Closed
		}
	}

	// Priority buckets
	labels := []string{"P0", "P1", "P2", "P3", "P4"}
	for p := 0; p < 5; p++ {
		data.Priorities = append(data.Priorities, execPriorityBucket{
			Label: labels[p],
			Count: priorities[p],
		})
	}

	// Sort blockers by priority
	sort.Slice(data.Blockers, func(i, j int) bool {
		return data.Blockers[i].Issue.Priority < data.Blockers[j].Issue.Priority
	})
	if len(data.Blockers) > 8 {
		data.Blockers = data.Blockers[:8]
	}

	// Sort epics by priority
	sort.Slice(data.Epics, func(i, j int) bool {
		if data.Epics[i].Epic.Priority != data.Epics[j].Epic.Priority {
			return data.Epics[i].Epic.Priority < data.Epics[j].Epic.Priority
		}
		return data.Epics[i].Epic.UpdatedAt.After(data.Epics[j].Epic.UpdatedAt)
	})
	if len(data.Epics) > 8 {
		data.Epics = data.Epics[:8]
	}

	// Agent leaderboard
	for agent, count := range agentTotals {
		data.Agents = append(data.Agents, velocityAgent{
			Name:   agent,
			Closed: count,
		})
	}
	sort.Slice(data.Agents, func(i, j int) bool {
		return data.Agents[i].Closed > data.Agents[j].Closed
	})
	if len(data.Agents) > 10 {
		data.Agents = data.Agents[:10]
	}

	s.render(w, r, "executive", data)
}
