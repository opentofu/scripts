# OpenTofu Scripts

This repository contains a collection of scripts and github actions which are used by OpenTofu to manage and maintain repositories and their contents. 
The scripts are made to be executed using GitHub Actions.

## You opened this repository because you want to...
* Blacklist a specific terraform-provider-* repository tag because the build is failing due to a reason outside of our control, check the last step from [`sync.yml`](./.github/workflows/sync.yml).
* Refresh the terraform-provider-* `.github/workflows` content, you might need to run [`./sh/reset_repos.sh`](./sh/reset_repos.sh) (details in [sh/readme.md](./sh/readme.md)).
* Disable newly introduced workflows from upstream on a terraform-provider-* repository that we are not interested in. You could run [`./sh/disable_unwanted_workflows.sh`](./sh/disable_unwanted_workflows.sh) (details in [sh/readme.md](./sh/readme.md)).
* There is a new provider that you want to fork on our side: 
  * run the [`fork.yml`](./.github/workflows/fork.yml) workflow for that specific upstream repo to fork it on the OpenTofu's organisation.
  * run the [`env.yml`](./.github/workflows/env.yml) workflow for the newly forked repo to setup the build environment with the GPG private key.
  * run the [`secret.yml`](./.github/workflows/secret.yml) workflow for the newly forked repo to configure other secrets than the GPG related information.
  * run the [`reset_repos.sh`](./sh/reset_repos.sh) script to configure the newly provider with the OpenTofu's specific workflows.
  * run the [`disable_unwanted_workflows.sh`](./sh/disable_unwanted_workflows.sh) to disable all the other workflows that are not needed by OpenTofu from that repo.


## Contents
The repository contains several directories and files:
- `.github/workflows`: Contains workflow files for GitHub Actions. These workflows automate various tasks and are either inherited by other repositories or executed directly from this repository. Details below.
- `go`: Contains various go scripts that are executed from different GitHub workflows.
  - [fork](./go/fork/readme.md)
  - [gpg](./go/gpg/readme.md)
  - [workflows](./go/workflows/readme.md)
  - [status](./go/status/readme.md)
  - [env](./go/env/readme.md)
  - [secret](./go/secret/readme.md)
  - [sign](./go/sign/readme.md)
- `sh`: Contains various Shell scripts for different operations. Details can be found in [sh/readme.md](./sh/readme.md).

The scripts in this repository are designed to work in conjunction with GitHub Actions, an automation feature provided by GitHub. For more information on how to use GitHub Actions, you can refer to the [GitHub Actions Cheat Sheet](https://resources.github.com/actions/github-actions-cheat/).

## How to

### Resign All Releases of a provider
Each provider has a GitHub action called [Artifacts Resign](https://github.com/opentofu/terraform-provider-waypoint/actions/workflows/resign.yml), 
which can be executed to resign all releases in that repository. 
The script executed during the resigning process for a provider can be found [here](https://github.com/opentofu/scripts/blob/main/go/sign/main.go).

> [!NOTE]
> Note: Keep in mind github rate limits when executing this action. 
> All repositories use the same PAT meaning that if the action is exected on all providers at the same time, the request limit of 5000 will be exceeded.

### Generate a New Private GPG Key
> [!WARNING]
> EXECUTING THIS WILL ERASE THE CURRENT KEY

Generate a new key using the [Run GPG script](https://github.com/opentofu/scripts/actions/workflows/gpg.yml) action. 
The script accepts inputs for testing, but by default, you should provide:
- Organization: `opentofu`
- Repo: `scripts`
- Secret Name: `GPG_PRIVATE_KEY` (Provide a different value if you do not wish to erase the current key)

After generating a new key, propagate it to all providers by calling 
[Update repository environments](https://github.com/opentofu/scripts/actions/workflows/env.yml) and using `terraform-provider-` as the repository prefix to match.

### Check if All Tags Have Releases
Easily check which tags have releases and which do not by using the 
[Check releases](https://github.com/opentofu/scripts/blob/main/.github/workflows/check_releases.yml) GitHub action.

## Details on the workflows from this repo
### [`secret.yml`](./.github/workflows/secret.yml)
Use this when you want to update or add a secret to a repository that you don't have access otherwise. 
### [`fork.yml`](./.github/workflows/fork.yml)
This can fork a repository from another organisation to the OpenTofu organisation.
### [`status.yml`](./.github/workflows/status.yml)
Prints the statuses of the workflows of specified repositories.
### [`check_releases.yml`](./.github/workflows/check_releases.yml)
Checks all the terraform-provider-* repos from the OpenTofu organisation and reports what tags have no release associated.
### [`env.yml`](./.github/workflows/env.yml)
Used to create a deployment environment in a newly forked repository that will contain the GPG key necessary for releasing new versions of that provider.
### [`sign.yml`](./.github/workflows/sign.yml)
Necessary when required to regenerate the `SHA256SUM.sig` of all the releases of a provider. Generally this should be used if *ever* the GPG key is changed for that provider repository. 
### [`gpg.yml`](./.github/workflows/gpg.yml)
This can generate new GPG keys for the provider repositories to use those for signing new releases.
> [!WARNING]
> From the looks of it, in the [gpg script](./go/gpg/main.go), the lifespan of such keys is hardcoded to 3 years.
> This script might need to be executed for all the repositories when the keys will expire.
### [`workflows.yml`](./.github/workflows/workflows.yml)
This can trigger a workflow in a terraform-provider-* repository. Generally, we don't have access to trigger the workflows manually in those repositories so we can use this one to do so.
### [`release.yml`](./.github/workflows/release.yml) (template)
> [!NOTE]
> This workflow is not meant to be executed directly from this repository but from the terraform-provider-* repositories.
> If such a repository does not have this workflow, consider the [`reset_repos.sh`](./sh/reset_repos.sh) script to add that.

A template workflow that is used by each provider via the [`artifact-release.yml`](./sh/workflows/artifact-release.yml) workflow
that creates a new release for the given tag of the repository that it's running against.
### [`sync.yml`](./.github/workflows/sync.yml) (template)
> [!NOTE]
> This workflow is not meant to be executed directly from this repository but from the terraform-provider-* repositories.
> If such a repository does not have this workflow, consider the [`reset_repos.sh`](./sh/reset_repos.sh) script to add that.

A template workflow that is used by each provider via the [`fork_sync.yml`](./sh/workflows/fork_sync.yml) workflow
that checks the upstream repository for new tags and if any missing on the OpenTofu's fork, it will create it and trigger
the [`artifact-release.yml`](./sh/workflows/artifact-release.yml) workflow for that tag.
### [`trigger.yml`](./.github/workflows/trigger.yml) (template)
> [!NOTE]
> This workflow is not meant to be executed directly from this repository but from the terraform-provider-* repositories.
> If such a repository does not have this workflow, consider the [`reset_repos.sh`](./sh/reset_repos.sh) script to add that.

A template workflow that is used by each provider via the [`artifact-release-trigger.yml`](./sh/workflows/artifact-release-trigger.yml) workflow
that is just another way to trigger the [`artifact-release.yml`](./sh/workflows/artifact-release.yml) workflow for a given tag.

## Note
This repository does not accept contributions. It's a collection of scripts used to manage OpenTofu repositories.

Please note that these scripts are specifically tailored for the needs of OpenTofu and may not be suitable for other use cases. 
