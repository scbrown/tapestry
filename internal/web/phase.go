package web

import (
	"log"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/scbrown/tapestry/internal/dolt"
)

type epicProgress struct {
	ID         string
	Title      string
	DB         string
	Priority   int
	Total      int
	Closed     int
	InProgress int
	Blocked    int
	Open       int
	Deferred   int
	Pct        float64
}

type phaseData struct {
	GeneratedAt time.Time
	Epics       []epicProgress
	TotalEpics  int
	Completed   int // 100% done
	InFlight    int // >0% and <100%
	NotStarted  int // 0%
	Rigs        []string
	FilterRig   string
}

func (s *Server) handlePhase(w http.ResponseWriter, r *http.Request) {
	data := phaseData{GeneratedAt: time.Now()}

	if s.ds == nil {
		s.render(w, r, "phase", data)
		return
	}

	ctx := r.Context()
	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("phase: list dbs: %v", err)
		s.render(w, r, "phase", data)
		return
	}

	filterRig := r.URL.Query().Get("rig")
	data.FilterRig = filterRig

	type dbEpics struct {
		db    string
		epics []dolt.Issue
		deps  []dolt.Dependency
	}

	results := make([]dbEpics, len(dbs))
	var wg sync.WaitGroup
	for i, db := range dbs {
		if filterRig != "" && db.Name != filterRig {
			continue
		}
		wg.Add(1)
		go func(i int, dbName string) {
			defer wg.Done()
			epics, err := s.ds.Epics(ctx, dbName)
			if err != nil {
				log.Printf("phase: epics %s: %v", dbName, err)
				return
			}
			deps, err := s.ds.AllChildDependencies(ctx, dbName)
			if err != nil {
				log.Printf("phase: deps %s: %v", dbName, err)
				return
			}
			results[i] = dbEpics{db: dbName, epics: epics, deps: deps}
		}(i, db.Name)
	}
	wg.Wait()

	rigSet := make(map[string]bool)
	for _, db := range dbs {
		rigSet[db.Name] = true
	}
	for rig := range rigSet {
		data.Rigs = append(data.Rigs, rig)
	}
	sort.Strings(data.Rigs)

	// Build child map: parent -> []child issues
	// We need to look up children for each epic. Use AllChildDependencies.
	for _, res := range results {
		if len(res.epics) == 0 {
			continue
		}

		// Build parent->children map from deps
		childIDs := make(map[string][]string) // parent -> child IDs
		for _, dep := range res.deps {
			if dep.Type == "child_of" {
				childIDs[dep.ToID] = append(childIDs[dep.ToID], dep.FromID)
			}
		}

		// Fetch all issues once for status lookup
		allIssues, err := s.ds.Issues(ctx, res.db, dolt.IssueFilter{Limit: 10000})
		if err != nil {
			continue
		}
		issueMap := make(map[string]*dolt.Issue, len(allIssues))
		for i := range allIssues {
			issueMap[allIssues[i].ID] = &allIssues[i]
		}

		for _, epic := range res.epics {
			children := childIDs[epic.ID]
			ep := epicProgress{
				ID:       epic.ID,
				Title:    epic.Title,
				DB:       res.db,
				Priority: epic.Priority,
				Total:    len(children),
			}

			for _, cid := range children {
				child, ok := issueMap[cid]
				if !ok {
					ep.Open++
					continue
				}
				switch child.Status {
				case "closed":
					ep.Closed++
				case "in_progress", "hooked":
					ep.InProgress++
				case "blocked":
					ep.Blocked++
				case "deferred":
					ep.Deferred++
				default:
					ep.Open++
				}
			}

			if ep.Total > 0 {
				ep.Pct = float64(ep.Closed) / float64(ep.Total) * 100
			}
			data.Epics = append(data.Epics, ep)
		}
	}

	// Sort by priority, then pct descending
	sort.Slice(data.Epics, func(i, j int) bool {
		if data.Epics[i].Priority != data.Epics[j].Priority {
			return data.Epics[i].Priority < data.Epics[j].Priority
		}
		return data.Epics[i].Pct > data.Epics[j].Pct
	})

	data.TotalEpics = len(data.Epics)
	for _, ep := range data.Epics {
		switch {
		case ep.Total > 0 && ep.Closed == ep.Total:
			data.Completed++
		case ep.Closed > 0 || ep.InProgress > 0:
			data.InFlight++
		default:
			data.NotStarted++
		}
	}

	s.render(w, r, "phase", data)
}

