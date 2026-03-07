package web

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
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
	Commits  []beadCommit
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

	if s.forgejo != nil {
		data.Commits = s.forgejo.searchCommitsForBead(ctx, id)
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
	BlockedItems    []briefingBlockedItem
	RecentlyClosed  []dolt.Issue
	AgentStats      []dolt.AgentStats
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
		blockedItems    []briefingBlockedItem
		recentlyClosed  []dolt.Issue
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
		data.BlockedItems = append(data.BlockedItems, r.blockedItems...)
		data.RecentlyClosed = append(data.RecentlyClosed, r.recentlyClosed...)
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

	agentMap := make(map[string]*dolt.AgentStats)
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
