package web

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/scbrown/tapestry/internal/dolt"
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

	// Previous month
	prev := time.Date(year, time.Month(month)-1, 1, 0, 0, 0, 0, time.UTC)
	data.PrevYear = prev.Year()
	data.PrevMonth = int(prev.Month())

	// Next month (only if not in the future)
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

		// Get top recently-updated closed issues
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

	// Search all configured databases
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
		// Get children via dependencies
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
	Status string
	Issues []dolt.Issue
}

func (s *Server) handleBeadList(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")

	ctx := r.Context()
	var allIssues []dolt.Issue

	for _, dbName := range s.databases() {
		issues, err := s.client.Issues(ctx, dbName, dolt.IssueFilter{
			Status: status,
			Limit:  100,
		})
		if err != nil {
			log.Printf("list %s: %v", dbName, err)
			continue
		}
		allIssues = append(allIssues, issues...)
	}

	data := beadListData{
		Status: status,
		Issues: allIssues,
	}

	s.render(w, "beads.html", data)
}
