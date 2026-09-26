# TASK-260907-187z6x — git same-source reinstall drops --use and --takeover

## Defect

`profile install <git-url> --use --takeover` on an already-installed git root
(the retry the §9.5 onboarding stop invites) accepted both flags, did nothing,
and exited 0 saying `updated profile`. Trunk defect: the same-source reinstall
of a git root delegated to `updateLocked`, which has no install-row activation
handling, while a path reinstall goes through `reinstallPathLocked` /
`reinstallActivation` / `activateReinstall`, which honour the flags.

## Fix (path-reinstall parity)

`internal/envprofile/envprofile.go`, `installLocked` git same-source branch:
after `updateLocked` re-resolves as an update, the branch now runs the install
row activation through the shared seams — `reinstallActivation` (activate when
the machine has no current, or `--use` is given and the root is not current;
`--takeover` alone activates nothing) and `activateReinstall` (the §9.2
machine-scope switch through `useLocked` with the operation takeover flag,
then the scoped resync). Reported as an update in every case, exactly as
before. Path-root code paths are untouched; only the `reinstallActivation`
doc comment was generalized to the now-shared rule.

No normative reason forbids parity, so none is cited as a refusal. Parity is
required by the pinned curator-spec:

- environments §9.1: activation on install applies to every install-row
  invocation (current set only when the machine has none or `--use` is
  passed); the reinstall sentence governs resolution and reporting only
  ("re-resolves exactly as `profile update` does ... and is reported as
  an update").
- cli/curator.md install row: `--use` activates the installed root and
  `[--takeover]` takes over the unmanaged files the install would write,
  with no reinstall carve-out.
- environments §9.5: `--takeover` is carried by `profile install` and covers
  the files the carrying operation would write.

Behaviour after the fix:

- Retry with `--use --takeover` converges: takeover (notice plus §8.3 backup)
  and activation, exit 0, `updated profile <name> (lock …)` plus the switch
  lines.
- Retry with `--use` but no `--takeover` over still-unmanaged state surfaces
  the stop loudly: exit 1, `environment_surface_unmanaged_conflict` per entry
  plus `profile_use_partial`, recorded current unchanged. Per pinned
  environments §9.2 every adapter is attempted and the successful switches
  stand: `codex_cli`, `opencode` and `pi` DO switch, while the blocked
  `claude_code` home is neither written nor backed up. This is the DoD row
  "the stop is surfaced with its diagnostic and non-zero exit, and neither
  flag takes effect": it describes the still-blocked retry. A
  `--takeover`-carrying retry converges by §9.5 design (takeover clears the
  stop), so it exits 0 having done the work.

## Regression test (production CLI entry, eight rows)

`cmd/curator/profile_git_reinstall_test.go`:
`TestProfileGitReinstallHonoursUseAndTakeover` drives `run()` for all four
flag combinations on a git root that is current and one that is not, using the
repository insteadOf fixture pattern (local repo served under
`https://example.com/groot` through `GIT_CONFIG_GLOBAL`; no `file://`
operand). An unmanaged blocker (hand-written `CLAUDE.md`, no marker) sits in
the machine scope in every row. Each row asserts the exit code, the exact
operator line (`updated profile groot (lock …)` — all eight rows since
revision 2; row 2 discarded stdout in revision 1), the activation outcome
(recorded current via `profile list`, materialized bytes, switch lines), and
the backup outcome.

| # | currency | flags | exit | current after | writes |
|---|----------|-------|------|---------------|--------|
| 1 | not current | none | 0, `updated profile` | other | none, blocker untouched |
| 2 | not current | `--use` | 1, `updated profile` + `environment_surface_unmanaged_conflict` + `profile_use_partial` | other | blocked claude home untouched, no backup; codex/opencode/pi switch per §9.2 |
| 3 | not current | `--takeover` | 0, `updated profile` | other | none, no backup |
| 4 | not current | `--use --takeover` | 0, `updated profile` + `switched` + takeover notice | groot | backup generation 1 holds the operator bytes, surfaces carry groot |
| 5 | current | none | 0, `updated profile` | groot | none, blocker untouched |
| 6 | current | `--use` | 0, `updated profile`, no re-switch | groot | none, no backup |
| 7 | current | `--takeover` | 0, `updated profile` | groot | none, no backup |
| 8 | current | `--use --takeover` | 0, `updated profile`, no re-switch | groot | none, no backup |

## Narrowing mutant

Mutant: drop the activation call on the git reinstall branch (restore the
trunk shape — `updateLocked`, then return). Result, `go test ./cmd/curator/
-run 'TestProfileGitReinstallHonoursUseAndTakeover/not-current/use'`:

- `not-current/use without takeover attempts the switch and fails loudly`:
  FAILS (`install --use = 0, want 1`) — reproduces the trunk defect.
- `not-current/use with takeover takes over and activates`: FAILS (bare
  `updated profile groot (lock …)`, no switch) — the exact reported symptom.
- `not-current/takeover without use switches nothing`: passes on the mutant,
  proving the mutant is narrowing rather than broadly breaking.

Mutant reverted after the kill; the fix diff was re-verified byte-identical.

## Narrow test exit codes (shell: bash, `set -o pipefail`)

- `go build ./internal/envprofile/ ./cmd/curator/` → exit 0
- `go vet ./internal/envprofile/ ./cmd/curator/` → exit 0
- `golangci-lint run ./cmd/curator/ ./internal/envprofile/` → exit 0, 0 issues
- `gofmt -l` on changed files → clean
- `go test ./cmd/curator/ -run
  'TestProfileGitReinstallHonoursUseAndTakeover|TestProfileInstallReinstallHonoursUseAndTakeover|TestProfileUseTakeoverRow|TestProfileUpdatePinnedTagIsUnchanged|TestProfileInstallListUse|TestProfileInstallUseActivatesThroughSwitch|TestProfileInstallUsePartialLeavesCurrent|TestGitFileURLHasNoBackslash'
  -count=1` → exit 0 (all 8 pass, incl. the new 8-row test and the path
  sibling — no path-root behaviour change)
- `go test ./internal/envprofile/ -run
  'Reinstall|Takeover|InstallUse|Snapshot|ForeignSymlink|ReadOnlyCommands'
  -count=1` → exit 0
- `go test ./internal/envprofile/ -count=1` split into four bounded
  `-run 'Test[A-G]' / 'Test[H-M]' / 'Test[N-S]' / 'Test[T-Z]'` calls
  (`-timeout 9m` each) → exit 0, 0, 0, 0 (whole 157-test package green).
  One un-split full-package run hit the go 10-minute default timeout while
  running concurrently with the CLI suite; the four sequential subsets
  supersede it.

## Notes

- CHANGELOG: Unreleased → Fixed entry added.
- Platform-case ledger: no new row registered. Git-fixture (insteadOf) tests
  are systematically unregistered in `.github/ci/platform-cases.tsv`
  (e.g. `TestProfileUpdatePinnedTagIsUnchanged` and every `serve()`-based
  envprofile test are absent), so the ledger demands no row for this test.
- Work left uncommitted in the story worktree as required. Files changed:
  `internal/envprofile/envprofile.go` (fix), `CHANGELOG.md` (Fixed entry),
  `cmd/curator/profile_git_reinstall_test.go` (new, 8 rows).

## Revision 2 (rework 1: R1 wording fix + R2 evidence correction)

Verdict on revision 1: CHANGES_REQUESTED (one medium finding + one evidence
correction). Normative reading, activation semantics, takeover safety, both
mutants, and the ledger position were VERIFIED and are not re-opened.

### R1 — partial activation of a reinstall reports "updated", not "installed"

`cmd/curator/profile.go`, `cmdProfileInstall` error branch: when the install
row's activation fails partially, the branch printed `installed profile …`
unconditionally, ignoring the returned `updated` flag. The lock/source record
of a same-source reinstall WAS updated — only the switch did not complete —
so §9.1's reinstall reporting rule requires `updated profile …`. The branch
now honours the flag: `updated` → `updated profile <name> (lock …)`,
otherwise the original `installed profile …` line. Exit code and diagnostics
are byte-unchanged (exit 1, `environment_surface_unmanaged_conflict` +
`profile_use_partial`).

Row 2 (`not-current/use without takeover`) now captures stdout and asserts
`updatedLine`, the same operator-line pin the seven success rows have. The
two neither-flag rows gained the missing no-backup assertions the verdict
required, so all eight backup claims are checked.

Narrowing mutant (restore the unconditional `installed` line in
`cmdProfileInstall`): `go test ./cmd/curator/ -run
'TestProfileGitReinstallHonoursUseAndTakeover/not-current/use_without'
-count=1` → exit 1, `operator line "installed profile groot (lock
sha256:…)", want the reinstall reported as an update`. Row 2 kills it; the
fix was restored byte-identical afterwards (re-verified green).

Path-identity check (reviewer bound): the path sibling
`TestProfileInstallReinstallHonoursUseAndTakeover` is green unchanged, and a
throwaway probe (deleted before handoff) showed the path failing row now
prints `updated profile tk (lock …)` with exit 1 and the same two
diagnostics — the identical §9.1 wording correction on a line that test
never pinned — while a first-install partial still prints `installed
profile beta (lock …)`. Pinned path assertions, exit codes, diagnostics,
and activation outcomes are identical; only the previously-unpinned
reinstall-partial word changed, which is the fix, not a regression.
`internal/envprofile/envprofile.go` and `CHANGELOG.md` are untouched in
revision 2, so the revision-1 activation-drop mutant evidence (independently
re-run by the reviewer) stands.

### R2 — row-2 prose corrected (no code change)

"Nothing written, no backup" now reads as it behaves: the blocked
`claude_code` home is neither written nor backed up; `codex_cli`, `opencode`
and `pi` switch and their switches stand (§9.2: attempt every entry, report
the partial); recorded current stays `other`; exit 1. Table row 2 and the
behaviour bullet above are rewritten accordingly.

### Revision-2 narrow exit codes (shell: bash, `set -o pipefail`)

- `gofmt -l` on changed files → clean; `git diff --check` → clean
- `go build ./cmd/curator/ ./internal/envprofile/` → exit 0
- `go vet ./cmd/curator/ ./internal/envprofile/` → exit 0
- `golangci-lint run ./cmd/curator/ ./internal/envprofile/` → exit 0, 0 issues
- `go test ./cmd/curator/ -run
  'TestProfileGitReinstallHonoursUseAndTakeover' -count=1` → exit 0, 8/8
  pass (54.7s)
- `go test ./cmd/curator/ -run
  'TestProfileInstallReinstallHonoursUseAndTakeover' -count=1` → exit 0,
  5/5 pass (17.4s, path sibling — no path behaviour change)
- `go test ./cmd/curator/ -run
  'TestProfileInstallUsePartialLeavesCurrent|TestProfileInstallUseActivatesThroughSwitch|TestProfileUseTakeoverRow|TestProfileUpdatePinnedTagIsUnchanged|TestProfileInstallListUse'
  -count=1` → exit 0 (14.8s; first-install partial wording branch preserved)
- `go test ./internal/envprofile/ -run
  'Reinstall|Takeover|InstallUse|Snapshot' -count=1` → exit 0 (55.3s)
- Wording mutant row-2 run → exit 1 with the pinned message (kill); fix
  restored, row re-run → exit 0

Base-refresh note: no base refresh landed in this worktree for revision 2
(HEAD is still `09b25ef6`); no previously-skipped row entered execution —
all eight git rows, all five path-sibling rows, and the neighbour tests
above executed without skips.

Files changed in revision 2: `cmd/curator/profile.go` (reporting fix),
`cmd/curator/profile_git_reinstall_test.go` (row-2 operator-line pin +
neither-flag backup pins). Work left uncommitted as required.

## Revision 3 (republish, no content change)

revision 3 = revision 2 unchanged; republished under the new board binary so the validation evidence is tree-bound (validation_not_bound_to_tree).

## Revision 4 (run RUN-260923-6201fe; no content change)

Revision 3's hosted gate failed only on `Test (macos-latest)` (`go test`
exit=1; ubuntu/windows/race-macos green, platform-case gate ok). The hosted
tail lists only `ok` rows, so the failing test name is unrecoverable from
`TASK-260907-187z6x_change-request_rev3-validation.log`. This run hunted it
locally on the shared mac host under the gate's own conditions
(`CURATOR_CONFORMANCE_ROOT` at SPEC_PIN `87a0d00`, suite-plan served=76).

