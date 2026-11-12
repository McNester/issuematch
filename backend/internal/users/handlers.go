package users

import (
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"issuematch/internal/platform"
)

// onUpdated is called with the user's id after their profile changes.
func Routes(r chi.Router, db *mongo.Database, jwtSecret string, onUpdated func(userID bson.ObjectID)) {
	h := handlers{NewStore(db), onUpdated}
	r.Group(func(r chi.Router) {
		r.Use(platform.RequireAuth(jwtSecret))
		r.Get("/users/me", h.me)
		r.Put("/users/me", h.updateMe)
	})
}

type handlers struct {
	store     Store
	onUpdated func(userID bson.ObjectID)
}

func (h handlers) me(w http.ResponseWriter, r *http.Request) {
	user, err := h.store.FindByID(r.Context(), platform.UserID(r))
	if errors.Is(err, mongo.ErrNoDocuments) {
		platform.WriteError(w, http.StatusNotFound, "user not found")
		return
	}
	if err != nil {
		platform.WriteServerError(w, err)
		return
	}
	platform.WriteJSON(w, http.StatusOK, user)
}

func (h handlers) updateMe(w http.ResponseWriter, r *http.Request) {
	var update ProfileUpdate
	if !platform.ReadJSON(w, r, &update) {
		return
	}
	update.Username = strings.TrimSpace(update.Username)
	update.Email = strings.TrimSpace(update.Email)
	if update.Username == "" || !strings.Contains(update.Email, "@") {
		platform.WriteError(w, http.StatusBadRequest, "username and a valid email are required")
		return
	}
	// Mongo would store nil slices as null and the UI expects arrays.
	if update.Languages == nil {
		update.Languages = []string{}
	}
	if update.Interests == nil {
		update.Interests = []string{}
	}

	user, err := h.store.UpdateProfile(r.Context(), platform.UserID(r), update)
	switch {
	case mongo.IsDuplicateKeyError(err):
		platform.WriteError(w, http.StatusConflict, "username or email is already taken")
	case errors.Is(err, mongo.ErrNoDocuments):
		platform.WriteError(w, http.StatusNotFound, "user not found")
	case err != nil:
		platform.WriteServerError(w, err)
	default:
		h.onUpdated(user.ID)
		platform.WriteJSON(w, http.StatusOK, user)
	}
}
