package recommendations

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"issuematch/internal/platform"
)

func (s *Service) Routes(r chi.Router, jwtSecret string) {
	r.Group(func(r chi.Router) {
		r.Use(platform.RequireAuth(jwtSecret))
		r.Get("/recommendations", s.list)
		r.Get("/recommendations/trending", s.listTrending)
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
	limit := limitFrom(r, 10)
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

func (s *Service) listTrending(w http.ResponseWriter, r *http.Request) {
	languages := []string{}
	if list := r.URL.Query().Get("languages"); list != "" {
		languages = strings.Split(list, ",")
	}
	items, err := s.trending(r.Context(), languages, limitFrom(r, 5))
	if err != nil {
		platform.WriteServerError(w, err)
		return
	}
	platform.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

func limitFrom(r *http.Request, fallback int) int {
	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil || limit < 1 {
		return fallback
	}
	return min(limit, 50)
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