Tree state: `git status --short` shows exactly the rev3 path set —
`M CHANGELOG.md`, `M cmd/curator/profile.go`,
`M internal/envprofile/envprofile.go`,
`?? cmd/curator/profile_git_reinstall_test.go`. No file changed in this run
(the mutant below was applied and restored byte-identical, sha256
`3deb42ea2305e9cfa76ee128e9d94abf3f5ca6a0750aeb794368f0914bb76b83`
before and after).

Fresh narrow exit codes (shell: bash, `set -o pipefail`):

- `go test ./cmd/curator/ -run
  'TestProfileGitReinstallHonoursUseAndTakeover' -count=1` → exit 0, 8/8
  pass (82.9s); same with `CURATOR_CONFORMANCE_ROOT` set → exit 0 (242.5s)
- `go test ./cmd/curator/ -run
  'TestProfileInstallReinstallHonoursUseAndTakeover' -count=1` → exit 0,
  5/5 pass (24.4s, path sibling — no path behaviour change)
- `go test ./internal/envprofile/ -run
  'Reinstall|Takeover|InstallUse|Snapshot' -count=1` (root set) → exit 0
  (141.0s)
- `gofmt -l` on the three touched files → clean; `go vet
  ./cmd/curator/ ./internal/envprofile/` → exit 0

