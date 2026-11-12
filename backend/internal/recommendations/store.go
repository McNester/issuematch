package recommendations

import (
	"context"
	"errors"
	"slices"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"issuematch/internal/issues"
)

const writeBatch = 1000

type similarity struct {
	IssueID   bson.ObjectID `bson:"_id"`
	Neighbors []Neighbor    `bson:"neighbors"`
	UpdatedAt time.Time     `bson:"updatedAt"`
}

type store struct {
	interactions *mongo.Collection
	similarities *mongo.Collection
	users        *mongo.Collection
}

// userHistory is served by the index { userId: 1, createdAt: -1 }.
func (s store) userHistory(ctx context.Context, userID bson.ObjectID) ([]interaction, error) {
	return s.findInteractions(ctx, bson.M{"userId": userID})
}

func (s store) allInteractions(ctx context.Context) ([]interaction, error) {
	return s.findInteractions(ctx, bson.M{})
}

func (s store) findInteractions(ctx context.Context, filter bson.M) ([]interaction, error) {
	cursor, err := s.interactions.Find(ctx, filter, options.Find().
		SetSort(bson.D{{Key: "userId", Value: 1}, {Key: "createdAt", Value: -1}}).
		SetProjection(bson.M{"_id": 0}))
	if err != nil {
		return nil, err
	}
	var found []interaction
	err = cursor.All(ctx, &found)
	return found, err
}

func (s store) neighborsOf(ctx context.Context, issueIDs []bson.ObjectID) (map[bson.ObjectID][]Neighbor, error) {
	cursor, err := s.similarities.Find(ctx, bson.M{"_id": bson.M{"$in": issueIDs}})
	if err != nil {
		return nil, err
	}
	var found []similarity
	if err := cursor.All(ctx, &found); err != nil {
		return nil, err
	}
	neighborsOf := map[bson.ObjectID][]Neighbor{}
	for _, item := range found {
		neighborsOf[item.IssueID] = item.Neighbors
	}
	return neighborsOf, nil
}

func (s store) saveSimilarities(ctx context.Context, similarities []similarity) error {
	for batch := range slices.Chunk(similarities, writeBatch) {
		writes := make([]mongo.WriteModel, len(batch))
		for i, item := range batch {
			writes[i] = mongo.NewReplaceOneModel().
				SetFilter(bson.M{"_id": item.IssueID}).
				SetReplacement(item).
				SetUpsert(true)
		}
		if _, err := s.similarities.BulkWrite(ctx, writes, options.BulkWrite().SetOrdered(false)); err != nil {
			return err
		}
	}
	return nil
}

func (s store) hasSimilarities(ctx context.Context) (bool, error) {
	count, err := s.similarities.EstimatedDocumentCount(ctx)
	return count > 0, err
}

func (s store) userLanguages(ctx context.Context, userID bson.ObjectID) ([]string, error) {
	var user struct {
		Languages []string `bson:"languages"`
	}
	err := s.users.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	// A token can outlive its user, for example after a reseed.
	if errors.Is(err, mongo.ErrNoDocuments) {
		err = nil
	}
	if user.Languages == nil {
		user.Languages = []string{}
	}
	return user.Languages, err
}

type popularIssue struct {
	Issue issues.Issue `bson:"issue"`
	Score float64      `bson:"score"`
}

// popular returns the open issues with the highest sum of interaction
// weights since the given time, those written in one of the given languages first.
func (s store) popular(ctx context.Context, since time.Time, languages []string, limit int) ([]popularIssue, error) {
	// People type "javascript" in their profile and issues say "JavaScript".
	lowercased := bson.A{}
	for _, language := range languages {
		lowercased = append(lowercased, strings.ToLower(strings.TrimSpace(language)))
	}
	var weightOfType bson.A
	for interactionType, weight := range weights {
		weightOfType = append(weightOfType, bson.M{
			"case": bson.M{"$eq": bson.A{"$type", interactionType}},
			"then": weight,
		})
	}
	cursor, err := s.interactions.Aggregate(ctx, mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"createdAt": bson.M{"$gte": since}}}},
		{{Key: "$group", Value: bson.M{
			"_id":   "$issueId",
			"score": bson.M{"$sum": bson.M{"$switch": bson.M{"branches": weightOfType, "default": 0}}},
		}}},
		{{Key: "$lookup", Value: bson.M{"from": "issues", "localField": "_id", "foreignField": "_id", "as": "issue"}}},
		{{Key: "$unwind", Value: "$issue"}},
		{{Key: "$match", Value: bson.M{"issue.status": "open"}}},
		{{Key: "$set", Value: bson.M{"inUserLanguage": bson.M{"$in": bson.A{bson.M{"$toLower": "$issue.language"}, lowercased}}}}},
		{{Key: "$sort", Value: bson.D{{Key: "inUserLanguage", Value: -1}, {Key: "score", Value: -1}, {Key: "_id", Value: 1}}}},
		{{Key: "$limit", Value: limit}},
	})
	if err != nil {
		return nil, err
	}
	var found []popularIssue
	err = cursor.All(ctx, &found)
	return found, err
}
