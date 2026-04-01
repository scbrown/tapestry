package web

import (
	"log"
	"net/http"
	"sort"
	"strings"

	"github.com/scbrown/tapestry/internal/dolt"
)

type ontologyData struct {
	Entities   []dolt.OntologyEntity
	Relations  []dolt.OntologyRelation
	Directives []dolt.OntologyDirective
	Decisions  []dolt.OntologyDecision
	Types      []dolt.OntologyTypeSummary
	Impact     []dolt.OntologyEntity
	ImpactSeed string

	// Graph data (JSON for JS visualization)
	GraphNodes string
	GraphEdges string

	// Filters
	FilterType string
	View       string // table, graph, impact, timeline, directives
	Database   string
	Err        string

	// Counts
	EntityCount    int
	RelationCount  int
	DirectiveCount int
	DecisionCount  int
}

func (s *Server) handleOntology(w http.ResponseWriter, r *http.Request) {
	if s.ds == nil {
		s.render(w, r, "ontology", ontologyData{Err: "No database connection"})
		return
	}

	ctx := r.Context()

	// Find the database with ontology tables
	dbName, err := s.ds.FindOntologyDatabase(ctx)
	if err != nil {
		log.Printf("ontology: find db: %v", err)
		s.render(w, r, "ontology", ontologyData{Err: "No ontology database found"})
		return
	}

	view := r.URL.Query().Get("view")
	if view == "" {
		view = "table"
	}
	filterType := r.URL.Query().Get("type")
	impactSeed := r.URL.Query().Get("impact")

	data := ontologyData{
		View:       view,
		FilterType: filterType,
		ImpactSeed: impactSeed,
		Database:   dbName,
	}

	// Always load type summaries for the sidebar
	types, err := s.ds.OntologyTypeSummaries(ctx, dbName)
	if err != nil {
		log.Printf("ontology: types: %v", err)
	}
	data.Types = types

	switch view {
	case "table":
		entities, err := s.ds.OntologyEntities(ctx, dbName, filterType)
		if err != nil {
			log.Printf("ontology: entities: %v", err)
			data.Err = err.Error()
		}
		data.Entities = entities
		data.EntityCount = len(entities)

	case "graph":
		entities, err := s.ds.OntologyEntities(ctx, dbName, "")
		if err != nil {
			log.Printf("ontology: graph entities: %v", err)
			data.Err = err.Error()
			break
		}
		relations, err := s.ds.OntologyRelations(ctx, dbName)
		if err != nil {
			log.Printf("ontology: graph relations: %v", err)
			data.Err = err.Error()
			break
		}
		data.Entities = entities
		data.Relations = relations
		data.EntityCount = len(entities)
		data.RelationCount = len(relations)
		data.GraphNodes = buildGraphNodes(entities)
		data.GraphEdges = buildGraphEdges(relations)

	case "impact":
		if impactSeed != "" {
			impact, err := s.ds.OntologyImpact(ctx, dbName, impactSeed)
			if err != nil {
				log.Printf("ontology: impact: %v", err)
				data.Err = err.Error()
			}
			data.Impact = impact
		}
		// Also load all entities for the dropdown
		entities, err := s.ds.OntologyEntities(ctx, dbName, "")
		if err != nil {
			log.Printf("ontology: impact entities: %v", err)
		}
		data.Entities = entities
		data.EntityCount = len(entities)

	case "timeline":
		directives, err := s.ds.OntologyDirectives(ctx, dbName)
		if err != nil {
			log.Printf("ontology: directives: %v", err)
			data.Err = err.Error()
		}
		data.Directives = directives
		data.DirectiveCount = len(directives)

		decisions, err := s.ds.OntologyDecisions(ctx, dbName)
		if err != nil {
			log.Printf("ontology: decisions: %v", err)
		}
		data.Decisions = decisions
		data.DecisionCount = len(decisions)

	case "directives":
		directives, err := s.ds.OntologyDirectives(ctx, dbName)
		if err != nil {
			log.Printf("ontology: directives: %v", err)
			data.Err = err.Error()
		}
		data.Directives = directives
		data.DirectiveCount = len(directives)
	}

	s.render(w, r, "ontology", data)
}

// buildGraphNodes creates a JSON array of {id, label, type, domain} for the
// graph visualization.
func buildGraphNodes(entities []dolt.OntologyEntity) string {
	if len(entities) == 0 {
		return "[]"
	}
	var b strings.Builder
	b.WriteString("[")
	for i, e := range entities {
		if i > 0 {
			b.WriteString(",")
		}
		b.WriteString(`{"id":"`)
		b.WriteString(jsonEscape(e.ID))
		b.WriteString(`","label":"`)
		b.WriteString(jsonEscape(e.Name))
		b.WriteString(`","type":"`)
		b.WriteString(jsonEscape(e.Type))
		b.WriteString(`","domain":"`)
		b.WriteString(jsonEscape(e.Domain))
		b.WriteString(`"}`)
	}
	b.WriteString("]")
	return b.String()
}

// buildGraphEdges creates a JSON array of {from, to, label} for the
// graph visualization.
func buildGraphEdges(relations []dolt.OntologyRelation) string {
	if len(relations) == 0 {
		return "[]"
	}

	// Deduplicate and limit to key predicates for readability
	keyPredicates := map[string]bool{
		"dependsOn":  true,
		"runningOn":  true,
		"runsService": true,
		"managedBy":  true,
		"memberOf":   true,
		"reportsTo":  true,
	}

	var filtered []dolt.OntologyRelation
	for _, r := range relations {
		if keyPredicates[r.Predicate] {
			filtered = append(filtered, r)
		}
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Predicate < filtered[j].Predicate
	})

	var b strings.Builder
	b.WriteString("[")
	for i, r := range filtered {
		if i > 0 {
			b.WriteString(",")
		}
		b.WriteString(`{"from":"`)
		b.WriteString(jsonEscape(r.SubjectID))
		b.WriteString(`","to":"`)
		b.WriteString(jsonEscape(r.ObjectID))
		b.WriteString(`","label":"`)
		b.WriteString(jsonEscape(r.Predicate))
		b.WriteString(`"}`)
	}
	b.WriteString("]")
	return b.String()
}

// jsonEscape escapes a string for safe embedding in JSON.
func jsonEscape(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	s = strings.ReplaceAll(s, "\n", `\n`)
	s = strings.ReplaceAll(s, "\r", `\r`)
	s = strings.ReplaceAll(s, "\t", `\t`)
	return s
}
