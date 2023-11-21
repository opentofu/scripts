package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/google/go-github/v54/github"
)

var (
	owner      = flag.String("owner", "", "Owner of the repo to be forked")
	repoPrefix = flag.String("repo-prefix", "", "GitHub repo name prefix")

	workflow = flag.String("workflow", "", "Workflow to show the status of, example: sync.yml")

	failedOnly = flag.Bool("failed-only", false, "Only show failed workflows")
	verbose    = flag.Bool("verbose", false, "Show verbose output")

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
	var (
		blue   = color.New(color.FgBlue).SprintFunc()
		yellow = color.New(color.FgYellow).SprintFunc()
		green  = color.New(color.FgGreen).SprintFunc()
		red    = color.New(color.FgRed).SprintFunc()
	)

	fmt.Println("\nChecking workflows...")
	fmt.Println("This may take a while...")

	status, err := getWorkflowStatus(context.Background(), repos)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\nname: total/queued/in progress/success/failure")

	var summary workflowStatus
	for name, workflowStatus := range status {
		if workflowStatus.FinishedAt.IsZero() {
			workflowStatus.FinishedAt = time.Now()
		}
		if !*failedOnly || workflowStatus.Failure > 0 {
			fmt.Printf("%s: %d/%s/%s/%s/%s\n", name, workflowStatus.Total, blue(workflowStatus.Queued), yellow(workflowStatus.InProgress), green(workflowStatus.Success), red(workflowStatus.Failure))
			if *verbose {
				if *failedOnly {
					for _, u := range workflowStatus.FailureURLs {
						fmt.Println(u)
					}
				} else {
					var dur time.Duration
					if !workflowStatus.StartedAt.IsZero() {
						dur = workflowStatus.FinishedAt.Sub(workflowStatus.StartedAt)
					}
					fmt.Printf("started: %s, duration: %s\n", workflowStatus.StartedAt.In(time.Local).Format(time.RFC3339), dur)
				}
			}
		}

		if workflowStatus.Total > 0 {
			summary.Total += workflowStatus.Total
			summary.Queued += workflowStatus.Queued
			summary.InProgress += workflowStatus.InProgress
			summary.Success += workflowStatus.Success
			summary.Failure += workflowStatus.Failure
			if summary.StartedAt.IsZero() || (!workflowStatus.StartedAt.IsZero() && workflowStatus.StartedAt.Before(summary.StartedAt)) {
				summary.StartedAt = workflowStatus.StartedAt
			}
			if workflowStatus.FinishedAt.After(summary.FinishedAt) {
				summary.FinishedAt = workflowStatus.FinishedAt
			}
		}
	}
	if !*failedOnly {
		fmt.Println("name: total/queued/in progress/success/failure")
		fmt.Printf("summary: %d/%s/%s/%s/%s\n", summary.Total, blue(summary.Queued), yellow(summary.InProgress), green(summary.Success), red(summary.Failure))
		if *verbose {
			fmt.Printf("started: %s, duration: %s\n", summary.StartedAt.In(time.Local).Format(time.RFC3339), summary.FinishedAt.Sub(summary.StartedAt))
		}
	}
}

type workflowStatus struct {
	Total      int
	Queued     int
	InProgress int
	Success    int
	Failure    int

	FailureURLs []string
	StartedAt   time.Time
	FinishedAt  time.Time
}

func getWorkflowStatus(ctx context.Context, repos []github.Repository) (map[string]workflowStatus, error) {
	status := map[string]workflowStatus{}
	for i := range repos {
		fmt.Println("Checking", repos[i].GetName(), "...")
		var workflowRuns []*github.WorkflowRun
		page := 1
		for {
			runs, _, err := c.Actions.ListWorkflowRunsByFileName(ctx, *repos[i].Owner.Login, *repos[i].Name, *workflow, &github.ListWorkflowRunsOptions{
				ListOptions: github.ListOptions{
					Page:    page,
					PerPage: 99,
				},
			})
			if err != nil {
				return nil, err
			}
			//log.Printf("%s: total=%d,page=%d,count=%d,got=%d", repos[i].GetName(), runs.GetTotalCount(), page, len(runs.WorkflowRuns), len(workflowRuns))
			workflowRuns = append(workflowRuns, runs.WorkflowRuns...)
			if len(workflowRuns) == runs.GetTotalCount() {
				break
			}
			page++
		}
		workflowStatus := workflowStatus{Total: len(workflowRuns)}
		for _, run := range workflowRuns {
			//log.Printf("%s: %s,%s", repos[i].GetName(), run.GetStatus(), run.GetConclusion())
			switch run.GetStatus() {
			case "queued":
				workflowStatus.Queued++
			case "completed":
				switch run.GetConclusion() {
				case "success":
					workflowStatus.Success++
				case "failure":
					workflowStatus.Failure++
					workflowStatus.FailureURLs = append(workflowStatus.FailureURLs, run.GetHTMLURL())
				}
				if run.GetUpdatedAt().After(workflowStatus.FinishedAt) {
					workflowStatus.FinishedAt = run.GetUpdatedAt().Time
				}
			default:
				workflowStatus.InProgress++
			}
			if workflowStatus.StartedAt.IsZero() || run.GetCreatedAt().Before(workflowStatus.StartedAt) {
				workflowStatus.StartedAt = run.GetCreatedAt().Time
			}
		}
		status[repos[i].GetName()] = workflowStatus
	}

	return status, nil
}

func getRepositories() ([]github.Repository, error) {
	found := make([]github.Repository, 0)
	nextPage := 1
	lastCount := 0
	for nextPage != 0 {
		fmt.Println("Checking page", nextPage, "...")
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
		fmt.Println("Found", len(found), "so far...")
		nextPage = resp.NextPage

		// We sorted by name, so if we have the same count twice, we're done.
		if lastCount != 0 && len(found) == lastCount {
			break
		}

		lastCount = len(found)
	}

	return found, nil
}
