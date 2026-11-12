package interactions

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

var Types = []string{"view", "bookmark", "working", "solved"}

type Interaction struct {
	ID        bson.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    bson.ObjectID `bson:"userId" json:"userId"`
	IssueID   bson.ObjectID `bson:"issueId" json:"issueId"`
	Type      string        `bson:"type" json:"type"`
	CreatedAt time.Time     `bson:"createdAt" json:"createdAt"`
}

type HistoryItem struct {
	ID        bson.ObjectID `bson:"_id" json:"id"`
	IssueID   bson.ObjectID `bson:"issueId" json:"issueId"`
	Type      string        `bson:"type" json:"type"`
	CreatedAt time.Time     `bson:"createdAt" json:"createdAt"`
	Title     string        `bson:"title" json:"title"`
}
