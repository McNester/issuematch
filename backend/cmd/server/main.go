package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"issuematch/internal/auth"
	"issuematch/internal/feed"
	"issuematch/internal/interactions"
	"issuematch/internal/issues"
	"issuematch/internal/platform"
	"issuematch/internal/users"
)

func main() {
	config, err := platform.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}
	db, err := platform.ConnectMongo(config.MongoURI)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	err = users.NewStore(db).EnsureIndexes(ctx)
	if err == nil {
		err = issues.NewStore(db).EnsureIndexes(ctx)
	}
	if err == nil {
		err = interactions.NewStore(db).EnsureIndexes(ctx)
	}
	if err != nil {
		log.Fatal(err)
	}

	go every(config.FeedInterval, func() {
		imported, err := feed.Run(ctx, issues.NewStore(db), config.FeedRepos, config.GitHubToken)
		log.Printf("feed: imported %d issues, error: %v", imported, err)
	})

	router := chi.NewRouter()
	router.Route("/api", func(r chi.Router) {
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			platform.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
		})
		auth.Routes(r, db, config.JWTSecret)
		users.Routes(r, db, config.JWTSecret)
		issues.Routes(r, db, config.JWTSecret)
		interactions.Routes(r, db, config.JWTSecret)
	})

	log.Println("listening on port", config.Port)
	log.Fatal(http.ListenAndServe(":"+config.Port, router))
}

// every runs the job after each interval. An interval of 0 turns it off.
func every(interval time.Duration, job func()) {
	if interval <= 0 {
		return
	}
	for range time.Tick(interval) {
		job()
	}
}
