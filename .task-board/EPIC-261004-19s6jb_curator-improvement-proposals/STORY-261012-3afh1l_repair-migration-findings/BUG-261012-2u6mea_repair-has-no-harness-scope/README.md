# BUG-261012-2u6mea: repair-has-no-harness-scope

## Description
There is no per-harness scope for repair or migration: env resolve has only --profile, --repair, --dry-run, --takeover and --format (cmd/curator/env.go:125-132, the same at rc.5 and main), and the identity migration always covers every harness (migrationAdapters with an empty argument, identity_migration.go:91,170). A repair meant for one harness (muse) therefore rewrites the claude_code and codex_cli records that an older Curator still serves.

## Scope
(define bug scope / affected area)

## Acceptance Criteria
Either a scoped repair that touches only the named environment and leaves the shared lock readable, or a spec statement that repair is profile-wide, with the dry-run output listing every record and the lock that would change
