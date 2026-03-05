package web

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/scbrown/tapestry/internal/dolt"
)

type monthlyData struct {
	Year             int
	Month            time.Month
	PrevYear         int
	PrevMonth        time.Month
	NextYear         int
	NextMonth        time.Month
	HasNext          bool
	Databases        []dbSummary
	TotalOpen        int
	TotalClosed      int
	TotalCreated     int
	TotalClosedMonth int
	Agents           map[string]int
	Err              string
}

type dbSummary struct {
	Name    string
	Counts  map[string]int
	Total   int
	Created int
	Closed  int
	Recent  []dolt.Issue
	Err     string
}

type beadData struct {
	Database string
	Issue    *dolt.Issue
	Comments []dolt.Comment
	Deps     []dolt.Dependency
	Err      string
}

type beadsListData struct {
	Issues    []dolt.Issue
	Total     int
	Status    string
	Rig       string
	Type      string
	Priority  string
	Assignee  string
	Rigs      []string
	Assignees []string
	Page      int
	Pages     int
	PageLinks []pageLink
	Err       string
}

type pageLink struct {
	Num    int
	URL    string
	Active bool
}

type searchData struct {
	Query  string
	Issues []dolt.Issue
	Err    string
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	s.handleMonthly(w, r, strconv.Itoa(now.Year()), fmt.Sprintf("%02d", now.Month()))
}

func (s *Server) handleMonthly(w http.ResponseWriter, r *http.Request, yearStr, monthStr string) {
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		http.Error(w, "invalid year", http.StatusBadRequest)
		return
	}

	monthNum, err := strconv.Atoi(monthStr)
	if err != nil || monthNum < 1 || monthNum > 12 {
		http.Error(w, "invalid month", http.StatusBadRequest)
		return
	}

	m := time.Month(monthNum)
	from := time.Date(year, m, 1, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 1, 0)

	prevTime := from.AddDate(0, -1, 0)
	nextTime := to

	data := monthlyData{
		Year:      year,
		Month:     m,
		PrevYear:  prevTime.Year(),
		PrevMonth: prevTime.Month(),
		NextYear:  nextTime.Year(),
		NextMonth: nextTime.Month(),
		HasNext:   nextTime.Before(time.Now().AddDate(0, 1, 0)),
		Agents:    make(map[string]int),
	}

	if s.ds == nil {
		data.Err = "No database connection configured"
		s.render(w, r, "monthly", data)
		return
	}

	ctx := r.Context()

	dbs, err := s.databases(ctx)
	if err != nil {
		data.Err = fmt.Sprintf("Failed to list databases: %v", err)
		s.render(w, r, "monthly", data)
		return
	}

	for _, db := range dbs {
		summary := dbSummary{Name: db.Name}

		counts, err := s.ds.CountByStatus(ctx, db.Name)
		if err != nil {
			summary.Err = err.Error()
			data.Databases = append(data.Databases, summary)
			continue
		}
		summary.Counts = counts
		for _, v := range counts {
			summary.Total += v
		}

		created, err := s.ds.CountCreatedInRange(ctx, db.Name, from, to)
		if err == nil {
			summary.Created = created
		}

		closed, err := s.ds.CountClosedInRange(ctx, db.Name, from, to)
		if err == nil {
			summary.Closed = closed
		}

		data.Databases = append(data.Databases, summary)
		data.TotalOpen += counts["open"]
		data.TotalClosed += counts["closed"] + counts["completed"]
		data.TotalCreated += summary.Created
		data.TotalClosedMonth += summary.Closed

		activity, err := s.ds.AgentActivityInRange(ctx, db.Name, from, to)
		if err == nil {
			for agent, count := range activity {
				data.Agents[agent] += count
			}
		}
	}

	s.render(w, r, "monthly", data)
}

func (s *Server) handleBead(w http.ResponseWriter, r *http.Request, database, id string) {
	data := beadData{Database: database}

	if s.ds == nil {
		data.Err = "No database connection configured"
		s.render(w, r, "bead", data)
		return
	}

	ctx := r.Context()

	issue, err := s.ds.IssueByID(ctx, database, id)
	if err != nil {
		data.Err = fmt.Sprintf("Failed to load issue: %v", err)
		s.render(w, r, "bead", data)
		return
	}
	if issue == nil {
		http.NotFound(w, r)
		return
	}
	data.Issue = issue

	comments, err := s.ds.Comments(ctx, database, id)
	if err == nil {
		data.Comments = comments
	}

	deps, err := s.ds.Dependencies(ctx, database, id)
	if err == nil {
		data.Deps = deps
	}

	s.render(w, r, "bead", data)
}

