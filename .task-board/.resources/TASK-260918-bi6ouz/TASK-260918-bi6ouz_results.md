# TASK-260918-bi6ouz — developer results

## Implementation

- Replaced the POSIX-only `dotfileStateDirs` list with one ordered §9.5 table for chezmoi, home-manager, yadm, stow, and dotbot across macOS, Linux, and Windows, including explicit `none` cells and the spec source column.
- Applied the native home plus XDG resolution policy: only non-empty absolute XDG paths override defaults; unset, empty, and relative values use the home-relative default. On Windows the production home lookup uses `os.UserHomeDir`, which reads `USERPROFILE`.
- Presence uses `Lstat`: only an existing non-symlink directory counts. Missing and failed inspections do not match a row, and failed inspections remain unknown rather than becoming absence. The scan continues in table order after a failed inspection and names the first later row known present.
- The suspected-manager warning names the first present manager and remains warning-only. Inventory emits it only when it found an unmanaged surface, as §9.5 requires.
- Added the byte fixture, native-platform resolution and lstat tests, and the vector test that drives `UseWithPolicy` through production inventory. Added CI platform rows, troubleshooting guidance, and an Unreleased changelog entry.

## Spec and conformance evidence

- Source checkout: `/Users/administrator/Developer/ReluxWorks/curator/curator-spec`, HEAD `eadb1c06480f4438f01b2b0a973caf775a188f41`, which contains `802caee548ddc8b19408746d26c7972d39b39cc2`.
- Byte fixture: `internal/envprofile/testdata/dotfile-manager-table-802caee.md`, copied from `protocol/environments.md` at `802caee548ddc8b19408746d26c7972d39b39cc2`; SHA-256 `6575912115cf5b87facc0a6981e2d17bbcc1716f6ca7db3e0da5ddbf98853670`.
- The local spec-main vector contains 22 cases: Linux 13, macOS 6, Windows 3. On this macOS host, all 6/22 macOS cases passed through `UseWithPolicy`; Linux and Windows vector cases remain for their native lanes.
- The committed CI `SPEC_PIN` is `dced9b8317e0e8af79edf2d0539b32bd22b6c85b`, which predates this vector. The vector test skips when `CURATOR_CONFORMANCE_ROOT` is unset or when the supplied root lacks the vector, using the existing `root-content` class and its `is a pre-revision root` matcher. No new skip class or gate-selftest row was needed. The hosted gate is reserved for handoff and has not been run manually.

## Validation before Revision 2

- The revision 1 focused table/vector/takeover command passed on the macOS lane with the local spec root (6/22 cases).
- A broad `go test ./internal/envprofile -count=1` attempt exited 1 after the Go 10-minute timeout (`601.208s`) while other host Go test processes were running. The named slow test was rerun alone and passed. This broad timeout was recorded as a failure, not a pass.
- Revision 1 recorded `go vet`, formatting, package build, and `golangci-lint` evidence; lint initially found one `ineffassign`, which was fixed and rerun green.

## Revision 2 — review rework

This section supersedes the revision 1 statement above that an inspection error stops the scan. `foreignManagerHintAt` now retains the first non-ENOENT lstat error and continues through the ordered table. If a later state directory is present, it returns that manager together with the retained error; if no row is present, it returns no hint and the error, so unknown is not recorded as absence.

