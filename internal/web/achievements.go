package web

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/scbrown/tapestry/internal/dolt"
)

type Achievement struct {
	ID          string
	Name        string
	Description string
	Icon        string
	Category    string
	Unlocked    bool
	UnlockedAt  time.Time
	UnlockedBy  string
	Note        string
}

type CategoryProgress struct {
	Name     string
	Total    int
	Unlocked int
}

type achievementPageData struct {
	Achievements []Achievement
	Categories   []CategoryProgress
	Selected     string
	Total        int
	Unlocked     int
	Percent      int
	Err          string
}

func (s *Server) handleAchievements(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	category := r.URL.Query().Get("category")

	if s.ds == nil {
		s.render(w, r, "achievements", achievementPageData{Err: "No database connection configured"})
		return
	}

	// Achievement data lives in the aegis database
	defs, unlocks, err := s.loadAchievements(ctx)
	if err != nil {
		log.Printf("achievements: %v", err)
		s.render(w, r, "achievements", achievementPageData{Err: "Could not load achievements"})
		return
	}

	achievements := mergeAchievements(defs, unlocks)

	// Build category progress
	catMap := map[string]*CategoryProgress{}
	for _, a := range achievements {
		cp, ok := catMap[a.Category]
		if !ok {
			cp = &CategoryProgress{Name: a.Category}
			catMap[a.Category] = cp
		}
		cp.Total++
		if a.Unlocked {
			cp.Unlocked++
		}
	}

	categoryOrder := []string{"infrastructure", "development", "operations", "milestones", "special"}
	var categories []CategoryProgress
	for _, name := range categoryOrder {
		if cp, ok := catMap[name]; ok {
			categories = append(categories, *cp)
		}
	}

	// Filter by category if requested
	if category != "" {
		var filtered []Achievement
		for _, a := range achievements {
			if a.Category == category {
				filtered = append(filtered, a)
			}
		}
		achievements = filtered
	}

	total := 0
	unlocked := 0
	for _, cp := range catMap {
		total += cp.Total
		unlocked += cp.Unlocked
	}

	pct := 0
	if total > 0 {
		pct = unlocked * 100 / total
	}

	s.render(w, r, "achievements", achievementPageData{
		Achievements: achievements,
		Categories:   categories,
		Selected:     category,
		Total:        total,
		Unlocked:     unlocked,
		Percent:      pct,
	})
}

func (s *Server) loadAchievements(ctx context.Context) ([]dolt.AchievementDef, []dolt.AchievementUnlock, error) {
	dbs, err := s.databases(ctx)
	if err != nil {
		return nil, nil, err
	}

	// Find the aegis database (achievement tables live there)
	db := ""
	for _, d := range dbs {
		if d.Name == "aegis" {
			db = "aegis"
			break
		}
	}
	if db == "" {
		return nil, nil, nil
	}

	defs, err := s.ds.AchievementDefs(ctx, db)
	if err != nil {
		return nil, nil, err
	}

	unlocks, err := s.ds.AchievementUnlocks(ctx, db)
	if err != nil {
		return nil, nil, err
	}

	return defs, unlocks, nil
}

func mergeAchievements(defs []dolt.AchievementDef, unlocks []dolt.AchievementUnlock) []Achievement {
	unlockMap := map[string]dolt.AchievementUnlock{}
	for _, u := range unlocks {
		unlockMap[u.ID] = u
	}

	var achievements []Achievement
	for _, d := range defs {
		a := Achievement{
			ID:          d.ID,
			Name:        d.Name,
			Description: d.Description,
			Icon:        d.Icon,
			Category:    d.Category,
		}
		if u, ok := unlockMap[d.ID]; ok {
			a.Unlocked = true
			a.UnlockedAt = u.UnlockedAt
			a.UnlockedBy = u.UnlockedBy
			a.Note = u.Note
		}
		achievements = append(achievements, a)
	}
	return achievements
}
