package web

import (
	"log"
	"net/http"
	"sort"
	"sync"
	"time"
)

type pulseHour struct {
	Hour     int
	Created  int
	Closed   int
	Comments int
	Changes  int
	Total    int
}

type pulseAgent struct {
	Name     string
	Created  int
	Closed   int
	Comments int
	Changes  int
	Total    int
}

type pulseData struct {
	GeneratedAt  time.Time
	Hours        []pulseHour
	Agents       []pulseAgent
	TotalCreated int
	TotalClosed  int
	TotalComments int
	TotalChanges int
	PeakHour     int
	PeakCount    int
	FilterRig    string
	Err          string
}

func (s *Server) handlePulse(w http.ResponseWriter, r *http.Request) {
	data := pulseData{GeneratedAt: time.Now()}

	if s.ds == nil {
		s.render(w, r, "pulse", data)
		return
	}

	ctx := r.Context()
	filterRig := r.URL.Query().Get("rig")
	data.FilterRig = filterRig

	dbs, err := s.databases(ctx)
	if err != nil {
		data.Err = err.Error()
		s.render(w, r, "pulse", data)
		return
	}

	now := time.Now()
	since := now.Add(-24 * time.Hour)

	// Initialize 24 hours
	hours := make([]pulseHour, 24)
	for i := range hours {
		hours[i].Hour = i
	}

	agentMap := map[string]*pulseAgent{}

	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, db := range dbs {
		if filterRig != "" && db.Name != filterRig {
			continue
		}
		wg.Add(1)
		go func(dbName string) {
			defer wg.Done()

			// Get issue diffs (creates, status changes)
			diffs, err := s.ds.IssueDiffSince(ctx, dbName, since)
			if err != nil {
				log.Printf("pulse: diff %s: %v", dbName, err)
				return
			}

			// Get comment diffs
			commentDiffs, err := s.ds.CommentDiffSince(ctx, dbName, since)
			if err != nil {
				log.Printf("pulse: comments %s: %v", dbName, err)
			}

			mu.Lock()
			defer mu.Unlock()

			for _, d := range diffs {
				h := d.ToCommitDate.Hour()
				if d.DiffType == "added" {
					hours[h].Created++
					data.TotalCreated++
					actor := d.ToOwner
					if actor == "" {
						actor = d.ToAssignee
					}
					if actor != "" {
						a := getOrCreatePulseAgent(agentMap, actor)
						a.Created++
						a.Total++
					}
				} else {
					// Status change
					if d.FromStatus != d.ToStatus {
						hours[h].Changes++
						data.TotalChanges++
						if d.ToStatus == "closed" {
							hours[h].Closed++
							data.TotalClosed++
							actor := d.ToAssignee
							if actor == "" {
								actor = d.ToOwner
							}
							if actor != "" {
								a := getOrCreatePulseAgent(agentMap, actor)
								a.Closed++
								a.Total++
							}
						} else {
							actor := d.ToAssignee
							if actor == "" {
								actor = d.ToOwner
							}
							if actor != "" {
								a := getOrCreatePulseAgent(agentMap, actor)
								a.Changes++
								a.Total++
							}
						}
					}
				}
			}

			for _, c := range commentDiffs {
				if c.DiffType != "added" {
					continue
				}
				h := c.ToCommitDate.Hour()
				hours[h].Comments++
				data.TotalComments++
				if c.ToAuthor != "" {
					a := getOrCreatePulseAgent(agentMap, c.ToAuthor)
					a.Comments++
					a.Total++
				}
			}
		}(db.Name)
	}
	wg.Wait()

	// Calculate totals per hour and find peak
	for i := range hours {
		hours[i].Total = hours[i].Created + hours[i].Closed + hours[i].Comments + hours[i].Changes
		if hours[i].Total > data.PeakCount {
			data.PeakCount = hours[i].Total
			data.PeakHour = hours[i].Hour
		}
	}

	data.Hours = hours

	// Convert agent map to sorted slice
	agents := make([]pulseAgent, 0, len(agentMap))
	for _, a := range agentMap {
		agents = append(agents, *a)
	}
	sort.Slice(agents, func(i, j int) bool {
		return agents[i].Total > agents[j].Total
	})
	data.Agents = agents

	s.render(w, r, "pulse", data)
}

func getOrCreatePulseAgent(m map[string]*pulseAgent, name string) *pulseAgent {
	a, ok := m[name]
	if !ok {
		a = &pulseAgent{Name: name}
		m[name] = a
	}
	return a
}
