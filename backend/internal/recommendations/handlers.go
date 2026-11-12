package recommendations

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"issuematch/internal/platform"
)

func (s *Service) Routes(r chi.Router, jwtSecret string) {
	r.Group(func(r chi.Router) {
		r.Use(platform.RequireAuth(jwtSecret))
		r.Get("/recommendations", s.list)
		r.With(platform.RequireAdmin).Post("/admin/recompute", s.recompute)
	})
}

func (s *Service) list(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	strategy := query.Get("strategy")
	if strategy == "" {
		strategy = "precomputed"
	}
	if strategy != "precomputed" && strategy != "live" {
		platform.WriteError(w, http.StatusBadRequest, "strategy must be precomputed or live")
		return
	}
	limit, err := strconv.Atoi(query.Get("limit"))
	if err != nil || limit < 1 {
		limit = 10
	}
	limit = min(limit, 50)
	useCache := query.Get("cache") != "off"
	userID := platform.UserID(r)

	if useCache {
		if response, found := s.cache.get(userID, strategy, limit); found {
			response.Cached = true
			platform.WriteJSON(w, http.StatusOK, response)
			return
		}
	}
	items, err := s.recommend(r.Context(), userID, strategy, limit)
	if err != nil {
		platform.WriteServerError(w, err)
		return
	}
	response := Response{Items: items, Strategy: strategy}
	if useCache {
		s.cache.put(userID, limit, response)
	}
	platform.WriteJSON(w, http.StatusOK, response)
}

func (s *Service) recompute(w http.ResponseWriter, r *http.Request) {
	started := time.Now()
	count, err := s.Recompute(r.Context())
	if err != nil {
		platform.WriteServerError(w, err)
		return
	}
	platform.WriteJSON(w, http.StatusOK, map[string]any{
		"issues":     count,
		"durationMs": time.Since(started).Milliseconds(),
	})
}
