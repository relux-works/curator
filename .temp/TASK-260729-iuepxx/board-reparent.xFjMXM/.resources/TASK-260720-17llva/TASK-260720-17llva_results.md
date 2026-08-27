# TASK-260720-17llva implementation results

## Scope

Extended only profiles/manager.md in curator-spec at required baseline 57c1f56846d221ecc55786bd3c2467ec32f11730. Existing prerequisite edits in protocol/core.md, SECURITY.md, and decisions/0004-compile-only-build-drivers.md were read as accepted inputs and left untouched.

## Contract coverage

The profile now normatively specifies immutable snapshot validation and build-source hashing before cache or compiler work, static build-root context exclusion, provider-first closure and collision gates, audit and registry ordering, trusted allowlisted Go release families, curator-go-toolchain-v1, private telemetry-off initialization, the exact five direct argv forms, the clean native vendor-only environment, full go-list dependency and input rejection, protected cache receipts and currentness, compiler-free dry-run, private build staging, deterministic project and manager-home locks, protected publication, one journal, consumer-last commit, reverse rollback, recovery, repair, status, and locked GC.

## Verification

PATH=.temp/TASK-260720-1nvomm/venv/bin:$PATH python3 tools/validate.py: passed, validated 30 schemas and 93 vector files.
PATH=.temp/TASK-260720-1nvomm/venv/bin:$PATH make validate: passed, including the scoped validator, 8 Python tests, and go test ./tools/....
git diff --check -- profiles/manager.md: passed.

No new test fixture was added because task ownership is limited to profiles/manager.md and the change is normative documentation. Existing repository validation and test gates cover this owned scope.