// handleBeadLookup handles 2-segment /bead/{id} URLs by searching across all
// databases for the bead. This avoids the N+1 query pattern of the old code
// by trying databases in a sensible order.
func (s *Server) handleBeadLookup(w http.ResponseWriter, r *http.Request, id string) {
	if s.ds == nil {
		http.Error(w, "no database", http.StatusServiceUnavailable)
		return
	}

	ctx := r.Context()

	dbs, err := s.databases(ctx)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	for _, db := range dbs {
		issue, err := s.ds.IssueByID(ctx, db.Name, id)
		if err != nil {
			continue
		}
		if issue != nil {
			s.handleBead(w, r, db.Name, id)
			return
		}
	}

	http.NotFound(w, r)
}

// handleBeadList serves the /beads page with filterable bead listing.
func (s *Server) handleBeadList(w http.ResponseWriter, r *http.Request) {
	data := beadsListData{
		Status:   r.URL.Query().Get("status"),
		Rig:      r.URL.Query().Get("rig"),
		Type:     r.URL.Query().Get("type"),
		Priority: r.URL.Query().Get("priority"),
		Assignee: r.URL.Query().Get("assignee"),
		Page:     1,
	}

	if p := r.URL.Query().Get("page"); p != "" {
		if pn, err := strconv.Atoi(p); err == nil && pn > 0 {
			data.Page = pn
		}
	}

	if s.ds == nil {
		data.Err = "No database connection configured"
		s.render(w, r, "beads", data)
		return
	}

	ctx := r.Context()

	dbs, err := s.databases(ctx)
	if err != nil {
		data.Err = fmt.Sprintf("Failed to list databases: %v", err)
		s.render(w, r, "beads", data)
		return
	}

	filter := dolt.IssueFilter{}
	if data.Status != "" {
		filter.Status = data.Status
	}
	if data.Type != "" {
		filter.Type = data.Type
	}
	if data.Priority != "" {
		if p, err := strconv.Atoi(data.Priority); err == nil {
			filter.Priority = p
		}
	}
	if data.Assignee != "" {
		filter.Assignee = data.Assignee
	}

	rigSet := make(map[string]bool)
	assigneeSet := make(map[string]bool)

	for _, db := range dbs {
		if data.Rig != "" && db.Name != data.Rig {
			continue
		}

		issues, err := s.ds.Issues(ctx, db.Name, filter)
		if err != nil {
			log.Printf("beads list: query %s: %v", db.Name, err)
			continue
		}

		for i := range issues {
			issues[i].Rig = db.Name
		}
		data.Issues = append(data.Issues, issues...)

		rigSet[db.Name] = true
		assignees, _ := s.ds.DistinctAssignees(ctx, db.Name)
		for _, a := range assignees {
			assigneeSet[a] = true
		}
	}

	data.Total = len(data.Issues)

	// Pagination: 50 per page
	const perPage = 50
	if data.Total > perPage {
		data.Pages = (data.Total + perPage - 1) / perPage
		start := (data.Page - 1) * perPage
		end := start + perPage
		if end > data.Total {
			end = data.Total
		}
		if start < data.Total {
			data.Issues = data.Issues[start:end]
		} else {
			data.Issues = nil
		}
	} else {
		data.Pages = 1
	}

	for rig := range rigSet {
		data.Rigs = append(data.Rigs, rig)
	}
	for a := range assigneeSet {
		data.Assignees = append(data.Assignees, a)
	}

	s.render(w, r, "beads", data)
}

// handleSearch serves the /search endpoint.
func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	data := searchData{Query: q}

	if s.ds == nil || q == "" {
		s.render(w, r, "search", data)
		return
	}

	ctx := r.Context()

	dbs, err := s.databases(ctx)
	if err != nil {
		data.Err = fmt.Sprintf("Failed to list databases: %v", err)
		s.render(w, r, "search", data)
		return
	}

	for _, db := range dbs {
		results, err := s.ds.SearchIssues(ctx, db.Name, q, 20)
		if err != nil {
			continue
		}
		for i := range results {
			results[i].Rig = db.Name
		}
		data.Issues = append(data.Issues, results...)
	}

	s.render(w, r, "search", data)
}

