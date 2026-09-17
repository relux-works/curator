# TASK-260910-1952mz — producer logbook rev4 (2026-09-17)

Repair of the CR rev3 hosted-gate failure (run 35189457421, all five lanes).
Full transcripts in `TASK-260910-1952mz_results.md` (rev4 appendix).

## Finding: PowerShell provider is blind to root-level links on macOS

`Get-Item -LiteralPath /tmp` fails with "Could not find item /tmp" under
pwsh 7.4.6 on macOS 15.7.4 (same for `/var`; `/private` also unlistable),
while .NET `([System.IO.DirectoryInfo]'/tmp').LinkTarget` correctly reports
`private/tmp`. Any future hook/test code that must observe symlinks must use
the .NET `LinkTarget` probe, not the provider. Whether this reproduces on CI
macOS runners is unknown (irrelevant now: the shipped code prefers .NET).

## Finding: `Split-Path -Parent` returns '' for single-component rooted paths

`Split-Path -Parent '/tmp'` → empty string on Unix pwsh. Feeding that into
`Join-Path -Path` throws a terminating binding error. The hook tracks the
parent accumulator directly instead. Sibling approve/status work must not
reintroduce `Split-Path -Parent` on shallow paths.

## Decision

Fix scoped to `Get-CuratorTrustCanonical` in `internal/shell/shell.go`; no
test/ledger/CHANGELOG change. The previously failing committed test now
passes unmodified, which is the strongest signal the production predicate —
not the test — was wrong.

## For the sibling (TASK-260910-3ungjy) and reviewers

- pwsh 7.4.6 macOS tarball works from `/tmp/pwsh` with no system install;
  all 14 vector cases execute locally with it (0 skips). Recommend the
  reviewer do the same instead of accepting ps1 SKIP evidence again.
- `remote-gate.sh` failure tails do not name the failing test; the
  `test-evidence-*` run artifacts' `go-test.json` streams do
  (`gh api repos/.../actions/runs/<id>/artifacts`, then grep `Action=fail`).
