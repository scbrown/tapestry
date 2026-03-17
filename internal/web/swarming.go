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

// swarmBead is a bead that multiple agents have touched
type swarmBead struct {
	ID        string
	Title     string
	Status    string
	Priority  int
	DB        string
	Agents    []string
	AgentCount int
	Comments  int
}

type swarmingData struct {
	GeneratedAt time.Time
	FilterRig   string

	// Beads with multiple agents assigned/commenting
	SwarmBeads []swarmBead

	// Stats
	TotalSwarmed int // beads with 2+ agents
	MaxAgents    int
	AvgAgents    float64

	Err string
}

func (s *Server) handleSwarming(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	filterRig := r.URL.Query().Get("rig")
	data := swarmingData{GeneratedAt: now, FilterRig: filterRig}

	if s.ds == nil {
		s.render(w, r, "swarming", data)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("swarming: list dbs: %v", err)
		data.Err = err.Error()
		s.render(w, r, "swarming", data)
		return
	}

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
				log.Printf("swarming %s: %v", dbName, err)
				return
			}

			for _, iss := range issues {
				if isNoise(iss.ID, iss.Title) {
					continue
				}
				if iss.Status == "closed" || iss.Status == "deferred" {
					continue
				}

				// Get comments to find all agents who've touched this bead
				comments, err := s.ds.Comments(ctx, dbName, iss.ID)
				if err != nil {
					continue
				}

				agents := make(map[string]bool)
				if iss.Assignee != "" {
					agents[iss.Assignee] = true
				}
				for _, c := range comments {
					if c.Author != "" {
						agents[c.Author] = true
					}
				}

				if len(agents) >= 2 {
					var agentList []string
					for a := range agents {
						agentList = append(agentList, a)
					}
					sort.Strings(agentList)

					mu.Lock()
					data.SwarmBeads = append(data.SwarmBeads, swarmBead{
						ID:         iss.ID,
						Title:      iss.Title,
						Status:     iss.Status,
						Priority:   iss.Priority,
						DB:         dbName,
						Agents:     agentList,
						AgentCount: len(agents),
						Comments:   len(comments),
					})
					mu.Unlock()
				}
			}
		}(db.Name)
	}
	wg.Wait()

	// Sort by agent count descending
	sort.Slice(data.SwarmBeads, func(i, j int) bool {
		if data.SwarmBeads[i].AgentCount != data.SwarmBeads[j].AgentCount {
			return data.SwarmBeads[i].AgentCount > data.SwarmBeads[j].AgentCount
		}
		return data.SwarmBeads[i].Comments > data.SwarmBeads[j].Comments
	})

	data.TotalSwarmed = len(data.SwarmBeads)
	if len(data.SwarmBeads) > 0 {
		total := 0
		for _, b := range data.SwarmBeads {
			total += b.AgentCount
			if b.AgentCount > data.MaxAgents {
				data.MaxAgents = b.AgentCount
			}
		}
		data.AvgAgents = float64(total) / float64(len(data.SwarmBeads))
	}

	// Limit display to top 50
	if len(data.SwarmBeads) > 50 {
		data.SwarmBeads = data.SwarmBeads[:50]
	}

	s.render(w, r, "swarming", data)
}
