package dolt

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// Issues queries issues from the specified database with optional filters.
func (c *Client) Issues(ctx context.Context, database string, f IssueFilter) ([]Issue, error) {
	query, args := buildIssueQuery(f, "")
	rows, err := c.queryDB(ctx, database, query, args...)
	if err != nil {
		return nil, fmt.Errorf("dolt: query issues: %w", err)
	}
	defer func() { _ = rows.Close() }()
	return scanIssues(rows)
}

// IssuesAsOf queries issues at a specific point in time using Dolt's
// AS OF clause.
func (c *Client) IssuesAsOf(ctx context.Context, database string, ts time.Time, f IssueFilter) ([]Issue, error) {
	asOf := ts.UTC().Format("2006-01-02T15:04:05")
	query, args := buildIssueQuery(f, asOf)
	rows, err := c.queryDB(ctx, database, query, args...)
	if err != nil {
		return nil, fmt.Errorf("dolt: query issues as of: %w", err)
	}
	defer func() { _ = rows.Close() }()
	return scanIssues(rows)
}

// IssueByID returns a single issue by ID from the specified database.
// Returns nil, nil if the issue is not found.
func (c *Client) IssueByID(ctx context.Context, database, id string) (*Issue, error) {
	query := "SELECT id, title, description, status, priority, issue_type, " +
		"COALESCE(owner,''), COALESCE(assignee,''), created_at, updated_at " +
		"FROM issues WHERE id = ? AND deleted_at IS NULL"
	rows, err := c.queryDB(ctx, database, query, id)
	if err != nil {
		return nil, fmt.Errorf("dolt: issue by id: %w", err)
	}
	defer func() { _ = rows.Close() }()

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("dolt: issue by id: %w", err)
		}
		return nil, nil // not found
	}

	var iss Issue
	if err := scanIssue(rows, &iss); err != nil {
		return nil, err
	}
	return &iss, nil
}

