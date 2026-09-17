# TASK-260910-1952mz — manager-hook-digest-pin (S6) — results

Role: developer. Spec: curator-spec `0da4020`, `profiles/manager.md` §8.1–§8.7.
Shipped profile: **`A-warning`** (warning release; `B-enforcing` follows in a later release).

## Per-file changes

- `internal/hookapproval/hookapproval.go` (new): manager-home approval state.
  Closed record `{path, sha256, approved_by, approved_at}` in
  `<manager-home>/hook-approvals.tsv` (TAB-separated, no header, sorted by
  path — a format the emitted hooks parse with `awk` / `.Split()`; documented
  in the package doc). `Canonicalize` (absolute + Clean, lexical only),
  `List`/`Lookup`/`Upsert`/`ApproveFile`/`Revoke`, atomic same-directory
  temp-file + rename writes, malformed/unreadable state fails closed (never
  an empty set). Never reads package/project/profile data.
- `internal/hookapproval/hookapproval_test.go` (new): digest vector,
  round-trip + sorted-file bytes, two-spellings-one-record, revoke-absent
  no-op, `ApproveFile` refusal without recording, open-shape rejections
  (short/uppercase/non-hex digest, forged approver, missing timestamp,
  separator in path), malformed-state fail-closed.
- `internal/shell/shell.go`: `TrustProfile` (`A-warning`/`B-enforcing`,
  closed) with `DefaultTrustProfile = A-warning`; `HookWithProfile` /
  `InstallHookWithProfile` (one generation-time option; `Hook`/`InstallHook`
  keep their signatures and default to A); closed diagnostics
  `shell_hook_env_unapproved` / `shell_hook_env_changed` (exact spelling).
  Both `posixHook` and `powershellHook` now compute the candidate SHA-256
  (`shasum -a 256` / `sha256sum` fallback; `Get-FileHash`), look up the
  manager-home record for the absolute path, source trusted bytes silently,
  and otherwise warn once per shell session (newline-framed marker variable)
  naming the absolute path and `curator hook approve <path>` while keeping
  activation green. A sources unapproved/changed with the warning; B refuses
  (unsets the activation guard so a later approval takes effect without
  leaving). Fail-closed: undigestable candidate, unreadable state, or missing
  digest tooling warns and skips sourcing under both profiles. Upward search,
  PATH save/restore, activation guard, Git-Bash config normalization, and the
  no-`dirname` rule are unchanged.
- `internal/envfiles/envfiles.go`: `WriteProject(projectRoot, home)` and
  `WriteGlobal(home)` record both digests as `approved_by: manager` at write
  time via `recordLive` (hashes the live bytes just written). The trust gate
  covers project files only (§8.1 closed scope); global records are inert for
  the gate and exist so every manager-written env file carries a digest.
- `internal/install/commit.go`: `runCommit` calls `recordPublishedEnvApprovals`
  after the journal commit, still under the home lock — every `ClassEnvFile`
  target the commit published is hashed live and recorded as manager. A
  recording failure warns (never fails the committed install). Dry-run never
  reaches `runCommit` (returns at install step 18), so planning stays pure.
- `internal/shell/shell_hook_trust_test.go` (new): `TestShellHookTrustVectors`
  (§8.7 recipe) and `TestShellHookRefusesHostileCheckout` (AC).
- `internal/shell/shell_test.go`: fixtures now use manager-recorded writes +
  `CURATOR_CONFIG` (nested-switch test additionally asserts trusted-silent);
  zsh reentrancy test seeds an operator approval; new `TestHookTrustGateShape`
  (default is A, open profile/shell rejected, every hook embeds the state
  filename, both diagnostics, the approve command, and exactly one profile).
- `internal/envfiles/envfiles_test.go`: new write signatures; both tests
  assert the manager record digest equals the live bytes.
- `internal/install/install_test.go`: `TestProjectInstallRecordsShellHookApprovals`
  drives the real `Project` entry point and asserts manager digests for both
  published project files; `TestProjectDryRunRecordsNoShellHookApprovals`
  asserts dry-run writes no approval state.
- `.github/ci/platform-cases.tsv`: parent row (all platforms, `root-content`
  skip tolerated, like `TestCandidateGoV1SourceAwareContract`) plus a `/*`
  row tolerating `host-capability` subtest skips (absent sh/bash/zsh or pwsh,
  or no native-spelling POSIX session on Windows).
- `CHANGELOG.md`: Unreleased S6 warning-release entry with migration hint and
  the B-follows note.

## How each AC line is met

- "Hook refuses foreign `.agents/env.sh` bytes" — B-enforcing hook refuses to
  source: `internal/shell/shell.go:306` (`_curator_trust_allow`), wired at
  `:394`; PowerShell twin at `:485`/`:538`.
- "tests simulate a hostile project checkout" —
  `TestShellHookRefusesHostileCheckout`
  (`internal/shell/shell_hook_trust_test.go:337`): hostile payload with a
  canary file + marker var; B asserts canary absent / marker unset / exactly
  one first-activation-only unapproved warning; A asserts sourced + warned.
- Digest pinning + warn-first + diagnostics + forged-record ignore — §8.7
  vector execution `TestShellHookTrustVectors`
  (`internal/shell/shell_hook_trust_test.go:69`): all 14 cases (approved /
  unapproved / changed × sh/ps1 × A/B + 2 forged-project-record) with
  two-activation sourced/diagnostic/warning-count assertions, `root-content`
  skip when the root lacks the vector, ledger rows as above.
- Manager-recorded digests — `internal/envfiles/envfiles.go:80,100,116` and
  `internal/install/commit.go:670,680`, proven by `TestWriteProject…` /
  `TestWriteGlobal` record assertions and
  `TestProjectInstallRecordsShellHookApprovals` through `install.Project`.

## Validation transcripts (exit codes real, `set -o pipefail`)

Shell for all commands: `bash`. Conformance root for test runs:
`CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1`.

- `go build ./...` → exit 0.
- `go vet ./...` → exit 0.
- `gofmt -l internal/ cmd/` → no output, exit 0 (two new files reformatted
  once with `gofmt -w`; pre-existing `.task-board/.resources` hits untouched).
