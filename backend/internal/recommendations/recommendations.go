package recommendations

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"issuematch/internal/issues"
)

type Item struct {
	Issue   issues.Issue `json:"issue"`
	Score   float64      `json:"score"`
	BasedOn string       `json:"basedOn"`
}

type Response struct {
	Items    []Item `json:"items"`
	Strategy string `json:"strategy"`
	Cached   bool   `json:"cached"`
}

type Service struct {
	store       store
	issues      issues.Store
	cache       cache
	recomputing sync.Mutex

	trendingMutex sync.Mutex
	trendingKept  map[string]keptTrending
}

func New(db *mongo.Database) *Service {
	return &Service{
		store: store{
			interactions: db.Collection("interactions"),
			similarities: db.Collection("similarities"),
			users:        db.Collection("users"),
		},
		issues: issues.NewStore(db),
		cache:  cache{entries: map[bson.ObjectID]cacheEntry{}},

		trendingKept: map[string]keptTrending{},
	}
}

// Forget drops the cached recommendations of a user whose history changed.
func (s *Service) Forget(userID bson.ObjectID) {
	s.cache.forget(userID)
}

// Recompute runs the similarity job and returns the number of issues it wrote.
func (s *Service) Recompute(ctx context.Context) (int, error) {
	s.recomputing.Lock()
	defer s.recomputing.Unlock()

	all, err := s.store.allInteractions(ctx)
	if err != nil {
		return 0, err
	}
	everyRating := newMatrix(recentRatingsByUser(all))
	now := time.Now().UTC()
	similarities := make([]similarity, 0, len(everyRating.byIssue))
	for issue := range everyRating.byIssue {
		similarities = append(similarities, similarity{issue, everyRating.neighbors(issue), now})
	}
	if err := s.store.saveSimilarities(ctx, similarities); err != nil {
		return 0, err
	}
	s.cache.clear()
	return len(similarities), nil
}

func (s *Service) RecomputeIfEmpty(ctx context.Context) error {
	filled, err := s.store.hasSimilarities(ctx)
	if err != nil || filled {
		return err
	}
	_, err = s.Recompute(ctx)
	return err
}

// recentRatingsByUser expects interactions sorted by user, newest first.
func recentRatingsByUser(all []interaction) map[bson.ObjectID]ratings {
	byUser := map[bson.ObjectID]ratings{}
	for start := 0; start < len(all); {
		end := start
		for end < len(all) && all[end].UserID == all[start].UserID {
			end++
		}
		byUser[all[start].UserID] = rate(recent(all[start:end]))
		start = end
	}
	return byUser
}

func (s *Service) recommend(ctx context.Context, userID bson.ObjectID, strategy string, limit int) ([]Item, error) {
	history, err := s.store.userHistory(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(history) == 0 {
		return s.popular(ctx, userID, limit)
	}
	rated := fade(rate(history), history, time.Now())
	ratedIssues := make([]bson.ObjectID, 0, len(rated))
	for issue := range rated {
		ratedIssues = append(ratedIssues, issue)
	}

	var neighborsOf map[bson.ObjectID][]Neighbor
	if strategy == "live" {
		neighborsOf, err = s.liveNeighbors(ctx, ratedIssues)
	} else {
		neighborsOf, err = s.store.neighborsOf(ctx, ratedIssues)
	}
	if err != nil {
		return nil, err
	}
	ranked := score(rated, neighborsOf, limit)
	// Issues that nobody else touched yet have no neighbors. A user whose
	// history holds only such issues gets the popular ones, like a new user.
	if len(ranked) == 0 {
		return s.popular(ctx, userID, limit)
	}

	// One query loads the recommended issues and the issues they are based on.
	var needed []bson.ObjectID
	for _, result := range ranked {
		needed = append(needed, result.issueID, result.basedOn)
	}
	loaded, err := s.issues.FindByIDs(ctx, needed)
	if err != nil {
		return nil, err
	}
	byID := map[bson.ObjectID]issues.Issue{}
	for _, issue := range loaded {
		byID[issue.ID] = issue
	}

	items := []Item{}
	for _, result := range ranked {
		if issue, found := byID[result.issueID]; found && issue.Status == "open" {
			items = append(items, Item{issue, result.score, byID[result.basedOn].Title})
		}
	}
	return items, nil
}

// liveNeighbors is the slow baseline: it rebuilds the similarities from the
// interactions collection on every request instead of reading the stored ones.
func (s *Service) liveNeighbors(ctx context.Context, issueIDs []bson.ObjectID) (map[bson.ObjectID][]Neighbor, error) {
	all, err := s.store.allInteractions(ctx)
	if err != nil {
		return nil, err
	}
	everyRating := newMatrix(recentRatingsByUser(all))
	neighborsOf := map[bson.ObjectID][]Neighbor{}
	for _, issue := range issueIDs {
		neighborsOf[issue] = everyRating.neighbors(issue)
	}
	return neighborsOf, nil
}

// popular is the cold start: a user without history has nothing to compare.
func (s *Service) popular(ctx context.Context, userID bson.ObjectID, limit int) ([]Item, error) {
	languages, err := s.store.userLanguages(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.popularItems(ctx, time.Time{}, languages, limit)
}

const (
	trendingWindow   = 7 * 24 * time.Hour
	trendingLifetime = 5 * time.Minute
)

// trending returns the most popular open issues of the last week, those in
// the given languages first. Everyone sees the same list, so it is computed
// once per language set and kept for a few minutes.
func (s *Service) trending(ctx context.Context, languages []string, limit int) ([]Item, error) {
	key := fmt.Sprint(languages, limit)
	s.trendingMutex.Lock()
	defer s.trendingMutex.Unlock()
	if kept, found := s.trendingKept[key]; found && time.Now().Before(kept.expires) {
		return kept.items, nil
	}
	items, err := s.popularItems(ctx, time.Now().Add(-trendingWindow), languages, limit)
	if err != nil {
		return nil, err
	}
	s.trendingKept[key] = keptTrending{items, time.Now().Add(trendingLifetime)}
	return items, nil
}

type keptTrending struct {
	items   []Item
	expires time.Time
}

func (s *Service) popularItems(ctx context.Context, since time.Time, languages []string, limit int) ([]Item, error) {
	found, err := s.store.popular(ctx, since, languages, limit)
	if err != nil {
		return nil, err
	}
	items := []Item{}
	for _, popular := range found {
		items = append(items, Item{Issue: popular.Issue, Score: popular.Score})
	}
	return items, nil
}
