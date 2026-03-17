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

type overflowAgent struct {
	Name       string
	OpenCount  int
	BlockedCnt int
	HighPriCnt int // P0+P1
	OldestDays int // age of oldest open bead
	Score      int // composite overload score
}

type overflowData struct {
	GeneratedAt time.Time
	Agents      []overflowAgent
	Threshold   int // highlight agents above this
	Rigs        []string
	FilterRig   string
	Err         string
}

func (s *Server) handleOverflow(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	data := overflowData{GeneratedAt: now, Threshold: 10}

	if s.ds == nil {
		s.render(w, r, "overflow", data)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("overflow: list dbs: %v", err)
		s.render(w, r, "overflow", overflowData{Err: err.Error(), GeneratedAt: now})
		return
	}

	filterRig := r.URL.Query().Get("rig")
	data.FilterRig = filterRig

	type agentAcc struct {
		open    int
		blocked int
		highPri int
		oldest  time.Time
	}

	agentMap := make(map[string]*agentAcc)
	rigSet := make(map[string]bool)
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
				log.Printf("overflow %s: %v", dbName, err)
				return
			}
			mu.Lock()
			defer mu.Unlock()
			for _, iss := range issues {
				if isNoise(iss.ID, iss.Title) || iss.Assignee == "" {
					continue
				}
				if iss.Status == "closed" || iss.Status == "deferred" {
					continue
				}
				rigSet[dbName] = true
				acc, ok := agentMap[iss.Assignee]
				if !ok {
					acc = &agentAcc{}
					agentMap[iss.Assignee] = acc
				}
				acc.open++
				if iss.Status == "blocked" {
					acc.blocked++
				}
				if iss.Priority >= 0 && iss.Priority <= 1 {
					acc.highPri++
				}
				if acc.oldest.IsZero() || iss.CreatedAt.Before(acc.oldest) {
					acc.oldest = iss.CreatedAt
				}
			}
		}(db.Name)
	}
	wg.Wait()

	for name, acc := range agentMap {
		oldestDays := 0
		if !acc.oldest.IsZero() {
			oldestDays = int(now.Sub(acc.oldest).Hours() / 24)
		}
		// Score: open count weighted by high-pri and blocked items
		score := acc.open + acc.highPri*2 + acc.blocked
		data.Agents = append(data.Agents, overflowAgent{
			Name:       name,
			OpenCount:  acc.open,
			BlockedCnt: acc.blocked,
			HighPriCnt: acc.highPri,
			OldestDays: oldestDays,
			Score:      score,
		})
	}

	sort.Slice(data.Agents, func(i, j int) bool {
		return data.Agents[i].Score > data.Agents[j].Score
	})

	for rig := range rigSet {
		data.Rigs = append(data.Rigs, rig)
	}
	sort.Strings(data.Rigs)

	s.render(w, r, "overflow", data)
}
