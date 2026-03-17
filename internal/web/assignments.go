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

type assignmentBead struct {
	ID       string
	DB       string
	Title    string
	Priority int
	Status   string
	AgeDays  int
}

type agentAssignment struct {
	Name  string
	Beads []assignmentBead
	Count int
}

type assignmentsData struct {
	GeneratedAt time.Time
	Agents      []agentAssignment
	Unassigned  []assignmentBead
	TotalBeads  int
	Rigs        []string
	FilterRig   string
	Err         string
}

func (s *Server) handleAssignments(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	data := assignmentsData{GeneratedAt: now}

	if s.ds == nil {
		s.render(w, r, "assignments", data)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("assignments: list dbs: %v", err)
		s.render(w, r, "assignments", assignmentsData{Err: err.Error(), GeneratedAt: now})
		return
	}

	filterRig := r.URL.Query().Get("rig")
	data.FilterRig = filterRig

	agentMap := make(map[string][]assignmentBead)
	var unassigned []assignmentBead
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
				log.Printf("assignments %s: %v", dbName, err)
				return
			}

			var localAgent map[string][]assignmentBead
			localAgent = make(map[string][]assignmentBead)
			var localUnassigned []assignmentBead
			hasData := false

			for _, iss := range issues {
				if isNoise(iss.ID, iss.Title) {
					continue
				}
				if iss.Status == "closed" || iss.Status == "deferred" {
					continue
				}

				hasData = true
				bead := assignmentBead{
					ID:       iss.ID,
					DB:       dbName,
					Title:    iss.Title,
					Priority: iss.Priority,
					Status:   iss.Status,
					AgeDays:  int(now.Sub(iss.CreatedAt).Hours() / 24),
				}

				if iss.Assignee == "" {
					localUnassigned = append(localUnassigned, bead)
				} else {
					localAgent[iss.Assignee] = append(localAgent[iss.Assignee], bead)
				}
			}

			mu.Lock()
			defer mu.Unlock()
			if hasData {
				rigSet[dbName] = true
			}
			unassigned = append(unassigned, localUnassigned...)
			for agent, beads := range localAgent {
				agentMap[agent] = append(agentMap[agent], beads...)
			}
		}(db.Name)
	}
	wg.Wait()

	// Sort beads within each agent by priority then age
	sortBeads := func(beads []assignmentBead) {
		sort.Slice(beads, func(i, j int) bool {
			if beads[i].Priority != beads[j].Priority {
				return beads[i].Priority < beads[j].Priority
			}
			return beads[i].AgeDays > beads[j].AgeDays
		})
	}

	var agents []agentAssignment
	total := 0
	for name, beads := range agentMap {
		sortBeads(beads)
		agents = append(agents, agentAssignment{
			Name:  name,
			Beads: beads,
			Count: len(beads),
		})
		total += len(beads)
	}

	sort.Slice(agents, func(i, j int) bool {
		return agents[i].Count > agents[j].Count
	})

	sortBeads(unassigned)
	total += len(unassigned)

	data.Agents = agents
	data.Unassigned = unassigned
	data.TotalBeads = total

	for rig := range rigSet {
		data.Rigs = append(data.Rigs, rig)
	}
	sort.Strings(data.Rigs)

	s.render(w, r, "assignments", data)
}
