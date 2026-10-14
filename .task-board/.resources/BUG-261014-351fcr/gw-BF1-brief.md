# THE ONLY CURRENT INSTRUCTION — BUG-261014-351fcr: audit migrated global skills under the write hash version (BF1, developer)

**Repository:** curator (Go), base `main` 84cf54b9. Spec: curator-spec main, `protocol/core.md` §8 (interim reader rule, opaque files) and `protocol/environments.md` §9.1 (strict profile audit), in `~/Developer/ReluxWorks/platform/curator-spec`.

**The defect.** `migrateGlobalSkills` (`internal/envprofile/envprofile.go`) audits a migrated legacy branch-declared global skill under `hashing.VersionV1`: once directly (`contextaudit.DetectAtVersion(..., hashing.VersionV1, ...)`) and once through `strictAuditMember`, because the migrated `contextresolve.Resolved` leaves `HashVersion` zero and `resolvedHashVersion` maps zero to V1. The lock the migration produces is written under `hashing.WriteVersion()` (v2). Under v1 any NUL byte is a blocking, non-waivable opaque-file finding, so a skill repository that tracks a gzip file or any binary can never migrate: `profile_source_invalid`. The `global add --revision` path already audits under `WriteVersion()` (`internal/contextresolve/contextresolve.go`).

**Do this.**
1. Audit migrated global skills under `hashing.WriteVersion()` in both places: the direct `DetectAtVersion` call and the `strictAuditMember` call (set `HashVersion` on the migrated `Resolved`). Match the `global add --revision` path; do not add a second version-selection rule.
2. Make sure the lock the migration writes records hash version 2 when the write version is v2, exactly as the normal path does.
3. No path-based exemption (no `.task-board/**` or similar). Core §8 applies regardless of directory.
4. **Tests (hermetic, temp dirs, local bare git repos):**
   - a legacy branch-declared global skill whose commit tracks a gzip file and a small binary migrates successfully under v2, and the lock records hash version 2;
   - the same fixture with the write version forced to v1 (use an existing seam, or add the smallest test seam) still refuses with the opaque-file finding;
   - reason about mutants: reverting either of the two call sites to V1 must fail a test.
5. **Follow-up check.** `strictAuditMember` passes `capabilities.ImplicitNone()`, so a migrated skill whose scripts contain URLs may then fail `audit.capability.network-undeclared`. Write a test that shows whether it does. If it does, do NOT fix it here: record it in the gaps file with the failing test name and the exact finding, so the orchestrator files a separate leaf.

**Rules:**
- R223: compile only locally (`go build ./...`, `go vet ./...`, and the focused `go test ./internal/envprofile/...`); hosted CI is the arbiter. Use `sh scripts/remote-gate.sh` for the full suite.
- Never touch `~/.curator` or any live home. Tests use temp dirs.
- Leave the candidate uncommitted in the managed Story worktree; the handoff snapshots it. Never pass `--no-gpg-sign` anywhere.
- Public repository: no private identifiers or local paths in code, tests or docs.

**Hand-off:**
1. Attach `BUG-261014-351fcr_gaps.md`: the changes, the tests, the remote-gate result, and the point-5 outcome.
2. Run `task-board handoff BUG-261014-351fcr --role developer`.
3. END YOUR TURN.