// Comments returns comments for the given issue.
func (c *Client) Comments(ctx context.Context, database, issueID string) ([]Comment, error) {
	query := "SELECT id, issue_id, author, text, created_at " +
		"FROM comments WHERE issue_id = ? ORDER BY created_at"
	rows, err := c.queryDB(ctx, database, query, issueID)
	if err != nil {
		return nil, fmt.Errorf("dolt: comments: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var comments []Comment
	for rows.Next() {
		var cm Comment
		if err := rows.Scan(&cm.ID, &cm.IssueID, &cm.Author, &cm.Body, &cm.CreatedAt); err != nil {
			return nil, fmt.Errorf("dolt: scan comment: %w", err)
		}
		comments = append(comments, cm)
	}
	return comments, rows.Err()
}

// Dependencies returns dependency edges for the given issue.
func (c *Client) Dependencies(ctx context.Context, database, issueID string) ([]Dependency, error) {
	query := "SELECT issue_id, depends_on, dep_type FROM dependencies WHERE issue_id = ? OR depends_on = ?"
	rows, err := c.queryDB(ctx, database, query, issueID, issueID)
	if err != nil {
		return nil, fmt.Errorf("dolt: dependencies: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var deps []Dependency
	for rows.Next() {
		var d Dependency
		if err := rows.Scan(&d.FromID, &d.ToID, &d.Type); err != nil {
			return nil, fmt.Errorf("dolt: scan dependency: %w", err)
		}
		deps = append(deps, d)
	}
	return deps, rows.Err()
}

// Diff returns changes between two revisions for a table.
// The from and to parameters can be commit hashes, branch names, or timestamps.
func (c *Client) Diff(ctx context.Context, database, table, from, to string) ([]DiffRow, error) {
	query := "SELECT diff_type, from_id, to_id, from_status, to_status, from_commit, to_commit " +
		"FROM dolt_diff(?, ?, ?)"
	rows, err := c.queryDB(ctx, database, query, table, from, to)
	if err != nil {
		return nil, fmt.Errorf("dolt: diff: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var diffs []DiffRow
	for rows.Next() {
		var d DiffRow
		if err := rows.Scan(&d.DiffType, &d.FromID, &d.ToID, &d.FromStatus, &d.ToStatus, &d.FromCommit, &d.ToCommit); err != nil {
			return nil, fmt.Errorf("dolt: scan diff: %w", err)
		}
		diffs = append(diffs, d)
	}
	return diffs, rows.Err()
}

// CountByStatus returns a map of status -> count for issues in the database.
func (c *Client) CountByStatus(ctx context.Context, database string) (map[string]int, error) {
	query := "SELECT status, COUNT(*) FROM issues WHERE deleted_at IS NULL " +
		"AND issue_type IN ('task','bug','epic') GROUP BY status"
	rows, err := c.queryDB(ctx, database, query)
	if err != nil {
		return nil, fmt.Errorf("dolt: count by status: %w", err)
	}
	defer func() { _ = rows.Close() }()

	counts := make(map[string]int)
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, fmt.Errorf("dolt: scan count: %w", err)
		}
		counts[status] = count
	}
	return counts, rows.Err()
}

// CountCreatedInRange counts issues created within the given time range.
func (c *Client) CountCreatedInRange(ctx context.Context, database string, start, end time.Time) (int, error) {
	query := "SELECT COUNT(*) FROM issues WHERE deleted_at IS NULL " +
		"AND issue_type IN ('task','bug','epic') " +
		"AND created_at >= ? AND created_at < ?"
	rows, err := c.queryDB(ctx, database, query, start, end)
	if err != nil {
		return 0, fmt.Errorf("dolt: count created in range: %w", err)
	}
	defer func() { _ = rows.Close() }()
	if !rows.Next() {
		return 0, rows.Err()
	}
	var count int
	if err := rows.Scan(&count); err != nil {
		return 0, fmt.Errorf("dolt: scan count: %w", err)
	}
	return count, rows.Err()
}

// CountClosedInRange counts issues closed (status='closed', updated) within the given time range.
func (c *Client) CountClosedInRange(ctx context.Context, database string, start, end time.Time) (int, error) {
	query := "SELECT COUNT(*) FROM issues WHERE deleted_at IS NULL " +
		"AND issue_type IN ('task','bug','epic') " +
		"AND status = 'closed' AND updated_at >= ? AND updated_at < ?"
	rows, err := c.queryDB(ctx, database, query, start, end)
	if err != nil {
		return 0, fmt.Errorf("dolt: count closed in range: %w", err)
	}
	defer func() { _ = rows.Close() }()
	if !rows.Next() {
		return 0, rows.Err()
	}
	var count int
	if err := rows.Scan(&count); err != nil {
		return 0, fmt.Errorf("dolt: scan count: %w", err)
	}
	return count, rows.Err()
}

// Epics returns all issues of type "epic" from the database.
func (c *Client) Epics(ctx context.Context, database string) ([]Issue, error) {
	return c.Issues(ctx, database, IssueFilter{Type: "epic"})
}

// EpicChildIDs returns the IDs of all direct children of the given epic.
func (c *Client) EpicChildIDs(ctx context.Context, database, epicID string) ([]string, error) {
	deps, err := c.Dependencies(ctx, database, epicID)
	if err != nil {
		return nil, err
	}
	var ids []string
	for _, d := range deps {
		if d.Type == "child_of" && d.ToID == epicID {
			ids = append(ids, d.FromID)
		}
	}
	return ids, nil
}

// AgentActivity returns aggregated stats per agent (by owner field).
func (c *Client) AgentActivity(ctx context.Context, database string) ([]AgentStats, error) {
	query := `SELECT COALESCE(owner,'(unowned)') AS agent,
		COUNT(*) AS total,
		SUM(CASE WHEN status = 'closed' THEN 1 ELSE 0 END) AS closed,
		SUM(CASE WHEN status = 'open' THEN 1 ELSE 0 END) AS open_count,
		SUM(CASE WHEN status IN ('in_progress','hooked') THEN 1 ELSE 0 END) AS in_progress
		FROM issues WHERE deleted_at IS NULL AND issue_type IN ('task','bug','epic')
		GROUP BY owner ORDER BY total DESC`
	rows, err := c.queryDB(ctx, database, query)
	if err != nil {
		return nil, fmt.Errorf("dolt: agent activity: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var agents []AgentStats
	for rows.Next() {
		var a AgentStats
		if err := rows.Scan(&a.Name, &a.Owned, &a.Closed, &a.Open, &a.InProgress); err != nil {
			return nil, fmt.Errorf("dolt: scan agent: %w", err)
		}
		agents = append(agents, a)
	}
	return agents, rows.Err()
}

// SearchIssues returns issues where the title or description matches the query string.
func (c *Client) SearchIssues(ctx context.Context, database, q string, limit int) ([]Issue, error) {
	if q == "" {
		return nil, nil
	}
	pattern := "%" + q + "%"
	query := "SELECT id, title, description, status, priority, issue_type, " +
		"COALESCE(owner,''), COALESCE(assignee,''), created_at, updated_at " +
		"FROM issues WHERE deleted_at IS NULL AND issue_type IN ('task','bug','epic') " +
		"AND (title LIKE ? OR description LIKE ?) " +
		"ORDER BY updated_at DESC"
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}
	rows, err := c.queryDB(ctx, database, query, pattern, pattern)
	if err != nil {
		return nil, fmt.Errorf("dolt: search issues: %w", err)
	}
	defer func() { _ = rows.Close() }()
	return scanIssues(rows)
}

// StatusHistory returns the status transitions for an issue by walking
// dolt_history_issues and detecting changes between consecutive commits.
func (c *Client) StatusHistory(ctx context.Context, database, issueID string) ([]StatusTransition, error) {
	query := "SELECT status, commit_date FROM dolt_history_issues WHERE id = ? ORDER BY commit_date"
	rows, err := c.queryDB(ctx, database, query, issueID)
	if err != nil {
		return nil, fmt.Errorf("dolt: status history: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var transitions []StatusTransition
	var prevStatus string
	first := true
	for rows.Next() {
		var status string
		var commitDate time.Time
		if err := rows.Scan(&status, &commitDate); err != nil {
			return nil, fmt.Errorf("dolt: scan status history: %w", err)
		}
		if first {
			transitions = append(transitions, StatusTransition{
				ToStatus:   status,
				CommitDate: commitDate,
			})
			prevStatus = status
			first = false
			continue
		}
		if status != prevStatus {
			transitions = append(transitions, StatusTransition{
				FromStatus: prevStatus,
				ToStatus:   status,
				CommitDate: commitDate,
			})
			prevStatus = status
		}
	}
	return transitions, rows.Err()
}

// DistinctAssignees returns distinct non-empty assignee values from the database.
func (c *Client) DistinctAssignees(ctx context.Context, database string) ([]string, error) {
	query := "SELECT DISTINCT COALESCE(assignee,'') FROM issues WHERE deleted_at IS NULL " +
		"AND issue_type IN ('task','bug','epic') AND assignee IS NOT NULL AND assignee != '' " +
		"ORDER BY assignee"
	rows, err := c.queryDB(ctx, database, query)
	if err != nil {
		return nil, fmt.Errorf("dolt: distinct assignees: %w", err)
	}
	defer func() { _ = rows.Close() }()
	var result []string
	for rows.Next() {
		var a string
		if err := rows.Scan(&a); err != nil {
			return nil, fmt.Errorf("dolt: scan assignee: %w", err)
		}
		result = append(result, a)
	}
	return result, rows.Err()
}

// buildIssueQuery constructs a SELECT for issues with optional filters
// and optional AS OF clause. Does NOT include USE prefix.
func buildIssueQuery(f IssueFilter, asOf string) (string, []any) {
	var b strings.Builder
	var args []any

	b.WriteString("SELECT id, title, description, status, priority, issue_type, " +
		"COALESCE(owner,''), COALESCE(assignee,''), created_at, updated_at FROM ")

	if asOf != "" {
		b.WriteString(fmt.Sprintf("issues AS OF '%s'", asOf))
	} else {
		b.WriteString("issues")
	}

	conditions := []string{"deleted_at IS NULL", "issue_type IN ('task','bug','epic')"}
	if f.Status != "" {
		conditions = append(conditions, "status = ?")
		args = append(args, f.Status)
	}
	if f.Priority != 0 {
		conditions = append(conditions, "priority = ?")
		args = append(args, f.Priority)
	}
	if f.Type != "" {
		conditions = append(conditions, "issue_type = ?")
		args = append(args, f.Type)
	}
	if f.Assignee != "" {
		conditions = append(conditions, "assignee = ?")
		args = append(args, f.Assignee)
	}
	if f.Owner != "" {
		conditions = append(conditions, "owner = ?")
		args = append(args, f.Owner)
	}
	if !f.UpdatedAfter.IsZero() {
		conditions = append(conditions, "updated_at >= ?")
		args = append(args, f.UpdatedAfter)
	}
	if !f.UpdatedBefore.IsZero() {
		conditions = append(conditions, "updated_at < ?")
		args = append(args, f.UpdatedBefore)
	}

	b.WriteString(" WHERE ")
	b.WriteString(strings.Join(conditions, " AND "))
	b.WriteString(" ORDER BY updated_at DESC")

	if f.Limit > 0 {
		b.WriteString(fmt.Sprintf(" LIMIT %d", f.Limit))
	}

	return b.String(), args
}

// scanIssue scans a single issue row from the given scanner.
func scanIssue(s interface{ Scan(...any) error }, iss *Issue) error {
	return s.Scan(
		&iss.ID, &iss.Title, &iss.Description, &iss.Status,
		&iss.Priority, &iss.Type, &iss.Owner, &iss.Assignee,
		&iss.CreatedAt, &iss.UpdatedAt,
	)
}

// scanIssues scans all rows into issues.
func scanIssues(rows interface {
	Next() bool
	Err() error
	Scan(...any) error
},
) ([]Issue, error) {
	var issues []Issue
	for rows.Next() {
		var iss Issue
		if err := scanIssue(rows, &iss); err != nil {
			return nil, fmt.Errorf("dolt: scan issue: %w", err)
		}
		issues = append(issues, iss)
	}
	return issues, rows.Err()
}
