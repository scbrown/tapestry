package web

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/scbrown/tapestry/internal/dolt"
)

type themeParksPageData struct {
	Parks     []dolt.ThemePark
	Rides     []dolt.Ride
	Visits    []dolt.ParkVisit
	Plans     []dolt.TripPlan
	Filter    string // "all", "visited", "wishlist"
	ParkID    string // filter rides by park
	Stats     themeParksStats
	Err       string
}

type themeParksStats struct {
	TotalParks   int
	Visited      int
	Wishlisted   int
	TotalRides   int
	Ridden       int
	TotalVisits  int
	PlannedTrips int
}

func (s *Server) handleThemeParks(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	filter := r.URL.Query().Get("filter")
	if filter == "" {
		filter = "all"
	}

	if s.ds == nil {
		s.render(w, r, "theme-parks", themeParksPageData{Err: "No database connection configured"})
		return
	}

	parks, rides, visits, plans, err := s.loadThemeParksData(ctx)
	if err != nil {
		log.Printf("theme-parks: %v", err)
		s.render(w, r, "theme-parks", themeParksPageData{Err: "Could not load theme park data"})
		return
	}

	// Compute stats
	var stats themeParksStats
	stats.TotalParks = len(parks)
	stats.TotalRides = len(rides)
	stats.TotalVisits = len(visits)
	stats.PlannedTrips = len(plans)
	for _, p := range parks {
		if p.Visited {
			stats.Visited++
		}
		if p.Wishlist {
			stats.Wishlisted++
		}
	}
	for _, r := range rides {
		if r.Ridden {
			stats.Ridden++
		}
	}

	// Apply filter
	if filter == "visited" {
		var filtered []dolt.ThemePark
		for _, p := range parks {
			if p.Visited {
				filtered = append(filtered, p)
			}
		}
		parks = filtered
	} else if filter == "wishlist" {
		var filtered []dolt.ThemePark
		for _, p := range parks {
			if p.Wishlist {
				filtered = append(filtered, p)
			}
		}
		parks = filtered
	}

	s.render(w, r, "theme-parks", themeParksPageData{
		Parks:  parks,
		Rides:  rides,
		Visits: visits,
		Plans:  plans,
		Filter: filter,
		Stats:  stats,
	})
}

func (s *Server) loadThemeParksData(ctx context.Context) (
	[]dolt.ThemePark, []dolt.Ride, []dolt.ParkVisit, []dolt.TripPlan, error,
) {
	dbs, err := s.databases(ctx)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	db := ""
	for _, d := range dbs {
		if d.Name == "aegis" {
			db = "aegis"
			break
		}
	}
	if db == "" {
		return nil, nil, nil, nil, nil
	}

	parks, err := s.ds.ThemeParks(ctx, db)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	rides, err := s.ds.Rides(ctx, db, "")
	if err != nil {
		return nil, nil, nil, nil, err
	}

	visits, err := s.ds.ParkVisits(ctx, db, "")
	if err != nil {
		return nil, nil, nil, nil, err
	}

	plans, err := s.ds.TripPlans(ctx, db)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	return parks, rides, visits, plans, nil
}
