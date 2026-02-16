package web

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/scbrown/tapestry/internal/aggregator"
	"github.com/scbrown/tapestry/internal/dolt"
	"github.com/scbrown/tapestry/internal/events"
	gitpkg "github.com/scbrown/tapestry/internal/git"
)

func (s *Server) handleMonthly(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	year, month := now.Year(), int(now.Month())

	if y := r.PathValue("year"); y != "" {
		var err error
		year, err = strconv.Atoi(y)
		if err != nil {
			http.Error(w, "bad year", http.StatusBadRequest)
			return
		}
	}
	if m := r.PathValue("month"); m != "" {
		var err error
		month, err = strconv.Atoi(m)
		if err != nil || month < 1 || month > 12 {
			http.Error(w, "bad month", http.StatusBadRequest)
			return
		}
	}

	rigFilter := r.URL.Query().Get("rig")
	data, err := aggregator.Monthly(r.Context(), s.client, s.databases(), year, month, rigFilter)
	if err != nil {
		log.Printf("monthly aggregation: %v", err)
		http.Error(w, "aggregation failed", http.StatusInternalServerError)
		return
	}

	s.render(w, "monthly.html", data)
}

type beadData struct {
	Issue         *dolt.Issue
	Comments      []dolt.Comment
	Children      []dolt.Issue
	Blocks        []dolt.Issue // issues this bead blocks
	BlockedBy     []dolt.Issue // issues that block this bead
	StatusHistory []dolt.StatusTransition
	Commits       []gitpkg.Commit
	RigName       string
}

func (s *Server) handleBead(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "missing bead id", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	for _, dbName := range s.databases() {
		issue, err := s.client.IssueByID(ctx, dbName, id)
		if err != nil {
			log.Printf("get %s from %s: %v", id, dbName, err)
			continue
		}
		if issue == nil {
			continue
		}

		comments, _ := s.client.Comments(ctx, dbName, id)
		deps, _ := s.client.Dependencies(ctx, dbName, id)
		var children []dolt.Issue
		var blocks []dolt.Issue
		var blockedBy []dolt.Issue
		for _, d := range deps {
			if d.Type == "child_of" && d.ToID == id {
				child, err := s.client.IssueByID(ctx, dbName, d.FromID)
				if err == nil && child != nil {
					children = append(children, *child)
				}
			} else if d.ToID == id {
				// something depends on this issue → this issue blocks it
				blocked, err := s.client.IssueByID(ctx, dbName, d.FromID)
				if err == nil && blocked != nil {
					blocks = append(blocks, *blocked)
				}
			} else if d.FromID == id {
				// this issue depends on something → blocked by it
				blocker, err := s.client.IssueByID(ctx, dbName, d.ToID)
				if err == nil && blocker != nil {
					blockedBy = append(blockedBy, *blocker)
				}
			}
		}

		statusHistory, _ := s.client.StatusHistory(ctx, dbName, id)
		commits := s.commitsForBead(id)

		data := beadData{
			Issue:         issue,
			Comments:      comments,
			Children:      children,
			Blocks:        blocks,
			BlockedBy:     blockedBy,
			StatusHistory: statusHistory,
			Commits:       commits,
			RigName:       dbName,
		}

		s.render(w, "bead.html", data)
		return
	}

	http.NotFound(w, r)
}

type issueRow struct {
	dolt.Issue
	Rig string
}

type pageLink struct {
	Num    int
	URL    string
	Active bool
}

type beadListData struct {
	Status   string
	Type     string
	Priority string
	Assignee string
	Rig      string

	Rigs      []string
	Assignees []string

	Issues    []issueRow
	Total     int
	Page      int
	PageSize  int
	Pages     int
	PageLinks []pageLink
}

