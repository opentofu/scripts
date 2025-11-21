# Re-sign releases
This script is used to regenerate (re-sign) the SHA256SUM.sig file for each release of the given repository based on the SHA256SUM file. 
It downloads checksum files, signs them locally, and uploads new signatures.

This script relies on having the `gpg` binary installed and configured with the GPG private key, which in the case of this
script, is done by the [sign.yml](../../.github/workflows/sign.yml) workflow that calls this script.

## Required environment variables
### `GITHUB_PAT`
A token that is used to interact with the GitHub API. 
The permissions that it needs to have:
* "Metadata" repository permissions (read)
* "Contents" repository permissions (read) - for listing releases and downloading assets
* "Contents" repository permissions (write) - for deleting old signatures and uploading new ones

## Flags

```shell
  -owner string
        The GitHub organisation (or user) where the repository can be found (required)
  -repo string
        The GitHub repository name for which the releases needs to be re-signed (required)
  -fingerprint string
        GPG fingerprint to use for signing (required)
```
