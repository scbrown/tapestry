// Package events reads Gas Town .events.jsonl files.
package events

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Event represents a single Gas Town event.
type Event struct {
	Timestamp  time.Time              `json:"ts"`
	Source     string                 `json:"source"`
	Type       string                 `json:"type"`
	Actor      string                 `json:"actor"`
	Payload    map[string]interface{} `json:"payload"`
	Visibility string                 `json:"visibility"`
}

// Filter controls which events are returned.
type Filter struct {
	Type  string // filter by event type (empty = all)
	Actor string // filter by actor (empty = all)
	After time.Time
	Limit int // max events (0 = no limit)
}

// ReadFile reads events from a .events.jsonl file, returning them in
// reverse chronological order (newest first).
func ReadFile(path string) ([]Event, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("events: open: %w", err)
	}
	defer func() { _ = f.Close() }()

	var events []Event
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 256*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var e Event
		if err := json.Unmarshal(line, &e); err != nil {
			continue // skip malformed lines
		}
		events = append(events, e)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("events: scan: %w", err)
	}

	sort.Slice(events, func(i, j int) bool {
		return events[i].Timestamp.After(events[j].Timestamp)
	})

	return events, nil
}

// ReadWorkspace reads events from a Gas Town workspace directory,
// looking for .events.jsonl at the workspace root.
func ReadWorkspace(wsPath string) ([]Event, error) {
	eventsFile := filepath.Join(wsPath, ".events.jsonl")
	return ReadFile(eventsFile)
}

// Apply filters events according to the filter criteria.
func Apply(events []Event, f Filter) []Event {
	var result []Event
	for _, e := range events {
		if f.Type != "" && e.Type != f.Type {
			continue
		}
		if f.Actor != "" && !strings.Contains(e.Actor, f.Actor) {
			continue
		}
		if !f.After.IsZero() && e.Timestamp.Before(f.After) {
			continue
		}
		result = append(result, e)
		if f.Limit > 0 && len(result) >= f.Limit {
			break
		}
	}
	return result
}

// Types returns the distinct event types from the list.
func Types(events []Event) []string {
	seen := make(map[string]bool)
	var types []string
	for _, e := range events {
		if !seen[e.Type] {
			seen[e.Type] = true
			types = append(types, e.Type)
		}
	}
	sort.Strings(types)
	return types
}

// PayloadString extracts a string value from an event's payload.
func PayloadString(e Event, key string) string {
	if v, ok := e.Payload[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// HandoffChain represents a sequence of handoff events for a single actor,
// showing the continuity of work across session boundaries.
type HandoffChain struct {
	Actor    string         // e.g., "aegis/crew/goldblum"
	Handoffs []HandoffEvent // chronological (oldest first)
}

// HandoffEvent is a single handoff within a chain.
type HandoffEvent struct {
	Timestamp time.Time
	Subject   string // from payload.subject
	// Duration since previous handoff (zero for first in chain)
	SessionDuration time.Duration
}

// BuildHandoffChains groups handoff events by actor and reconstructs
// the chain of session handoffs. Each chain represents one actor's
// sequence of context cycles.
//
// Events must be sorted (any order); chains are returned sorted by
// most recent handoff first. Each chain's handoffs are chronological.
func BuildHandoffChains(allEvents []Event) []HandoffChain {
	// Group handoff events by actor (chronological order)
	byActor := make(map[string][]Event)
	for _, e := range allEvents {
		if e.Type != "handoff" {
			continue
		}
		byActor[e.Actor] = append(byActor[e.Actor], e)
	}

	var chains []HandoffChain
	for actor, evts := range byActor {
		// Sort chronologically (oldest first)
		sort.Slice(evts, func(i, j int) bool {
			return evts[i].Timestamp.Before(evts[j].Timestamp)
		})

		chain := HandoffChain{Actor: actor}
		for i, e := range evts {
			he := HandoffEvent{
				Timestamp: e.Timestamp,
				Subject:   PayloadString(e, "subject"),
			}
			if i > 0 {
				he.SessionDuration = e.Timestamp.Sub(evts[i-1].Timestamp)
			}
			chain.Handoffs = append(chain.Handoffs, he)
		}
		chains = append(chains, chain)
	}

	// Sort chains by most recent handoff (most active actors first)
	sort.Slice(chains, func(i, j int) bool {
		iLast := chains[i].Handoffs[len(chains[i].Handoffs)-1].Timestamp
		jLast := chains[j].Handoffs[len(chains[j].Handoffs)-1].Timestamp
		return iLast.After(jLast)
	})

	return chains
}

// ChainStats summarizes handoff activity for an actor.
type ChainStats struct {
	Actor          string
	TotalHandoffs  int
	AvgSessionTime time.Duration
	LastHandoff    time.Time
	LastSubject    string
}

// ChainSummary returns summary statistics for each handoff chain.
func ChainSummary(chains []HandoffChain) []ChainStats {
	var stats []ChainStats
	for _, c := range chains {
		s := ChainStats{
			Actor:         c.Actor,
			TotalHandoffs: len(c.Handoffs),
			LastHandoff:   c.Handoffs[len(c.Handoffs)-1].Timestamp,
			LastSubject:   c.Handoffs[len(c.Handoffs)-1].Subject,
		}
		var totalDur time.Duration
		durCount := 0
		for _, h := range c.Handoffs {
			if h.SessionDuration > 0 {
				totalDur += h.SessionDuration
				durCount++
			}
		}
		if durCount > 0 {
			s.AvgSessionTime = totalDur / time.Duration(durCount)
		}
		stats = append(stats, s)
	}
	return stats
}
