# TASK-260729-2sxx7k: validate-accepted-curator-go-snapshot-on-linux

## Description
Run a non-gating native Linux validation of the independently accepted Curator Go composite snapshot from TASK-260720-1ljev5 while ssh lev is available.

## Scope
Use the preserved stable .temp/TASK-260720-1ljev5/worktree, never the concurrently changing TASK-260720-1nlmvv tree. Inventory ssh lev first; transfer a deterministic source archive into a private mktemp directory; verify source identity; use only an already installed approved Go toolchain; run native build, vet, focused/full tests and focused race gates as resources permit; capture exact exit codes and platform evidence; remove remote temporary bytes. No package install, PATH/profile/system mutation, product edit, release claim, stage, commit, publish or pin. This evidence must be rerun on the final integrated candidate.

## Acceptance Criteria
Linux OS/architecture/filesystem and trusted Go root are recorded; deterministic accepted snapshot identity is verified remotely; native Go gates have exact exit evidence or a precise preflight failure; no remote persistent state remains; task-scoped evidence explicitly states non-gating/non-release status.
