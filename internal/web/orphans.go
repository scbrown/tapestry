package web

import (
	"log"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/scbrown/tapestry/internal/dolt"
)

type orphanItem struct {
	Rig    string
	Issue  dolt.Issue
	Age    int
	Reason string // "no owner", "no labels", "no context"
}

type orphansData struct {
	GeneratedAt time.Time
	Items       []orphanItem
	Total       int
	Rigs        []string
	FilterRig   string
	SortBy      string
	Assignees   []string
}

func (s *Server) handleOrphans(w http.ResponseWriter, r *http.Request) {
	data := orphansData{GeneratedAt: time.Now()}

	if s.ds == nil {
		s.render(w, r, "orphans", data)
		return
	}

	ctx := r.Context()
	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("orphans: list dbs: %v", err)
		s.render(w, r, "orphans", data)
		return
	}

	filterRig := r.URL.Query().Get("rig")
	sortBy := r.URL.Query().Get("sort")
	if sortBy == "" {
		sortBy = "age"
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
		items     []orphanItem
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

			// Get open issues
			open, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{Status: "open", Limit: 5000})
			if err != nil {
				log.Printf("orphans: %s: %v", dbName, err)
				return
			}

			var items []orphanItem
			sem := make(chan struct{}, 10)
			var mu sync.Mutex
			var labWg sync.WaitGroup

			for _, iss := range open {
				labWg.Add(1)
				sem <- struct{}{}
				go func(issue dolt.Issue) {
					defer labWg.Done()
					defer func() { <-sem }()

					noOwner := issue.Owner == ""
					noAssignee := issue.Assignee == ""
					noDesc := len(issue.Description) < 10

					// Check labels
					labels, err := s.ds.LabelsForIssue(ctx, dbName, issue.ID)
					noLabels := err != nil || len(labels) == 0

					// Must be missing at least 2 of: owner, assignee, labels, description
					missingCount := 0
					if noOwner {
						missingCount++
					}
					if noAssignee {
						missingCount++
					}
					if noLabels {
						missingCount++
					}
					if noDesc {
						missingCount++
					}

					if missingCount >= 3 {
						reason := "no"
						parts := []string{}
						if noOwner {
							parts = append(parts, "owner")
						}
						if noAssignee {
							parts = append(parts, "assignee")
						}
						if noLabels {
							parts = append(parts, "labels")
						}
						if noDesc {
							parts = append(parts, "description")
						}
						for i, p := range parts {
							if i == 0 {
								reason += " " + p
							} else {
								reason += ", " + p
							}
						}

						age := int(now.Sub(issue.CreatedAt).Hours() / 24)
						mu.Lock()
						items = append(items, orphanItem{
							Rig:    dbName,
							Issue:  issue,
							Age:    age,
							Reason: reason,
						})
						mu.Unlock()
					}
				}(iss)
			}
			labWg.Wait()

			results[idx] = dbResult{rig: dbName, items: items, assignees: assignees}
		}(i, db.Name)
	}
	wg.Wait()

	var allItems []orphanItem
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

	// Sort
	switch sortBy {
	case "priority":
		sort.Slice(allItems, func(i, j int) bool {
			if allItems[i].Issue.Priority != allItems[j].Issue.Priority {
				return allItems[i].Issue.Priority < allItems[j].Issue.Priority
			}
			return allItems[i].Age > allItems[j].Age
		})
	case "reason":
		sort.Slice(allItems, func(i, j int) bool {
			if allItems[i].Reason != allItems[j].Reason {
				return allItems[i].Reason < allItems[j].Reason
			}
			return allItems[i].Age > allItems[j].Age
		})
	case "rig":
		sort.Slice(allItems, func(i, j int) bool {
			if allItems[i].Rig != allItems[j].Rig {
				return allItems[i].Rig < allItems[j].Rig
			}
			return allItems[i].Age > allItems[j].Age
		})
	default: // "age"
		sort.Slice(allItems, func(i, j int) bool {
			return allItems[i].Age > allItems[j].Age
		})
	}

	if len(allItems) > 100 {
		allItems = allItems[:100]
	}

	data.Items = allItems
	data.Total = len(allItems)
	s.render(w, r, "orphans", data)
}
