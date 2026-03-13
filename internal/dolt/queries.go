package dolt

import (
	"context"
	"encoding/json"
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
		"FROM issues WHERE id = ?"
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
	query := "SELECT issue_id, depends_on_id, type FROM dependencies WHERE issue_id = ? OR depends_on_id = ?"
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
	query := "SELECT status, COUNT(*) FROM issues " +
		"WHERE issue_type IN ('task','bug','epic') GROUP BY status"
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

// CountByPriorityStatus returns issue counts grouped by priority and status.
func (c *Client) CountByPriorityStatus(ctx context.Context, database string) ([]PriorityStatusCount, error) {
	query := "SELECT priority, status, COUNT(*) FROM issues " +
		"WHERE issue_type IN ('task','bug','epic') GROUP BY priority, status ORDER BY priority ASC, status"
	rows, err := c.queryDB(ctx, database, query)
	if err != nil {
		return nil, fmt.Errorf("dolt: count by priority status: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var result []PriorityStatusCount
	for rows.Next() {
		var psc PriorityStatusCount
		if err := rows.Scan(&psc.Priority, &psc.Status, &psc.Count); err != nil {
			return nil, fmt.Errorf("dolt: scan priority status count: %w", err)
		}
		result = append(result, psc)
	}
	return result, rows.Err()
}

// CountByAssigneeStatus returns issue counts grouped by assignee and status.
func (c *Client) CountByAssigneeStatus(ctx context.Context, database string) ([]AssigneeStatusCount, error) {
	query := "SELECT COALESCE(NULLIF(assignee,''), '(unassigned)'), status, COUNT(*) FROM issues " +
		"WHERE issue_type IN ('task','bug','epic') GROUP BY assignee, status ORDER BY assignee, status"
	rows, err := c.queryDB(ctx, database, query)
	if err != nil {
		return nil, fmt.Errorf("dolt: count by assignee status: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var result []AssigneeStatusCount
	for rows.Next() {
		var asc AssigneeStatusCount
		if err := rows.Scan(&asc.Assignee, &asc.Status, &asc.Count); err != nil {
			return nil, fmt.Errorf("dolt: scan assignee status count: %w", err)
		}
		result = append(result, asc)
	}
	return result, rows.Err()
}

// CountCreatedInRange counts issues created within the given time range.
func (c *Client) CountCreatedInRange(ctx context.Context, database string, start, end time.Time) (int, error) {
	query := "SELECT COUNT(*) FROM issues WHERE issue_type IN ('task','bug','epic') " +
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
	query := "SELECT COUNT(*) FROM issues WHERE issue_type IN ('task','bug','epic') " +
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

// AgentActivity returns aggregated stats per agent (by assignee field).
// Uses assignee rather than owner because owner contains legacy identities
// (sentinel@aegis.svc, scbrown3@gmail.com) while assignee has current
// Gas Town crew names (aegis/crew/arnold, etc.).
func (c *Client) AgentActivity(ctx context.Context, database string) ([]AgentStats, error) {
	query := `SELECT COALESCE(assignee,'(unowned)') AS agent,
		COUNT(*) AS total,
		SUM(CASE WHEN status = 'closed' THEN 1 ELSE 0 END) AS closed,
		SUM(CASE WHEN status = 'open' THEN 1 ELSE 0 END) AS open_count,
		SUM(CASE WHEN status IN ('in_progress','hooked') THEN 1 ELSE 0 END) AS in_progress
		FROM issues WHERE issue_type IN ('task','bug','epic')
		AND assignee IS NOT NULL AND assignee <> ''
		AND updated_at >= NOW() - INTERVAL 7 DAY
		GROUP BY assignee ORDER BY in_progress DESC, total DESC`
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

// AgentActivityInRange returns the number of issues updated per assignee within
// the given time range. Only issues with a non-empty assignee are counted.
func (c *Client) AgentActivityInRange(ctx context.Context, database string, from, to time.Time) (map[string]int, error) {
	query := "SELECT assignee, COUNT(*) FROM issues " +
		"WHERE issue_type IN ('task','bug','epic') " +
		"AND updated_at >= ? AND updated_at < ? AND assignee != '' " +
		"GROUP BY assignee ORDER BY COUNT(*) DESC"
	rows, err := c.queryDB(ctx, database, query, from, to)
	if err != nil {
		return nil, fmt.Errorf("dolt: agent activity in range: %w", err)
	}
	defer func() { _ = rows.Close() }()

	activity := make(map[string]int)
	for rows.Next() {
		var agent string
		var count int
		if err := rows.Scan(&agent, &count); err != nil {
			return nil, fmt.Errorf("dolt: scan agent activity: %w", err)
		}
		activity[agent] = count
	}
	return activity, rows.Err()
}

// SearchIssues returns issues where the title or description matches the query string.
func (c *Client) SearchIssues(ctx context.Context, database, q string, limit int) ([]Issue, error) {
	if q == "" {
		return nil, nil
	}
	pattern := "%" + q + "%"
	query := "SELECT id, title, description, status, priority, issue_type, " +
		"COALESCE(owner,''), COALESCE(assignee,''), created_at, updated_at " +
		"FROM issues WHERE issue_type IN ('task','bug','epic') " +
		"AND (id LIKE ? OR title LIKE ? OR description LIKE ?) " +
		"ORDER BY updated_at DESC"
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}
	rows, err := c.queryDB(ctx, database, query, pattern, pattern, pattern)
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
	query := "SELECT DISTINCT COALESCE(assignee,'') FROM issues " +
		"WHERE issue_type IN ('task','bug','epic') AND assignee IS NOT NULL AND assignee != '' " +
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

// Decisions returns all issues of type "decision" from the database.
func (c *Client) Decisions(ctx context.Context, database string) ([]Issue, error) {
	query := "SELECT id, title, description, status, priority, issue_type, " +
		"COALESCE(owner,''), COALESCE(assignee,''), created_at, updated_at " +
		"FROM issues WHERE issue_type = 'decision' " +
		"ORDER BY updated_at DESC"
	rows, err := c.queryDB(ctx, database, query)
	if err != nil {
		return nil, fmt.Errorf("dolt: decisions: %w", err)
	}
	defer func() { _ = rows.Close() }()
	return scanIssues(rows)
}

// LabelsForIssue returns all labels attached to an issue.
func (c *Client) LabelsForIssue(ctx context.Context, database, issueID string) ([]string, error) {
	query := "SELECT label FROM labels WHERE issue_id = ? ORDER BY label"
	rows, err := c.queryDB(ctx, database, query, issueID)
	if err != nil {
		return nil, fmt.Errorf("dolt: labels: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var labels []string
	for rows.Next() {
		var l string
		if err := rows.Scan(&l); err != nil {
			return nil, fmt.Errorf("dolt: scan label: %w", err)
		}
		labels = append(labels, l)
	}
	return labels, rows.Err()
}

// DistinctLabels returns all unique labels with issue counts, ordered by count descending.
func (c *Client) DistinctLabels(ctx context.Context, database string) ([]LabelCount, error) {
	query := "SELECT label, COUNT(*) as cnt FROM labels GROUP BY label ORDER BY cnt DESC, label ASC"
	rows, err := c.queryDB(ctx, database, query)
	if err != nil {
		return nil, fmt.Errorf("dolt: distinct labels: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var result []LabelCount
	for rows.Next() {
		var lc LabelCount
		if err := rows.Scan(&lc.Label, &lc.Count); err != nil {
			return nil, fmt.Errorf("dolt: scan label count: %w", err)
		}
		result = append(result, lc)
	}
	return result, rows.Err()
}

// IssuesByLabel returns all non-closed issues with the given label.
func (c *Client) IssuesByLabel(ctx context.Context, database, label string) ([]Issue, error) {
	query := `SELECT i.id, i.title, i.status, i.priority, i.issue_type,
		COALESCE(i.owner,''), COALESCE(i.assignee,''),
		i.created_at, i.updated_at, COALESCE(i.description,'')
		FROM issues i JOIN labels l ON l.issue_id = i.id
		WHERE l.label = ?
		ORDER BY i.priority ASC, i.updated_at DESC`
	rows, err := c.queryDB(ctx, database, query, label)
	if err != nil {
		return nil, fmt.Errorf("dolt: issues by label: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var result []Issue
	for rows.Next() {
		var iss Issue
		if err := rows.Scan(&iss.ID, &iss.Title, &iss.Status, &iss.Priority, &iss.Type,
			&iss.Owner, &iss.Assignee,
			&iss.CreatedAt, &iss.UpdatedAt, &iss.Description); err != nil {
			return nil, fmt.Errorf("dolt: scan issue by label: %w", err)
		}
		result = append(result, iss)
	}
	return result, rows.Err()
}

// MetadataForIssue returns parsed metadata for the given issue from the JSON
// metadata column. Returns an empty IssueMetadata (not nil) if not found or empty.
func (c *Client) MetadataForIssue(ctx context.Context, database, issueID string) (*IssueMetadata, error) {
	query := "SELECT COALESCE(metadata, '{}') FROM issues WHERE id = ?"
	rows, err := c.queryDB(ctx, database, query, issueID)
	if err != nil {
		return &IssueMetadata{}, fmt.Errorf("dolt: metadata: %w", err)
	}
	defer func() { _ = rows.Close() }()

	if !rows.Next() {
		return &IssueMetadata{}, rows.Err()
	}
	var raw string
	if err := rows.Scan(&raw); err != nil {
		return &IssueMetadata{}, fmt.Errorf("dolt: scan metadata: %w", err)
	}
	var meta IssueMetadata
	if err := json.Unmarshal([]byte(raw), &meta); err != nil {
		return &IssueMetadata{}, nil // ignore malformed JSON
	}
	return &meta, nil
}

// ChildIssues returns issues that are children of the given parent (epic).
func (c *Client) ChildIssues(ctx context.Context, database, parentID string) ([]Issue, error) {
	query := "SELECT i.id, i.title, i.description, i.status, i.priority, i.issue_type, " +
		"COALESCE(i.owner,''), COALESCE(i.assignee,''), i.created_at, i.updated_at " +
		"FROM issues i JOIN dependencies d ON i.id = d.issue_id " +
		"WHERE d.depends_on_id = ? AND d.type = 'child_of' " +
		"ORDER BY i.priority, i.id"
	rows, err := c.queryDB(ctx, database, query, parentID)
	if err != nil {
		return nil, fmt.Errorf("dolt: child issues: %w", err)
	}
	defer func() { _ = rows.Close() }()
	return scanIssues(rows)
}

// AddLabel inserts a label for an issue. Ignores duplicates.
func (c *Client) AddLabel(ctx context.Context, database, issueID, label string) error {
	query := "INSERT IGNORE INTO labels (issue_id, label) VALUES (?, ?)"
	conn, err := c.db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("dolt: conn: %w", err)
	}
	defer func() { _ = conn.Close() }()
	if _, err := conn.ExecContext(ctx, fmt.Sprintf("USE `%s`", database)); err != nil {
		return fmt.Errorf("dolt: use %s: %w", database, err)
	}
	if _, err := conn.ExecContext(ctx, query, issueID, label); err != nil {
		return fmt.Errorf("dolt: add label: %w", err)
	}
	return nil
}

// RemoveLabel deletes a label from an issue.
func (c *Client) RemoveLabel(ctx context.Context, database, issueID, label string) error {
	query := "DELETE FROM labels WHERE issue_id = ? AND label = ?"
	conn, err := c.db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("dolt: conn: %w", err)
	}
	defer func() { _ = conn.Close() }()
	if _, err := conn.ExecContext(ctx, fmt.Sprintf("USE `%s`", database)); err != nil {
		return fmt.Errorf("dolt: use %s: %w", database, err)
	}
	if _, err := conn.ExecContext(ctx, query, issueID, label); err != nil {
		return fmt.Errorf("dolt: remove label: %w", err)
	}
	return nil
}

// AddComment inserts a new comment on an issue.
func (c *Client) AddComment(ctx context.Context, database, issueID, author, body string) error {
	query := "INSERT INTO comments (issue_id, author, text, created_at) VALUES (?, ?, ?, NOW())"
	conn, err := c.db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("dolt: conn: %w", err)
	}
	defer func() { _ = conn.Close() }()
	if _, err := conn.ExecContext(ctx, fmt.Sprintf("USE `%s`", database)); err != nil {
		return fmt.Errorf("dolt: use %s: %w", database, err)
	}
	if _, err := conn.ExecContext(ctx, query, issueID, author, body); err != nil {
		return fmt.Errorf("dolt: add comment: %w", err)
	}
	return nil
}

// UpdateStatus sets the status of an issue.
func (c *Client) UpdateStatus(ctx context.Context, database, issueID, status string) error {
	query := "UPDATE issues SET status = ?, updated_at = NOW() WHERE id = ?"
	conn, err := c.db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("dolt: conn: %w", err)
	}
	defer func() { _ = conn.Close() }()
	if _, err := conn.ExecContext(ctx, fmt.Sprintf("USE `%s`", database)); err != nil {
		return fmt.Errorf("dolt: use %s: %w", database, err)
	}
	if _, err := conn.ExecContext(ctx, query, status, issueID); err != nil {
		return fmt.Errorf("dolt: update status: %w", err)
	}
	return nil
}

func (c *Client) UpdatePriority(ctx context.Context, database, issueID string, priority int) error {
	query := "UPDATE issues SET priority = ?, updated_at = NOW() WHERE id = ?"
	conn, err := c.db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("dolt: conn: %w", err)
	}
	defer func() { _ = conn.Close() }()
	if _, err := conn.ExecContext(ctx, fmt.Sprintf("USE `%s`", database)); err != nil {
		return fmt.Errorf("dolt: use %s: %w", database, err)
	}
	if _, err := conn.ExecContext(ctx, query, priority, issueID); err != nil {
		return fmt.Errorf("dolt: update priority: %w", err)
	}
	return nil
}

func (c *Client) UpdateAssignee(ctx context.Context, database, issueID, assignee string) error {
	query := "UPDATE issues SET assignee = ?, updated_at = NOW() WHERE id = ?"
	conn, err := c.db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("dolt: conn: %w", err)
	}
	defer func() { _ = conn.Close() }()
	if _, err := conn.ExecContext(ctx, fmt.Sprintf("USE `%s`", database)); err != nil {
		return fmt.Errorf("dolt: use %s: %w", database, err)
	}
	if _, err := conn.ExecContext(ctx, query, assignee, issueID); err != nil {
		return fmt.Errorf("dolt: update assignee: %w", err)
	}
	return nil
}

// BlockedIssues returns issues that have unresolved depends_on dependencies.
// An issue is "blocked" if it depends on at least one non-closed issue.
func (c *Client) BlockedIssues(ctx context.Context, database string) ([]BlockedIssue, error) {
	query := `SELECT i.id, i.title, i.status, i.priority, i.issue_type,
		COALESCE(i.owner,''), COALESCE(i.assignee,''),
		i.created_at, i.updated_at,
		blocker.id, blocker.title, blocker.status,
		COALESCE(blocker.owner,''), COALESCE(blocker.assignee,'')
		FROM dependencies d
		JOIN issues i ON i.id = d.issue_id
		JOIN issues blocker ON blocker.id = d.depends_on_id
		WHERE d.type = 'depends_on'
		AND i.status != 'closed'
		AND blocker.status != 'closed'
		AND i.issue_type IN ('task','bug','epic')
		ORDER BY i.priority ASC, i.updated_at DESC`
	rows, err := c.queryDB(ctx, database, query)
	if err != nil {
		return nil, fmt.Errorf("dolt: blocked issues: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var result []BlockedIssue
	for rows.Next() {
		var bi BlockedIssue
		if err := rows.Scan(
			&bi.Issue.ID, &bi.Issue.Title, &bi.Issue.Status,
			&bi.Issue.Priority, &bi.Issue.Type,
			&bi.Issue.Owner, &bi.Issue.Assignee,
			&bi.Issue.CreatedAt, &bi.Issue.UpdatedAt,
			&bi.Blocker.ID, &bi.Blocker.Title, &bi.Blocker.Status,
			&bi.Blocker.Owner, &bi.Blocker.Assignee,
		); err != nil {
			return nil, fmt.Errorf("dolt: scan blocked issue: %w", err)
		}
		result = append(result, bi)
	}
	return result, rows.Err()
}

// AllChildDependencies returns all child_of dependency edges in the database.
// Used for building task hierarchy trees efficiently without N+1 queries.
func (c *Client) AllChildDependencies(ctx context.Context, database string) ([]Dependency, error) {
	query := "SELECT issue_id, depends_on_id, type FROM dependencies WHERE type = 'child_of'"
	rows, err := c.queryDB(ctx, database, query)
	if err != nil {
		return nil, fmt.Errorf("dolt: all child deps: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var deps []Dependency
	for rows.Next() {
		var d Dependency
		if err := rows.Scan(&d.FromID, &d.ToID, &d.Type); err != nil {
			return nil, fmt.Errorf("dolt: scan child dep: %w", err)
		}
		deps = append(deps, d)
	}
	return deps, rows.Err()
}

// AllDependenciesWithIssues returns all dependency edges with issue details for both sides.
func (c *Client) AllDependenciesWithIssues(ctx context.Context, database string) ([]DepEdge, error) {
	query := `SELECT d.issue_id, i1.title, i1.status, i1.priority, i1.issue_type,
		COALESCE(i1.owner,''), COALESCE(i1.assignee,''), i1.created_at, i1.updated_at,
		d.depends_on_id, i2.title, i2.status, i2.priority, i2.issue_type,
		COALESCE(i2.owner,''), COALESCE(i2.assignee,''), i2.created_at, i2.updated_at,
		d.type
		FROM dependencies d
		JOIN issues i1 ON i1.id = d.issue_id
		JOIN issues i2 ON i2.id = d.depends_on_id
		ORDER BY d.type, i1.priority ASC, i1.updated_at DESC`
	rows, err := c.queryDB(ctx, database, query)
	if err != nil {
		return nil, fmt.Errorf("dolt: all deps with issues: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var edges []DepEdge
	for rows.Next() {
		var e DepEdge
		if err := rows.Scan(
			&e.From.ID, &e.From.Title, &e.From.Status, &e.From.Priority, &e.From.Type,
			&e.From.Owner, &e.From.Assignee, &e.From.CreatedAt, &e.From.UpdatedAt,
			&e.To.ID, &e.To.Title, &e.To.Status, &e.To.Priority, &e.To.Type,
			&e.To.Owner, &e.To.Assignee, &e.To.CreatedAt, &e.To.UpdatedAt,
			&e.Type,
		); err != nil {
			return nil, fmt.Errorf("dolt: scan dep edge: %w", err)
		}
		edges = append(edges, e)
	}
	return edges, rows.Err()
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

	var conditions []string
	if f.Type != "" {
		conditions = append(conditions, "issue_type = ?")
		args = append(args, f.Type)
	} else {
		conditions = append(conditions, "issue_type IN ('task','bug','epic','feature','decision')")
	}
	if f.Status != "" {
		conditions = append(conditions, "status = ?")
		args = append(args, f.Status)
	}
	if f.Priority != 0 {
		conditions = append(conditions, "priority = ?")
		args = append(args, f.Priority)
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

// AchievementDefs returns all achievement definitions ordered by sort_order.
func (c *Client) AchievementDefs(ctx context.Context, database string) ([]AchievementDef, error) {
	query := `SELECT id, name, description, icon, category, trigger_expr, sort_order
		FROM achievement_defs ORDER BY sort_order`
	rows, err := c.queryDB(ctx, database, query)
	if err != nil {
		return nil, fmt.Errorf("dolt: achievement defs: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var defs []AchievementDef
	for rows.Next() {
		var d AchievementDef
		if err := rows.Scan(&d.ID, &d.Name, &d.Description, &d.Icon, &d.Category, &d.TriggerExpr, &d.SortOrder); err != nil {
			return nil, fmt.Errorf("dolt: scan achievement def: %w", err)
		}
		defs = append(defs, d)
	}
	return defs, rows.Err()
}

// AchievementUnlocks returns all unlocked achievements.
func (c *Client) AchievementUnlocks(ctx context.Context, database string) ([]AchievementUnlock, error) {
	query := `SELECT id, unlocked_at, COALESCE(unlocked_by,''), COALESCE(note,'')
		FROM achievement_unlocks ORDER BY unlocked_at DESC`
	rows, err := c.queryDB(ctx, database, query)
	if err != nil {
		return nil, fmt.Errorf("dolt: achievement unlocks: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var unlocks []AchievementUnlock
	for rows.Next() {
		var u AchievementUnlock
		if err := rows.Scan(&u.ID, &u.UnlockedAt, &u.UnlockedBy, &u.Note); err != nil {
			return nil, fmt.Errorf("dolt: scan achievement unlock: %w", err)
		}
		unlocks = append(unlocks, u)
	}
	return unlocks, rows.Err()
}

// ThemeParks returns all theme parks ordered by name.
func (c *Client) ThemeParks(ctx context.Context, database string) ([]ThemePark, error) {
	query := `SELECT id, name, COALESCE(location,''), COALESCE(region,''),
		COALESCE(website,''), COALESCE(notes,''), COALESCE(rating,0),
		COALESCE(visited,0), COALESCE(wishlist,0), created_at, updated_at
		FROM theme_parks ORDER BY name`
	rows, err := c.queryDB(ctx, database, query)
	if err != nil {
		return nil, fmt.Errorf("dolt: theme parks: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var parks []ThemePark
	for rows.Next() {
		var p ThemePark
		if err := rows.Scan(&p.ID, &p.Name, &p.Location, &p.Region,
			&p.Website, &p.Notes, &p.Rating,
			&p.Visited, &p.Wishlist, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("dolt: scan theme park: %w", err)
		}
		parks = append(parks, p)
	}
	return parks, rows.Err()
}

// Rides returns rides for a specific park, or all rides if parkID is empty.
func (c *Client) Rides(ctx context.Context, database, parkID string) ([]Ride, error) {
	query := `SELECT id, park_id, name, COALESCE(type,''), COALESCE(thrill_level,0),
		COALESCE(height_req_inches,0), COALESCE(notes,''), COALESCE(rating,0),
		COALESCE(ridden,0), COALESCE(wishlist,0), created_at, updated_at
		FROM rides`
	var args []any
	if parkID != "" {
		query += " WHERE park_id = ?"
		args = append(args, parkID)
	}
	query += " ORDER BY name"
	rows, err := c.queryDB(ctx, database, query, args...)
	if err != nil {
		return nil, fmt.Errorf("dolt: rides: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var rides []Ride
	for rows.Next() {
		var r Ride
		if err := rows.Scan(&r.ID, &r.ParkID, &r.Name, &r.Type, &r.ThrillLevel,
			&r.HeightReqInches, &r.Notes, &r.Rating,
			&r.Ridden, &r.Wishlist, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, fmt.Errorf("dolt: scan ride: %w", err)
		}
		rides = append(rides, r)
	}
	return rides, rows.Err()
}

// ParkVisits returns visit history, optionally filtered by park.
func (c *Client) ParkVisits(ctx context.Context, database, parkID string) ([]ParkVisit, error) {
	query := `SELECT v.id, v.park_id, p.name, v.visit_date,
		COALESCE(v.party_size,0), COALESCE(v.weather,''), COALESCE(v.crowd_level,''),
		COALESCE(v.highlights,''), COALESCE(v.notes,''), COALESCE(v.overall_rating,0),
		v.created_at
		FROM park_visits v JOIN theme_parks p ON v.park_id = p.id`
	var args []any
	if parkID != "" {
		query += " WHERE v.park_id = ?"
		args = append(args, parkID)
	}
	query += " ORDER BY v.visit_date DESC"
	rows, err := c.queryDB(ctx, database, query, args...)
	if err != nil {
		return nil, fmt.Errorf("dolt: park visits: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var visits []ParkVisit
	for rows.Next() {
		var v ParkVisit
		if err := rows.Scan(&v.ID, &v.ParkID, &v.ParkName, &v.VisitDate,
			&v.PartySize, &v.Weather, &v.CrowdLevel,
			&v.Highlights, &v.Notes, &v.OverallRating, &v.CreatedAt); err != nil {
			return nil, fmt.Errorf("dolt: scan park visit: %w", err)
		}
		visits = append(visits, v)
	}
	return visits, rows.Err()
}

// TripPlans returns upcoming trip plans.
func (c *Client) TripPlans(ctx context.Context, database string) ([]TripPlan, error) {
	query := `SELECT t.id, t.park_id, p.name, COALESCE(t.planned_date, CURRENT_DATE),
		COALESCE(t.status,'planned'), COALESCE(t.priority_rides,''),
		COALESCE(t.budget_estimate,0), COALESCE(t.notes,''),
		t.created_at, t.updated_at
		FROM trip_plans t JOIN theme_parks p ON t.park_id = p.id
		ORDER BY t.planned_date`
	rows, err := c.queryDB(ctx, database, query)
	if err != nil {
		return nil, fmt.Errorf("dolt: trip plans: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var plans []TripPlan
	for rows.Next() {
		var t TripPlan
		if err := rows.Scan(&t.ID, &t.ParkID, &t.ParkName, &t.PlannedDate,
			&t.Status, &t.PriorityRides, &t.BudgetEstimate,
			&t.Notes, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, fmt.Errorf("dolt: scan trip plan: %w", err)
		}
		plans = append(plans, t)
	}
	return plans, rows.Err()
}

// RecentComments returns the most recent comments across all issues in a database.
func (c *Client) RecentComments(ctx context.Context, database string, limit int) ([]Comment, error) {
	if limit <= 0 {
		limit = 50
	}
	query := "SELECT id, issue_id, author, text, created_at " +
		"FROM comments ORDER BY created_at DESC LIMIT ?"
	rows, err := c.queryDB(ctx, database, query, limit)
	if err != nil {
		return nil, fmt.Errorf("dolt: recent comments: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var comments []Comment
	for rows.Next() {
		var cm Comment
		if err := rows.Scan(&cm.ID, &cm.IssueID, &cm.Author, &cm.Body, &cm.CreatedAt); err != nil {
			return nil, fmt.Errorf("dolt: scan recent comment: %w", err)
		}
		comments = append(comments, cm)
	}
	return comments, rows.Err()
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
