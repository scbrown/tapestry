package web

import (
	"log"
	"net/http"
	"sort"
	"sync"
	"time"
)

type agentStreak struct {
	Name           string
	CurrentStreak  int
	LongestStreak  int
	ActiveDays30d  int
	TotalDays30d   int
	LastActiveDate string
}

type streaksData struct {
	GeneratedAt time.Time
	Agents      []agentStreak
	FilterRig   string
	Rigs        []string
	SortBy      string
	Err         string
}

func (s *Server) handleStreaks(w http.ResponseWriter, r *http.Request) {
	data := streaksData{GeneratedAt: time.Now()}

	if s.ds == nil {
		s.render(w, r, "streaks", data)
		return
	}

	ctx := r.Context()
	data.FilterRig = r.URL.Query().Get("rig")

	dbs, err := s.databases(ctx)
	if err != nil {
		data.Err = err.Error()
		s.render(w, r, "streaks", data)
		return
	}

	var rigNames []string
	for _, db := range dbs {
		rigNames = append(rigNames, db.Name)
	}
	sort.Strings(rigNames)
	data.Rigs = rigNames

	now := time.Now()
	thirtyDaysAgo := now.AddDate(0, 0, -30)

	// Collect activity per agent per day
	// key: agent name, value: set of active dates (YYYY-MM-DD)
	agentDays := map[string]map[string]bool{}

	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, db := range dbs {
		if data.FilterRig != "" && db.Name != data.FilterRig {
			continue
		}
		wg.Add(1)
		go func(dbName string) {
			defer wg.Done()

			diffs, err := s.ds.IssueDiffSince(ctx, dbName, thirtyDaysAgo)
			if err != nil {
				log.Printf("streaks: diffs %s: %v", dbName, err)
				return
			}

			commentDiffs, err := s.ds.CommentDiffSince(ctx, dbName, thirtyDaysAgo)
			if err != nil {
				log.Printf("streaks: comments %s: %v", dbName, err)
			}

			mu.Lock()
			defer mu.Unlock()

			for _, d := range diffs {
				actor := d.ToAssignee
				if actor == "" {
					actor = d.ToOwner
				}
				if actor == "" {
					continue
				}
				day := d.ToCommitDate.Format("2006-01-02")
				if agentDays[actor] == nil {
					agentDays[actor] = map[string]bool{}
				}
				agentDays[actor][day] = true
			}

			for _, c := range commentDiffs {
				if c.DiffType != "added" || c.ToAuthor == "" {
					continue
				}
				day := c.ToCommitDate.Format("2006-01-02")
				if agentDays[c.ToAuthor] == nil {
					agentDays[c.ToAuthor] = map[string]bool{}
				}
				agentDays[c.ToAuthor][day] = true
			}
		}(db.Name)
	}
	wg.Wait()

	// Calculate streaks for each agent
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	agents := make([]agentStreak, 0, len(agentDays))

	for name, days := range agentDays {
		as := agentStreak{
			Name:          name,
			ActiveDays30d: len(days),
			TotalDays30d:  30,
		}

		// Sort days to find streaks
		sortedDays := make([]string, 0, len(days))
		for d := range days {
			sortedDays = append(sortedDays, d)
		}
		sort.Strings(sortedDays)

		if len(sortedDays) > 0 {
			as.LastActiveDate = sortedDays[len(sortedDays)-1]
		}

		// Calculate current streak (counting back from today)
		currentStreak := 0
		for d := today; ; d = d.AddDate(0, 0, -1) {
			if days[d.Format("2006-01-02")] {
				currentStreak++
			} else {
				break
			}
		}
		as.CurrentStreak = currentStreak

		// Calculate longest streak
		longest := 0
		streak := 0
		prevDate := time.Time{}
		for _, ds := range sortedDays {
			t, _ := time.Parse("2006-01-02", ds)
			if !prevDate.IsZero() && t.Sub(prevDate) == 24*time.Hour {
				streak++
			} else {
				streak = 1
			}
			if streak > longest {
				longest = streak
			}
			prevDate = t
		}
		as.LongestStreak = longest

		agents = append(agents, as)
	}

	sortBy := r.URL.Query().Get("sort")
	if sortBy == "" {
		sortBy = "streak"
	}
	data.SortBy = sortBy

	switch sortBy {
	case "longest":
		sort.Slice(agents, func(i, j int) bool {
			if agents[i].LongestStreak != agents[j].LongestStreak {
				return agents[i].LongestStreak > agents[j].LongestStreak
			}
			return agents[i].CurrentStreak > agents[j].CurrentStreak
		})
	case "active":
		sort.Slice(agents, func(i, j int) bool {
			if agents[i].ActiveDays30d != agents[j].ActiveDays30d {
				return agents[i].ActiveDays30d > agents[j].ActiveDays30d
			}
			return agents[i].CurrentStreak > agents[j].CurrentStreak
		})
	case "name":
		sort.Slice(agents, func(i, j int) bool {
			return agents[i].Name < agents[j].Name
		})
	default: // streak
		sort.Slice(agents, func(i, j int) bool {
			if agents[i].CurrentStreak != agents[j].CurrentStreak {
				return agents[i].CurrentStreak > agents[j].CurrentStreak
			}
			return agents[i].ActiveDays30d > agents[j].ActiveDays30d
		})
	}

	data.Agents = agents

	s.render(w, r, "streaks", data)
}
