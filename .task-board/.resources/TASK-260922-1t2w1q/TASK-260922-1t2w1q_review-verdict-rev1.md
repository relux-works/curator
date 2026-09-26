# TASK-260922-1t2w1q revision 1 review

Verdict: CHANGES_REQUESTED. Route to to-dev. No human decision or external blocker.

Candidate tree 6007a2366be66471542e400ab84b9cb81f569a3a, base 637e5b43608b76664ed36cedc8f484c389669a2e. Reviewed a git-archive disposable copy under the assigned worktree. All 11 changed paths byte-match the candidate in both workspace and copy. No production edits or commits.

## Required fixes

1. **P1 — apply bypasses the prior-plan requirement.** internal/envprofile/migrate.go:790 checks drift only when Expect is nonempty; cmd/curator/envmigrate.go:57 accepts --apply without --expect. Independent real CLI test TestReviewerMigrationCLI/missing-plan provisions the operator Pi fixture, points its recorded link at the old root, invokes --apply without ever planning, and observes exit 0 plus the changed link. Require a prior complete plan identity on every apply entry, refuse missing evidence before mutation, update repair hints/docs, and replace the current test that deliberately blesses no-expect apply. Binding review note explicitly requires apply to require a plan.

2. **P1 — the plan is printed after mutation.** cmd/curator/envmigrate.go:58-60 calls ApplyMigration to completion before Fprint. TestReviewerMigrationCLI/print-before-write supplies a stdout writer that checks the link on the first write: it already targets agent/auth.json. Existing test only checks that final output contains plan text and cannot prove ordering. Print the locked, revalidated plan before the first mutation and fail if output cannot be delivered; add a temporal CLI assertion (not final-string containment).

3. **P1 — link mutation is neither atomic nor durably journaled, and syscall failure escapes rollback.** internal/envprofile/migrate.go:859-872 removes the old link before creating its replacement and appends the operation to applied only after success. If Symlink fails, the removed link is absent from rollback. TestReviewerFailedRelinkRollback reproduces this using an OS-rejected oversized target: applied=0, rollback=nil, old link missing. This is a helper-level syscall failure probe, not claimed CLI reachability of that oversized fixture; the same production sequence is invoked at ApplyMigration:811 and exposes ordinary symlink failures/crash windows. Additionally, relink-only plans never call op.publish; only marker bytes are journaled at :822-826, after link changes. Implement the required durable migration journal and temp-link/atomic replacement with recovery of partial apply, including the currently failing operation. Exercise syscall-boundary failure and process interruption, not only InjectFault after a successful operation. A failed marker publication currently explicitly leaves link changes standing, contrary to the advertised rollback guarantee.

4. **P2 — plan hash omits old-marker identity.** internal/envprofile/migrate.go:592-625 hashes a projection of mode/entries/operations, never markerRaw or a full marker digest. TestReviewerMarkerDrift plans through PlanMigration, changes valid marker profile.lock_sha256, then calls ApplyMigration with the old hash; apply succeeds. Include the inventoried marker identity in the plan hash and refuse this drift with no mutation. Do not read credential bytes to implement this check.

## Independent evidence

Shell zsh, set -o pipefail, Go -count=1, bounded timeouts. Attached logs and reproducible Go probe files.
- Focused exact-candidate envprofile migration + changed F-C1 rows: PASS, 7.681s, exit 0.
- Exact-candidate CLI TestEnvMigrate* and TestEnvResolveRepairNeedsMigration: PASS, 34.245s, exit 0.
- Independent CLI probes: 2/2 contract assertions fail, exit 1, 17.441s.
- Independent marker drift and syscall rollback probes: 2/2 contract assertions fail, exit 1, 0.746s. Initial marker probe used whitespace; final attached probe changes the valid profile.lock_sha256 field. One rerun from the outer module failed package setup; corrected cwd rerun produced the behavioral failures above, not counted as a gate failure.
- git diff --check exact delta: PASS.
- GitHub gate run 35696877595 succeeded; queried head 7e8cadf1f70d86335db5540c1e75918d3e962475 resolves to the exact candidate tree. Hosted Ubuntu/macOS/Windows test jobs, Ubuntu/macOS race jobs, lint, interop/naming and gate-self-test jobs succeeded. Rose-air and candidate-suite jobs skipped. Reused this evidence; did not rerun full landing suite.

Read F-C1 results and revision-4 accepted verdict and F-C2 producer results. The explicit no-migration-in-repair requirement supersedes F-C1 repair behavior; the updated repair refusal rows pass. No new credential-byte reads/copies were found in the production migration path (marker and native configuration reads are separate). No native host credentials were used by reviewer fixtures.

Bounds: rejection does not certify all acceptance criteria. No independent replay of producer mutation campaign; skip-lock kill is not established by this review. Producer attached results end with a literal truncation marker in the mutant table, so missing validation/mutant details are unknown, not accepted. Committed CLI cases cover regular-file refusal; other conflict and rollback rows primarily drive library entries. Existing no-copy walk excludes manager state, so it does not prove absence of secret copies anywhere. F-C3 classification follow-through remains out of scope. Durable process-crash behavior is a static gap here, not a claimed executed crash experiment.

Run goal queried: none. Attach this verdict and supporting evidence before setting to-dev. A fresh producer revision and independent review are required.
