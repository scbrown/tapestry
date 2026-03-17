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

type complexityItem struct {
	ID          string
	DB          string
	Title       string
	Status      string
	Priority    int
	Assignee    string
	DepCount    int
	CommentCount int
	DescLen     int
	Score       int // composite complexity score
}

type complexityData struct {
	GeneratedAt time.Time

	Items      []complexityItem
	TotalBeads int

	// Distribution
	HighComplexity int // score >= 10
	MedComplexity  int // score 5-9
	LowComplexity  int // score < 5

	Rigs      []string
	FilterRig string
	Err       string
}

func complexityScore(deps, comments, descLen int) int {
	score := deps * 3 // each dependency is worth 3 points
	score += comments * 2 // each comment is worth 2 points
	if descLen > 500 {
		score += 3
	} else if descLen > 200 {
		score += 2
	} else if descLen > 50 {
		score += 1
	}
	return score
}

func (s *Server) handleComplexity(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	filterRig := r.URL.Query().Get("rig")
	data := complexityData{GeneratedAt: now, FilterRig: filterRig}

	if s.ds == nil {
		s.render(w, r, "complexity", data)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("complexity: list dbs: %v", err)
		data.Err = err.Error()
		s.render(w, r, "complexity", data)
		return
	}

	for _, db := range dbs {
		data.Rigs = append(data.Rigs, db.Name)
	}
	sort.Strings(data.Rigs)

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
				log.Printf("complexity %s: %v", dbName, err)
				return
			}

			// Get all dependencies for this DB
			allDeps, err := s.ds.AllDependenciesWithIssues(ctx, dbName)
			if err != nil {
				log.Printf("complexity %s deps: %v", dbName, err)
				allDeps = nil
			}

			// Count deps per issue
			depCounts := make(map[string]int)
			for _, d := range allDeps {
				depCounts[d.From.ID]++
				depCounts[d.To.ID]++
			}

			for _, iss := range issues {
				if isNoise(iss.ID, iss.Title) {
					continue
				}
				if iss.Status == "closed" || iss.Status == "deferred" {
					continue
				}

				comments, _ := s.ds.Comments(ctx, dbName, iss.ID)

				item := complexityItem{
					ID:           iss.ID,
					DB:           dbName,
					Title:        iss.Title,
					Status:       iss.Status,
					Priority:     iss.Priority,
					Assignee:     iss.Assignee,
					DepCount:     depCounts[iss.ID],
					CommentCount: len(comments),
					DescLen:      len(iss.Description),
				}
				item.Score = complexityScore(item.DepCount, item.CommentCount, item.DescLen)

				mu.Lock()
				data.Items = append(data.Items, item)
				data.TotalBeads++
				if item.Score >= 10 {
					data.HighComplexity++
				} else if item.Score >= 5 {
					data.MedComplexity++
				} else {
					data.LowComplexity++
				}
				mu.Unlock()
			}
		}(db.Name)
	}
	wg.Wait()

	sort.Slice(data.Items, func(i, j int) bool {
		return data.Items[i].Score > data.Items[j].Score
	})

	if len(data.Items) > 50 {
		data.Items = data.Items[:50]
	}

	s.render(w, r, "complexity", data)
}
