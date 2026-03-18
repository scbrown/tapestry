package events

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestReadFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".events.jsonl")

	content := `{"ts":"2026-03-17T10:00:00Z","source":"test","type":"handoff","actor":"aegis/crew/arnold","payload":{"subject":"session 1"}}
{"ts":"2026-03-17T11:00:00Z","source":"test","type":"sling","actor":"aegis/crew/goldblum","payload":{"bead":"aegis-abc"}}
{"ts":"2026-03-17T12:00:00Z","source":"test","type":"handoff","actor":"aegis/crew/arnold","payload":{"subject":"session 2"}}
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	events, err := ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if len(events) != 3 {
		t.Fatalf("got %d events, want 3", len(events))
	}
	// Should be reverse chronological (newest first)
	if events[0].Type != "handoff" || events[0].Actor != "aegis/crew/arnold" {
		t.Errorf("first event should be newest handoff, got type=%s actor=%s", events[0].Type, events[0].Actor)
	}
	if !events[0].Timestamp.After(events[1].Timestamp) {
		t.Error("events should be sorted newest first")
	}
}

func TestReadFile_Empty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".events.jsonl")
	if err := os.WriteFile(path, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	events, err := ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile empty: %v", err)
	}
	if len(events) != 0 {
		t.Errorf("got %d events from empty file, want 0", len(events))
	}
}

func TestReadFile_MalformedLines(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".events.jsonl")

	content := `{"ts":"2026-03-17T10:00:00Z","source":"test","type":"good","actor":"a","payload":{}}
not valid json
{"ts":"2026-03-17T11:00:00Z","source":"test","type":"also-good","actor":"b","payload":{}}
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	events, err := ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile with malformed: %v", err)
	}
	if len(events) != 2 {
		t.Errorf("got %d events, want 2 (malformed line skipped)", len(events))
	}
}

