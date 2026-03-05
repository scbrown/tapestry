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
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Rig         string // database name, set by caller
}

// Comment represents an issue comment row.
type Comment struct {
	ID        int64
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

// IssueDiffRow represents a row from dolt_diff('issues', ...) with fields
// needed for bead lifecycle event extraction.
type IssueDiffRow struct {
	DiffType     string // "added", "modified", "removed"
	ToID         string
	ToTitle      string
	ToStatus     string
	ToOwner      string
	ToAssignee   string
	FromStatus   string
	FromOwner    string
	FromAssignee string
	ToCommitDate time.Time
}

// CommentDiffRow represents a row from dolt_diff('comments', ...) with fields
// needed for comment timeline event extraction.
type CommentDiffRow struct {
	DiffType     string // "added", "modified", "removed"
	ToID         string
	ToIssueID    string
	ToAuthor     string
	ToBody       string
	ToCommitDate time.Time
}

// DatabaseInfo describes a beads database on the Dolt server.
type DatabaseInfo struct {
	Name string
}

// IssueFilter controls which issues are returned by a query.
type IssueFilter struct {
	Status        string    // filter by status (empty = all)
	Priority      int       // filter by priority (0 = all)
	Type          string    // filter by type (empty = all)
	Assignee      string    // filter by assignee (empty = all)
	Owner         string    // filter by owner (empty = all)
	Limit         int       // max rows (0 = no limit)
	UpdatedAfter  time.Time // only issues updated after this time (zero = no filter)
	UpdatedBefore time.Time // only issues updated before this time (zero = no filter)
}

// AgentStats holds per-agent issue statistics from a database.
type AgentStats struct {
	Name       string
	Owned      int
	Closed     int
	Open       int
	InProgress int
}

// StatusTransition represents a status change for an issue over time.
type StatusTransition struct {
	FromStatus string
	ToStatus   string
	CommitDate time.Time
}

// EpicProgress tracks completion of an epic and its children.
type EpicProgress struct {
	Total  int
	Closed int
}

// BlockedIssue pairs a blocked issue with the issue that blocks it.
type BlockedIssue struct {
	Issue   Issue
	Blocker Issue
}

// AchievementDef describes an achievement that can be unlocked.
type AchievementDef struct {
	ID          string
	Name        string
	Description string
	Icon        string
	Category    string
	TriggerExpr string
	SortOrder   int
}

// AchievementUnlock records when an achievement was unlocked.
type AchievementUnlock struct {
	ID         string
	UnlockedAt time.Time
	UnlockedBy string
	Note       string
}
