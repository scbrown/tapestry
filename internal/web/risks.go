package web

import (
	"log"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/scbrown/tapestry/internal/dolt"
)

type riskItem struct {
	Issue    dolt.Issue
	Rig      string
	Reason   string
	Severity string // "critical", "warning", "info"
	AgeDays  int
}

type risksData struct {
	GeneratedAt time.Time
	Items       []riskItem
	Critical    int
	Warning     int
	Info        int
	Rigs        []string
	FilterRig   string
	SortBy      string
	Assignees   []string
	Err         string
}

func (s *Server) handleRisks(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	data := risksData{GeneratedAt: now}

	if s.ds == nil {
		s.render(w, r, "risks", data)
		return
	}

	ctx := r.Context()
	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("risks: list dbs: %v", err)
		s.render(w, r, "risks", risksData{Err: err.Error(), GeneratedAt: now})
		return
	}

	type dbResult struct {
		items     []riskItem
		assignees []string
	}

	staleThreshold := now.AddDate(0, 0, -7)
	veryStaleThreshold := now.AddDate(0, 0, -14)

	results := make([]dbResult, len(dbs))
	var wg sync.WaitGroup
	for i, db := range dbs {
		wg.Add(1)
		go func(i int, dbName string) {
			defer wg.Done()
			var r dbResult
			r.assignees, _ = s.ds.DistinctAssignees(ctx, dbName)

			issues, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{Limit: 500})
			if err != nil {
				log.Printf("risks %s: %v", dbName, err)
				results[i] = r
				return
			}

			for _, iss := range issues {
				if isNoise(iss.ID, iss.Title) || iss.Status == "closed" || iss.Status == "deferred" {
					continue
				}

				age := int(now.Sub(iss.CreatedAt).Hours() / 24)

				// Critical: P0/P1 that are stale
				if iss.Priority <= 1 && (iss.Status == "open" || iss.Status == "in_progress") &&
					iss.UpdatedAt.Before(staleThreshold) {
					severity := "warning"
					reason := "P" + string(rune('0'+iss.Priority)) + " stale 7+ days"
					if iss.UpdatedAt.Before(veryStaleThreshold) {
						severity = "critical"
						reason = "P" + string(rune('0'+iss.Priority)) + " stale 14+ days"
					}
					r.items = append(r.items, riskItem{
						Issue: iss, Rig: dbName, Reason: reason, Severity: severity, AgeDays: age,
					})
				}

				// Warning: High priority with no assignee
				if iss.Priority <= 2 && iss.Status != "blocked" && iss.Assignee == "" &&
					iss.UpdatedAt.Before(staleThreshold) {
					r.items = append(r.items, riskItem{
						Issue: iss, Rig: dbName, Reason: "P" + string(rune('0'+iss.Priority)) + " unassigned and stale",
						Severity: "warning", AgeDays: age,
					})
				}

				// Warning: Blocked for a long time
				if iss.Status == "blocked" && iss.Priority <= 2 && iss.UpdatedAt.Before(staleThreshold) {
					r.items = append(r.items, riskItem{
						Issue: iss, Rig: dbName, Reason: "Blocked 7+ days",
						Severity: "warning", AgeDays: age,
					})
				}

				// Info: Very old open items (30+ days)
				if (iss.Status == "open" || iss.Status == "in_progress") &&
					iss.Priority <= 3 && age > 30 {
					r.items = append(r.items, riskItem{
						Issue: iss, Rig: dbName, Reason: "Open 30+ days",
						Severity: "info", AgeDays: age,
					})
				}
			}

			results[i] = r
		}(i, db.Name)
	}
	wg.Wait()

	// Aggregate and deduplicate (prefer higher severity)
	seen := make(map[string]int) // issue ID -> index in items
	rigSet := make(map[string]bool)
	assigneeSet := make(map[string]bool)

	for _, r := range results {
		for _, a := range r.assignees {
			assigneeSet[a] = true
		}
		for _, item := range r.items {
			rigSet[item.Rig] = true
			key := item.Rig + "/" + item.Issue.ID
			if idx, ok := seen[key]; ok {
				// Keep the higher severity one
				if severityRank(item.Severity) > severityRank(data.Items[idx].Severity) {
					data.Items[idx] = item
				}
			} else {
				seen[key] = len(data.Items)
				data.Items = append(data.Items, item)
			}
		}
	}

	// Build rig list
	for rig := range rigSet {
		data.Rigs = append(data.Rigs, rig)
	}
	sort.Strings(data.Rigs)

	// Build assignee list
	for a := range assigneeSet {
		data.Assignees = append(data.Assignees, a)
	}
	sort.Strings(data.Assignees)

	// Apply rig filter
	filterRig := r.URL.Query().Get("rig")
	sortBy := r.URL.Query().Get("sort")
	if sortBy == "" {
		sortBy = "severity"
	}
	data.FilterRig = filterRig
	data.SortBy = sortBy
	if filterRig != "" {
		filtered := data.Items[:0]
		for _, item := range data.Items {
			if item.Rig == filterRig {
				filtered = append(filtered, item)
			}
		}
		data.Items = filtered
	}

	switch sortBy {
	case "priority":
		sort.Slice(data.Items, func(i, j int) bool {
			if data.Items[i].Issue.Priority != data.Items[j].Issue.Priority {
				return data.Items[i].Issue.Priority < data.Items[j].Issue.Priority
			}
			return severityRank(data.Items[i].Severity) > severityRank(data.Items[j].Severity)
		})
	case "age":
		sort.Slice(data.Items, func(i, j int) bool {
			return data.Items[i].AgeDays > data.Items[j].AgeDays
		})
	case "rig":
		sort.Slice(data.Items, func(i, j int) bool {
			if data.Items[i].Rig != data.Items[j].Rig {
				return data.Items[i].Rig < data.Items[j].Rig
			}
			return severityRank(data.Items[i].Severity) > severityRank(data.Items[j].Severity)
		})
	default: // "severity"
		sort.Slice(data.Items, func(i, j int) bool {
			si, sj := severityRank(data.Items[i].Severity), severityRank(data.Items[j].Severity)
			if si != sj {
				return si > sj
			}
			if data.Items[i].Issue.Priority != data.Items[j].Issue.Priority {
				return data.Items[i].Issue.Priority < data.Items[j].Issue.Priority
			}
			return data.Items[i].AgeDays > data.Items[j].AgeDays
		})
	}

	// Count severities
	for _, item := range data.Items {
		switch item.Severity {
		case "critical":
			data.Critical++
		case "warning":
			data.Warning++
		case "info":
			data.Info++
		}
	}

	// Cap at 50
	if len(data.Items) > 50 {
		data.Items = data.Items[:50]
	}

	s.render(w, r, "risks", data)
}

func severityRank(s string) int {
	switch s {
	case "critical":
		return 3
	case "warning":
		return 2
	case "info":
		return 1
	default:
		return 0
	}
}
