// Package aggregator computes summary data from beads databases.
package aggregator

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/scbrown/tapestry/internal/dolt"
)

// MonthlySummary contains all data for the monthly landing page.
type MonthlySummary struct {
	Year       int
	Month      time.Month
	MonthName  string
	PrevYear   int
	PrevMonth  int
	NextYear   int
	NextMonth  int
	HasNext    bool
	RigFilter  string
	AllRigs    []string
	Rigs       []RigSummary
	TotalStats Stats
	Epics      []EpicRow
	Agents     []AgentRow
	Weeks      []WeeklyTrend
}

// Stats holds issue counts for a period.
type Stats struct {
	Created        int
	Closed         int
	Open           int
	InFlight       int
	CompletionRate int // percentage of all issues that are closed (0-100)
}

// WeeklyTrend holds burndown data for a week within a month.
type WeeklyTrend struct {
	Label   string
	Created int
	Closed  int
	Net     int // Created - Closed (negative means more closed than created)
}

// RigSummary holds per-rig statistics and top completions.
type RigSummary struct {
	Name      string
	Stats     Stats
	Top       []dolt.Issue
	AllClosed int // all-time closed count (used for completion rate)
}

// EpicRow shows an epic with its completion progress.
type EpicRow struct {
	Issue    dolt.Issue
	RigName  string
	Progress dolt.EpicProgress
}

// AgentRow shows per-agent activity.
type AgentRow struct {
	Name       string
	Owned      int
	Closed     int
	Open       int
	InProgress int
}

// Monthly computes the monthly summary for the given year, month, and
// optional rig filter. The databases slice should list all beads_* database
// names to query.
func Monthly(ctx context.Context, client *dolt.Client, databases []string, year, month int, rigFilter string) (*MonthlySummary, error) {
	now := time.Now()
	monthStart := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)

	prev := monthStart.AddDate(0, -1, 0)
	next := monthStart.AddDate(0, 1, 0)

	s := &MonthlySummary{
		Year:      year,
		Month:     time.Month(month),
		MonthName: time.Month(month).String(),
		PrevYear:  prev.Year(),
		PrevMonth: int(prev.Month()),
		NextYear:  next.Year(),
		NextMonth: int(next.Month()),
		HasNext:   next.Before(now) || (next.Year() == now.Year() && next.Month() <= now.Month()),
		RigFilter: rigFilter,
		AllRigs:   databases,
	}

	agentMerged := make(map[string]*AgentRow)
	var total Stats
	var totalAllClosed int

	for _, dbName := range databases {
		rigName := RigDisplayName(dbName)
		if rigFilter != "" && rigName != rigFilter && dbName != rigFilter {
			continue
		}

		rs := computeRigStats(ctx, client, dbName, monthStart, next)
		s.Rigs = append(s.Rigs, rs)

		total.Open += rs.Stats.Open
		total.Created += rs.Stats.Created
		total.Closed += rs.Stats.Closed
		total.InFlight += rs.Stats.InFlight
		totalAllClosed += rs.AllClosed

		collectEpics(ctx, client, dbName, &s.Epics)
		mergeAgentStats(ctx, client, dbName, agentMerged)
	}

	allTotal := total.Open + total.InFlight + totalAllClosed
	if allTotal > 0 {
		total.CompletionRate = totalAllClosed * 100 / allTotal
	}
	s.TotalStats = total

	// Compute weekly burndown trends
	s.Weeks = computeWeeklyTrends(ctx, client, databases, monthStart, next, rigFilter)

	// Sort epics by progress (most complete children first)
	sort.Slice(s.Epics, func(i, j int) bool {
		return s.Epics[i].Progress.Closed > s.Epics[j].Progress.Closed
	})
	if len(s.Epics) > 10 {
		s.Epics = s.Epics[:10]
	}

	// Build sorted agent list
	for _, row := range agentMerged {
		s.Agents = append(s.Agents, *row)
	}
	sort.Slice(s.Agents, func(i, j int) bool {
		return s.Agents[i].Owned > s.Agents[j].Owned
	})

	return s, nil
}

// RigDisplayName strips the "beads_" prefix from a database name.
func RigDisplayName(dbName string) string {
	return strings.TrimPrefix(dbName, "beads_")
}

