package web

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/scbrown/tapestry/internal/dolt"
)

// DecisionOption represents a single choice in a decision.
type DecisionOption struct {
	Key         string
	Description string
	Recommended bool
}

// DecisionView wraps an issue with parsed decision metadata.
type DecisionView struct {
	Issue       dolt.Issue
	RigName     string
	Labels      []string
	Options     []DecisionOption
	Deadline    *time.Time
	DefaultKey  string
	State       string // pending, decided, expired
	Response    string // chosen option key
	RespondedBy string
	RespondedAt string
	Channel     string
	ContextBead string
	Requester   string
}

type decisionsData struct {
	Decisions  []DecisionView
	Filter     string // pending, decided, expired, "" (all)
	Total      int
	Pending    int
	Decided    int
	Expired    int
}

func (s *Server) handleDecisions(w http.ResponseWriter, r *http.Request) {
	filter := r.URL.Query().Get("filter")
	ctx := r.Context()

	var allDecisions []DecisionView

	for _, dbName := range s.databases() {
		issues, err := s.client.Decisions(ctx, dbName)
		if err != nil {
			log.Printf("decisions %s: %v", dbName, err)
			continue
		}

		for _, iss := range issues {
			labels, err := s.client.LabelsForIssue(ctx, dbName, iss.ID)
			if err != nil {
				log.Printf("labels %s/%s: %v", dbName, iss.ID, err)
			}

			dv := parseDecisionView(iss, dbName, labels)
			allDecisions = append(allDecisions, dv)
		}
	}

	// Sort by deadline (most urgent first), then by updated_at
	sort.Slice(allDecisions, func(i, j int) bool {
		di, dj := allDecisions[i], allDecisions[j]
		if di.Deadline != nil && dj.Deadline != nil {
			return di.Deadline.Before(*dj.Deadline)
		}
		if di.Deadline != nil {
			return true
		}
		if dj.Deadline != nil {
			return false
		}
		return di.Issue.UpdatedAt.After(dj.Issue.UpdatedAt)
	})

	// Count by state
	var pending, decided, expired int
	for _, d := range allDecisions {
		switch d.State {
		case "pending":
			pending++
		case "decided":
			decided++
		case "expired":
			expired++
		}
	}

	// Apply filter
	if filter != "" {
		var filtered []DecisionView
		for _, d := range allDecisions {
			if d.State == filter {
				filtered = append(filtered, d)
			}
		}
		allDecisions = filtered
	}

	s.render(w, "decisions.html", decisionsData{
		Decisions: allDecisions,
		Filter:    filter,
		Total:     pending + decided + expired,
		Pending:   pending,
		Decided:   decided,
		Expired:   expired,
	})
}

func (s *Server) handleDecisionRespond(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "missing decision id", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form data", http.StatusBadRequest)
		return
	}
	choice := r.FormValue("choice")
	if choice == "" {
		http.Error(w, "missing choice", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// Find the decision across databases
	for _, dbName := range s.databases() {
		issue, err := s.client.IssueByID(ctx, dbName, id)
		if err != nil || issue == nil {
			continue
		}

		// Record the response via labels and comment
		now := time.Now().UTC()
		responseLabels := []string{
			fmt.Sprintf("decision:response:%s", choice),
			"decision:decided",
			fmt.Sprintf("decision:responded-via:tapestry"),
			fmt.Sprintf("decision:responded-at:%s", now.Format(time.RFC3339)),
		}
		for _, label := range responseLabels {
			if err := s.client.AddLabel(ctx, dbName, id, label); err != nil {
				log.Printf("add label %s/%s %s: %v", dbName, id, label, err)
			}
		}
		// Remove pending label
		if err := s.client.RemoveLabel(ctx, dbName, id, "decision:pending"); err != nil {
			log.Printf("remove pending label %s/%s: %v", dbName, id, err)
		}

		// Add a comment recording the decision
		comment := fmt.Sprintf("Decision recorded via Tapestry: **%s**\nResponded at: %s",
			choice, now.Format("2006-01-02 15:04"))
		if err := s.client.AddComment(ctx, dbName, id, "tapestry", comment); err != nil {
			log.Printf("add comment %s/%s: %v", dbName, id, err)
		}

		// Return HTMX partial: confirmation card
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, `<div class="decision-card decided" id="decision-%s">
  <div class="decision-header">
    <h3><a href="/bead/%s">%s</a></h3>
    <span class="decision-state decided">decided</span>
  </div>
  <p class="decision-confirmed">Recorded: <strong>%s</strong></p>
  <p class="text-dim">via Tapestry at %s</p>
</div>`, id, id, issue.Title, choice, now.Format("15:04"))
		return
	}

	http.NotFound(w, r)
}

