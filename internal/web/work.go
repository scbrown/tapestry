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

type workData struct {
	Mode       string
	ShowClosed bool
	TotalCount int
	Repos      []repoSection
	Priorities []prioritySection
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

	s.render(w, r, "work", data)
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
