package interactions

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Store struct {
	interactions *mongo.Collection
	issues       *mongo.Collection
}

func NewStore(db *mongo.Database) Store {
	return Store{db.Collection("interactions"), db.Collection("issues")}
}

func (s Store) EnsureIndexes(ctx context.Context) error {
	_, err := s.interactions.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "userId", Value: 1}, {Key: "createdAt", Value: -1}}},
		{Keys: bson.D{{Key: "issueId", Value: 1}, {Key: "type", Value: 1}}},
		{
			Keys:    bson.D{{Key: "userId", Value: 1}, {Key: "issueId", Value: 1}, {Key: "type", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	})
	return err
}

func (s Store) IssueExists(ctx context.Context, issueID bson.ObjectID) (bool, error) {
	count, err := s.issues.CountDocuments(ctx, bson.M{"_id": issueID})
	return count > 0, err
}

// Record keeps one document per user, issue and type. Recording the same
// interaction again only moves its createdAt forward.
func (s Store) Record(ctx context.Context, userID, issueID bson.ObjectID, interactionType string) (Interaction, error) {
	var recorded Interaction
	err := s.interactions.FindOneAndUpdate(ctx,
		bson.M{"userId": userID, "issueId": issueID, "type": interactionType},
		bson.M{"$set": bson.M{"createdAt": time.Now().UTC()}},
		options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After),
	).Decode(&recorded)
	return recorded, err
}

func (s Store) InsertMany(ctx context.Context, interactions []Interaction) error {
	_, err := s.interactions.InsertMany(ctx, interactions)
	return err
}

func (s Store) History(ctx context.Context, userID bson.ObjectID) ([]HistoryItem, error) {
	cursor, err := s.interactions.Aggregate(ctx, mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"userId": userID}}},
		{{Key: "$sort", Value: bson.M{"createdAt": -1}}},
		{{Key: "$lookup", Value: bson.M{
			"from":         "issues",
			"localField":   "issueId",
			"foreignField": "_id",
			"as":           "issue",
		}}},
		{{Key: "$set", Value: bson.M{"title": bson.M{"$first": "$issue.title"}}}},
		{{Key: "$unset", Value: "issue"}},
	})
	if err != nil {
		return nil, err
	}
	history := []HistoryItem{}
	err = cursor.All(ctx, &history)
	return history, err
}
