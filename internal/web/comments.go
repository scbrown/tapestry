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
	GeneratedAt  time.Time
	Comments     []commentEntry
	Total        int
	Authors      []string // distinct authors for filter
	FilterAuthor string   // currently selected author filter
	Rigs         []string
	FilterRig    string
	SortBy       string
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
	rigSet := make(map[string]bool)
	for _, db := range dbs {
		rigSet[db.Name] = true
	}
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

	for rig := range rigSet {
		data.Rigs = append(data.Rigs, rig)
	}
	sort.Strings(data.Rigs)

	// Apply rig filter if specified
	data.FilterRig = r.URL.Query().Get("rig")
	if data.FilterRig != "" {
		filtered := data.Comments[:0]
		for _, c := range data.Comments {
			if c.Rig == data.FilterRig {
				filtered = append(filtered, c)
			}
		}
		data.Comments = filtered
	}

	sortBy := r.URL.Query().Get("sort")
	if sortBy == "" {
		sortBy = "date"
	}
	data.SortBy = sortBy

	switch sortBy {
	case "author":
		sort.Slice(data.Comments, func(i, j int) bool {
			if data.Comments[i].Author != data.Comments[j].Author {
				return data.Comments[i].Author < data.Comments[j].Author
			}
			return data.Comments[i].Created.After(data.Comments[j].Created)
		})
	case "rig":
		sort.Slice(data.Comments, func(i, j int) bool {
			if data.Comments[i].Rig != data.Comments[j].Rig {
				return data.Comments[i].Rig < data.Comments[j].Rig
			}
			return data.Comments[i].Created.After(data.Comments[j].Created)
		})
	default: // "date"
		sort.Slice(data.Comments, func(i, j int) bool {
			return data.Comments[i].Created.After(data.Comments[j].Created)
		})
	}

	// Cap at 100
	if len(data.Comments) > 100 {
		data.Comments = data.Comments[:100]
	}

	// Collect distinct authors for filter
	authorSet := make(map[string]bool)
	for _, c := range data.Comments {
		if c.Author != "" {
			authorSet[c.Author] = true
		}
	}
	for a := range authorSet {
		data.Authors = append(data.Authors, a)
	}
	sort.Strings(data.Authors)

	// Apply author filter if specified
	filterAuthor := r.URL.Query().Get("author")
	if filterAuthor != "" {
		data.FilterAuthor = filterAuthor
		filtered := data.Comments[:0]
		for _, c := range data.Comments {
			if c.Author == filterAuthor {
				filtered = append(filtered, c)
			}
		}
		data.Comments = filtered
	}

	data.Total = len(data.Comments)
	s.render(w, r, "comments", data)
}
