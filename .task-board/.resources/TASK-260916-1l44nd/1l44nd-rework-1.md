# TASK-260916-1l44nd rework 1 (orchestrator, binding)

Revision 1 gate FAILED (run 35667106874): `Race (ubuntu-latest)` and `Lint`; Test lanes green.

1. Linux (Race ubuntu, ~30 rows, one root cause): every session refuses
   `script_execution_capability_evidence_invalid: cannot add the Landlock rule for "/dev/null"`
   (`internal/scriptworker/landlock_linux.go`). Landlock path rules are `landlock_add_rule(LANDLOCK_RULE_PATH_BENEATH)`
   on an O_PATH fd of the path; `/dev/null` is a character device on devtmpfs — a rule on it
   with filesystem access rights it cannot carry (e.g. `LANDLOCK_ACCESS_FS_READ_DIR`/`MAKE_*`/
   `TRUNCATE` on a non-directory, or rights above the ruleset's `handled_access_fs`, or an ABI
   the runner kernel lacks) returns EINVAL. Fix the real cause: only pass the access rights that
   apply to the object type (file vs directory) and that are within the handled set for the
   probed ABI; for the standard streams / null device prefer not needing a rule at all (Landlock
   does not restrict already-open fds — bind `/dev/null` BEFORE `landlock_restrict_self`, or add
   a file-typed rule with `READ_FILE|WRITE_FILE` only). Prove the probe/apply pair on the
   ubuntu runner (the Test lane passed because the row set differs under `-race`? — check why the
   Test lane did not hit it and make the rows lane-independent).
2. Lint (golangci-lint v7, all in `internal/scriptworker`): `capabilities.go:334` QF1002 tagged
   switch; `inventory_unix.go:50` unused `platform` param; `landlock_linux.go:115,125` G103 unsafe
   — add the `// #nosec G103 -- <why: raw syscall args for landlock_*; no memory escapes>` audit
   comments the repo uses elsewhere; `worker_test.go:726` unused `request` param.
Continue from the revision-1 tree (no checkout/clean/stash); append "Revision 2" to results.md
(cause, fix, lane status); republish only on a green gate. No new scope; keep the honest
done/undone list.
