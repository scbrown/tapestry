// Package events extracts bead lifecycle events from Dolt diff history.
//
// It queries dolt_diff for the issues and comments tables to produce a
// chronological stream of events: bead creation, status changes, closures,
// assignment changes, and comment additions.
package events

import (
	"context"
	"sort"
	"time"

	"github.com/scbrown/tapestry/internal/dolt"
)

// Action describes what happened in an event.
const (
	ActionCreated    = "created"
	ActionClosed     = "closed"
	ActionReopened   = "reopened"
	ActionStatus     = "status_change"
	ActionAssigned   = "assigned"
	ActionCommented  = "commented"
	ActionRemoved    = "removed"
)

// Event represents a single bead lifecycle event extracted from Dolt
// diff history.
type Event struct {
	Timestamp time.Time
	Actor     string
	Action    string
	BeadID    string
	Title     string
	Details   map[string]string
}

// EventFilter controls which events are returned.
type EventFilter struct {
	BeadID string // filter by bead ID (empty = all)
	Actor  string // filter by actor (empty = all)
	Action string // filter by action type (empty = all)
	Limit  int    // max events (0 = no limit)
}

// MonthlyStats holds aggregated event counts for a calendar month.
type MonthlyStats struct {
	Month    time.Time      // first day of the month (UTC)
	Created  int            // beads created
	Closed   int            // beads closed
	Comments int            // comments added
	ByActor  map[string]int // total events per actor
	ByAction map[string]int // total events per action type
}

// AgentStats holds per-agent activity statistics.
type AgentStats struct {
	Agent    string
	Created  int
	Closed   int
	Comments int
	Total    int
}

// Source extracts lifecycle events from a Dolt-backed beads database.
type Source struct {
	client *dolt.Client
}

// NewSource creates an event source backed by the given Dolt client.
func NewSource(client *dolt.Client) *Source {
	return &Source{client: client}
}

// Events returns bead lifecycle events between two Dolt revisions.
// The from and to parameters can be commit hashes, branch names, or
// timestamps (same as dolt_diff).
func (s *Source) Events(ctx context.Context, database, from, to string) ([]Event, error) {
	return s.events(ctx, database, from, to, EventFilter{})
}

// EventsFiltered returns filtered bead lifecycle events between two revisions.
func (s *Source) EventsFiltered(ctx context.Context, database, from, to string, f EventFilter) ([]Event, error) {
	return s.events(ctx, database, from, to, f)
}

func (s *Source) events(ctx context.Context, database, from, to string, f EventFilter) ([]Event, error) {
	issueDiffs, err := s.client.IssueDiffs(ctx, database, from, to)
	if err != nil {
		return nil, err
	}

	commentDiffs, err := s.client.CommentDiffs(ctx, database, from, to)
	if err != nil {
		return nil, err
	}

	var events []Event

	for _, d := range issueDiffs {
		events = append(events, issueEvents(d)...)
	}

	for _, d := range commentDiffs {
		if d.DiffType != "added" {
			continue
		}
		events = append(events, Event{
			Timestamp: d.ToCommitDate,
			Actor:     d.ToAuthor,
			Action:    ActionCommented,
			BeadID:    d.ToIssueID,
			Details: map[string]string{
				"comment_id": d.ToID,
				"body":       truncate(d.ToBody, 200),
			},
		})
	}

	// Sort chronologically.
	sort.Slice(events, func(i, j int) bool {
		return events[i].Timestamp.Before(events[j].Timestamp)
	})

	events = applyFilter(events, f)
	return events, nil
}

// issueEvents extracts zero or more events from a single issue diff row.
func issueEvents(d dolt.IssueDiffRow) []Event {
	var events []Event
	base := Event{
		Timestamp: d.ToCommitDate,
		BeadID:    d.ToID,
		Title:     d.ToTitle,
	}

	switch d.DiffType {
	case "added":
		ev := base
		ev.Action = ActionCreated
		ev.Actor = d.ToOwner
		events = append(events, ev)

	case "removed":
		ev := base
		ev.Action = ActionRemoved
		ev.Actor = d.ToOwner
		events = append(events, ev)

	case "modified":
		// Status change.
		if d.FromStatus != d.ToStatus {
			ev := base
			ev.Actor = d.ToOwner
			ev.Details = map[string]string{
				"from_status": d.FromStatus,
				"to_status":   d.ToStatus,
			}

			switch {
			case d.ToStatus == "closed":
				ev.Action = ActionClosed
			case d.FromStatus == "closed":
				ev.Action = ActionReopened
			default:
				ev.Action = ActionStatus
			}
			events = append(events, ev)
		}

		// Assignee change.
		if d.FromAssignee != d.ToAssignee {
			ev := base
			ev.Action = ActionAssigned
			ev.Actor = d.ToOwner
			ev.Details = map[string]string{
				"from_assignee": d.FromAssignee,
				"to_assignee":   d.ToAssignee,
			}
			events = append(events, ev)
		}
	}

	return events
}

// applyFilter removes events that don't match the filter criteria.
func applyFilter(events []Event, f EventFilter) []Event {
	if f.BeadID == "" && f.Actor == "" && f.Action == "" && f.Limit == 0 {
		return events
	}

	var filtered []Event
	for _, ev := range events {
		if f.BeadID != "" && ev.BeadID != f.BeadID {
			continue
		}
		if f.Actor != "" && ev.Actor != f.Actor {
			continue
		}
		if f.Action != "" && ev.Action != f.Action {
			continue
		}
		filtered = append(filtered, ev)
		if f.Limit > 0 && len(filtered) >= f.Limit {
			break
		}
	}
	return filtered
}

// Monthly aggregates events into per-month statistics.
func Monthly(events []Event) []MonthlyStats {
	months := make(map[time.Time]*MonthlyStats)

	for _, ev := range events {
		key := time.Date(ev.Timestamp.Year(), ev.Timestamp.Month(), 1, 0, 0, 0, 0, time.UTC)
		ms, ok := months[key]
		if !ok {
			ms = &MonthlyStats{
				Month:    key,
				ByActor:  make(map[string]int),
				ByAction: make(map[string]int),
			}
			months[key] = ms
		}

		ms.ByAction[ev.Action]++
		if ev.Actor != "" {
			ms.ByActor[ev.Actor]++
		}

		switch ev.Action {
		case ActionCreated:
			ms.Created++
		case ActionClosed:
			ms.Closed++
		case ActionCommented:
			ms.Comments++
		}
	}

	result := make([]MonthlyStats, 0, len(months))
	for _, ms := range months {
		result = append(result, *ms)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Month.Before(result[j].Month)
	})

	return result
}

// AgentActivity computes per-agent statistics from a set of events.
func AgentActivity(events []Event) []AgentStats {
	agents := make(map[string]*AgentStats)

	for _, ev := range events {
		if ev.Actor == "" {
			continue
		}
		as, ok := agents[ev.Actor]
		if !ok {
			as = &AgentStats{Agent: ev.Actor}
			agents[ev.Actor] = as
		}

		as.Total++
		switch ev.Action {
		case ActionCreated:
			as.Created++
		case ActionClosed:
			as.Closed++
		case ActionCommented:
			as.Comments++
		}
	}

	result := make([]AgentStats, 0, len(agents))
	for _, as := range agents {
		result = append(result, *as)
	}

	// Sort by total activity descending.
	sort.Slice(result, func(i, j int) bool {
		return result[i].Total > result[j].Total
	})

	return result
}

// truncate shortens s to at most n bytes, appending "..." if truncated.
func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