- `go test -count=1 ./internal/hookapproval/... ./internal/envfiles/...` →
  exit 0 (`ok` both).
- `go test ./internal/shell/...` (with root) → exit 0. Focused run:
  `TestShellHookTrustVectors` 8/8 POSIX subtests PASS, 6/6 ps1 subtests SKIP
  (`PowerShell is unavailable (no pwsh on this runner)`, host-capability);
  `TestShellHookRefusesHostileCheckout/posix` PASS, `/powershell` SKIP (same
  reason). Full package run also green.
- `go test -count=1 ./internal/install/ -run
  'TestProjectInstallRecordsShellHookApprovals|TestProjectDryRunRecordsNoShellHookApprovals'`
  → exit 0 (both PASS).
- `go test -count=1 ./cmd/curator/...` → exit 0 via bounded `-run` chunks
  (163 tests; single full-package run exceeds the shell yield, so chunked per
  the headless-run rule; every chunk `ok`):
  `TestS` (incl. `TestShellInitPrintsHooks`,
  `TestShellInitInstallCachesHookWithoutConfig`) 86s;
  `TestP` 151s; `Test(Classify|Currentness|CheckFails|CLI)` 89s;
  `TestConfig` 32s; `TestCompiledProjectStatusAndUntrustedRecovery|
  TestCompiledProjectRepairsCorruptCompiledState` 329s;
  `TestCompiledProjectRestoresCacheWhenCommitFails|
  TestCompiledInstallFollowsTheNativeControlInventoryExactly` 578s;
  `Test[ABDEG]` 447s; `Test[HIMNORTU]` 8.5s.
  Two chunk attempts failed environmentally and were rerun green (`TestMain`:
  host GOROOT test-lock 5-minute deadline while another agent's run held the
  shared-host lock; rerun after the holder exited) — no test assertion failed
  in any chunk; all 163 `cmd/curator` tests observed PASS across the chunks.
- `golangci-lint run ./internal/hookapproval/... ./internal/shell/...
  ./internal/envfiles/... ./internal/install/...` → `0 issues.`, exit 0.
- `bash .github/ci/ledger-consistency.sh` (with root) → `241 rows checked`,
  `ok`, exit 0; new rows resolve on all three GOOS.

## Deliberately out of scope (per brief)

- `curator hook approve|approvals|revoke` and the `curator status` /
  `curator env status` posture rows (sibling TASK-260910-3ungjy). The
  `hookapproval` API (`Upsert`/`ApproveFile`/`Revoke`/`List`) is shaped for it.
- E2/E4/S4 findings, SPEC_PIN (untouched), ax integration, proposals
  0014–0018, tags/releases.
- PowerShell execution where `pwsh` is absent (legitimate skip, no stub).

## Known bounds / handoff notes for the sibling + reviewers

1. PowerShell hook verified by close review + CI, not locally: no `pwsh` is
   obtainable on this host (Homebrew `powershell` cask missing, preview cask
   fails on this Tier-3 host). The ps1 vector subtests + hostile ps1 subtest
   execute on Windows CI (pwsh preinstalled); POSIX subtests skip there by
   design. One real bug was caught and fixed in review (`[char] + $null`
   first-use concatenation → now `[string][char]10`).
2. Canonicalization is lexical only (absolute + Clean, no symlink resolution):
   symlink-distinct spellings are distinct keys (fail-closed). Rationale:
   identical rule on the Go and shell sides; the warning names the exact path
   to approve, so a miss self-heals.
3. Git-Bash-on-Windows spelling: the POSIX hook looks up the candidate
   as-found (`/c/...` form), while Go canonicalizes native (`C:\...`) form.
   Under A-warning this only warns; the sibling approve command should consider
   recording the spelling the warning names (a `/c/...` CLI path needs
   care — `filepath.Abs` on Windows mis-resolves it).
4. Unreadable-state / undigestable-candidate edges fail closed (warn as
   `shell_hook_env_unapproved`, skip sourcing under both profiles) per the
   "unreadable is never absence" rule; no vector covers them.
5. `WriteGlobal` records global env files although the §8.1 gate covers
   project files only (DoD "whenever it writes" read literally); the records
   are inert for the gate.

## Spec gaps

None — §8.1–§8.7 implemented as written. Items 2–4 above are documented
implementation bounds, not spec contradictions.

---

## Revision 2 — CR rev1 validation repair (RUN-260917-abbfd5)

CR rev1 failed `scripts/remote-gate.sh` (run 35174707552): `go test` exit 1
on Test ubuntu / Race ubuntu / Test windows, green on macOS. The gate-evidence
artifacts name exactly one failing test,
`internal/shell TestShellHookRefusesHostileCheckout`, with two distinct
harness-only root causes (no production-code change in this revision):

1. **Ubuntu + Race ubuntu, `/posix`**: `sh` activation died with
   `/usr/bin/sh: 183: hook: Syntax error: "(" unexpected (expecting "fi")`.
   Line 183 of the generated hook is the pre-existing bash-only
   `PROMPT_COMMAND=(...)` array integration. The harness resolved bare `sh`
   (dash on Ubuntu) but generated the bash hook flavor — `Hook()` only admits
   `bash`/`zsh`/`powershell`, and every pre-existing execution test resolves
   `bash`/`zsh` explicitly, so production never sources this hook under dash.
   Fix (`shell_hook_trust_test.go`): `lookPathTrustShell` resolves `bash`
   then `zsh`, never `sh`; the dead `case "sh"` args branch in
   `runTrustShell` is removed.
2. **Windows, `/posix`**: `warning count = 0, want 1` — the subtest ran under
   Git Bash, which reports the `/c/...` spelling while the assertion expects
   the native `C:\...` spelling. Fix: skip on Windows with the same
   host-capability reason the vector POSIX runner already uses
   (`this host cannot create a POSIX shell session with native path
   spellings ...`); mechanically verified to classify as
   `host-capability`/`allow`, and Tier 2 admits it without a ledger row
   because the hostile test is not a ledger case.
