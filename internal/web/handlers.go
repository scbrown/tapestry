package web

import (
	"fmt"
	"net/http"
	"strconv"
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

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	http.Redirect(w, r, fmt.Sprintf("/%d/%02d", now.Year(), now.Month()), http.StatusFound)
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

		recent, err := s.ds.Issues(ctx, db.Name, dolt.IssueFilter{Limit: 5})
		if err == nil {
			summary.Recent = recent
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
