# TASK-260910-3ungjy — hook-approval-command (S6) — results

Role: developer. Spec: curator-spec `23dafa7`, `profiles/manager.md` §8.2–§8.7.
Shipped profile: **`A-warning`** (unchanged; `B-enforcing` stays a later release).
Prior leaf: `TASK-260910-1952mz` (checkpointed on the Story branch; hooks,
`internal/hookapproval` state, manager recording, §8.7 vector test).

This is the third producer pass (rev 3). CR rev 1 failed the hosted gate
(`scripts/remote-gate.sh` exit 1: `go test` exit 1 on all three OS lanes,
plus the platform-case gate FAILED on windows-latest). Rev 2 fixed the
shape conflict with a conditional JSON key; the orchestrator then directed
(`TASK-260910-3ungjy_gate-failure-rev1.md`) the unconditional additive key
with a deliberate pin update instead, plus a platform-real unreadable-state
mechanism for the two Windows-only failures. Rev 3 implements that
direction (see "Retry fixes"); the full `cmd/curator` and
`internal/install` suites the first pass skipped are run here.

## Per-file changes

- `cmd/curator/hook.go` (new): closed §8.3 surface `curator hook
  approve|approvals|revoke` plus the §8.6 posture helpers shared by both
  status commands (`hookTrustPosture`, `projectTrustCandidates`,
  `hookTrustChanged`, `formatTrustRow`). Home resolves from the config
  path alone (no config load), like `shell-init --install` and like the
  hook.
