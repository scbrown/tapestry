package web

import (
	"log"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/scbrown/tapestry/internal/dolt"
)

type dogPileItem struct {
	Rig            string
	Issue          dolt.Issue
	StatusChanges  int
	CommentCount   int
	HeatScore      int // statusChanges*3 + commentCount*2
}

type dogPileData struct {
	GeneratedAt time.Time
	Items       []dogPileItem
	Total       int
	Rigs        []string
	FilterRig   string
	Window      string
	Assignees   []string
}

func (s *Server) handleDogPile(w http.ResponseWriter, r *http.Request) {
	data := dogPileData{GeneratedAt: time.Now()}

	if s.ds == nil {
		s.render(w, r, "dog-pile", data)
		return
	}

	ctx := r.Context()
	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("dog-pile: list dbs: %v", err)
		s.render(w, r, "dog-pile", data)
		return
	}

	filterRig := r.URL.Query().Get("rig")
	window := r.URL.Query().Get("window")
	if window == "" {
		window = "7d"
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
	case "24h":
		since = time.Now().AddDate(0, 0, -1)
	case "30d":
		since = time.Now().AddDate(0, 0, -30)
	default:
		since = time.Now().AddDate(0, 0, -7)
	}

	type dbResult struct {
		rig   string
		items []dogPileItem
	}
	results := make([]dbResult, len(dbs))
	var wg sync.WaitGroup

	for i, db := range dbs {
		if filterRig != "" && db.Name != filterRig {
			continue
		}
		wg.Add(1)
		go func(idx int, dbName string) {
			defer wg.Done()
			assignees, _ := s.ds.DistinctAssignees(ctx, dbName)
			if len(assignees) > 0 {
				data.Assignees = append(data.Assignees, assignees...)
			}
			var items []dogPileItem

			// Count status changes per issue
			statusCounts := map[string]int{}
			issueTitles := map[string]string{}
			issueData := map[string]dolt.Issue{}

			diffs, err := s.ds.IssueDiffSince(ctx, dbName, since)
			if err != nil {
				log.Printf("dog-pile: %s diffs: %v", dbName, err)
			} else {
				for _, d := range diffs {
					if d.FromStatus != d.ToStatus && d.FromStatus != "" {
						statusCounts[d.ToID]++
					}
					issueTitles[d.ToID] = d.ToTitle
				}
			}

			// Count comments per issue
			commentCounts := map[string]int{}
			cdiffs, err := s.ds.CommentDiffSince(ctx, dbName, since)
			if err != nil {
				log.Printf("dog-pile: %s comments: %v", dbName, err)
			} else {
				for _, c := range cdiffs {
					if c.DiffType == "added" {
						commentCounts[c.ToIssueID]++
					}
				}
			}

			// Merge all issue IDs
			allIDs := map[string]bool{}
			for id := range statusCounts {
				allIDs[id] = true
			}
			for id := range commentCounts {
				allIDs[id] = true
			}

			// Fetch issue details for IDs with activity
			for id := range allIDs {
				sc := statusCounts[id]
				cc := commentCounts[id]
				heat := sc*3 + cc*2
				if heat < 4 { // minimum threshold
					continue
				}
				iss, ok := issueData[id]
				if !ok {
					fetched, err := s.ds.IssueByID(ctx, dbName, id)
					if err != nil || fetched == nil {
						// Use partial data from diffs
						iss = dolt.Issue{ID: id, Title: issueTitles[id]}
					} else {
						iss = *fetched
					}
				}
				items = append(items, dogPileItem{
					Rig:           dbName,
					Issue:         iss,
					StatusChanges: sc,
					CommentCount:  cc,
					HeatScore:     heat,
				})
			}

			results[idx] = dbResult{rig: dbName, items: items}
		}(i, db.Name)
	}
	wg.Wait()

	sort.Strings(data.Assignees)

	var allItems []dogPileItem
	for _, r := range results {
		allItems = append(allItems, r.items...)
	}

	sort.Slice(allItems, func(i, j int) bool {
		return allItems[i].HeatScore > allItems[j].HeatScore
	})

	if len(allItems) > 50 {
		allItems = allItems[:50]
	}

	data.Items = allItems
	data.Total = len(allItems)
	s.render(w, r, "dog-pile", data)
}
