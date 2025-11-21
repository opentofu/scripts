# Display GitHub Actions workflow status

This script is used to display the status of GitHub Actions workflow runs across multiple repositories matching a prefix. 
It shows counts of queued, in-progress, successful, and failed runs with optional filtering.

## Required environment variables
### `GITHUB_PAT`
A token that is used to interact with the GitHub API. 
The permissions that it needs to have:
* "Metadata" repository permissions (read)
* "Actions" repository permissions (read) - for reading workflow runs

## Flags

```shell
  -owner string
        Owner of the repositories (required)
  -repo-prefix string
        GitHub repo name prefix to filter repositories (required)
  -workflow string
        Workflow filename to show status of, e.g., fork_sync.yml (required)
  -failed-only
        Only show failed workflows (default false)
  -verbose
        Show verbose output including timestamps and duration (default false)
```
