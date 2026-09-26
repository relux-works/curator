# STORY-260922-1cenbr: credential-modes-0017-implementation

## Description
Implementation of the adopted Decision 0017 (environment credential modes; curator-spec main 05053cd7: environments 7.4/7.7/8.4.1/10.1/10.4, manager 12.4/12.5) in Curator (Go): fix-first credential-link repairs, the explicit inspect -> plan -> apply migration under the manager lock, and production-entry tests on temporary stores covering both 0017 hazards including the operator-observed dangling Pi link. Follow-ups F-C1, F-C2, F-C3 from TASK-260921-3qcjsy results (as amended by TASK-260922-23ahj2).

## Scope
internal/envprofile managed.go/status.go and the env CLI surface (status, resolve, repair/migrate); docs; no schema change to the frozen v1 marker (credential-record members are F-S1)

## Acceptance Criteria
shared->isolated leaves no stale link; a regular file at a link path refuses with environment_credential_conflict naming the path; a dangling or mis-targeted link is reported detached by env status and resolve; the explicit migration moves a Pi wrong-target home to ~/.pi/agent with bytes preserved and mode intact, printing the plan before applying; codex isolated admitted only under effective file storage (absent cli_auth_credentials_store => file); every refusal has a mutant-killed production-entry row
