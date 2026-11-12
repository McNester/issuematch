package users

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type User struct {
	ID           bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Username     string        `bson:"username" json:"username"`
	Email        string        `bson:"email" json:"email"`
	PasswordHash string        `bson:"passwordHash" json:"-"`
	Role         string        `bson:"role" json:"role"`
	Bio          string        `bson:"bio" json:"bio"`
	Languages    []string      `bson:"languages" json:"languages"`
	Interests    []string      `bson:"interests" json:"interests"`
	CreatedAt    time.Time     `bson:"createdAt" json:"createdAt"`
}