func (s *Server) handleBeadList(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	status := q.Get("status")
	typeFilter := q.Get("type")
	assignee := q.Get("assignee")
	rig := q.Get("rig")
	priorityStr := q.Get("priority")

	var priorityFilter int
	if priorityStr != "" {
		if p, err := strconv.Atoi(priorityStr); err == nil && p > 0 {
			priorityFilter = p
		}
	}

	page := 1
	if p := q.Get("page"); p != "" {
		if n, err := strconv.Atoi(p); err == nil && n > 0 {
			page = n
		}
	}
	const pageSize = 50

	ctx := r.Context()
	var allIssues []issueRow
	assigneeSet := make(map[string]bool)

	for _, dbName := range s.databases() {
		// Collect assignees for filter dropdown (from all rigs, unfiltered)
		if assignees, err := s.client.DistinctAssignees(ctx, dbName); err == nil {
			for _, a := range assignees {
				assigneeSet[a] = true
			}
		}

		if rig != "" && dbName != rig {
			continue
		}
		issues, err := s.client.Issues(ctx, dbName, dolt.IssueFilter{
			Status:   status,
			Type:     typeFilter,
			Assignee: assignee,
			Priority: priorityFilter,
			Limit:    500,
		})
		if err != nil {
			log.Printf("list %s: %v", dbName, err)
			continue
		}
		for _, iss := range issues {
			allIssues = append(allIssues, issueRow{Issue: iss, Rig: dbName})
		}
	}

	sort.Slice(allIssues, func(i, j int) bool {
		return allIssues[i].UpdatedAt.After(allIssues[j].UpdatedAt)
	})

	total := len(allIssues)
	pages := (total + pageSize - 1) / pageSize
	if pages == 0 {
		pages = 1
	}
	if page > pages {
		page = pages
	}

	start := (page - 1) * pageSize
	end := start + pageSize
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}
	paged := allIssues[start:end]

	var assigneeList []string
	for a := range assigneeSet {
		assigneeList = append(assigneeList, a)
	}
	sort.Strings(assigneeList)

	var pageLinks []pageLink
	for i := 1; i <= pages; i++ {
		pageLinks = append(pageLinks, pageLink{
			Num:    i,
			URL:    beadFilterURL(status, typeFilter, rig, priorityStr, assignee, i),
			Active: i == page,
		})
	}

	data := beadListData{
		Status:    status,
		Type:      typeFilter,
		Priority:  priorityStr,
		Assignee:  assignee,
		Rig:       rig,
		Rigs:      s.databases(),
		Assignees: assigneeList,
		Issues:    paged,
		Total:     total,
		Page:      page,
		PageSize:  pageSize,
		Pages:     pages,
		PageLinks: pageLinks,
	}

	if r.Header.Get("HX-Request") != "" {
		s.renderPartial(w, "beads.html", "beads-results", data)
		return
	}
	s.render(w, "beads.html", data)
}

func beadFilterURL(status, typ, rig, priority, assignee string, page int) string {
	v := url.Values{}
	if status != "" {
		v.Set("status", status)
	}
	if typ != "" {
		v.Set("type", typ)
	}
	if rig != "" {
		v.Set("rig", rig)
	}
	if priority != "" {
		v.Set("priority", priority)
	}
	if assignee != "" {
		v.Set("assignee", assignee)
	}
	if page > 1 {
		v.Set("page", strconv.Itoa(page))
	}
	if len(v) == 0 {
		return "/beads"
	}
	return "/beads?" + v.Encode()
}

type epicData struct {
	Issue    *dolt.Issue
	Children []dolt.Issue
	Progress dolt.EpicProgress
	RigName  string
	Comments []dolt.Comment
}

func (s *Server) handleEpic(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "missing epic id", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	for _, dbName := range s.databases() {
		issue, err := s.client.IssueByID(ctx, dbName, id)
		if err != nil {
			log.Printf("epic %s from %s: %v", id, dbName, err)
			continue
		}
		if issue == nil {
			continue
		}

		childIDs, _ := s.client.EpicChildIDs(ctx, dbName, id)
		var children []dolt.Issue
		var progress dolt.EpicProgress
		for _, cid := range childIDs {
			child, err := s.client.IssueByID(ctx, dbName, cid)
			if err == nil && child != nil {
				children = append(children, *child)
				progress.Total++
				if child.Status == "closed" {
					progress.Closed++
				}
			}
		}

		comments, _ := s.client.Comments(ctx, dbName, id)

		s.render(w, "epic.html", epicData{
			Issue:    issue,
			Children: children,
			Progress: progress,
			RigName:  dbName,
			Comments: comments,
		})
		return
	}

	http.NotFound(w, r)
}

