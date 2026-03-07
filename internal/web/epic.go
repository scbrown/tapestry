package web

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/scbrown/tapestry/internal/dolt"
)

type epicDetailData struct {
	Issue    dolt.Issue
	RigName  string
	Children []dolt.Issue
	Progress dolt.EpicProgress
	Comments []dolt.Comment
	Err      string
}

func (s *Server) handleEpicDetail(w http.ResponseWriter, r *http.Request, id string) {
	data := epicDetailData{}

	if s.ds == nil {
		data.Err = "No database connection configured"
		s.render(w, r, "epic", data)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	dbs, err := s.databases(ctx)
	if err != nil {
		data.Err = fmt.Sprintf("Failed to list databases: %v", err)
		s.render(w, r, "epic", data)
		return
	}

	// Find the epic across all databases
	var foundDB string
	for _, db := range dbs {
		issue, err := s.ds.IssueByID(ctx, db.Name, id)
		if err != nil {
			continue
		}
		if issue != nil {
			data.Issue = *issue
			foundDB = db.Name
			data.RigName = strings.TrimPrefix(db.Name, "beads_")
			break
		}
	}

	if foundDB == "" {
		http.NotFound(w, r)
		return
	}

	// Get child dependencies
	childDeps, err := s.ds.AllChildDependencies(ctx, foundDB)
	if err != nil {
		log.Printf("epic: child deps %s: %v", foundDB, err)
	}

	var childIDs []string
	for _, dep := range childDeps {
		if dep.ToID == id {
			childIDs = append(childIDs, dep.FromID)
		}
	}

	for _, childID := range childIDs {
		child, err := s.ds.IssueByID(ctx, foundDB, childID)
		if err != nil || child == nil {
			continue
		}
		data.Children = append(data.Children, *child)
		data.Progress.Total++
		if child.Status == "closed" {
			data.Progress.Closed++
		}
	}

	sort.Slice(data.Children, func(i, j int) bool {
		statusOrder := map[string]int{"in_progress": 0, "hooked": 0, "open": 1, "blocked": 2, "closed": 3}
		if statusOrder[data.Children[i].Status] != statusOrder[data.Children[j].Status] {
			return statusOrder[data.Children[i].Status] < statusOrder[data.Children[j].Status]
		}
		return data.Children[i].Priority < data.Children[j].Priority
	})

	comments, err := s.ds.Comments(ctx, foundDB, id)
	if err == nil {
		data.Comments = comments
	}

	s.render(w, r, "epic", data)
}
