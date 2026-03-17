package web

import (
	"log"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/scbrown/tapestry/internal/dolt"
)

type impactAgent struct {
	Name        string
	Closed7d    int
	Closed30d   int
	Created7d   int
	Created30d  int
	Comments7d  int
	Comments30d int
	Score7d     float64 // weighted impact score
	Score30d    float64
}

type impactData struct {
	GeneratedAt time.Time
	Agents      []impactAgent
	FilterRig   string
	Err         string
}

func (s *Server) handleImpact(w http.ResponseWriter, r *http.Request) {
	data := impactData{GeneratedAt: time.Now()}

	if s.ds == nil {
		s.render(w, r, "impact", data)
		return
	}

	ctx := r.Context()
	data.FilterRig = r.URL.Query().Get("rig")

	dbs, err := s.databases(ctx)
	if err != nil {
		data.Err = err.Error()
		s.render(w, r, "impact", data)
		return
	}

	now := time.Now()
	sevenDaysAgo := now.AddDate(0, 0, -7)
	thirtyDaysAgo := now.AddDate(0, 0, -30)

	agentMap := map[string]*impactAgent{}

	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, db := range dbs {
		if data.FilterRig != "" && db.Name != data.FilterRig {
			continue
		}
		wg.Add(1)
		go func(dbName string) {
			defer wg.Done()

			// Get issue diffs for 30 days
			diffs, err := s.ds.IssueDiffSince(ctx, dbName, thirtyDaysAgo)
			if err != nil {
				log.Printf("impact: diffs %s: %v", dbName, err)
				return
			}

			// Get comment diffs for 30 days
			commentDiffs, err := s.ds.CommentDiffSince(ctx, dbName, thirtyDaysAgo)
			if err != nil {
				log.Printf("impact: comments %s: %v", dbName, err)
			}

			// Build issue priority map
			priMap := map[string]int{}
			issues, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{Limit: 5000})
			if err == nil {
				for _, iss := range issues {
					priMap[iss.ID] = iss.Priority
				}
			}

			mu.Lock()
			defer mu.Unlock()

			for _, d := range diffs {
				if d.DiffType == "added" {
					actor := d.ToOwner
					if actor == "" {
						actor = d.ToAssignee
					}
					if actor == "" {
						continue
					}
					a := impactGetOrCreate(agentMap, actor)
					a.Created30d++
					if d.ToCommitDate.After(sevenDaysAgo) {
						a.Created7d++
					}
				} else if d.ToStatus == "closed" && d.FromStatus != "closed" {
					actor := d.ToAssignee
					if actor == "" {
						actor = d.ToOwner
					}
					if actor == "" {
						continue
					}
					a := impactGetOrCreate(agentMap, actor)
					a.Closed30d++
					if d.ToCommitDate.After(sevenDaysAgo) {
						a.Closed7d++
					}
				}
			}

			for _, c := range commentDiffs {
				if c.DiffType != "added" || c.ToAuthor == "" {
					continue
				}
				a := impactGetOrCreate(agentMap, c.ToAuthor)
				a.Comments30d++
				if c.ToCommitDate.After(sevenDaysAgo) {
					a.Comments7d++
				}
			}
		}(db.Name)
	}
	wg.Wait()

	// Calculate impact scores
	// Score = closures*3 + creations*1 + comments*0.5
	agents := make([]impactAgent, 0, len(agentMap))
	for _, a := range agentMap {
		a.Score7d = float64(a.Closed7d)*3 + float64(a.Created7d) + float64(a.Comments7d)*0.5
		a.Score30d = float64(a.Closed30d)*3 + float64(a.Created30d) + float64(a.Comments30d)*0.5
		agents = append(agents, *a)
	}

	sort.Slice(agents, func(i, j int) bool {
		return agents[i].Score30d > agents[j].Score30d
	})

	data.Agents = agents

	s.render(w, r, "impact", data)
}

func impactGetOrCreate(m map[string]*impactAgent, name string) *impactAgent {
	a, ok := m[name]
	if !ok {
		a = &impactAgent{Name: name}
		m[name] = a
	}
	return a
}
