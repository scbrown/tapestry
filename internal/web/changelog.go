package web

import (
	"log"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/scbrown/tapestry/internal/dolt"
)

type changelogWeek struct {
	Start   time.Time
	End     time.Time
	Label   string // "Mar 10 – Mar 16"
	Items   []changelogItem
	Count   int
	ByType  map[string]int // bug, task, epic
}

type changelogItem struct {
	Issue dolt.Issue
	Rig   string
}

type changelogData struct {
	GeneratedAt    time.Time
	Weeks          []changelogWeek
	WeekCount      int
	TotalClosed    int
	FilterRig      string
	FilterType     string
	FilterPriority string
	Rigs           []string
	Types          []string
	QNoRig         string // query string without rig param
	QNoType        string // query string without type param
	QNoPriority    string // query string without priority param
	QAll           string // query string with all params
	Err            string
}

func (s *Server) handleChangelog(w http.ResponseWriter, r *http.Request) {
	data := changelogData{GeneratedAt: time.Now(), WeekCount: 8}

	if s.ds == nil {
		s.render(w, r, "changelog", data)
		return
	}

	ctx := r.Context()
	data.FilterRig = r.URL.Query().Get("rig")
	data.FilterType = r.URL.Query().Get("type")
	data.FilterPriority = r.URL.Query().Get("priority")

	dbs, err := s.databases(ctx)
	if err != nil {
		data.Err = err.Error()
		s.render(w, r, "changelog", data)
		return
	}

	// Collect all closed issues from last 8 weeks
	now := time.Now()
	eightWeeksAgo := now.AddDate(0, 0, -56)

	type issueResult struct {
		issue dolt.Issue
		rig   string
	}

	var allIssues []issueResult
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, db := range dbs {
		if data.FilterRig != "" && db.Name != data.FilterRig {
			continue
		}
		wg.Add(1)
		go func(dbName string) {
			defer wg.Done()
			issues, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{
				Status: "closed",
				Limit:  5000,
			})
			if err != nil {
				log.Printf("changelog: issues %s: %v", dbName, err)
				return
			}

			var filtered []issueResult
			for _, iss := range issues {
				if iss.UpdatedAt.After(eightWeeksAgo) {
					filtered = append(filtered, issueResult{issue: iss, rig: dbName})
				}
			}

			mu.Lock()
			allIssues = append(allIssues, filtered...)
			mu.Unlock()
		}(db.Name)
	}
	wg.Wait()

	// Build rig list
	var rigNames []string
	for _, db := range dbs {
		rigNames = append(rigNames, db.Name)
	}
	sort.Strings(rigNames)
	data.Rigs = rigNames

	// Collect distinct types
	typeSet := map[string]bool{}
	for _, ir := range allIssues {
		if ir.issue.Type != "" {
			typeSet[ir.issue.Type] = true
		}
	}
	var types []string
	for t := range typeSet {
		types = append(types, t)
	}
	sort.Strings(types)
	data.Types = types

	// Apply type filter
	if data.FilterType != "" {
		filtered := allIssues[:0]
		for _, ir := range allIssues {
			if ir.issue.Type == data.FilterType {
				filtered = append(filtered, ir)
			}
		}
		allIssues = filtered
	}

	// Apply priority filter
	if data.FilterPriority != "" {
		if pri, err := strconv.Atoi(data.FilterPriority); err == nil {
			filtered := allIssues[:0]
			for _, ir := range allIssues {
				if ir.issue.Priority == pri {
					filtered = append(filtered, ir)
				}
			}
			allIssues = filtered
		}
	}

	// Group by week (Monday start)
	weekMap := map[string]*changelogWeek{}
	for _, ir := range allIssues {
		// Find the Monday of the week this issue was updated
		weekStart := weekStartDate(ir.issue.UpdatedAt)
		key := weekStart.Format("2006-01-02")

		w, ok := weekMap[key]
		if !ok {
			weekEnd := weekStart.AddDate(0, 0, 6)
			w = &changelogWeek{
				Start:  weekStart,
				End:    weekEnd,
				Label:  weekStart.Format("Jan 02") + " – " + weekEnd.Format("Jan 02"),
				ByType: map[string]int{},
			}
			weekMap[key] = w
		}

		w.Items = append(w.Items, changelogItem{Issue: ir.issue, Rig: ir.rig})
		w.Count++
		w.ByType[ir.issue.Type]++
		data.TotalClosed++
	}

	// Sort weeks by start date descending
	weeks := make([]changelogWeek, 0, len(weekMap))
	for _, w := range weekMap {
		// Sort items within week by priority then title
		sort.Slice(w.Items, func(i, j int) bool {
			if w.Items[i].Issue.Priority != w.Items[j].Issue.Priority {
				return w.Items[i].Issue.Priority < w.Items[j].Issue.Priority
			}
			return w.Items[i].Issue.Title < w.Items[j].Issue.Title
		})
		weeks = append(weeks, *w)
	}
	sort.Slice(weeks, func(i, j int) bool {
		return weeks[i].Start.After(weeks[j].Start)
	})

	data.Weeks = weeks

	// Build pre-computed query strings for filter links
	data.QNoRig = changelogQuery("", data.FilterType, data.FilterPriority)
	data.QNoType = changelogQuery(data.FilterRig, "", data.FilterPriority)
	data.QNoPriority = changelogQuery(data.FilterRig, data.FilterType, "")
	data.QAll = changelogQuery(data.FilterRig, data.FilterType, data.FilterPriority)

	s.render(w, r, "changelog", data)
}

func changelogQuery(rig, typ, priority string) string {
	v := url.Values{}
	if rig != "" {
		v.Set("rig", rig)
	}
	if typ != "" {
		v.Set("type", typ)
	}
	if priority != "" {
		v.Set("priority", priority)
	}
	if len(v) == 0 {
		return ""
	}
	return "?" + v.Encode()
}

func weekStartDate(t time.Time) time.Time {
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7 // Sunday = 7
	}
	monday := t.AddDate(0, 0, -(weekday - 1))
	return time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, time.UTC)
}
