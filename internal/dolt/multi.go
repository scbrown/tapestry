package dolt

import (
	"context"
	"errors"
	"fmt"
)

// ForEachDatabase calls fn for each beads_* database on the server.
// Errors from individual databases are collected; iteration continues
// through all databases. Returns a joined error if any calls failed.
func (c *Client) ForEachDatabase(ctx context.Context, fn func(ctx context.Context, database string) error) error {
	dbs, err := c.ListBeadsDatabases(ctx)
	if err != nil {
		return fmt.Errorf("dolt: list databases: %w", err)
	}
	var errs []error
	for _, db := range dbs {
		if err := fn(ctx, db.Name); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", db.Name, err))
		}
	}
	return errors.Join(errs...)
}

// FindIssue searches for an issue by ID across all beads_* databases.
// Returns the issue and the database it was found in, or nil if not found.
func (c *Client) FindIssue(ctx context.Context, id string) (*Issue, string, error) {
	dbs, err := c.ListBeadsDatabases(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("dolt: list databases: %w", err)
	}
	for _, db := range dbs {
		issue, err := c.IssueByID(ctx, db.Name, id)
		if err != nil {
			continue // skip databases with errors
		}
		if issue != nil {
			return issue, db.Name, nil
		}
	}
	return nil, "", nil
}

// EpicChildren returns the child issues and progress for the given epic.
func (c *Client) EpicChildren(ctx context.Context, database, epicID string) ([]Issue, EpicProgress, error) {
	childIDs, err := c.EpicChildIDs(ctx, database, epicID)
	if err != nil {
		return nil, EpicProgress{}, err
	}
	var children []Issue
	var progress EpicProgress
	for _, cid := range childIDs {
		child, err := c.IssueByID(ctx, database, cid)
		if err != nil || child == nil {
			continue
		}
		children = append(children, *child)
		progress.Total++
		if child.Status == "closed" {
			progress.Closed++
		}
	}
	return children, progress, nil
}