3. **Windows, `/powershell`**: `stderr lacks the activation probe` — pwsh on
   Windows frames stderr with CRLF, so the `"PROBE-1\n"` split missed.
   Fix: both assert helpers (`assertTrustOutcome`, `assertHostileWarning`)
   normalize `\r\n` → `\n` on entry (also hardens the future ps1 vector
   cases on Windows CI).

Also: the `platform-cases.tsv` S6 comment now says `no bash/zsh` (was
`no sh/bash/zsh`); no ledger row added or changed otherwise.

### Rev2 validation transcripts (exit codes real, `set -o pipefail`, bash)

Conformance root:
`CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1`.

- `gofmt -l internal/shell/ .github/ci` → no output, exit 0.
- `go build ./...` → exit 0.
- `go vet ./internal/shell/... ./internal/hookapproval/...` → exit 0.
- `go test -count=1 ./internal/shell/... ./internal/hookapproval/...
  ./internal/envfiles/...` → exit 0 (`ok` all three; shell 6.7s).
- Focused: `TestShellHookTrustVectors` 8/8 POSIX subtests PASS (under bash
  3.2), 6/6 ps1 subtests SKIP (no pwsh on host);
  `TestShellHookRefusesHostileCheckout/posix` PASS, `/powershell` SKIP.
- Throwaway CRLF probe (in-package, run once, then deleted): both assert
  helpers accept the exact CRLF-framed Windows stderr bytes → PASS, exit 0.
- `golangci-lint run ./internal/shell/...` → `0 issues.`, exit 0.
- `bash .github/ci/ledger-consistency.sh /tmp/ledger-ev` → `241 rows
  checked`, `ok`, exit 0.
- Not rerun in rev2: `./internal/install/...` and `./cmd/curator/...`
  (untouched by this revision — test-only change in one file; rev1 chunks
  all green, and the handoff remote-gate reruns the full suite including
  the ubuntu/windows lanes that reproduce the original failures).

Shipped profile unchanged: **`A-warning`** (B implemented + vector-covered;
flip is a later release).
## Revision 3 — review-verdict-rev2 rework R1–R4

Verdict `TASK-260910-1952mz_review-verdict-rev2.md` (changes_requested, route
to-dev) is the authority for this round; rework brief
`TASK-260910-1952mz_rework-rev3.md` scopes the four corrections. Everything
the verdict marks PASS stays; shipped profile stays **`A-warning`**.

### R2 (P1) — POSIX hook repaired, gate restored (was: weakened test)

- `internal/shell/shell.go`: the emitted POSIX hook is strictly
  POSIX-parseable again. Bash array `PROMPT_COMMAND` integration moved
  behind a `BASH_VERSION`-guarded `eval` (`shell.go:204`), so `sh`/`dash`
  never parse array syntax; both `${cfg//\\//}` substitutions replaced
  with `tr '\\' '/'` (`_curator_global_env_file`,
  `_curator_hook_approvals_file`); `_curator_trust_digest` gains the
  `openssl dgst -sha256` fallback (`shell.go:304`). No arrays, `[[ ]]`,
  `function`, process substitution, or `dirname` in the emitted hook
  (verified by grep over generated bytes).
- `internal/shell/shell_hook_trust_test.go`: `posixTrustShells` resolves
  `sh`, `dash` (when present), `bash`, `zsh` (deduped by resolved binary)
  and every POSIX trust execution — vectors, hostile, malformed, symlink —
  runs under each (`runPosixTrustActivation`, `runTrustShell` with `sh`/
  `dash` args + `$ENV` cleared). New `TestGeneratedHookParsesUnderPOSIXShells`
  (`:569`): 8 hook variants (bash/zsh × A/B × global on/off) under each
  available `sh`/`dash`/`bash`/`zsh -n`, skipping only absent interpreters.
- Evidence: before, `/bin/dash -n` on the generated hook exited 2
  (`Syntax error: "(" unexpected`, generated line 183, reproduced); after,
  all 16 `dash`/`sh`/`bash`/`zsh -n` checks exit 0. Dash functional probes:
  approved+B sourced silent, unapproved+B refused+warned,
  unapproved+A sourced+warned, changed+B refused with
  `shell_hook_env_changed`, alias-dir+B sourced silent, bash array and
  string `PROMPT_COMMAND` behavior preserved.
- Adjacent hardening in the rewritten integration: the eval'd array branch
  drops `nounset` around the array expansion and restores it (macOS bash
  3.2 aborts on an empty `"${PROMPT_COMMAND[@]}"` under `set -u` —
  reproduced, exit 127 — while the hook must stay sourceable from
  `set -u` profiles). Covered by
  `TestBashHookToleratesEmptyPromptCommandArrayUnderNounset`
  (`shell_test.go`, asserts single entry + `u` still in `$-`).
- Carried bound (verdict asked, rework brief did not scope): the
  Windows-only POSIX skips stay. They are worded as host capability, the
  rework brief scopes R2 to the sh/dash repair + coverage, and
  MSYS/native identity bridging cannot be verified from this macOS host;
  the hosted Windows lane still executes the full PowerShell matrix.
  Flag for the orchestrator whether a Windows-capable round should lift
  them.

### R1 (P1) — hooks validate the closed record, not just path+digest

- `internal/hookapproval/hookapproval.go:46-60`: new shared grammar
  constants `RecordFieldCount`, `SHA256HexLength`, `TimestampShape` (POSIX
  ERE, also valid RE2/.NET); `Validate`/`List` use them.
- `internal/shell/shell.go:518` `expandTrustTokens` renders the emitted
  hooks from those constants. POSIX `_curator_trust_recorded` (`:329`)
  accepts a line only with exact field count, 64-char lowercase hex
  digest, `approved_by` in the closed set, and an RFC 3339 timestamp
  passing shape + range checks (month/day/leap-year/hour/min/sec/offset);
  non-matching and invalid lines are skipped, unreadable/non-regular
  state stays return-2 (refuse under both profiles). PowerShell
  `Get-CuratorTrustRecorded` (`:597`) enforces the same four checks
  (shape + `DateTimeOffset.TryParse` for ranges). Malformed candidate
  lines behave as unapproved (no valid record authorizes). Both `allow`
  gates now default an unset profile variable to the baked profile
  instead of degrading B to A.
