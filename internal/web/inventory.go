package web

import (
	"log"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/scbrown/tapestry/internal/dolt"
)

type inventoryData struct {
	GeneratedAt time.Time
	TotalBeads  int
	ByStatus    []statusCount
	ByType      []typeCount
	ByRig       []rigCount
	RigCount    int
	Rigs        []string
	FilterRig   string
}

type statusCount struct {
	Status string
	Count  int
}

type typeCount struct {
	Type  string
	Count int
}

type rigCount struct {
	Rig   string
	Total int
	Open  int
	Closed int
}

func (s *Server) handleInventory(w http.ResponseWriter, r *http.Request) {
	data := inventoryData{GeneratedAt: time.Now()}

	if s.ds == nil {
		s.render(w, r, "inventory", data)
		return
	}

	ctx := r.Context()
	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("inventory: list dbs: %v", err)
		s.render(w, r, "inventory", data)
		return
	}

	type dbResult struct {
		rig    string
		issues []dolt.Issue
	}
	results := make([]dbResult, len(dbs))
	var wg sync.WaitGroup
	for i, db := range dbs {
		wg.Add(1)
		go func(i int, dbName string) {
			defer wg.Done()
			issues, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{Limit: 5000})
			if err != nil {
				log.Printf("inventory: %s: %v", dbName, err)
				return
			}
			results[i] = dbResult{rig: dbName, issues: issues}
		}(i, db.Name)
	}
	wg.Wait()

	filterRig := r.URL.Query().Get("rig")
	var rigs []string
	for _, r := range results {
		if len(r.issues) > 0 {
			rigs = append(rigs, r.rig)
		}
	}
	sort.Strings(rigs)
	data.Rigs = rigs
	data.FilterRig = filterRig

	statusMap := map[string]int{}
	typeMap := map[string]int{}
	rigMap := map[string]*rigCount{}

	for _, r := range results {
		if filterRig != "" && r.rig != filterRig {
			continue
		}
		rc, ok := rigMap[r.rig]
		if !ok {
			rc = &rigCount{Rig: r.rig}
			rigMap[r.rig] = rc
		}
		for _, iss := range r.issues {
			data.TotalBeads++
			statusMap[iss.Status]++
			if iss.Type != "" {
				typeMap[iss.Type]++
			} else {
				typeMap["(none)"]++
			}
			rc.Total++
			if iss.Status == "closed" {
				rc.Closed++
			} else {
				rc.Open++
			}
		}
	}

	for s, c := range statusMap {
		data.ByStatus = append(data.ByStatus, statusCount{Status: s, Count: c})
	}
	sort.Slice(data.ByStatus, func(i, j int) bool {
		return data.ByStatus[i].Count > data.ByStatus[j].Count
	})

	for t, c := range typeMap {
		data.ByType = append(data.ByType, typeCount{Type: t, Count: c})
	}
	sort.Slice(data.ByType, func(i, j int) bool {
		return data.ByType[i].Count > data.ByType[j].Count
	})

	for _, rc := range rigMap {
		data.ByRig = append(data.ByRig, *rc)
	}
	sort.Slice(data.ByRig, func(i, j int) bool {
		return data.ByRig[i].Total > data.ByRig[j].Total
	})

	data.RigCount = len(data.ByRig)
	s.render(w, r, "inventory", data)
}
