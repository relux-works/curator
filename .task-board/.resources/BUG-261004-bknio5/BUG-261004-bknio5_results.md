# BUG-261004-bknio5 — gc-sweeps-live-runtime-on-uncertain-marks

Developer review packet. Work remains uncommitted on the provided Story branch.

The uncertainty check now runs before runtime deletion in `scopes.Collect`. Any uncertainty skips runtime and both build-cache sweeps while retaining the existing conservative registry/consumer pruning. Existing source diagnostics explain why the reference set is incomplete; a new warning explicitly says the runtime sweep was skipped. The runtime-only `CollectRuntime` API also refuses an uncertain sweep and reports its reasons through its error result.

The production CLI path is `run(["gc"]) → cmdGC → collectUnderLock → scopes.Collect`. The regressions use the actual home locks and `runtimestore.WriteBinShim`, with executable fixtures and launcher executions before and after each pass.

## Acceptance evidence

| AC | Proof |
| --- | --- |
| 1 | 3/3 CLI uncertainty rows preserve live and orphan runtime over two passes; consumer registry bytes remain unchanged. 11/11 local internal conservative cases retain runtime, never call the recording build cache, keep consumers, and exercise the runtime-only refusal. |
| 2 | Truncated registry, invalid marker, and unreadable marker each call the production CLI twice and run the working shim before/after GC. The unreadable marker is a directory at the marker path, producing a genuine portable read failure even under privileged execution. |
| 3 | 1/1 complete-reference CLI control removes the orphan on pass one, retains the live runtime, and preserves the working shim on both passes. Existing normal collection tests also pass in the package suite. |
| 4 | Each uncertain pass checks the source diagnostic, "runtime sweep skipped", incomplete-reference explanation, and the existing build-cache skip warning. |

Coverage: 4/4 requested CLI rows, 8/8 GC invocations, 2/2 requested mutants caught. This is a scoped N1 regression claim, not a general security certification.

## Commands and real exits

All local Go commands used `GOFLAGS=-work`. Local toolchain: `go1.26.0 darwin/amd64`. No command was piped through tee.

| Command (local) | Exit | Interpretation |
| --- | ---: | --- |
| `env -u GOROOT GOFLAGS=-work GOMAXPROCS=2 go test -p 2 ./cmd/curator -run '^TestGCPreservesRuntimeOnUncertainMarks$' -count=1 -v` before fix | 1 | Expected red on main: all uncertainty rows delete live runtime; control passes. |
| `env -u GOROOT GOFLAGS=-work GOMAXPROCS=2 go test -p 2 ./internal/scopes -run '^TestCollectStaysFailSafe' -count=1 -v` before fix | 1 | Expected red: uncertain scopes remove runtime. |
| `env -u GOROOT GOFLAGS=-work GOMAXPROCS=2 go test -p 2 ./internal/scopes -run '^TestCollectStaysFailSafeAcrossConsecutivePasses/corrupt_consumer_registry$' -count=1` before fix | 1 | Expected red: runtime-only API also sweeps rather than refusing. |
| `env -u GOROOT GOFLAGS=-work GOMAXPROCS=2 go test -p 2 ./cmd/curator ./internal/scopes -run '^(TestGCPreservesRuntimeOnUncertainMarks\|TestCollectStaysFailSafe)' -count=1 -v` after fix and after restoring mutants | 0, 0 | Green regressions; verbose restored run attached. |
| `env -u GOROOT GOFLAGS=-work GOMAXPROCS=2 go test -p 2 ./internal/scopes ./internal/runtimestore -count=1 -timeout=5m` | 0 | Both package suites pass. |
| `env -u GOROOT GOFLAGS=-work GOMAXPROCS=2 go test -p 2 ./cmd/curator -run '^TestGC' -count=1 -timeout=5m` | 0 | Relevant CLI suite passes, including lock serialization and live-process build retention (93.033s reported for package). |
| `env -u GOROOT GOFLAGS=-work GOMAXPROCS=2 go build -p 2 -o .temp/BUG-261004-bknio5/curator ./cmd/curator` | 0 | Production CLI builds. |
| `env -u GOROOT GOFLAGS=-work GOMAXPROCS=2 go vet -p 2 ./cmd/curator ./internal/scopes ./internal/runtimestore` | 0 | Relevant vet checks pass. |
| `env -u GOROOT GOFLAGS=-work GOMAXPROCS=2 golangci-lint run --timeout=5m` | 3 | Existing parallel linter held its shared lock; this is a failed attempt. |
| Same lint with `--allow-parallel-runners` | 1 | Shared lint cache returned stale findings from other worktrees; this is a failed attempt. No product changes were made to suppress those findings. |
| `env -u GOROOT GOFLAGS=-work GOMAXPROCS=2 GOLANGCI_LINT_CACHE="$PWD/.temp/BUG-261004-bknio5/lint-cache" golangci-lint run --allow-parallel-runners --timeout=5m` | 0 | Isolated task-local cache: 0 issues. |
| CLI regression with uncertainty guard reordered after runtime sweep | 1 | Expected red, mutant caught by all three uncertainty rows. |
| Truncated-registry CLI regression with only registry-read uncertainty omitted | 1 | Expected red, narrowed-source mutant caught. |
| `git diff --check` | 0 | Patch whitespace checks pass. |
| `git diff --exit-code b78a1b2a358babffd8499416ceca5a24a844aaad --` | 0 | Restored tracked source matches hosted snapshot. |
| Assert empty `gofmt -l` output for all three changed Go files | 0 | Changed files formatted. |

