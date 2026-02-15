// Package aggregator computes summary data from beads databases.
package aggregator

import (
	"context"
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
}

// Stats holds issue counts for a period.
type Stats struct {
	Created  int
	Closed   int
	Open     int
	InFlight int
}

// RigSummary holds per-rig statistics and top completions.
type RigSummary struct {
	Name  string
	Stats Stats
	Top   []dolt.Issue
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

		collectEpics(ctx, client, dbName, &s.Epics)
		mergeAgentStats(ctx, client, dbName, agentMerged)
	}

	s.TotalStats = total

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

	stats := Stats{
		Open:     counts["open"],
		InFlight: counts["in_progress"] + counts["hooked"],
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
		Name:  dbName,
		Stats: stats,
		Top:   top,
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
