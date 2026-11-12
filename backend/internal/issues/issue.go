package issues

import (
	"slices"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

var (
	Categories   = []string{"bug", "feature", "documentation", "refactoring"}
	Difficulties = []string{"easy", "medium", "hard"}
	Statuses     = []string{"open", "closed"}
)

type Issue struct {
	ID          bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Title       string        `bson:"title" json:"title"`
	Description string        `bson:"description" json:"description"`
	Category    string        `bson:"category" json:"category"`
	Language    string        `bson:"language" json:"language"`
	Difficulty  string        `bson:"difficulty" json:"difficulty"`
	Labels      []string      `bson:"labels" json:"labels"`
	Project     Project       `bson:"project" json:"project"`
	Status      string        `bson:"status" json:"status"`
	ExternalID  string        `bson:"externalId,omitempty" json:"externalId,omitempty"`
	CreatedAt   time.Time     `bson:"createdAt" json:"createdAt"`
}

type Project struct {
	Name  string `bson:"name" json:"name"`
	URL   string `bson:"url" json:"url"`
	Stars int    `bson:"stars" json:"stars"`
}

// validationError returns an empty string when the issue is valid.
func (i Issue) validationError() string {
	switch {
	case strings.TrimSpace(i.Title) == "":
		return "title is required"
	case !slices.Contains(Categories, i.Category):
		return "category must be one of: " + strings.Join(Categories, ", ")
	case !slices.Contains(Difficulties, i.Difficulty):
		return "difficulty must be one of: " + strings.Join(Difficulties, ", ")
	case !slices.Contains(Statuses, i.Status):
		return "status must be one of: " + strings.Join(Statuses, ", ")
	}
	return ""
}
