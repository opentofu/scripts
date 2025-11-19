# Trigger GitHub Actions workflows

This script is used to trigger GitHub Actions workflow dispatch events across multiple repositories matching a prefix. 
It optionally filters to specific repositories.

## Required environment variables
### `GITHUB_PAT`
A token that is used to interact with the GitHub API. 
The permissions that it needs to have:
* "Metadata" repository permissions (read)
* "Actions" repository permissions (write) - for triggering workflows

## Flags

```shell
  -owner string
        Owner of the repositories that the workflow will be triggered on (required)
  -repo-prefix string
        GitHub repo name prefix to filter repositories (required)
  -repo-filter string
        Comma-separated list of specific repository names to trigger, e.g., repo1,repo2,repo3. This is used to be sure that from the filtered ones it triggers workflow only of some specific ones. If this is empty, all repositories returned from listing with `-repo-prefix` will have the workflow triggered. 
  -workflow string
        Workflow filename to trigger, e.g., sync.yml (required)
```
