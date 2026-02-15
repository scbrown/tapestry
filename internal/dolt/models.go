package dolt

import "time"

// Issue represents a beads issue row.
type Issue struct {
	ID          string
	Title       string
	Description string
	Status      string
	Priority    int
	Type        string
	Owner       string
	Assignee    string
	Labels      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Comment represents an issue comment row.
type Comment struct {
	ID        string
	IssueID   string
	Author    string
	Body      string
	CreatedAt time.Time
}

// Dependency represents a relationship between two issues.
type Dependency struct {
	FromID string
	ToID   string
	Type   string
}

// DiffRow represents a single row from a Dolt diff query.
type DiffRow struct {
	DiffType   string // "added", "modified", "removed"
	FromID     string
	ToID       string
	FromStatus string
	ToStatus   string
	FromCommit string
	ToCommit   string
}

// DatabaseInfo describes a beads database on the Dolt server.
type DatabaseInfo struct {
	Name string
}

// IssueFilter controls which issues are returned by a query.
type IssueFilter struct {
	Status   string // filter by status (empty = all)
	Priority int    // filter by priority (0 = all)
	Type     string // filter by type (empty = all)
	Assignee string // filter by assignee (empty = all)
	Limit    int    // max rows (0 = no limit)
}
