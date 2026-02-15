package dolt

import (
	"context"
	"fmt"
	"time"
)

// IssueDiff returns issue-level changes between two revisions using Dolt's
// dolt_diff table function. Revisions can be commit hashes, branch names,
// or timestamps formatted as "2006-01-02T15:04:05".
func (c *Client) IssueDiff(ctx context.Context, database, from, to string) ([]IssueDiffRow, error) {
	query := `SELECT diff_type,
		COALESCE(to_id,''), COALESCE(to_title,''), COALESCE(to_status,''),
		COALESCE(to_owner,''), COALESCE(to_assignee,''),
		COALESCE(from_status,''), COALESCE(from_owner,''), COALESCE(from_assignee,''),
		to_commit_date
		FROM dolt_diff(?, ?, ?)`
	rows, err := c.queryDB(ctx, database, query, "issues", from, to)
	if err != nil {
		return nil, fmt.Errorf("dolt: issue diff: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var diffs []IssueDiffRow
	for rows.Next() {
		var d IssueDiffRow
		if err := rows.Scan(
			&d.DiffType, &d.ToID, &d.ToTitle, &d.ToStatus,
			&d.ToOwner, &d.ToAssignee,
			&d.FromStatus, &d.FromOwner, &d.FromAssignee,
			&d.ToCommitDate,
		); err != nil {
			return nil, fmt.Errorf("dolt: scan issue diff: %w", err)
		}
		diffs = append(diffs, d)
	}
	return diffs, rows.Err()
}

// CommentDiff returns comment-level changes between two revisions using
// Dolt's dolt_diff table function.
func (c *Client) CommentDiff(ctx context.Context, database, from, to string) ([]CommentDiffRow, error) {
	query := `SELECT diff_type,
		COALESCE(to_id,''), COALESCE(to_issue_id,''),
		COALESCE(to_author,''), COALESCE(to_text,''),
		to_commit_date
		FROM dolt_diff(?, ?, ?)`
	rows, err := c.queryDB(ctx, database, query, "comments", from, to)
	if err != nil {
		return nil, fmt.Errorf("dolt: comment diff: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var diffs []CommentDiffRow
	for rows.Next() {
		var d CommentDiffRow
		if err := rows.Scan(
			&d.DiffType, &d.ToID, &d.ToIssueID,
			&d.ToAuthor, &d.ToBody,
			&d.ToCommitDate,
		); err != nil {
			return nil, fmt.Errorf("dolt: scan comment diff: %w", err)
		}
		diffs = append(diffs, d)
	}
	return diffs, rows.Err()
}

// IssueDiffSince returns issue changes since a given timestamp.
func (c *Client) IssueDiffSince(ctx context.Context, database string, since time.Time) ([]IssueDiffRow, error) {
	from := since.UTC().Format("2006-01-02T15:04:05")
	return c.IssueDiff(ctx, database, from, "HEAD")
}

// CommentDiffSince returns comment changes since a given timestamp.
func (c *Client) CommentDiffSince(ctx context.Context, database string, since time.Time) ([]CommentDiffRow, error) {
	from := since.UTC().Format("2006-01-02T15:04:05")
	return c.CommentDiff(ctx, database, from, "HEAD")
}