- Tests: `TestShellHookRejectsMalformedRecords` (`shell_hook_trust_test.go:665`,
  14 variants — field counts, forged/empty/uppercase approver,
  upper/short/non-hex digest, 6 timestamp classes — × POSIX shells × A/B
  plus pwsh where present; the Go reader must reject every seeded line,
  proving reader/hook agreement); `TestHookEmbedsClosedRecordGrammar`
  (`:627`, hook text carries the constants verbatim, no `__CURATOR_`
  placeholder survives); `TestTimestampShapeMirrorsRFC3339`
  (`hookapproval_test.go:281`, shape accepts everything Go accepts and
  rejects structural garbage; range corners documented as covered by the
  hook range checks + end-to-end cases).
- The verdict's three B-enforcing probes (path/digest-only line,
  `approved_by=project`, invalid timestamp) were re-run against the new
  hook under dash: all refused with `shell_hook_env_unapproved`, and all
  three classes are committed negative cases.

### R3 (P2) — one record for two spellings of one file

- `Canonicalize` (`hookapproval.go:84`) now resolves `Abs` +
  `EvalSymlinks`, falling back to longest-existing-ancestor resolution so
  not-yet-existing paths still record; package doc updated (no longer
  lexical-only).
- Hooks canonicalize before lookup and warn on the canonical path:
  POSIX `_curator_trust_canonical` (`shell.go:259`, `cd -P`/`pwd -P` +
  bounded `readlink` loop) and PowerShell `Get-CuratorTrustCanonical`
  (`shell.go:571`, `GetFullPath` + `.Target` loop).
- Tests: `TestAliasAndRealPathShareOneRecord`
  (`hookapproval_test.go:108`, record via alias, observe via real, one
  record, revoke via either); `TestShellHookTrustResolvesSymlinkedProject`
  (`shell_hook_trust_test.go:820`, activation from the alias dir sources
  silently under A and B, POSIX shells + pwsh). Existing round-trip /
  two-spellings tests updated to resolved-tempdir expectations.

### R4 (P2) — failed publication preserves the last valid state

- `writeRecords` (`hookapproval.go:298`) is one `renameFile` onto the
  target (POSIX atomic rename; `os.Rename` replaces on Windows); the
  remove-and-retry fallback is deleted, so a failed publish leaves prior
  bytes untouched. `Revoke` (`:269`) now reports `removed=false` when the
  publication fails (the record is still present).
- Test: `TestFailedPublicationPreservesLastValidState`
  (`hookapproval_test.go:232`) stubs `renameFile` to fail and proves
  `Upsert`/`Revoke` error, file bytes identical, `List` still returns the
  prior set, and no staged temp files remain.

### Rev3 validation transcripts (exit codes real, `set -o pipefail`, bash)

Conformance root for test runs:
`CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1`.

- `go build ./...` → exit 0.
- `go vet ./...` → exit 0.
- `gofmt -l internal/ cmd/` → no output, exit 0.
- `go test -count=1 ./internal/shell/ ./internal/hookapproval/...
  ./internal/envfiles/...` → exit 0 (`ok` all three; shell 33.0s final
  run incl. the new nounset regression test, hookapproval 0.65s,
  envfiles 1.3s).
  - `TestShellHookTrustVectors`: 8/8 POSIX cases PASS, each under
    sh+dash+bash+zsh (32 executions); 6/6 ps1 SKIP (no pwsh).
  - `TestShellHookRejectsMalformedRecords`: 14/14 variants PASS (≈112
    POSIX executions across shells × A/B); ps1 SKIP (no pwsh).
  - `TestShellHookRefusesHostileCheckout/posix` ×4 shells PASS;
    `/powershell` SKIP. `TestShellHookTrustResolvesSymlinkedProject`
    posix ×4 shells PASS; powershell SKIP.
  - `TestGeneratedHookParsesUnderPOSIXShells` (32 checks),
    `TestHookEmbedsClosedRecordGrammar`, all pre-existing shell tests:
    PASS. Skip census: 23 SKIP lines, all host-capability (22 no-pwsh,
    1 Windows-only); zero POSIX skips on this host.
- `./internal/install/` (whole package, green across bounded chunks):
  the `go test` wrapper runs (full package, then the `Test[A-C]` chunk)
  were killed by the package timeout with ZERO test output — diagnosed
  as environmental, not a test failure (see the host-conditions note
  below). Reran via a pre-compiled warm test binary (`go test -c -o
  /tmp/install.test`, 0.86s, exit 0; then the binary directly with
  `-test.run` masks, `-test.timeout 240-280s` each), all exit 0:
  `TestA` minus the two slowest conformance tests PASS (22/22);
  `TestAuthoritativeCacheRejectionsAreRebuiltNeverAdopted` solo PASS
  (83.8s); `TestAuthoritativeDryRunCasesMutateNothingPersistent`
  `project-upgrade` / `global-upgrade` /
  `compiled-cache-miss-is-read-only` PASS (1.05s / 0.91s / 14.57s);
  `TestB`, `TestC`, `Test[D-G]`, `Test[H-M]`, `Test[N-R]`, `Test[S-Z]`
  each PASS. Warning/approval-adjacent five
  (`TestProjectInstallRecordsShellHookApprovals`,
  `TestProjectDryRunRecordsNoShellHookApprovals`,
  `TestPostCommitMaintenanceWarningsReachTheResult`,
  `TestMaintenanceFailureAfterCommitIsAWarning`,
  `TestSessionReleaseFailureWarnsWithoutFailingACommittedInstall`)
  additionally rerun together → exit 0, 5/5 PASS. No test failed on an
  assertion in any run; every kill carried zero test output and zero
  own-code frames (see host note).
- `./cmd/curator/ -run TestShellInit`: the `go test` wrapper run was
  likewise timeout-killed with zero output by the first-exec stall;
  rerun via the warm compiled binary → exit 0, both PASS
  (`TestShellInitPrintsHooks` 0.02s,
  `TestShellInitInstallCachesHookWithoutConfig` 0.23s; wall 3m51s,
  ~0 CPU — the delay was the stall, proven by the passing re-exec).
  Full `./cmd/curator/` suite accepted from rev1 (all 163 green across
  chunks; this round changes no `cmd` code and no `cmd` test asserts
  hook text beyond shell-init, which was rerun green).