// ── Executive Status ────────────────────────────────────────────

type statusData struct {
	GeneratedAt    time.Time
	ActiveWork     []dolt.Issue
	KeeperPipeline []keeperGroup
	BlockedWork    []blockedRow
	RecentClosed   []dolt.Issue
	ActionItems    []actionItem
}

type keeperGroup struct {
	Name       string
	Open       int
	InProgress int
}

type blockedRow struct {
	Issue        dolt.Issue
	Blocker      dolt.Issue
	BlockerOwner string
	IsHuman      bool
}

type actionItem struct {
	Issue      dolt.Issue
	Reason     string
	WaitingFor string
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	data := statusData{GeneratedAt: time.Now()}

	if s.ds == nil {
		s.render(w, r, "status", data)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()
	yesterday := time.Now().Add(-24 * time.Hour)

	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("status: list dbs: %v", err)
		s.render(w, r, "status", data)
		return
	}

	for _, db := range dbs {
		// 1. Active Initiatives: P0/P1 in_progress or open
		for _, status := range []string{"in_progress", "open"} {
			issues, err := s.ds.Issues(ctx, db.Name, dolt.IssueFilter{
				Status:   status,
				Priority: 1,
				Limit:    30,
			})
			if err != nil {
				log.Printf("status: active %s %s: %v", status, db.Name, err)
				continue
			}
			for _, iss := range issues {
				if isNoise(iss.ID, iss.Title) {
					continue
				}
				if status == "in_progress" || (status == "open" && iss.Priority <= 1) {
					data.ActiveWork = append(data.ActiveWork, iss)
				}
			}
		}

		// 2. Keeper Pipeline: group by assignee
		inFlight, err := s.ds.Issues(ctx, db.Name, dolt.IssueFilter{Limit: 200})
		if err != nil {
			log.Printf("status: pipeline %s: %v", db.Name, err)
		} else {
			keeperMap := make(map[string]*keeperGroup)
			for _, iss := range inFlight {
				if iss.Status == "closed" || isNoise(iss.ID, iss.Title) {
					continue
				}
				agent := iss.Assignee
				if agent == "" {
					agent = iss.Owner
				}
				if agent == "" {
					continue
				}
				kg, ok := keeperMap[agent]
				if !ok {
					kg = &keeperGroup{Name: agent}
					keeperMap[agent] = kg
				}
				switch iss.Status {
				case "open":
					kg.Open++
				case "in_progress", "hooked":
					kg.InProgress++
				}
			}
			for _, kg := range keeperMap {
				data.KeeperPipeline = append(data.KeeperPipeline, *kg)
			}
		}

		// 3. Blocked Work
		blocked, err := s.ds.BlockedIssues(ctx, db.Name)
		if err != nil {
			log.Printf("status: blocked %s: %v", db.Name, err)
		} else {
			for _, bi := range blocked {
				if isNoise(bi.Issue.ID, bi.Issue.Title) || isNoise(bi.Blocker.ID, bi.Blocker.Title) {
					continue
				}
				owner := bi.Blocker.Assignee
				if owner == "" {
					owner = bi.Blocker.Owner
				}
				data.BlockedWork = append(data.BlockedWork, blockedRow{
					Issue:        bi.Issue,
					Blocker:      bi.Blocker,
					BlockerOwner: owner,
					IsHuman:      isHumanOwner(owner),
				})
			}
		}

		// 4. Recent Completions (last 24h)
		recent, err := s.ds.Issues(ctx, db.Name, dolt.IssueFilter{
			Status:       "closed",
			UpdatedAfter: yesterday,
			Limit:        20,
		})
		if err != nil {
			log.Printf("status: recent %s: %v", db.Name, err)
		} else {
			for _, iss := range recent {
				if !isNoise(iss.ID, iss.Title) {
					data.RecentClosed = append(data.RecentClosed, iss)
				}
			}
		}

		// 5. Action Items: beads owned by or assigned to human
		humanIssues, err := s.ds.Issues(ctx, db.Name, dolt.IssueFilter{Limit: 100})
		if err != nil {
			log.Printf("status: action items %s: %v", db.Name, err)
		} else {
			for _, iss := range humanIssues {
				if iss.Status == "closed" || isNoise(iss.ID, iss.Title) {
					continue
				}
				reason := ""
				waitingFor := ""
				if isHumanOwner(iss.Owner) {
					reason = "Human-owned issue"
					waitingFor = iss.Owner
				} else if isHumanOwner(iss.Assignee) {
					reason = "Assigned to human"
					waitingFor = iss.Assignee
				}
				if reason != "" {
					data.ActionItems = append(data.ActionItems, actionItem{
						Issue:      iss,
						Reason:     reason,
						WaitingFor: waitingFor,
					})
				}
			}
		}
	}

	// Sort active work by priority then updated
	sort.Slice(data.ActiveWork, func(i, j int) bool {
		if data.ActiveWork[i].Priority != data.ActiveWork[j].Priority {
			return data.ActiveWork[i].Priority < data.ActiveWork[j].Priority
		}
		return data.ActiveWork[i].UpdatedAt.After(data.ActiveWork[j].UpdatedAt)
	})
	if len(data.ActiveWork) > 20 {
		data.ActiveWork = data.ActiveWork[:20]
	}

	sort.Slice(data.KeeperPipeline, func(i, j int) bool {
		return data.KeeperPipeline[i].InProgress > data.KeeperPipeline[j].InProgress
	})

	sort.Slice(data.RecentClosed, func(i, j int) bool {
		return data.RecentClosed[i].UpdatedAt.After(data.RecentClosed[j].UpdatedAt)
	})
	if len(data.RecentClosed) > 50 {
		data.RecentClosed = data.RecentClosed[:50]
	}

	sort.Slice(data.ActionItems, func(i, j int) bool {
		return data.ActionItems[i].Issue.Priority < data.ActionItems[j].Issue.Priority
	})

	s.render(w, r, "status", data)
}

