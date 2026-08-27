# TASK-260720-17llva review verdict

## Verdict

Changes requested; route to `to-dev`.

## Required finding

P1 lifecycle-ordering mismatch: `profiles/manager.md` lines 263-270 acquires the manager-home mutation lock and recovers every incomplete journal before network access, audit, fingerprinting, and compilation. The attached normative activity model instead requires read-only planning and all private `go list`/`go build` staging to finish first; only after every private build succeeds may the manager acquire the home lock, recover journals, and revalidate shared state. The task AC independently requires real installs to build misses before mutation and requires a build failure to preserve the prior installation byte-for-byte. Early recovery can rewrite installation targets, consumers, journals, or live-cache state before a later preflight/build failure. Lines 277-280 redefine the protected prior state as the state after recovery and therefore do not satisfy that guarantee.

## Rework required

Remove the pre-build recovery pass and its exception. Keep the deterministic project locks, perform planning and private staging without shared mutation, and after successful staging acquire the manager-home lock, recover journals, and revalidate the cache/shared targets. If recovery changes any assumption, release the lock and restart from the earliest affected read or build. Ensure the failed-private-build branch leaves installation, consumers, and live caches exactly as they were when the operation began. Reconcile the recovery section and lock-order language with that single post-build recovery point.

## Coverage and validation

The remaining reviewed additions cover the five fixed Go argv forms, clean environment and process graph, source gates, dependency/input rejection, protected-cache trust, dry-run, publication, rollback, repair, status, and locked GC. `PATH=.temp/TASK-260720-1nvomm/venv/bin:$PATH python3 tools/validate.py` passed: 30 schemas and 93 vector files. `PATH=.temp/TASK-260720-1nvomm/venv/bin:$PATH make validate` passed: the scoped validator, 8 Python tests, and `go test ./tools/...`. `git diff --check -- profiles/manager.md` passed. The unqualified host Python lacked `jsonschema`; this is an environment dependency issue, while the task-prepared validation environment is green.