- F1 regression: `TestTakeoverSymlinkedDotfileManagerStateDoesNotWarn` drives `UseWithPolicy(..., Policy{Takeover: true})` with an unmanaged Claude surface and a symlink to a real directory at the resolved chezmoi state path. The production warning must remain absent. On Windows, symlink creation may skip only with the existing `host-capability` class if unprivileged creation is unavailable; Linux and macOS require the symlink fixture.
- F2 regression: `TestForeignManagerHintLstatDiscipline/failed_inspection_continues_to_later_yadm_directory` injects EACCES for chezmoi, creates a real `~/.local/share/yadm` directory, and delegates subsequent probes to `os.Lstat`. It returns `yadm` while retaining the permission error. The production warning uses this returned manager name; the production takeover warning path is exercised by `TestTakeoverWarnsDotfileHeuristic` and the conformance vectors.
- Unknown case: `TestForeignManagerHintLstatDiscipline/unreadable_only_candidate_yields_unknown_without_a_hint` injects an unreadable first candidate and absence for the others. It asserts an empty hint, retained permission error, and continued inspection of every resolvable row; no absence record is produced.
- Narrowing mutant, F1: in a disposable copy, changed the production call from `os.Lstat` to `os.Stat`. `go test ./internal/envprofile -run '^TestTakeoverSymlinkedDotfileManagerStateDoesNotWarn$' -count=1` exited **1** because the test observed `environment_foreign_manager_suspected: chezmoi`.
- Narrowing mutant, F2: in a separate disposable copy, restored the immediate return on the first non-ENOENT lstat error. `go test ./internal/envprofile -run '^TestForeignManagerHintLstatDiscipline$/failed_inspection_continues_to_later_yadm_directory$' -count=1` exited **1** because the later yadm row was not reached. Both mutant failures are expected and show that their named regression rows kill the narrowing changes.

## Revision 2 validation

All commands ran as standalone processes in zsh with `set -o pipefail` where shown; none used a pipeline.

- `set -o pipefail; env CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1 go test ./internal/envprofile -run 'Dotfile|ForeignManager|Takeover' -count=1 -v` — **exit 0**, 46.973s. Six macOS vector cases passed; table fixture, path resolution, five lstat subtests, warning suppression, and production takeover cases passed.
- `set -o pipefail; env CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1 go test -race ./internal/envprofile -run 'Dotfile|ForeignManager|Takeover' -count=1` — **exit 0**, 65.518s.
- `set -o pipefail; go test ./internal/envprofile -run '^TestForeignManagerHintLstatDiscipline$' -count=1 -v` — **exit 0** after adding the real yadm directory fixture.
- `set -o pipefail; go build -o /tmp/TASK-260918-bi6ouz-curator ./cmd/curator` — **exit 0**.
- `set -o pipefail; golangci-lint run ./internal/envprofile` — **exit 0**, 0 issues.
- `git diff --check` — **exit 0**.
- The F1 and F2 mutant commands above both exited 1 as expected in disposable copies. They are reported as killed mutants, not passing validation commands.

## Platform and handoff state

The current host is macOS. Linux and Windows lanes, including Windows table cells and the Windows symlink-capability outcome, are unverified here and remain for the hosted handoff gate. Changes remain uncommitted in the assigned Story worktree. Hosted CI is expected to run once through `task-board handoff`.

## Revision 2 — producer rerun (RUN-260923-fcc2e3)

Shell: zsh. The current worktree was verified on macOS after tightening the simulated permission failure to `os.ErrPermission`.

- `env CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1 go test ./internal/envprofile -run 'Dotfile|ForeignManager|Takeover' -count=1 -v` — **exit 0**, package reported 32.231s. The vector test executed **6 of 22** cases for macOS; fixture/table, path resolution, all five lstat discipline subtests, and takeover/repair cases passed.
- `env CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1 go test -race ./internal/envprofile -run 'Dotfile|ForeignManager|Takeover' -count=1` — **exit 0**, package reported 40.492s.
- `go build -o /tmp/TASK-260918-bi6ouz-curator ./cmd/curator` — **exit 0**.
- `golangci-lint run ./internal/envprofile` — **exit 0**, 0 issues.
- `gofmt -l internal/envprofile/managed.go internal/envprofile/takeover_test.go internal/envprofile/managed_dotfile_test.go internal/envprofile/managed_dotfile_conformance_test.go` — **exit 0**, no files reported.
- `git diff --check` — **exit 0**.

Mutant replay in disposable copies of the candidate tree (expected-red gates):

