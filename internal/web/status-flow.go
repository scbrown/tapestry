package web

import (
	"log"
	"net/http"
	"sort"
	"sync"
	"time"
)

type statusTransition struct {
	From  string
	To    string
	Count int
	Pct   float64
}

type statusFlowData struct {
	GeneratedAt time.Time
	Transitions []statusTransition
	Total       int
	Rigs        []string
	FilterRig   string
	Window      string
	SortBy      string
	Err         string
}

func (s *Server) handleStatusFlow(w http.ResponseWriter, r *http.Request) {
	data := statusFlowData{GeneratedAt: time.Now(), Window: "30d"}

	if s.ds == nil {
		s.render(w, r, "status-flow", data)
		return
	}

	ctx := r.Context()
	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("status-flow: list dbs: %v", err)
		s.render(w, r, "status-flow", data)
		return
	}

	filterRig := r.URL.Query().Get("rig")
	window := r.URL.Query().Get("window")
	if window == "" {
		window = "30d"
	}
	data.FilterRig = filterRig
	data.Window = window

	var rigs []string
	for _, db := range dbs {
		rigs = append(rigs, db.Name)
	}
	sort.Strings(rigs)
	data.Rigs = rigs

	var since time.Time
	switch window {
	case "7d":
		since = time.Now().AddDate(0, 0, -7)
	case "90d":
		since = time.Now().AddDate(0, 0, -90)
	default:
		since = time.Now().AddDate(0, 0, -30)
	}

	// Count transitions
	type transKey struct{ from, to string }
	transMap := make(map[transKey]int)
	var mu sync.Mutex
	var wg sync.WaitGroup
	totalCount := 0

	for _, db := range dbs {
		if filterRig != "" && db.Name != filterRig {
			continue
		}
		wg.Add(1)
		go func(dbName string) {
			defer wg.Done()
			diffs, err := s.ds.IssueDiffSince(ctx, dbName, since)
			if err != nil {
				log.Printf("status-flow: %s diff: %v", dbName, err)
				return
			}

			mu.Lock()
			defer mu.Unlock()
			for _, d := range diffs {
				if d.FromStatus == "" || d.FromStatus == d.ToStatus {
					continue
				}
				key := transKey{d.FromStatus, d.ToStatus}
				transMap[key]++
				totalCount++
			}
		}(db.Name)
	}
	wg.Wait()

	var transitions []statusTransition
	for key, count := range transMap {
		pct := 0.0
		if totalCount > 0 {
			pct = float64(count) / float64(totalCount) * 100
		}
		transitions = append(transitions, statusTransition{
			From:  key.from,
			To:    key.to,
			Count: count,
			Pct:   pct,
		})
	}

	sortBy := r.URL.Query().Get("sort")
	if sortBy == "" {
		sortBy = "count"
	}
	data.SortBy = sortBy

	switch sortBy {
	case "from":
		sort.Slice(transitions, func(i, j int) bool {
			if transitions[i].From != transitions[j].From {
				return transitions[i].From < transitions[j].From
			}
			return transitions[i].Count > transitions[j].Count
		})
	case "to":
		sort.Slice(transitions, func(i, j int) bool {
			if transitions[i].To != transitions[j].To {
				return transitions[i].To < transitions[j].To
			}
			return transitions[i].Count > transitions[j].Count
		})
	case "pct":
		sort.Slice(transitions, func(i, j int) bool {
			return transitions[i].Pct > transitions[j].Pct
		})
	default: // count
		sort.Slice(transitions, func(i, j int) bool {
			return transitions[i].Count > transitions[j].Count
		})
	}

	data.Transitions = transitions
	data.Total = totalCount
	s.render(w, r, "status-flow", data)
}
