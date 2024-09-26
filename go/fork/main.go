package main

import (
	"container/list"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/go-github/v54/github"
)

var (
	forkOwner = flag.String("fork-owner", "", "Orginal owner of the repo to be forked")

	newUserOwner = flag.String("new-user-owner", "", "New user owner of the forked repo")
	newOrgOwner  = flag.String("new-org-owner", "", "New organization owner of the forked repo")

	repoPrefix   = flag.String("repo-prefix", "", "GitHub repo name prefix")
	includeForks = flag.Bool("include-forks", false, "Include forked repos when searching for repos")

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

	err = forkAll(context.Background(), repos)
	if err != nil {
		log.Fatalf("failed to fork repositories: %v", err)
	}
}

func getRepositories() ([]github.Repository, error) {
	found := make([]github.Repository, 0)
	nextPage := 1
	lastCount := 0
	for nextPage != 0 {
		log.Println("Checking page", nextPage, "...")
		repos, resp, err := c.Repositories.List(context.Background(), *forkOwner, &github.RepositoryListOptions{
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
			if repo.Fork != nil && *repo.Fork && !*includeForks {
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

func forkAll(ctx context.Context, repos []github.Repository) error {
	log.Println("Forking", len(repos), "repositories...")

	l := list.New()
	for _, r := range repos {
		l.PushBack(*r.Name)
	}

	t := newTimer(time.Second * 30)

	e := l.Front()
	for e != nil {
		repo := fmt.Sprint(e.Value)
		if t.triesExausted() {
			return fmt.Errorf("exhausted all tries while forking %s", repo)
		}

		ok, err := forkOne(ctx, repo)
		if err != nil {
			return err
		}

		if !ok {
			log.Printf("Rate limited or failed, will retry...")
			t.wait()
			continue
		}

		e = e.Next()
		t.reset()
	}

	return nil
}

func forkOne(ctx context.Context, repo string) (bool, error) {
	if *forkOwner != "" || *newOrgOwner != "" {
		check := forkOwner
		if *newOrgOwner != "" {
			check = newOrgOwner
		}
		ok, err := repositoryExists(ctx, *check, repo)
		if err != nil {
			return false, err
		}

		if ok {
			log.Println("Skipping", repo, "because it already exists in", forkOwner)
			return true, nil
		}
	}

	log.Println("Forking", repo, "repository...")
	opts := &github.RepositoryCreateForkOptions{}
	if *newOrgOwner != "" {
		opts.Organization = *newOrgOwner
	}

	_, _, err := c.Repositories.CreateFork(ctx, *forkOwner, repo, opts)
	if err != nil {
		if strings.Contains(err.Error(), "403 was submitted too quickly") || strings.Contains(err.Error(), "500") {
			return false, nil
		}
		if strings.Contains(err.Error(), "try again later") {
			return true, nil
		}

		return false, err
	}

	return true, nil
}

func repositoryExists(ctx context.Context, owner string, repo string) (bool, error) {
	_, _, err := c.Repositories.Get(ctx, owner, repo)
	if err != nil {
		if strings.Contains(err.Error(), "404 Not Found") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

type timer struct {
	tries    int
	maxTries int

	waitTime time.Duration
}

func newTimer(waitTime time.Duration) *timer {
	return &timer{
		maxTries: 10,
		waitTime: waitTime,
	}
}

func (t *timer) triesExausted() bool {
	return t.tries == t.maxTries
}

func (t *timer) wait() {
	log.Printf("Waiting %s ...", t.waitTime*time.Duration(t.tries))

	time.Sleep(t.waitTime * time.Duration(t.tries))
	t.tries++
}

func (t *timer) reset() {
	t.tries = 1
}