- F1 Stat mutant in `/tmp/TASK-260918-bi6ouz-r2-f1`: replaced the production `os.Lstat` argument with `os.Stat`. `go test ./internal/envprofile -run '^TestTakeoverSymlinkedDotfileManagerStateDoesNotWarn$' -count=1` — **exit 1**; the production takeover row observed `environment_foreign_manager_suspected: chezmoi`.
- F2 abort mutant in `/tmp/TASK-260918-bi6ouz-r2-f2`: restored immediate return on non-ENOENT lstat failure. `go test ./internal/envprofile -run '^TestForeignManagerHintLstatDiscipline$/failed_inspection_continues_to_later_yadm_directory$' -count=1` — **exit 1**; the row failed because the later present yadm directory was not reached.

The full `internal/envprofile` package suite was not repeated in this run. The earlier broad attempt recorded above timed out at 601.208s; this rework used the bounded behavior-focused package and race slices. Linux (13 vector cases) and Windows (3 vector cases and the Windows table cells) remain for their hosted lanes; the handoff landing gate is reserved for the single hosted run.

## Prior handoff gate attempt

The previously attached revision 2 handoff validation log records hosted run 35900187922. Ubuntu, Windows, race, lint, and platform-case checks succeeded; the macOS `go test` stage failed on `internal/install/TestBuildRootExcludedBeforeLocaleRenderingWithoutCompilerExecution`, which reported that it could not resolve its locally-created `build-skill` commit as a tree. The task-specific `internal/envprofile` rows passed; the pinned rc.12 vector absence was tolerated by the declared `root-content` class. This failure is outside the files changed for this task. I reran the named test directly in the current worktree: `go test ./internal/install -run '^TestBuildRootExcludedBeforeLocaleRenderingWithoutCompilerExecution$' -count=1 -v` — **exit 0**, 7.647s. The gate's original full log and downloaded evidence are retained as the attached validation artifact; I am leaving the suite rerun to the handoff workflow.

## Revision 4 (carry-forward republish)

Revision 3 was ACCEPTED on content (review verdict rev3, run 35911633000 green). Trunk moved to `1511b345`; the Story worktree carries the accepted delta uncommitted after `worktree converge`. No code or test changes were made in this run; content is revision 3.

Carry-forward verification (rev3 base `48da2690`, rev3 patch `TASK-260918-bi6ouz_change-request_rev3.patch`):

- Trunk touched 2 of 8 patch paths since the rev3 base: `CHANGELOG.md` (commits `b1e296ef`, `f9e0e710`) and `.github/ci/platform-cases.tsv` (commit `f9e0e710`). The other 6 paths are untouched by trunk.
- The 6 trunk-untouched paths are byte-identical to revision 3 (verified by applying the rev3 patch to a disposable worktree at `48da2690` and `cmp` against this worktree, 6/6 identical, scratch worktree removed afterwards):
  - `docs/troubleshooting.md`
  - `internal/envprofile/managed.go`
  - `internal/envprofile/takeover_test.go`
  - `internal/envprofile/managed_dotfile_test.go` (new file, untracked)
  - `internal/envprofile/managed_dotfile_conformance_test.go` (new file, untracked)
  - `internal/envprofile/testdata/dotfile-manager-table-802caee.md` (new file, untracked)
- Intersecting paths keep both sides, no conflict markers (`grep` for `<<<<<<<`/`>>>>>>>`/`^=======$` clean):
  - `CHANGELOG.md`: rev3 dotfile-manager entry present (line ~10) alongside trunk's GoReleaser rc-channel guard and directory-component-fold entries.
  - `.github/ci/platform-cases.tsv`: rev3 envprofile rows present (lines 425-430: table-fixture, path-resolution, lstat-discipline, symlinked-takeover, vectors, state-alone) alongside trunk's directory-component-fold rows and the pre-existing `TestTakeoverWarnsDotfileHeuristic` row.

Focused bounded run in this worktree (standalone process, no pipe chain):

- `go test ./internal/envprofile -run 'Dotfile|ForeignManager|Takeover' -count=1` — **exit 0** (package 18.281s).

All Definition-of-Done items were already checked at revision 3 and remain checked; no unchecked item cites this revision. Work left uncommitted in the Story worktree for handoff.
## Revision 5 (carry-forward republish)

