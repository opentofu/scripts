#!/bin/bash
# This script processes all the repos given in a file and disables any workflow that is not in the directory of workflows.
#
# To run this, you need to have `gh` tool installed: https://cli.github.com/ and to be logged in accordingly.
#
# The repositories in the given file should be in the following format: <owner>/<repo> (e.g.: opentofu/terraform-provider-aws)
# To know what workflows to disable, it lists the files inside the workflows directory and for any workflow file found in
# any given repo, it checks to be named the same with one in the workflows dir. If not, it will disable the workflow.
while getopts r:w: flag
do
    case "${flag}" in
        r) REPOS_FILE=${OPTARG};;
        w) WORKFLOW_DIR=${OPTARG};;
	*) echo "Usage: $0 -r <REPOS_FILE> -w <WORKFLOW_DIR>" >&2; exit 1;;
    esac
done

if [ -z "$REPOS_FILE" ] || [ -z "$WORKFLOW_DIR" ]; then
  echo "Usage: $0 -r <REPOS_FILE> -w <WORKFLOW_DIR>"
  exit 1
fi

if [ ! -f "$REPOS_FILE" ]; then
  echo "File '$REPOS_FILE' not found."
  exit 1
fi

wanted_workflows="$(ls -A1 "${WORKFLOW_DIR}")"

while IFS= read -r repo; do
  echo "Processing ${repo}"
  while IFS= read -r wf_file; do
    printf "\tProcessing %s from %s\n" "${wf_file}" "${repo}"
    found=$(echo "${wanted_workflows}" | grep "^${wf_file}$")
    if [ -z "${found}" ]; then
      printf "\t\tDisable %s from %s\n" "${wf_file}" "${repo}"
      gh workflow disable -R "${repo}" "${wf_file}"
    fi
    printf "\tProcessing %s from %s (DONE)\n" "${wf_file}" "${repo}"
  done <<< "$(gh workflow list --repo "${repo}" --json path --jq '.[].path' | rev | cut -d'/' -f1 | rev)"
  printf "Processing %s (DONE)\n" "${repo}"
done < "$REPOS_FILE"
