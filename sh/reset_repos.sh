#!/bin/bash
#
# Example usage ./reset-repos.sh -u fork_urls.txt -w /Users/tomas/testdir/workflows
set -o errexit

while getopts u:w: flag
do
    case "${flag}" in
        u) URL_FILE=${OPTARG};;
        w) WORKFLOW_DIR=${OPTARG};;
	*) echo "Usage: $0 -u <URL_FILE> -w <WORKFLOW_DIR>" >&2; exit 1;;
    esac
done

if [ -z "$URL_FILE" ] || [ -z "$WORKFLOW_DIR" ]; then
  echo "Usage: $0 -u <URL_FILE> -w <WORKFLOW_DIR>"
  exit 1
fi

if [ ! -f "$URL_FILE" ]; then
  echo "File '$URL_FILE' not found."
  exit 1
fi

while IFS= read -r FORK_URL; do
  UPSTREAM_URL=${FORK_URL//opentofu/terraform-providers}

  echo "Got fork URL: $FORK_URL"
  echo "Got upstream URL: $UPSTREAM_URL"

  git clone "$FORK_URL"
  REPO_NAME="$(basename "$FORK_URL" .git)"
  cd "$REPO_NAME"

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
