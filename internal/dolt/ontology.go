package dolt

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"
)

// OntologyEntity represents a row from ontology_entities.
type OntologyEntity struct {
	ID        string
	Type      string
	Name      string
	Domain    string
	JSONLD    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// OntologyRelation represents a row from ontology_relations.
type OntologyRelation struct {
	SubjectID string
	Predicate string
	ObjectID  string
}

// OntologyDirective represents a row from ontology_directives.
type OntologyDirective struct {
	ID         string
	Rule       string
	IssuedBy   string
	IssuedDate *time.Time
	Priority   string
	BeadID     string
	CreatedAt  time.Time
}

// OntologyDecision represents a row from ontology_decisions.
type OntologyDecision struct {
	ID        string
	Decision  string
	Rationale string
	MadeBy    string
	Date      *time.Time
	BeadID    string
	CreatedAt time.Time
}

// OntologyTypeSummary holds count per entity type.
type OntologyTypeSummary struct {
	Type  string
	Count int
}

// queryFQ runs a query using the shared connection pool with fully-qualified
// table names (database.table) instead of USE + unqualified names. This avoids
// Dolt revision resolution issues where USE may not see recently created tables.
func (c *Client) queryFQ(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return c.db.QueryContext(ctx, query, args...)
}

// fqTable returns a fully-qualified table reference: `database`.`table`.
func fqTable(database, table string) string {
	return fmt.Sprintf("`%s`.`%s`", database, table)
}

// OntologyEntities returns all entities from ontology_entities, optionally
// filtered by type.
func (c *Client) OntologyEntities(ctx context.Context, database, filterType string) ([]OntologyEntity, error) {
	t := fqTable(database, "ontology_entities")
	q := fmt.Sprintf("SELECT id, type, name, COALESCE(domain,''), COALESCE(jsonld,'{}'), created_at, updated_at FROM %s", t)
	var args []any
	if filterType != "" {
		q += " WHERE type = ?"
		args = append(args, filterType)
	}
	q += " ORDER BY type, name"

	rows, err := c.queryFQ(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("ontology entities: %w", err)
	}
	defer rows.Close()

	var out []OntologyEntity
	for rows.Next() {
		var e OntologyEntity
		if err := rows.Scan(&e.ID, &e.Type, &e.Name, &e.Domain, &e.JSONLD, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, fmt.Errorf("ontology entities scan: %w", err)
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// OntologyRelations returns all relations from ontology_relations.
func (c *Client) OntologyRelations(ctx context.Context, database string) ([]OntologyRelation, error) {
	t := fqTable(database, "ontology_relations")
	rows, err := c.queryFQ(ctx,
		fmt.Sprintf("SELECT subject_id, predicate, object_id FROM %s ORDER BY predicate, subject_id", t))
	if err != nil {
		return nil, fmt.Errorf("ontology relations: %w", err)
	}
	defer rows.Close()

	var out []OntologyRelation
	for rows.Next() {
		var r OntologyRelation
		if err := rows.Scan(&r.SubjectID, &r.Predicate, &r.ObjectID); err != nil {
			return nil, fmt.Errorf("ontology relations scan: %w", err)
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// OntologyDirectives returns all directives from ontology_directives.
func (c *Client) OntologyDirectives(ctx context.Context, database string) ([]OntologyDirective, error) {
	t := fqTable(database, "ontology_directives")
	rows, err := c.queryFQ(ctx,
		fmt.Sprintf("SELECT id, rule, COALESCE(issued_by,''), issued_date, COALESCE(priority,''), COALESCE(bead_id,''), created_at FROM %s ORDER BY created_at DESC", t))
	if err != nil {
		return nil, fmt.Errorf("ontology directives: %w", err)
	}
	defer rows.Close()

	var out []OntologyDirective
	for rows.Next() {
		var d OntologyDirective
		if err := rows.Scan(&d.ID, &d.Rule, &d.IssuedBy, &d.IssuedDate, &d.Priority, &d.BeadID, &d.CreatedAt); err != nil {
			return nil, fmt.Errorf("ontology directives scan: %w", err)
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

// OntologyDecisions returns all decisions from ontology_decisions.
func (c *Client) OntologyDecisions(ctx context.Context, database string) ([]OntologyDecision, error) {
	t := fqTable(database, "ontology_decisions")
	rows, err := c.queryFQ(ctx,
		fmt.Sprintf("SELECT id, decision, COALESCE(rationale,''), COALESCE(made_by,''), date, COALESCE(bead_id,''), created_at FROM %s ORDER BY date DESC", t))
	if err != nil {
		return nil, fmt.Errorf("ontology decisions: %w", err)
	}
	defer rows.Close()

	var out []OntologyDecision
	for rows.Next() {
		var d OntologyDecision
		if err := rows.Scan(&d.ID, &d.Decision, &d.Rationale, &d.MadeBy, &d.Date, &d.BeadID, &d.CreatedAt); err != nil {
			return nil, fmt.Errorf("ontology decisions scan: %w", err)
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

// OntologyTypeSummaries returns entity counts grouped by type.
func (c *Client) OntologyTypeSummaries(ctx context.Context, database string) ([]OntologyTypeSummary, error) {
	t := fqTable(database, "ontology_entities")
	rows, err := c.queryFQ(ctx,
		fmt.Sprintf("SELECT type, COUNT(*) FROM %s GROUP BY type ORDER BY COUNT(*) DESC", t))
	if err != nil {
		return nil, fmt.Errorf("ontology type summaries: %w", err)
	}
	defer rows.Close()

	var out []OntologyTypeSummary
	for rows.Next() {
		var s OntologyTypeSummary
		if err := rows.Scan(&s.Type, &s.Count); err != nil {
			return nil, fmt.Errorf("ontology type summaries scan: %w", err)
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// OntologyImpact returns all entity IDs transitively dependent on the given
// entity (follows dependsOn and runningOn edges in reverse).
func (c *Client) OntologyImpact(ctx context.Context, database, entityID string) ([]OntologyEntity, error) {
	relTable := fqTable(database, "ontology_relations")
	entTable := fqTable(database, "ontology_entities")

	// Get direct dependents first, then recurse in Go. Max 5 hops.
	seen := map[string]bool{entityID: true}
	frontier := []string{entityID}

	for depth := 0; depth < 5 && len(frontier) > 0; depth++ {
		var next []string
		for _, eid := range frontier {
			rows, err := c.queryFQ(ctx,
				fmt.Sprintf("SELECT subject_id FROM %s WHERE object_id = ? AND predicate IN ('dependsOn','runningOn','runsService')", relTable), eid)
			if err != nil {
				return nil, fmt.Errorf("ontology impact: %w", err)
			}
			for rows.Next() {
				var sid string
				if err := rows.Scan(&sid); err != nil {
					rows.Close()
					return nil, fmt.Errorf("ontology impact scan: %w", err)
				}
				if !seen[sid] {
					seen[sid] = true
					next = append(next, sid)
				}
			}
			rows.Close()
		}
		frontier = next
	}

	// Fetch full entities for everything except the seed
	delete(seen, entityID)
	if len(seen) == 0 {
		return nil, nil
	}

	var out []OntologyEntity
	for id := range seen {
		rows, err := c.queryFQ(ctx,
			fmt.Sprintf("SELECT id, type, name, COALESCE(domain,''), COALESCE(jsonld,'{}'), created_at, updated_at FROM %s WHERE id = ?", entTable), id)
		if err != nil {
			return nil, fmt.Errorf("ontology impact entity: %w", err)
		}
		if rows.Next() {
			var e OntologyEntity
			if err := rows.Scan(&e.ID, &e.Type, &e.Name, &e.Domain, &e.JSONLD, &e.CreatedAt, &e.UpdatedAt); err != nil {
				rows.Close()
				return nil, fmt.Errorf("ontology impact entity scan: %w", err)
			}
			out = append(out, e)
		}
		rows.Close()
	}
	return out, nil
}

// hasOntologyTables returns true if the database has ontology_entities table.
func (c *Client) hasOntologyTables(ctx context.Context, database string) bool {
	t := fqTable(database, "ontology_entities")
	rows, err := c.queryFQ(ctx, fmt.Sprintf("SELECT 1 FROM %s LIMIT 1", t))
	if err != nil {
		log.Printf("ontology: hasOntologyTables(%s): %v", database, err)
		return false
	}
	defer rows.Close()
	return true
}

// FindOntologyDatabase returns the first database that has ontology tables.
// It checks "aegis" first (the primary ontology database), then falls back
// to scanning all beads databases.
func (c *Client) FindOntologyDatabase(ctx context.Context) (string, error) {
	// Fast path: check aegis directly
	if c.hasOntologyTables(ctx, "aegis") {
		return "aegis", nil
	}
	// Fallback: scan all databases
	all, err := c.ListDatabases(ctx, "")
	if err != nil {
		return "", err
	}
	for _, db := range all {
		if isSystemDatabase(db.Name) || db.Name == "aegis" {
			continue
		}
		if c.hasOntologyTables(ctx, db.Name) {
			return db.Name, nil
		}
	}
	return "", fmt.Errorf("no database with ontology tables found")
}
