# TASK-260905-30zs8t rework report 1 — stage (a) core findings F1–F6

Branch: `feat/agent-environments-stage-a` (curator worktree
`/Users/iv/Developer/ReluxWorks/.worktrees/curator-stage-a-core`).
Base of rework: `7238412c`. The acquisition has since landed on curator
main (`e8038558`, PR #58), so per the brief the branch was rebased onto
`origin/main` (`a2406dfe`) with `git rebase -S origin/main`: the 7
acquisition commits were skipped as already applied, the 8 stage-a/rework
commits replayed with zero conflicts, and
`git range-diff bb14375a..71ad8988 origin/main..HEAD` maps all 7
pre-existing commits 1:1 (`=`, identical content). New head `ac9d0037`;
all commits `G oparin@me.com`:

- `0f15cce1` F5 + F6 (ranges, weights)
- `44586376` F4 (honest skip class)
- `cfbac813` F1 + F2 + F3 (audit scope, default migration, lock + journal)
- `ac9d0037` cleanup (drop the unread operation field)

Spec authority: curator-spec main `f39f4a9`
(`protocol/environments.md` rev 1.1, schemas v1, conformance v1).
No push, no tag, no PR.

## Finding → disposition

- **F1 (blocking) — FIXED.** `contextaudit.Detect` now runs on the package
  root below a member directory through one shared `packageRoot` helper
  (`internal/envprofile/gitsource.go`), used by every consumer of a
  resolved member: the Update new-member pre-check, `auditAndStore` (detect,
  manifest load, MCP command check), and the switch materialization load.
  The Update pre-check additionally ensures the store entry before auditing
  (it previously read a possibly-absent path, which scans clean).
  Regression cover: `TestInstallDirectoryRejectsSecretUnderSubdirectory`
  (`Install` with `Directory: "sub"`, secret under `sub/`) and
  `TestInstallTransitiveDirectoryRejectsSecret` (clean root requiring a
  `requires.contexts[].directory` member with a secret) — each must fail
  installation with the blocking `context-secret-material` finding.
- **F2 (blocking) — FIXED.** `ensureDefault` materializes the synthesized
  contextless local root (`agent-context.json`: name `default`, version
  `0.0.0`, no `context` member — hence no root-context surface) into the
  store via `contextstore.EnsureState` and pins the resulting state hash;
  it migrates the machine's global skill set at migration time as skill lock
  members (git tag/revision declarations: fetched, pinned, stored, audited;
  see bounds). The lock writes before the install record so a failed
  migration retries instead of stranding a sourceless profile.
  Regression cover: `TestDefaultProfileMaterializesOnFreshHome` drives
  exactly the four reviewer-named paths on an isolated home — `List`,
  `Use default`, `Sync`, `Use --clear --env claude_code` — asserts markers
  in all four adapter homes, no root-context file, and migration
  idempotence; `TestEnsureDefaultMigratesGlobalSkills` migrates one
  tag-pinned skill from a local git remote and asserts the committed,
  stored pin.
- **F3 (major) — FIXED with one stated subset bound.** Every profile
  mutation (`Install`, `Update`, `Remove`, `Use`, `Sync`, plus `List` for
  first-use migration safety) holds the manager-home mutation lock
  (`managerlock` + `AcquireHomeOnly`, bounded 30 s wait,
  `environment_lock_unavailable` on contention) and recovers incomplete
  journals on entry; manager-home records (install records, locks, current
  and scope pointers) publish through `internal/transaction` as one Plan
  per operation (`internal/envprofile/lock.go`). Public entry points are
  thin locking wrappers over `*Locked` inners so the lock is never
  re-acquired (update→resync→use threads one operation).
  Subset bound, stated in the `switch.go` package doc and here: the
  per-entry agent-home payloads are not transaction targets in this stage.
  One Plan commits all-or-nothing with rollback, while §9.2 requires every
  entry attempted with partial success persisted as `profile_use_partial`;
  and the engine owns its backups as sidecars it deletes on success, while
  §8.3 requires the versioned `.agent-environment-backup/<n>/` generations
  with retention. Entries stay direct writes under the held lock with the
  verified M11 observable shape; re-running `profile use` converges the
  scope from the lock (the recorded current moves, journaled, only on full
  success). Follow-up: per-entry journaled plans when the engine grows a
  partial-persistence mode or the backup generations move into it.
  Regression cover: `TestMutationLockContention` (held lock →
  `environment_lock_unavailable` via `Install`),
  `TestOperationsLeaveNoJournalResidue` (no residue, byte-correct records),
  `TestIncompleteJournalRecoversOnEntry` (planted prepared journal
  completes on the next operation's entry).
- **F4 (major) — FIXED.** The stage-(b) skip no longer claims the phantom
  `CURATOR_STAGE_B` opt-in: new `stage-deferred` class in
  `skip-classes.tsv` ("the surface belongs to a later delivery stage"),
  ledger row 187 moved to it with a truthful reason, test prints the honest
  reason. `grep -rn CURATOR_STAGE_B` outside scratch: no hits. The 7
  sub-skips are exactly the reviewer's list (3 referenced-*, 4 mcp-*), all
  genuinely stage (b); no in-scope surface hidden.
- **F5 (minor) — FIXED.** `>`/`<` on a bare wildcard (`>*`, `<*`, `>x`,
  `<X`) are rejected as `profile_source_invalid` (node-semver reads them
  as match-nothing; §1.4 admits neither reading). Note: the published
  schema regex still admits the spellings, but the parser is the
  enforcement point — a manifest carrying one fails with
  `context_manifest_invalid` — so no silent meaning is substituted
  anywhere; a spec-schema tightening is flagged for the spec owner.
  Regression cover: `TestBareWildcardComparatorsAreRejected`.
- **F6 (minor) — FIXED by applying, no erratum needed.** Effective weights
  (§6 rules 1–3: manifest → agreeing direct-requirer edges → root map)
  now compute for every closure member kind; the overlay override is
  retained (overlays name context members only, so it is a no-op for other
  kinds). Regression cover: `TestRootWeightsApplyToSkillMembers` (root
  `weights: {"sk": 900}` over a skill member locks weight 900).
- **F7 — recorded bound, unchanged.** Scoped waivers still have no
  production path (both audit call sites pass nil; no machine-config
  surface). Operator surface arrives with the stage carrying
  manager-config schema 2 — the same stage that wires live direct
  declarations (`global add`/`remove` into the lock) and fetch-only
  `global update|upgrade`.

## Further bounds stated (no silent gaps)

- Migration carries git tag/revision global skills only: branch-pinned and
  local (sourceless) skills have no representable pin (a skill member pins
  a commit with its source; a state pin admits only a context member) and
  await the skill-pipeline stage. A global skill added after migration does
  not enter the default lock until the live-declaration wiring lands
  (`Update default` stays `profile_update_blocked`).
- Scope-clear (`use --clear`) is one unlink under the held lock, not a
  journaled absence target: crash-before retries, crash-after has
  converged; both states are retry-safe.
- `Remove` holds the lock; the profile directory goes with one `RemoveAll`
  under it (atomic at the manager's observation granularity), so no
  journal is needed for the removal itself.

## New tests (9)

`TestInstallDirectoryRejectsSecretUnderSubdirectory`,
`TestInstallTransitiveDirectoryRejectsSecret`,
`TestDefaultProfileMaterializesOnFreshHome`,
`TestEnsureDefaultMigratesGlobalSkills`,
`TestMutationLockContention`,
`TestOperationsLeaveNoJournalResidue`,
`TestIncompleteJournalRecoversOnEntry`,
`TestBareWildcardComparatorsAreRejected`,
`TestRootWeightsApplyToSkillMembers`.

## Mutants (all killed; reverts verified clean by `git diff` + green re-runs)

| Mutant | Narrowing (gate stays, admits exactly one) | Named failing test | Exit |
|---|---|---|---|
| M-F1: snapshot-root audit restored at both Detect sites | directory-addressed secrets admitted | TestInstallDirectoryRejectsSecretUnderSubdirectory + TestInstallTransitiveDirectoryRejectsSecret FAIL; TestInstallRejectsSecretMember still passes | 1 |
| M-F2-A: migrated skills dropped from the lock | skill-less default lock admitted | TestEnsureDefaultMigratesGlobalSkills FAILS (1 member) | 1 |
| M-F2-B: empty-tree hash pinned instead of the store key | sourceless pin admitted | both F2 tests FAIL; fresh-home fails with the reviewer's exact `agent-context.json is absent at …/e3b0c442…` | 1 |
| M-F3-lock: proceed unlocked + direct-write fallback | concurrent mutation admitted | TestMutationLockContention FAILS (install proceeds); residue + recovery still pass | 1 |
| M-F3-recover: Recover dropped from lock entry | crashed journal ignored | TestIncompleteJournalRecoversOnEntry FAILS (current "") | 1 |
| M-F5: wildcard >/ < back to match-everything | `>*` admitted | TestBareWildcardComparatorsAreRejected FAILS | 1 |
| M-F6: context-only weight computation restored | ignored skill weight admitted | TestRootWeightsApplyToSkillMembers FAILS (weight 0, want 900) | 1 |

No survivors.

## Gate outputs (real exit codes, standalone processes)

- `go build ./...` — 0 (also `GOOS=windows` 0, `GOOS=linux` 0)
- `go vet ./...` — 0
- `gofmt -l cmd internal` — clean
- `golangci-lint run` on envprofile + pkgversion + contextresolve — 0
  issues. (interop shows 4 findings in `context_resolution_test.go`, a
  file this rework does not touch; verified present at `7238412c` via
  stash — pre-existing, reported not fixed.)
- `go test -count=1 -race` on pkgversion, contextresolve, contextlock,
  contextstore, contextaudit, contextmaterialize, envmarker — all ok;
  `go test -count=1 -race ./internal/envprofile/` — ok
- Vector families, candidate root
  (`CURATOR_CONFORMANCE_ROOT=…/curator-spec/conformance/v1`): detectors,
  header, monolithic, resolution, versions — all PASS; exactly 7
  sub-skips, all `stage-deferred`/`tolerated-by-ledger` in
  `skips-observed.tsv`; 0 FAIL
- `bash .github/ci/gate-selftest.sh` — 0 (81 passed)
- `bash .github/ci/no-broad-suppression.sh` — 0
- `bash .github/ci/ledger-consistency.sh` — 0 (103 rows, linux/darwin/windows)
- `bash .github/ci/platform-case-gate.sh` on the combined internal +
  cmd/curator `-json` stream — gate verdict ok (26 skips recorded, all
  classified; `TestConformanceEnvironmentsMonolithic` ok with the 7
  `stage-deferred` sub-skips)
- `go test -count=1 -timeout 25m ./internal/...` — 0 (pre-rebase;
  range-diff proves identical rework content; all touched packages plus
  interop re-ran green on the rebased head — see below)
- `go test -count=1 -timeout 30m ./cmd/curator/` — 0 on the pre-rebase
  tree (`ok … 307.351s`, zero fail events in the `-json` stream) and 0 on
  the rebased head `ac9d0037` (`ok … 314.808s`, verdict read from the run
  log; exit code appended by the run wrapper itself)
- Post-rebase head `ac9d0037`, rerun by hand: `go build ./...` 0,
  `go vet` on the four touched packages 0, `gofmt -l` clean,
  `go test -count=1` on envprofile, pkgversion, contextresolve,
  contextlock, contextstore, contextaudit, contextmaterialize, envmarker —
  all ok, interop with candidate root — ok. (`GOOS=windows`/`linux`
  builds ran green on the pre-rebase tree; the rebase replayed identical
  content, verified by range-diff.)
- Reran-vs-accepted: the performer reran every gate above on the rework
  tree (this report's exit codes); cited from attached evidence only the
  producer's `./cmd/curator` timing context, never its verdict. The two
  `cmd/curator` runs here are both first-hand (pre- and post-rebase).

## AC coverage

Rework items F1–F6: 6 of 6 delivered with the dispositions above (F3 with
the stated agent-home subset bound); F7 stays a recorded bound. Every
production call site is named in its test's doc comment. CLI rows:
`profile list`, `profile use default`, `profile sync`,
`profile use --clear --env <id>` now work on a fresh home (driven in
`TestDefaultProfileMaterializesOnFreshHome`); all other stage-(a) rows
unchanged and green in the full suite.
