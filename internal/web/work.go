package web

import (
	"context"
	"log"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/scbrown/tapestry/internal/dolt"
)

// epicTree groups an epic with its child issues and progress.
type epicTree struct {
	Epic     dolt.Issue
	Rig      string
	Children []dolt.Issue
	Progress dolt.EpicProgress
}

type repoSection struct {
	Name     string
	Expanded bool
	Stats    repoStats
	Epics    []epicTree
	Tasks    []dolt.Issue
}

type repoStats struct {
	Open       int
	InProgress int
	Closed     int
	Total      int
}

type prioritySection struct {
	Priority int
	Label    string
	Count    int
	Epics    []epicTree
	Tasks    []dolt.Issue
}

type agentSection struct {
	Name   string
	Stats  repoStats
	Issues []dolt.Issue
}

type workData struct {
	Mode       string
	ShowClosed bool
	TotalCount int
	Repos      []repoSection
	Priorities []prioritySection
	Agents     []agentSection
	Rigs       []string
	FilterRig  string
}

func (s *Server) handleWork(w http.ResponseWriter, r *http.Request) {
	mode := r.URL.Query().Get("mode")
	if mode == "" {
		mode = "repo"
	}
	showClosed := r.URL.Query().Get("closed") == "1"

	data := workData{Mode: mode, ShowClosed: showClosed}

	if s.ds == nil {
		s.render(w, r, "work", data)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 20*time.Second)
	defer cancel()

	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("work: list dbs: %v", err)
		s.render(w, r, "work", data)
		return
	}

	type dbResult struct {
		rigName string
		epics   []epicTree
		tasks   []dolt.Issue
	}

	results := make([]dbResult, len(dbs))
	var wg sync.WaitGroup
	for i, db := range dbs {
		wg.Add(1)
		go func(i int, dbName string) {
			defer wg.Done()
			r := dbResult{rigName: dbName}

			allIssues, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{Limit: 500})
			if err != nil {
				log.Printf("work: issues %s: %v", dbName, err)
				results[i] = r
				return
			}

			childDeps, err := s.ds.AllChildDependencies(ctx, dbName)
			if err != nil {
				log.Printf("work: child deps %s: %v", dbName, err)
			}

			// Build lookup maps
			issueMap := make(map[string]dolt.Issue, len(allIssues))
			for _, iss := range allIssues {
				iss.Rig = dbName
				issueMap[iss.ID] = iss
			}

			// Map child -> parent
			childToParent := make(map[string]string)
			parentChildren := make(map[string][]string)
			for _, dep := range childDeps {
				childToParent[dep.FromID] = dep.ToID
				parentChildren[dep.ToID] = append(parentChildren[dep.ToID], dep.FromID)
			}

			// Build epic trees
			for _, iss := range allIssues {
				if iss.Type != "epic" {
					continue
				}
				if !showClosed && iss.Status == "closed" {
					continue
				}
				if isNoise(iss.ID, iss.Title) {
					continue
				}

				et := epicTree{Epic: iss, Rig: dbName}
				for _, childID := range parentChildren[iss.ID] {
					if child, ok := issueMap[childID]; ok {
						if !showClosed && child.Status == "closed" {
							et.Progress.Total++
							et.Progress.Closed++
							continue
						}
						child.Rig = dbName
						et.Children = append(et.Children, child)
						et.Progress.Total++
						if child.Status == "closed" {
							et.Progress.Closed++
						}
					}
				}

				sort.Slice(et.Children, func(a, b int) bool {
					if et.Children[a].Priority != et.Children[b].Priority {
						return et.Children[a].Priority < et.Children[b].Priority
					}
					return et.Children[a].UpdatedAt.After(et.Children[b].UpdatedAt)
				})

				r.epics = append(r.epics, et)
			}

			// Standalone tasks (not epics, not children of epics)
			for _, iss := range allIssues {
				if iss.Type == "epic" {
					continue
				}
				if _, isChild := childToParent[iss.ID]; isChild {
					continue
				}
				if !showClosed && iss.Status == "closed" {
					continue
				}
				if isNoise(iss.ID, iss.Title) {
					continue
				}
				iss.Rig = dbName
				r.tasks = append(r.tasks, iss)
			}

			sort.Slice(r.epics, func(a, b int) bool {
				if r.epics[a].Epic.Priority != r.epics[b].Epic.Priority {
					return r.epics[a].Epic.Priority < r.epics[b].Epic.Priority
				}
				return r.epics[a].Epic.UpdatedAt.After(r.epics[b].Epic.UpdatedAt)
			})
			sort.Slice(r.tasks, func(a, b int) bool {
				if r.tasks[a].Priority != r.tasks[b].Priority {
					return r.tasks[a].Priority < r.tasks[b].Priority
				}
				return r.tasks[a].UpdatedAt.After(r.tasks[b].UpdatedAt)
			})

			results[i] = r
		}(i, db.Name)
	}
	wg.Wait()

	// Collect rigs and apply rig filter
	rigSet := make(map[string]bool)
	for _, r := range results {
		if len(r.epics) > 0 || len(r.tasks) > 0 {
			rigSet[r.rigName] = true
		}
	}
	for rig := range rigSet {
		data.Rigs = append(data.Rigs, rig)
	}
	sort.Strings(data.Rigs)

	data.FilterRig = r.URL.Query().Get("rig")
	if data.FilterRig != "" {
		filtered := results[:0]
		for _, res := range results {
			if res.rigName == data.FilterRig {
				filtered = append(filtered, res)
			}
		}
		results = filtered
	}

	if mode == "repo" {
		for _, r := range results {
			if len(r.epics) == 0 && len(r.tasks) == 0 {
				continue
			}
			sec := repoSection{
				Name:  strings.TrimPrefix(r.rigName, "beads_"),
				Epics: r.epics,
				Tasks: r.tasks,
			}

			for _, et := range r.epics {
				countIssueStats(&sec.Stats, et.Epic)
				for _, c := range et.Children {
					countIssueStats(&sec.Stats, c)
				}
			}
			for _, t := range r.tasks {
				countIssueStats(&sec.Stats, t)
			}
			sec.Stats.Total = sec.Stats.Open + sec.Stats.InProgress + sec.Stats.Closed
			sec.Expanded = sec.Stats.InProgress > 0

			data.Repos = append(data.Repos, sec)
		}
		sort.Slice(data.Repos, func(i, j int) bool {
			return data.Repos[i].Stats.InProgress > data.Repos[j].Stats.InProgress
		})
	} else if mode == "agent" {
		// Agent mode — group by assignee
		aMap := make(map[string]*agentSection)
		for _, r := range results {
			for _, et := range r.epics {
				agent := shortActorName(et.Epic.Assignee)
				if agent == "" {
					agent = shortActorName(et.Epic.Owner)
				}
				if agent == "" {
					agent = "(unassigned)"
				}
				as := getOrCreateAgent(aMap, agent)
				countIssueStats(&as.Stats, et.Epic)
				for _, c := range et.Children {
					a2 := shortActorName(c.Assignee)
					if a2 == "" {
						a2 = agent
					}
					as2 := getOrCreateAgent(aMap, a2)
					as2.Issues = append(as2.Issues, c)
					countIssueStats(&as2.Stats, c)
				}
				as.Issues = append(as.Issues, et.Epic)
			}
			for _, t := range r.tasks {
				agent := shortActorName(t.Assignee)
				if agent == "" {
					agent = shortActorName(t.Owner)
				}
				if agent == "" {
					agent = "(unassigned)"
				}
				as := getOrCreateAgent(aMap, agent)
				as.Issues = append(as.Issues, t)
				countIssueStats(&as.Stats, t)
			}
		}
		for _, as := range aMap {
			as.Stats.Total = as.Stats.Open + as.Stats.InProgress + as.Stats.Closed
			sort.Slice(as.Issues, func(i, j int) bool {
				if as.Issues[i].Priority != as.Issues[j].Priority {
					return as.Issues[i].Priority < as.Issues[j].Priority
				}
				return as.Issues[i].UpdatedAt.After(as.Issues[j].UpdatedAt)
			})
			data.Agents = append(data.Agents, *as)
		}
		sort.Slice(data.Agents, func(i, j int) bool {
			return data.Agents[i].Stats.InProgress > data.Agents[j].Stats.InProgress
		})
	} else {
		// Priority mode
		pMap := make(map[int]*prioritySection)
		for _, r := range results {
			for _, et := range r.epics {
				p := et.Epic.Priority
				ps := getOrCreatePriority(pMap, p)
				ps.Epics = append(ps.Epics, et)
				ps.Count += 1 + len(et.Children)
			}
			for _, t := range r.tasks {
				p := t.Priority
				ps := getOrCreatePriority(pMap, p)
				ps.Tasks = append(ps.Tasks, t)
				ps.Count++
			}
		}
		for _, ps := range pMap {
			data.Priorities = append(data.Priorities, *ps)
		}
		sort.Slice(data.Priorities, func(i, j int) bool {
			return data.Priorities[i].Priority < data.Priorities[j].Priority
		})
	}

	for _, r := range data.Repos {
		data.TotalCount += r.Stats.Total
	}
	for _, p := range data.Priorities {
		data.TotalCount += p.Count
	}
	for _, a := range data.Agents {
		data.TotalCount += a.Stats.Total
	}

	s.render(w, r, "work", data)
}