Revision 4 was ACCEPTED on content (review verdict rev4, remote gate run 35924170007 green). Trunk moved to `a48f584c`; the Story worktree carries the accepted delta uncommitted after `worktree converge`. No code or test changes were made in this run; content is revision 4, except the CHANGELOG entry was reverted per the 2026-09-24 orchestrator CHANGELOG policy (file now equals trunk; entry text preserved below for release prep).

Carry-forward verification (rev4 base `1511b345`, rev4 candidate tree `9fe12649`, rev4 patch `TASK-260918-bi6ouz_change-request_rev4.patch`, 8 paths):

- Trunk touched 3 of 8 patch paths since the rev4 base (`git diff 1511b345..a48f584c`): `CHANGELOG.md`, `.github/ci/platform-cases.tsv`, `docs/troubleshooting.md`. The other 5 paths are untouched by trunk.
- The 5 trunk-untouched paths are byte-identical to revision 4 (`git show 9fe12649:<path>` vs worktree, `cmp` clean 5/5):
  - `internal/envprofile/managed.go`
  - `internal/envprofile/takeover_test.go`
  - `internal/envprofile/managed_dotfile_test.go` (new file, untracked)
  - `internal/envprofile/managed_dotfile_conformance_test.go` (new file, untracked)
  - `internal/envprofile/testdata/dotfile-manager-table-802caee.md` (new file, untracked)
- Intersecting paths keep both sides, no conflict markers (`grep` for `<<<<<<<`/`>>>>>>>`/`^=======$` clean across all 5 tracked files):
  - `CHANGELOG.md` (pre-revert): rev4 dotfile-manager `### Changed` hunk present (1x) alongside trunk's R5 script-worker and E4 entries; reverted afterwards per policy (see below).
  - `.github/ci/platform-cases.tsv`: all 6 rev4 envprofile rows present (1x each: TableMatchesSpecFixture, StatePathResolution, ForeignManagerHintLstatDiscipline, TakeoverSymlinkedDotfileManagerStateDoesNotWarn, Vectors, StateAloneDoesNotWarn) alongside trunk's godriver/scriptpolicy/scriptworker rows (e.g. TestWorkerRejectsMissingPrivateJobHandle, TestEnforcedShapesProceedPastAdmission, TestMandatoryControlsMatchTheSuite, 1x each).
  - `docs/troubleshooting.md`: rev4 `## Dotfile-manager onboarding warning` section (line 3) alongside trunk's `## Enforced script execution diagnostics` section (line 583).
- No stray root `TASK-*/BUG-*.md`, `test/`, or `ledger/` paths (`ls` confirms absent).

## CHANGELOG entry (for release prep)

The following entry was reverted from `CHANGELOG.md` per the orchestrator 2026-09-24 policy (file now equals trunk `a48f584c`); the release-prep leaf writes all entries:

### Changed

- The takeover warning environment_foreign_manager_suspected now follows
  the closed environments §9.5 dotfile-manager table in spec order:
  chezmoi, home-manager, yadm, stow, dotbot across macOS, Linux, and
  Windows, including explicit none cells. State paths resolve through
  absolute XDG overrides or the specified home defaults, and presence
  requires a non-symlink directory found with lstat; failed inspection
  leaves that row unknown while scanning later rows and does not become
  absence (§8.4.1). The warning names the first present manager and remains
  best-effort.

Focused bounded run in this worktree (standalone process):
- `go test ./internal/envprofile/... -count=1` — **exit 1**, FAIL after 600.863s (Go default 10-minute test timeout; tail shows a `testing` cleanup/`os.RemoveAll` stack; same signature as the revision 2 recorded broad attempt at 601.208s). Reported as failing; not presented as passing.
- `go test ./internal/envprofile -run 'Dotfile|ForeignManager|Takeover' -count=1 -v` — **exit 0**, 38.254s. Table fixture, path resolution, all lstat discipline subtests, state-alone, symlinked-takeover, and production takeover cases PASS; `TestDotfileManagerVectors` SKIP (no `CURATOR_CONFORMANCE_ROOT` set; declared `root-content` class).
- `go vet ./internal/envprofile` — **exit 0**. `gofmt -l internal/envprofile/` and `git diff --check` — **exit 0**, clean.
- Go content is byte-identical to revision 4 (5/5 `cmp` clean against candidate tree `9fe12649`), which passed the hosted remote gate (run 35924170007, success); the only post-acceptance edit is the CHANGELOG revert (non-Go) plus trunk's own merged rows.