type epicsListData struct {
	Epics []epicSummary
}

type epicSummary struct {
	Issue    dolt.Issue
	RigName  string
	Progress dolt.EpicProgress
}

func (s *Server) handleEpicsList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var allEpics []epicSummary

	for _, dbName := range s.databases() {
		issues, err := s.client.Epics(ctx, dbName)
		if err != nil {
			log.Printf("epics %s: %v", dbName, err)
			continue
		}
		for _, iss := range issues {
			childIDs, _ := s.client.EpicChildIDs(ctx, dbName, iss.ID)
			var prog dolt.EpicProgress
			for _, cid := range childIDs {
				child, err := s.client.IssueByID(ctx, dbName, cid)
				if err == nil && child != nil {
					prog.Total++
					if child.Status == "closed" {
						prog.Closed++
					}
				}
			}
			allEpics = append(allEpics, epicSummary{
				Issue:    iss,
				RigName:  dbName,
				Progress: prog,
			})
		}
	}

	s.render(w, "epics.html", epicsListData{Epics: allEpics})
}

type agentsData struct {
	Agents []agentRow
}

type agentRow struct {
	Name       string
	Owned      int
	Closed     int
	Open       int
	InProgress int
}

func (s *Server) handleAgents(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	merged := make(map[string]*agentRow)

	for _, dbName := range s.databases() {
		stats, err := s.client.AgentActivity(ctx, dbName)
		if err != nil {
			log.Printf("agents %s: %v", dbName, err)
			continue
		}
		for _, a := range stats {
			if a.Name == "(unowned)" {
				continue
			}
			row, ok := merged[a.Name]
			if !ok {
				row = &agentRow{Name: a.Name}
				merged[a.Name] = row
			}
			row.Owned += a.Owned
			row.Closed += a.Closed
			row.Open += a.Open
			row.InProgress += a.InProgress
		}
	}

	var agents []agentRow
	for _, row := range merged {
		agents = append(agents, *row)
	}
	sort.Slice(agents, func(i, j int) bool {
		return agents[i].Owned > agents[j].Owned
	})

	s.render(w, "agents.html", agentsData{Agents: agents})
}

type agentDetailData struct {
	Name   string
	Stats  agentRow
	Issues []dolt.Issue
}