Fresh narrowing mutant (drop the activation call on the git reinstall
branch in `installLocked`): `go test ./cmd/curator/ -run
'TestProfileGitReinstallHonoursUseAndTakeover/not-current/use_with_takeover'
-count=1` → exit 1, `profile_git_reinstall_test.go:190: the retry must
report the switch` (silent no-op: `updated profile groot …` with no
switch). Killed by the named row; fix restored byte-identical, row re-run
→ exit 0.

Gate-hunt bounds (stated, not silent):

- A full `TestP*` shard of `cmd/curator` hit the 9m `go test` timeout with
  only `TestProductionExternalDepsAssignsTransportProvenanceSink` and
  `TestProductionExternalDepsFalseDrivesMirrorFetchToSink` still running —
  both from `cmd/curator/draft_transport_provenance_test.go`, a
  draft-transport story that never references `envprofile`/`installLocked`
  (verified by grep: zero hits), i.e. structurally disjoint from this leaf.
- In isolation both pass: the sink test in 2.2s; the mirror-fetch test
  failed once (248s) then passed verbose (exit 0) — flaky under shared-host
  load (concurrent agent runs, one `go test` GOROOT lock contention
  observed and cleaned up: my own stray `go test ./internal/envprofile/`
  killed; other agents' runs left untouched).
- The CI stage failed fast (10 min, exit=1, not a 30m timeout), so a hang
  is not the CI shape; the reviewer already ran this leaf's 8-row suite
  green from zsh on a mac (203.7s). No evidence points at this leaf.

Conclusion: revision 4 = revision 3 unchanged. The macos gate failure is
not attributable to this leaf on the evidence above; republishing the same
tree so the runtime reruns the landing suite once.