All Definition-of-Done items were already checked at revisions 3–4 and remain checked; no unchecked item cites this revision. Work left uncommitted in the Story worktree for handoff (4 modified + 3 new paths after the CHANGELOG revert).

## Revision 5 — confirmation rerun (RUN-260924-553d6d)

Second developer run under the same carry-5 instruction (trunk `a48f584c`). Re-verified from scratch; no source, test, doc, TSV, or CHANGELOG edits made in this run.

- Trunk touch since rev4 base `1511b345`: 3 of 8 patch paths (`CHANGELOG.md`, `.github/ci/platform-cases.tsv`, `docs/troubleshooting.md`); other 5 untouched.
- 5 trunk-untouched paths byte-identical to rev4 candidate tree `9fe12649` (`cmp` clean 5/5): `internal/envprofile/managed.go`, `internal/envprofile/takeover_test.go`, `internal/envprofile/managed_dotfile_test.go` (new), `internal/envprofile/managed_dotfile_conformance_test.go` (new), `internal/envprofile/testdata/dotfile-manager-table-802caee.md` (new).
- Intersecting paths keep both sides, no conflict markers: all 6 rev4 TSV rows present 1x each alongside trunk rows; `docs/troubleshooting.md` keeps both the rev4 dotfile-manager section (line 3) and trunk's enforced-script section (line 583); `CHANGELOG.md` equals trunk (`git diff HEAD -- CHANGELOG.md` empty, entry text preserved in the "## CHANGELOG entry (for release prep)" section above).
- No stray root `TASK-*`/`BUG-*`, `test/`, or `ledger/` paths.
- Fresh bounded validation in this worktree (standalone processes, no pipes): `go test ./internal/envprofile -run 'Dotfile|ForeignManager|Takeover' -count=1` — **exit 0** (72.643s); `go vet ./internal/envprofile` — **exit 0**; `gofmt -l internal/envprofile/` clean and `git diff --check` clean — **exit 0**.
- Full-package `go test ./internal/envprofile/... -count=1` was not rerun here: the section above records its **exit 1** after 600.863s (Go 10-minute timeout, same signature as revision 2). Linux/Windows lanes remain for the hosted handoff gate.
- All 12 DoD checklist items remain checked; no unchecked item cites this revision. Work left uncommitted in the Story worktree for handoff.

## Revision 6 (carry-forward republish)

Developer run under the carry-6 instruction (trunk now `0a628621`; rev5 base `a48f584c`, rev5 candidate tree `914478b1` per the rev5 verdict). No repo file edits made in this run — the converged worktree already carried the accepted delta; verification only.

Per-path verification against `TASK-260918-bi6ouz_change-request_rev5.patch` (7 paths; `git patch-id --stable` per file, rev5 vs worktree diff):

- `internal/envprofile/managed.go` — trunk untouched (`git log a48f584c..HEAD` empty for this path). Patch-id `2597ae9e…` identical on both sides: byte-identical to rev5.
- `internal/envprofile/takeover_test.go` — trunk untouched. Patch-id `e044e103…` identical: byte-identical to rev5.
- `internal/envprofile/managed_dotfile_test.go` (new), `internal/envprofile/managed_dotfile_conformance_test.go` (new), `internal/envprofile/testdata/dotfile-manager-table-802caee.md` (new) — trunk untouched; `cmp` clean against rev5 post-images reconstructed from the patch (shas `e90a1b9e…`, `9a26c07f…`, `65759121…` match on both sides).
- `.github/ci/platform-cases.tsv` — intersecting (trunk added 1 row since rev5 base). Patch-id `e202806c…` identical; worktree keeps both sides with no conflict markers: all 6 rev5 envprofile rows present 1x each (TableMatchesSpecFixture, StatePathResolution, ForeignManagerHintLstatDiscipline, TakeoverSymlinkedDotfileManagerStateDoesNotWarn, Vectors, StateAloneDoesNotWarn) alongside trunk's `TestExecutableIdentityCasesAtProductionEntry` row.
- `docs/troubleshooting.md` — intersecting (trunk edited elsewhere since rev5 base). Patch-id `d9f573f5…` identical; both sides present with no conflict markers: rev5 `## Dotfile-manager onboarding warning` section (line 3) alongside trunk's `## Skillfile source and transport diagnostics` section (line 384).
- `CHANGELOG.md` — equals trunk (`git status --porcelain -- CHANGELOG.md` empty, as required by the 2026-09-24 policy; rev5 carries no CHANGELOG hunk). Entry text preserved verbatim in the "## CHANGELOG entry (for release prep)" section above.
- No stray root `TASK-*`/`BUG-*`, `test/`, or `ledger/` paths (`ls` confirms absent).

