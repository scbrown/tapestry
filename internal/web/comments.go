package web

import (
	"log"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/scbrown/tapestry/internal/dolt"
)

type commentEntry struct {
	Rig     string
	IssueID string
	Author  string
	Body    string
	Created time.Time
}

type commentsData struct {
	GeneratedAt time.Time
	Comments    []commentEntry
	Total       int
}

func (s *Server) handleComments(w http.ResponseWriter, r *http.Request) {
	data := commentsData{GeneratedAt: time.Now()}

	if s.ds == nil {
		s.render(w, r, "comments", data)
		return
	}

	ctx := r.Context()
	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("comments: list dbs: %v", err)
		s.render(w, r, "comments", data)
		return
	}

	type dbResult struct {
		rig      string
		comments []dolt.Comment
	}
	results := make([]dbResult, len(dbs))
	var wg sync.WaitGroup
	for i, db := range dbs {
		wg.Add(1)
		go func(i int, dbName string) {
			defer wg.Done()
			comments, err := s.ds.RecentComments(ctx, dbName, 50)
			if err != nil {
				log.Printf("comments: %s: %v", dbName, err)
				return
			}
			results[i] = dbResult{rig: dbName, comments: comments}
		}(i, db.Name)
	}
	wg.Wait()

	// Merge and sort by created descending
	for _, r := range results {
		for _, c := range r.comments {
			data.Comments = append(data.Comments, commentEntry{
				Rig:     r.rig,
				IssueID: c.IssueID,
				Author:  c.Author,
				Body:    c.Body,
				Created: c.CreatedAt,
			})
		}
	}

	sort.Slice(data.Comments, func(i, j int) bool {
		return data.Comments[i].Created.After(data.Comments[j].Created)
	})

	// Cap at 100
	if len(data.Comments) > 100 {
		data.Comments = data.Comments[:100]
	}

	data.Total = len(data.Comments)
	s.render(w, r, "comments", data)
}
