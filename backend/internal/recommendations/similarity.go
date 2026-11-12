package recommendations

import (
	"bytes"
	"cmp"
	"math"
	"slices"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

const (
	neighborsPerIssue = 20
	// Without this limit one very active user creates millions of issue pairs.
	recentIssuesPerUser = 50
	ratingHalfLifeDays  = 7
)

var weights = map[string]float64{"view": 1, "bookmark": 3, "working": 4, "solved": 5}

type interaction struct {
	UserID    bson.ObjectID `bson:"userId"`
	IssueID   bson.ObjectID `bson:"issueId"`
	Type      string        `bson:"type"`
	CreatedAt time.Time     `bson:"createdAt"`
}

// ratings maps an issue to one user's rating of it.
type ratings map[bson.ObjectID]float64

// rate turns one user's interactions into ratings: the highest weight per issue.
func rate(history []interaction) ratings {
	rated := ratings{}
	for _, item := range history {
		rated[item.IssueID] = max(rated[item.IssueID], weights[item.Type])
	}
	return rated
}

// fade lowers the rating of issues the user has not touched for a while, so
// that what they did today shapes the recommendations more than what they did
// last month. Age counts in whole days to keep the result stable within a day.
// The history must be sorted newest first.
func fade(rated ratings, history []interaction, now time.Time) ratings {
	faded := ratings{}
	for _, item := range history {
		if _, done := faded[item.IssueID]; done {
			continue
		}
		daysAgo := math.Floor(now.Sub(item.CreatedAt).Hours() / 24)
		faded[item.IssueID] = rated[item.IssueID] * math.Pow(0.5, max(daysAgo, 0)/ratingHalfLifeDays)
	}
	return faded
}

// recent keeps only the interactions with the user's most recently touched issues.
// The history must be sorted newest first.
func recent(history []interaction) []interaction {
	issues := map[bson.ObjectID]bool{}
	for index, item := range history {
		issues[item.IssueID] = true
		if len(issues) > recentIssuesPerUser {
			return history[:index]
		}
	}
	return history
}

type Neighbor struct {
	IssueID bson.ObjectID `bson:"issueId"`
	Score   float64       `bson:"score"`
}

// matrix holds every rating twice, by user and by issue, so that both
// directions can be walked without scanning.
type matrix struct {
	byUser  map[bson.ObjectID]ratings
	byIssue map[bson.ObjectID]map[bson.ObjectID]float64
	norms   map[bson.ObjectID]float64
}

func newMatrix(byUser map[bson.ObjectID]ratings) matrix {
	byIssue := map[bson.ObjectID]map[bson.ObjectID]float64{}
	sumsOfSquares := map[bson.ObjectID]float64{}
	for user, rated := range byUser {
		for issue, rating := range rated {
			if byIssue[issue] == nil {
				byIssue[issue] = map[bson.ObjectID]float64{}
			}
			byIssue[issue][user] = rating
			sumsOfSquares[issue] += rating * rating
		}
	}
	norms := map[bson.ObjectID]float64{}
	for issue, sum := range sumsOfSquares {
		norms[issue] = math.Sqrt(sum)
	}
	return matrix{byUser, byIssue, norms}
}

// neighbors returns the issues most similar to the given one by cosine similarity.
func (m matrix) neighbors(issue bson.ObjectID) []Neighbor {
	dotProducts := map[bson.ObjectID]float64{}
	for user, rating := range m.byIssue[issue] {
		for other, otherRating := range m.byUser[user] {
			if other != issue {
				dotProducts[other] += rating * otherRating
			}
		}
	}

	found := make([]Neighbor, 0, len(dotProducts))
	for other, dotProduct := range dotProducts {
		found = append(found, Neighbor{other, dotProduct / (m.norms[issue] * m.norms[other])})
	}
	slices.SortFunc(found, func(a, b Neighbor) int {
		return cmp.Or(cmp.Compare(b.Score, a.Score), compareIDs(a.IssueID, b.IssueID))
	})
	return found[:min(len(found), neighborsPerIssue)]
}

type scored struct {
	issueID bson.ObjectID
	score   float64
	// basedOn is the rated issue that added the most to the score.
	basedOn bson.ObjectID
}

// score ranks the neighbors of the issues a user rated:
// score(j) = sum over rated issues i of sim(i, j) * rating(i).
func score(rated ratings, neighborsOf map[bson.ObjectID][]Neighbor, limit int) []scored {
	// Map order is random and float sums depend on order. A fixed order makes
	// the live and precomputed strategies give identical scores.
	ratedIssues := make([]bson.ObjectID, 0, len(rated))
	for issue := range rated {
		ratedIssues = append(ratedIssues, issue)
	}
	slices.SortFunc(ratedIssues, compareIDs)

	candidates := map[bson.ObjectID]*scored{}
	largestPart := map[bson.ObjectID]float64{}
	for _, issue := range ratedIssues {
		for _, neighbor := range neighborsOf[issue] {
			if _, seen := rated[neighbor.IssueID]; seen {
				continue
			}
			candidate := candidates[neighbor.IssueID]
			if candidate == nil {
				candidate = &scored{issueID: neighbor.IssueID}
				candidates[neighbor.IssueID] = candidate
			}
			part := neighbor.Score * rated[issue]
			candidate.score += part
			if part > largestPart[neighbor.IssueID] {
				largestPart[neighbor.IssueID] = part
				candidate.basedOn = issue
			}
		}
	}

	ranked := make([]scored, 0, len(candidates))
	for _, candidate := range candidates {
		ranked = append(ranked, *candidate)
	}
	slices.SortFunc(ranked, func(a, b scored) int {
		return cmp.Or(cmp.Compare(b.score, a.score), compareIDs(a.issueID, b.issueID))
	})
	return ranked[:min(len(ranked), limit)]
}

func compareIDs(a, b bson.ObjectID) int {
	return bytes.Compare(a[:], b[:])
}
