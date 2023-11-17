# OpenTofu Scripts

This repository contains a collection of scripts and github actions which are used by OpenTofu to manage and maintain repositories and their contents. The scripts are made to be executed using GitHub Actions.

## Contents

The repository contains several directories and files:

- `.github/workflows`: Contains workflow files for GitHub Actions. These workflows automate various tasks and are either inherited by other repositores or executed directly from this repository.

- `go`: Contains Go scripts.

- `sh`: Contains Shell scripts.

The scripts in this repository are designed to work in conjunction with GitHub Actions, an automation feature provided by GitHub. For more information on how to use GitHub Actions, you can refer to the [GitHub Actions Cheat Sheet](https://resources.github.com/actions/github-actions-cheat/).

## How to 

### Add a New Provider Repository

When you add a provider repository, it needs to be configured to automatically sync with upstream and produce new releases. To achieve this, execute the [Update repository environments](https://github.com/opentofu/scripts/actions/workflows/env.yml) action. This action configures the repository by creating an environment and setting the GPG key used for signing releases. Additionally, include [our custom GitHub workflows](https://github.com/opentofu/terraform-provider-waypoint/commit/cda4c700d64bf5c8cefed5cce0a59aea43462baf) in the new provider repository. You can easily accomplish this by using the [reset_repos.sh](https://github.com/opentofu/scripts/blob/main/sh/reset_repos.sh) script.

### Resign All Releases 

Each provider has a GitHub action called [Artifacts Resign](https://github.com/opentofu/terraform-provider-waypoint/actions/workflows/resign.yml), which can be executed to resign all releases in that repository. The script executed during the resigning process for a provider can be found [here](https://github.com/opentofu/scripts/blob/main/go/sign/main.go).

Note: Keep in mind github rate limits when executing this action. All repositories use the same PAT meaning that if the action is exected on all providers at the same time, the request limit of 5000 will be exceeded.

### Generate a New Private GPG Key

**NOTE: EXECUTING THIS WILL ERASE THE CURRENT KEY**

Generate a new key using the [Run GPG script](https://github.com/opentofu/scripts/actions/workflows/gpg.yml) action. The script accepts inputs for testing, but by default, you should provide:
- Organization: `opentofu`
- Repo: `scripts`
- Secret Name: `GPG_PRIVATE_KEY` (Provide a different value if you do not wish to erase the current key)

After generating a new key, propagate it to all providers by calling [Update repository environments](https://github.com/opentofu/scripts/actions/workflows/env.yml) and using `terraform-provider-` as the repository prefix to match.

### Check if All Tags Have Releases

Easily check which tags have releases and which do not by using the [Check releases](https://github.com/opentofu/scripts/blob/main/.github/workflows/check_releases.yml) GitHub action.

## Note

This repository does not accept contributions. It's a collection of scripts used to manage OpenTofu repositories.

Please note that these scripts are specifically tailored for the needs of OpenTofu and may not be suitable for other use cases. 
