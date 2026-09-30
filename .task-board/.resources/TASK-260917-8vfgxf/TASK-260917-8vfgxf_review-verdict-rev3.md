# TASK-260917-8vfgxf — review verdict rev3: ACCEPTED

CR-TASK-260917-8vfgxf-3, base 0e3169bb, tree 0a10a669 (the worktree's write-tree reproduces 0a10a669 exactly), 10 paths.

## F1 (rev2): fixed
- In `internal/audit/audit.go` `gate`, `opaquescan.NULPaths` now runs for every subject before the `!cfg.Audit.Enabled` early return. That return now hands back `warnings, errs` instead of `nil, nil`. Everything else in the audit (the canary, the detectors, revocation and the cache) stays opt-in.
- Every admission site that goes through `audit.Gate`/`GateReadOnly` inherits the rule: install.go:580, global.go:198, main.go:1896/2198, and `strictAuditMember` (envprofile.go:1712).
- `contextaudit.Detect` also scans the whole snapshot. Its opaque findings cannot be waived. The auditMember, auditNewUpdateMember and migrateGlobalSkills errors now name the finding class and the file.
- When the scan fails, the snapshot is blocked. A failed read is never treated as absence.
- Symlinks and directories are not flagged.

## F2 (rev2): fixed
`auditFindingMessage` appends `(file: <path>)` to the message. The install-lane row asserts both `critical audit.opaque.nul-byte` and the file path.

## Rows (default config, audit disabled), re-run by the reviewer
- internal/install `-run 'NUL|Opaque'`: ok. The project and global lanes each cover the single-file and split-file colliding trees, which have equal `hashing.ContentSHA256` (the test asserts the equality). Both are refused and the target is not materialized. A deep `docs/deep/opaque.unsupported` file is refused. A NUL-free tree installs with status ok.
- internal/envprofile `-run 'NUL|Opaque|Colliding|TestManagerOwnedAbsenceReadsAreGuarded'`: ok. This covers the colliding trees, a deep NUL file in the context-path snapshot, the update lane, and the stateread guard.
- cmd/curator `-run 'NUL|Opaque'`: ok. The production external audit and the CLI both block with audit disabled.
- internal/audit, internal/contextaudit: ok.

## Mutants (disposable copy, `go test ./internal/install -run 'NUL|Opaque' -count=1`)
- M1: rule gated behind `Audit.Enabled` again → rc=1, killed.
- M2: rule limited to `assets/` → rc=1, killed (split-file-collision and deep-docs rows fail).
- M3: `opaqueNULReport` demoted to `DecisionWarn` → rc=1, killed.
- Restored tree → rc=0.

## Other checks
- Stateread guard: the allowlist entry moved from `detect` to `detectWithOpaquePaths` and was not widened. The new file reads go through internal/stateread (`opaquescan`).
- `git ls-files -co` shows no Windows-reserved names (the grep printed nothing).
- No CHANGELOG/LOGBOOK edits and no stray files.
- rc.13: the published spec has no wording or vectors for the interim rule. The rule is being added in TASK-260917-2vapkz (core §8), which is still in review, so none could be cited.

## Residuals (not blocking)
- The `RulesetVersion` bump to "2" invalidates cached verdicts. That is intended.
- `DetectFiles` (in-memory) and `Detect` (on disk) build their opaque findings on different paths. Both are covered by contextaudit tests.
