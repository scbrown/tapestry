package web

import (
	"log"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/scbrown/tapestry/internal/dolt"
	"github.com/scbrown/tapestry/internal/events"
)

type monthlyData struct {
	Year       int
	Month      time.Month
	MonthName  string
	PrevYear   int
	PrevMonth  int
	NextYear   int
	NextMonth  int
	HasNext    bool
	Rigs       []rigViewData
	TotalStats statusCounts
}

type statusCounts struct {
	Created  int
	Closed   int
	Open     int
	InFlight int
}

type rigViewData struct {
	Name  string
	Stats statusCounts
	Top   []dolt.Issue
}

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

	data := monthlyData{
		Year:      year,
		Month:     time.Month(month),
		MonthName: time.Month(month).String(),
	}

	prev := time.Date(year, time.Month(month)-1, 1, 0, 0, 0, 0, time.UTC)
	data.PrevYear = prev.Year()
	data.PrevMonth = int(prev.Month())

	next := time.Date(year, time.Month(month)+1, 1, 0, 0, 0, 0, time.UTC)
	data.NextYear = next.Year()
	data.NextMonth = int(next.Month())
	data.HasNext = next.Before(now) || (next.Year() == now.Year() && next.Month() <= now.Month())

	ctx := r.Context()
	var total statusCounts

	for _, dbName := range s.databases() {
		counts, err := s.client.CountByStatus(ctx, dbName)
		if err != nil {
			log.Printf("counts %s: %v", dbName, err)
			continue
		}

		stats := statusCounts{
			Open:     counts["open"],
			InFlight: counts["in_progress"] + counts["hooked"],
			Closed:   counts["closed"],
		}

		top, err := s.client.Issues(ctx, dbName, dolt.IssueFilter{
			Status: "closed",
			Limit:  10,
		})
		if err != nil {
			log.Printf("top %s: %v", dbName, err)
		}

		data.Rigs = append(data.Rigs, rigViewData{
			Name:  dbName,
			Stats: stats,
			Top:   top,
		})

		total.Open += stats.Open
		total.Closed += stats.Closed
		total.InFlight += stats.InFlight
	}
	data.TotalStats = total

	s.render(w, "monthly.html", data)
}

type beadData struct {
	Issue    *dolt.Issue
	Comments []dolt.Comment
	Children []dolt.Issue
	RigName  string
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
		for _, d := range deps {
			if d.Type == "child_of" && d.ToID == id {
				child, err := s.client.IssueByID(ctx, dbName, d.FromID)
				if err == nil && child != nil {
					children = append(children, *child)
				}
			}
		}

		data := beadData{
			Issue:    issue,
			Comments: comments,
			Children: children,
			RigName:  dbName,
		}

		s.render(w, "bead.html", data)
		return
	}

	http.NotFound(w, r)
}

type beadListData struct {
	Status   string
	Type     string
	Assignee string
	Rig      string
	Rigs     []string
	Issues   []dolt.Issue
}

func (s *Server) handleBeadList(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	status := q.Get("status")
	typeFilter := q.Get("type")
	assignee := q.Get("assignee")
	rig := q.Get("rig")

	ctx := r.Context()
	var allIssues []dolt.Issue

	for _, dbName := range s.databases() {
		if rig != "" && dbName != rig {
			continue
		}
		issues, err := s.client.Issues(ctx, dbName, dolt.IssueFilter{
			Status:   status,
			Type:     typeFilter,
			Assignee: assignee,
			Limit:    200,
		})
		if err != nil {
			log.Printf("list %s: %v", dbName, err)
			continue
		}
		allIssues = append(allIssues, issues...)
	}

	data := beadListData{
		Status:   status,
		Type:     typeFilter,
		Assignee: assignee,
		Rig:      rig,
		Rigs:     s.databases(),
		Issues:   allIssues,
	}

	s.render(w, "beads.html", data)
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
