package dolt

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// Rows wraps sql.Rows with connection lifecycle management.
// When Close() is called, both the rows and the underlying connection
// are returned to the pool.
type Rows struct {
	*sql.Rows
	conn *sql.Conn
}

// Close closes the rows and returns the connection to the pool.
func (r *Rows) Close() error {
	rowsErr := r.Rows.Close()
	connErr := r.conn.Close()
	if rowsErr != nil {
		return rowsErr
	}
	return connErr
}

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
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
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

// isSystemDatabase returns true for databases that never contain beads data.
func isSystemDatabase(name string) bool {
	switch name {
	case "information_schema", "mysql":
		return true
	}
	return strings.HasPrefix(name, "dolt_")
}

// isLegacyDatabase returns true for pre-Gas Town databases that contain
// stale data with incorrect agent names and shouldn't be shown in Tapestry.
func isLegacyDatabase(name string) bool {
	switch name {
	case "beads":
		return true
	}
	return false
}

// hasIssuesTable checks whether a database has an issues table.
func (c *Client) hasIssuesTable(ctx context.Context, database string) bool {
	rows, err := c.db.QueryContext(ctx,
		fmt.Sprintf("SHOW TABLES FROM `%s` LIKE 'issues'", database))
	if err != nil {
		return false
	}
	defer func() { _ = rows.Close() }()
	return rows.Next()
}

// ListBeadsDatabases returns all databases containing beads data.
// It discovers databases with the "beads_" prefix automatically and also
// probes other non-system databases for an issues table (legacy databases
// like "aegis" or "gastown" that predate the beads_ naming convention).
func (c *Client) ListBeadsDatabases(ctx context.Context) ([]DatabaseInfo, error) {
	all, err := c.ListDatabases(ctx, "")
	if err != nil {
		return nil, err
	}
	var result []DatabaseInfo
	for _, db := range all {
		if isLegacyDatabase(db.Name) {
			continue
		}
		if strings.HasPrefix(db.Name, "beads_") {
			result = append(result, db)
			continue
		}
		if isSystemDatabase(db.Name) {
			continue
		}
		if c.hasIssuesTable(ctx, db.Name) {
			result = append(result, db)
		}
	}
	return result, nil
}

// queryDB gets a dedicated connection, switches to the database, and runs
// the query. The returned *Rows closes the connection when Close() is called.
func (c *Client) queryDB(ctx context.Context, database, query string, args ...any) (*Rows, error) {
	conn, err := c.db.Conn(ctx)
	if err != nil {
		return nil, fmt.Errorf("dolt: conn: %w", err)
	}
	// Switch database on this connection
	if _, err := conn.ExecContext(ctx, fmt.Sprintf("USE `%s`", database)); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("dolt: use %s: %w", database, err)
	}
	rows, err := conn.QueryContext(ctx, query, args...)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	return &Rows{Rows: rows, conn: conn}, nil
}
