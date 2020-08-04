package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/google/go-github/v32/github"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/oauth2"
)

// SearchPerPage is specified for pagination
const SearchPerPage = 100

// CollectIntervalSeconds specifies the interval for collecting data from GitHub
const CollectIntervalSeconds = 300

// LanguageLabels contain label names should be detected as languages
var LanguageLabels = map[string]bool{
	"ruby":       true,
	"javascript": true,
	"python":     true,
	"elixir":     true,
	"rust":       true,
	"java":       true,
	"go":         true,
	"elm":        true,
}

const namespace = "github_dependabot"

var (
	githubUsername string
	githubReponame string

	client *github.Client
	ctx    = context.Background()

	openPullRequestsGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "open_pull_requests",
			Help:      "Open Pull Requests sent by dependabot",
		},
		[]string{"username", "reponame", "library", "language", "from_version", "to_version", "directory", "security"},
	)
)

func init() {
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
	)
	httpClient := oauth2.NewClient(context.Background(), src)
	client = github.NewClient(httpClient)

}

func searchIssues() chan *github.Issue {
	githubUsername = os.Getenv("GITHUB_USERNAME")
	if githubUsername == "" {
		log.Fatal("GITHUB_USERNAME is not set")
	}

	githubReponame = os.Getenv("GITHUB_REPONAME")
	if githubReponame == "" {
		log.Fatal("GITHUB_REPONAME is not set")
	}

	issueChan := make(chan *github.Issue)

	go func() {
		var lastCreated *time.Time
		for {
			query := fmt.Sprintf("repo:%s/%s is:pr is:open label:dependencies", githubUsername, githubReponame)
			if lastCreated != nil {
				query = query + " created:>" + lastCreated.Format(time.RFC3339)
			}

			opts := &github.SearchOptions{
				Sort:  "created",
				Order: "asc",
				ListOptions: github.ListOptions{
					Page:    1,
					PerPage: SearchPerPage,
				},
			}
			result, _, err := client.Search.Issues(ctx, query, opts)
			if err != nil {
				log.Fatalf("Failed to fetch search result: %s", err)
			}

			for _, issue := range result.Issues {
				issueChan <- issue
				lastCreated = issue.CreatedAt
			}

			if len(result.Issues) < SearchPerPage {
				break
			}
		}

		close(issueChan)
	}()

	return issueChan
}

type dependabotPullRequest struct {
	Library     string
	Language    string
	FromVersion string
	ToVersion   string
	Directory   string
	Security    bool
}

var pattern = regexp.MustCompile(`^(\[Security\] )?(?:Bump|Update) (?P<library>\S+)(?: requirement)? from (?P<from_version>.+?) to (?P<to_version>.+?)(?: in /(?P<directory>.+?))?$`)

func parseDependabotPullRequest(issue *github.Issue) (*dependabotPullRequest, error) {
	matches := pattern.FindStringSubmatch(issue.GetTitle())

	if len(matches) > 0 {
		var security bool
		if matches[1] != "" {
			security = true
		}

		var language string
		for _, label := range issue.Labels {
			if _, ok := LanguageLabels[label.GetName()]; ok {
				language = label.GetName()
			}
		}

		return &dependabotPullRequest{
			Library:     matches[2],
			Language:    language,
			FromVersion: matches[3],
			ToVersion:   matches[4],
			Directory:   matches[5],
			Security:    security,
		}, nil
	}

	return nil, errors.New("Pattern not matched")
}

func main() {
	go collectTicker()

	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe("0.0.0.0:8000", nil))
}

func collect() {
	for issue := range searchIssues() {
		pr, err := parseDependabotPullRequest(issue)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to parse: #%d: %s\n", issue.GetNumber(), issue.GetTitle())
			continue
		}

		var security = "false"
		if pr.Security {
			security = "true"
		}

		labels := prometheus.Labels{
			"username":     githubUsername,
			"reponame":     githubReponame,
			"library":      pr.Library,
			"language":     pr.Language,
			"from_version": pr.FromVersion,
			"to_version":   pr.ToVersion,
			"directory":    pr.Directory,
			"security":     security,
		}
		openPullRequestsGauge.With(labels).Set(1)
	}
}

func collectTicker() {
	collected := false

	for {
		collect()

		if !collected {
			prometheus.MustRegister(
				openPullRequestsGauge,
			)
		}

		collected = true
		time.Sleep(CollectIntervalSeconds * time.Second)
	}
}
