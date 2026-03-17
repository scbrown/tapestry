package web

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sort"
	"sync"
	"time"
)

// labelPair tracks how often two labels co-occur on the same bead
type labelPair struct {
	LabelA string
	LabelB string
	Count  int
}

type pairFreqData struct {
	GeneratedAt time.Time

	Pairs      []labelPair
	TotalPairs int
	TopLabel   string // most frequently paired label

	Err string
}

func (s *Server) handlePairFreq(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	data := pairFreqData{GeneratedAt: now}

	if s.ds == nil {
		s.render(w, r, "pair-freq", data)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	dbs, err := s.databases(ctx)
	if err != nil {
		log.Printf("pair-freq: list dbs: %v", err)
		data.Err = err.Error()
		s.render(w, r, "pair-freq", data)
		return
	}

	// Collect all issue -> labels mappings
	pairCounts := make(map[string]int) // "labelA|labelB" -> count
	labelFreq := make(map[string]int)  // label -> number of pairings

	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, db := range dbs {
		wg.Add(1)
		go func(dbName string) {
			defer wg.Done()

			labels, err := s.ds.DistinctLabels(ctx, dbName)
			if err != nil {
				log.Printf("pair-freq %s: labels: %v", dbName, err)
				return
			}

			// For each label, get its issues
			labelIssues := make(map[string]map[string]bool) // issueID -> set of labels
			for _, lc := range labels {
				issues, err := s.ds.IssuesByLabel(ctx, dbName, lc.Label)
				if err != nil {
					continue
				}
				for _, iss := range issues {
					if isNoise(iss.ID, iss.Title) {
						continue
					}
					if labelIssues[iss.ID] == nil {
						labelIssues[iss.ID] = make(map[string]bool)
					}
					labelIssues[iss.ID][lc.Label] = true
				}
			}

			// Count co-occurrences
			localPairs := make(map[string]int)
			for _, lbls := range labelIssues {
				sorted := make([]string, 0, len(lbls))
				for l := range lbls {
					sorted = append(sorted, l)
				}
				sort.Strings(sorted)
				for i := 0; i < len(sorted); i++ {
					for j := i + 1; j < len(sorted); j++ {
						key := fmt.Sprintf("%s|%s", sorted[i], sorted[j])
						localPairs[key]++
					}
				}
			}

			mu.Lock()
			for key, count := range localPairs {
				pairCounts[key] += count
			}
			mu.Unlock()
		}(db.Name)
	}
	wg.Wait()

	// Convert to sorted list
	for key, count := range pairCounts {
		parts := splitPairKey(key)
		if len(parts) == 2 {
			data.Pairs = append(data.Pairs, labelPair{
				LabelA: parts[0],
				LabelB: parts[1],
				Count:  count,
			})
			labelFreq[parts[0]] += count
			labelFreq[parts[1]] += count
		}
	}

	sort.Slice(data.Pairs, func(i, j int) bool {
		return data.Pairs[i].Count > data.Pairs[j].Count
	})

	data.TotalPairs = len(data.Pairs)

	// Find most frequently paired label
	maxFreq := 0
	for l, f := range labelFreq {
		if f > maxFreq {
			maxFreq = f
			data.TopLabel = l
		}
	}

	// Limit to top 50
	if len(data.Pairs) > 50 {
		data.Pairs = data.Pairs[:50]
	}

	s.render(w, r, "pair-freq", data)
}

func splitPairKey(key string) []string {
	for i, c := range key {
		if c == '|' {
			return []string{key[:i], key[i+1:]}
		}
	}
	return nil
}
