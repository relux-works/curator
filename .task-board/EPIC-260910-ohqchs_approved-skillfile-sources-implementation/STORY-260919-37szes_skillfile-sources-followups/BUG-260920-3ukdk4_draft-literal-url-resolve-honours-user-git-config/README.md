# BUG-260920-3ukdk4: draft-literal-url-resolve-honours-user-git-config

## Description
SECURITY (from TASK-260910-1xya7x review rev3, corpus case v2-user-insteadof-ignored): curator project resolve on a literal git: source (cmd/curator project_resolve.go -> gitops.Clone/Fetch) inherits the operator's git configuration; a hostile url.<evil>.insteadOf=<declared> redirects the clone and the lock binds the evil commit (row in internal/crossconformance/draftsources_semantic_v2_test.go reproduces). The resolved lane already isolates user config (TestUserConfigIgnoredByResolvedLane); the draft literal-URL lane must too (repository-transport: user configuration never consulted).

## Scope
cmd/curator project resolve/refresh Git acquisition path, internal/gitops clone/fetch environment isolation

## Acceptance Criteria
Draft literal-URL clone/fetch runs with user/system git config isolated (GIT_CONFIG_GLOBAL/GIT_CONFIG_SYSTEM or equivalent, HOME-independent), the corpus case v2-user-insteadof-ignored passes at the production entry (the known-gap marker in crossconformance flips to driven-pass), a narrowing mutant re-enabling user config is killed; legacy v1 lane byte-identical.
