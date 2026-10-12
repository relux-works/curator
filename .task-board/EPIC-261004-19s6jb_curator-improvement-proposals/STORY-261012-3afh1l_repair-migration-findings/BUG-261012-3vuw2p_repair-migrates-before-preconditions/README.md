# BUG-261012-3vuw2p: repair-migrates-before-preconditions

## Description
env resolve <env> --repair (v0.15.0-rc.4) is not atomic. repairUnderLock publishes the identity migration first (managed.go:2572-2580: lock to schema 2 / hash 2, sibling env markers to v3). Only afterwards does it check the target env preconditions, e.g. the muse native credential target (effectivePassthrough / isCredentialRefusal, managed.go:2647-2651). A refusal such as environment_credential_conflict therefore leaves the profile half-migrated: 0.14 can no longer read it, and the requested env is still unprovisioned. Found by an isolated dry run of the Mac mini cut-over (a schema-1 lock, claude_code/codex_cli/pi current, muse unprovisioned, muse auth target absent).

## Scope
(define bug scope / affected area)

## Acceptance Criteria
A refused --repair leaves lock.json and all sibling env markers byte-identical; the identity migration publishes only after every target-env precondition passes, or rolls back atomically
