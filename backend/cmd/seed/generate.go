package main

import (
	"fmt"
	"math/rand/v2"
	"strings"
	"time"

	"issuematch/internal/issues"
)

var languages = []string{"Go", "Java", "JavaScript", "Python", "Rust"}

var projectOwners = []string{"northwind", "papertrail", "quarry", "tidewater", "lumen", "harbor", "cobalt", "meridian"}

var projectNames = []string{
	"httpclient", "docs-site", "quarry-db", "jobs", "reports", "gateway", "cache", "scheduler",
	"auth-kit", "metrics", "queue", "cli", "storage", "search", "config", "mailer",
}

var components = []string{
	"connection pool", "query planner", "scheduler", "request router", "config loader", "cache layer",
	"retry policy", "log formatter", "session store", "rate limiter", "migration runner", "file watcher",
	"JSON encoder", "auth middleware", "worker queue", "metrics exporter", "CLI parser", "index builder",
	"health check", "upload handler", "template renderer", "websocket hub", "backup job", "search endpoint",
}

var failures = []string{
	"leaks memory", "returns stale data", "panics", "deadlocks", "drops requests", "ignores the timeout",
	"reports the wrong status", "loses the last item", "uses too much memory", "fails silently",
	"retries forever", "corrupts the output", "blocks the event loop", "skips validation",
}

var conditions = []string{
	"when a request times out", "under heavy load", "after a restart", "on empty input",
	"when the network drops", "with more than 1,000 items", "on Windows", "when two writers race",
	"if the config file is missing", "during a rolling deploy", "with unicode input", "after the cache expires",
}

var capabilities = []string{
	"CSV export", "dark mode", "pagination", "structured logging", "a dry-run flag", "gzip compression",
	"OpenTelemetry tracing", "a health endpoint", "bulk import", "retry with backoff", "webhook support",
	"a progress bar", "config reload without restart", "per-user rate limits", "memory usage metrics",
}

var topics = []string{
	"the retry policy options", "the configuration file format", "how to run the tests", "the public API",
	"error codes", "the release process", "the plugin interface", "memory tuning", "the upgrade path",
	"authentication setup", "the caching rules", "deployment with Docker",
}

var cleanups = []string{
	"Split %s into smaller packages", "Remove duplicated code in the %s", "Replace global state in the %s",
	"Simplify error handling in the %s", "Extract an interface from the %s", "Rename unclear functions in the %s",
	"Move the %s out of the main module", "Delete dead code from the %s",
}

type project struct {
	issues.Project
	language string
}

func generateProjects(random *rand.Rand) []project {
	var projects []project
	for _, owner := range projectOwners {
		for _, name := range projectNames {
			projects = append(projects, project{
				Project: issues.Project{
					Name:  owner + "/" + name,
					URL:   "https://github.com/" + owner + "/" + name,
					Stars: 50 + random.IntN(20000),
				},
				language: pick(random, languages),
			})
		}
	}
	return projects
}

func generateIssue(random *rand.Rand, projects []project) issues.Issue {
	project := pick(random, projects)
	category := pick(random, issues.Categories)
	difficulty := pick(random, issues.Difficulties)
	component := pick(random, components)

	var title, description string
	switch category {
	case "bug":
		failure, condition := pick(random, failures), pick(random, conditions)
		title = fmt.Sprintf("%s %s %s", capitalize(component), failure, condition)
		description = fmt.Sprintf("The %s %s %s. It happens on the latest release.\n\n"+
			"Steps to reproduce: start the service with the default settings and repeat the action a few times.\n\n"+
			"Expected behavior: the %s keeps working and reports a clear error.", component, failure, condition, component)
	case "feature":
		capability := pick(random, capabilities)
		title = fmt.Sprintf("Add %s to the %s", capability, component)
		description = fmt.Sprintf("Users have asked for %s in the %s. Today they work around it with scripts.\n\n"+
			"A first version can be small. It needs tests and a short section in the docs.", capability, component)
	case "documentation":
		topic := pick(random, topics)
		title = fmt.Sprintf("Document %s for the %s", topic, component)
		description = fmt.Sprintf("The docs do not explain %s for the %s. New contributors ask about it often.\n\n"+
			"Add a page with one working example and link it from the README.", topic, component)
	case "refactoring":
		title = fmt.Sprintf(pick(random, cleanups), component)
		description = fmt.Sprintf("The %s has grown and is hard to change. Tests for it are slow and brittle.\n\n"+
			"The behavior must stay the same. Do the change in small commits.", component)
	}

	labels := []string{}
	if difficulty == "easy" && random.IntN(2) == 0 {
		labels = append(labels, "good first issue")
	}
	if random.IntN(4) == 0 {
		labels = append(labels, "help wanted")
	}

	return issues.Issue{
		Title:       title,
		Description: description,
		Category:    category,
		Language:    project.language,
		Difficulty:  difficulty,
		Labels:      labels,
		Project:     project.Project,
		Status:      "open",
		CreatedAt:   time.Now().UTC().Add(-time.Duration(random.IntN(90*24)) * time.Hour),
	}
}

func pick[T any](random *rand.Rand, from []T) T {
	return from[random.IntN(len(from))]
}

func capitalize(text string) string {
	return strings.ToUpper(text[:1]) + text[1:]
}