Worktree freshness (step 4): `git diff --name-only HEAD -- . ':!.task-board'` lists exactly the 4 tracked rev5 paths (`.github/ci/platform-cases.tsv`, `docs/troubleshooting.md`, `internal/envprofile/managed.go`, `internal/envprofile/takeover_test.go`); the remaining 3 rev5 paths are untracked new files verified identical above. `Makefile`, `README.md`, and `CHANGELOG.md` are clean — no trunk revert. Not a stale snapshot.

Fresh bounded validation in this worktree (standalone processes, exit codes preserved without pipes):

- `go test ./internal/envprofile -run 'Dotfile|ForeignManager|Takeover' -count=1 -v` — **exit 0**, `ok 45.210s`. All listed cases PASS (StatePathResolution incl. all 9 subtests, LstatDiscipline incl. all 5 subtests, StateAloneDoesNotWarn, UseWithoutTakeover, UseTakeoverBacksUp, TakeoverWarnsDotfileHeuristic, TakeoverSymlinked, RepairTakeover).
- `go test ./internal/envprofile -run 'TestDotfileManagerTableMatchesSpecFixture|TestDotfileManagerVectors' -count=1 -v` — **exit 0**, `ok 1.083s`. TableMatchesSpecFixture PASS; Vectors SKIP (`CURATOR_CONFORMANCE_ROOT` unset; declared `root-content` class).
- `go vet ./internal/envprofile` — **exit 0**. `gofmt -l internal/envprofile/` clean and `git diff --check` clean — **exit 0**.
- Full-package `go test ./internal/envprofile/... -count=1` was not run to completion here: the package exceeds the single headless-call bound (rev4 recorded **exit 1** after 600.863s on the Go default 10-minute timeout); rev5 passed the hosted remote gate (run 36016682383, success, incl. mac/ubuntu/windows lanes), and this revision changes zero Go bytes versus rev5. Linux/Windows lanes remain for the handoff gate.
- All 12 DoD checklist items remain checked; no unchecked item cites this revision. Work left uncommitted in the Story worktree for handoff.

## Revision 6 — autonomous recovery (RUN-260926-d61388)

Successor run after Change Request CR-TASK-260918-bi6ouz-6 revision 6
validation failed: remote gate run 36205109346 (trunk `0a628621`) reported
`Test (ubuntu-latest): failure` and `Race (ubuntu-latest): failure` with
`test-gate: go test exit=1, platform-case gate exit=0`. The prior Revision 6
section above verified the carry; this section fixes the gate failure. No
production file changed: `managed.go`, `takeover_test.go`, and the
`dotfile-manager-table-802caee.md` fixture remain byte-identical to the
accepted rev5 tree `914478b1` (`cmp` clean, re-verified after the fix).

