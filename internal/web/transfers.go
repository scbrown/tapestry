package web

import (
	"log"
	"net/http"
	"sort"
	"sync"
	"time"
)

type transferEvent struct {
	Rig       string
	Timestamp time.Time
	IssueID   string
	Title     string
	From      string
	To        string
}

type transferData struct {
	GeneratedAt time.Time
	Transfers   []transferEvent
	Total       int
	Rigs        []string
	FilterRig   string
	Window      string // "7d", "30d", "90d"
	Err         string
}

func (s *Server) handleTransfers(w http.ResponseWriter, r *http.Request) {
	data := transferData{GeneratedAt: time.Now(), Window: "30d"}

	if s.ds == nil {
		s.render(w, r, "transfers", data)
		return
	}

	ctx := r.Context()
	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("transfers: list dbs: %v", err)
		s.render(w, r, "transfers", data)
		return
	}

	filterRig := r.URL.Query().Get("rig")
	window := r.URL.Query().Get("window")
	if window == "" {
		window = "30d"
	}
	data.FilterRig = filterRig
	data.Window = window

	var rigs []string
	for _, db := range dbs {
		rigs = append(rigs, db.Name)
	}
	sort.Strings(rigs)
	data.Rigs = rigs

	var since time.Time
	switch window {
	case "7d":
		since = time.Now().AddDate(0, 0, -7)
	case "90d":
		since = time.Now().AddDate(0, 0, -90)
	default:
		since = time.Now().AddDate(0, 0, -30)
	}

	type dbResult struct {
		events []transferEvent
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
			diffs, err := s.ds.IssueDiffSince(ctx, dbName, since)
			if err != nil {
				log.Printf("transfers: %s diff: %v", dbName, err)
				return
			}

			var events []transferEvent
			for _, d := range diffs {
				if d.FromAssignee == d.ToAssignee || (d.FromAssignee == "" && d.ToAssignee == "") {
					continue
				}
				from := d.FromAssignee
				if from == "" {
					from = "(unassigned)"
				}
				to := d.ToAssignee
				if to == "" {
					to = "(unassigned)"
				}
				events = append(events, transferEvent{
					Rig:       dbName,
					Timestamp: d.ToCommitDate,
					IssueID:   d.ToID,
					Title:     d.ToTitle,
					From:      from,
					To:        to,
				})
			}
			results[idx] = dbResult{events: events}
		}(i, db.Name)
	}
	wg.Wait()

	var all []transferEvent
	for _, r := range results {
		all = append(all, r.events...)
	}

	sort.Slice(all, func(i, j int) bool {
		return all[i].Timestamp.After(all[j].Timestamp)
	})

	if len(all) > 200 {
		all = all[:200]
	}

	data.Transfers = all
	data.Total = len(all)
	s.render(w, r, "transfers", data)
}
