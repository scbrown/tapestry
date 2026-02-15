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
	defer f.Close()

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
