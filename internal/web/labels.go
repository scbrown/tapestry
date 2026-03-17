package web

import (
	"log"
	"net/http"
	"sort"
	"sync"

	"github.com/scbrown/tapestry/internal/dolt"
)

type labelEntry struct {
	Label string
	Count int
	Rig   string
}

type labelsData struct {
	FilterRig string
	Rigs      []string
	Labels    []labelEntry
	Total     int
	Filter    string // selected label
	Issues    []labelIssueEntry
	Assignees []string
	Err       string
}

type labelIssueEntry struct {
	Issue dolt.Issue
	Rig   string
}

func (s *Server) handleLabels(w http.ResponseWriter, r *http.Request) {
	filter := r.URL.Query().Get("label")
	filterRig := r.URL.Query().Get("rig")

	if s.ds == nil {
		s.render(w, r, "labels", labelsData{})
		return
	}

	ctx := r.Context()
	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("labels: list dbs: %v", err)
		s.render(w, r, "labels", labelsData{Err: err.Error()})
		return
	}

	var rigs []string
	for _, db := range dbs {
		rigs = append(rigs, db.Name)
	}
	sort.Strings(rigs)

	type dbResult struct {
		labels    []labelEntry
		issues    []labelIssueEntry
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
			var res dbResult

			lcs, err := s.ds.DistinctLabels(ctx, dbName)
			if err != nil {
				log.Printf("labels %s: %v", dbName, err)
				results[i] = res
				return
			}
			for _, lc := range lcs {
				res.labels = append(res.labels, labelEntry{
					Label: lc.Label,
					Count: lc.Count,
					Rig:   dbName,
				})
			}

			if filter != "" {
				issues, err := s.ds.IssuesByLabel(ctx, dbName, filter)
				if err != nil {
					log.Printf("labels %s issues: %v", dbName, err)
				} else {
					for _, iss := range issues {
						res.issues = append(res.issues, labelIssueEntry{Issue: iss, Rig: dbName})
					}
				}
				res.assignees, _ = s.ds.DistinctAssignees(ctx, dbName)
			}

			results[i] = res
		}(i, db.Name)
	}
	wg.Wait()

	// Merge labels across rigs by name
	merged := map[string]int{}
	var allIssues []labelIssueEntry
	assigneeSet := make(map[string]bool)
	for _, r := range results {
		for _, le := range r.labels {
			merged[le.Label] += le.Count
		}
		allIssues = append(allIssues, r.issues...)
		for _, a := range r.assignees {
			assigneeSet[a] = true
		}
	}
	var assignees []string
	for a := range assigneeSet {
		assignees = append(assignees, a)
	}
	sort.Strings(assignees)

	var labels []labelEntry
	for label, count := range merged {
		labels = append(labels, labelEntry{Label: label, Count: count})
	}
	sort.Slice(labels, func(i, j int) bool {
		if labels[i].Count != labels[j].Count {
			return labels[i].Count > labels[j].Count
		}
		return labels[i].Label < labels[j].Label
	})

	sort.Slice(allIssues, func(i, j int) bool {
		if allIssues[i].Issue.Priority != allIssues[j].Issue.Priority {
			return allIssues[i].Issue.Priority < allIssues[j].Issue.Priority
		}
		return allIssues[i].Issue.UpdatedAt.After(allIssues[j].Issue.UpdatedAt)
	})

	s.render(w, r, "labels", labelsData{
		FilterRig: filterRig,
		Rigs:      rigs,
		Labels:    labels,
		Total:     len(labels),
		Filter:    filter,
		Issues:    allIssues,
		Assignees: assignees,
	})
}
