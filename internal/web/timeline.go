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

type timelineEvent struct {
	Time     time.Time
	Type     string // "created", "closed", "status_change", "comment", "reassigned"
	BeadID   string
	BeadDB   string
	Title    string
	Actor    string
	Detail   string // e.g. "open → closed" or comment body snippet
	Priority int
}

type timelineData struct {
	GeneratedAt time.Time
	Events      []timelineEvent
	Window      string // "6h", "12h", "24h", "48h"
	FilterRig   string
	Total       int
	Err         string
}

func (s *Server) handleTimeline(w http.ResponseWriter, r *http.Request) {
	data := timelineData{GeneratedAt: time.Now()}

	if s.ds == nil {
		s.render(w, r, "timeline", data)
		return
	}

	ctx := r.Context()
	data.FilterRig = r.URL.Query().Get("rig")
	data.Window = r.URL.Query().Get("window")
	if data.Window == "" {
		data.Window = "24h"
	}

	var duration time.Duration
	switch data.Window {
	case "6h":
		duration = 6 * time.Hour
	case "12h":
		duration = 12 * time.Hour
	case "48h":
		duration = 48 * time.Hour
	default:
		duration = 24 * time.Hour
		data.Window = "24h"
	}

	since := time.Now().Add(-duration)

	dbs, err := s.databases(ctx)
	if err != nil {
		data.Err = err.Error()
		s.render(w, r, "timeline", data)
		return
	}

	var allEvents []timelineEvent
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, db := range dbs {
		if data.FilterRig != "" && db.Name != data.FilterRig {
			continue
		}
		wg.Add(1)
		go func(dbName string) {
			defer wg.Done()

			// Build issue ID → title lookup
			titleMap := map[string]string{}
			priMap := map[string]int{}
			issues, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{Limit: 5000})
			if err == nil {
				for _, iss := range issues {
					titleMap[iss.ID] = iss.Title
					priMap[iss.ID] = iss.Priority
				}
			}

			// Issue diffs
			diffs, err := s.ds.IssueDiffSince(ctx, dbName, since)
			if err != nil {
				log.Printf("timeline: diffs %s: %v", dbName, err)
			} else {
				var events []timelineEvent
				for _, d := range diffs {
					title := titleMap[d.ToID]
					if title == "" {
						title = d.ToTitle
					}
					pri := priMap[d.ToID]

					if d.DiffType == "added" {
						events = append(events, timelineEvent{
							Time:     d.ToCommitDate,
							Type:     "created",
							BeadID:   d.ToID,
							BeadDB:   dbName,
							Title:    title,
							Actor:    firstNonEmpty(d.ToOwner, d.ToAssignee),
							Detail:   d.ToStatus,
							Priority: pri,
						})
					} else if d.FromStatus != d.ToStatus {
						evType := "status_change"
						if d.ToStatus == "closed" {
							evType = "closed"
						}
						events = append(events, timelineEvent{
							Time:     d.ToCommitDate,
							Type:     evType,
							BeadID:   d.ToID,
							BeadDB:   dbName,
							Title:    title,
							Actor:    firstNonEmpty(d.ToAssignee, d.ToOwner),
							Detail:   d.FromStatus + " → " + d.ToStatus,
							Priority: pri,
						})
					}
					if d.FromAssignee != d.ToAssignee && d.DiffType != "added" {
						events = append(events, timelineEvent{
							Time:     d.ToCommitDate,
							Type:     "reassigned",
							BeadID:   d.ToID,
							BeadDB:   dbName,
							Title:    title,
							Actor:    firstNonEmpty(d.ToAssignee, d.ToOwner),
							Detail:   shortActorStr(d.FromAssignee) + " → " + shortActorStr(d.ToAssignee),
							Priority: pri,
						})
					}
				}
				mu.Lock()
				allEvents = append(allEvents, events...)
				mu.Unlock()
			}

			// Comment diffs
			commentDiffs, err := s.ds.CommentDiffSince(ctx, dbName, since)
			if err != nil {
				log.Printf("timeline: comments %s: %v", dbName, err)
			} else {
				var events []timelineEvent
				for _, c := range commentDiffs {
					if c.DiffType != "added" {
						continue
					}
					title := titleMap[c.ToIssueID]
					pri := priMap[c.ToIssueID]
					snippet := c.ToBody
					if len(snippet) > 120 {
						snippet = snippet[:120] + "..."
					}
					events = append(events, timelineEvent{
						Time:     c.ToCommitDate,
						Type:     "comment",
						BeadID:   c.ToIssueID,
						BeadDB:   dbName,
						Title:    title,
						Actor:    c.ToAuthor,
						Detail:   snippet,
						Priority: pri,
					})
				}
				mu.Lock()
				allEvents = append(allEvents, events...)
				mu.Unlock()
			}
		}(db.Name)
	}
	wg.Wait()

	// Sort by time descending (most recent first)
	sort.Slice(allEvents, func(i, j int) bool {
		return allEvents[i].Time.After(allEvents[j].Time)
	})

	data.Total = len(allEvents)
	// Cap at 500 events
	if len(allEvents) > 500 {
		allEvents = allEvents[:500]
	}
	data.Events = allEvents

	s.render(w, r, "timeline", data)
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

func shortActorStr(s string) string {
	if s == "" {
		return "—"
	}
	parts := strings.Split(s, "/")
	return parts[len(parts)-1]
}