func TestReadFile_NotFound(t *testing.T) {
	_, err := ReadFile("/nonexistent/path")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestReadWorkspace(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".events.jsonl")
	content := `{"ts":"2026-03-17T10:00:00Z","source":"test","type":"handoff","actor":"a","payload":{}}
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	events, err := ReadWorkspace(dir)
	if err != nil {
		t.Fatalf("ReadWorkspace: %v", err)
	}
	if len(events) != 1 {
		t.Errorf("got %d events, want 1", len(events))
	}
}

func makeEvents() []Event {
	return []Event{
		{Timestamp: time.Date(2026, 3, 17, 12, 0, 0, 0, time.UTC), Type: "handoff", Actor: "aegis/crew/arnold", Payload: map[string]interface{}{"subject": "s3"}},
		{Timestamp: time.Date(2026, 3, 17, 11, 0, 0, 0, time.UTC), Type: "sling", Actor: "aegis/crew/goldblum", Payload: map[string]interface{}{"bead": "aegis-abc"}},
		{Timestamp: time.Date(2026, 3, 17, 10, 0, 0, 0, time.UTC), Type: "handoff", Actor: "aegis/crew/arnold", Payload: map[string]interface{}{"subject": "s2"}},
		{Timestamp: time.Date(2026, 3, 16, 15, 0, 0, 0, time.UTC), Type: "handoff", Actor: "aegis/crew/arnold", Payload: map[string]interface{}{"subject": "s1"}},
		{Timestamp: time.Date(2026, 3, 16, 10, 0, 0, 0, time.UTC), Type: "nudge", Actor: "mayor/", Payload: map[string]interface{}{"target": "aegis/crew/arnold"}},
	}
}

func TestApply_TypeFilter(t *testing.T) {
	events := makeEvents()
	result := Apply(events, Filter{Type: "handoff"})
	if len(result) != 3 {
		t.Errorf("got %d handoff events, want 3", len(result))
	}
	for _, e := range result {
		if e.Type != "handoff" {
			t.Errorf("got type %q, want handoff", e.Type)
		}
	}
}

func TestApply_ActorFilter(t *testing.T) {
	events := makeEvents()
	result := Apply(events, Filter{Actor: "goldblum"})
	if len(result) != 1 {
		t.Errorf("got %d events for goldblum, want 1", len(result))
	}
}

func TestApply_AfterFilter(t *testing.T) {
	events := makeEvents()
	after := time.Date(2026, 3, 17, 0, 0, 0, 0, time.UTC)
	result := Apply(events, Filter{After: after})
	if len(result) != 3 {
		t.Errorf("got %d events after %v, want 3", len(result), after)
	}
}

func TestApply_Limit(t *testing.T) {
	events := makeEvents()
	result := Apply(events, Filter{Limit: 2})
	if len(result) != 2 {
		t.Errorf("got %d events with limit 2, want 2", len(result))
	}
}

func TestApply_CombinedFilters(t *testing.T) {
	events := makeEvents()
	result := Apply(events, Filter{
		Type:  "handoff",
		Actor: "arnold",
		After: time.Date(2026, 3, 17, 0, 0, 0, 0, time.UTC),
		Limit: 1,
	})
	if len(result) != 1 {
		t.Fatalf("got %d events, want 1", len(result))
	}
	if result[0].Actor != "aegis/crew/arnold" {
		t.Errorf("got actor %q, want aegis/crew/arnold", result[0].Actor)
	}
}

func TestApply_NoFilter(t *testing.T) {
	events := makeEvents()
	result := Apply(events, Filter{})
	if len(result) != len(events) {
		t.Errorf("got %d events with no filter, want %d", len(result), len(events))
	}
}

func TestTypes(t *testing.T) {
	events := makeEvents()
	types := Types(events)
	if len(types) != 3 {
		t.Fatalf("got %d types, want 3", len(types))
	}
	// Types should be sorted
	for i := 1; i < len(types); i++ {
		if types[i] < types[i-1] {
			t.Errorf("types not sorted: %v", types)
			break
		}
	}
}

func TestTypes_Empty(t *testing.T) {
	types := Types(nil)
	if len(types) != 0 {
		t.Errorf("got %d types from nil, want 0", len(types))
	}
}

func TestPayloadString(t *testing.T) {
	e := Event{Payload: map[string]interface{}{
		"subject": "hello",
		"count":   42,
	}}

	if got := PayloadString(e, "subject"); got != "hello" {
		t.Errorf("PayloadString(subject) = %q, want %q", got, "hello")
	}
	if got := PayloadString(e, "missing"); got != "" {
		t.Errorf("PayloadString(missing) = %q, want empty", got)
	}
	if got := PayloadString(e, "count"); got != "" {
		t.Errorf("PayloadString(count) = %q, want empty (not a string)", got)
	}
}

func TestPayloadString_NilPayload(t *testing.T) {
	e := Event{}
	if got := PayloadString(e, "anything"); got != "" {
		t.Errorf("PayloadString with nil payload = %q, want empty", got)
	}
}

func TestBuildHandoffChains(t *testing.T) {
	events := makeEvents()
	chains := BuildHandoffChains(events)

	if len(chains) != 1 {
		t.Fatalf("got %d chains, want 1 (only arnold has handoffs)", len(chains))
	}

	chain := chains[0]
	if chain.Actor != "aegis/crew/arnold" {
		t.Errorf("chain actor = %q, want aegis/crew/arnold", chain.Actor)
	}
	if len(chain.Handoffs) != 3 {
		t.Fatalf("got %d handoffs in chain, want 3", len(chain.Handoffs))
	}

	// Should be chronological (oldest first)
	for i := 1; i < len(chain.Handoffs); i++ {
		if chain.Handoffs[i].Timestamp.Before(chain.Handoffs[i-1].Timestamp) {
			t.Error("handoffs should be chronological (oldest first)")
			break
		}
	}

	// First handoff has zero session duration
	if chain.Handoffs[0].SessionDuration != 0 {
		t.Errorf("first handoff duration = %v, want 0", chain.Handoffs[0].SessionDuration)
	}

	// Subsequent handoffs have positive session duration
	if chain.Handoffs[1].SessionDuration <= 0 {
		t.Errorf("second handoff duration = %v, want positive", chain.Handoffs[1].SessionDuration)
	}
}

func TestBuildHandoffChains_MultipleActors(t *testing.T) {
	events := []Event{
		{Timestamp: time.Date(2026, 3, 17, 12, 0, 0, 0, time.UTC), Type: "handoff", Actor: "aegis/crew/arnold"},
		{Timestamp: time.Date(2026, 3, 17, 11, 0, 0, 0, time.UTC), Type: "handoff", Actor: "aegis/crew/goldblum"},
		{Timestamp: time.Date(2026, 3, 17, 10, 0, 0, 0, time.UTC), Type: "handoff", Actor: "aegis/crew/arnold"},
		{Timestamp: time.Date(2026, 3, 16, 15, 0, 0, 0, time.UTC), Type: "handoff", Actor: "aegis/crew/goldblum"},
	}

	chains := BuildHandoffChains(events)
	if len(chains) != 2 {
		t.Fatalf("got %d chains, want 2", len(chains))
	}

	// Most recent handoff first
	if chains[0].Actor != "aegis/crew/arnold" {
		t.Errorf("first chain actor = %q, want arnold (most recent)", chains[0].Actor)
	}
}

func TestBuildHandoffChains_NoHandoffs(t *testing.T) {
	events := []Event{
		{Timestamp: time.Date(2026, 3, 17, 10, 0, 0, 0, time.UTC), Type: "sling", Actor: "a"},
		{Timestamp: time.Date(2026, 3, 17, 11, 0, 0, 0, time.UTC), Type: "nudge", Actor: "b"},
	}

	chains := BuildHandoffChains(events)
	if len(chains) != 0 {
		t.Errorf("got %d chains from non-handoff events, want 0", len(chains))
	}
}

func TestBuildHandoffChains_Empty(t *testing.T) {
	chains := BuildHandoffChains(nil)
	if len(chains) != 0 {
		t.Errorf("got %d chains from nil, want 0", len(chains))
	}
}

func TestChainSummary(t *testing.T) {
	events := makeEvents()
	chains := BuildHandoffChains(events)
	stats := ChainSummary(chains)

	if len(stats) != 1 {
		t.Fatalf("got %d stats, want 1", len(stats))
	}

	s := stats[0]
	if s.Actor != "aegis/crew/arnold" {
		t.Errorf("stats actor = %q, want aegis/crew/arnold", s.Actor)
	}
	if s.TotalHandoffs != 3 {
		t.Errorf("total handoffs = %d, want 3", s.TotalHandoffs)
	}
	if s.AvgSessionTime <= 0 {
		t.Errorf("avg session time = %v, want positive", s.AvgSessionTime)
	}
	if s.LastSubject != "s3" {
		t.Errorf("last subject = %q, want s3", s.LastSubject)
	}
}

func TestChainSummary_Empty(t *testing.T) {
	stats := ChainSummary(nil)
	if len(stats) != 0 {
		t.Errorf("got %d stats from nil, want 0", len(stats))
	}
}

func TestRigs(t *testing.T) {
	events := makeEvents()
	rigs := Rigs(events)
	if len(rigs) != 2 {
		t.Fatalf("got %d rigs, want 2 (aegis, mayor)", len(rigs))
	}
	// Should be sorted
	if rigs[0] != "aegis" || rigs[1] != "mayor" {
		t.Errorf("rigs = %v, want [aegis mayor]", rigs)
	}
}

func TestRigs_Empty(t *testing.T) {
	rigs := Rigs(nil)
	if len(rigs) != 0 {
		t.Errorf("got %d rigs from nil, want 0", len(rigs))
	}
}

func TestRigs_NoSlash(t *testing.T) {
	events := []Event{
		{Actor: "standalone"},
	}
	rigs := Rigs(events)
	if len(rigs) != 1 || rigs[0] != "standalone" {
		t.Errorf("rigs = %v, want [standalone]", rigs)
	}
}

func TestHandoffEvent_Subject(t *testing.T) {
	events := makeEvents()
	chains := BuildHandoffChains(events)
	if len(chains) == 0 {
		t.Fatal("no chains")
	}
	// Check subjects are extracted from payload
	for _, h := range chains[0].Handoffs {
		if h.Subject == "" {
			t.Error("expected non-empty subject from handoff events")
		}
	}
}
