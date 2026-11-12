package feed

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"issuematch/internal/issues"
)

var errRateLimited = errors.New("GitHub rate limit hit")

type githubRepo struct {
	FullName string `json:"full_name"`
	URL      string `json:"html_url"`
	Language string `json:"language"`
	Stars    int    `json:"stargazers_count"`
}

type githubIssue struct {
	Number      int       `json:"number"`
	Title       string    `json:"title"`
	Body        string    `json:"body"`
	CreatedAt   time.Time `json:"created_at"`
	PullRequest *struct{} `json:"pull_request"`
	Labels      []struct {
		Name string `json:"name"`
	} `json:"labels"`
}

// Run imports the open issues of every repository and returns how many it stored.
func Run(ctx context.Context, store issues.Store, repos []string, token string) (int, error) {
	imported := 0
	for _, repoName := range repos {
		repoName = strings.TrimSpace(repoName)

		var repo githubRepo
		var open []githubIssue
		err := fetch(ctx, token, "/repos/"+repoName, &repo)
		if err == nil {
			err = fetch(ctx, token, "/repos/"+repoName+"/issues?state=open&per_page=100", &open)
		}
		if errors.Is(err, errRateLimited) {
			log.Println("feed: GitHub rate limit hit, stopping")
			return imported, nil
		}
		if err != nil {
			return imported, err
		}

		for _, item := range open {
			// GitHub lists pull requests together with issues.
			if item.PullRequest != nil {
				continue
			}
			if err := store.UpsertImported(ctx, toIssue(repo, item)); err != nil {
				return imported, err
			}
			imported++
		}
	}
	return imported, nil
}

func fetch(ctx context.Context, token, path string, into any) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com"+path, nil)
	if err != nil {
		return err
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	rateLimited := response.StatusCode == http.StatusTooManyRequests ||
		(response.StatusCode == http.StatusForbidden && response.Header.Get("X-RateLimit-Remaining") == "0")
	if rateLimited {
		return errRateLimited
	}
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("GitHub answered %s for %s", response.Status, path)
	}
	return json.NewDecoder(response.Body).Decode(into)
}

func toIssue(repo githubRepo, item githubIssue) issues.Issue {
	labels := []string{}
	for _, label := range item.Labels {
		labels = append(labels, label.Name)
	}
	return issues.Issue{
		Title:       item.Title,
		Description: item.Body,
		Category:    categoryFromLabels(labels),
		Language:    repo.Language,
		Difficulty:  difficultyFromLabels(labels),
		Labels:      labels,
		Project:     issues.Project{Name: repo.FullName, URL: repo.URL, Stars: repo.Stars},
		Status:      "open",
		ExternalID:  fmt.Sprintf("github:%s#%d", repo.FullName, item.Number),
		CreatedAt:   item.CreatedAt,
	}
}

func categoryFromLabels(labels []string) string {
	switch {
	case anyContains(labels, "bug"):
		return "bug"
	case anyContains(labels, "doc"):
		return "documentation"
	case anyContains(labels, "refactor", "cleanup", "tech debt"):
		return "refactoring"
	}
	return "feature"
}

func difficultyFromLabels(labels []string) string {
	switch {
	case anyContains(labels, "good first issue", "easy", "beginner"):
		return "easy"
	case anyContains(labels, "hard", "complex"):
		return "hard"
	}
	return "medium"
}

func anyContains(labels []string, words ...string) bool {
	for _, label := range labels {
		for _, word := range words {
			if strings.Contains(strings.ToLower(label), word) {
				return true
			}
		}
	}
	return false
}
