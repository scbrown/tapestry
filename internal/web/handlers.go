package web

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/scbrown/tapestry/internal/dolt"
	"github.com/scbrown/tapestry/internal/events"
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
	Database      string
	Issue         *dolt.Issue
	Labels        []string
	Comments      []dolt.Comment
	Deps          []dolt.Dependency
	Commits       []beadCommit
	Metadata      *dolt.IssueMetadata
	StatusHistory []dolt.StatusTransition
	Children      []dolt.Issue
	Assignees     []string          // known assignees for reassign dropdown
	DispatchInfo  map[string]string // parsed key-value metadata from description prefix
	CleanDesc     string            // description with metadata prefix stripped
	Err           string
}

type beadsListData struct {
	Issues    []dolt.Issue
	Total     int
	Status    string
	Rig       string
	Type      string
	Priority  string
	Assignee  string
	Sort      string
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

	summaries := make([]dbSummary, len(dbs))
	activities := make([]map[string]int, len(dbs))
	var wg sync.WaitGroup
	for i, db := range dbs {
		wg.Add(1)
		go func(i int, dbName string) {
			defer wg.Done()
			summary := dbSummary{Name: dbName}

			counts, err := s.ds.CountByStatus(ctx, dbName)
			if err != nil {
				summary.Err = err.Error()
				summaries[i] = summary
				return
			}
			summary.Counts = counts
			for _, v := range counts {
				summary.Total += v
			}

			created, err := s.ds.CountCreatedInRange(ctx, dbName, from, to)
			if err == nil {
				summary.Created = created
			}

			closed, err := s.ds.CountClosedInRange(ctx, dbName, from, to)
			if err == nil {
				summary.Closed = closed
			}

			summaries[i] = summary

			activity, err := s.ds.AgentActivityInRange(ctx, dbName, from, to)
			if err == nil {
				activities[i] = activity
			}
		}(i, db.Name)
	}
	wg.Wait()

	for i := range summaries {
		data.Databases = append(data.Databases, summaries[i])
		if summaries[i].Counts != nil {
			data.TotalOpen += summaries[i].Counts["open"]
			data.TotalClosed += summaries[i].Counts["closed"] + summaries[i].Counts["completed"]
		}
		data.TotalCreated += summaries[i].Created
		data.TotalClosedMonth += summaries[i].Closed
		for agent, count := range activities[i] {
			data.Agents[agent] += count
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

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

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
	data.DispatchInfo, data.CleanDesc = parseDescriptionMetadata(issue.Description)

	labels, err := s.ds.LabelsForIssue(ctx, database, id)
	if err == nil {
		data.Labels = labels
	}

	comments, err := s.ds.Comments(ctx, database, id)
	if err == nil {
		data.Comments = comments
	}

	deps, err := s.ds.Dependencies(ctx, database, id)
	if err == nil {
		data.Deps = deps
	}

	meta, err := s.ds.MetadataForIssue(ctx, database, id)
	if err == nil {
		data.Metadata = meta
	}

	// StatusHistory queries dolt_history_issues which scans the full commit
	// graph — expensive on databases with many commits (e.g. aegis has 23k+).
	// Bound with a short timeout so it degrades gracefully.
	histCtx, histCancel := context.WithTimeout(ctx, 2*time.Second)
	history, err := s.ds.StatusHistory(histCtx, database, id)
	histCancel()
	if err == nil {
		data.StatusHistory = history
	}

	if issue.Type == "epic" {
		children, err := s.ds.ChildIssues(ctx, database, id)
		if err == nil {
			data.Children = children
		}
	}

	if s.forgejo != nil {
		commitCtx, commitCancel := context.WithTimeout(ctx, 2*time.Second)
		data.Commits = s.forgejo.searchCommitsForBead(commitCtx, id)
		commitCancel()
	}

	assignees, _ := s.ds.DistinctAssignees(ctx, database)
	data.Assignees = assignees

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

	// Try to match the bead ID prefix to a database name for fast lookup.
	// e.g. "aegis-abc" → try "beads_aegis" first, "hq-abc" → try "beads_gt" first.
	prefix := id
	if idx := strings.IndexByte(id, '-'); idx > 0 {
		prefix = id[:idx]
	}
	guessDB := "beads_" + prefix
	if prefix == "hq" {
		guessDB = "beads_gt"
	}

	// Reorder: put the guessed database first to avoid scanning all databases.
	ordered := make([]dolt.DatabaseInfo, 0, len(dbs))
	for _, db := range dbs {
		if db.Name == guessDB {
			ordered = append([]dolt.DatabaseInfo{db}, ordered...)
		} else {
			ordered = append(ordered, db)
		}
	}

	for _, db := range ordered {
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
		Sort:     r.URL.Query().Get("sort"),
		Page:     1,
	}
	if data.Sort == "" {
		data.Sort = "updated"
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

	// Filter databases by rig if specified
	var targetDBs []dolt.DatabaseInfo
	for _, db := range dbs {
		if data.Rig != "" && db.Name != data.Rig {
			continue
		}
		targetDBs = append(targetDBs, db)
	}

	type beadListResult struct {
		issues    []dolt.Issue
		rigName   string
		assignees []string
	}

	beadResults := make([]beadListResult, len(targetDBs))
	var wg sync.WaitGroup
	for i, db := range targetDBs {
		wg.Add(1)
		go func(i int, dbName string) {
			defer wg.Done()
			r := beadListResult{rigName: dbName}

			issues, err := s.ds.Issues(ctx, dbName, filter)
			if err != nil {
				log.Printf("beads list: query %s: %v", dbName, err)
				beadResults[i] = r
				return
			}

			for j := range issues {
				issues[j].Rig = dbName
			}
			r.issues = issues
			r.assignees, _ = s.ds.DistinctAssignees(ctx, dbName)
			beadResults[i] = r
		}(i, db.Name)
	}
	wg.Wait()

	rigSet := make(map[string]bool)
	assigneeSet := make(map[string]bool)
	for _, r := range beadResults {
		data.Issues = append(data.Issues, r.issues...)
		rigSet[r.rigName] = true
		for _, a := range r.assignees {
			assigneeSet[a] = true
		}
	}

	// Sort before pagination
	switch data.Sort {
	case "priority":
		sort.Slice(data.Issues, func(i, j int) bool {
			if data.Issues[i].Priority != data.Issues[j].Priority {
				return data.Issues[i].Priority < data.Issues[j].Priority
			}
			return data.Issues[i].UpdatedAt.After(data.Issues[j].UpdatedAt)
		})
	case "oldest":
		sort.Slice(data.Issues, func(i, j int) bool {
			return data.Issues[i].UpdatedAt.Before(data.Issues[j].UpdatedAt)
		})
	case "created":
		sort.Slice(data.Issues, func(i, j int) bool {
			return data.Issues[i].CreatedAt.After(data.Issues[j].CreatedAt)
		})
	default: // "updated" — newest first
		sort.Slice(data.Issues, func(i, j int) bool {
			return data.Issues[i].UpdatedAt.After(data.Issues[j].UpdatedAt)
		})
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

	// Build page links
	if data.Pages > 1 {
		baseParams := r.URL.Query()
		for p := 1; p <= data.Pages; p++ {
			baseParams.Set("page", strconv.Itoa(p))
			data.PageLinks = append(data.PageLinks, pageLink{
				Num:    p,
				URL:    "/beads?" + baseParams.Encode(),
				Active: p == data.Page,
			})
		}
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

	type searchResult struct {
		issues []dolt.Issue
	}
	results := make([]searchResult, len(dbs))
	var wg sync.WaitGroup
	for i, db := range dbs {
		wg.Add(1)
		go func(i int, dbName string) {
			defer wg.Done()
			issues, err := s.ds.SearchIssues(ctx, dbName, q, 20)
			if err != nil {
				return
			}
			for j := range issues {
				issues[j].Rig = dbName
			}
			results[i] = searchResult{issues: issues}
		}(i, db.Name)
	}
	wg.Wait()

	for _, r := range results {
		data.Issues = append(data.Issues, r.issues...)
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

	type statusDBResult struct {
		activeWork      []dolt.Issue
		keeperPipeline  []keeperGroup
		blockedWork     []blockedRow
		recentClosed    []dolt.Issue
		actionItems     []actionItem
	}

	statusResults := make([]statusDBResult, len(dbs))
	var wg sync.WaitGroup
	for i, db := range dbs {
		wg.Add(1)
		go func(i int, dbName string) {
			defer wg.Done()
			var r statusDBResult

			// 1. Active Initiatives: P0/P1 in_progress or open
			for _, status := range []string{"in_progress", "open"} {
				issues, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{
					Status:   status,
					Priority: 1,
					Limit:    30,
				})
				if err != nil {
					log.Printf("status: active %s %s: %v", status, dbName, err)
					continue
				}
				for _, iss := range issues {
					if isNoise(iss.ID, iss.Title) {
						continue
					}
					iss.Rig = dbName
					if status == "in_progress" || (status == "open" && iss.Priority <= 1) {
						r.activeWork = append(r.activeWork, iss)
					}
				}
			}

			// 2. Keeper Pipeline: group by assignee
			inFlight, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{Limit: 200})
			if err != nil {
				log.Printf("status: pipeline %s: %v", dbName, err)
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
					r.keeperPipeline = append(r.keeperPipeline, *kg)
				}
			}

			// 3. Blocked Work
			blocked, err := s.ds.BlockedIssues(ctx, dbName)
			if err != nil {
				log.Printf("status: blocked %s: %v", dbName, err)
			} else {
				for _, bi := range blocked {
					if isNoise(bi.Issue.ID, bi.Issue.Title) || isNoise(bi.Blocker.ID, bi.Blocker.Title) {
						continue
					}
					owner := bi.Blocker.Assignee
					if owner == "" {
						owner = bi.Blocker.Owner
					}
					bi.Issue.Rig = dbName
					bi.Blocker.Rig = dbName
					r.blockedWork = append(r.blockedWork, blockedRow{
						Issue:        bi.Issue,
						Blocker:      bi.Blocker,
						BlockerOwner: owner,
						IsHuman:      isHumanOwner(owner),
					})
				}
			}

			// 4. Recent Completions (last 24h)
			recent, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{
				Status:       "closed",
				UpdatedAfter: yesterday,
				Limit:        20,
			})
			if err != nil {
				log.Printf("status: recent %s: %v", dbName, err)
			} else {
				for _, iss := range recent {
					if !isNoise(iss.ID, iss.Title) {
						iss.Rig = dbName
						r.recentClosed = append(r.recentClosed, iss)
					}
				}
			}

			// 5. Action Items: beads owned by or assigned to human
			humanIssues, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{Limit: 100})
			if err != nil {
				log.Printf("status: action items %s: %v", dbName, err)
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
						iss.Rig = dbName
						r.actionItems = append(r.actionItems, actionItem{
							Issue:      iss,
							Reason:     reason,
							WaitingFor: waitingFor,
						})
					}
				}
			}

			statusResults[i] = r
		}(i, db.Name)
	}
	wg.Wait()

	for _, r := range statusResults {
		data.ActiveWork = append(data.ActiveWork, r.activeWork...)
		data.KeeperPipeline = append(data.KeeperPipeline, r.keeperPipeline...)
		data.BlockedWork = append(data.BlockedWork, r.blockedWork...)
		data.RecentClosed = append(data.RecentClosed, r.recentClosed...)
		data.ActionItems = append(data.ActionItems, r.actionItems...)
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

type briefingBlockedItem struct {
	Issue       dolt.Issue
	BlockerID   string
	BlockerDesc string
	BlockerRig  string
	Owner       string
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
	UnclaimedQueue  []dolt.Issue
	BlockedItems    []briefingBlockedItem
	RecentlyClosed  []dolt.Issue
	AgentStats      []dolt.AgentStats
	StaleWork       []dolt.Issue
	HLA             hlaStatus
	FreshAttention  time.Time
	FreshInFlight   time.Time
	FreshClosed     time.Time
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

	type briefingDBResult struct {
		openCount       int
		inProgressCount int
		closedCount     int
		totalBeads      int
		createdToday    int
		closedToday     int
		needsAttention  []dolt.Issue
		inFlight        []dolt.Issue
		unclaimedQueue  []dolt.Issue
		blockedItems    []briefingBlockedItem
		recentlyClosed  []dolt.Issue
		staleWork       []dolt.Issue
		agents          []dolt.AgentStats
	}

	results := make([]briefingDBResult, len(dbs))
	var wg sync.WaitGroup
	for i, db := range dbs {
		wg.Add(1)
		go func(i int, dbName string) {
			defer wg.Done()
			var r briefingDBResult

			counts, err := s.ds.CountByStatus(ctx, dbName)
			if err != nil {
				log.Printf("briefing: counts %s: %v", dbName, err)
				results[i] = r
				return
			}
			r.openCount = counts["open"]
			r.inProgressCount = counts["in_progress"] + counts["hooked"]
			r.closedCount = counts["closed"] + counts["completed"]
			for _, v := range counts {
				r.totalBeads += v
			}

			created, err := s.ds.CountCreatedInRange(ctx, dbName, todayStart, todayEnd)
			if err == nil {
				r.createdToday = created
			}
			closed, err := s.ds.CountClosedInRange(ctx, dbName, todayStart, todayEnd)
			if err == nil {
				r.closedToday = closed
			}

			issues, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{Limit: 100})
			if err != nil {
				log.Printf("briefing: issues %s: %v", dbName, err)
				results[i] = r
				return
			}
			staleThreshold := now.AddDate(0, 0, -7)
			for j := range issues {
				issues[j].Rig = dbName
			}
			for _, iss := range issues {
				if iss.Status == "closed" || isNoise(iss.ID, iss.Title) {
					continue
				}
				if isHumanOwner(iss.Owner) || isHumanOwner(iss.Assignee) {
					r.needsAttention = append(r.needsAttention, iss)
				}
				if (iss.Status == "in_progress" || iss.Status == "hooked") && iss.Priority <= 1 {
					r.inFlight = append(r.inFlight, iss)
				}
				if iss.Status == "open" && iss.Priority <= 2 && iss.Assignee == "" {
					r.unclaimedQueue = append(r.unclaimedQueue, iss)
				}
				if iss.Priority <= 2 && iss.UpdatedAt.Before(staleThreshold) &&
					(iss.Status == "open" || iss.Status == "in_progress") {
					r.staleWork = append(r.staleWork, iss)
				}
			}

			// Blocked items (P1/P2 non-closed items with unresolved blockers)
			blocked, err := s.ds.BlockedIssues(ctx, dbName)
			if err == nil {
				for _, bi := range blocked {
					if isNoise(bi.Issue.ID, bi.Issue.Title) || bi.Issue.Priority > 2 {
						continue
					}
					owner := bi.Blocker.Assignee
					if owner == "" {
						owner = bi.Blocker.Owner
					}
					r.blockedItems = append(r.blockedItems, briefingBlockedItem{
						Issue:       bi.Issue,
						BlockerID:   bi.Blocker.ID,
						BlockerDesc: bi.Blocker.Title,
						BlockerRig:  dbName,
						Owner:       owner,
					})
				}
			}

			recent, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{
				Status:       "closed",
				UpdatedAfter: yesterday,
				Limit:        20,
			})
			if err == nil {
				for _, iss := range recent {
					if !isNoise(iss.ID, iss.Title) {
						iss.Rig = dbName
						r.recentlyClosed = append(r.recentlyClosed, iss)
					}
				}
			}

			r.agents, _ = s.ds.AgentActivity(ctx, dbName)
			results[i] = r
		}(i, db.Name)
	}
	wg.Wait()

	agentMap := make(map[string]*dolt.AgentStats)
	for _, r := range results {
		data.OpenCount += r.openCount
		data.InProgressCount += r.inProgressCount
		data.ClosedCount += r.closedCount
		data.TotalBeads += r.totalBeads
		data.CreatedToday += r.createdToday
		data.ClosedToday += r.closedToday
		data.NeedsAttention = append(data.NeedsAttention, r.needsAttention...)
		data.InFlight = append(data.InFlight, r.inFlight...)
		data.UnclaimedQueue = append(data.UnclaimedQueue, r.unclaimedQueue...)
		data.BlockedItems = append(data.BlockedItems, r.blockedItems...)
		data.RecentlyClosed = append(data.RecentlyClosed, r.recentlyClosed...)
		data.StaleWork = append(data.StaleWork, r.staleWork...)
		for _, a := range r.agents {
			if existing, ok := agentMap[a.Name]; ok {
				existing.Owned += a.Owned
				existing.Closed += a.Closed
				existing.Open += a.Open
				existing.InProgress += a.InProgress
			} else {
				cp := a
				agentMap[a.Name] = &cp
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
	for _, iss := range data.NeedsAttention {
		if iss.UpdatedAt.After(data.FreshAttention) {
			data.FreshAttention = iss.UpdatedAt
		}
	}

	sort.Slice(data.UnclaimedQueue, func(i, j int) bool {
		return data.UnclaimedQueue[i].Priority < data.UnclaimedQueue[j].Priority
	})
	if len(data.UnclaimedQueue) > 15 {
		data.UnclaimedQueue = data.UnclaimedQueue[:15]
	}

	sort.Slice(data.BlockedItems, func(i, j int) bool {
		return data.BlockedItems[i].Issue.Priority < data.BlockedItems[j].Issue.Priority
	})
	if len(data.BlockedItems) > 15 {
		data.BlockedItems = data.BlockedItems[:15]
	}

	sort.Slice(data.InFlight, func(i, j int) bool {
		return data.InFlight[i].Priority < data.InFlight[j].Priority
	})
	if len(data.InFlight) > 20 {
		data.InFlight = data.InFlight[:20]
	}
	for _, iss := range data.InFlight {
		if iss.UpdatedAt.After(data.FreshInFlight) {
			data.FreshInFlight = iss.UpdatedAt
		}
	}

	sort.Slice(data.RecentlyClosed, func(i, j int) bool {
		return data.RecentlyClosed[i].UpdatedAt.After(data.RecentlyClosed[j].UpdatedAt)
	})
	if len(data.RecentlyClosed) > 50 {
		data.RecentlyClosed = data.RecentlyClosed[:50]
	}
	for _, iss := range data.RecentlyClosed {
		if iss.UpdatedAt.After(data.FreshClosed) {
			data.FreshClosed = iss.UpdatedAt
		}
	}

	// Sort stale by oldest update first (most stale at top)
	sort.Slice(data.StaleWork, func(i, j int) bool {
		return data.StaleWork[i].UpdatedAt.Before(data.StaleWork[j].UpdatedAt)
	})
	if len(data.StaleWork) > 10 {
		data.StaleWork = data.StaleWork[:10]
	}

	s.render(w, r, "briefing", data)
}

// ── Agents ──────────────────────────────────────────────────

type agentRow struct {
	dolt.AgentStats
	LastActive    time.Time
	TotalHandoffs int
}

type agentsData struct {
	Agents []agentRow
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

	agentResults := make([][]dolt.AgentStats, len(dbs))
	var wgAgents sync.WaitGroup
	for i, db := range dbs {
		wgAgents.Add(1)
		go func(i int, dbName string) {
			defer wgAgents.Done()
			agents, err := s.ds.AgentActivity(ctx, dbName)
			if err != nil {
				log.Printf("agents: activity %s: %v", dbName, err)
				return
			}
			agentResults[i] = agents
		}(i, db.Name)
	}
	wgAgents.Wait()

	agentMap := make(map[string]*agentRow)
	for _, agents := range agentResults {
		for _, a := range agents {
			if a.Name == "(unowned)" || a.Name == "" {
				continue
			}
			key := shortActorName(a.Name)
			if existing, ok := agentMap[key]; ok {
				existing.Owned += a.Owned
				existing.Closed += a.Closed
				existing.Open += a.Open
				existing.InProgress += a.InProgress
			} else {
				merged := a
				merged.Name = key
				agentMap[key] = &agentRow{AgentStats: merged}
			}
		}
	}

	// Enrich with handoff data
	if s.workspacePath != "" {
		allEvents, err := events.ReadWorkspace(s.workspacePath)
		if err == nil {
			chains := events.BuildHandoffChains(allEvents)
			summary := events.ChainSummary(chains)
			for _, cs := range summary {
				key := shortActorName(cs.Actor)
				if row, ok := agentMap[key]; ok {
					row.LastActive = cs.LastHandoff
					row.TotalHandoffs = cs.TotalHandoffs
				}
			}
		}
	}

	for _, a := range agentMap {
		data.Agents = append(data.Agents, *a)
	}
	sort.Slice(data.Agents, func(i, j int) bool {
		if data.Agents[i].InProgress != data.Agents[j].InProgress {
			return data.Agents[i].InProgress > data.Agents[j].InProgress
		}
		return data.Agents[i].LastActive.After(data.Agents[j].LastActive)
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

// handleBeadStatusUpdate handles POST /bead/{db}/{id}/status to update a bead's status.
// Returns an HTMX partial with the updated status badge and action buttons.
func (s *Server) handleBeadStatusUpdate(w http.ResponseWriter, r *http.Request, database, id string) {
	if s.ds == nil {
		http.Error(w, "no database", http.StatusServiceUnavailable)
		return
	}

	newStatus := r.FormValue("status")
	validStatuses := map[string]bool{
		"open": true, "in_progress": true, "closed": true, "blocked": true, "deferred": true,
	}
	if !validStatuses[newStatus] {
		http.Error(w, "invalid status", http.StatusBadRequest)
		return
	}

	if err := s.ds.UpdateStatus(r.Context(), database, id, newStatus); err != nil {
		log.Printf("bead status update: %v", err)
		http.Error(w, "update failed", http.StatusInternalServerError)
		return
	}

	// Return updated action bar as HTMX partial
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("HX-Trigger", fmt.Sprintf(`{"showToast":"Bead %s"}`, newStatus))
	fmt.Fprintf(w, `<div class="bead-actions" id="bead-actions">
		<span class="badge %s">%s</span>`, statusClassName(newStatus), newStatus)
	switch newStatus {
	case "open", "in_progress", "blocked", "deferred":
		fmt.Fprintf(w, `
		<button hx-post="/bead/%s/%s/status" hx-vals='{"status":"closed"}' hx-target="#bead-actions" hx-swap="outerHTML" class="btn btn-sm btn-close">Close</button>`, database, id)
	case "closed":
		fmt.Fprintf(w, `
		<button hx-post="/bead/%s/%s/status" hx-vals='{"status":"open"}' hx-target="#bead-actions" hx-swap="outerHTML" class="btn btn-sm btn-reopen">Reopen</button>`, database, id)
	}
	fmt.Fprint(w, `</div>`)
}

func (s *Server) handleBeadComment(w http.ResponseWriter, r *http.Request, database, id string) {
	if s.ds == nil {
		http.Error(w, "no database", http.StatusServiceUnavailable)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	body := strings.TrimSpace(r.FormValue("body"))
	if body == "" {
		http.Error(w, "comment body required", http.StatusBadRequest)
		return
	}

	// Sanitize: limit body length
	if len(body) > 4000 {
		body = body[:4000]
	}

	author := "tapestry-web"

	if err := s.ds.AddComment(r.Context(), database, id, author, body); err != nil {
		log.Printf("bead comment: %v", err)
		http.Error(w, "failed to add comment", http.StatusInternalServerError)
		return
	}

	// Return the new comment as HTMX partial, plus clear the form
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	now := time.Now()
	fmt.Fprintf(w, `<div class="comment">
		<div class="comment-header">
			<strong>%s</strong>
			<time>%s</time>
		</div>
		<pre class="comment-body">%s</pre>
	</div>`,
		template.HTMLEscapeString(author),
		template.HTMLEscapeString(now.Format("Jan 2, 2006 15:04")),
		template.HTMLEscapeString(body),
	)
}

func (s *Server) handleBeadPriorityUpdate(w http.ResponseWriter, r *http.Request, database, id string) {
	if s.ds == nil {
		http.Error(w, "no database", http.StatusServiceUnavailable)
		return
	}

	pStr := r.FormValue("priority")
	priority := -1
	switch pStr {
	case "0":
		priority = 0
	case "1":
		priority = 1
	case "2":
		priority = 2
	case "3":
		priority = 3
	case "4":
		priority = 4
	default:
		http.Error(w, "invalid priority (0-4)", http.StatusBadRequest)
		return
	}

	if err := s.ds.UpdatePriority(r.Context(), database, id, priority); err != nil {
		log.Printf("bead priority update: %v", err)
		http.Error(w, "update failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("HX-Trigger", fmt.Sprintf(`{"showToast":"Priority set to P%d"}`, priority))
	fmt.Fprintf(w, `<span class="priority-badge p%d">P%d</span>`, priority, priority)
}

func (s *Server) handleBeadAssigneeUpdate(w http.ResponseWriter, r *http.Request, database, id string) {
	if s.ds == nil {
		http.Error(w, "no database", http.StatusServiceUnavailable)
		return
	}

	assignee := strings.TrimSpace(r.FormValue("assignee"))
	if len(assignee) > 200 {
		http.Error(w, "assignee too long", http.StatusBadRequest)
		return
	}

	if err := s.ds.UpdateAssignee(r.Context(), database, id, assignee); err != nil {
		log.Printf("bead assignee update: %v", err)
		http.Error(w, "update failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if assignee != "" {
		w.Header().Set("HX-Trigger", fmt.Sprintf(`{"showToast":"Assigned to %s"}`, template.HTMLEscapeString(assignee)))
	}
	if assignee == "" {
		fmt.Fprint(w, `<span class="text-dim">unassigned</span>`)
	} else {
		short := assignee
		if idx := strings.Index(short, "@"); idx > 0 {
			short = short[:idx]
		} else if parts := strings.Split(short, "/"); len(parts) > 1 {
			short = parts[len(parts)-1]
		}
		fmt.Fprintf(w, `<span>%s</span>`, template.HTMLEscapeString(short))
	}
}

func (s *Server) handleBeadLabelAdd(w http.ResponseWriter, r *http.Request, database, id string) {
	if s.ds == nil {
		http.Error(w, "no database", http.StatusServiceUnavailable)
		return
	}

	label := strings.TrimSpace(r.FormValue("label"))
	if label == "" {
		http.Error(w, "label required", http.StatusBadRequest)
		return
	}
	// Sanitize: alphanumeric, dashes, underscores only, max 50 chars
	if len(label) > 50 {
		label = label[:50]
	}
	for _, c := range label {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_') {
			http.Error(w, "label must be alphanumeric with dashes/underscores", http.StatusBadRequest)
			return
		}
	}

	if err := s.ds.AddLabel(r.Context(), database, id, label); err != nil {
		log.Printf("bead label add: %v", err)
		http.Error(w, "failed to add label", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("HX-Trigger", fmt.Sprintf(`{"showToast":"Label '%s' added"}`, template.HTMLEscapeString(label)))
	fmt.Fprintf(w, `<a href="/labels?label=%s" class="badge label">%s</a> `,
		template.HTMLEscapeString(label), template.HTMLEscapeString(label))
}

func (s *Server) handleBeadTitleUpdate(w http.ResponseWriter, r *http.Request, database, id string) {
	if s.ds == nil {
		http.Error(w, "no database", http.StatusServiceUnavailable)
		return
	}

	title := strings.TrimSpace(r.FormValue("title"))
	if title == "" {
		http.Error(w, "title required", http.StatusBadRequest)
		return
	}
	if len(title) > 500 {
		title = title[:500]
	}

	if err := s.ds.UpdateTitle(r.Context(), database, id, title); err != nil {
		log.Printf("bead title update: %v", err)
		http.Error(w, "update failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("HX-Trigger", `{"showToast":"Title updated"}`)
	fmt.Fprintf(w, `<span class="bead-title-text">%s</span>
<button class="btn btn-xs bead-title-edit" onclick="this.parentElement.querySelector('.bead-title-form').style.display='inline-flex'; this.parentElement.querySelector('.bead-title-text').style.display='none'; this.style.display='none'; this.parentElement.querySelector('.bead-title-form input').focus();">edit</button>
<form class="bead-title-form" style="display:none" hx-post="/bead/%s/%s/title" hx-target="#bead-title" hx-swap="innerHTML">
  <input type="text" name="title" value="%s" class="bead-title-input">
  <button type="submit" class="btn btn-xs">save</button>
  <button type="button" class="btn btn-xs" onclick="this.parentElement.style.display='none'; this.parentElement.parentElement.querySelector('.bead-title-text').style.display=''; this.parentElement.parentElement.querySelector('.bead-title-edit').style.display='';">cancel</button>
</form>`,
		template.HTMLEscapeString(title), database, template.HTMLEscapeString(id),
		template.HTMLEscapeString(title))
}

// handleBatchStatus handles POST /batch/status to update multiple beads at once.
// Expects form fields: ids[] (format: "db/id"), status.
func (s *Server) handleBatchStatus(w http.ResponseWriter, r *http.Request) {
	if s.ds == nil {
		http.Error(w, "no database", http.StatusServiceUnavailable)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	newStatus := r.FormValue("status")
	validStatuses := map[string]bool{
		"open": true, "closed": true, "deferred": true,
	}
	if !validStatuses[newStatus] {
		http.Error(w, "invalid status", http.StatusBadRequest)
		return
	}

	ids := r.Form["ids[]"]
	if len(ids) == 0 {
		http.Error(w, "no beads specified", http.StatusBadRequest)
		return
	}
	if len(ids) > 100 {
		http.Error(w, "too many beads (max 100)", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	updated := 0
	for _, ref := range ids {
		parts := strings.SplitN(ref, "/", 2)
		if len(parts) != 2 {
			continue
		}
		db, id := parts[0], parts[1]
		if err := s.ds.UpdateStatus(ctx, db, id, newStatus); err != nil {
			log.Printf("batch status: %s/%s: %v", db, id, err)
			continue
		}
		updated++
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("HX-Trigger", fmt.Sprintf(`{"showToast":"%d beads %s"}`, updated, newStatus))
	fmt.Fprintf(w, `<div class="text-dim">%d beads updated to %s</div>`, updated, template.HTMLEscapeString(newStatus))
}

func statusClassName(s string) string {
	switch s {
	case "open":
		return "status-open"
	case "closed", "completed":
		return "status-closed"
	case "in_progress", "hooked":
		return "status-progress"
	case "blocked":
		return "status-blocked"
	case "deferred":
		return "status-deferred"
	default:
		return ""
	}
}

// parseDescriptionMetadata extracts key: value lines from the start of a
// description (used by beads for attached_molecule, dispatched_by, etc.)
// and returns them as a map plus the remaining description text.
func parseDescriptionMetadata(desc string) (map[string]string, string) {
	if desc == "" {
		return nil, ""
	}
	lines := strings.Split(desc, "\n")
	info := make(map[string]string)
	bodyStart := 0
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			bodyStart = i + 1
			continue
		}
		idx := strings.Index(line, ": ")
		if idx > 0 && idx < 30 && !strings.Contains(line[:idx], " ") {
			key := line[:idx]
			info[key] = line[idx+2:]
			bodyStart = i + 1
		} else {
			bodyStart = i
			break
		}
	}
	if len(info) == 0 {
		return nil, desc
	}
	remaining := strings.TrimSpace(strings.Join(lines[bodyStart:], "\n"))
	return info, remaining
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
