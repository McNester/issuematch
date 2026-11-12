package auth

import (
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"

	"issuematch/internal/platform"
	"issuematch/internal/users"
)

func TestPasswordHashing(t *testing.T) {
	hash, err := HashPassword("password123")
	if err != nil {
		t.Fatal(err)
	}
	if hash == "password123" {
		t.Error("password is stored as plain text")
	}
	if !passwordMatches(hash, "password123") {
		t.Error("the right password does not match its hash")
	}
	if passwordMatches(hash, "password124") {
		t.Error("a wrong password matches the hash")
	}
}

func TestTokenRoundTrip(t *testing.T) {
	user := users.User{ID: bson.NewObjectID(), Role: "admin"}
	token, err := newToken("secret", user)
	if err != nil {
		t.Fatal(err)
	}

	claims, err := platform.ParseToken("secret", token)
	if err != nil {
		t.Fatal(err)
	}
	if claims.Subject != user.ID.Hex() || claims.Role != "admin" {
		t.Errorf("got subject %q and role %q", claims.Subject, claims.Role)
	}

	if _, err := platform.ParseToken("other secret", token); err == nil {
		t.Error("a token signed with another secret is accepted")
	}
}
