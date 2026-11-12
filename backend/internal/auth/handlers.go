package auth

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"issuematch/internal/platform"
	"issuematch/internal/users"
)

func Routes(r chi.Router, db *mongo.Database, jwtSecret string) {
	h := handlers{users.NewStore(db), jwtSecret}
	r.Post("/auth/register", h.register)
	r.Post("/auth/login", h.login)
}

type handlers struct {
	users     users.Store
	jwtSecret string
}

func (h handlers) register(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Username  string   `json:"username"`
		Email     string   `json:"email"`
		Password  string   `json:"password"`
		Languages []string `json:"languages"`
	}
	if !platform.ReadJSON(w, r, &body) {
		return
	}
	body.Username = strings.TrimSpace(body.Username)
	body.Email = strings.TrimSpace(body.Email)
	if body.Username == "" || !strings.Contains(body.Email, "@") {
		platform.WriteError(w, http.StatusBadRequest, "username and a valid email are required")
		return
	}
	if len(body.Password) < 8 {
		platform.WriteError(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}
	if body.Languages == nil {
		body.Languages = []string{}
	}

	hash, err := HashPassword(body.Password)
	if err != nil {
		platform.WriteServerError(w, err)
		return
	}
	user := users.User{
		Username:     body.Username,
		Email:        body.Email,
		PasswordHash: hash,
		Role:         "user",
		Languages:    body.Languages,
		Interests:    []string{},
		CreatedAt:    time.Now().UTC(),
	}
	err = h.users.Insert(r.Context(), &user)
	if mongo.IsDuplicateKeyError(err) {
		platform.WriteError(w, http.StatusConflict, "username or email is already taken")
		return
	}
	if err != nil {
		platform.WriteServerError(w, err)
		return
	}
	h.writeSession(w, http.StatusCreated, user)
}

func (h handlers) login(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if !platform.ReadJSON(w, r, &body) {
		return
	}
	user, err := h.users.FindByEmail(r.Context(), strings.TrimSpace(body.Email))
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		platform.WriteServerError(w, err)
		return
	}
	if err != nil || !passwordMatches(user.PasswordHash, body.Password) {
		platform.WriteError(w, http.StatusUnauthorized, "wrong email or password")
		return
	}
	h.writeSession(w, http.StatusOK, user)
}

func (h handlers) writeSession(w http.ResponseWriter, status int, user users.User) {
	token, err := newToken(h.jwtSecret, user)
	if err != nil {
		platform.WriteServerError(w, err)
		return
	}
	platform.WriteJSON(w, status, map[string]any{"token": token, "user": user})
}
