# Set GitHub Actions secret

This script is used to set a GitHub Actions secret across multiple repositories that match a given prefix.

## Required environment variables
### `GITHUB_PAT`
A token that is used to interact with the GitHub API. 
The permissions that it needs to have:
* "Metadata" repository permissions (read)
* "Secrets" repository permissions (read)
* "Secrets" repository permissions (write) - for managing repository secrets

### `SECRET_VALUE`
The value to set for the secret in the targeted repositories.
You need to use this env var to provide the value of the secret while using the flag `-secret-key` to provide the secret name.

## Flags

```shell
  -owner string
        Owner of the repo to set the secret in (required)
  -repo-prefix string
        GitHub repo name prefix to filter repositories that the secret will be configured on (required)
  -secret-key string
        Name for the secret key (required)
```
