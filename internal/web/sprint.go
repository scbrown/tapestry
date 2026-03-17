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

type sprintItem struct {
	Issue dolt.Issue
	Rig   string
}

type sprintData struct {
	GeneratedAt time.Time
	WeekStart   time.Time
	WeekEnd     time.Time
	WeekLabel   string
	PrevWeek    string
	NextWeek    string
	IsThisWeek  bool

	CreatedItems []sprintItem
	ClosedItems  []sprintItem
	ActiveItems  []sprintItem

	TotalCreated int
	TotalClosed  int
	TotalActive  int
	NetChange    int

	ClosedByAgent  map[string]int
	CreatedByAgent map[string]int

	Rigs      []string
	FilterRig string
	Assignees []string
}

func weekStart(t time.Time) time.Time {
	// ISO week: Monday is day 0
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7 // Sunday = 7
	}
	return time.Date(t.Year(), t.Month(), t.Day()-(weekday-1), 0, 0, 0, 0, t.Location())
}

func (s *Server) handleSprint(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	data := sprintData{GeneratedAt: now}

	// Parse week from query param, default to this week
	var target time.Time
	if ws := r.URL.Query().Get("week"); ws != "" {
		parsed, err := time.Parse("2006-01-02", ws)
		if err == nil {
			target = parsed
		}
	}
	if target.IsZero() {
		target = now
	}

	ws := weekStart(target)
	we := ws.AddDate(0, 0, 7)

	data.WeekStart = ws
	data.WeekEnd = we
	data.WeekLabel = ws.Format("Jan 2") + " – " + we.Add(-24*time.Hour).Format("Jan 2, 2006")
	data.PrevWeek = ws.AddDate(0, 0, -7).Format("2006-01-02")
	data.NextWeek = ws.AddDate(0, 0, 7).Format("2006-01-02")

	thisWeekStart := weekStart(now)
	data.IsThisWeek = ws.Equal(thisWeekStart)

	if s.ds == nil {
		s.render(w, r, "sprint", data)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 20*time.Second)
	defer cancel()

	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("sprint: list dbs: %v", err)
		s.render(w, r, "sprint", data)
		return
	}

	filterRig := r.URL.Query().Get("rig")
	var rigs []string
	for _, db := range dbs {
		rigs = append(rigs, db.Name)
	}
	sort.Strings(rigs)
	data.Rigs = rigs
	data.FilterRig = filterRig

	type dbResult struct {
		created   []sprintItem
		closed    []sprintItem
		active    []sprintItem
		assignees []string
	}

	results := make([]dbResult, len(dbs))
	var wg sync.WaitGroup
	for i, db := range dbs {
		if filterRig != "" && db.Name != filterRig {
			continue
		}
		wg.Add(1)
		go func(i int, dbName string) {
			defer wg.Done()
			var r dbResult
			rig := rigDisplayName(dbName)

			r.assignees, _ = s.ds.DistinctAssignees(ctx, dbName)

			// Get all issues to filter
			all, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{Limit: 2000})
			if err != nil {
				log.Printf("sprint: issues %s: %v", dbName, err)
				results[i] = r
				return
			}

			for _, iss := range all {
				if isNoise(iss.ID, iss.Title) {
					continue
				}
				item := sprintItem{Issue: iss, Rig: rig}

				// Created this week
				if !iss.CreatedAt.Before(ws) && iss.CreatedAt.Before(we) {
					r.created = append(r.created, item)
				}

				// Closed this week
				if (iss.Status == "closed" || iss.Status == "completed") &&
					!iss.UpdatedAt.Before(ws) && iss.UpdatedAt.Before(we) {
					r.closed = append(r.closed, item)
				}

				// Active this week (in_progress/hooked with updates in range)
				if (iss.Status == "in_progress" || iss.Status == "hooked") &&
					!iss.UpdatedAt.Before(ws) && iss.UpdatedAt.Before(we) {
					r.active = append(r.active, item)
				}
			}

			results[i] = r
		}(i, db.Name)
	}
	wg.Wait()

	// Collect assignees
	assigneeSet := make(map[string]bool)
	for _, r := range results {
		for _, a := range r.assignees {
			if a != "" {
				assigneeSet[a] = true
			}
		}
	}
	for a := range assigneeSet {
		data.Assignees = append(data.Assignees, a)
	}
	sort.Strings(data.Assignees)

	// Aggregate
	data.CreatedByAgent = make(map[string]int)
	data.ClosedByAgent = make(map[string]int)
	seen := make(map[string]bool)

	for _, r := range results {
		for _, item := range r.created {
			if !seen["c:"+item.Issue.ID] {
				seen["c:"+item.Issue.ID] = true
				data.CreatedItems = append(data.CreatedItems, item)
				if agent := item.Issue.Owner; agent != "" {
					data.CreatedByAgent[agent]++
				}
			}
		}
		for _, item := range r.closed {
			if !seen["x:"+item.Issue.ID] {
				seen["x:"+item.Issue.ID] = true
				data.ClosedItems = append(data.ClosedItems, item)
				agent := item.Issue.Assignee
				if agent == "" {
					agent = item.Issue.Owner
				}
				if agent != "" {
					data.ClosedByAgent[agent]++
				}
			}
		}
		for _, item := range r.active {
			if !seen["a:"+item.Issue.ID] {
				seen["a:"+item.Issue.ID] = true
				data.ActiveItems = append(data.ActiveItems, item)
			}
		}
	}

	// Sort by priority
	sortByPri := func(items []sprintItem) {
		sort.Slice(items, func(i, j int) bool {
			if items[i].Issue.Priority != items[j].Issue.Priority {
				return items[i].Issue.Priority < items[j].Issue.Priority
			}
			return items[i].Issue.UpdatedAt.After(items[j].Issue.UpdatedAt)
		})
	}
	sortByPri(data.CreatedItems)
	sortByPri(data.ClosedItems)
	sortByPri(data.ActiveItems)

	data.TotalCreated = len(data.CreatedItems)
	data.TotalClosed = len(data.ClosedItems)
	data.TotalActive = len(data.ActiveItems)
	data.NetChange = data.TotalCreated - data.TotalClosed

	s.render(w, r, "sprint", data)
}
