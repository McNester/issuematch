package interactions

import (
	"net/http"
	"slices"
	"strings"

	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"issuematch/internal/platform"
)

// onRecorded is called with the user's id after each new interaction.
func Routes(r chi.Router, db *mongo.Database, jwtSecret string, onRecorded func(userID bson.ObjectID)) {
	h := handlers{NewStore(db), onRecorded}
	r.Group(func(r chi.Router) {
		r.Use(platform.RequireAuth(jwtSecret))
		r.Post("/interactions", h.record)
		r.Get("/users/me/interactions", h.history)
	})
}

type handlers struct {
	store      Store
	onRecorded func(userID bson.ObjectID)
}

func (h handlers) record(w http.ResponseWriter, r *http.Request) {
	var body struct {
		IssueID string `json:"issueId"`
		Type    string `json:"type"`
	}
	if !platform.ReadJSON(w, r, &body) {
		return
	}
	if !slices.Contains(Types, body.Type) {
		platform.WriteError(w, http.StatusBadRequest, "type must be one of: "+strings.Join(Types, ", "))
		return
	}
	issueID, err := bson.ObjectIDFromHex(body.IssueID)
	if err != nil {
		platform.WriteError(w, http.StatusBadRequest, "issueId is not valid")
		return
	}
	exists, err := h.store.IssueExists(r.Context(), issueID)
	if err != nil {
		platform.WriteServerError(w, err)
		return
	}
	if !exists {
		platform.WriteError(w, http.StatusNotFound, "issue not found")
		return
	}

	recorded, err := h.store.Record(r.Context(), platform.UserID(r), issueID, body.Type)
	if err != nil {
		platform.WriteServerError(w, err)
		return
	}
	h.onRecorded(recorded.UserID)
	platform.WriteJSON(w, http.StatusCreated, recorded)
}

func (h handlers) history(w http.ResponseWriter, r *http.Request) {
	history, err := h.store.History(r.Context(), platform.UserID(r))
	if err != nil {
		platform.WriteServerError(w, err)
		return
	}
	platform.WriteJSON(w, http.StatusOK, history)
}
