package web

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/scbrown/tapestry/internal/dolt"
)

type loadBalanceAgent struct {
	Name       string
	Open       int
	InProgress int
	Blocked    int
	HighPri    int // P0+P1
	Total      int
	Score      float64 // composite load score
	ScorePct   float64 // 0-100 for bar width
	Status     string  // "overloaded", "busy", "balanced", "idle"
}

type loadBalData struct {
	GeneratedAt time.Time
	Agents      []loadBalanceAgent
	Total       int
	AvgScore    float64
	Rigs        []string
	FilterRig   string
	SortBy      string
	Err         string
}

func (s *Server) handleLoadBalance(w http.ResponseWriter, r *http.Request) {
	data := loadBalData{GeneratedAt: time.Now()}

	if s.ds == nil {
		s.render(w, r, "load-balance", data)
		return
	}

	ctx := r.Context()
	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("load-balance: list dbs: %v", err)
		s.render(w, r, "load-balance", data)
		return
	}

	filterRig := r.URL.Query().Get("rig")
	data.FilterRig = filterRig

	var rigs []string
	for _, db := range dbs {
		rigs = append(rigs, db.Name)
	}
	sort.Strings(rigs)
	data.Rigs = rigs

	type agentIssues struct {
		open       int
		inProgress int
		blocked    int
		highPri    int
	}
	agentMap := make(map[string]*agentIssues)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, db := range dbs {
		if filterRig != "" && db.Name != filterRig {
			continue
		}
		wg.Add(1)
		go func(dbName string) {
			defer wg.Done()
			issues, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{Limit: 5000})
			if err != nil {
				log.Printf("load-balance: issues %s: %v", dbName, err)
				return
			}

			mu.Lock()
			defer mu.Unlock()
			for _, iss := range issues {
				if iss.Status == "closed" || iss.Status == "deferred" || isNoise(iss.ID, iss.Title) {
					continue
				}
				assignee := iss.Assignee
				if assignee == "" {
					assignee = iss.Owner
				}
				if assignee == "" {
					continue
				}
				a, ok := agentMap[assignee]
				if !ok {
					a = &agentIssues{}
					agentMap[assignee] = a
				}
				switch iss.Status {
				case "open":
					a.open++
				case "in_progress", "hooked":
					a.inProgress++
				case "blocked":
					a.blocked++
				}
				if iss.Priority <= 1 {
					a.highPri++
				}
			}
		}(db.Name)
	}
	wg.Wait()

	var agents []loadBalanceAgent
	for name, a := range agentMap {
		total := a.open + a.inProgress + a.blocked
		score := loadScore(a.open, a.inProgress, a.blocked, a.highPri)
		agents = append(agents, loadBalanceAgent{
			Name:       name,
			Open:       a.open,
			InProgress: a.inProgress,
			Blocked:    a.blocked,
			HighPri:    a.highPri,
			Total:      total,
			Score:      score,
		})
	}

	sortBy := r.URL.Query().Get("sort")
	if sortBy == "" {
		sortBy = "score"
	}
	data.SortBy = sortBy

	switch sortBy {
	case "total":
		sort.Slice(agents, func(i, j int) bool {
			return agents[i].Total > agents[j].Total
		})
	case "active":
		sort.Slice(agents, func(i, j int) bool {
			return agents[i].InProgress > agents[j].InProgress
		})
	case "blocked":
		sort.Slice(agents, func(i, j int) bool {
			return agents[i].Blocked > agents[j].Blocked
		})
	case "highpri":
		sort.Slice(agents, func(i, j int) bool {
			return agents[i].HighPri > agents[j].HighPri
		})
	case "name":
		sort.Slice(agents, func(i, j int) bool {
			return agents[i].Name < agents[j].Name
		})
	default: // "score"
		sort.Slice(agents, func(i, j int) bool {
			return agents[i].Score > agents[j].Score
		})
	}

	// Compute bar widths and statuses
	if len(agents) > 0 {
		maxScore := agents[0].Score
		if maxScore < 1 {
			maxScore = 1
		}
		var totalScore float64
		for i := range agents {
			agents[i].ScorePct = math.Min(100, (agents[i].Score/maxScore)*100)
			totalScore += agents[i].Score
			switch {
			case agents[i].Score > maxScore*0.8:
				agents[i].Status = "overloaded"
			case agents[i].Score > maxScore*0.5:
				agents[i].Status = "busy"
			case agents[i].Score > maxScore*0.2:
				agents[i].Status = "balanced"
			default:
				agents[i].Status = "idle"
			}
		}
		data.AvgScore = totalScore / float64(len(agents))
	}

	data.Agents = agents
	data.Total = len(agents)
	s.render(w, r, "load-balance", data)
}

func loadScore(open, inProgress, blocked, highPri int) float64 {
	return float64(inProgress)*3 + float64(blocked)*2 + float64(highPri)*2 + float64(open)*0.5
}

func fmtScore(s float64) string {
	return fmt.Sprintf("%.0f", s)
}
