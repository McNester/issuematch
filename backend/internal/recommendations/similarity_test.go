package recommendations

import (
	"math"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
)

var (
	alice, bob, carol    = bson.NewObjectID(), bson.NewObjectID(), bson.NewObjectID()
	poolBug, cacheBug    = bson.NewObjectID(), bson.NewObjectID()
	docsPage, exportTask = bson.NewObjectID(), bson.NewObjectID()
)

// Alice and Bob both like the two bugs. Carol only cares about the docs page.
func tinyMatrix() matrix {
	return newMatrix(map[bson.ObjectID]ratings{
		alice: {poolBug: 5, cacheBug: 3},
		bob:   {poolBug: 4, cacheBug: 4, exportTask: 1},
		carol: {docsPage: 5, exportTask: 1},
	})
}

func TestRateKeepsTheHighestWeight(t *testing.T) {
	rated := rate([]interaction{
		{IssueID: poolBug, Type: "view"},
		{IssueID: poolBug, Type: "solved"},
		{IssueID: poolBug, Type: "bookmark"},
		{IssueID: cacheBug, Type: "view"},
	})
	if rated[poolBug] != 5 || rated[cacheBug] != 1 {
		t.Errorf("got %v", rated)
	}
}

func TestCosineSimilarity(t *testing.T) {
	neighbors := tinyMatrix().neighbors(poolBug)

	if len(neighbors) != 2 {
		t.Fatalf("got %d neighbors, want 2: the docs page shares no user with the pool bug", len(neighbors))
	}
	// (5*3 + 4*4) / (sqrt(25+16) * sqrt(9+16))
	wantCacheBug := 31 / (math.Sqrt(41) * 5)
	if neighbors[0].IssueID != cacheBug || !almostEqual(neighbors[0].Score, wantCacheBug) {
		t.Errorf("closest neighbor is %v, want the cache bug with %f", neighbors[0], wantCacheBug)
	}
	// (4*1) / (sqrt(41) * sqrt(1+1))
	wantExportTask := 4 / (math.Sqrt(41) * math.Sqrt(2))
	if neighbors[1].IssueID != exportTask || !almostEqual(neighbors[1].Score, wantExportTask) {
		t.Errorf("second neighbor is %v, want the export task with %f", neighbors[1], wantExportTask)
	}
}

func TestScoreSkipsSeenIssuesAndSumsOverRatedOnes(t *testing.T) {
	neighborsOf := map[bson.ObjectID][]Neighbor{
		poolBug:  {{cacheBug, 0.9}, {exportTask, 0.5}},
		docsPage: {{exportTask, 0.2}, {cacheBug, 0.1}},
	}
	rated := ratings{poolBug: 5, docsPage: 1}

	ranked := score(rated, neighborsOf, 10)

	if len(ranked) != 2 {
		t.Fatalf("got %d results, want 2", len(ranked))
	}
	// 0.9*5 + 0.1*1
	if ranked[0].issueID != cacheBug || !almostEqual(ranked[0].score, 4.6) || ranked[0].basedOn != poolBug {
		t.Errorf("first result is %+v", ranked[0])
	}
	// 0.5*5 + 0.2*1
	if ranked[1].issueID != exportTask || !almostEqual(ranked[1].score, 2.7) {
		t.Errorf("second result is %+v", ranked[1])
	}

	seenCacheBug := ratings{poolBug: 5, cacheBug: 1}
	for _, result := range score(seenCacheBug, neighborsOf, 10) {
		if result.issueID == cacheBug {
			t.Error("an issue the user already rated is recommended")
		}
	}
}

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) < 1e-9
}
