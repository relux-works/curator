# TASK-260922-1ejkxv review verdict — CR rev1: ACCEPTED

Reviewer: claude-opus-5-5 (low). Candidate tree 029bdd27 over base eadb1c0, checked out in a disposable clone (git read-tree; write-tree == 029bdd27).

## Checks
1. Schema: `system-config-v2` `environments.isolation.<profile>.<env>` enum is now `["shared","isolated"]`. Shape and whole-map replacement under manager §1 rule 2 are unchanged, and no new knob was added, so 0017 choice 5 still holds. Manager §1 rule 1 text is updated.
2. Environments §12.2 states both directions. Silence resolves to the locked direction. Explicit `shared` under an isolated lock is refused with the new code `environment_isolation_lock_conflict`, which is registered in the environments diagnostics tables and in manager §12.4/§12.5. An existing shared passthrough is refused with `environment_credential_conflict`, is never unlinked or migrated silently, and points to the F-C2 inspect→plan→apply migration.
3. Cases: valid-isolation-isolated-direction (renamed from the old invalid case), valid-isolation-shared-direction, invalid-isolation-unknown-direction ("automatic"), invalid-isolation-both-directions (array). The merged hunks from 1xbrz6 are additive only in the shared files.
4. results.md names the follow-up leaf "Enforce the isolated environment lock in the environment manager" with `environments.isolation` and `environment_isolation_lock_conflict`.
5. The gate table covers all three `validate:` lines plus regenerate-check, and the interrupted exit 130 is labelled as not green.

## Independent reruns (in the clone, zsh, pipefail)
- `python3 tools/validate.py` → exit 0, "validated 64 schemas and 1169 vector files"
- `go test ./tools/generate-vectors/` → ok
- `make regenerate-check` → exit 0 (no diff)

## Mutants
- M1: enum narrowed to ["shared"] → FAIL: valid-isolation-isolated-direction expected valid. Caught.
- M2: enum widened to any string → FAIL: invalid-isolation-unknown-direction expected invalid. Caught.

## Non-blocking observation
Manager §1 rule 2 still reads "overrides a user value with a warning". The isolated-lock refusal of an explicit `shared` is carved out in rule 1 and environments §12.2, as the brief requires, but rule 2 itself does not cross-reference that exception. The curator follow-up can cite §12.2.

Process note: one of my validate/sed commands ran in the Story worktree by mistake because a `cd` failed. The sed was reverted from the index (`git checkout --`), and `git status` confirmed the worktree is identical to the staged candidate. regenerate-check there produced no diff.
