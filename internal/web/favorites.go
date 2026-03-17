package web

import "net/http"

func (s *Server) handleFavorites(w http.ResponseWriter, r *http.Request) {
	s.render(w, r, "favorites", struct{}{})
}