- `golangci-lint run ./internal/shell/... ./internal/hookapproval/...
  ./internal/envfiles/... ./internal/install/... ./cmd/curator/...` →
  `0 issues.`, exit 0 (two `revive` unused-parameter nits fixed).
- `bash .github/ci/ledger-consistency.sh` → `241 rows checked`, `ok`,
  exit 0 (no ledger change: no new ledger case; skips stay inside the
  existing `host-capability` row).
- Narrowing mutants (disposable, `cmp`-verified restore): M1 neutered
  POSIX `approved_by` check → forged-approver test FAILs (exit 1,
  killed); M2 bypassed hook canonicalization → symlink test FAILs
  (exit 1, killed); M3 restored rev2 remove-and-retry →
  failure-preservation test FAILs (exit 1, killed). 3/3 killed,
  0 survivors.

### Rev3 scope notes

- No `SPEC_PIN` change, no vendored spec bytes, no ax/proposal content,
  no CLI/status additions (still sibling `TASK-260910-3ungjy`), no
  CHANGELOG change (the Unreleased S6 warning-release entry already
  covers this work; no new user-visible surface). Files touched this
  round: `internal/shell/shell.go`, `internal/shell/shell_test.go`
  (one nounset regression test),
  `internal/shell/shell_hook_trust_test.go`,
  `internal/hookapproval/hookapproval.go`,
  `internal/hookapproval/hookapproval_test.go`.
- Known intentional divergence: Go `time.Parse` accepts absurd zone
  offsets (e.g. `+24:00`); the hooks range-check offsets (fail closed).
  Unreachable in practice — the manager always writes `Z` timestamps —
  and probed/documented, not patched around.
- PowerShell execution remains unverified on this host (no pwsh); the
  ps1 matrix (vectors, malformed, hostile, symlink) runs on Windows CI.
- Pre-existing corner, unchanged: a project at the filesystem root yields a
  `//.agents/...` candidate spelling that bash preserves (dash `cd -P`
  normalizes it) while Go stores the `Clean`ed single-slash form, so bash
  warns once for a root-level project. Identical in rev2; fixing it
  completely would need `//x` collapsing that risks Cygwin UNC paths —
  out of scope for this round, reported not patched.
- Host conditions during rev3 validation (shared Tier-3 macOS host, other
  agents' suites + spawn-runners + remote-gate active concurrently):
  fresh test binaries intermittently stall multi-minute with ~0 CPU on
  first exec (two independent `install.test` binaries observed at 0:00
  CPU after minutes; same-inode re-exec 0.022s; `cmd.test` wall 3m51s
  for 0.25s of tests) — macOS first-exec security assessment delay,
  plus fsync contention (one install-test goroutine observed parked in
  `F_FULLFSYNC`) and slow builder tests (84s solo). Mitigation used:
  `go test -c` once, then the warm binary directly with `-test.run`
  masks and `-test.timeout 240-280s` per chunk. Every timeout-kill cited
  above carried zero test output and zero own-code frames; the one
  inspected stuck test passes solo (exit 0).

## Revision 4 — CR rev3 hosted-gate repair (RUN-260917-e0f2de)

CR rev3 failed the hosted gate on all five lanes (run 35189457421; the
`remote-gate.sh` tail shows only `go test exit=1`, so the failing test
was identified from the `test-evidence-*` artifacts' `go-test.json`
streams): exactly one test fails on every lane —
`TestShellHookTrustResolvesSymlinkedProject/powershell`. POSIX (incl.
the new sh/dash coverage), malformed-record, hostile, and vector suites
were green everywhere; Lint and the platform-case gate were green.

Root cause: the rev3 `Get-CuratorTrustCanonical`
(`internal/shell/shell.go`) probed `.Target` only on the final candidate
path. A symlink at an ancestor — the test's `alias-project` directory,
and in general any aliased project dir — is invisible to `Get-Item` on
the file, so the hook looked up the unresolved alias spelling, missed
the real-path record, and warned `shell_hook_env_unapproved` (under B
the activation then refused, so `sourced1=1` never printed). Locally the
subtest always skipped (no pwsh on this host), which is why rev3 went
out green.

Fix (`internal/shell/shell.go`, `Get-CuratorTrustCanonical` only — no
Go-code, test, ledger, or CHANGELOG change):

- Component-wise EvalSymlinks semantics: walk every ancestor, expand a
  link target, reattach the unprocessed tail, restart the scan over the
  expanded absolute path; 32-hop bound; unlistable components stay
  lexical (Go ancestor-fallback parity).
- Link probe order per component: .NET `DirectoryInfo`/`FileInfo`
  `.LinkTarget` first, `Get-Item .Target` as fallback. The .NET probe is
  required, not cosmetic: on this macOS host the filesystem provider
  cannot list root-level links at all (`Get-Item -LiteralPath /tmp` →
  "Could not find item /tmp") while `.LinkTarget` reports `private/tmp`
  correctly. `Get-Item` stays for runtimes without `LinkTarget`
  (Windows PowerShell 5.1).
- Parent tracked as the previous accumulator, never `Split-Path
  -Parent`: `Split-Path -Parent "/tmp"` returns an empty string on
  Unix, which made `Join-Path` throw a terminating binding error that
  the outer catch swallowed into a silent no-resolution (found by a
  noisy-catch probe during this round; every relative-target link at a
  top-level directory hit it).

Verification (all with real pwsh this round — 7.4.6 macOS tarball
extracted to `/tmp/pwsh`, nothing installed system-wide):

- `TestShellHookTrustResolvesSymlinkedProject/powershell` PASS (5.0s);
  posix sh+dash+bash+zsh PASS.
- `TestShellHookTrustVectors`: 14/14 cases PASS, 0 SKIP — the first run
  in this task's history to execute the ps1 matrix locally (6 ps1 cases
  under pwsh, 8 sh cases × 4 shells).
