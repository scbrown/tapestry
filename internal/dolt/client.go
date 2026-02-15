package dolt

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

// Client manages a connection pool to a Dolt server and provides
// typed query methods for beads databases.
type Client struct {
	db  *sql.DB
	cfg Config
}

// New opens a connection pool to the Dolt server described by cfg.
func New(cfg Config) (*Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	db, err := sql.Open("mysql", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("dolt: open: %w", err)
	}
	return &Client{db: db, cfg: cfg}, nil
}

// Ping verifies the server is reachable.
func (c *Client) Ping(ctx context.Context) error {
	return c.db.PingContext(ctx)
}

// Close closes the underlying connection pool.
func (c *Client) Close() error {
	return c.db.Close()
}

// DB returns the underlying *sql.DB for advanced use cases.
func (c *Client) DB() *sql.DB {
	return c.db
}

// ListDatabases returns all database names on the server that start with
// the given prefix. Pass "" to list all databases.
func (c *Client) ListDatabases(ctx context.Context, prefix string) ([]DatabaseInfo, error) {
	rows, err := c.db.QueryContext(ctx, "SHOW DATABASES")
	if err != nil {
		return nil, fmt.Errorf("dolt: list databases: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var dbs []DatabaseInfo
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("dolt: scan database: %w", err)
		}
		if prefix == "" || strings.HasPrefix(name, prefix) {
			dbs = append(dbs, DatabaseInfo{Name: name})
		}
	}
	return dbs, rows.Err()
}

// ListBeadsDatabases returns databases whose names start with "beads_".
func (c *Client) ListBeadsDatabases(ctx context.Context) ([]DatabaseInfo, error) {
	return c.ListDatabases(ctx, "beads_")
}

// useDB returns a query prefix that switches to the given database.
// Callers must use this in the same connection (same transaction or
// multiStatements DSN).
func useDB(db string) string {
	return fmt.Sprintf("USE `%s`; ", db)
}
