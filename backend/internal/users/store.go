package users

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Store struct {
	users *mongo.Collection
}

func NewStore(db *mongo.Database) Store {
	return Store{db.Collection("users")}
}

func (s Store) EnsureIndexes(ctx context.Context) error {
	_, err := s.users.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "email", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "username", Value: 1}}, Options: options.Index().SetUnique(true)},
	})
	return err
}

func (s Store) Insert(ctx context.Context, user *User) error {
	user.ID = bson.NewObjectID()
	_, err := s.users.InsertOne(ctx, user)
	return err
}

func (s Store) InsertMany(ctx context.Context, users []User) error {
	_, err := s.users.InsertMany(ctx, users)
	return err
}

func (s Store) FindByID(ctx context.Context, id bson.ObjectID) (User, error) {
	var user User
	err := s.users.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	return user, err
}

func (s Store) FindByEmail(ctx context.Context, email string) (User, error) {
	var user User
	err := s.users.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	return user, err
}

type ProfileUpdate struct {
	Username  string   `bson:"username" json:"username"`
	Email     string   `bson:"email" json:"email"`
	Bio       string   `bson:"bio" json:"bio"`
	Languages []string `bson:"languages" json:"languages"`
	Interests []string `bson:"interests" json:"interests"`
}

func (s Store) UpdateProfile(ctx context.Context, id bson.ObjectID, update ProfileUpdate) (User, error) {
	var user User
	err := s.users.FindOneAndUpdate(ctx, bson.M{"_id": id}, bson.M{"$set": update},
		options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&user)
	return user, err
}
