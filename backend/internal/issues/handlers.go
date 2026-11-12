package issues

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"issuematch/internal/platform"
)

func Routes(r chi.Router, db *mongo.Database, jwtSecret string) {
	h := handlers{NewStore(db)}
	r.Get("/issues", h.list)
	r.Get("/issues/{id}", h.get)
	r.Group(func(r chi.Router) {
		r.Use(platform.RequireAuth(jwtSecret), platform.RequireAdmin)
		r.Post("/issues", h.create)
		r.Put("/issues/{id}", h.update)
		r.Delete("/issues/{id}", h.delete)
	})
}

type handlers struct {
	store Store
}

func (h handlers) list(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	search := Search{
		Text:       query.Get("q"),
		Category:   query.Get("category"),
		Language:   query.Get("language"),
		Difficulty: query.Get("difficulty"),
		Label:      query.Get("label"),
		Page:       positiveInt(query.Get("page"), 1),
		Limit:      min(positiveInt(query.Get("limit"), 20), 100),
	}
	items, total, err := h.store.Search(r.Context(), search)
	if err != nil {
		platform.WriteServerError(w, err)
		return
	}
	platform.WriteJSON(w, http.StatusOK, map[string]any{
		"items": items,
		"total": total,
		"page":  search.Page,
		"limit": search.Limit,
	})
}

func positiveInt(text string, fallback int) int {
	number, err := strconv.Atoi(text)
	if err != nil || number < 1 {
		return fallback
	}
	return number
}

func (h handlers) get(w http.ResponseWriter, r *http.Request) {
	id, ok := idFromPath(w, r)
	if !ok {
		return
	}
	issue, err := h.store.FindByID(r.Context(), id)
	writeIssue(w, http.StatusOK, issue, err)
}

func (h handlers) create(w http.ResponseWriter, r *http.Request) {
	issue, ok := readIssue(w, r)
	if !ok {
		return
	}
	err := h.store.Insert(r.Context(), &issue)
	writeIssue(w, http.StatusCreated, issue, err)
}

func (h handlers) update(w http.ResponseWriter, r *http.Request) {
	id, ok := idFromPath(w, r)
	if !ok {
		return
	}
	issue, ok := readIssue(w, r)
	if !ok {
		return
	}
	updated, err := h.store.Update(r.Context(), id, issue)
	writeIssue(w, http.StatusOK, updated, err)
}

func (h handlers) delete(w http.ResponseWriter, r *http.Request) {
	id, ok := idFromPath(w, r)
	if !ok {
		return
	}
	deleted, err := h.store.Delete(r.Context(), id)
	switch {
	case err != nil:
		platform.WriteServerError(w, err)
	case !deleted:
		platform.WriteError(w, http.StatusNotFound, "issue not found")
	default:
		w.WriteHeader(http.StatusNoContent)
	}
}

func idFromPath(w http.ResponseWriter, r *http.Request) (bson.ObjectID, bool) {
	id, err := bson.ObjectIDFromHex(chi.URLParam(r, "id"))
	if err != nil {
		platform.WriteError(w, http.StatusNotFound, "issue not found")
		return id, false
	}
	return id, true
}

func readIssue(w http.ResponseWriter, r *http.Request) (Issue, bool) {
	issue := Issue{Status: "open", Labels: []string{}}
	if !platform.ReadJSON(w, r, &issue) {
		return issue, false
	}
	if message := issue.validationError(); message != "" {
		platform.WriteError(w, http.StatusBadRequest, message)
		return issue, false
	}
	if issue.Labels == nil {
		issue.Labels = []string{}
	}
	// The client does not choose these.
	issue.ID = bson.ObjectID{}
	issue.ExternalID = ""
	issue.CreatedAt = time.Now().UTC()
	return issue, true
}

func writeIssue(w http.ResponseWriter, status int, issue Issue, err error) {
	switch {
	case errors.Is(err, mongo.ErrNoDocuments):
		platform.WriteError(w, http.StatusNotFound, "issue not found")
	case err != nil:
		platform.WriteServerError(w, err)
	default:
		platform.WriteJSON(w, status, issue)
	}
}
