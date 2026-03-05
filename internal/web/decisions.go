package web

import (
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
	Decisions []DecisionView
	Filter    string // pending, decided, expired, "" (all)
	Total     int
	Pending   int
	Decided   int
	Expired   int
}

func (s *Server) handleDecisions(w http.ResponseWriter, r *http.Request) {
	filter := r.URL.Query().Get("filter")
	ctx := r.Context()

	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("decisions: list dbs: %v", err)
		s.render(w, r, "decisions", decisionsData{})
		return
	}

	var allDecisions []DecisionView

	for _, db := range dbs {
		issues, err := s.ds.Decisions(ctx, db.Name)
		if err != nil {
			log.Printf("decisions %s: %v", db.Name, err)
			continue
		}

		for _, iss := range issues {
			labels, err := s.ds.LabelsForIssue(ctx, db.Name, iss.ID)
			if err != nil {
				log.Printf("labels %s/%s: %v", db.Name, iss.ID, err)
			}

			dv := parseDecisionView(iss, db.Name, labels)
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

	s.render(w, r, "decisions", decisionsData{
		Decisions: allDecisions,
		Filter:    filter,
		Total:     pending + decided + expired,
		Pending:   pending,
		Decided:   decided,
		Expired:   expired,
	})
}

// parseDecisionView extracts decision metadata from labels and description.
func parseDecisionView(issue dolt.Issue, rigName string, labels []string) DecisionView {
	dv := DecisionView{
		Issue:   issue,
		RigName: rigName,
		Labels:  labels,
		State:   "pending",
	}

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

	if dv.State == "pending" && dv.Deadline != nil && time.Now().After(*dv.Deadline) {
		dv.State = "expired"
	}

	dv.Options = parseOptions(issue.Description)

	return dv
}

var optionRe = regexp.MustCompile(`(?m)^\s*([A-Z]):\s*(.+?)(?:\s*\[RECOMMENDED\])?\s*$`)

// parseOptions extracts decision options from the description text.
func parseOptions(description string) []DecisionOption {
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
		opt.Description = strings.TrimSuffix(opt.Description, "[RECOMMENDED]")
		opt.Description = strings.TrimSpace(opt.Description)
		options = append(options, opt)
	}

	return options
}