Root cause (from the gate's own evidence artifact
`test-evidence-ubuntu-latest`, `test/go-test-served.json`): the only failing
events are `TestDotfileManagerVectors/home-manager-linux-unreadable-quiet`
(and its parent), with `takeover blocked: profile_use_partial`. Every other
task row on that lane is `ok`, including all remaining vector cases. The
vector file is new at the gate's SPEC_PIN `dcc7f01` (absent at the rev5 pin
`dced9b8`, where the vector test skipped as a pre-revision root), so this
case never executed before. Mechanism, reproduced on this host with a scratch
test (since removed): the case pins `XDG_CONFIG_HOME` empty and chmods
`<operatorHome>/.config` to 0 for the unreadable home-manager state, but the
opencode adapter falls back to `$HOME/.config/opencode` when
`XDG_CONFIG_HOME` is empty (`NativeHome`, switch.go), so its entry fails with
`mkdir .../.config/opencode: permission denied` and the machine switch
reports `profile_use_partial`. The hint itself was already correct (failed
inspection stays unknown, no warning); the casualty is the adapter entry, not
the heuristic. The collision is topological — hint and adapter read the same
live variable — so the harness, not production, is fixed.

Fix (harness only, 2 files):

- `internal/envprofile/managed_dotfile_conformance_test.go`: new
  `isolateLiveXDGFromUnreadableFixture`, called by
  `runDotfileManagerVectorCase` after fixture setup. When an `unreadable`
  state sits under the `XDG_CONFIG_HOME` root (the only XDG root an adapter
  derives its home from), the live process variable moves to a writable
  sandbox for the switch; fixtures, resolved-path assertions, and the live
  hint's other rows keep the vector-mapped environment. The quiet
  expectation is identical whether that row reads absent or failed-unknown,
  and the failed-inspection-continues rule stays covered by
  `TestForeignManagerHintLstatDiscipline` with injected errors.
- Same file: new regression test
  `TestUnreadableDotfileStateKeepsTakeoverQuiet`, which builds the
  unreadable home-manager fixture exactly like the vector harness and
  asserts the machine takeover succeeds with no
  `environment_foreign_manager_suspected` warning. Portable: skips with the
  existing guards where the artefact cannot be built, and its Windows skip
  reason classifies as `host-capability`.
- `.github/ci/platform-cases.tsv`: one ledger row for the new test
  (`linux,darwin` must-run, `windows` skip `host-capability`), keeping
  `ledger-consistency.sh` green (363 rows, exit 0).

Negative evidence (disposable in-place mutants, file restored byte-identical
after each, `cmp` verified):

- Delete mutant (isolation call replaced by a discard): new test exits 1
  with `profile_use_partial` — the exact gate signature.
- Narrowing mutant (predicate `XDG_CONFIG_HOME` → `XDG_DATA_HOME`):
  exits 1 with `profile_use_partial`, proving the test covers the sharing
  class, not just the call's existence.

Validation in this worktree (standalone processes, real exit codes):

- `go test ./internal/envprofile -run 'Dotfile|ForeignManager|Takeover|UnreadableDotfile' -count=1` — exit 0 (`ok 17.409s` on the final state; earlier full-mask run `ok 29.992s` with all PASS).
- `CURATOR_CONFORMANCE_ROOT=<curator-spec@dcc7f01>/conformance/v1 go test ./internal/envprofile -run 'TestDotfileManagerVectors|TestDotfileManagerTableMatchesSpecFixture' -count=1 -v` — exit 0; all 6 macOS vector cases PASS against the same pin that failed the gate, table fixture matches. The Linux-only unreadable case remains for the hosted linux lane; the regression test exercises its identical mechanism here (linux and macOS home-manager cells share the `XDG_CONFIG_HOME` shape).
- `bash .github/ci/ledger-consistency.sh` — exit 0 (`ledger-consistency: ok`).
- `go vet ./internal/envprofile` — exit 0. `gofmt -l internal/envprofile/` clean, `git diff --check` clean — exit 0.
- Full-package `go test ./internal/envprofile/... -count=1` not rerun to completion: exceeds the single headless-call bound (recorded exit 1 at the Go 10-minute timeout in rev2/rev5 evidence); the hosted handoff gate runs it on all lanes.
- `CHANGELOG.md` still equals trunk; no stray root `TASK-*`/`BUG-*`, `test/`, or `ledger/` paths. Worktree file set is still exactly the 7 rev5 patch paths (4 tracked + 3 new); only two of them changed in this run, both inside this task's scope.

All 12 DoD checklist items remain checked. Work left uncommitted in the Story worktree for handoff.
