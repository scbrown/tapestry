package dolt

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// Issues queries issues from the specified database with optional filters.
func (c *Client) Issues(ctx context.Context, database string, f IssueFilter) ([]Issue, error) {
	query, args := buildIssueQuery(database, f, "")
	return c.queryIssues(ctx, query, args)
}

// IssuesAsOf queries issues at a specific point in time using Dolt's
// AS OF clause.
func (c *Client) IssuesAsOf(ctx context.Context, database string, ts time.Time, f IssueFilter) ([]Issue, error) {
	asOf := ts.UTC().Format("2006-01-02T15:04:05")
	query, args := buildIssueQuery(database, f, asOf)
	return c.queryIssues(ctx, query, args)
}

// IssueByID returns a single issue by ID from the specified database.
// Returns nil, nil if the issue is not found.
func (c *Client) IssueByID(ctx context.Context, database, id string) (*Issue, error) {
	query := useDB(database) +
		"SELECT id, title, description, status, priority, type, owner, assignee, labels, created_at, updated_at " +
		"FROM issues WHERE id = ?"
	rows, err := c.db.QueryContext(ctx, query, id)
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
	query := useDB(database) +
		"SELECT id, issue_id, author, body, created_at " +
		"FROM comments WHERE issue_id = ? ORDER BY created_at"
	rows, err := c.db.QueryContext(ctx, query, issueID)
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
	query := useDB(database) +
		"SELECT from_id, to_id, type FROM deps WHERE from_id = ? OR to_id = ?"
	rows, err := c.db.QueryContext(ctx, query, issueID, issueID)
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
	query := useDB(database) +
		"SELECT diff_type, from_id, to_id, from_status, to_status, from_commit, to_commit " +
		"FROM dolt_diff(?, ?, ?)"
	rows, err := c.db.QueryContext(ctx, query, table, from, to)
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

// CountByStatus returns a map of status → count for issues in the database.
func (c *Client) CountByStatus(ctx context.Context, database string) (map[string]int, error) {
	query := useDB(database) +
		"SELECT status, COUNT(*) FROM issues GROUP BY status"
	rows, err := c.db.QueryContext(ctx, query)
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

// IssueDiffs returns issue-level changes between two revisions, including
// status, owner, and assignee fields needed for event extraction.
func (c *Client) IssueDiffs(ctx context.Context, database, from, to string) ([]IssueDiffRow, error) {
	query := useDB(database) +
		"SELECT diff_type, to_id, to_title, to_status, to_owner, to_assignee, " +
		"from_status, from_owner, from_assignee, to_commit_date " +
		"FROM dolt_diff(?, ?, ?) ORDER BY to_commit_date"
	rows, err := c.db.QueryContext(ctx, query, "issues", from, to)
	if err != nil {
		return nil, fmt.Errorf("dolt: issue diffs: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var diffs []IssueDiffRow
	for rows.Next() {
		var d IssueDiffRow
		if err := rows.Scan(
			&d.DiffType, &d.ToID, &d.ToTitle, &d.ToStatus, &d.ToOwner, &d.ToAssignee,
			&d.FromStatus, &d.FromOwner, &d.FromAssignee, &d.ToCommitDate,
		); err != nil {
			return nil, fmt.Errorf("dolt: scan issue diff: %w", err)
		}
		diffs = append(diffs, d)
	}
	return diffs, rows.Err()
}

// CommentDiffs returns comment-level changes between two revisions, including
// author and body fields needed for comment timeline extraction.
func (c *Client) CommentDiffs(ctx context.Context, database, from, to string) ([]CommentDiffRow, error) {
	query := useDB(database) +
		"SELECT diff_type, to_id, to_issue_id, to_author, to_body, to_commit_date " +
		"FROM dolt_diff(?, ?, ?) ORDER BY to_commit_date"
	rows, err := c.db.QueryContext(ctx, query, "comments", from, to)
	if err != nil {
		return nil, fmt.Errorf("dolt: comment diffs: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var diffs []CommentDiffRow
	for rows.Next() {
		var d CommentDiffRow
		if err := rows.Scan(
			&d.DiffType, &d.ToID, &d.ToIssueID, &d.ToAuthor, &d.ToBody, &d.ToCommitDate,
		); err != nil {
			return nil, fmt.Errorf("dolt: scan comment diff: %w", err)
		}
		diffs = append(diffs, d)
	}
	return diffs, rows.Err()
}

// buildIssueQuery constructs a SELECT for issues with optional filters
// and optional AS OF clause.
func buildIssueQuery(database string, f IssueFilter, asOf string) (string, []any) {
	var b strings.Builder
	var args []any

	b.WriteString(useDB(database))
	b.WriteString("SELECT id, title, description, status, priority, type, owner, assignee, labels, created_at, updated_at FROM ")

	if asOf != "" {
		b.WriteString(fmt.Sprintf("issues AS OF '%s'", asOf))
	} else {
		b.WriteString("issues")
	}

	var conditions []string
	if f.Status != "" {
		conditions = append(conditions, "status = ?")
		args = append(args, f.Status)
	}
	if f.Priority != 0 {
		conditions = append(conditions, "priority = ?")
		args = append(args, f.Priority)
	}
	if f.Type != "" {
		conditions = append(conditions, "type = ?")
		args = append(args, f.Type)
	}
	if f.Assignee != "" {
		conditions = append(conditions, "assignee = ?")
		args = append(args, f.Assignee)
	}

	if len(conditions) > 0 {
		b.WriteString(" WHERE ")
		b.WriteString(strings.Join(conditions, " AND "))
	}

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
		&iss.Labels, &iss.CreatedAt, &iss.UpdatedAt,
	)
}

// queryIssues executes a query and returns scanned issues.
func (c *Client) queryIssues(ctx context.Context, query string, args []any) ([]Issue, error) {
	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("dolt: query issues: %w", err)
	}
	defer func() { _ = rows.Close() }()

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
