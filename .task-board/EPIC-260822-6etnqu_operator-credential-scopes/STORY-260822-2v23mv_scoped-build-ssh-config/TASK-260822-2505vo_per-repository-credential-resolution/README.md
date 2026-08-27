# TASK-260822-2505vo: per-repository-credential-resolution

## Description
Resolve SSH credentials per external build repository in the install path (internal/install/external.go + buildrepo policy construction): explicit flags/env selection covers every repository; otherwise the longest matching build_ssh scope for the repository's canonical identity; repositories on https or local substitution need none. Identity/known_hosts values from config expand '~'. Empty selection for an ssh repository fails closed with the protocol error build_repository_ssh_credential_missing.

## Scope
(define task scope)

## Acceptance Criteria
Precedence covered by tests (flags/env beat config; config scope selected per repository; https/local skip); pinned-agent, agent-only, and identity-only selections all reach the SSH policy; go test green.
