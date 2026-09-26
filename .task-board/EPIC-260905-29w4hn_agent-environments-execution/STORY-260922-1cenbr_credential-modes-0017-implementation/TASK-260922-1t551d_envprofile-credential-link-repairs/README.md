# TASK-260922-1t551d: envprofile-credential-link-repairs

## Description
F-C1: Fix-first credential-link repairs per environments 7.4/10.1 and manager 12.4: unlink only a recorded symlink that still targets the declared native store, preserving native bytes; refuse on a detached or regular file at the link path with environment_credential_conflict naming the path; shared to isolated removes the stale link; env status and resolve report a dangling or mis-targeted credential link as detached with environment_credential_conflict-class wording, never silence; codex_cli isolated admitted under effective file storage only, absent cli_auth_credentials_store means file, keyring or auto means environment_isolated_unsupported; marker entries record only what the frozen v1 schema allows, the extended record is F-S1.

## Scope
internal/envprofile and the env CLI surface; docs/troubleshooting; CHANGELOG

## Acceptance Criteria
shared to isolated leaves no stale link; a regular file at a link path refuses with the conflict diagnostic naming the path; the dangling link case from the operator machine is reported detached by env status and resolve; absent cli_auth_credentials_store resolves to file and admits isolated; a narrowing mutant per refusal
