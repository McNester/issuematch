package platform

import (
	"errors"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	MongoURI          string
	JWTSecret         string
	Port              string
	FeedRepos         []string
	FeedInterval      time.Duration
	GitHubToken       string
	RecomputeInterval time.Duration
}

var defaultFeedRepos = "golang/go,kubernetes/kubernetes,spring-projects/spring-boot,elastic/elasticsearch," +
	"facebook/react,nodejs/node,python/cpython,pallets/flask,rust-lang/rust,tokio-rs/tokio"

func LoadConfig() (Config, error) {
	// A missing .env file is fine: the settings can come from the environment.
	_ = godotenv.Load()

	feedInterval, err := time.ParseDuration(env("FEED_INTERVAL", "6h"))
	if err != nil {
		return Config{}, errors.New("FEED_INTERVAL is not a duration")
	}
	recomputeInterval, err := time.ParseDuration(env("RECOMPUTE_INTERVAL", "10m"))
	if err != nil {
		return Config{}, errors.New("RECOMPUTE_INTERVAL is not a duration")
	}

	config := Config{
		MongoURI:          env("MONGO_URI", "mongodb://localhost:27017"),
		JWTSecret:         os.Getenv("JWT_SECRET"),
		Port:              env("PORT", "8080"),
		FeedRepos:         strings.Split(env("FEED_REPOS", defaultFeedRepos), ","),
		FeedInterval:      feedInterval,
		GitHubToken:       os.Getenv("GITHUB_TOKEN"),
		RecomputeInterval: recomputeInterval,
	}
	if config.JWTSecret == "" {
		return Config{}, errors.New("JWT_SECRET is required")
	}
	return config, nil
}

func env(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
