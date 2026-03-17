package web

import (
	"log"
	"net/http"
	"sort"
	"sync"
	"time"
)

type auditEvent struct {
	Rig       string
	Timestamp time.Time
	Kind      string // "status", "created", "comment", "reassign"
	IssueID   string
	Title     string
	Detail    string // e.g. "open → closed", "new comment by aegis/arnold"
}

type auditLogData struct {
	GeneratedAt time.Time
	Events      []auditEvent
	Total       int
	Rigs        []string
	FilterRig   string
	Window      string // "24h", "7d", "30d"
}

func (s *Server) handleAuditLog(w http.ResponseWriter, r *http.Request) {
	data := auditLogData{GeneratedAt: time.Now()}

	if s.ds == nil {
		s.render(w, r, "audit-log", data)
		return
	}

	ctx := r.Context()
	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("audit-log: list dbs: %v", err)
		s.render(w, r, "audit-log", data)
		return
	}

	filterRig := r.URL.Query().Get("rig")
	window := r.URL.Query().Get("window")
	if window == "" {
		window = "7d"
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
	case "24h":
		since = time.Now().AddDate(0, 0, -1)
	case "30d":
		since = time.Now().AddDate(0, 0, -30)
	default:
		since = time.Now().AddDate(0, 0, -7)
	}

	type dbEvents struct {
		rig    string
		events []auditEvent
	}
	results := make([]dbEvents, len(dbs))
	var wg sync.WaitGroup

	for i, db := range dbs {
		if filterRig != "" && db.Name != filterRig {
			continue
		}
		wg.Add(1)
		go func(idx int, dbName string) {
			defer wg.Done()
			var evts []auditEvent

			// Issue diffs (status changes, new issues, reassignments)
			diffs, err := s.ds.IssueDiffSince(ctx, dbName, since)
			if err != nil {
				log.Printf("audit-log: %s issue diff: %v", dbName, err)
			} else {
				for _, d := range diffs {
					evt := auditEvent{
						Rig:       dbName,
						Timestamp: d.ToCommitDate,
						IssueID:   d.ToID,
						Title:     d.ToTitle,
					}
					switch {
					case d.DiffType == "added":
						evt.Kind = "created"
						evt.Detail = "Created as " + d.ToStatus
					case d.FromStatus != d.ToStatus && d.FromStatus != "":
						evt.Kind = "status"
						evt.Detail = d.FromStatus + " → " + d.ToStatus
					case d.FromAssignee != d.ToAssignee:
						evt.Kind = "reassign"
						from := d.FromAssignee
						if from == "" {
							from = "unassigned"
						}
						to := d.ToAssignee
						if to == "" {
							to = "unassigned"
						}
						evt.Detail = from + " → " + to
					default:
						evt.Kind = "modified"
						evt.Detail = "Updated"
					}
					evts = append(evts, evt)
				}
			}

			// Comment diffs
			cdiffs, err := s.ds.CommentDiffSince(ctx, dbName, since)
			if err != nil {
				log.Printf("audit-log: %s comment diff: %v", dbName, err)
			} else {
				for _, c := range cdiffs {
					if c.DiffType != "added" {
						continue
					}
					body := c.ToBody
					if len(body) > 80 {
						body = body[:80] + "..."
					}
					evts = append(evts, auditEvent{
						Rig:       dbName,
						Timestamp: c.ToCommitDate,
						Kind:      "comment",
						IssueID:   c.ToIssueID,
						Title:     body,
						Detail:    "by " + c.ToAuthor,
					})
				}
			}

			results[idx] = dbEvents{rig: dbName, events: evts}
		}(i, db.Name)
	}
	wg.Wait()

	var allEvents []auditEvent
	for _, r := range results {
		allEvents = append(allEvents, r.events...)
	}

	sort.Slice(allEvents, func(i, j int) bool {
		return allEvents[i].Timestamp.After(allEvents[j].Timestamp)
	})

	if len(allEvents) > 500 {
		allEvents = allEvents[:500]
	}

	data.Events = allEvents
	data.Total = len(allEvents)
	s.render(w, r, "audit-log", data)
}
