package web

import (
	"log"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/scbrown/tapestry/internal/dolt"
)

type dupGroup struct {
	Key    string
	Count  int
	Issues []dolt.Issue
}

type duplicatesData struct {
	GeneratedAt time.Time
	Groups      []dupGroup
	TotalDups   int
	TotalGroups int
}

// normalizeTitle strips common prefixes like "[AUTO] " and the alert name
// to find duplicates with the same root alert.
func normalizeTitle(title string) string {
	t := title
	// Strip [AUTO] prefix
	t = strings.TrimPrefix(t, "[AUTO] ")
	// For alert-style titles like "AlertName: details", use just the alert name
	if idx := strings.Index(t, ":"); idx > 0 && idx < 60 {
		t = strings.TrimSpace(t[:idx])
	}
	return strings.ToLower(t)
}

func (s *Server) handleDuplicates(w http.ResponseWriter, r *http.Request) {
	data := duplicatesData{GeneratedAt: time.Now()}

	if s.ds == nil {
		s.render(w, r, "duplicates", data)
		return
	}

	ctx := r.Context()
	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("duplicates: list dbs: %v", err)
		s.render(w, r, "duplicates", data)
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
			issues, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{Limit: 3000})
			if err != nil {
				log.Printf("duplicates: %s: %v", dbName, err)
				return
			}
			results[i] = dbResult{issues: issues}
		}(i, db.Name)
	}
	wg.Wait()

	// Group open/in_progress issues by normalized title
	groups := map[string][]dolt.Issue{}
	for i, r := range results {
		for _, iss := range r.issues {
			if iss.Status == "closed" || iss.Status == "deferred" {
				continue
			}
			iss.Rig = dbs[i].Name
			key := normalizeTitle(iss.Title)
			groups[key] = append(groups[key], iss)
		}
	}

	// Only keep groups with 2+ issues (actual duplicates)
	for key, issues := range groups {
		if len(issues) < 2 {
			continue
		}
		// Sort by created date (newest first)
		sort.Slice(issues, func(i, j int) bool {
			return issues[i].CreatedAt.After(issues[j].CreatedAt)
		})
		// Limit to 10 per group for display
		shown := issues
		if len(shown) > 10 {
			shown = shown[:10]
		}
		data.Groups = append(data.Groups, dupGroup{
			Key:    key,
			Count:  len(issues),
			Issues: shown,
		})
		data.TotalDups += len(issues)
	}

	// Sort groups by count descending
	sort.Slice(data.Groups, func(i, j int) bool {
		return data.Groups[i].Count > data.Groups[j].Count
	})

	data.TotalGroups = len(data.Groups)

	// Limit to top 30 groups
	if len(data.Groups) > 30 {
		data.Groups = data.Groups[:30]
	}

	s.render(w, r, "duplicates", data)
}