// ── Epics Page ────────────────────────────────────────────

type epicsData struct {
	Epics     []epicTree
	Rigs      []string
	FilterRig string
	Assignees []string
	SortBy    string
}

func (s *Server) handleEpics(w http.ResponseWriter, r *http.Request) {
	data := epicsData{}

	if s.ds == nil {
		s.render(w, r, "epics", data)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("epics: list dbs: %v", err)
		s.render(w, r, "epics", data)
		return
	}

	type epicsDBResult struct {
		epics     []epicTree
		assignees []string
	}
	results := make([]epicsDBResult, len(dbs))
	var wg sync.WaitGroup
	for i, db := range dbs {
		wg.Add(1)
		go func(i int, dbName string) {
			defer wg.Done()
			var r epicsDBResult
			r.assignees, _ = s.ds.DistinctAssignees(ctx, dbName)

			epics, err := s.ds.Epics(ctx, dbName)
			if err != nil {
				log.Printf("epics: %s: %v", dbName, err)
				results[i] = r
				return
			}

			childDeps, _ := s.ds.AllChildDependencies(ctx, dbName)
			parentChildren := make(map[string][]string)
			for _, dep := range childDeps {
				parentChildren[dep.ToID] = append(parentChildren[dep.ToID], dep.FromID)
			}

			issueMap := make(map[string]dolt.Issue)
			allIssues, _ := s.ds.Issues(ctx, dbName, dolt.IssueFilter{Limit: 500})
			for _, iss := range allIssues {
				issueMap[iss.ID] = iss
			}

			// Also include closed issues for accurate progress calculation
			closedIssues, _ := s.ds.Issues(ctx, dbName, dolt.IssueFilter{Status: "closed", Limit: 500})
			for _, iss := range closedIssues {
				issueMap[iss.ID] = iss
			}

			for _, epic := range epics {
				if isNoise(epic.ID, epic.Title) {
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
				r.epics = append(r.epics, et)
			}
			results[i] = r
		}(i, db.Name)
	}
	wg.Wait()

	rigSet := make(map[string]bool)
	assigneeSet := make(map[string]bool)
	for _, r := range results {
		data.Epics = append(data.Epics, r.epics...)
		for _, e := range r.epics {
			rigSet[e.Rig] = true
		}
		for _, a := range r.assignees {
			assigneeSet[a] = true
		}
	}

	for rig := range rigSet {
		data.Rigs = append(data.Rigs, rig)
	}
	sort.Strings(data.Rigs)
	for a := range assigneeSet {
		data.Assignees = append(data.Assignees, a)
	}
	sort.Strings(data.Assignees)

	data.FilterRig = r.URL.Query().Get("rig")
	if data.FilterRig != "" {
		filtered := data.Epics[:0]
		for _, e := range data.Epics {
			if e.Rig == data.FilterRig {
				filtered = append(filtered, e)
			}
		}
		data.Epics = filtered
	}

	sortBy := r.URL.Query().Get("sort")
	if sortBy == "" {
		sortBy = "priority"
	}
	data.SortBy = sortBy

	switch sortBy {
	case "updated":
		sort.Slice(data.Epics, func(i, j int) bool {
			return data.Epics[i].Epic.UpdatedAt.After(data.Epics[j].Epic.UpdatedAt)
		})
	case "created":
		sort.Slice(data.Epics, func(i, j int) bool {
			return data.Epics[i].Epic.CreatedAt.After(data.Epics[j].Epic.CreatedAt)
		})
	case "status":
		sort.Slice(data.Epics, func(i, j int) bool {
			if data.Epics[i].Epic.Status != data.Epics[j].Epic.Status {
				return data.Epics[i].Epic.Status < data.Epics[j].Epic.Status
			}
			return data.Epics[i].Epic.Priority < data.Epics[j].Epic.Priority
		})
	case "progress":
		sort.Slice(data.Epics, func(i, j int) bool {
			pi, pj := 0.0, 0.0
			if data.Epics[i].Progress.Total > 0 {
				pi = float64(data.Epics[i].Progress.Closed) / float64(data.Epics[i].Progress.Total)
			}
			if data.Epics[j].Progress.Total > 0 {
				pj = float64(data.Epics[j].Progress.Closed) / float64(data.Epics[j].Progress.Total)
			}
			return pi > pj
		})
	default: // priority
		sort.Slice(data.Epics, func(i, j int) bool {
			if data.Epics[i].Epic.Priority != data.Epics[j].Epic.Priority {
				return data.Epics[i].Epic.Priority < data.Epics[j].Epic.Priority
			}
			return data.Epics[i].Epic.UpdatedAt.After(data.Epics[j].Epic.UpdatedAt)
		})
	}

	s.render(w, r, "epics", data)
}

func countIssueStats(s *repoStats, iss dolt.Issue) {
	switch iss.Status {
	case "open":
		s.Open++
	case "in_progress", "hooked":
		s.InProgress++
	case "closed":
		s.Closed++
	}
}

func getOrCreateAgent(m map[string]*agentSection, name string) *agentSection {
	if as, ok := m[name]; ok {
		return as
	}
	as := &agentSection{Name: name}
	m[name] = as
	return as
}

func getOrCreatePriority(m map[int]*prioritySection, p int) *prioritySection {
	if ps, ok := m[p]; ok {
		return ps
	}
	label := "P" + strings.TrimLeft(string(rune('0'+p)), "0")
	if p == 0 {
		label = "Unset"
	}
	ps := &prioritySection{Priority: p, Label: label}
	m[p] = ps
	return ps
}
