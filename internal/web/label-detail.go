package web

import (
	"log"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/scbrown/tapestry/internal/dolt"
)

type labelDetailStats struct {
	Open       int
	Closed     int
	InProgress int
	Blocked    int
	Deferred   int
	Total      int
}

type labelDetailRigGroup struct {
	Rig    string
	Issues []dolt.Issue
	Stats  labelDetailStats
}

type labelDetailData struct {
	GeneratedAt time.Time
	Label       string
	Groups      []labelDetailRigGroup
	Stats       labelDetailStats
	Rigs        []string
	FilterRig   string
}

func (s *Server) handleLabelDetail(w http.ResponseWriter, r *http.Request, label string) {
	data := labelDetailData{GeneratedAt: time.Now(), Label: label}

	if s.ds == nil {
		s.render(w, r, "label-detail", data)
		return
	}

	ctx := r.Context()
	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("label-detail: list dbs: %v", err)
		s.render(w, r, "label-detail", data)
		return
	}

	filterRig := r.URL.Query().Get("rig")
	data.FilterRig = filterRig

	var rigs []string
	for _, db := range dbs {
		rigs = append(rigs, db.Name)
	}
	sort.Strings(rigs)
	data.Rigs = rigs

	type dbResult struct {
		rig    string
		issues []dolt.Issue
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
			issues, err := s.ds.IssuesByLabel(ctx, dbName, label)
			if err != nil {
				log.Printf("label-detail: %s: %v", dbName, err)
				return
			}
			results[idx] = dbResult{rig: dbName, issues: issues}
		}(i, db.Name)
	}
	wg.Wait()

	var totalStats labelDetailStats
	for _, r := range results {
		if len(r.issues) == 0 {
			continue
		}
		var stats labelDetailStats
		for _, iss := range r.issues {
			stats.Total++
			switch iss.Status {
			case "open":
				stats.Open++
			case "closed":
				stats.Closed++
			case "in_progress", "hooked":
				stats.InProgress++
			case "blocked":
				stats.Blocked++
			case "deferred":
				stats.Deferred++
			}
		}
		totalStats.Open += stats.Open
		totalStats.Closed += stats.Closed
		totalStats.InProgress += stats.InProgress
		totalStats.Blocked += stats.Blocked
		totalStats.Deferred += stats.Deferred
		totalStats.Total += stats.Total

		// Sort by priority then status
		sort.Slice(r.issues, func(i, j int) bool {
			if r.issues[i].Priority != r.issues[j].Priority {
				return r.issues[i].Priority < r.issues[j].Priority
			}
			return r.issues[i].Status < r.issues[j].Status
		})

		data.Groups = append(data.Groups, labelDetailRigGroup{
			Rig:    r.rig,
			Issues: r.issues,
			Stats:  stats,
		})
	}

	sort.Slice(data.Groups, func(i, j int) bool {
		return data.Groups[i].Rig < data.Groups[j].Rig
	})

	data.Stats = totalStats
	s.render(w, r, "label-detail", data)
}
