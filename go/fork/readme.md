# Fork GitHub repositories

This script is used to fork GitHub repositories from one owner to another (user or organization) based on a repository name prefix. It includes rate limiting and retry logic.

## Required environment variables
### `GITHUB_PAT`
A token that is used to interact with the GitHub API. 
The permissions that it needs to have:
* "Metadata" repository permissions (read)
* "Administration" repository permissions (write)
* "Contents" repository permissions (read)

## Flags

```shell
  -fork-owner string
        The GitHub organisation/user where the repository should be forked from
  -new-org-owner string
        The GitHub organisation/user where the repository should be forked to
  -include-forks (default: false)
        When listing the repositories to be forked, by setting this to `true` will include also the repositories that are forked from another one
  -repo-prefix string
        Value to filter what repositories should be processed. This should be just a regular string and not a regex
```