Both mutants were restored from a saved green copy with byte equality asserted before the green replay. The broader package, CLI, build, vet, and lint evidence is reused after restoration because the source/test/configuration identity is identical; no previously attached evidence replaced these developer-run commands.

Local conformance-root-dependent cases use their existing skip paths because `CURATOR_CONFORMANCE_ROOT` is unset. The full pinned conformance and race lanes are delegated to the repository's hosted workflow; local package greens are not claimed as a full pinned conformance gate. The normal CLI control can emit the existing fail-safe process-inventory warning on this host; its runtime deletion/retention assertions still execute, and internal recording-cache assertions prove uncertainty prevents build sweeping.

## Hosted candidate

Base/main (also freshly advertised by origin): `934952a45953587a1d4184b692b3fb4ee401e732`.
Detached verification snapshot: `b78a1b2a358babffd8499416ceca5a24a844aaad`.
Candidate tree: `31d96981112f3a682357b8064b28ed8ba5c1d42b`.

[Hosted CI run 37169362606](https://github.com/relux-works/curator/actions/runs/37169362606).

Hosted verdict: **success**, all 11 required jobs green (three OS test lanes, two race lanes, lint, interop, naming and three gate self-tests). `gh run view 37169362606 --exit-status --json status,conclusion,headSha,jobs --jq '{status,conclusion,headSha,jobs:[.jobs[]|{name,conclusion}]}'` returned **exit 0** and identifies the exact snapshot above. This is the bounded hosted gate verdict command; earlier status queries were observations, not passing-gate claims.

The uploaded Ubuntu `test-evidence-ubuntu-latest/test/go-test.json` was downloaded and parsed directly. A Python assertion of the exact required subtest set and the CLI package pass event returned **exit 0**: **4/4** N1 CLI rows passed in the full hosted CLI suite. Raw stream SHA-256: `09e2c435956aafeb876e6d1d130ea307c317e42dbae787b5121cc370196f9aea`. The workflow artifact is the source evidence; its corpus is not duplicated on the board.

The repository's remote-gate snapshot/push/wait protocol was split into bounded calls, with a neutral detached commit message to avoid publishing the workspace's personal path. The Story branch, HEAD and real index were untouched. The workflow executes its committed CI configuration and immutable conformance pin. The optional candidate-conformance and main-only self-hosted lane are excluded by the existing gate-branch rules; they are not claimed as executed.

## Host observations and findings

`syspolicyd` was inspected using `launchctl print system/com.apple.security.syspolicy`. Successive-crash count was 392 before red runs, after red runs, after the longer local checks, and after both mutation checks; runs count stayed 495 in those samples. The closing sample after the hosted wait was **394 successive crashes / 497 runs**, an increase of 2 since the last local-validation sample. The cause of the increase is unknown; no attribution to this change or the remote workflow is established. These observations measure the counter only and do not infer broader host health.

Findings and the lint-cache anomaly are recorded in board notes and this packet. `LOGBOOK.md` and `CHANGELOG.md` remain untouched per the task-specific instruction; the board notes are the persistent findings record for this run.

Changed-file SHA-256:
- `internal/scopes/gc.go`: `8434a74f0de547d885928fca0530e602774c1b91727122a954ac996dfe26c8c8`
- `internal/scopes/gc_conservative_test.go`: `979b7fea9559b1ba4fb88509de6df76cc896908ed2891fca6c0e78be7f6182c1`
- `cmd/curator/gc_test.go`: `4d96944613108a36c1a025b96fea2b1ea306b7ac2fdc08132b7a40ca820773b3`


After the closing CI verdict, local commands stopped returning promptly, including shell-only probes. Non-PTY probes and a cosmetic packet-formatting command were interrupted with exit 130; the final TTY probe exited 1 when interrupted. No validation gate was canceled or substituted. Execution subsequently returned; the board update exited 0 and the restored tracked tree was rechecked against the hosted snapshot with exit 0 before handoff. The cause of this execution delay is unknown.
