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

	dbs, err := s.ds.ListBeadsDatabases(ctx)
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

	dbs, err := s.ds.ListBeadsDatabases(ctx)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	// Also check non-beads_ databases that might contain issues
	allDBs, _ := s.allDatabases(ctx)

	// Try beads_ databases first, then others
	tried := make(map[string]bool)
	for _, db := range dbs {
		tried[db.Name] = true
		issue, err := s.ds.IssueByID(ctx, db.Name, id)
		if err != nil {
			continue
		}
		if issue != nil {
			s.handleBead(w, r, db.Name, id)
			return
		}
	}
	for _, db := range allDBs {
		if tried[db] {
			continue
		}
		issue, err := s.ds.IssueByID(ctx, db, id)
		if err != nil {
			continue
		}
		if issue != nil {
			s.handleBead(w, r, db, id)
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

	dbs, err := s.ds.ListBeadsDatabases(ctx)
	if err != nil {
		data.Err = fmt.Sprintf("Failed to list databases: %v", err)
		s.render(w, r, "beads", data)
		return
	}

	// Also include non-beads_ databases
	allDBs, _ := s.allDatabases(ctx)
	dbSet := make(map[string]bool)
	for _, db := range dbs {
		dbSet[db.Name] = true
	}
	for _, db := range allDBs {
		if !dbSet[db] {
			dbs = append(dbs, dolt.DatabaseInfo{Name: db})
			dbSet[db] = true
		}
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

	dbs, err := s.ds.ListBeadsDatabases(ctx)
	if err != nil {
		data.Err = fmt.Sprintf("Failed to list databases: %v", err)
		s.render(w, r, "search", data)
		return
	}

	allDBs, _ := s.allDatabases(ctx)
	dbSet := make(map[string]bool)
	for _, db := range dbs {
		dbSet[db.Name] = true
	}
	for _, db := range allDBs {
		if !dbSet[db] {
			dbs = append(dbs, dolt.DatabaseInfo{Name: db})
		}
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

// allDatabases returns database names that contain an issues table
// but don't start with "beads_" (to complement ListBeadsDatabases).
func (s *Server) allDatabases(ctx context.Context) ([]string, error) {
	// Check well-known databases that aren't prefixed with beads_
	known := []string{"aegis", "gastown", "tapestry", "bobbin"}
	var result []string
	for _, db := range known {
		// Quick probe: try to read a single issue
		_, err := s.ds.IssueByID(ctx, db, "__probe__")
		if err == nil {
			result = append(result, db)
		}
	}
	return result, nil
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

	dbs, err := s.ds.ListBeadsDatabases(ctx)
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
				if isMolecule(iss.Title) {
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
				if iss.Status == "closed" || isMolecule(iss.Title) {
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
				if isMolecule(bi.Issue.Title) || isMolecule(bi.Blocker.Title) {
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
				if !isMolecule(iss.Title) {
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
				if iss.Status == "closed" || isMolecule(iss.Title) {
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
	if len(data.RecentClosed) > 15 {
		data.RecentClosed = data.RecentClosed[:15]
	}

	sort.Slice(data.ActionItems, func(i, j int) bool {
		return data.ActionItems[i].Issue.Priority < data.ActionItems[j].Issue.Priority
	})

	s.render(w, r, "status", data)
}

func isMolecule(title string) bool {
	return strings.HasPrefix(title, "mol-")
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
