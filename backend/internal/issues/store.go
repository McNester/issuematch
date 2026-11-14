package issues

import (
	"context"
	"errors"
	"regexp"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Store struct {
	issues *mongo.Collection
}

func NewStore(db *mongo.Database) Store {
	return Store{db.Collection("issues")}
}

func (s Store) EnsureIndexes(ctx context.Context) error {
	_, err := s.issues.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "title", Value: "text"}, {Key: "description", Value: "text"}},
			// By default a text index reads the document's "language" field as the
			// human language of the text and rejects values such as "Go".
			Options: options.Index().SetLanguageOverride("textLanguage"),
		},
		{Keys: bson.D{{Key: "category", Value: 1}, {Key: "language", Value: 1}, {Key: "difficulty", Value: 1}}},
		{Keys: bson.D{{Key: "labels", Value: 1}}},
		{Keys: bson.D{{Key: "externalId", Value: 1}}, Options: options.Index().SetUnique(true).SetSparse(true)},
	})
	return err
}

type Search struct {
	Text       string
	Category   string
	Language   string
	Difficulty string
	Label      string
	Page       int
	Limit      int
}

func (s Store) Search(ctx context.Context, search Search) ([]Issue, int64, error) {
	newestFirst := bson.D{{Key: "createdAt", Value: -1}}
	if search.Text == "" {
		return s.find(ctx, search, bson.M{}, newestFirst)
	}

	byTextScore := bson.D{{Key: "score", Value: bson.M{"$meta": "textScore"}}}
	items, total, err := s.find(ctx, search, bson.M{"$text": bson.M{"$search": search.Text}}, byTextScore)

	// The benchmark drops the text index to measure its effect. Without it
	// $text fails, so the search falls back to scanning with a regex.
	var commandError mongo.CommandError
	if errors.As(err, &commandError) && commandError.Name == "IndexNotFound" {
		pattern := bson.M{"$regex": regexp.QuoteMeta(search.Text), "$options": "i"}
		anyField := bson.M{"$or": bson.A{bson.M{"title": pattern}, bson.M{"description": pattern}}}
		return s.find(ctx, search, anyField, newestFirst)
	}
	return items, total, err
}

func (s Store) find(ctx context.Context, search Search, filter bson.M, sort bson.D) ([]Issue, int64, error) {
	if search.Category != "" {
		filter["category"] = search.Category
	}
	if search.Language != "" {
		filter["language"] = search.Language
	}
	if search.Difficulty != "" {
		filter["difficulty"] = search.Difficulty
	}
	if search.Label != "" {
		filter["labels"] = search.Label
	}

	total, err := s.issues.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	cursor, err := s.issues.Find(ctx, filter, options.Find().
		SetSort(sort).
		SetSkip(int64((search.Page-1)*search.Limit)).
		SetLimit(int64(search.Limit)))
	if err != nil {
		return nil, 0, err
	}
	items := []Issue{}
	err = cursor.All(ctx, &items)
	return items, total, err
}

func (s Store) FindByID(ctx context.Context, id bson.ObjectID) (Issue, error) {
	var issue Issue
	err := s.issues.FindOne(ctx, bson.M{"_id": id}).Decode(&issue)
	return issue, err
}

func (s Store) FindByIDs(ctx context.Context, ids []bson.ObjectID) ([]Issue, error) {
	cursor, err := s.issues.Find(ctx, bson.M{"_id": bson.M{"$in": ids}})
	if err != nil {
		return nil, err
	}
	var found []Issue
	err = cursor.All(ctx, &found)
	return found, err
}

// Related returns open issues with the same language and category, newest first.
func (s Store) Related(ctx context.Context, issue Issue, limit int) ([]Issue, error) {
	cursor, err := s.issues.Find(ctx,
		bson.M{"category": issue.Category, "language": issue.Language, "status": "open", "_id": bson.M{"$ne": issue.ID}},
		options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}, {Key: "_id", Value: 1}}).SetLimit(int64(limit)))
	if err != nil {
		return nil, err
	}
	var related []Issue
	err = cursor.All(ctx, &related)
	return related, err
}

func (s Store) Insert(ctx context.Context, issue *Issue) error {
	issue.ID = bson.NewObjectID()
	_, err := s.issues.InsertOne(ctx, issue)
	return err
}

func (s Store) InsertMany(ctx context.Context, issues []Issue) error {
	_, err := s.issues.InsertMany(ctx, issues)
	return err
}

// Update replaces the editable fields and keeps _id, externalId and createdAt.
func (s Store) Update(ctx context.Context, id bson.ObjectID, issue Issue) (Issue, error) {
	fields := bson.M{
		"title":       issue.Title,
		"description": issue.Description,
		"category":    issue.Category,
		"language":    issue.Language,
		"difficulty":  issue.Difficulty,
		"labels":      issue.Labels,
		"project":     issue.Project,
		"status":      issue.Status,
	}
	var updated Issue
	err := s.issues.FindOneAndUpdate(ctx, bson.M{"_id": id}, bson.M{"$set": fields},
		options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&updated)
	return updated, err
}

// UpsertImported inserts an imported issue or refreshes the one with the same externalId.
func (s Store) UpsertImported(ctx context.Context, issue Issue) error {
	_, err := s.issues.UpdateOne(ctx, bson.M{"externalId": issue.ExternalID}, bson.M{"$set": issue},
		options.UpdateOne().SetUpsert(true))
	return err
}

func (s Store) Delete(ctx context.Context, id bson.ObjectID) (bool, error) {
	result, err := s.issues.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return false, err
	}
	return result.DeletedCount == 1, nil
}

func (s Store) All(ctx context.Context) ([]Issue, error) {
	cursor, err := s.issues.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	var all []Issue
	err = cursor.All(ctx, &all)
	return all, err
}
