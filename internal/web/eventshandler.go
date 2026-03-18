package web

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/scbrown/tapestry/internal/events"
)

type eventsPageData struct {
	Events      []events.Event
	Types       []string
	Rigs        []string
	TypeFilter  string
	ActorFilter string
	RigFilter   string
	Total       int
	Err         string
}

type handoffsPageData struct {
	Stats       []events.ChainStats
	Chains      []events.HandoffChain
	TotalChains int
	ActorFilter string
	Err         string
}

func (s *Server) loadEvents() ([]events.Event, error) {
	if s.workspacePath == "" {
		return nil, nil
	}
	return events.ReadWorkspace(s.workspacePath)
}

func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	typeFilter := r.URL.Query().Get("type")
	actorFilter := r.URL.Query().Get("actor")
	rigFilter := r.URL.Query().Get("rig")

	data := eventsPageData{
		TypeFilter:  typeFilter,
		ActorFilter: actorFilter,
		RigFilter:   rigFilter,
	}

	allEvents, err := s.loadEvents()
	if err != nil {
		log.Printf("events: load: %v", err)
		data.Err = "Could not load events file"
		s.render(w, r, "events", data)
		return
	}

	if allEvents == nil {
		data.Err = "No workspace path configured"
		s.render(w, r, "events", data)
		return
	}

	data.Types = events.Types(allEvents)
	data.Rigs = events.Rigs(allEvents)
	data.Total = len(allEvents)

	// Build actor filter: if rig filter is set, match actor prefix
	effectiveActor := actorFilter
	if rigFilter != "" && actorFilter == "" {
		effectiveActor = rigFilter + "/"
	}

	// Apply filters
	filtered := events.Apply(allEvents, events.Filter{
		Type:  typeFilter,
		Actor: effectiveActor,
		After: time.Now().Add(-7 * 24 * time.Hour), // last 7 days
		Limit: 200,
	})
	data.Events = filtered

	s.render(w, r, "events", data)
}

func (s *Server) handleHandoffs(w http.ResponseWriter, r *http.Request) {
	actorFilter := r.URL.Query().Get("actor")

	data := handoffsPageData{
		ActorFilter: actorFilter,
	}

	allEvents, err := s.loadEvents()
	if err != nil {
		log.Printf("handoffs: load: %v", err)
		data.Err = "Could not load events file"
		s.render(w, r, "handoffs", data)
		return
	}

	if allEvents == nil {
		data.Err = "No workspace path configured"
		s.render(w, r, "handoffs", data)
		return
	}

	chains := events.BuildHandoffChains(allEvents)

	// Filter by actor if requested
	if actorFilter != "" {
		var filtered []events.HandoffChain
		for _, c := range chains {
			if strings.Contains(c.Actor, actorFilter) {
				filtered = append(filtered, c)
			}
		}
		chains = filtered
	}

	data.TotalChains = len(chains)
	data.Stats = events.ChainSummary(chains)
	data.Chains = chains

	s.render(w, r, "handoffs", data)
}