- Throwaway edge probe of the emitted B hook (kept in `/tmp`, not
  committed): alias-dir, relative file link, plain, dangling, 2-cycle,
  missing trailing, `..`, `/etc/hosts` — all terminate; every
  reachable path resolves byte-identical to Go `Canonicalize`
  (dangling/cycle spellings differ lexically but both sides fail closed
  with an unapproved warning, and neither can name an existing file, so
  the divergence is unreachable via `Test-Path`). Emitted A/B hooks
  parse with 0 errors under the PowerShell parser API.
- Full narrow gates on the final tree (exit codes real, `bash`,
  `CURATOR_CONFORMANCE_ROOT=.../curator-spec/conformance/v1`,
  pwsh on `PATH`): `go build ./...` → exit 0; `go vet ./...` →
  exit 0; `gofmt -l internal/ cmd/` → no output, exit 0;
  `go test -count=1 ./internal/shell/... ./internal/hookapproval/...
  ./internal/envfiles/...` → exit 0 (`ok` all three; shell 78.0s);
  `go test -count=1 ./internal/install/` (whole package, one bounded
  call, 327.7s) → exit 0 — run before the final `$parent` one-liner,
  which is valid because `internal/install` never consumes the hook
  templates (verified by grep; only `cmd/curator/main.go:1870`
  installs hooks); `go test -count=1 ./cmd/curator/ -run
  TestShellInit` → exit 0 (rerun after the final edit, both PASS);
  `golangci-lint run ./internal/shell/... ./internal/envfiles/...
  ./internal/hookapproval/... ./internal/install/...` → `0 issues.`,
  exit 0.
- Files touched this round: `internal/shell/shell.go` only
  (`Get-CuratorTrustCanonical` + two comment blocks). No ledger,
  SPEC_PIN, CHANGELOG, or sibling-surface change.

---

## Revision 5 — review-verdict-rev4 rework: Windows / Git Bash identity and coverage

Verdict `TASK-260910-1952mz_review-verdict-rev4.md` (changes_requested, route
to-dev) is the authority; rework brief `TASK-260910-1952mz_rework-rev5.md`
scopes the one P1 correction. R1, R2 (Unix), R3 (local alias), R4 stay as
the verdict marked PASS; shipped profile stays **`A-warning`**. No CLI/status
change (still sibling `TASK-260910-3ungjy`), no SPEC_PIN change, no vendored
spec bytes, no CHANGELOG change (the Unreleased S6 warning-release entry
already covers this work; no new user-visible surface beyond the Git Bash
identity repair).

### 1. One identity for native Go and Git Bash lookup

- `internal/hookapproval/hookapproval.go:87-100` documents the single
  Windows identity rule next to `Canonicalize`: on Windows the canonical
  spelling is the native absolute path with backslash separators and an
  uppercase drive letter (UNC keeps `\\`), and record matching folds case.
  `normalizeWindowsDrive` (`:146`) uppercases the drive after symlink
  resolution; `pathsEqual` (`:162`) is exact on Unix and `EqualFold` on
  Windows, used by `Lookup` (`:260`), `Upsert` (`:283`), and `Revoke`
  (`:332`). Unix behavior is byte-identical (normalization is a Windows-only
  no-op there, fold is exact).
- `internal/shell/shell.go` POSIX hook: new `_curator_trust_is_windows_shell`
  (`:312`, `uname -s` matches `MINGW*|MSYS*|CYGWIN*`) and
  `_curator_trust_identity` (`:325`, cites `hookapproval.Canonicalize`):
  under MSYS/Git Bash/Cygwin the resolved candidate is mapped through
  `cygpath -w` (`:331`, CR-stripped), the drive letter uppercased, and the
  native spelling becomes the lookup and warning identity; elsewhere the
  resolved spelling passes through. When `cygpath` is absent the function
  fails and `_curator_trust_allow` (`:467`) warns `shell_hook_env_unapproved`
  and refuses under BOTH profiles — never a silent source (documented in the
  hook comment at `:321-324`). `_curator_trust_recorded` takes a fold flag
  and compares `tolower($1) == tolower(want)` under MSYS (`:423`), exact
  elsewhere. Strictly POSIX (`case`, `command -v`, `tr`/`cut`; no arrays or
  `[[ ]]`); `sh -n`/`dash -n`/`bash -n` green on all four generated variants.
- PowerShell hook untouched (native-to-native already; rev4 gate green).

### 2. Windows POSIX coverage runs instead of GOOS skips

- `internal/shell/shell_hook_trust_test.go`: the four unconditional
  `runtime.GOOS == "windows"` skips (vector POSIX runner, hostile, malformed,
  symlink-alias posix subtests) are deleted. `posixTrustShells` (`:400`)
  probes `bash`/`sh` (+`.exe`) on Windows via `LookPath` plus the well-known
  Git for Windows install roots (`C:\Program Files\Git\bin|usr\bin\...`),
  so the vector, hostile, malformed, and alias cases run wherever Git Bash
  is actually present; the skip (`:459`) names the absent interpreter
  (`bash is unavailable (Git Bash absent on this runner)`, host-capability
  class), never the GOOS alone. `runTrustShell` (`:472`) folds the `.exe`
  suffix so Git Bash takes the same argument shapes as its Unix names.
- Warning assertions compare by the spelling the hook prints on Windows
  without weakening sourced/diagnostic/count: `assertTrustWarningWindows`
  (`:245`) counts the `curator hook approve ` prefix exactly, requires every
  warned path token (`warnedApprovalPaths`, `:304`) to normalize
  (`normalizeWindowsPathToken`, `:287`: separators unified, leading
  `/c/...` and `/cygdrive/c/...` converted to drive form, lowercased) to the
  expected native candidate (`windowsWarningSpellings`, `:276`), and keeps
  the first/second-activation split exact; `assertHostileWarning` (`:682`)
  branches the same way. Unix assertions are byte-identical to rev4. CRLF
  normalization retained on both branches.
- `.github/ci/platform-cases.tsv:150-157`: the S6 comment and the `/*` row
  now state the real condition (interpreter genuinely absent; on Windows Git
  Bash present → POSIX subtests must run, absent → host-capability skip).
  No row added or removed; `ledger-consistency.sh` green (241 rows).

### 3. Cross-spelling proof test

