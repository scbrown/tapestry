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

type recapItem struct {
	Issue dolt.Issue
	Rig   string
}

type recapData struct {
	GeneratedAt time.Time
	Date        time.Time
	DateLabel   string
	PrevDate    string
	NextDate    string
	IsToday     bool

	CreatedItems []recapItem
	ClosedItems  []recapItem
	ActiveItems  []recapItem

	CreatedByAgent map[string]int
	ClosedByAgent  map[string]int

	TotalCreated int
	TotalClosed  int
	TotalActive  int
	NetChange    int

	Rigs      []string
	FilterRig string
}

func (s *Server) handleRecap(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	data := recapData{GeneratedAt: now}

	// Parse date from query param, default to today
	dateStr := r.URL.Query().Get("date")
	var targetDate time.Time
	if dateStr != "" {
		parsed, err := time.Parse("2006-01-02", dateStr)
		if err == nil {
			targetDate = parsed
		}
	}
	if targetDate.IsZero() {
		targetDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	} else {
		targetDate = time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 0, 0, 0, 0, now.Location())
	}

	data.Date = targetDate
	data.DateLabel = targetDate.Format("Monday, January 2, 2006")
	data.PrevDate = targetDate.AddDate(0, 0, -1).Format("2006-01-02")
	data.NextDate = targetDate.AddDate(0, 0, 1).Format("2006-01-02")

	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	data.IsToday = targetDate.Equal(todayStart)

	if s.ds == nil {
		s.render(w, r, "recap", data)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 20*time.Second)
	defer cancel()

	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("recap: list dbs: %v", err)
		s.render(w, r, "recap", data)
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

	dayStart := targetDate
	dayEnd := targetDate.AddDate(0, 0, 1)

	type dbResult struct {
		created []recapItem
		closed  []recapItem
		active  []recapItem
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

			// Get all issues updated on this day (captures closed + status changes)
			updated, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{
				UpdatedAfter:  dayStart,
				UpdatedBefore: dayEnd,
				Limit:         500,
			})
			if err != nil {
				log.Printf("recap: updated %s: %v", dbName, err)
				results[i] = r
				return
			}

			for _, iss := range updated {
				if isNoise(iss.ID, iss.Title) {
					continue
				}
				item := recapItem{Issue: iss, Rig: rig}
				if iss.Status == "closed" || iss.Status == "completed" {
					r.closed = append(r.closed, item)
				}
			}

			// Get all issues to find created-on-this-day
			all, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{Limit: 1000})
			if err != nil {
				log.Printf("recap: all %s: %v", dbName, err)
				results[i] = r
				return
			}
			for _, iss := range all {
				if isNoise(iss.ID, iss.Title) {
					continue
				}
				if !iss.CreatedAt.Before(dayStart) && iss.CreatedAt.Before(dayEnd) {
					r.created = append(r.created, recapItem{Issue: iss, Rig: rig})
				}
				if (iss.Status == "in_progress" || iss.Status == "hooked") &&
					!iss.UpdatedAt.Before(dayStart) && iss.UpdatedAt.Before(dayEnd) {
					r.active = append(r.active, recapItem{Issue: iss, Rig: rig})
				}
			}

			results[i] = r
		}(i, db.Name)
	}
	wg.Wait()

	// Aggregate
	data.CreatedByAgent = make(map[string]int)
	data.ClosedByAgent = make(map[string]int)
	seen := make(map[string]bool)

	for _, r := range results {
		for _, item := range r.created {
			if !seen["c:"+item.Issue.ID] {
				seen["c:"+item.Issue.ID] = true
				data.CreatedItems = append(data.CreatedItems, item)
				agent := item.Issue.Owner
				if agent != "" {
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
	sortByPriority := func(items []recapItem) {
		sort.Slice(items, func(i, j int) bool {
			if items[i].Issue.Priority != items[j].Issue.Priority {
				return items[i].Issue.Priority < items[j].Issue.Priority
			}
			return items[i].Issue.UpdatedAt.After(items[j].Issue.UpdatedAt)
		})
	}
	sortByPriority(data.CreatedItems)
	sortByPriority(data.ClosedItems)
	sortByPriority(data.ActiveItems)

	data.TotalCreated = len(data.CreatedItems)
	data.TotalClosed = len(data.ClosedItems)
	data.TotalActive = len(data.ActiveItems)
	data.NetChange = data.TotalCreated - data.TotalClosed

	s.render(w, r, "recap", data)
}

func rigDisplayName(dbName string) string {
	if len(dbName) > 6 && dbName[:6] == "beads_" {
		return dbName[6:]
	}
	return dbName
}