// parseDecisionView extracts decision metadata from labels and description.
func parseDecisionView(issue dolt.Issue, rigName string, labels []string) DecisionView {
	dv := DecisionView{
		Issue:   issue,
		RigName: rigName,
		Labels:  labels,
		State:   "pending", // default
	}

	// Parse labels
	for _, l := range labels {
		switch {
		case l == "decision:pending":
			dv.State = "pending"
		case l == "decision:decided":
			dv.State = "decided"
		case strings.HasPrefix(l, "decision:response:"):
			dv.Response = strings.TrimPrefix(l, "decision:response:")
		case strings.HasPrefix(l, "decision:responded-by:"):
			dv.RespondedBy = strings.TrimPrefix(l, "decision:responded-by:")
		case strings.HasPrefix(l, "decision:responded-via:"):
			dv.Channel = strings.TrimPrefix(l, "decision:responded-via:")
		case strings.HasPrefix(l, "decision:responded-at:"):
			dv.RespondedAt = strings.TrimPrefix(l, "decision:responded-at:")
		case strings.HasPrefix(l, "decision:deadline:"):
			ts := strings.TrimPrefix(l, "decision:deadline:")
			if t, err := time.Parse(time.RFC3339, ts); err == nil {
				dv.Deadline = &t
			}
		case strings.HasPrefix(l, "decision:default:"):
			dv.DefaultKey = strings.TrimPrefix(l, "decision:default:")
		case strings.HasPrefix(l, "decision:context-bead:"):
			dv.ContextBead = strings.TrimPrefix(l, "decision:context-bead:")
		case strings.HasPrefix(l, "decision:requester:"):
			dv.Requester = strings.TrimPrefix(l, "decision:requester:")
		}
	}

	// Check if expired
	if dv.State == "pending" && dv.Deadline != nil && time.Now().After(*dv.Deadline) {
		dv.State = "expired"
	}

	// Also use due_at from issue if no deadline label
	if dv.Deadline == nil && !issue.UpdatedAt.IsZero() {
		// No explicit deadline — leave nil
	}

	// Parse options from description
	dv.Options = parseOptions(issue.Description)

	return dv
}

var optionRe = regexp.MustCompile(`(?m)^\s*([A-Z]):\s*(.+?)(?:\s*\[RECOMMENDED\])?\s*$`)

// parseOptions extracts decision options from the description text.
func parseOptions(description string) []DecisionOption {
	// Look for OPTIONS: section
	idx := strings.Index(strings.ToUpper(description), "OPTIONS:")
	if idx < 0 {
		return nil
	}
	section := description[idx:]

	var options []DecisionOption
	for _, match := range optionRe.FindAllStringSubmatch(section, -1) {
		opt := DecisionOption{
			Key:         match[1],
			Description: strings.TrimSpace(match[2]),
			Recommended: strings.Contains(match[0], "[RECOMMENDED]"),
		}
		// Clean [RECOMMENDED] from description
		opt.Description = strings.TrimSuffix(opt.Description, "[RECOMMENDED]")
		opt.Description = strings.TrimSpace(opt.Description)
		options = append(options, opt)
	}

	return options
}
