package web

import (
	"testing"
	"time"

	"github.com/scbrown/tapestry/internal/dolt"
)

func TestParseOptions(t *testing.T) {
	desc := `CONTEXT: We need to choose an approach for caching.
RELATED: aegis-xxx — unblocks cache implementation
OPTIONS:
  A: Redis cluster [RECOMMENDED]
  B: In-memory LRU
  C: No caching
DEFAULT: B (if no response by deadline)`

	opts := parseOptions(desc)
	if len(opts) != 3 {
		t.Fatalf("expected 3 options, got %d", len(opts))
	}

	if opts[0].Key != "A" {
		t.Errorf("expected key A, got %s", opts[0].Key)
	}
	if opts[0].Description != "Redis cluster" {
		t.Errorf("expected 'Redis cluster', got %q", opts[0].Description)
	}
	if !opts[0].Recommended {
		t.Error("expected option A to be recommended")
	}

	if opts[1].Key != "B" || opts[1].Description != "In-memory LRU" {
		t.Errorf("option B: got key=%s desc=%q", opts[1].Key, opts[1].Description)
	}
	if opts[1].Recommended {
		t.Error("option B should not be recommended")
	}

	if opts[2].Key != "C" || opts[2].Description != "No caching" {
		t.Errorf("option C: got key=%s desc=%q", opts[2].Key, opts[2].Description)
	}
}

func TestParseOptionsNoSection(t *testing.T) {
	desc := "Just a regular description with no options."
	opts := parseOptions(desc)
	if len(opts) != 0 {
		t.Fatalf("expected 0 options, got %d", len(opts))
	}
}

func TestParseDecisionView(t *testing.T) {
	issue := dolt.Issue{
		ID:        "aegis-test1",
		Title:     "Decision: Which cache?",
		Status:    "open",
		Priority:  1,
		Type:      "decision",
		UpdatedAt: time.Now(),
	}

	futureDeadline := time.Now().Add(48 * time.Hour).UTC().Format(time.RFC3339)
	labels := []string{
		"decision:pending",
		"decision:deadline:" + futureDeadline,
		"decision:default:B",
		"decision:requester:aegis/crew/goldblum",
		"decision:context-bead:aegis-xxx",
	}

	dv := parseDecisionView(issue, "beads_aegis", labels)

	if dv.State != "pending" {
		t.Errorf("expected state=pending, got %s", dv.State)
	}
	if dv.DefaultKey != "B" {
		t.Errorf("expected default=B, got %s", dv.DefaultKey)
	}
	if dv.Requester != "aegis/crew/goldblum" {
		t.Errorf("expected requester=aegis/crew/goldblum, got %s", dv.Requester)
	}
	if dv.ContextBead != "aegis-xxx" {
		t.Errorf("expected context-bead=aegis-xxx, got %s", dv.ContextBead)
	}
	if dv.Deadline == nil {
		t.Fatal("expected non-nil deadline")
	}
}

func TestExpiredState(t *testing.T) {
	issue := dolt.Issue{
		ID:        "aegis-exp1",
		Title:     "Decision: Expired one",
		Status:    "open",
		Type:      "decision",
		UpdatedAt: time.Now(),
	}

	pastTime := time.Now().Add(-24 * time.Hour).Format(time.RFC3339)
	labels := []string{
		"decision:pending",
		"decision:deadline:" + pastTime,
	}

	dv := parseDecisionView(issue, "beads_aegis", labels)

	if dv.State != "expired" {
		t.Errorf("expected state=expired for past deadline, got %s", dv.State)
	}
}

func TestDecidedState(t *testing.T) {
	issue := dolt.Issue{
		ID:        "aegis-dec1",
		Title:     "Decision: Already decided",
		Status:    "open",
		Type:      "decision",
		UpdatedAt: time.Now(),
	}

	labels := []string{
		"decision:decided",
		"decision:response:A",
		"decision:responded-by:stiwi",
		"decision:responded-via:telegram",
	}

	dv := parseDecisionView(issue, "beads_aegis", labels)

	if dv.State != "decided" {
		t.Errorf("expected state=decided, got %s", dv.State)
	}
	if dv.Response != "A" {
		t.Errorf("expected response=A, got %s", dv.Response)
	}
	if dv.RespondedBy != "stiwi" {
		t.Errorf("expected respondedBy=stiwi, got %s", dv.RespondedBy)
	}
	if dv.Channel != "telegram" {
		t.Errorf("expected channel=telegram, got %s", dv.Channel)
	}
}