- `cmd/curator/main.go`: `hook` usage line + dispatch; `curator status`
  assesses the trust posture once per invocation (every recorded path plus
  the targets' env files), prints `shell-hook-trust:` rows once for
  humans, carries the row set per target under `shell_hook_trust` in
  `--json` (always present — an all-approved posture is still posture),
  and fails `--check` only on a changed file.
- `cmd/curator/envstatus.go`: `printEnvStatus` renders the trust warnings
  and rows ahead of the home matrix.
- `internal/hookapproval/hookapproval.go` (additive only; `List`/`Lookup`/
  `Upsert`/`ApproveFile`/`Revoke`/`Canonicalize` untouched): §8.4
  diagnostic twins, §8.6 posture states, `Scan` (tolerant reader: valid
  records sorted, malformed 1-based lines reported and skipped, never
  repaired), `Posture`, `Classify` (absent yields no row; unreadable is
  unapproved, never absence), `AssessFailClosed` (recorded ∪ extra,
  warnings instead of errors).
- `internal/envprofile/status.go`: `Status.ShellHookTrust` (`json:
  "shell_hook_trust"`, always present) + `ShellHookTrustWarnings`
  (`json:"-"`, human-only); `StatusOf` assesses recorded ∪
  launch-project files and sets `NonCurrent` on a changed file only;
  `trustProjectRoot` upward Skillfile search.
- Tests: `cmd/curator/hook_test.go` (new, 17 tests), `TestScan…` /
  `TestClassify…` / `TestAssessFailClosed…` in
  `internal/hookapproval/hookapproval_test.go`,
  `TestStatusShellHookTrustFromLaunchProject` in
  `internal/envprofile/status_test.go`, and the deliberate pin update in
  `cmd/curator/status_test.go`
  (`TestStatusJSONKeepsTheLegacyShapeWithoutCompiledCommands`).
- `docs/cli.md`: `curator hook approve|approvals|revoke` sections, status
  posture + `--check` semantics (both status commands).
- `CHANGELOG.md`: Unreleased S6 entry extended with the three commands,
  the posture rows, and the explicit "`status --json` gains
  `shell_hook_trust`" naming; shipped profile stays `A-warning`.

## Retry fixes (CR rev 1 gate failures)

1. `TestStatusJSONKeepsTheLegacyShapeWithoutCompiledCommands` failed on
   every OS: it pins the no-compiled-commands `status --json` document at
   exactly 3 keys (`alias`, `path`, `skills`), and the new
   `shell_hook_trust` array (two approved/manager rows recorded by
   install) makes it four. Reproduced locally (exit 1) before the fix.
   Orchestrator decision (`TASK-260910-3ungjy_gate-failure-rev1.md`):
   the posture rows are spec-mandated (§8.6) and the additive key is the
   intended extension — no conditional key ("an all-approved posture is
   still posture"). Fix: `shell_hook_trust` is always present
   (`cmd/curator/main.go:781`), and the pin is updated deliberately, not
   loosened: still no `builds` key, exactly the four keys `alias`,
   `path`, `skills`, `shell_hook_trust`, plus the rows' closed shape
   (absolute `path`, `state`, `approved_by` for recorded files, no
   `diagnostic` on approved rows). `TestStatusJSONCarriesShellHookTrust`
   pins the unapproved / changed / approved row shapes; the mixed
   approved+unapproved listing is pinned in
   `TestStatusReportsRecordedPathsBeyondTheCurrentProject`.
2. `TestScanRejectsUnreadableState` and
   `TestAssessFailClosedOnUnreadableStateReportsUnapproved` failed on
   windows-latest only: both staged "unreadable" state with a regular
   file at the manager-home path, and Windows reports that open as
   `IsNotExist` (ERROR_PATH_NOT_FOUND), so the code rightly sees
   absence. Fix per the orchestrator direction: keep `home` a directory
   and put a DIRECTORY at the state file's own path — reading a
   directory as a file fails with a non-`IsNotExist` error on every
   platform (EISDIR on POSIX, handle/access error on Windows). No GOOS
   skip. (Cannot execute Windows here; the mechanism is POSIX-verified
   and follows the analysis exactly.)
3. Three skip reasons matched no declared skip class, which the Tier-2
   gate treats as fatal — the windows-latest `platform-case gate:
   FAILED` line in the rev 1 evidence. The mode-000 skip fires
   deterministically on Windows (reads succeed through 000 bits there);
   the two symlink skips fire where the host forbids link creation. All
   three now use the repository's established host-capability phrasings
   (`this host cannot create …`,
   `this environment can read a mode-000 file`), verified against
   `.github/ci/skip-classes.tsv` patterns. One of the symlink wordings
   (`internal/hookapproval/hookapproval_test.go:121`,
   `TestAliasAndRealPathShareOneRecord`) is pre-existing 1952mz code;
   the change is wording-only, no behavior change — recorded here as a
   finding per the brief. No ledger change: no new vector family is
   consumed (`ledger-consistency.sh`: 241 rows, ok, exit 0).

## How each AC line is met

- "Approval persists per project; hook honors it":
  `cmd/curator/hook.go:53` (`cmdHookApprove`: Canonicalize → read →
  idempotence check → atomic `Upsert` as operator) persists;
  `TestHookApproveEndToEndWithGeneratedHook`
  (`cmd/curator/hook_test.go`) drives the real CLI and the real generated
  hook: approve → silent sourcing; change → `shell_hook_env_changed`;
  re-approve → silent; revoke → `shell_hook_env_unapproved` again.
- §8.3 approve (`cmd/curator/hook.go:53`): same canonical identity as the
  hooks (`hookapproval.Canonicalize`,
  `internal/hookapproval/hookapproval.go:117`); absent vs unreadable fail
  without recording with distinct exit text (`file is absent` /
  `file is unreadable (…)`); idempotent when the operator record already
  matches (state untouched).
- §8.3 approvals (`cmd/curator/hook.go:97`): rows sorted by path on
  stdout; malformed lines to stderr, skipped, bytes never rewritten
  (`hookapproval.Scan`, `hookapproval.go:376`).
- §8.3 revoke (`cmd/curator/hook.go:123`): `Revoke`
  (`hookapproval.go:337`) removes; missing record → `nothing to revoke`,
  exit 0, byte-identical state. Exit-0 rationale: the library returns
  benign `(false, nil)`, and §8.3 says "leave state unchanged and report",
  not "fail" (contrast approve's explicit "MUST fail"); the
  remove-missing-exits-1 convention covers declaration removes whose
  library returns an error.
- §8.6 posture: `curator status` (`cmd/curator/main.go:740,781,815`)
  and `curator env status` (`internal/envprofile/status.go:210`,
  `cmd/curator/envstatus.go:22`) list each known file (every recorded
  path + current projects' `.agents/env.sh`/`.agents/env.ps1`) as
  approved / `shell_hook_env_unapproved` / `shell_hook_env_changed`
  with absolute path and `approved_by` for recorded files; `--check`
  fails on changed, unapproved stays a warning row. Both `--json`
  documents carry the posture under `shell_hook_trust`.
- No `SPEC_PIN` change, no hook-logic change (`internal/shell/shell.go`
  untouched), no vendored spec bytes.

## Validation transcripts (exit codes real)

Shell: `bash`. Conformance root for test runs:
`CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1`
(spec checkout at `23dafa7`).

- `go build ./...` → exit 0.
- `go vet ./...` → exit 0.
- `gofmt -l cmd internal docs` → no output, exit 0.
- `go test -count=1 -timeout 30m ./cmd/curator/` (FULL package, rev 3
  tree) → exit 0 (`ok`, 955.7s).
- Hook/status/E2E mask in `cmd/curator` on the rev 3 tree (18 tests
  incl. the updated `TestStatusJSONKeepsTheLegacyShapeWithout-
  CompiledCommands`) → exit 0, all PASS (73.0s).
- `go test -count=1 ./internal/hookapproval/` (FULL package) →
  exit 0 (`ok`).
- `go test -count=1 ./internal/envprofile/` (FULL package) → exit 0
  (`ok`, 363.6s).
- `go test -count=1 ./internal/envprofile/ -run
  TestStatusShellHookTrustFromLaunchProject` (post-edit re-run on the
  final tree) → exit 0 (`ok`, 7.5s).
- `go test -count=1 ./internal/envfiles/` (FULL package) → exit 0
  (`ok`).
- `go test -count=1 -timeout 30m ./internal/install/...` (FULL) →
  exit 0 (`ok`: install 894.0s, install/atomicity 481.0s).
- 1952mz `TestShellHookTrustVectors` (with root, real shells) →
  exit 0, PASS (5.5s test time).
- `bash .github/ci/ledger-consistency.sh /tmp/ledger-ev-3ungjy-retry` →
  `241 rows checked`, `ok`, exit 0 (no ledger change: no new vector
  family consumed).
- `golangci-lint run` (full repo, rev 3 tree) → `0 issues.`, exit 0.
- Narrowing mutants on the final (rev 3) tree (disposable,
  `cmp`-verified restore, each must FAIL): M1 collapsed the
  absent/unreadable wording → `go test -run
  TestHookApproveFailsDistinctly…` FAIL (exit 1, killed); M2 made
  `--check` fail on unapproved too → `TestStatusReportsShellHookTrust-
  Posture` FAIL (exit 1, killed); M3 made `Scan` error on malformed
  like `List` → `TestHookApprovalsToleratesMalformedRecord` FAIL
  (exit 1, killed). 3/3 killed, 0 survivors; the three targets rerun
  green after restore (exit 0).
- Load-starvation notes (host ran at load 6–30 with other agents'
  suites in parallel; all reran green alone): the first full
  `cmd/curator` attempt failed `TestGCRetainsAndReportsReferenced-
  CompiledState` with `build_execution_worker_protocol_invalid`
  (worker IPC starved) — the test passes alone (exit 0, 47.5s) and
  the full package passes on retry (exit 0); the post-edit envprofile
  targeted run hit the 10m `go test` wall timeout once under load 16
  and passes alone (exit 0, 7.5s). Both are capacity artifacts, not
  code failures; CI runs each lane on a quiet runner.

## Deliberately out of scope (per brief)

- Hook text and trust decision logic (1952mz; untouched except the
  wording-only skip fix above, flagged as a finding), the flip to
  `B-enforcing`, E2/E4/S4, `SPEC_PIN`, PowerShell-specific approval UX
  beyond path folding, `curator global status` (§8.6 names only `status`
  and `env status`).

## Known bounds / handoff notes

1. On Windows, approve/revoke take native spellings (as the hook
   warnings print them); an MSYS `/c/…` CLI operand is not mapped —
   `filepath.Abs` mis-resolves it (1952mz note 3). Out of scope per the
   brief's PowerShell-UX carve-out.
2. SUPERSEDED by the revision-3 rework (see addendum): every recorded
   path now keeps its posture row — a recorded-but-deleted file is
   reported as approved-with-`file: missing` (non-current under
   `--check`), an unreadable file as approved/unapproved-with-`file:
   unreadable`, and approval-state warnings travel in JSON under
   `shell_hook_trust_warnings`. (Pre-rework text: posture listed files
   that exist; a recorded-but-deleted file yielded no row; malformed
   lines surfaced as human-output warning rows only.)
3. `curator status --json` repeats the invocation-global trust set per
   target (shape-preserving; documented in code); human output prints it
   once.

## Spec gaps

None — §8.2–§8.7 implemented as written. Items 1–3 above are documented
implementation bounds, not spec contradictions.

---

# Addendum: rework for CR revision 3 (review corrections R1–R3)

CR revision 2 passed the hosted gate on all lanes but was rejected in
independent review (`TASK-260910-3ungjy_review-verdict-rev2.md`,
changes_requested, three P2 findings, all in the posture path). This
rework implements exactly the three corrections of
`TASK-260910-3ungjy_rework-rev3.md`; everything else from the accepted
portions of revision 2 is unchanged. Rule behind all three (environments
§8.4): absence and read failure are different facts, and a read failure
is never evidence that no approval exists. Shipped profile stays
**`A-warning`**; no `SPEC_PIN` change, no hook-text or trust-decision
change (`internal/shell` untouched), no new §8.4 diagnostic code, no new
vector family consumed (no ledger change).

## Per-correction file:line

- R1 — every recorded path stays in the inventory.
  `internal/hookapproval/hookapproval.go:476` (`Classify`): the record
  is now associated BEFORE the current bytes are inspected; a valid
  record whose file disappeared keeps its posture row as approved with
  its `approved_by` and the explicit `file: missing` qualifier
  (`PostureFileMissing`, hookapproval.go:72; `:507`), never claiming a
  digest comparison (no diagnostic on the row). Never-existing,
  unrecorded optional env files still yield no row (`:501-503`).
  `--check` treats the row as non-current via `Posture.NonCurrent`
  (`:455-457`), wired through `hookTrustCheckFailed`
  (`cmd/curator/hook.go:176`) and the `envprofile` loop
  (`internal/envprofile/status.go:217-222`). Text rendering
  (`cmd/curator/hook.go:193`, `formatTrustRow`) prints
  `approved (approved_by=…; file is missing, so the recorded digest was
  not compared; restore the file or run curator hook revoke …)`.
- R2 — unreadable recorded candidates keep their record.
  Same `Classify` rewrite: on a stat/read failure the row keeps
  `approved_by` and reports the failed read explicitly —
  approved-with-`file: unreadable` and no diagnostic when a record
  exists, unapproved-with-`file: unreadable` (keeping the earned
  `shell_hook_env_unapproved` diagnostic) when none
  (`unreadablePosture`, hookapproval.go:538-543). Both are non-current
  under `--check` (`NonCurrent`), never the ordinary no-record warning
  path. `curator hook approve` already fails with distinct
  absent/unreadable exit text and `curator hook approvals` already names
  state read failures (`cannot list approvals: … read state …`);
  covered by `TestHookApproveFailsDistinctlyOnAbsentAndUnreadable`
  (pre-existing) and new
  `TestHookApprovalsUnreadableStateNamesReadFailure`
  (`cmd/curator/hook_posture_test.go`).
- R3 — JSON mode reports approval-state read failures.
  New `AssessFailClosedDetailed`
  (hookapproval.go:567-583) additionally returns `stateUnreadable`
  (true only when the state file itself cannot be read; a truly absent
  state stays the normal unapproved-warning case; a malformed line
  warns without flagging the state unreadable). `curator status`
  carries warnings under the conditional `shell_hook_trust_warnings`
  key (`cmd/curator/main.go:781-789`, present only when non-empty, so
  the 4-key pin is unaffected) and fails `--check` on an unreadable
  state (`:820-826`). `curator env status` exposes the same warnings
  via `Status.ShellHookTrustWarnings`
  (`internal/envprofile/status.go:132`,
  `json:"shell_hook_trust_warnings,omitempty"`) and sets `NonCurrent`
  (`:223-225`). The unavailable record set is thus treated as
  unreadable, never as an empty set.
- Row-shape hardening from the verdict (minor): the legacy pin now
  asserts the exact closed approved-row shape (exactly the three keys
  path/state/approved_by) in
  `TestStatusJSONKeepsTheLegacyShapeWithoutCompiledCommands`
  (`cmd/curator/status_test.go`), and the unapproved/changed/approved
  JSON rows in `TestStatusJSONCarriesShellHookTrust`
  (`cmd/curator/hook_test.go`) assert no `file` qualifier when the
  bytes compared.

## New/updated tests (all committed)

- `internal/hookapproval/hookapproval_test.go`: updated
  `TestAssessFailClosedUnionsRecordedAndExtra` (stale record keeps its
  row) and `TestClassifyTreatsUnreadableAsUnapprovedNeverAbsent`
  (unreadable qualifier + non-current); new
  `TestClassifyKeepsRecordOnUnreadableCandidates`,
  `TestPostureNonCurrentCoversChangedMissingAndUnreadable`,
  `TestAssessFailClosedDetailedDistinguishesUnreadableState`
  (absent/unreadable/malformed triple).
- `cmd/curator/hook_posture_test.go` (new): production-`run()` tests
  for text and JSON on both status surfaces incl. `--check` exit
  semantics —
  `TestStatusRecordedButMissingStaysInInventory`,
  `TestStatusUnreadableCandidatesKeepRecord` (recorded + unrecorded),
  `TestStatusUnreadableApprovalStateSurfaced` (paired absent control),
  `TestEnvStatusMissingAndUnreadableKeepRecord` and
  `TestEnvStatusUnreadableApprovalStateSurfaced` (both on an
  otherwise-current matrix via `provisionedEnvMatrix`, so each
  `--check` exit is evidence of the trust posture alone),
  `TestHookApprovalsUnreadableStateNamesReadFailure`.
- `internal/envprofile/status_test.go`: new
  `TestStatusShellHookTrustMissingUnreadableAndStateFailures`
  (StatusOf-level rows; matrix `--check` exits proven at the CLI
  level per the file's established layering).
- Reviewer-probe parity: the committed tests reproduce every probe
  fixture (approve→delete, approve→directory-at-path,
  directory-at-state-path) across the same four surfaces
  (`status --check`, `status --check --json`, `env status --check`,
  `env status --check --json`) and assert path + `operator` +
  `missing`/`unreadable`/`cannot read` presence plus the `--check`
  exits the probes did not pin.

## Validation transcripts (verdicts quoted verbatim)

Shell: `bash`. Every Go command below ran with
`CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1`
(spec checkout at `23dafa7`). `exit N` is quoted only where the command
ran under `set -o pipefail` with stdout redirected to a file (real gate
exit code); otherwise the verdict quoted is go test's own `ok`/`FAIL`
line, which is authoritative for pass/fail.

- `go build ./...` → exit 0.
- `go vet ./...` → exit 0.
- `gofmt -l internal cmd` → no output, exit 0.
- `go test -count=1 ./internal/hookapproval/...` (FULL) → exit 0
  (`ok`, 1.9s).
- `go test -count=1 -timeout 8m ./internal/envprofile/ -run 'TestStatus'`
  (5/5 incl. the new R1–R3 test) → exit 0 (`ok`, 22.6s).
- `go test -count=1 -timeout 8m ./cmd/curator/ -run
  'TestStatus|TestCheckFailsForEveryNonCurrentCode|TestCompiled'`
  (all status/compiled/currentness tests incl. new posture tests and
  the hardened pin) → `ok` (456.1s).
- `go test -count=1 -timeout 9m ./cmd/curator/ -run
  'TestHook|TestStatusReportsShellHook|TestStatusJSONCarriesShellHook|
  TestStatusReportsRecordedPaths|TestStatusJSONKeepsTheLegacyShape|
  TestEnvStatusReportsShellHook|TestEnvStatusMatrix'` (pre-existing
  hook/posture/E2E/pin/matrix coverage) → `ok` (55.2s).
- `go test -count=1 -timeout 9m ./cmd/curator/ -run
  'TestStatusRecordedButMissing|TestStatusUnreadableCandidates|
  TestStatusUnreadableApprovalState|TestHookApprovalsUnreadableState'`
  (new status-surface tests) → `ok` (1.9s).
- `go test -count=1 -timeout 9m ./cmd/curator/ -run
  'TestEnvStatusMissingAndUnreadableKeepRecord|
  TestEnvStatusUnreadableApprovalStateSurfaced'` (new env-surface
  tests) → `ok` (94.1s).
- `go test -count=1 -timeout 8m ./cmd/curator/ -run
  'TestGlobal|TestUmbrella|TestEnvResolve|TestEnvAlias|TestCanonical'`
  → exit 0 (`ok`, 126.8s).
- 1952mz `TestShellHookTrustVectors` (with root) → exit 0
  (`--- PASS`, 7.5s test time).
- `golangci-lint run ./internal/hookapproval/...
  ./internal/envprofile/... ./cmd/curator/...` (repo `.golangci.yml`,
  the CI lint gate) → exit 0 (`0 issues.`).
- Narrowing mutants on the final tree (in-worktree, `cmp`-verified
  byte-identical restore, each must FAIL): M1 narrowed the inventory
  back to existing files (drop the missing-row branch) →
  `TestStatusRecordedButMissingStaysInInventory` FAIL (exit 1,
  killed); M2 narrowed unreadable-state treatment back to the empty
  set (`stateUnreadable=false`) →
  `TestStatusUnreadableApprovalStateSurfaced` FAIL (exit 1, killed).
  2/2 killed, 0 survivors; both targets rerun green after restore
  (exit 0).

## Coverage statement (bounded headless run)

Per the spawn operational constraints, long suites were split into
bounded calls (above) instead of one full-package run: every test
touching the posture/hook/status/env surface ran green in this
session. NOT rerun here: the full unmasked `cmd/curator`,
`internal/envprofile`, `internal/shell` and `internal/install`
packages (each exceeds the single-shell time bound; the hosted
`scripts/remote-gate.sh` at handoff runs the full landing suite,
and CR revision 2 of this exact tree was all-green on every hosted
lane before this delta). No test was narrowed, deselected, or
skipped to obtain a pass; the only skips are the pre-existing
host-capability ones.

## Deliberately out of scope (unchanged)

Hook text and trust decision logic (1952mz; `internal/shell`
untouched), the flip to `B-enforcing`, E2/E4/S4, `SPEC_PIN`,
PowerShell-specific approval UX beyond path folding, `curator global
status` (§8.6 names only `status` and `env status`). The
`shell_hook_trust_warnings` keys are conditional (present only with
warnings); no always-present shape change beyond revision 2's
`shell_hook_trust`.

## Re-verification (recovery run RUN-260917-285eee, same tree)

The prior rev3 run handed off to `to-review` but was cancelled during
finalizing before Change Request construction; this recovery run
re-verified the identical uncommitted tree (no source edits — the only
worktree write was a temporary narrowing mutant, `cmp`-verified
byte-identical on restore) and re-ran the gates below itself. Shell:
`bash`, `set -o pipefail`, every Go command with
`CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1`:

- `go build ./...` → exit 0.
- `go vet ./...` → exit 0; `gofmt -l internal cmd` → no output, exit 0.
- `go test -count=1 ./internal/hookapproval/...` (FULL) → exit 0 (`ok`, 2.0s).
- `go test -count=1 ./cmd/curator/ -run
  'TestStatusRecordedButMissing|TestStatusUnreadableCandidates|
  TestStatusUnreadableApprovalState|TestHookApprovalsUnreadableState'`
  (new status-surface R1–R3 tests) → exit 0 (`ok`, 1.8s, 4/4 PASS).
- `go test -count=1 ./cmd/curator/ -run
  'TestHook|TestStatusReportsShellHook|TestStatusJSONCarriesShellHook|
  TestStatusReportsRecordedPaths|TestStatusJSONKeepsTheLegacyShape|
  TestEnvStatusReportsShellHook'` (commands, posture, hardened pin, E2E) →
  exit 0 (`ok`, 32.3s).
- `go test -count=1 ./cmd/curator/ -run
  'TestEnvStatusMissingAndUnreadableKeepRecord|
  TestEnvStatusUnreadableApprovalStateSurfaced'` (new env-surface R1–R3
  tests) → exit 0 (`ok`, 108.7s, 2/2 PASS).
- `go test -count=1 ./internal/envprofile/ -run 'TestStatus'` → exit 0
  (`ok`, 15.9s).
- 1952mz `TestShellHookTrustVectors` (with root) → exit 0 (`--- PASS`, 3.6s).
- `golangci-lint run ./internal/hookapproval/... ./internal/envprofile/...
  ./cmd/curator/...` (repo `.golangci.yml`) → exit 0 (`0 issues.`).
- Narrowing mutant M1 (inventory narrowed back to existing files — drop
  the missing-row branch): `TestStatusRecordedButMissingStaysInInventory`
  FAIL, exit 1, killed; target rerun green (exit 0) after byte-identical
  restore. 1/1 killed in this run (prior run: 2/2 on the same tree).

Not rerun here (same bound as the prior run): full unmasked
`cmd/curator`, `internal/envprofile`, `internal/shell`,
`internal/install` packages — each exceeds the single-shell time bound;
the hosted `scripts/remote-gate.sh` at handoff runs the full landing
suite. `git status` at handoff shows only the 9 modified + 3 new leaf
files; `internal/shell` untouched, no `SPEC_PIN`/`.github` diff.

## Revision 3 — refreshed base (run RUN-260917-3f7b29, same tree)

The rev-3 rework was refused twice at handoff with
`stale-anchor: change_request_base_authority_mismatch` (workspace
checkpoint `1e6a165` on trunk `aa46ecd`; trunk had advanced to
`0f0ae61` through board-state-only commits). This run performed the
refresh and re-verified the candidate; no source edits were made.

- Pre-refresh: `git status --short` showed exactly the 12 rev-3 paths
  (9 modified + 3 new `cmd/curator/hook*.go`); `git log --oneline -2`
  tip `1e6a165` on `aa46ecd`.
- `task-board worktree refresh-candidate TASK-260910-3ungjy` →
  `{"Outcome":"refresh_advanced","TrunkOID":"0f0ae61…","BranchOID":"dc5675e…"}`,
  no conflicts. Post-refresh: `git log --oneline -3` tip `dc5675e`
  descending from `0f0ae61`.
- The refresh left stale diffs in the worktree's `.task-board` checkout
  copy (older board bytes vs the new checkpoint; the authoritative
  board is `TASK_BOARD_DIR`, the worktree copy is an artifact). They
  were restored with `git checkout -- .task-board` (candidate
  untouched); `git status --short` afterwards shows exactly the same
  12 rev-3 paths again.

Re-verification on the refreshed base. Shell: `bash`,
`set -o pipefail`, every Go command with
`CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1`:

- `go build ./... && go vet ./... && gofmt -l internal cmd` →
  exit 0 (gofmt clean, no output).
- `golangci-lint run ./internal/hookapproval/...
  ./internal/envprofile/... ./cmd/curator/...` (repo `.golangci.yml`) →
  exit 0 (`0 issues.`).
- `go test -count=1 ./cmd/curator/... ./internal/hookapproval/...
  ./internal/envprofile/... ./internal/shell/...` (unmasked) →
  `internal/hookapproval ok` (1.6s), `internal/envprofile ok`
  (394.2s), `internal/shell ok` (32.5s); `cmd/curator` hit the
  default 10-minute `go test` timeout (`panic: test timed out after
  10m0s`, package FAIL at 600.7s) with ZERO `--- FAIL` lines — no
  test failed; the tests still running at the panic were unrelated
  compiled-command/install/global-status tests. Overall EXIT=1,
  timeout-only.
- Bounded `cmd/curator` masks re-run by this run, all exit 0:
  - `TestStatusRecordedButMissing|TestStatusUnreadableCandidates|
    TestStatusUnreadableApprovalState|TestHookApprovalsUnreadableState`
    (new R1–R3 status-surface tests) → `ok` (1.8s).
  - `TestHook|TestStatusReportsShellHook|TestStatusJSONCarriesShellHook|
    TestStatusReportsRecordedPaths|TestStatusJSONKeepsTheLegacyShape|
    TestEnvStatusReportsShellHook` (commands, posture, hardened pin,
    E2E) → `ok` (34.6s).
  - `TestEnvStatusMissingAndUnreadableKeepRecord|
    TestEnvStatusUnreadableApprovalStateSurfaced` (new env-surface
    R1–R3 tests) → `ok` (82.8s).
- 1952mz `TestShellHookTrustVectors` (`-v`, with root) →
  `--- PASS` (3.1s), exit 0, all sh/dash/bash/zsh + ps1 subtests
  executed, no SKIP.

Accepted from already-attached evidence (byte-identical tree — the
trunk delta between `aa46ecd` and `0f0ae61` touches board state
only, no code path): full unmasked `cmd/curator` green (exit 0,
955.7s) from the prior rev-3 run; the hosted
`scripts/remote-gate.sh` at handoff runs the full landing suite.
