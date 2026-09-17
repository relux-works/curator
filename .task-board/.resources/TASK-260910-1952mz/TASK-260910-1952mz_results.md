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
