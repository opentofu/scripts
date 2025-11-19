# Configure the deployment environment

This script is used to configure the environment for the given repository name or for any repository named with a prefix indicated by the `-repo-prefix` flag.

## Required environment variables
### `GITHUB_PAT`
A token that is used to interact with the GitHub API. 
The permissions that it needs to have:
* "Metadata" repository permissions (read)
* "Administration" repository permissions (write)
* "Environments" repository permissions (read)
* "Environments" repository permissions (write)

### `SECRET_VALUE`
This represents the value of the secret that wants to be created or updated in the deployment environment of the targeted repository.
You need to use this to provide the value of the secret while using the flag `-secret-name` to provide the secret name.

## Flags

```shell
  -env-name string
        GitHub env name to be created or updated (default "gpg")
  -org string
        GitHub organization name where the targeted repository(/repositories) can be found
  -repo string
        GitHub repository name (exact match)
  -repo-prefix string
        GitHub repository name prefix
  -secret-name string
        GitHub organization secret name (default "GPG_PRIVATE_KEY") 
```