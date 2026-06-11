#!/bin/bash
# This checks the indicated files for the given copyright header.
#
# Some relevant things this script does:
# * From the given header, it removes any 4 digit number, as it is considered to be the year which can be different
#   from file to file.
# * It skips any Golang generated file
#
# Requires 3 inputs:
# * A string containing the header to be checked against. (e.g.: "Copyright (c) The OpenTofu Authors\nSPDX-License-Identifier: MPL-2.0\nCopyright (c) 2023 HashiCorp, Inc.\nSPDX-License-Identifier: MPL-2.0\n")
# * A string containing space separated patterns for the paths to be included in the scan (e.g.: '"*.go" "*.proto"')
# * A string containing space separated patterns for files or paths to ignore (e.g.: '"*/.git*" "*/vendor/*" "*/node_modules/*"')
#
# Example usage: ./copyright_check.sh "Copyright (c) The OpenTofu Authors\nSPDX-License-Identifier: MPL-2.0\nCopyright (c) 2023 HashiCorp, Inc.\nSPDX-License-Identifier: MPL-2.0\n" '"*.go" "*.proto"' '"*/.git*" "*/vendor/*" "*/node_modules/*"'


is_generated() {
  if grep -q -E '(^.{1,2} Code generated .* DO NOT EDIT\.\r?$)' "$1"; then
    return 0
  fi
  return 1
}
# Removes the year from the expected header and inlines everything. Additionally, replaces any double space with a single one
cleanup_header() {
  echo "$1" | sed -E 's/[0-9]{4}-[0-9]{4}//g;s/[0-9]{4}//g;' | tr '\n' ' ' | sed 's/  / /g'
}

header="Copyright (c) The OpenTofu Authors
SPDX-License-Identifier: MPL-2.0
Copyright (c) 2023 HashiCorp, Inc.
SPDX-License-Identifier: MPL-2.0
"
if [ -n "${1}" ]; then
  header="${1}"
fi
# Process the header to check against and print it
header_to_use=$(printf "%s" "${header}")
header_lines=$(echo "${header_to_use}" | wc -l)
clean_header="$(cleanup_header "${header_to_use}")"
printf "Header to check:\n---\n%s\n---\n" "${clean_header}"

# shellcheck disable=SC2206
include_patterns=(${2})
# shellcheck disable=SC2206
ignore_paths=(${3})

find_args=(".")
if [ ${#ignore_paths[@]} -gt 0 ]; then
  find_args+=("(")
  for i in "${!ignore_paths[@]}"; do
    if [ "$i" -gt 0 ]; then find_args+=("-o"); fi
    find_args+=("-path" "${ignore_paths[$i]}")
  done
  find_args+=(")" "-prune" "-o")
fi

find_args+=("(")
for i in "${!include_patterns[@]}"; do
  if [ "$i" -gt 0 ]; then find_args+=("-o"); fi
  find_args+=("-name" "${include_patterns[$i]}")
done
find_args+=(")" "-print")

echo "Find command ready: find ${find_args[*]}"
echo "Scanning files for copyright headers..."

mismatched_count=0
scanned_count=0
while IFS= read -r file; do
  ((scanned_count++))

  if is_generated "${file}"; then
  echo "Skipping: '${file}' is a generated file."
    continue
  fi

  # Read the first N lines of the source file depending on the template length.
  # Also, cleanup the comment syntax from the headers and inline everything
  file_header="$(head -n "${header_lines}" "${file}" | sed -e 's#^//##' -e 's/^#//' -e 's/^\s*//' || true)"
  clean_file_header="$(cleanup_header "${file_header}")"

  if ! echo "${clean_file_header}" | grep -q "^${clean_header}$"; then
    printf "ERROR: %s missing or not matching the expected header.\n\tWanted:\n\t\t%s\n\tGot:\n\t\t%s\n" "${file}" "${clean_header}" "${clean_file_header}"
    ((mismatched_count++))
  fi
done < <(find "${find_args[@]}")

printf "\nScan Summary: scanned=%d, mismatched=%d\n" ${scanned_count} ${mismatched_count}

if [ "${mismatched_count}" -gt 0 ]; then
  exit 1
fi
exit 0