// ── Briefing ────────────────────────────────────────────────

type hlaStatus struct {
	Online        bool
	Error         string
	RecentCount   int
	LastDirective time.Time
	TrackerBead   string
}

type briefingData struct {
	GeneratedAt     time.Time
	OpenCount       int
	InProgressCount int
	ClosedCount     int
	TotalBeads      int
	ClosedToday     int
	CreatedToday    int
	NeedsAttention  []dolt.Issue
	InFlight        []dolt.Issue
	RecentlyClosed  []dolt.Issue
	AgentStats      []dolt.AgentStats
	HLA             hlaStatus
}

func (s *Server) handleBriefing(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	todayEnd := todayStart.AddDate(0, 0, 1)
	yesterday := now.Add(-24 * time.Hour)

	data := briefingData{
		GeneratedAt: now,
		HLA: hlaStatus{
			Online:      false,
			Error:       "HLA archive not yet integrated — pending SSH auth fix",
			TrackerBead: "aegis-fjsnsc",
		},
	}

	if s.ds == nil {
		s.render(w, r, "briefing", data)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("briefing: list dbs: %v", err)
		s.render(w, r, "briefing", data)
		return
	}

	agentMap := make(map[string]*dolt.AgentStats)

	for _, db := range dbs {
		counts, err := s.ds.CountByStatus(ctx, db.Name)
		if err != nil {
			log.Printf("briefing: counts %s: %v", db.Name, err)
			continue
		}
		data.OpenCount += counts["open"]
		data.InProgressCount += counts["in_progress"] + counts["hooked"]
		data.ClosedCount += counts["closed"] + counts["completed"]
		for _, v := range counts {
			data.TotalBeads += v
		}

		created, err := s.ds.CountCreatedInRange(ctx, db.Name, todayStart, todayEnd)
		if err == nil {
			data.CreatedToday += created
		}
		closed, err := s.ds.CountClosedInRange(ctx, db.Name, todayStart, todayEnd)
		if err == nil {
			data.ClosedToday += closed
		}

		// Needs attention: human-owned P0/P1 open beads
		issues, err := s.ds.Issues(ctx, db.Name, dolt.IssueFilter{Limit: 100})
		if err != nil {
			log.Printf("briefing: issues %s: %v", db.Name, err)
			continue
		}
		for _, iss := range issues {
			if iss.Status == "closed" || isNoise(iss.ID, iss.Title) {
				continue
			}
			if isHumanOwner(iss.Owner) || isHumanOwner(iss.Assignee) {
				data.NeedsAttention = append(data.NeedsAttention, iss)
			}
			if (iss.Status == "in_progress" || iss.Status == "hooked") && iss.Priority <= 1 {
				data.InFlight = append(data.InFlight, iss)
			}
		}

		// Recently closed (24h)
		recent, err := s.ds.Issues(ctx, db.Name, dolt.IssueFilter{
			Status:       "closed",
			UpdatedAfter: yesterday,
			Limit:        20,
		})
		if err == nil {
			for _, iss := range recent {
				if !isNoise(iss.ID, iss.Title) {
					data.RecentlyClosed = append(data.RecentlyClosed, iss)
				}
			}
		}

		// Agent stats
		agents, err := s.ds.AgentActivity(ctx, db.Name)
		if err == nil {
			for _, a := range agents {
				if existing, ok := agentMap[a.Name]; ok {
					existing.Owned += a.Owned
					existing.Closed += a.Closed
					existing.Open += a.Open
					existing.InProgress += a.InProgress
				} else {
					copy := a
					agentMap[a.Name] = &copy
				}
			}
		}
	}

	for _, a := range agentMap {
		data.AgentStats = append(data.AgentStats, *a)
	}
	sort.Slice(data.AgentStats, func(i, j int) bool {
		return data.AgentStats[i].Owned > data.AgentStats[j].Owned
	})

	sort.Slice(data.NeedsAttention, func(i, j int) bool {
		return data.NeedsAttention[i].Priority < data.NeedsAttention[j].Priority
	})
	if len(data.NeedsAttention) > 15 {
		data.NeedsAttention = data.NeedsAttention[:15]
	}

	sort.Slice(data.InFlight, func(i, j int) bool {
		return data.InFlight[i].Priority < data.InFlight[j].Priority
	})
	if len(data.InFlight) > 20 {
		data.InFlight = data.InFlight[:20]
	}

	sort.Slice(data.RecentlyClosed, func(i, j int) bool {
		return data.RecentlyClosed[i].UpdatedAt.After(data.RecentlyClosed[j].UpdatedAt)
	})
	if len(data.RecentlyClosed) > 50 {
		data.RecentlyClosed = data.RecentlyClosed[:50]
	}

	s.render(w, r, "briefing", data)
}

