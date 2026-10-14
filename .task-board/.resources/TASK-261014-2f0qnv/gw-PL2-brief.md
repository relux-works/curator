# THE ONLY CURRENT INSTRUCTION — TASK-261014-2f0qnv: implement `env resolve --repair --preserve-lock` (PL2, developer)

**Repository:** curator (Go), base `main`. Spec: curator-spec main `6648e18` (PR #146), in `~/Developer/ReluxWorks/platform/curator-spec`. Read `protocol/environments.md` §1.3 (lock identity migration paragraph), §10.1 ("Preserving the lock", the `--preserve-lock` grammar and the lock-changed race rule), §10.4 (the two rows `environment_repair_needs_lock_migration` and the extended `environment_repair_failed`), §12 (the status note), §13, and `profiles/manager.md` §12.5. Vectors: `conformance/environments-repair-preserve-lock/vectors.json` (10 cases).

**Current code.** `internal/envprofile/managed.go`: `Resolve` branches into `repairUnderLock` when `req.Repair && hashing.WriteVersion() == VersionV2 && lock.ContentHashVersion() == VersionV1`, and `repairUnderLock` runs `prepareIdentityMigration` + `publish`, which rewrites the lock. CLI flags are in `cmd/curator/env.go` (`cmdEnvResolve`).

**Do this.**
1. **Flag.** Add `--preserve-lock` to `env resolve`. It is valid only with `--repair`: without it, return the existing usage error before profile resolution, locking or writes, like the existing `--takeover`/`--dry-run` checks. Update the usage line and `cli` docs/help in this repository where `env resolve` flags are listed.
2. **Pre-check.** Under `--repair --preserve-lock`, decide whether the repair would change the lock bytes before `beginOperation` (the repair lock) and before any write, staging or store rebuild. Today that means a v1 lock with a v2 writer. Structure it as one predicate, "repair would change the lock hash", so a later lock-changing step plugs in. If it would, refuse with `environment_repair_needs_lock_migration`. Emit no fragment, write nothing, and name the operator command exactly as §10.1 specifies (rerun without `--preserve-lock`). Add the diagnostic constant next to the existing `Diag*` constants.
3. **Order relative to store rebuild.** `Resolve` can call `rebuildUntrustedStoreEntry` before the migration branch. Under `--preserve-lock`, make sure the refusal happens before any such rebuild write whenever the lock would change. Follow the spec's ordering text exactly; if the spec and the vectors disagree, stop and record it in the gaps file.
4. **Race.** Under `--preserve-lock`, after taking the repair lock, re-read the lock and compare its hash with the pre-check value. If it differs, refuse with `environment_repair_failed` (no fragment, no writes), per the §10.4 row.
5. **Otherwise** the repair behaves exactly as `--repair`, and the lock hash after equals the lock hash before. A current home emits without taking the lock, as today. Plain `--repair` keeps migrating a v1 lock (now spec-admitted, §1.3/§10.1).
6. **Tests (hermetic, temp homes; never a live home):** one test per vector case where this repository's harness can express it: refuse on v1+v2-writer for stale, unprovisioned and current homes, with no lock acquired, no writes and no fragment; v2 repair with equal lock hashes; the usage error without `--repair`; a current home with the lock held elsewhere still emits; plain-repair migration; the lock-changed race via a test seam; `env status` read-only on a current v1-lock home. Reason about mutants: moving the refusal after `beginOperation`, or dropping the race re-check, must fail a test. If this repository runs curator-spec vectors (search for conformance runners), wire the new family in.

**Rules:**
- R223: compile only locally (`go build ./...`, `go vet ./...`, focused `go test ./internal/envprofile/... ./cmd/curator/...`); hosted CI is the arbiter. Use `sh scripts/remote-gate.sh` for the full suite.
- Never touch `~/.curator` or any live home.
- Leave the candidate uncommitted in the managed Story worktree; the handoff snapshots it. Never pass `--no-gpg-sign`.
- Public repository: no private identifiers, seat names, host names or local paths in code, tests or docs.

**Hand-off:**
1. Attach `TASK-261014-2f0qnv_gaps.md`: the changes, a vector-to-test map, the remote-gate result, any spec/vector disagreement.
2. Run `task-board handoff TASK-261014-2f0qnv --role developer`.
3. END YOUR TURN.
