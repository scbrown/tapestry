package web

import (
	"log"
	"net/http"
	"sort"
	"sync"

	"github.com/scbrown/tapestry/internal/dolt"
)

type kanbanCard struct {
	ID       string
	Title    string
	Priority int
	Assignee string
	Rig      string
	Type     string
}

type kanbanColumn struct {
	Status string
	Cards  []kanbanCard
}

type kanbanData struct {
	Columns []kanbanColumn
	Total   int
	Err     string
}

var kanbanStatuses = []string{"open", "in_progress", "hooked", "blocked"}

func (s *Server) handleKanban(w http.ResponseWriter, r *http.Request) {
	if s.ds == nil {
		s.render(w, r, "kanban", kanbanData{})
		return
	}

	ctx := r.Context()
	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("kanban: list dbs: %v", err)
		s.render(w, r, "kanban", kanbanData{Err: err.Error()})
		return
	}

	type dbResult struct {
		issues []dolt.Issue
	}

	results := make([]dbResult, len(dbs))
	var wg sync.WaitGroup
	for i, db := range dbs {
		wg.Add(1)
		go func(i int, dbName string) {
			defer wg.Done()
			issues, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{Limit: 500})
			if err != nil {
				log.Printf("kanban %s: %v", dbName, err)
				return
			}
			for j := range issues {
				issues[j].Rig = dbName
			}
			results[i] = dbResult{issues: issues}
		}(i, db.Name)
	}
	wg.Wait()

	// Build columns by status
	byStatus := map[string][]kanbanCard{}
	total := 0
	for _, r := range results {
		for _, iss := range r.issues {
			found := false
			for _, s := range kanbanStatuses {
				if iss.Status == s {
					found = true
					break
				}
			}
			if !found {
				continue
			}
			card := kanbanCard{
				ID:       iss.ID,
				Title:    iss.Title,
				Priority: iss.Priority,
				Assignee: iss.Assignee,
				Rig:      iss.Rig,
				Type:     iss.Type,
			}
			byStatus[iss.Status] = append(byStatus[iss.Status], card)
			total++
		}
	}

	// Sort cards within each column by priority (ascending)
	for status := range byStatus {
		cards := byStatus[status]
		sort.Slice(cards, func(i, j int) bool {
			return cards[i].Priority < cards[j].Priority
		})
		byStatus[status] = cards
	}

	var columns []kanbanColumn
	for _, status := range kanbanStatuses {
		columns = append(columns, kanbanColumn{
			Status: status,
			Cards:  byStatus[status],
		})
	}

	s.render(w, r, "kanban", kanbanData{
		Columns: columns,
		Total:   total,
	})
}
