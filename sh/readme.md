# Shell Scripts

This directory, contains a collection of shell scripts that perform various tasks.

## Contents

- `check_releases.sh`: Can be used to check if all repository tags have releases with artefacts.

- `reset_repos.sh`: Can be used to reset a list of repositories to their upstream. It will also fetch tags and add additional files. 

- `workflows`: Is a collection of workflows which every provider should include by default. They enable automatic syncing, releasing using go releaser and additional workflows.
