#!/bin/bash
# This script processes all the repos prefixed with `terraform-provider` from a given GitHub organisation
# and disables any workflow that is not in ./workflows.
#
# To run this, you need to have `gh` tool installed: https://cli.github.com/ and to be logged in accordingly.
usage () {
    cat <<HELP_USAGE
$0 <github org>

<github org>  The organisation (or owner) of the repositories to be processed
HELP_USAGE
}

if [ $# -ne 1 ]; then
  usage;
  exit 1
fi
owner="$1"

if [ -z "${owner}" ]; then
  echo "empty github organisation"
  usage;
  exit 1
fi

wanted_workflows="$(ls -A1 ./workflows)"

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
done <<< "$(gh repo list "${owner}" --json name,isArchived,owner --limit 300 --jq '.[] | select(.name | startswith("terraform-provider")) | select (.isArchived == false) | .owner.login+"/"+.name')"