func (s *Server) handleAgent(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if name == "" {
		http.Error(w, "missing agent name", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	var stats agentRow
	stats.Name = name
	var allIssues []dolt.Issue

	for _, dbName := range s.databases() {
		issues, err := s.client.Issues(ctx, dbName, dolt.IssueFilter{
			Owner: name,
			Limit: 100,
		})
		if err != nil {
			log.Printf("agent %s %s: %v", name, dbName, err)
			continue
		}
		for _, iss := range issues {
			stats.Owned++
			switch iss.Status {
			case "closed":
				stats.Closed++
			case "open":
				stats.Open++
			case "in_progress", "hooked":
				stats.InProgress++
			}
		}
		allIssues = append(allIssues, issues...)
	}

	s.render(w, "agent.html", agentDetailData{
		Name:   name,
		Stats:  stats,
		Issues: allIssues,
	})
}

type eventsData struct {
	Events      []events.Event
	Types       []string
	TypeFilter  string
	ActorFilter string
	Total       int
}

func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	typeFilter := q.Get("type")
	actorFilter := q.Get("actor")

	var allEvents []events.Event
	for _, ws := range s.cfg.Workspace {
		if ws.Path == "" {
			continue
		}
		evts, err := events.ReadWorkspace(ws.Path)
		if err != nil {
			log.Printf("events %s: %v", ws.Path, err)
			continue
		}
		allEvents = append(allEvents, evts...)
	}

	sort.Slice(allEvents, func(i, j int) bool {
		return allEvents[i].Timestamp.After(allEvents[j].Timestamp)
	})

	types := events.Types(allEvents)

	filtered := events.Apply(allEvents, events.Filter{
		Type:  typeFilter,
		Actor: actorFilter,
		Limit: 200,
	})

	s.render(w, "events.html", eventsData{
		Events:      filtered,
		Types:       types,
		TypeFilter:  typeFilter,
		ActorFilter: actorFilter,
		Total:       len(allEvents),
	})
}

// ── Search ──────────────────────────────────────────────────────

type searchData struct {
	Query   string
	Results []dolt.Issue
	Total   int
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")

	var allResults []dolt.Issue
	if q != "" {
		ctx := r.Context()
		for _, dbName := range s.databases() {
			issues, err := s.client.SearchIssues(ctx, dbName, q, 50)
			if err != nil {
				log.Printf("search %s: %v", dbName, err)
				continue
			}
			allResults = append(allResults, issues...)
		}
		sort.Slice(allResults, func(i, j int) bool {
			return allResults[i].UpdatedAt.After(allResults[j].UpdatedAt)
		})
		if len(allResults) > 100 {
			allResults = allResults[:100]
		}
	}

	s.render(w, "search.html", searchData{
		Query:   q,
		Results: allResults,
		Total:   len(allResults),
	})
}

// ── Handoff Chains ──────────────────────────────────────────────

type handoffsData struct {
	Chains      []events.HandoffChain
	Stats       []events.ChainStats
	ActorFilter string
	TotalChains int
}

func (s *Server) handleHandoffs(w http.ResponseWriter, r *http.Request) {
	actorFilter := r.URL.Query().Get("actor")

	allEvents := s.readAllEvents()
	chains := events.BuildHandoffChains(allEvents)

	if actorFilter != "" {
		var filtered []events.HandoffChain
		for _, c := range chains {
			if strings.Contains(c.Actor, actorFilter) {
				filtered = append(filtered, c)
			}
		}
		chains = filtered
	}

	s.render(w, "handoffs.html", handoffsData{
		Chains:      chains,
		Stats:       events.ChainSummary(chains),
		ActorFilter: actorFilter,
		TotalChains: len(chains),
	})
}

// ── Digest Export ───────────────────────────────────────────────

func (s *Server) handleDigest(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	year, month := now.Year(), int(now.Month())

	if y := r.PathValue("year"); y != "" {
		var err error
		year, err = strconv.Atoi(y)
		if err != nil {
			http.Error(w, "bad year", http.StatusBadRequest)
			return
		}
	}
	if m := r.PathValue("month"); m != "" {
		var err error
		month, err = strconv.Atoi(m)
		if err != nil || month < 1 || month > 12 {
			http.Error(w, "bad month", http.StatusBadRequest)
			return
		}
	}

	format := r.URL.Query().Get("format")
	if format == "" {
		format = "markdown"
	}

	ctx := r.Context()
	monthStart := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	monthEnd := monthStart.AddDate(0, 1, 0)

	var md strings.Builder
	md.WriteString(fmt.Sprintf("# Digest: %s %d\n\n", time.Month(month).String(), year))
	md.WriteString(fmt.Sprintf("Generated: %s\n\n", now.Format("2006-01-02 15:04")))

	// Per-rig stats
	md.WriteString("## Beads Summary\n\n")
	md.WriteString("| Rig | Open | In Progress | Closed | Total |\n")
	md.WriteString("|-----|------|-------------|--------|-------|\n")

	var totalOpen, totalInFlight, totalClosed int
	for _, dbName := range s.databases() {
		counts, err := s.client.CountByStatus(ctx, dbName)
		if err != nil {
			continue
		}
		rigName := strings.TrimPrefix(dbName, "beads_")
		open := counts["open"]
		inFlight := counts["in_progress"] + counts["hooked"]
		closed := counts["closed"]
		total := open + inFlight + closed
		md.WriteString(fmt.Sprintf("| %s | %d | %d | %d | %d |\n", rigName, open, inFlight, closed, total))
		totalOpen += open
		totalInFlight += inFlight
		totalClosed += closed
	}
	md.WriteString(fmt.Sprintf("| **Total** | **%d** | **%d** | **%d** | **%d** |\n\n",
		totalOpen, totalInFlight, totalClosed, totalOpen+totalInFlight+totalClosed))

	// Top completions
	md.WriteString("## Top Completions\n\n")
	for _, dbName := range s.databases() {
		issues, err := s.client.Issues(ctx, dbName, dolt.IssueFilter{Status: "closed", Limit: 10})
		if err != nil || len(issues) == 0 {
			continue
		}
		rigName := strings.TrimPrefix(dbName, "beads_")
		md.WriteString(fmt.Sprintf("### %s\n\n", rigName))
		for _, iss := range issues {
			if iss.UpdatedAt.Before(monthStart) || iss.UpdatedAt.After(monthEnd) {
				continue
			}
			md.WriteString(fmt.Sprintf("- **%s** %s (P%d, %s)\n", iss.ID, iss.Title, iss.Priority, iss.Assignee))
		}
		md.WriteString("\n")
	}

	// Agent activity
	md.WriteString("## Agent Activity\n\n")
	md.WriteString("| Agent | Owned | Closed | Open | In Progress |\n")
	md.WriteString("|-------|-------|--------|------|-------------|\n")

	agentMap := make(map[string]dolt.AgentStats)
	for _, dbName := range s.databases() {
		stats, err := s.client.AgentActivity(ctx, dbName)
		if err != nil {
			continue
		}
		for _, st := range stats {
			existing := agentMap[st.Name]
			existing.Name = st.Name
			existing.Owned += st.Owned
			existing.Closed += st.Closed
			existing.Open += st.Open
			existing.InProgress += st.InProgress
			agentMap[st.Name] = existing
		}
	}
	for _, st := range agentMap {
		md.WriteString(fmt.Sprintf("| %s | %d | %d | %d | %d |\n",
			st.Name, st.Owned, st.Closed, st.Open, st.InProgress))
	}
	md.WriteString("\n")

	// Handoff summary
	allEvents := s.readAllEvents()
	chains := events.BuildHandoffChains(allEvents)
	if len(chains) > 0 {
		md.WriteString("## Handoff Activity\n\n")
		md.WriteString("| Agent | Handoffs | Avg Session | Last Handoff |\n")
		md.WriteString("|-------|----------|-------------|-------------|\n")
		for _, st := range events.ChainSummary(chains) {
			if st.LastHandoff.Before(monthStart) {
				continue
			}
			avg := "-"
			if st.AvgSessionTime > 0 {
				avg = st.AvgSessionTime.Round(time.Minute).String()
			}
			md.WriteString(fmt.Sprintf("| %s | %d | %s | %s |\n",
				st.Actor, st.TotalHandoffs, avg, st.LastHandoff.Format("2006-01-02 15:04")))
		}
		md.WriteString("\n")
	}

	md.WriteString("---\n*Generated by Tapestry*\n")

	if format == "json" {
		w.Header().Set("Content-Type", "application/json")
		digestJSON := map[string]interface{}{
			"year":       year,
			"month":      month,
			"month_name": time.Month(month).String(),
			"generated":  now.Format(time.RFC3339),
			"content":    md.String(),
		}
		data, err := encodeJSON(digestJSON)
		if err != nil {
			http.Error(w, "json encode failed", http.StatusInternalServerError)
			return
		}
		_, _ = w.Write(data)
	} else {
		w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
		w.Header().Set("Content-Disposition",
			fmt.Sprintf("attachment; filename=\"digest-%d-%02d.md\"", year, month))
		_, _ = fmt.Fprint(w, md.String())
	}
}

// readAllEvents reads events from all configured workspaces.
func (s *Server) readAllEvents() []events.Event {
	var all []events.Event
	for _, ws := range s.cfg.Workspace {
		if ws.Path == "" {
			continue
		}
		evts, err := events.ReadWorkspace(ws.Path)
		if err != nil {
			log.Printf("events %s: %v", ws.Path, err)
			continue
		}
		all = append(all, evts...)
	}
	sort.Slice(all, func(i, j int) bool {
		return all[i].Timestamp.After(all[j].Timestamp)
	})
	return all
}

func encodeJSON(v interface{}) ([]byte, error) {
	return json.MarshalIndent(v, "", "  ")
}

// ── Git Commits ─────────────────────────────────────────────────

// readAllCommits parses git logs from all configured workspace paths.
func (s *Server) readAllCommits() []gitpkg.Commit {
	var all []gitpkg.Commit
	for _, ws := range s.cfg.Workspace {
		if ws.Path == "" {
			continue
		}
		commits, err := gitpkg.ParseWorkspace(ws.Path, 200)
		if err != nil {
			log.Printf("git %s: %v", ws.Path, err)
			continue
		}
		all = append(all, commits...)
	}
	// Apply config-based repo URL overrides (takes precedence over auto-detected)
	if len(s.cfg.Repos) > 0 {
		gitpkg.SetCommitURLs(all, s.cfg.Repos)
	}
	sort.Slice(all, func(i, j int) bool {
		return all[i].Timestamp.After(all[j].Timestamp)
	})
	return all
}

// commitsForBead returns commits that reference a specific bead ID.
func (s *Server) commitsForBead(beadID string) []gitpkg.Commit {
	all := s.readAllCommits()
	return gitpkg.CommitsForBead(all, beadID)
}

type commitsData struct {
	Commits      []gitpkg.Commit
	RecentLinked []gitpkg.Commit
	TotalCommits int
	LinkedCount  int
}

func (s *Server) handleCommits(w http.ResponseWriter, r *http.Request) {
	all := s.readAllCommits()
	linked := gitpkg.RecentWithBeads(all, 50)

	limit := 100
	display := all
	if len(display) > limit {
		display = display[:limit]
	}

	s.render(w, "commits.html", commitsData{
		Commits:      display,
		RecentLinked: linked,
		TotalCommits: len(all),
		LinkedCount:  len(linked),
	})
}

// ── Briefing ────────────────────────────────────────────────────

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
	RecentCommits   []gitpkg.Commit
	AgentStats      []dolt.AgentStats
}