- `TestShellHookTrustNativeRecordAuthorizesMSYSSpelling`
  (`shell_hook_trust_test.go:1059`, Windows-only): proves Git Bash observes
  MSYS spelling (`pwd` starts with `/`), one native manager record sources
  silently under both `A-warning` and `B-enforcing`, and changed bytes are
  refused under B (sourced with the `shell_hook_env_changed` warning under
  A). Skips elsewhere as platform-control (`... is exercised on Windows`).
- `TestWindowsDriveIdentity` (`hookapproval_test.go:340`, Windows-only):
  stored spelling carries an uppercase drive letter; a fully lowercased
  spelling of the same (not-yet-existing, mixed-case) file resolves to the
  one record for lookup and revoke.
- `TestWindowsWarningNormalization` (`:1133`, all platforms): native,
  lower-drive, forward-slash, `/c/`, `/C/`, `/cygdrive/c/` spellings fold
  to one file; a different drive does not; warned-path extraction names the
  MSYS token.
- `TestPosixHookEmitsWindowsIdentity` (`:1173`, all platforms): pins the
  emission (`_curator_trust_identity`, MSYS case, `cygpath -w`, the tolower
  lookup, the `hookapproval.Canonicalize` citation) under both profiles.

### Rev5 validation transcripts (exit codes real, `set -o pipefail`, bash)

Conformance root for test runs:
`CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1`,
`PATH=/tmp/pwsh/app:$PATH` (real pwsh 7.4.6, pre-existing tarball).

- `go build ./...` → exit 0.
- `go vet ./...` → exit 0.
- `gofmt -l internal/ cmd/` → no output, exit 0.
- `GOOS=windows go build ./internal/shell/ ./internal/hookapproval/` → exit 0.
- `GOOS=windows go vet ./internal/shell/ ./internal/hookapproval/` → exit 0.
- `go test -count=1 ./internal/shell/... ./internal/hookapproval/...
  ./internal/envfiles/...` → exit 0 (`ok` all three; shell 80.5s,
  hookapproval 1.9s, envfiles 1.8s):
  - `TestShellHookTrustVectors`: 14/14 PASS, 0 SKIP (8 sh × sh/dash/bash/zsh,
    6 ps1 under real pwsh).
  - `TestShellHookRefusesHostileCheckout` posix+powershell PASS;
    `TestShellHookTrustResolvesSymlinkedProject` posix+powershell PASS;
    `TestShellHookRejectsMalformedRecords` PASS (62.9s);
    `TestGeneratedHookParsesUnderPOSIXShells` (32 checks),
    `TestHookEmbedsClosedRecordGrammar`, `TestWindowsWarningNormalization`,
    `TestPosixHookEmitsWindowsIdentity` PASS.
  - `TestShellHookTrustNativeRecordAuthorizesMSYSSpelling` SKIP on macOS
    (Windows-only, platform-control); `TestWindowsDriveIdentity` SKIP on
    macOS (Windows-only). Both execute on the hosted Windows lane.
- `go test -count=1 ./internal/install/ -run
  'TestProjectInstallRecordsShellHookApprovals|TestProjectDryRunRecordsNoShellHookApprovals'`
  → exit 0 (both PASS, 5.6s).
- `go test -count=1 ./cmd/curator/ -run TestShellInit` → exit 0 (both PASS).
- Empty-root `TestShellHookTrustVectors` → SKIP (root-content), exit 0.
- `golangci-lint run ./internal/shell/... ./internal/hookapproval/...
  ./internal/envfiles/... ./internal/install/...` → `0 issues.`, exit 0
  (one `misspell` false-positive on a `D:\...` fixture word avoided by
  renaming the fixture).
- `bash .github/ci/ledger-consistency.sh` → `241 rows checked`, `ok`, exit 0.
- MSYS simulation probe (throwaway `/tmp/msys-probe.sh`, stubbed
  `uname`→MINGW64 + `cygpath -w`, real `sh` + emitted hooks): 19/19 PASS —
  approved native record sourced silent under A and B; lower-drive +
  upper-component record folded to silent; unapproved refused under B with a
  single native-path warning (+ approve command) and sourced+warned under A;
  changed bytes refused under B with `shell_hook_env_changed`; cygpath
  absent refused with warning under BOTH profiles (fail-closed, never
  silent). Kept in `/tmp`, not committed: it exercises the branch logic,
  while real Windows proof is the hosted gate at handoff.
- Narrowing mutants (disposable, `cmp`-verified restore): M1 neutered the
  tolower fold → `TestPosixHookEmitsWindowsIdentity` FAILs (exit 1,
  killed); M2 broke `cygpath -w` → same test FAILs (exit 1, killed). 2/2
  killed, 0 survivors. Bound: emission-shape mutants; Windows behavioral
  mutants await the hosted lane.

### Rev5 scope notes

- Files touched this round: `internal/hookapproval/hookapproval.go`,
  `internal/hookapproval/hookapproval_test.go`, `internal/shell/shell.go`
  (POSIX trust part only; PowerShell bytes untouched),
  `internal/shell/shell_hook_trust_test.go`,
  `.github/ci/platform-cases.tsv` (comment + behaviour text only).
- Known bound: the `tolower` fold is POSIX-awk ASCII-oriented while Go
  `EqualFold` is Unicode; both sides observe the same filesystem bytes in
  practice (same file, same case), with the fold covering drive-letter and
  spelling-case drift. A `//`-rooted project corner from rev3 is unchanged.
- Windows behavior (Git Bash execution, cross-spelling, drive fold) is
  proven by the hosted gate at handoff, which runs the Windows lane with
  Git Bash and pwsh present.

---

## Revision 6 — CR rev5 hosted-gate repair (Windows cross-spelling marker)

CR rev5 failed the hosted gate on exactly one lane (run 35201254365;
`test-gate: go test exit=1, platform-case gate exit=0` on Test
windows-latest; all other lanes green). The `test-evidence-windows-latest`
artifact's `go-test-served.json` stream names exactly one failing test —
the new Windows-only `TestShellHookTrustNativeRecordAuthorizesMSYSSpelling`
(all 4 Git Bash subtests: `bash`, `sh`, `bash#01`, `sh#01`). Every other
POSIX-under-Git-Bash suite the rev5 round enabled (hostile, malformed,
symlink-alias) passed on Windows, as did `TestWindowsDriveIdentity`.