func computeRigStats(ctx context.Context, client *dolt.Client, dbName string, monthStart, monthEnd time.Time) RigSummary {
	counts, err := client.CountByStatus(ctx, dbName)
	if err != nil {
		log.Printf("counts %s: %v", dbName, err)
		return RigSummary{Name: dbName}
	}

	allClosed := counts["closed"]
	stats := Stats{
		Open:     counts["open"],
		InFlight: counts["in_progress"] + counts["hooked"],
	}

	allTotal := stats.Open + stats.InFlight + allClosed
	if allTotal > 0 {
		stats.CompletionRate = allClosed * 100 / allTotal
	}

	if n, err := client.CountCreatedInRange(ctx, dbName, monthStart, monthEnd); err == nil {
		stats.Created = n
	}
	if n, err := client.CountClosedInRange(ctx, dbName, monthStart, monthEnd); err == nil {
		stats.Closed = n
	}

	top, err := client.Issues(ctx, dbName, dolt.IssueFilter{
		Status:        "closed",
		UpdatedAfter:  monthStart,
		UpdatedBefore: monthEnd,
		Limit:         10,
	})
	if err != nil {
		log.Printf("top %s: %v", dbName, err)
	}

	return RigSummary{
		Name:      dbName,
		Stats:     stats,
		Top:       top,
		AllClosed: allClosed,
	}
}

func collectEpics(ctx context.Context, client *dolt.Client, dbName string, epics *[]EpicRow) {
	epicIssues, err := client.Epics(ctx, dbName)
	if err != nil {
		return
	}
	for _, epic := range epicIssues {
		_, prog, err := client.EpicChildren(ctx, dbName, epic.ID)
		if err != nil || prog.Total == 0 {
			continue
		}
		*epics = append(*epics, EpicRow{
			Issue:    epic,
			RigName:  dbName,
			Progress: prog,
		})
	}
}

func computeWeeklyTrends(ctx context.Context, client *dolt.Client, databases []string, monthStart, monthEnd time.Time, rigFilter string) []WeeklyTrend {
	// Split month into weeks (1-7, 8-14, 15-21, 22-end)
	var weeks []WeeklyTrend
	weekStart := monthStart
	for weekStart.Before(monthEnd) {
		weekEnd := weekStart.AddDate(0, 0, 7)
		if weekEnd.After(monthEnd) {
			weekEnd = monthEnd
		}
		label := fmt.Sprintf("%s %d–%d",
			monthStart.Month().String()[:3],
			weekStart.Day(),
			weekEnd.AddDate(0, 0, -1).Day())
		if weekEnd.Equal(monthEnd) && weekEnd.Day() == 1 {
			// Last day rolled into next month; use month's last day
			label = fmt.Sprintf("%s %d–%d",
				monthStart.Month().String()[:3],
				weekStart.Day(),
				monthEnd.AddDate(0, 0, -1).Day())
		}

		w := WeeklyTrend{Label: label}
		for _, dbName := range databases {
			if rigFilter != "" {
				rigName := RigDisplayName(dbName)
				if rigName != rigFilter && dbName != rigFilter {
					continue
				}
			}
			if n, err := client.CountCreatedInRange(ctx, dbName, weekStart, weekEnd); err == nil {
				w.Created += n
			}
			if n, err := client.CountClosedInRange(ctx, dbName, weekStart, weekEnd); err == nil {
				w.Closed += n
			}
		}
		w.Net = w.Created - w.Closed
		weeks = append(weeks, w)

		weekStart = weekEnd
	}
	return weeks
}

func mergeAgentStats(ctx context.Context, client *dolt.Client, dbName string, merged map[string]*AgentRow) {
	stats, err := client.AgentActivity(ctx, dbName)
	if err != nil {
		return
	}
	for _, a := range stats {
		if a.Name == "(unowned)" {
			continue
		}
		row, ok := merged[a.Name]
		if !ok {
			row = &AgentRow{Name: a.Name}
			merged[a.Name] = row
		}
		row.Owned += a.Owned
		row.Closed += a.Closed
		row.Open += a.Open
		row.InProgress += a.InProgress
	}
}
