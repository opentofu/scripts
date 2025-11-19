#!/bin/bash
# This checks if each tag from of a repository has an associated release entry.
#
# Requires 2 inputs:
# * A file that should contain the repositories that will do the check against.
# * An environment variable "GITHUB_PAT" that should point to a GitHub Access Token with enough rights to read the tags and releases of the repos contained by the file above.
#
# The repositories in the given file should be in the following format: <owner>/<repo> (e.g.: opentofu/terraform-provider-aws)
#
# Example usage: GITHUB_PAT=<token> ./check_releases.sh repos_file

[ -z "${GITHUB_PAT}" ] && echo "Expected GITHUB_PAT environment variable is missing" && exit 1
[ $# -ne 1 ] && echo "Usage: $0 <REPOS_FILE>" && exit 1

GITHUB_TOKEN="$GITHUB_PAT"
REPO_LIST_FILE="$1"

# Read each line from the file
while IFS= read -r repo; do
  OWNER="$(echo "$repo" | cut -d'/' -f1)"
  REPO="$(echo "$repo" | cut -d'/' -f2)"

  echo "Checking releases for $OWNER/$REPO..."

  # Make the API request to list tags with the PAT in the header
  tags=$(curl -s -H "Authorization: token $GITHUB_TOKEN" \
                  "https://api.github.com/repos/${OWNER}/${REPO}/tags" | jq -r '.[].name')

  # Loop through the tags and check if there's a corresponding release
  for tag in $tags; do
    # Make the API request to list releases for the tag with the PAT in the header
    releases=$(curl -s -H "Authorization: token $GITHUB_TOKEN" \
                      "https://api.github.com/repos/${OWNER}/${REPO}/releases/tags/${tag}" | jq -r '.tag_name')

    # Check if a release exists for the tag
    if [ -n "$releases" ]; then
      echo "Release found for tag: $tag in $OWNER/$REPO"
    else
      echo "No release found for tag: $tag in $OWNER/$REPO"
    fi
  done
  echo "Done checking releases for $OWNER/$REPO"
done < "$REPO_LIST_FILE"