Root cause (harness-only, no production change): the cross-spelling test's
changed-bytes phase rewrites the fixture to
`export CURATOR_PROJECT_ENV=2`, but `assertTrustOutcome` hardcoded the
sourced marker to `"1"`. The hook behaved exactly right — the CI stderr
shows `shell_hook_env_changed` with the native path plus one warning, and
stdout shows `sourced1=2`/`sourced2=2` (A-warning sources the NEW bytes
with a warning) — while the assertion demanded `sourced1=1`. Notably the
two approved-phase activations (silent sourcing under A AND B through
MSYS spelling against one native record) passed on Windows, i.e. the rev5
identity repair itself is proven working; only this assertion was wrong.
All vector fixtures export `=1` even in changed form (v2 adds a second
variable), so no vector case could have caught the hardcoded marker.

Fix (`internal/shell/shell_hook_trust_test.go` only — production bytes
identical to rev5):

- `hookTrustCase` gains harness-only `SourcedMarker string` (`json:"-"`,
  `:50-56`): expected `${CURATOR_PROJECT_ENV}` value when `Sourced` is
  true; empty means `"1"`, so every decoded vector case is unaffected.
- `assertTrustOutcome` (`:197-205`) honors `SourcedMarker` when set.
- The cross-spelling `expectA` (`:1131-1134`) sets `SourcedMarker = "2"`,
  keeping the stronger proof (the NEW `=2` bytes are sourced under A,
  not stale state) instead of weakening the fixture back to `=1`.
- `gofmt -w` realignment of the struct block (comment splits the field
  alignment); verified the rev6-vs-rev5 delta is exactly these hunks.

Proof the fix targets the failure (throwaway in-package probe
`zz_rev6_probe_test.go`, run once, archived as
`TASK-260910-1952mz_rev6-probe_test.go`, then deleted from the tree):

- Old expectation (no marker) against the byte-exact CI stdout/stderr
  pair (CRLF kept): FAIL, exit 1, with the identical CI message
  `stdout lacks "sourced1=1"` — the diagnosis reproduces.
- New expectation (`SourcedMarker = "2"`) against the same pair: PASS,
  exit 0 — the fix accepts exactly the CI-observed correct behavior.

### Rev6 validation transcripts (exit codes real, `bash`, `set -o pipefail` where piped)

Conformance root for test runs:
`CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1`,
`PATH=/tmp/pwsh/app:$PATH` (real pwsh 7.4.6, pre-existing tarball).

- `go build ./...` → exit 0.
- `go vet ./...` → exit 0.
- `gofmt -l internal/ cmd/` → no output, exit 0 (after one `gofmt -w`
  on the touched test file).
- `GOOS=windows go build ./internal/shell/ ./internal/hookapproval/` →
  exit 0; `GOOS=windows go vet` on the same → exit 0.
- `go test -count=1 ./internal/hookapproval/... ./internal/envfiles/...`
  → exit 0 (`ok` both; 161.8s/162.3s wall — see host note).
- `./internal/shell/` via a pre-compiled warm binary (`go test -c -o
  /tmp/shell-rev6.test`, exit 0) in bounded `-test.run` chunks, all
  exit 0: shape/grammar/normalization/emission/parses/variants (7 PASS);
  pre-existing behavior (10 PASS + `TestPowerShellHookRunsOnEveryPrompt`
  SKIP, pre-existing Windows-only); `TestShellHookTrustVectors` 14/14
  PASS with 0 SKIP (8 sh × sh/dash/bash/zsh + 6 ps1 under real pwsh);
  `TestShellHookRejectsMalformedRecords` 99 subtests PASS;
  `TestShellHookRefusesHostileCheckout` + `TestShellHookTrustResolvesSymlinkedProject`
  posix+powershell PASS; `TestShellHookTrustNativeRecordAuthorizesMSYSSpelling`
  SKIP on macOS (Windows-only, executes on the hosted lane).
- Empty-root `TestShellHookTrustVectors` → SKIP (root-content), exit 0.
- `go test -count=1 ./internal/install/ -run
  'TestProjectInstallRecordsShellHookApprovals|TestProjectDryRunRecordsNoShellHookApprovals'`
  → exit 0 (both PASS, 13.0s).
- `go test -count=1 ./cmd/curator/ -run TestShellInit` → exit 0
  (both PASS, 0.66s).
- `golangci-lint run ./internal/shell/... ./internal/hookapproval/...
  ./internal/envfiles/... ./internal/install/...` → `0 issues.`,
  exit 0.
- `bash .github/ci/ledger-consistency.sh /tmp/ledger-ev` →
  `241 rows checked`, `ok`, exit 0 (no ledger change this round).
- Not rerun in rev6: full `./internal/install/` and `./cmd/curator/`
  suites (untouched by this round — one test-file-only change; rev4 ran
  both green and the handoff remote-gate reruns the full suite
  including the Windows lane that reproduces the original failure).

### Rev6 scope notes

- Files touched this round: `internal/shell/shell_hook_trust_test.go`
  only. No production, ledger, SPEC_PIN, CHANGELOG, or sibling-surface
  change (the Unreleased S6 warning-release entry already covers this
  work). Shipped profile stays **`A-warning`**.
- Host conditions during rev6 validation (shared Tier-3 macOS host):
  a plain `go test ./internal/shell/...` wrapper run hit the 600s
  package timeout mid-suite (zero assertion failures — the dump shows
  the run parked in a shell activation), and the tiny hookapproval /
  envfiles suites took ~162s each (rev5: ~2s). Both are the documented
  first-exec/environmental stall, not a code slowdown: the change is
  three string-comparison lines. Mitigation used, as in rev3: `go test
  -c` once (0.8s), then the warm same-inode binary directly with
  `-test.run` masks (`-test.list` 0.4s; every chunk green above).
- Windows proof of the corrected assertion is the hosted gate at
  handoff, which runs the cross-spelling test on the Windows lane with
  Git Bash present; the local probe above proves the fixed assertion
  accepts the exact bytes that lane produced.
