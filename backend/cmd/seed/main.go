package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand/v2"
	"slices"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"

	"issuematch/internal/auth"
	"issuematch/internal/feed"
	"issuematch/internal/interactions"
	"issuematch/internal/issues"
	"issuematch/internal/platform"
	"issuematch/internal/recommendations"
	"issuematch/internal/users"
)

type size struct {
	issues, users, interactions int
}

var sizes = map[string]size{
	"small":  {1_000, 200, 20_000},
	"medium": {5_000, 1_000, 200_000},
	"large":  {10_000, 5_000, 1_000_000},
}

// One view brings on average 0.3 bookmarks, 0.1 working and 0.05 solved.
const interactionsPerView = 1.45

const insertBatch = 10_000

func main() {
	sizeName := flag.String("size", "small", "small, medium or large")
	realOnly := flag.Bool("real", false, "keep only issues imported from GitHub, generate none")
	flag.Parse()
	size, known := sizes[*sizeName]
	if !known {
		log.Fatal("unknown size: ", *sizeName)
	}

	config, err := platform.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}
	db, err := platform.ConnectMongo(config.MongoURI)
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	must(db.Drop(ctx))

	userStore, issueStore, interactionStore := users.NewStore(db), issues.NewStore(db), interactions.NewStore(db)
	must(userStore.EnsureIndexes(ctx))
	must(issueStore.EnsureIndexes(ctx))
	must(interactionStore.EnsureIndexes(ctx))

	// A fixed seed makes two runs produce the same data.
	random := rand.New(rand.NewPCG(1, 2))

	// One page holds 100 entries and many of them are pull requests, so a
	// seed with only real issues reads more pages.
	feedPages := 1
	if *realOnly {
		feedPages = 5
	}
	imported, err := feed.Run(ctx, issueStore, config.FeedRepos, config.GitHubToken, feedPages)
	if err != nil {
		log.Println("feed failed:", err)
	}
	if *realOnly && imported == 0 {
		log.Fatal("no issues imported from GitHub; set GITHUB_TOKEN in backend/.env and try again")
	}
	if missing := size.issues - imported; missing > 0 && !*realOnly {
		projects := generateProjects(random)
		generated := make([]issues.Issue, missing)
		for i := range generated {
			generated[i] = generateIssue(random, projects)
		}
		must(issueStore.InsertMany(ctx, generated))
	}
	allIssues, err := issueStore.All(ctx)
	must(err)

	admin := newUser("admin", "admin@issuematch.local", hash("admin12345"), "admin", nil, nil)
	demo := newUser("demo", "demo@issuematch.local", hash("demo12345"), "user", []string{"Go", "Java"}, []string{"backend", "databases"})
	demo.Bio = "Backend developer"

	// Hashing is slow on purpose, so the synthetic users share one hash.
	sharedHash := hash("password123")
	people := []users.User{admin, demo}
	for i := range size.users {
		name := fmt.Sprintf("user%04d", i+1)
		people = append(people, newUser(name, name+"@example.com", sharedHash, "user",
			pickSome(random, languages), pickSome(random, issues.Categories)))
	}
	must(userStore.InsertMany(ctx, people))

	generator := interactionGenerator{random: random, issues: allIssues}
	viewsPerUser := int(float64(size.interactions) / float64(size.users) / interactionsPerView)
	var pending []interactions.Interaction
	total := 0
	flush := func() {
		if len(pending) > 0 {
			must(interactionStore.InsertMany(ctx, pending))
			total += len(pending)
			pending = pending[:0]
		}
	}
	for _, person := range people[2:] {
		// Synthetic users keep their preferred categories in interests.
		pending = append(pending, generator.forUser(person.ID, person.Languages, person.Interests, viewsPerUser)...)
		if len(pending) >= insertBatch {
			flush()
		}
	}
	pending = append(pending, generator.forUser(demo.ID, demo.Languages, []string{"bug", "feature"}, 20)...)
	flush()

	withNeighbors, err := recommendations.New(db).Recompute(ctx)
	must(err)

	fmt.Printf("issues: %d (%d imported from GitHub)\n", len(allIssues), imported)
	fmt.Printf("users: %d\n", len(people))
	fmt.Printf("interactions: %d\n", total)
	fmt.Printf("similarities: %d\n", withNeighbors)
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func hash(password string) string {
	hashed, err := auth.HashPassword(password)
	must(err)
	return hashed
}

func newUser(username, email, passwordHash, role string, languages, interests []string) users.User {
	return users.User{
		ID:           bson.NewObjectID(),
		Username:     username,
		Email:        email,
		PasswordHash: passwordHash,
		Role:         role,
		Languages:    slices.Concat([]string{}, languages),
		Interests:    slices.Concat([]string{}, interests),
		CreatedAt:    time.Now().UTC(),
	}
}

// pickSome returns one or two different values.
func pickSome(random *rand.Rand, from []string) []string {
	order := random.Perm(len(from))
	picked := []string{from[order[0]]}
	if random.IntN(2) == 0 {
		picked = append(picked, from[order[1]])
	}
	return picked
}

type interactionGenerator struct {
	random *rand.Rand
	issues []issues.Issue
}

// forUser picks 80% of the viewed issues among those that match the user's
// languages and categories and 20% at random, so that users with the same
// preferences end up with overlapping histories.
func (g interactionGenerator) forUser(userID bson.ObjectID, languages, categories []string, views int) []interactions.Interaction {
	var matching []int
	for index, issue := range g.issues {
		if slices.Contains(languages, issue.Language) && slices.Contains(categories, issue.Category) {
			matching = append(matching, index)
		}
	}
	g.random.Shuffle(len(matching), func(i, j int) { matching[i], matching[j] = matching[j], matching[i] })
	viewed := matching[:min(len(matching), views*8/10)]
	for len(viewed) < min(views, len(g.issues)) {
		if index := g.random.IntN(len(g.issues)); !slices.Contains(viewed, index) {
			viewed = append(viewed, index)
		}
	}

	var result []interactions.Interaction
	for _, index := range viewed {
		at := time.Now().UTC().Add(-time.Duration(g.random.IntN(30*24*60)) * time.Minute)
		record := func(interactionType string) {
			result = append(result, interactions.Interaction{
				ID:        bson.NewObjectID(),
				UserID:    userID,
				IssueID:   g.issues[index].ID,
				Type:      interactionType,
				CreatedAt: at,
			})
			at = at.Add(time.Minute)
		}
		record("view")
		if g.random.Float64() < 0.30 {
			record("bookmark")
		}
		if g.random.Float64() < 0.10 {
			record("working")
		}
		if g.random.Float64() < 0.05 {
			record("solved")
		}
	}
	return result
}
