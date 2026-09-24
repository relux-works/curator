# BUG-260924-5p8b0z — integration precondition check (bound developer run)

Board status at check time: `integrating` (left untouched per assignment).
HEAD (Story-branch base): `5b326aa3` — matches the accepted base; no trunk-intersection work done here.
`worktree integrate` / `worktree checkpoint` NOT executed (forbidden by the superseding
Integration Assignment; the runner performs the bound landing). No files changed, no commits,
no status writes, no handoff call in this run.

## Candidate content verified (uncommitted worktree state)
- `M internal/scriptworker/exec.go` — the fix (4 insertions, 1 deletion):
  `resolveExecForPlatform` now requires `managerEnvironment != nil` before deriving the
  trusted System32 root:
  `if platform == "windows" && useDefaultSearchDirs && managerEnvironment != nil`
  A nil manager snapshot may still supply ambient SYSTEMROOT for the default search list
  (via `parseHostEnvironment` → `os.Environ()`), but that ambient value no longer confers
  hard-link trust.
- `?? internal/scriptworker/exec_identity_conformance_test.go` (new)
- `?? internal/scriptworker/testdata/executable_identity_cases.json` (new, 8 cases)
- No CHANGELOG edit. No other paths in `git status --short`.

## Validation (each run directly, `set -o pipefail`, real exit codes)
Shell: bash. Note: the default `GOCACHE` on this host is corrupt
(`could not import internal/coverage/rtcov ... no such file or directory`, exit 1 on first
attempt — environment issue, not code); all gates below reran green with a fresh
`GOCACHE="$(mktemp -d)/gocache"`.

| Command | Exit |
|---|---|
| `go test ./internal/scriptworker/ -run TestExecutableIdentityCasesAtProductionEntry -v -count=1` | 0 — all 8 subtests PASS, incl. `windows-exec-uncaptured-systemroot-hardlinks` (rejected) and the one accepting System32 case |
| `go test ./internal/scriptworker/ -count=1` (full package) | 0 (`ok ... 125.440s`) |
| `go vet ./internal/scriptworker/` | 0 |
| `gofmt -l` on the two `.go` files | 0 (no output = clean) |
| `go build ./internal/scriptworker/` | 0 |
| Fixture `sha256sum` | `124e0075...104292a2`, matches the pinned `executableIdentityFixtureSHA256` asserted by the test |

## Mutant kill (drop the captured-SystemRoot condition)
Re-applied the mutant locally (`managerEnvironment != nil` removed), ran only the target case,
then restored the fix (`git diff --stat HEAD` back to 1 file changed):
`go test ./internal/scriptworker/ -run TestExecutableIdentityCasesAtProductionEntry/windows-exec-uncaptured-systemroot-hardlinks`
→ FAIL, exit 1:
`production resolver accepted = true, want false (reason systemroot-not-manager-captured; platform_owned=true)`
— the exact hosted-run symptom. Mutant killed on the local lane (Windows exercised via the
GOOS/env seam: `deriveProfileForPlatform(..., "windows")`, the production derivation path
behind `DeriveProfile`).

## Gap-ledger / other-cases check
- `uncaptured` appears only in the new test driver + fixture; it is NOT in the test's
  preserved-gap switch (only `windows-exec-noncomponent-store-hardlinks` and
  `windows-exec-unowned-file-hardlinks` remain listed as owned R5 gaps) — the uncaptured
  gap row is gone and the case is enforced-reject.
- The other 7 identity cases behave per fixture (single accepting case still accepts;
  interpreter cases still reject multiply-linked targets).

## CHANGELOG entry (for release prep — do NOT edit CHANGELOG.md in a leaf)
`scriptworker`: require a manager-captured SYSTEMROOT for the Windows System32 hard-link
allowance; ambient/uncaptured SYSTEMROOT values no longer confer hard-link trust
(`windows-exec-uncaptured-systemroot-hardlinks` now rejected at the production entry).