// ── Agents ──────────────────────────────────────────────────

type agentsData struct {
	Agents []dolt.AgentStats
}

func (s *Server) handleAgents(w http.ResponseWriter, r *http.Request) {
	data := agentsData{}

	if s.ds == nil {
		s.render(w, r, "agents", data)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("agents: list dbs: %v", err)
		s.render(w, r, "agents", data)
		return
	}

	agentMap := make(map[string]*dolt.AgentStats)
	for _, db := range dbs {
		agents, err := s.ds.AgentActivity(ctx, db.Name)
		if err != nil {
			log.Printf("agents: activity %s: %v", db.Name, err)
			continue
		}
		for _, a := range agents {
			if a.Name == "(unowned)" || a.Name == "" {
				continue
			}
			// Merge by short name to combine duplicate identities
			key := shortActorName(a.Name)
			if existing, ok := agentMap[key]; ok {
				existing.Owned += a.Owned
				existing.Closed += a.Closed
				existing.Open += a.Open
				existing.InProgress += a.InProgress
			} else {
				merged := a
				merged.Name = key
				agentMap[key] = &merged
			}
		}
	}

	for _, a := range agentMap {
		data.Agents = append(data.Agents, *a)
	}
	sort.Slice(data.Agents, func(i, j int) bool {
		return data.Agents[i].InProgress > data.Agents[j].InProgress
	})

	s.render(w, r, "agents", data)
}

// shortActorName extracts a display name from an email or path-style identity.
func shortActorName(s string) string {
	if s == "" {
		return "—"
	}
	if idx := strings.Index(s, "@"); idx > 0 {
		return s[:idx]
	}
	parts := strings.Split(s, "/")
	return parts[len(parts)-1]
}

func isMolecule(title string) bool {
	return strings.HasPrefix(title, "mol-")
}

func isWisp(id, title string) bool {
	if strings.Contains(id, "wisp") {
		return true
	}
	lower := strings.ToLower(title)
	return strings.HasPrefix(lower, "🤝 handoff") ||
		strings.HasPrefix(lower, "handoff:") ||
		strings.HasPrefix(title, "gt-wisp-")
}

func isNoise(id, title string) bool {
	return isMolecule(title) || isWisp(id, title)
}

func isHumanOwner(owner string) bool {
	if owner == "" {
		return false
	}
	if strings.Contains(owner, "@") {
		return true
	}
	lower := strings.ToLower(owner)
	return lower == "stiwi" || lower == "braino" || strings.HasPrefix(lower, "scbrown")
}
