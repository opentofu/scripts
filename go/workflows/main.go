package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/google/go-github/v54/github"
)

var (
	owner      = flag.String("owner", "", "Owner of the repo to be forked")
	repoPrefix = flag.String("repo-prefix", "", "GitHub repo name prefix")
	repoFilter = flag.String("repo-filter", "", "Repository list to filter repositories, example: repo1,repo2,repo3")
	workflow   = flag.String("workflow", "", "Workflow to trigger, example: sync.yml")

	pat string

	c *github.Client
)

func main() {
	pat = os.Getenv("GITHUB_PAT")
	if pat == "" {
		panic("GITHUB_PAT environment variable not set")
	}

	flag.Parse()

	if repoPrefix == nil || *repoPrefix == "" {
		panic("repo-prefix flag not set")
	}

	c = github.NewTokenClient(context.Background(), pat)

	repos, err := getRepositories()
	if err != nil {
		log.Fatalf("failed to get repositories: %v", err)
	}

	if repoFilter != nil && *repoFilter != "" {
		var reposToTrigger []github.Repository
		repoList := strings.Split(*repoFilter, ",")

		for _, r := range repos {
			for _, repo := range repoList {
				if *r.Name != repo {
					continue
				}

				reposToTrigger = append(reposToTrigger, r)
				break
			}
		}

		repos = reposToTrigger
	}

	ctx := context.Background()
	for _, r := range repos {
		fmt.Println("Triggering workflow in", *r.Name, "...")

		_, err := c.Actions.CreateWorkflowDispatchEventByFileName(ctx, *owner, r.GetName(), *workflow, github.CreateWorkflowDispatchEventRequest{
			Ref: *r.DefaultBranch,
		})
		if err != nil {
			log.Fatalf("failed to trigger workflow: %v", err)
		}
	}
}

func getRepositories() ([]github.Repository, error) {
	found := make([]github.Repository, 0)
	nextPage := 1
	lastCount := 0
	for nextPage != 0 {
		log.Println("Checking page", nextPage, "...")
		repos, resp, err := c.Repositories.List(context.Background(), *owner, &github.RepositoryListOptions{
			ListOptions: github.ListOptions{
				Page:    nextPage,
				PerPage: 100,
			},
			Sort:       "full_name",
			Visibility: "public",
			Direction:  "desc",
		})
		if err != nil {
			return nil, err
		}

		for _, repo := range repos {
			if repo.Name == nil {
				continue
			}

			if strings.HasPrefix(*repo.Name, *repoPrefix) {
				found = append(found, *repo)
			}
		}
		log.Println("Found", len(found), "so far...")
		nextPage = resp.NextPage

		// We sorted by name, so if we have the same count twice, we're done.
		if lastCount != 0 && len(found) == lastCount {
			break
		}

		lastCount = len(found)
	}

	return found, nil
}
