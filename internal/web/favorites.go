package web

import (
	"encoding/json"
	"net/http"
	"sync"
)

func (s *Server) handleFavorites(w http.ResponseWriter, r *http.Request) {
	s.render(w, r, "favorites", struct{}{})
}

type favLookupReq struct {
	DB string `json:"db"`
	ID string `json:"id"`
}

type favLookupResp struct {
	DB       string `json:"db"`
	ID       string `json:"id"`
	Title    string `json:"title"`
	Status   string `json:"status"`
	Priority int    `json:"priority"`
	Assignee string `json:"assignee"`
	Found    bool   `json:"found"`
}

func (s *Server) handleFavoritesLookup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	var reqs []favLookupReq
	if err := json.NewDecoder(r.Body).Decode(&reqs); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}

	if len(reqs) > 100 {
		reqs = reqs[:100]
	}

	ctx := r.Context()
	results := make([]favLookupResp, len(reqs))
	var wg sync.WaitGroup

	for i, req := range reqs {
		results[i] = favLookupResp{DB: req.DB, ID: req.ID}
		if s.ds == nil {
			continue
		}
		wg.Add(1)
		go func(i int, db, id string) {
			defer wg.Done()
			iss, err := s.ds.IssueByID(ctx, db, id)
			if err != nil || iss == nil {
				return
			}
			results[i].Title = iss.Title
			results[i].Status = iss.Status
			results[i].Priority = iss.Priority
			results[i].Assignee = iss.Assignee
			results[i].Found = true
		}(i, req.DB, req.ID)
	}
	wg.Wait()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}
