# Shell Scripts

This directory, contains a collection of shell scripts that perform various against provider repositories.

## Contents

This directory contains a list of workflows that are meant to be added (by using the existing scripts) to any `terraform-provider-*` repo.
These can be found in [`./workflows`](./workflows):
* [`artifact-release.yml`](./workflows/artifact-release.yml): this uses goreleaser to generate new releases for the provider repo that is executed on.
* [`artifact-release-trigger.yml`](./workflows/artifact-release-trigger.yml): allows manual triggering of the [`artifact-release.yml`](./workflows/artifact-release.yml) from the terraform-provider repo.
* [`fork_sync.yml`](./workflows/fork_sync.yml): checks the OpenTofu's repo against the upstream fork and triggers [`artifact-release.yml`](./workflows/artifact-release.yml) for any missing releases on the repo.
* [`resign.yml`](./workflows/resign.yml): signs again the artifacts of the releases of the terraform-provider-* repo.

The critical one is `fork_sync.yml` because that keeps the forked repos up to date with the upstream.

### [`check_releases.sh`](./check_releases.sh)
Can be used to find out what tags of a provider repository does not have a release created.
For more details on the input arguments, check the script.

### [`reset_repos.sh`](./reset_repos.sh)
For each repo in a given file, it's cloning it, setting the upstream repo (by replacing `opentofu` with `hashicorp` namespace) and 
resets the given repo main branch to the upstream one. After this has been done, it adds the workflow files above mentioned
into `.github/workflows` and adds a new commit and pushes it.
This is really useful when wanting to propagate (or refresh) the workflows managed by OpenTofu to all the `terraform-provider-*` repos.

### [`disable_unwanted_workflows.sh`](./disable_unwanted_workflows.sh)
This gets a file with repositories and disables all of their workflows that are not in the above mentioned workflows directory.
This is meant to be used to disable any workflow that is inherited from the forked repo that is not necessary in the OpenTofu world.
