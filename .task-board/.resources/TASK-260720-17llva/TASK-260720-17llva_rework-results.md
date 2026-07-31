# TASK-260720-17llva rework results

## Scope

Reworked only `profiles/manager.md` in `curator-spec` at the required baseline `57c1f56846d221ecc55786bd3c2467ec32f11730`. Existing prerequisite edits in `protocol/core.md`, `SECURITY.md`, and `decisions/0004-compile-only-build-drivers.md` remained untouched.

## Reviewer finding resolved

- Removed the pre-build manager-home recovery pass and its shared-mutation exception.
- Required every current-operation cache miss to be preflighted, built, and verified in private staging before recovery or any other shared-state mutation.
- Required failed preflight/build operations to leave the installation, consumers, and live caches byte-for-byte as they were when the operation began.
- Placed recovery at the start of the single manager-home locked publication phase after private builds succeed.
- Required recovery and locked revalidation drift to release the home lock and restart from the earliest affected read or build.
- Reconciled install/repair recovery and lock-order wording with that one post-build recovery point while retaining home-lock-only standalone recovery and GC.

## Acceptance coverage

The profile retains the previously reviewed source gates, exact five Go argv forms, clean environment and closed process graph, complete `go list` dependency/input rejection, protected cache trust and receipts, compiler-free dry-run, deterministic publication, one journal, consumer-last commit, reverse rollback, repair, status/currentness, and locked GC requirements.

## Verification

- `PATH=.temp/TASK-260720-1nvomm/venv/bin:$PATH python3 tools/validate.py` — passed; validated 30 schemas and 93 vector files.
- `PATH=.temp/TASK-260720-1nvomm/venv/bin:$PATH make validate` — passed; scoped validation, 8 Python tests, and `go test ./tools/...` all green.
- `git diff --check -- profiles/manager.md` — passed.

No test fixture was added because the owned scope is limited to the normative `profiles/manager.md`; the existing repository validators and tests cover this documentation-only change.
