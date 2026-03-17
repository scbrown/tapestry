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

type chainInfo struct {
	Length int
	Path   []chainNode
}

type chainNode struct {
	ID     string
	Title  string
	Status string
	DB     string
}

type chainsData struct {
	GeneratedAt time.Time
	FilterRig   string
	Rigs        []string

	// Longest chains
	LongestChains []chainInfo

	// Stats
	TotalDeps    int
	MaxDepth     int
	AvgDepth     float64
	ChainsOver3  int // chains with depth > 3
	CircularRefs int // cycles detected

	// Beads with most dependents (most blocking)
	TopBlockers []blockerInfo

	Err string
}

type blockerInfo struct {
	ID         string
	Title      string
	Status     string
	DB         string
	Dependents int
}

func (s *Server) handleChains(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	filterRig := r.URL.Query().Get("rig")
	data := chainsData{GeneratedAt: now, FilterRig: filterRig}

	if s.ds == nil {
		s.render(w, r, "chains", data)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("chains: list dbs: %v", err)
		data.Err = err.Error()
		s.render(w, r, "chains", data)
		return
	}

	for _, db := range dbs {
		data.Rigs = append(data.Rigs, db.Name)
	}
	sort.Strings(data.Rigs)

	// Collect all dependency edges and issues across databases
	type edge struct {
		from string // issue that depends (child)
		to   string // issue depended on (parent)
	}

	var allEdges []edge
	issueMap := make(map[string]dolt.Issue)
	issueDB := make(map[string]string)
	dependentCount := make(map[string]int) // id -> number of issues depending on it

	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, db := range dbs {
		if filterRig != "" && db.Name != filterRig {
			continue
		}
		wg.Add(1)
		go func(dbName string) {
			defer wg.Done()

			deps, err := s.ds.AllDependenciesWithIssues(ctx, dbName)
			if err != nil {
				log.Printf("chains %s: deps: %v", dbName, err)
				return
			}

			issues, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{Limit: 2000})
			if err != nil {
				log.Printf("chains %s: issues: %v", dbName, err)
				return
			}

			mu.Lock()
			for _, d := range deps {
				allEdges = append(allEdges, edge{from: d.From.ID, to: d.To.ID})
				dependentCount[d.To.ID]++
			}
			for _, iss := range issues {
				if isNoise(iss.ID, iss.Title) {
					continue
				}
				issueMap[iss.ID] = iss
				issueDB[iss.ID] = dbName
			}
			mu.Unlock()
		}(db.Name)
	}
	wg.Wait()

	data.TotalDeps = len(allEdges)

	// Build adjacency list: for each issue, what does it depend on?
	dependsOn := make(map[string][]string)
	for _, e := range allEdges {
		dependsOn[e.from] = append(dependsOn[e.from], e.to)
	}

	// Find all chain roots (issues that have dependencies but nothing depends on them)
	hasDependents := make(map[string]bool)
	for _, e := range allEdges {
		hasDependents[e.to] = true
	}

	// Compute chain depth from each node via DFS
	depthCache := make(map[string]int)
	visiting := make(map[string]bool)
	var circularCount int

	var computeDepth func(id string) int
	computeDepth = func(id string) int {
		if d, ok := depthCache[id]; ok {
			return d
		}
		if visiting[id] {
			circularCount++
			return 0
		}
		visiting[id] = true
		maxChild := 0
		for _, dep := range dependsOn[id] {
			cd := computeDepth(dep)
			if cd > maxChild {
				maxChild = cd
			}
		}
		delete(visiting, id)
		depth := maxChild + 1
		depthCache[id] = depth
		return depth
	}

	// Compute depth for all nodes that have dependencies
	for id := range dependsOn {
		computeDepth(id)
	}

	data.CircularRefs = circularCount

	// Find max depth and average
	var totalDepth int
	var depthCount int
	for _, d := range depthCache {
		if d > data.MaxDepth {
			data.MaxDepth = d
		}
		totalDepth += d
		depthCount++
		if d > 3 {
			data.ChainsOver3++
		}
	}
	if depthCount > 0 {
		data.AvgDepth = float64(totalDepth) / float64(depthCount)
	}

	// Build longest chains by tracing from deepest nodes
	type depthID struct {
		id    string
		depth int
	}
	var allDepths []depthID
	for id, d := range depthCache {
		allDepths = append(allDepths, depthID{id, d})
	}
	sort.Slice(allDepths, func(i, j int) bool {
		return allDepths[i].depth > allDepths[j].depth
	})

	// Trace top 10 chains
	seen := make(map[string]bool)
	for _, di := range allDepths {
		if len(data.LongestChains) >= 10 {
			break
		}
		if seen[di.id] {
			continue
		}

		var path []chainNode
		current := di.id
		visited := make(map[string]bool)
		for current != "" && !visited[current] {
			visited[current] = true
			seen[current] = true
			iss, ok := issueMap[current]
			node := chainNode{ID: current, DB: issueDB[current]}
			if ok {
				node.Title = iss.Title
				node.Status = iss.Status
			}
			path = append(path, node)

			// Follow deepest dependency
			deps := dependsOn[current]
			if len(deps) == 0 {
				break
			}
			best := ""
			bestD := -1
			for _, d := range deps {
				if depthCache[d] > bestD {
					bestD = depthCache[d]
					best = d
				}
			}
			current = best
		}

		if len(path) > 1 {
			data.LongestChains = append(data.LongestChains, chainInfo{
				Length: len(path),
				Path:   path,
			})
		}
	}

	// Top blockers (issues with most dependents)
	type blockCount struct {
		id    string
		count int
	}
	var blockerList []blockCount
	for id, count := range dependentCount {
		blockerList = append(blockerList, blockCount{id, count})
	}
	sort.Slice(blockerList, func(i, j int) bool {
		return blockerList[i].count > blockerList[j].count
	})
	for i := 0; i < len(blockerList) && i < 15; i++ {
		bc := blockerList[i]
		iss := issueMap[bc.id]
		data.TopBlockers = append(data.TopBlockers, blockerInfo{
			ID:         bc.id,
			Title:      iss.Title,
			Status:     iss.Status,
			DB:         issueDB[bc.id],
			Dependents: bc.count,
		})
	}

	s.render(w, r, "chains", data)
}
