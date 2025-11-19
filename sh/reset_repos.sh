#!/bin/bash
# For each repo in a given file, it's cloning it, setting the upstream repo (by replacing `opentofu` with `hashicorp` namespace) and
# resets the given repo main branch to the upstream one. After this has been done, it adds the workflow files from the -w argument directory
# into `.github/workflows` and adds a new commit and pushes it.
#
# Each line of the file given by the -u argument should be a fully qualified url of the repository that wants to be processed.
# E.g.: https://github.com/opentofu/terraform-provider-arukas

# Example usage: ./reset-repos.sh -u fork_urls -w "$(pwd)/workflows"
set -o errexit

while getopts u:w: flag
do
    case "${flag}" in
        u) URL_FILE=${OPTARG};;
        w) WORKFLOW_DIR=${OPTARG};;
	*) echo "Usage: $0 -u <URL_FILE> -w <WORKFLOW_DIR>" >&2; exit 1;;
    esac
done

{ [ -z "$URL_FILE" ] || [ -z "$WORKFLOW_DIR" ]; } && echo "Usage: $0 -u <URL_FILE> -w <WORKFLOW_DIR>" && exit 1
[ ! -f "$URL_FILE" ] && echo "File '$URL_FILE' not found." && exit 1

while IFS= read -r FORK_URL; do
  # Note: Not all providers exist in the same "hashicorp" org, some are in "terraform-providers".
  # When providing a list of URLs, make sure to not include different upstream orgs.
  UPSTREAM_URL=${FORK_URL//opentofu/hashicorp}

  echo "Got fork URL: $FORK_URL"
  echo "Got upstream URL: $UPSTREAM_URL"

  git clone "$FORK_URL"
  REPO_NAME="$(basename "$FORK_URL" .git)"
  cd "$REPO_NAME"

  git config user.name "OpenTofu Core Development Team"
  git config user.email "core@opentofu.org"

  git remote add upstream "$UPSTREAM_URL"
  git fetch upstream

  MAIN_BRANCH=$(git remote show upstream | grep "HEAD branch" | cut -d ":" -f 2 | xargs)

  git checkout "$MAIN_BRANCH"

  echo "Resetting $MAIN_BRANCH to upstream/$MAIN_BRANCH"

  git reset --hard upstream/"$MAIN_BRANCH"

  echo "Committing workflow changes to $FORK_URL"

  if [ ! -d ./.github/workflows ]; then
    mkdir -p ./.github/workflows
  fi

  cp -r "$WORKFLOW_DIR/." ./.github/workflows/
  git add .
  git commit -m "Apply GitHub workflow changes"

  git push origin "$MAIN_BRANCH" --force

  cd ..

  echo "Done! Removing $REPO_NAME"

  rm -rf "$REPO_NAME"

done < "$URL_FILE"
