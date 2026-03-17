package web

import (
	"context"
	"log"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/scbrown/tapestry/internal/dolt"
)

type debtItem struct {
	ID        string
	DB        string
	Title     string
	Priority  int
	AgeDays   int
	Assignee  string
	Status    string
	CreatedAt time.Time
}

type debtData struct {
	GeneratedAt time.Time

	// Ratios
	TotalOpen   int
	BugCount    int
	TaskCount   int
	EpicCount   int
	DeferCount  int
	BugPct      int
	DeferPct    int

	// Old bugs (open bugs > 14 days)
	OldBugs []debtItem

	// Deferred pile (longest deferred items)
	DeferredPile []debtItem

	// Untyped items (no type set — metadata gap)
	UntypedCount int

	// Unprioritized (priority < 0 or missing)
	UnprioritizedCount int

	Rigs      []string
	FilterRig string
	Err       string
}

func (s *Server) handleDebt(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	data := debtData{GeneratedAt: now}

	if s.ds == nil {
		s.render(w, r, "debt", data)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("debt: list dbs: %v", err)
		s.render(w, r, "debt", debtData{Err: err.Error(), GeneratedAt: now})
		return
	}

	filterRig := r.URL.Query().Get("rig")
	data.FilterRig = filterRig

	type dbAccum struct {
		bugs     []debtItem
		deferred []debtItem
		bugCnt   int
		taskCnt  int
		epicCnt  int
		deferCnt int
		openCnt  int
		untyped  int
		unpri    int
	}

	results := make([]dbAccum, len(dbs))
	var wg sync.WaitGroup

	for i, db := range dbs {
		if filterRig != "" && db.Name != filterRig {
			continue
		}
		wg.Add(1)
		go func(idx int, dbName string) {
			defer wg.Done()
			var acc dbAccum

			issues, err := s.ds.Issues(ctx, dbName, dolt.IssueFilter{Limit: 2000})
			if err != nil {
				log.Printf("debt %s: %v", dbName, err)
				return
			}

			for _, iss := range issues {
				if isNoise(iss.ID, iss.Title) {
					continue
				}
				if iss.Status == "closed" {
					continue
				}

				acc.openCnt++
				age := int(now.Sub(iss.CreatedAt).Hours() / 24)

				switch iss.Type {
				case "bug":
					acc.bugCnt++
					if iss.Status != "deferred" && age > 14 {
						acc.bugs = append(acc.bugs, debtItem{
							ID: iss.ID, DB: dbName, Title: iss.Title,
							Priority: iss.Priority, AgeDays: age,
							Assignee: iss.Assignee, Status: iss.Status,
							CreatedAt: iss.CreatedAt,
						})
					}
				case "task":
					acc.taskCnt++
				case "epic":
					acc.epicCnt++
				case "":
					acc.untyped++
				}

				if iss.Priority < 0 {
					acc.unpri++
				}

				if iss.Status == "deferred" {
					acc.deferCnt++
					acc.deferred = append(acc.deferred, debtItem{
						ID: iss.ID, DB: dbName, Title: iss.Title,
						Priority: iss.Priority, AgeDays: age,
						Assignee: iss.Assignee, Status: iss.Status,
						CreatedAt: iss.CreatedAt,
					})
				}
			}

			results[idx] = acc
		}(i, db.Name)
	}
	wg.Wait()

	rigSet := make(map[string]bool)
	var allBugs, allDeferred []debtItem

	for idx, acc := range results {
		data.TotalOpen += acc.openCnt
		data.BugCount += acc.bugCnt
		data.TaskCount += acc.taskCnt
		data.EpicCount += acc.epicCnt
		data.DeferCount += acc.deferCnt
		data.UntypedCount += acc.untyped
		data.UnprioritizedCount += acc.unpri
		allBugs = append(allBugs, acc.bugs...)
		allDeferred = append(allDeferred, acc.deferred...)
		if acc.openCnt > 0 {
			rigSet[dbs[idx].Name] = true
		}
	}

	if data.TotalOpen > 0 {
		data.BugPct = pct(data.BugCount, data.TotalOpen)
		data.DeferPct = pct(data.DeferCount, data.TotalOpen)
	}

	// Sort old bugs by age descending, take top 20
	sort.Slice(allBugs, func(i, j int) bool { return allBugs[i].AgeDays > allBugs[j].AgeDays })
	if len(allBugs) > 20 {
		allBugs = allBugs[:20]
	}
	data.OldBugs = allBugs

	// Sort deferred by age descending, take top 20
	sort.Slice(allDeferred, func(i, j int) bool { return allDeferred[i].AgeDays > allDeferred[j].AgeDays })
	if len(allDeferred) > 20 {
		allDeferred = allDeferred[:20]
	}
	data.DeferredPile = allDeferred

	for rig := range rigSet {
		data.Rigs = append(data.Rigs, rig)
	}
	sort.Strings(data.Rigs)

	s.render(w, r, "debt", data)
}
