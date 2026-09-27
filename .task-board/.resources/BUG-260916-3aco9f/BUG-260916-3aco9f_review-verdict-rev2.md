# BUG-260916-3aco9f review verdict — CR rev2: ACCEPTED

Reviewer: claude-opus-5-5 (low). Candidate tree eff08565 vs base bd3c0f43 (2 test files, no production change).

## Reproduced independently (zsh, set -o pipefail)
- `go test ./cmd/curator -run 'TestProfileInstall(ActivationMatrix|StopThenRetry|ReinstallUsePreserves)' -count=1` → ok (247.7s). 3 shapes × 3 kinds × 4 flag cases = 36 matrix rows, 3 stop-then-retry rows and 3 project-scope rows, all driven through run() (`curator profile install`).
- `go test ./internal/registry -run TestSnapshotZeroClockSkewIsLiteral -count=3` → ok.

## Mutant (drop activation for the git shape)
envprofile.go git same-source branch: `else if activate {` → `else if activate && false {`.
Killed by 10 rows: git-root and git-collection × {unchanged, changed lock} × {use, use-takeover}, plus StopThenRetry/git-root and /git-collection. path-root rows still pass (separate reinstallPathLocked path), which is expected. Source restored byte-identical (git diff is empty).

## Coverage vs. the brief
- Operand kinds: installOperandKind only accepts path or git (envprofile.go:1592). "repository/Skillfile schema-2 collection" maps to git + --directory (git-collection); there is no registry operand for profile install. Stated bound, not a gap.
- "Changed-source reinstall": the matrix uses same source with a changed lock. A *different* source under the same name is refused with DiagNameTaken (envprofile.go:739) before any activation runs, so --use/--takeover don't apply there. Bound: no row asserts the refusal together with the flags.
- Scopes: the project-scope (codex_cli --env) divergence stays intact under machine --use reinstall (TestProfileInstallReinstallUsePreservesProjectScope).
- Only tests changed. Every shape already passed on the base because 187z6x fixed git and stage (c) fixed path. The mutant confirms the rows depend on the production activation step.

## registry_test.go change
The test now creates the persistent snapshot caches and writes their state catalogs before it samples `now`, so cache setup time can't push the injected-now 500ms boundary. The assertions are unchanged: zero skew still refuses and the default skew still absorbs. This is a justified flake fix and nothing is weakened.

## Not in this leaf
Closing issue #73 with the landed commit and adding the curator-spec conformance vector are for the orchestrator/integration. The CLI-level vector is TestProfileInstallStopThenRetryConformanceVector.