func (s *Server) handleBriefing(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	data := briefingData{
		GeneratedAt: now,
	}

	for _, db := range s.databases() {
		// Status counts
		counts, err := s.client.CountByStatus(ctx, db)
		if err != nil {
			log.Printf("briefing: count by status %s: %v", db, err)
			continue
		}
		for status, n := range counts {
			switch status {
			case "open":
				data.OpenCount += n
			case "in_progress", "hooked":
				data.InProgressCount += n
			case "closed":
				data.ClosedCount += n
			}
			data.TotalBeads += n
		}

		// Closed today
		closedToday, err := s.client.CountClosedInRange(ctx, db, todayStart, now)
		if err == nil {
			data.ClosedToday += closedToday
		}

		// Created today
		createdToday, err := s.client.CountCreatedInRange(ctx, db, todayStart, now)
		if err == nil {
			data.CreatedToday += createdToday
		}

		// Needs attention: P1 open beads with no assignee or tagged needs-human
		unassigned, err := s.client.Issues(ctx, db, dolt.IssueFilter{
			Status:   "open",
			Priority: 1,
			Limit:    10,
		})
		if err == nil {
			for _, iss := range unassigned {
				if iss.Assignee == "" || strings.Contains(strings.ToLower(iss.Title), "needs-human") {
					data.NeedsAttention = append(data.NeedsAttention, iss)
				}
			}
		}

		// In flight: in_progress P1 beads
		inFlight, err := s.client.Issues(ctx, db, dolt.IssueFilter{
			Status:   "in_progress",
			Priority: 1,
			Limit:    10,
		})
		if err == nil {
			data.InFlight = append(data.InFlight, inFlight...)
		}

		// Recently closed (last 24h)
		yesterday := now.Add(-24 * time.Hour)
		recent, err := s.client.Issues(ctx, db, dolt.IssueFilter{
			Status:       "closed",
			UpdatedAfter: yesterday,
			Limit:        15,
		})
		if err == nil {
			data.RecentlyClosed = append(data.RecentlyClosed, recent...)
		}

		// Agent activity
		agents, err := s.client.AgentActivity(ctx, db)
		if err == nil {
			data.AgentStats = append(data.AgentStats, agents...)
		}
	}

	// Recent commits (last 20)
	allCommits := s.readAllCommits()
	limit := 20
	if len(allCommits) > limit {
		allCommits = allCommits[:limit]
	}
	data.RecentCommits = allCommits

	s.render(w, "briefing.html", data)
}
