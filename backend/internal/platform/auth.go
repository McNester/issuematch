package platform

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Claims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

func ParseToken(secret, token string) (Claims, error) {
	var claims Claims
	_, err := jwt.ParseWithClaims(token, &claims, func(*jwt.Token) (any, error) {
		return []byte(secret), nil
	}, jwt.WithValidMethods([]string{"HS256"}))
	return claims, err
}

type caller struct {
	id   bson.ObjectID
	role string
}

type callerKey struct{}

func RequireAuth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, found := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
			if !found {
				WriteError(w, http.StatusUnauthorized, "missing token")
				return
			}
			claims, err := ParseToken(secret, token)
			if err != nil {
				WriteError(w, http.StatusUnauthorized, "invalid token")
				return
			}
			id, err := bson.ObjectIDFromHex(claims.Subject)
			if err != nil {
				WriteError(w, http.StatusUnauthorized, "invalid token")
				return
			}
			ctx := context.WithValue(r.Context(), callerKey{}, caller{id, claims.Role})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireAdmin must run after RequireAuth.
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Context().Value(callerKey{}).(caller).role != "admin" {
			WriteError(w, http.StatusForbidden, "admin only")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// UserID works only behind RequireAuth.
func UserID(r *http.Request) bson.ObjectID {
	return r.Context().Value(callerKey{}).(caller).id
}
