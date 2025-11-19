# Generate GPG key pair

This script is used to generate a new GPG key pair (4096-bit RSA) and store the private key as an encrypted GitHub Actions secret in a repository.

## Required environment variables
### `GITHUB_PAT`
A token that is used to interact with the GitHub API. 
The permissions that it needs to have:
* "Metadata" repository permissions (read)
* "Secrets" repository permissions (read)
* "Secrets" repository permissions (write) - for creating/updating repository secrets

## Flags

```shell
  -org string
        GitHub organization (or user) that owns the repository for which the secret needs to be configured
  -repo string
        GitHub repository on which the secret will be configured (default "scripts")
  -secret string
        Name that will be used to store the secret in the repository (default "GPG_PRIVATE_KEY")
  -gpg-comment string
        GPG comment (default "This is the key used to sign opentofu providers")
  -gpg-email string
        GPG comment (default "your.email@example.com")
  -gpg-name string
        GPG name (default "opentofu")
```
