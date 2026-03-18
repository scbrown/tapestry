package web

import (
	"log"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/scbrown/tapestry/internal/dolt"
)

type quickWinItem struct {
	Rig        string
	Issue      dolt.Issue
	Score      int // lower = easier to knock out
	AgeDays    int
	DepCount   int
	CommentCnt int
}

type quickWinsData struct {
	GeneratedAt time.Time
	Items       []quickWinItem
	Total       int
	Rigs        []string
	FilterRig   string
	SortBy      string
	Assignees   []string
}

func (s *Server) handleQuickWins(w http.ResponseWriter, r *http.Request) {
	data := quickWinsData{GeneratedAt: time.Now()}

	if s.ds == nil {
		s.render(w, r, "quick-wins", data)
		return
	}

	ctx := r.Context()
	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("quick-wins: list dbs: %v", err)
		s.render(w, r, "quick-wins", data)
		return
	}

	filterRig := r.URL.Query().Get("rig")
	sortBy := r.URL.Query().Get("sort")
	if sortBy == "" {
		sortBy = "priority"
	}
	data.SortBy = sortBy
	data.FilterRig = filterRig
	var rigs []string
	for _, db := range dbs {
		rigs = append(rigs, db.Name)
	}
	sort.Strings(rigs)
	data.Rigs = rigs
	data.FilterRig = filterRig

	now := time.Now()

	type dbResult struct {
		rig       string
		items     []quickWinItem
		assignees []string
	}
	results := make([]dbResult, len(dbs))
	var wg sync.WaitGroup

	for i, db := range dbs {
		if filterRig != "" && db.Name != filterRig {
			continue
		}
		wg.Add(1)
		go func(idx int, dbName string) {
			defer wg.Done()
			assignees, _ := s.ds.DistinctAssignees(ctx, dbName)

			// Get open issues (tasks and bugs, not epics)
			open, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{Status: "open", Limit: 5000})
			if err != nil {
				log.Printf("quick-wins: %s: %v", dbName, err)
				return
			}

			var items []quickWinItem
			for _, iss := range open {
				if iss.Type == "epic" || iss.Type == "decision" {
					continue
				}

				ageDays := int(now.Sub(iss.CreatedAt).Hours() / 24)

				// Get dependency count
				deps, err := s.ds.Dependencies(ctx, dbName, iss.ID)
				depCount := 0
				if err == nil {
					depCount = len(deps)
				}

				// Get comment count
				comments, err := s.ds.Comments(ctx, dbName, iss.ID)
				commentCnt := 0
				if err == nil {
					commentCnt = len(comments)
				}

				// Compute complexity score: deps are heavy, comments indicate discussion
				// Lower title length = likely simpler
				titleLen := len(iss.Title)
				descLen := len(iss.Description)
				score := depCount*5 + commentCnt*2
				if titleLen > 80 {
					score += 2
				}
				if descLen > 500 {
					score += 3
				}

				// Only include low-score items (quick wins)
				if score <= 10 && depCount == 0 {
					items = append(items, quickWinItem{
						Rig:        dbName,
						Issue:      iss,
						Score:      score,
						AgeDays:    ageDays,
						DepCount:   depCount,
						CommentCnt: commentCnt,
					})
				}
			}

			results[idx] = dbResult{rig: dbName, items: items, assignees: assignees}
		}(i, db.Name)
	}
	wg.Wait()

	var allItems []quickWinItem
	assigneeSet := make(map[string]bool)
	for _, r := range results {
		allItems = append(allItems, r.items...)
		for _, a := range r.assignees {
			assigneeSet[a] = true
		}
	}
	for a := range assigneeSet {
		data.Assignees = append(data.Assignees, a)
	}
	sort.Strings(data.Assignees)

	switch sortBy {
	case "score":
		sort.Slice(allItems, func(i, j int) bool {
			return allItems[i].Score < allItems[j].Score
		})
	case "age":
		sort.Slice(allItems, func(i, j int) bool {
			return allItems[i].AgeDays > allItems[j].AgeDays
		})
	case "rig":
		sort.Slice(allItems, func(i, j int) bool {
			if allItems[i].Rig != allItems[j].Rig {
				return allItems[i].Rig < allItems[j].Rig
			}
			return allItems[i].Issue.Priority < allItems[j].Issue.Priority
		})
	default: // "priority"
		sort.Slice(allItems, func(i, j int) bool {
			if allItems[i].Issue.Priority != allItems[j].Issue.Priority {
				return allItems[i].Issue.Priority < allItems[j].Issue.Priority
			}
			return allItems[i].Score < allItems[j].Score
		})
	}

	if len(allItems) > 100 {
		allItems = allItems[:100]
	}

	data.Items = allItems
	data.Total = len(allItems)
	s.render(w, r, "quick-wins", data)
}
