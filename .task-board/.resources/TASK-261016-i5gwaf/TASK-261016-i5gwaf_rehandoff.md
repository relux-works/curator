# TASK-261016-i5gwaf — trunk convergence re-handoff

Ready for review. The converged README was read alongside the new agent-login reference, environment-config pointer and tools documentation. No duplicate, contradictory text or broken local links was found, so this run makes no further candidate edits. All inherited M2 changes remain uncommitted.

Candidate checkpoint: `dccac3199aa5c543c589c129f5ef2bc437523a2a`; `git rev-list --count HEAD..main` = 0. This supersedes the prior handoff's stale-worktree bound for these fresh checks; historical measurements in the qualification record retain their original checkpoint identity. Default runtime/test/config paths are byte-identical to the previous checkpoint (`git diff --exit-code 0d00a109ad1da77a5506747c89f24af9442eb341 -- cmd internal go.mod go.sum .github scripts Makefile`, exit 0). No commit, switch, rebase, merge or integration was performed on the Story branch.

Normative main was read at `15d970111277b1f68b3f7a4ddd63bc339b2c0a3f`: environments §§7.1 P1–P4, 7.9 and 10.2; CIP-0017 implementation leaf 2 and minimal probe items 1–4; Muse v4 conformance README. CIP-0017 and the conformance README are unchanged from the prior pinned read. New environment text concerns status inspection, store diagnostics and the successor carrier, outside M2. No conflict with M2's explicit separate credential/model evidence boundary was found. The independent reference CLI remains intentionally pinned to the record's `d2a942cf17e4eb017c45475442da1456d8dfbe6c` snapshot contract.

## Fresh commands and real exits

All gates ran directly with log redirection, without tee. Vendor/reference locators were supplied privately, as documented in `.research/muse148/README.md`:

```sh
MUSE148_BINARY="$(python3 -c 'import json; print(json.load(open(".temp/TASK-261016-i5gwaf/identity.json"))["binary_path"])')" MUSE148_CURATOR="$PWD/.temp/TASK-261016-i5gwaf/curator-rehandoff" MUSE148_RUN_ROOT="$PWD/.temp/TASK-261016-i5gwaf/replay-rehandoff-01" MUSE148_SPEC="$PWD/.temp/TASK-261016-i5gwaf/spec-main" MUSE148_SPEC_PYTHON="$PWD/.temp/TASK-261016-i5gwaf/reference-venv/bin/python" TMPDIR="$PWD/.temp/TASK-261016-i5gwaf/tmp" go test ./.research/muse148 -count=1 -v -timeout=5m
```

| Command | Exit | Evidence |
|---|---:|---|
| `go build -o .temp/TASK-261016-i5gwaf/curator-rehandoff ./cmd/curator` | 0 | build-rehandoff-01.log |
| Full opt-in command above | 0 | replay-rehandoff-01.log; no skips, 44 real vendor calls, 42 accepted exits 0 and 2 intended parser refusals exits 2 |
| `go build ./...` | 0 | build-all-rehandoff-01.log |
| `go vet ./.research/muse148` | 0 | vet-rehandoff-01.log |
| `golangci-lint run ./.research/muse148/...` | 0 | lint-rehandoff-01.log; 0 issues |
| `lychee --offline --include-fragments --no-progress README.md docs/environment-config.md docs/agent-logins.md docs/qualification/muse-1.4.8-R6256.1.md .research/muse148/README.md` | 0 | links-rehandoff-01.log; zero errors |
| `git diff --check` | 0 | diff-check-rehandoff-02.log |
| `MUSE148_OMIT_EXCLUSION=exec go test ./.research/muse148 -run '^TestMuse148LoaderMatrix$/exec/(linked\|snapshot)/skills' -count=1 -v` with fresh run root and documented locators | **1** | plant-exec-rehandoff-01.log; expected-red required-exclusion failure with user_skills=1, no-exclusion controls pass on both roots |
| `MUSE148_OMIT_EXCLUSION=interactive go test ./.research/muse148 -run '^TestMuse148LoaderMatrix$/interactive/(linked\|snapshot)/skills' -count=1 -v` with fresh run root and documented locators | 0 | plant-interactive-rehandoff-01.log; explicit survivor on both roots, not exclusion protection |
| `sh scripts/remote-gate.sh` | **1** | remote-gate-rehandoff-01.log; attempt 1: 19 success, 1 Windows failure, 2 configured optional skips |
| `gh run rerun 38869688515 --job 116870600742` | 0 | Dispatch only; reruns the failed Windows job at the unchanged snapshot, not proof of a passing test |
| `gh run watch 38869688515 --interval 30 --exit-status` | 0 | hosted-rerun-watch-rehandoff.log; attempt 2 completed/success, 20 successful jobs and 2 configured optional skips |

Hosted gate snapshot `a88ce7997152b97e99389b6bbf9303e5e210e75f` is author-signed (verified G); the existing task-local Git wrapper signs commit-tree while retaining configured human identity, as in the previous handoff. Story HEAD and index remain untouched. An independently rebuilt temporary index yielded exactly the hosted snapshot tree (comparison exit 0). Hosted run: https://github.com/relux-works/curator/actions/runs/38869688515. Unlike the previous rework's reuse, this full gate is freshly rerun after convergence; the hidden probe is separately exercised above.

One exploratory binary `cmp -s` exited 1: the fresh Curator embeds the converged module pseudo-version and VCS revision/time. `go version -m` confirmed those build-info differences; runtime/test/config source comparison exited 0. The fresh replay used the rebuilt binary, not the old executable. An initial unquoted GitHub API argument failed in shell glob expansion (exit 1); the quoted standalone reads succeeded. Neither diagnostic is a failed test or counted as passing validation.

## Identity, measured result and coverage

Exact Muse remains `1.4.8-R6256.1`, SHA-256 `5721763b4723f961709cd28f69aa049763965b46c4d4d70f883a0a90e5ddf39c`; macOS 27.0 arm64, Go 1.25.5. Its exact executable locator stays in ignored private `identity.json`. Probe, reviewer regression and reference checker SHA-256 remain respectively `78ecf817e292c2a7032bea08cf403dc9f2e86bd015eefdb3caf82abd02f94da1`, `bfedfd84093101834626e55476bb7361b80e8b9a955593da7f4cffc29f977103`, `524a8a00d2ce68be6084d3c87ac39a2f7a051cf3b064216fc5e7f3cca50dc1e6`. Frozen root/manifest/launch bytes are unchanged and revalidated by both the Go and pinned reference CLI tests.

Every Muse invocation uses offline echo, auto-update disabled, env -i, disposable HOME/all four XDG parents and tested OS network/Keychain deny. No login, logout, auth command or native store inspection ran. The local echo turn is expressly not remote inference or model evidence. Fresh sanitized argv, exits and loader records for all calls accompany this handoff.

**6 of 6 in-scope AC rows driven; 5 of 6 satisfy the loader contract.** Exec exclusion passes. Interactive skill exclusion FAILS on both representations; a passing regression witness records that vendor failure and does not authorize release. Section 7.9 interactive release gate stays pending.

| AC observation | Named test / production entrypoint | Result |
|---|---|---|
| Exec acceptance; independent interactive flag/string-false refusals | TestMuse148ParserRefusals, TestMuse148LoaderMatrix / real exec and TUI parsers | Expected exit 2 refusals before loader, accepted boolean/exec controls exit 0 |
| Root absent/present rules fallback, persistent skills | TestMuse148LoaderMatrix / real rules.context_load and skills.catalog_load | All separate sources and both exclusion states reproduce |
| Required exclusion with linked and snapshot root | Matrix, TestMuse148InteractiveExclusionRegression / real loader and banner | Exec 1→0; interactive 1→1 FAILURE on both |
| Interactive boolean without root | Matrix interactive/absent / real TUI loader | Rules fallback and foreign skill remain |
| Real Curator link and frozen regular snapshot | TestFrozenSnapshot, TestFrozenSnapshotSourceInputs, TestReferenceSnapshotPhysicalClosure, matrix / Curator profile install + env resolve, reference CLI, Muse loader | Real link, root identity and independent physical regular closure pass |
| Workspace residual | TestMuse148WorkspaceResidual / real loader and TUI /rules | Order/count transition reproduced; model-visible precedence unknown |

No catalog-keyed surface table was supplied (brief gap). Every supplied observation row remains mapped in the canonical record. Out-of-contract rows: credential/backend/login/refresh and model-visible removal (M2 Separate evidence); model-visible workspace precedence (M2 workspace residual); export provenance, capture locks, publication, destination admission and runtime-B (M2 frozen fixture boundary); actual env materialize→Muse (M2 separate SNAP-CLI row); runtime consumer refusals/qualification constants (M1/M3 ownership). No new exclusions from scope.

## Regression and mutation evidence

All prior test names remain. TestReviewContradictoryObservation, actual-source physical-fixture controls and independent CLI listed-symlink/dangling-link refusals passed freshly beside regular controls, asserting unchanged rejected inputs. The final-source eight patch-mutant results in `.research/muse148/TASK-261016-i5gwaf_review-rework.md` is reused, not rerun: source, tests, patches, reference contract, configuration and relevant toolchain/platform identities are unchanged. The exec and interactive omission plants were freshly rerun with the rebuilt Curator because its binary identity changed. These add eight real vendor calls to the 44-case green full replay. No static source scanner is present.

| Mutant | Gate narrowed to admit | Named failing test | Result / survivor bound |
|---|---|---|---|
| checker-duplicate-count | zero-then-one duplicate count | TestReviewContradictoryObservation/count-zero-then-one | Killed, exit 1 |
| checker-duplicate-outcome | failed-then-ready duplicate outcome | TestReviewContradictoryObservation/outcome-failed-then-ready | Killed, exit 1 |
| snapshot-go-listed-link | listed root symlink | TestFrozenSnapshot/listed-symlink and TestFrozenSnapshotSourceInputs/listed-symlink | Killed, exit 1 |
| snapshot-reference-listed-link | listed root symlink at real CLI | TestReferenceSnapshotPhysicalClosure/listed-symlink | Killed, exit 1 |
| snapshot-reference-dangling-link | omit unlisted dangling link | TestReferenceSnapshotPhysicalClosure/unlisted-dangling-link | Killed, exit 1 |
| checker-one-skill | one foreign skill as zero | TestObservationChecker | Killed, exit 1 |
| snapshot-extra-entry | ignore home/extra | TestFrozenSnapshot/extra | Killed, exit 1 |
| offline-loopback | outbound loopback only | TestMuse148OfflineBoundary | Killed, exit 1; no Muse started under mutant |
| exec omission | omit required exclusion | TestMuse148LoaderMatrix/exec/(linked\|snapshot)/skills/exclude-true | Freshly killed on both, exit 1; controls pass |
| interactive omission | omit required boolean | None — survivor | Fresh exit 0; vendor retains skill with or without boolean on both roots, so this cannot prove interactive protection |

No candidate paths changed during re-handoff. Existing out-of-module README/LOGBOOK/UNRESOLVED_QUESTIONS documentation remains justified by discoverability, measured anomaly and required logbook recording. No runtime constants, admission gates or credential policy changes.

## Full-gate attempt 1 diagnostic and logbook entry

The first fresh full gate is RED, exit 1. The only failed test is `internal/registry::TestSnapshotFutureBoundToleratesCheckerLatency/a_slow_sibling's_latency_is_inside_the_bound`, at registry_test.go:1071: the zero-skew future timestamp was accepted instead of refused. That group took 8.48s; it uses 1.5s sleeps and a 3s future timestamp, so scheduler latency is a plausible cause, not an established external CI outage. The full uploaded Windows Go JSON stream was downloaded and inspected; its only failing package is internal/registry. Platform-case gate itself exited 0, with 322 declared skips. The remote gate log's tail alone omitted the assertion, so it is not used as complete failure evidence.

The registry source/tests/configuration are unchanged from the prior checkpoint and outside M2. No registry code or timing bound was changed. One failed-job rerun is the recorded diagnostic reason for repeating Windows checks; the 19 successful jobs retain their legitimate hosted evidence at the same snapshot. This is neither a local mirror nor a synthetic status. While the rerun was pending, the full-gate and Tests green checklist items were unchecked. Attempt 2 now legitimately reports success on the unchanged snapshot, so those properties are green. The original shell full-gate command still has real exit 1; it is not relabeled 0. The first failure remains in the evidence; no timing-test fix or external-outage claim is made.

## Terminal hosted evidence and role handoff

Attempt 2: completed/success, exact head `a88ce7997152b97e99389b6bbf9303e5e210e75f`, 20 success and 2 configured optional skips. Only the failed Windows job was repeated; the hosting platform retained the other 19 green results on the unchanged source/test/configuration identity. The observer command exited 0. This is a legitimate hosted rerun, not a local mirror or synthetic status, and the original script exit 1 remains reported above.

The complete new Windows artifact `12037072271` was downloaded by its exact ID, independently of the identically named attempt-1 artifact `12034463164`. Commands `gh run view 38869688515 --json status,conclusion,attempt,headSha,url,jobs` and `gh api repos/relux-works/curator/actions/artifacts/12037072271/zip` both exited 0. The complete uploaded Go JSON stream has zero fail actions, and explicit pass actions for `TestSnapshotFutureBoundToleratesCheckerLatency/a_slow_sibling's_latency_is_inside_the_bound`, its parent, and `internal/registry`. Those pass records accompany this note. The first failure did not reproduce in this retry; its cause remains uncorrected outside this leaf, and it has not been erased from the evidence.

A fresh temporary-index comparison after both hosted attempts still equals the signed hosted snapshot tree (exit 0), and Story HEAD remains at the converged checkpoint with the inherited delta uncommitted. No code, tests, fixtures or tracked documentation changed in this re-handoff run. The findings/logbook entry above travels as this task-scoped outcome; no direct control-root file edit was made.

All five handoff preconditions are accounted for: no catalog surface table supplied (explicit brief gap and six-row AC coverage map above); hosted property suites green after the named retry; all prior reviewer regression names retained and freshly green; out-of-contract rows and clauses declared; rework diff bounded (no additional candidate edits). The persistent checklist and the developer-handoff acknowledgement are authoritative lifecycle records.

## Validation log appendix

Complete focused/plant logs follow. Empty build/vet/whitespace logs accompany exit codes in the command table; the full hosted Go streams remain in the canonical run artifacts and task-local ignored storage. The registry pass excerpt is explicitly selected from the complete attempt-2 stream, which was checked for zero failures. The optional native runner label is redacted in public hosted metadata.

### replay-rehandoff-01.log

```text
=== RUN   TestObservationChecker
--- PASS: TestObservationChecker (0.00s)
=== RUN   TestMuse148OfflineBoundary
--- PASS: TestMuse148OfflineBoundary (0.26s)
=== RUN   TestFrozenSnapshot
=== RUN   TestFrozenSnapshot/regular
=== RUN   TestFrozenSnapshot/altered
=== RUN   TestFrozenSnapshot/extra
=== RUN   TestFrozenSnapshot/listed-symlink
=== RUN   TestFrozenSnapshot/unlisted-dangling-link
=== RUN   TestFrozenSnapshot/missing
--- PASS: TestFrozenSnapshot (0.04s)
    --- PASS: TestFrozenSnapshot/regular (0.03s)
    --- PASS: TestFrozenSnapshot/altered (0.01s)
    --- PASS: TestFrozenSnapshot/extra (0.00s)
    --- PASS: TestFrozenSnapshot/listed-symlink (0.00s)
    --- PASS: TestFrozenSnapshot/unlisted-dangling-link (0.00s)
    --- PASS: TestFrozenSnapshot/missing (0.00s)
=== RUN   TestFrozenSnapshotSourceInputs
=== RUN   TestFrozenSnapshotSourceInputs/regular
=== RUN   TestFrozenSnapshotSourceInputs/listed-symlink
=== RUN   TestFrozenSnapshotSourceInputs/unlisted-dangling-link
--- PASS: TestFrozenSnapshotSourceInputs (0.02s)
    --- PASS: TestFrozenSnapshotSourceInputs/regular (0.01s)
    --- PASS: TestFrozenSnapshotSourceInputs/listed-symlink (0.01s)
    --- PASS: TestFrozenSnapshotSourceInputs/unlisted-dangling-link (0.01s)
=== RUN   TestReferenceSnapshotPhysicalClosure
=== RUN   TestReferenceSnapshotPhysicalClosure/regular
=== RUN   TestReferenceSnapshotPhysicalClosure/listed-symlink
=== RUN   TestReferenceSnapshotPhysicalClosure/unlisted-dangling-link
--- PASS: TestReferenceSnapshotPhysicalClosure (1.10s)
    --- PASS: TestReferenceSnapshotPhysicalClosure/regular (0.61s)
    --- PASS: TestReferenceSnapshotPhysicalClosure/listed-symlink (0.24s)
    --- PASS: TestReferenceSnapshotPhysicalClosure/unlisted-dangling-link (0.24s)
=== RUN   TestMuse148ParserRefusals
=== RUN   TestMuse148ParserRefusals/interactive-exec-flag
=== RUN   TestMuse148ParserRefusals/interactive-string-false
--- PASS: TestMuse148ParserRefusals (0.37s)
    --- PASS: TestMuse148ParserRefusals/interactive-exec-flag (0.13s)
    --- PASS: TestMuse148ParserRefusals/interactive-string-false (0.06s)
=== RUN   TestMuse148LoaderMatrix
=== RUN   TestMuse148LoaderMatrix/exec/absent/claude-rules/exclude-false
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=1 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/absent/claude-rules/exclude-true
    replay_test.go:790: vendor exit=0; user_sources=0 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/absent/codex-rules/exclude-false
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=1 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/absent/codex-rules/exclude-true
    replay_test.go:790: vendor exit=0; user_sources=0 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/absent/skills/exclude-false
    replay_test.go:790: vendor exit=0; user_sources=0 foreign_sources=0 user_skills=1; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/absent/skills/exclude-true
    replay_test.go:790: vendor exit=0; user_sources=0 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/linked/claude-rules/exclude-false
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/linked/claude-rules/exclude-true
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/linked/codex-rules/exclude-false
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/linked/codex-rules/exclude-true
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/linked/skills/exclude-false
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=1; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/linked/skills/exclude-true
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/snapshot/claude-rules/exclude-false
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/snapshot/claude-rules/exclude-true
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/snapshot/codex-rules/exclude-false
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/snapshot/codex-rules/exclude-true
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/snapshot/skills/exclude-false
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=1; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/snapshot/skills/exclude-true
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/absent/claude-rules/exclude-false
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=1 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/absent/claude-rules/exclude-true
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=1 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/absent/codex-rules/exclude-false
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=1 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/absent/codex-rules/exclude-true
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=1 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/absent/skills/exclude-false
    replay_test.go:790: vendor exit=0; user_sources=0 foreign_sources=0 user_skills=1; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/absent/skills/exclude-true
    replay_test.go:790: vendor exit=0; user_sources=0 foreign_sources=0 user_skills=1; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/linked/claude-rules/exclude-false
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/linked/claude-rules/exclude-true
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/linked/codex-rules/exclude-false
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/linked/codex-rules/exclude-true
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/linked/skills/exclude-false
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=1; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/linked/skills/exclude-true
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=1; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/snapshot/claude-rules/exclude-false
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/snapshot/claude-rules/exclude-true
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/snapshot/codex-rules/exclude-false
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/snapshot/codex-rules/exclude-true
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/snapshot/skills/exclude-false
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=1; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/snapshot/skills/exclude-true
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=1; no model claim
--- PASS: TestMuse148LoaderMatrix (54.35s)
    --- PASS: TestMuse148LoaderMatrix/exec/absent/claude-rules/exclude-false (1.06s)
    --- PASS: TestMuse148LoaderMatrix/exec/absent/claude-rules/exclude-true (0.87s)
    --- PASS: TestMuse148LoaderMatrix/exec/absent/codex-rules/exclude-false (1.09s)
    --- PASS: TestMuse148LoaderMatrix/exec/absent/codex-rules/exclude-true (0.83s)
    --- PASS: TestMuse148LoaderMatrix/exec/absent/skills/exclude-false (0.87s)
    --- PASS: TestMuse148LoaderMatrix/exec/absent/skills/exclude-true (0.91s)
    --- PASS: TestMuse148LoaderMatrix/exec/linked/claude-rules/exclude-false (0.85s)
    --- PASS: TestMuse148LoaderMatrix/exec/linked/claude-rules/exclude-true (1.06s)
    --- PASS: TestMuse148LoaderMatrix/exec/linked/codex-rules/exclude-false (0.92s)
    --- PASS: TestMuse148LoaderMatrix/exec/linked/codex-rules/exclude-true (0.93s)
    --- PASS: TestMuse148LoaderMatrix/exec/linked/skills/exclude-false (0.98s)
    --- PASS: TestMuse148LoaderMatrix/exec/linked/skills/exclude-true (1.05s)
    --- PASS: TestMuse148LoaderMatrix/exec/snapshot/claude-rules/exclude-false (0.96s)
    --- PASS: TestMuse148LoaderMatrix/exec/snapshot/claude-rules/exclude-true (1.01s)
    --- PASS: TestMuse148LoaderMatrix/exec/snapshot/codex-rules/exclude-false (0.92s)
    --- PASS: TestMuse148LoaderMatrix/exec/snapshot/codex-rules/exclude-true (1.09s)
    --- PASS: TestMuse148LoaderMatrix/exec/snapshot/skills/exclude-false (1.04s)
    --- PASS: TestMuse148LoaderMatrix/exec/snapshot/skills/exclude-true (1.02s)
    --- PASS: TestMuse148LoaderMatrix/interactive/absent/claude-rules/exclude-false (1.65s)
    --- PASS: TestMuse148LoaderMatrix/interactive/absent/claude-rules/exclude-true (1.97s)
    --- PASS: TestMuse148LoaderMatrix/interactive/absent/codex-rules/exclude-false (1.96s)
    --- PASS: TestMuse148LoaderMatrix/interactive/absent/codex-rules/exclude-true (2.07s)
    --- PASS: TestMuse148LoaderMatrix/interactive/absent/skills/exclude-false (1.76s)
    --- PASS: TestMuse148LoaderMatrix/interactive/absent/skills/exclude-true (1.97s)
    --- PASS: TestMuse148LoaderMatrix/interactive/linked/claude-rules/exclude-false (1.96s)
    --- PASS: TestMuse148LoaderMatrix/interactive/linked/claude-rules/exclude-true (1.96s)
    --- PASS: TestMuse148LoaderMatrix/interactive/linked/codex-rules/exclude-false (1.87s)
    --- PASS: TestMuse148LoaderMatrix/interactive/linked/codex-rules/exclude-true (1.98s)
    --- PASS: TestMuse148LoaderMatrix/interactive/linked/skills/exclude-false (1.98s)
    --- PASS: TestMuse148LoaderMatrix/interactive/linked/skills/exclude-true (2.18s)
    --- PASS: TestMuse148LoaderMatrix/interactive/snapshot/claude-rules/exclude-false (2.07s)
    --- PASS: TestMuse148LoaderMatrix/interactive/snapshot/claude-rules/exclude-true (2.08s)
    --- PASS: TestMuse148LoaderMatrix/interactive/snapshot/codex-rules/exclude-false (1.89s)
    --- PASS: TestMuse148LoaderMatrix/interactive/snapshot/codex-rules/exclude-true (2.02s)
    --- PASS: TestMuse148LoaderMatrix/interactive/snapshot/skills/exclude-false (1.99s)
    --- PASS: TestMuse148LoaderMatrix/interactive/snapshot/skills/exclude-true (1.98s)
=== RUN   TestMuse148InteractiveExclusionRegression
=== RUN   TestMuse148InteractiveExclusionRegression/linked
    replay_test.go:820: vendor exit=0; required interactive exclusion FAILED (user_skills=1); see docs/qualification/muse-1.4.8-R6256.1.md
=== RUN   TestMuse148InteractiveExclusionRegression/snapshot
    replay_test.go:820: vendor exit=0; required interactive exclusion FAILED (user_skills=1); see docs/qualification/muse-1.4.8-R6256.1.md
--- PASS: TestMuse148InteractiveExclusionRegression (5.05s)
    --- PASS: TestMuse148InteractiveExclusionRegression/linked (2.07s)
    --- PASS: TestMuse148InteractiveExclusionRegression/snapshot (2.07s)
=== RUN   TestMuse148WorkspaceResidual
=== RUN   TestMuse148WorkspaceResidual/exec/linked
=== RUN   TestMuse148WorkspaceResidual/exec/snapshot
=== RUN   TestMuse148WorkspaceResidual/interactive/linked
=== RUN   TestMuse148WorkspaceResidual/interactive/snapshot
--- PASS: TestMuse148WorkspaceResidual (8.12s)
    --- PASS: TestMuse148WorkspaceResidual/exec/linked (0.99s)
    --- PASS: TestMuse148WorkspaceResidual/exec/snapshot (1.11s)
    --- PASS: TestMuse148WorkspaceResidual/interactive/linked (2.95s)
    --- PASS: TestMuse148WorkspaceResidual/interactive/snapshot (2.34s)
=== RUN   TestReviewContradictoryObservation
=== RUN   TestReviewContradictoryObservation/duplicate-count
=== RUN   TestReviewContradictoryObservation/count-zero-then-one
=== RUN   TestReviewContradictoryObservation/outcome-failed-then-ready
=== RUN   TestReviewContradictoryObservation/duplicate-outcome
=== RUN   TestReviewContradictoryObservation/duplicate-event
=== RUN   TestReviewContradictoryObservation/quoted-decoy
=== RUN   TestReviewContradictoryObservation/malformed-count
=== RUN   TestReviewContradictoryObservation/overflow-count
=== RUN   TestReviewContradictoryObservation/count-one-then-zero
=== RUN   TestReviewContradictoryObservation/outcome-ready-then-failed
--- PASS: TestReviewContradictoryObservation (0.00s)
    --- PASS: TestReviewContradictoryObservation/duplicate-count (0.00s)
    --- PASS: TestReviewContradictoryObservation/count-zero-then-one (0.00s)
    --- PASS: TestReviewContradictoryObservation/outcome-failed-then-ready (0.00s)
    --- PASS: TestReviewContradictoryObservation/duplicate-outcome (0.00s)
    --- PASS: TestReviewContradictoryObservation/duplicate-event (0.00s)
    --- PASS: TestReviewContradictoryObservation/quoted-decoy (0.00s)
    --- PASS: TestReviewContradictoryObservation/malformed-count (0.00s)
    --- PASS: TestReviewContradictoryObservation/overflow-count (0.00s)
    --- PASS: TestReviewContradictoryObservation/count-one-then-zero (0.00s)
    --- PASS: TestReviewContradictoryObservation/outcome-ready-then-failed (0.00s)
PASS
ok  	github.com/relux-works/curator/.research/muse148	69.679s

```

### plant-exec-rehandoff-01.log

```text
=== RUN   TestMuse148LoaderMatrix
=== RUN   TestMuse148LoaderMatrix/exec/linked/skills/exclude-false
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=1; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/linked/skills/exclude-true
    replay_test.go:788: skills.catalog_load observed user_skills=1; want 0
=== RUN   TestMuse148LoaderMatrix/exec/snapshot/skills/exclude-false
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=1; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/snapshot/skills/exclude-true
    replay_test.go:788: skills.catalog_load observed user_skills=1; want 0
--- FAIL: TestMuse148LoaderMatrix (6.25s)
    --- PASS: TestMuse148LoaderMatrix/exec/linked/skills/exclude-false (1.72s)
    --- FAIL: TestMuse148LoaderMatrix/exec/linked/skills/exclude-true (0.99s)
    --- PASS: TestMuse148LoaderMatrix/exec/snapshot/skills/exclude-false (1.41s)
    --- FAIL: TestMuse148LoaderMatrix/exec/snapshot/skills/exclude-true (1.04s)
FAIL
FAIL	github.com/relux-works/curator/.research/muse148	6.567s
FAIL

```

### plant-interactive-rehandoff-01.log

```text
=== RUN   TestMuse148LoaderMatrix
=== RUN   TestMuse148LoaderMatrix/interactive/linked/skills/exclude-false
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=1; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/linked/skills/exclude-true
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=1; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/snapshot/skills/exclude-false
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=1; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/snapshot/skills/exclude-true
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=1; no model claim
--- PASS: TestMuse148LoaderMatrix (9.90s)
    --- PASS: TestMuse148LoaderMatrix/interactive/linked/skills/exclude-false (2.17s)
    --- PASS: TestMuse148LoaderMatrix/interactive/linked/skills/exclude-true (2.72s)
    --- PASS: TestMuse148LoaderMatrix/interactive/snapshot/skills/exclude-false (2.10s)
    --- PASS: TestMuse148LoaderMatrix/interactive/snapshot/skills/exclude-true (1.84s)
PASS
ok  	github.com/relux-works/curator/.research/muse148	10.191s

```

### build-rehandoff-01.log

```text

```

### build-all-rehandoff-01.log

```text

```

### vet-rehandoff-01.log

```text

```

### lint-rehandoff-01.log

```text
0 issues.

```

### links-rehandoff-01.log

```text
🔍 59 Total (in 15ms) 🔗 52 Unique ✅ 42 OK 🚫 0 Errors 👻 17 Excluded


```

### diff-check-rehandoff-02.log

```text

```

### default-source-convergence-identity.log

```text

```

### source-identity-rehandoff.log

```text
78ecf817e292c2a7032bea08cf403dc9f2e86bd015eefdb3caf82abd02f94da1  .research/muse148/replay_test.go
bfedfd84093101834626e55476bb7361b80e8b9a955593da7f4cffc29f977103  .research/muse148/review_attack_test.go
524a8a00d2ce68be6084d3c87ac39a2f7a051cf3b064216fc5e7f3cca50dc1e6  .research/muse148/reference-snapshot-check.py

```

### platform-rehandoff.log

```text
27.0
arm64

```

### registry-attempt2-pass.json

```text
[
  {
    "Time": "2026-10-16T08:53:15.9713308Z",
    "Action": "pass",
    "Package": "github.com/relux-works/curator/internal/registry",
    "Test": "TestSnapshotFutureBoundToleratesCheckerLatency/mint_during_a_slow_fetch_is_accepted_under_zero_skew",
    "Elapsed": 1.56
  },
  {
    "Time": "2026-10-16T08:53:17.5668694Z",
    "Action": "pass",
    "Package": "github.com/relux-works/curator/internal/registry",
    "Test": "TestSnapshotFutureBoundToleratesCheckerLatency/a_slow_sibling's_latency_is_inside_the_bound/behind_the_post-fetch_clock_is_accepted",
    "Elapsed": 1.6
  },
  {
    "Time": "2026-10-16T08:53:20.6735743Z",
    "Action": "pass",
    "Package": "github.com/relux-works/curator/internal/registry",
    "Test": "TestSnapshotFutureBoundToleratesCheckerLatency/a_slow_sibling's_latency_is_inside_the_bound",
    "Elapsed": 4.7
  },
  {
    "Time": "2026-10-16T08:53:20.6988841Z",
    "Action": "pass",
    "Package": "github.com/relux-works/curator/internal/registry",
    "Test": "TestSnapshotFutureBoundToleratesCheckerLatency/now_plus_skew_plus_two_seconds_still_refuses",
    "Elapsed": 0.03
  },
  {
    "Time": "2026-10-16T08:53:20.6988841Z",
    "Action": "pass",
    "Package": "github.com/relux-works/curator/internal/registry",
    "Test": "TestSnapshotFutureBoundToleratesCheckerLatency",
    "Elapsed": 6.28
  }
]

```

### hosted-rehandoff-attempt2.json

```text
{"attempt":2,"conclusion":"success","headSha":"a88ce7997152b97e99389b6bbf9303e5e210e75f","jobs":[{"completedAt":"2026-10-16T09:23:17Z","conclusion":"success","databaseId":116886088789,"name":"Test (windows-latest)","startedAt":"2026-10-16T08:33:28Z","status":"completed","steps":[{"completedAt":"2026-10-16T08:33:34Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-16T08:33:29Z","status":"completed"},{"completedAt":"2026-10-16T08:33:58Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-16T08:33:34Z","status":"completed"},{"completedAt":"2026-10-16T08:34:02Z","conclusion":"success","name":"Checkout authoritative protocol suite","number":3,"startedAt":"2026-10-16T08:33:58Z","status":"completed"},{"completedAt":"2026-10-16T08:34:39Z","conclusion":"success","name":"Run actions/setup-go@v5","number":4,"startedAt":"2026-10-16T08:34:02Z","status":"completed"},{"completedAt":"2026-10-16T08:34:44Z","conclusion":"success","name":"Run actions/setup-node@v4","number":5,"startedAt":"2026-10-16T08:34:39Z","status":"completed"},{"completedAt":"2026-10-16T08:34:44Z","conclusion":"success","name":"Verify pnpm pin matches Go constant","number":6,"startedAt":"2026-10-16T08:34:44Z","status":"completed"},{"completedAt":"2026-10-16T08:34:50Z","conclusion":"success","name":"Install pinned pnpm via npm","number":7,"startedAt":"2026-10-16T08:34:44Z","status":"completed"},{"completedAt":"2026-10-16T08:34:51Z","conclusion":"success","name":"Verify rust pin matches Go constant","number":8,"startedAt":"2026-10-16T08:34:50Z","status":"completed"},{"completedAt":"2026-10-16T08:35:01Z","conclusion":"success","name":"Install pinned Rust toolchain via rustup","number":9,"startedAt":"2026-10-16T08:34:51Z","status":"completed"},{"completedAt":"2026-10-16T08:35:01Z","conclusion":"success","name":"Verify Go toolchain identity","number":10,"startedAt":"2026-10-16T08:35:01Z","status":"completed"},{"completedAt":"2026-10-16T08:35:01Z","conclusion":"success","name":"Verify versioned Go install compatibility","number":11,"startedAt":"2026-10-16T08:35:01Z","status":"completed"},{"completedAt":"2026-10-16T08:35:01Z","conclusion":"skipped","name":"gofmt check","number":12,"startedAt":"2026-10-16T08:35:01Z","status":"completed"},{"completedAt":"2026-10-16T08:35:20Z","conclusion":"success","name":"go vet","number":13,"startedAt":"2026-10-16T08:35:01Z","status":"completed"},{"completedAt":"2026-10-16T08:36:07Z","conclusion":"success","name":"Ledger consistency","number":14,"startedAt":"2026-10-16T08:35:20Z","status":"completed"},{"completedAt":"2026-10-16T09:23:08Z","conclusion":"success","name":"go test + platform-case gate","number":15,"startedAt":"2026-10-16T08:36:07Z","status":"completed"},{"completedAt":"2026-10-16T09:23:09Z","conclusion":"success","name":"Upload gate evidence","number":16,"startedAt":"2026-10-16T09:23:08Z","status":"completed"},{"completedAt":"2026-10-16T09:23:10Z","conclusion":"success","name":"Post Run actions/setup-node@v4","number":29,"startedAt":"2026-10-16T09:23:09Z","status":"completed"},{"completedAt":"2026-10-16T09:23:10Z","conclusion":"success","name":"Post Run actions/setup-go@v5","number":30,"startedAt":"2026-10-16T09:23:10Z","status":"completed"},{"completedAt":"2026-10-16T09:23:12Z","conclusion":"success","name":"Post Checkout authoritative protocol suite","number":31,"startedAt":"2026-10-16T09:23:10Z","status":"completed"},{"completedAt":"2026-10-16T09:23:15Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":32,"startedAt":"2026-10-16T09:23:12Z","status":"completed"},{"completedAt":"2026-10-16T09:23:15Z","conclusion":"success","name":"Complete job","number":33,"startedAt":"2026-10-16T09:23:15Z","status":"completed"}],"url":"https://github.com/relux-works/curator/actions/runs/38869688515/job/116886088789"},{"completedAt":"2026-10-16T07:49:50Z","conclusion":"skipped","databaseId":116886089253,"name":"Test (optional native platform)","startedAt":"2026-10-16T08:33:24Z","status":"completed","steps":[],"url":"https://github.com/relux-works/curator/actions/runs/38869688515/job/116886089253"},{"completedAt":"2026-10-16T07:50:41Z","conclusion":"success","databaseId":116886089756,"name":"Go 1.25.5 driver (ubuntu-latest)","startedAt":"2026-10-16T07:49:53Z","status":"completed","steps":[{"completedAt":"2026-10-16T07:49:55Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-16T07:49:54Z","status":"completed"},{"completedAt":"2026-10-16T07:50:04Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-16T07:49:55Z","status":"completed"},{"completedAt":"2026-10-16T07:50:05Z","conclusion":"success","name":"Checkout authoritative protocol suite","number":3,"startedAt":"2026-10-16T07:50:04Z","status":"completed"},{"completedAt":"2026-10-16T07:50:17Z","conclusion":"success","name":"Run actions/setup-go@v5","number":4,"startedAt":"2026-10-16T07:50:05Z","status":"completed"},{"completedAt":"2026-10-16T07:50:17Z","conclusion":"success","name":"Verify family toolchain identity","number":5,"startedAt":"2026-10-16T07:50:17Z","status":"completed"},{"completedAt":"2026-10-16T07:50:22Z","conclusion":"success","name":"Driver qualification and reproducibility","number":6,"startedAt":"2026-10-16T07:50:17Z","status":"completed"},{"completedAt":"2026-10-16T07:50:24Z","conclusion":"success","name":"Driver vet","number":7,"startedAt":"2026-10-16T07:50:22Z","status":"completed"},{"completedAt":"2026-10-16T07:50:37Z","conclusion":"success","name":"Build manager","number":8,"startedAt":"2026-10-16T07:50:24Z","status":"completed"},{"completedAt":"2026-10-16T07:50:37Z","conclusion":"success","name":"Post Run actions/setup-go@v5","number":14,"startedAt":"2026-10-16T07:50:37Z","status":"completed"},{"completedAt":"2026-10-16T07:50:37Z","conclusion":"success","name":"Post Checkout authoritative protocol suite","number":15,"startedAt":"2026-10-16T07:50:37Z","status":"completed"},{"completedAt":"2026-10-16T07:50:38Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":16,"startedAt":"2026-10-16T07:50:37Z","status":"completed"},{"completedAt":"2026-10-16T07:50:38Z","conclusion":"success","name":"Complete job","number":17,"startedAt":"2026-10-16T07:50:38Z","status":"completed"}],"url":"https://github.com/relux-works/curator/actions/runs/38869688515/job/116886089756"},{"completedAt":"2026-10-16T07:49:50Z","conclusion":"skipped","databaseId":116886090010,"name":"Candidate suite (${{ matrix.os }})","startedAt":"2026-10-16T08:33:24Z","status":"completed","steps":[],"url":"https://github.com/relux-works/curator/actions/runs/38869688515/job/116886090010"},{"completedAt":"2026-10-16T07:51:33Z","conclusion":"success","databaseId":116886090219,"name":"Gate self-test (windows-latest)","startedAt":"2026-10-16T07:50:22Z","status":"completed","steps":[{"completedAt":"2026-10-16T07:50:23Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-16T07:50:22Z","status":"completed"},{"completedAt":"2026-10-16T07:50:38Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-16T07:50:23Z","status":"completed"},{"completedAt":"2026-10-16T07:51:29Z","conclusion":"success","name":"Self-test the CI gates","number":3,"startedAt":"2026-10-16T07:50:38Z","status":"completed"},{"completedAt":"2026-10-16T07:51:31Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":6,"startedAt":"2026-10-16T07:51:29Z","status":"completed"},{"completedAt":"2026-10-16T07:51:31Z","conclusion":"success","name":"Complete job","number":7,"startedAt":"2026-10-16T07:51:31Z","status":"completed"}],"url":"https://github.com/relux-works/curator/actions/runs/38869688515/job/116886090219"},{"completedAt":"2026-10-16T07:53:52Z","conclusion":"success","databaseId":116886090295,"name":"Go 1.27.1 driver (macos-latest)","startedAt":"2026-10-16T07:51:09Z","status":"completed","steps":[{"completedAt":"2026-10-16T07:51:11Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-16T07:51:09Z","status":"completed"},{"completedAt":"2026-10-16T07:51:21Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-16T07:51:11Z","status":"completed"},{"completedAt":"2026-10-16T07:51:22Z","conclusion":"success","name":"Checkout authoritative protocol suite","number":3,"startedAt":"2026-10-16T07:51:21Z","status":"completed"},{"completedAt":"2026-10-16T07:51:48Z","conclusion":"success","name":"Run actions/setup-go@v5","number":4,"startedAt":"2026-10-16T07:51:22Z","status":"completed"},{"completedAt":"2026-10-16T07:51:48Z","conclusion":"success","name":"Verify family toolchain identity","number":5,"startedAt":"2026-10-16T07:51:48Z","status":"completed"},{"completedAt":"2026-10-16T07:53:27Z","conclusion":"success","name":"Driver qualification and reproducibility","number":6,"startedAt":"2026-10-16T07:51:48Z","status":"completed"},{"completedAt":"2026-10-16T07:53:29Z","conclusion":"success","name":"Driver vet","number":7,"startedAt":"2026-10-16T07:53:27Z","status":"completed"},{"completedAt":"2026-10-16T07:53:44Z","conclusion":"success","name":"Build manager","number":8,"startedAt":"2026-10-16T07:53:29Z","status":"completed"},{"completedAt":"2026-10-16T07:53:45Z","conclusion":"success","name":"Post Run actions/setup-go@v5","number":14,"startedAt":"2026-10-16T07:53:44Z","status":"completed"},{"completedAt":"2026-10-16T07:53:45Z","conclusion":"success","name":"Post Checkout authoritative protocol suite","number":15,"startedAt":"2026-10-16T07:53:45Z","status":"completed"},{"completedAt":"2026-10-16T07:53:46Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":16,"startedAt":"2026-10-16T07:53:45Z","status":"completed"},{"completedAt":"2026-10-16T07:53:47Z","conclusion":"success","name":"Complete job","number":17,"startedAt":"2026-10-16T07:53:46Z","status":"completed"}],"url":"https://github.com/relux-works/curator/actions/runs/38869688515/job/116886090295"},{"completedAt":"2026-10-16T08:14:55Z","conclusion":"success","databaseId":116886090346,"name":"Go 1.25.5 driver (macos-latest)","startedAt":"2026-10-16T08:13:10Z","status":"completed","steps":[{"completedAt":"2026-10-16T08:13:11Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-16T08:13:10Z","status":"completed"},{"completedAt":"2026-10-16T08:13:20Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-16T08:13:11Z","status":"completed"},{"completedAt":"2026-10-16T08:13:21Z","conclusion":"success","name":"Checkout authoritative protocol suite","number":3,"startedAt":"2026-10-16T08:13:20Z","status":"completed"},{"completedAt":"2026-10-16T08:13:40Z","conclusion":"success","name":"Run actions/setup-go@v5","number":4,"startedAt":"2026-10-16T08:13:21Z","status":"completed"},{"completedAt":"2026-10-16T08:13:41Z","conclusion":"success","name":"Verify family toolchain identity","number":5,"startedAt":"2026-10-16T08:13:40Z","status":"completed"},{"completedAt":"2026-10-16T08:14:39Z","conclusion":"success","name":"Driver qualification and reproducibility","number":6,"startedAt":"2026-10-16T08:13:41Z","status":"completed"},{"completedAt":"2026-10-16T08:14:40Z","conclusion":"success","name":"Driver vet","number":7,"startedAt":"2026-10-16T08:14:39Z","status":"completed"},{"completedAt":"2026-10-16T08:14:50Z","conclusion":"success","name":"Build manager","number":8,"startedAt":"2026-10-16T08:14:40Z","status":"completed"},{"completedAt":"2026-10-16T08:14:50Z","conclusion":"success","name":"Post Run actions/setup-go@v5","number":14,"startedAt":"2026-10-16T08:14:50Z","status":"completed"},{"completedAt":"2026-10-16T08:14:51Z","conclusion":"success","name":"Post Checkout authoritative protocol suite","number":15,"startedAt":"2026-10-16T08:14:50Z","status":"completed"},{"completedAt":"2026-10-16T08:14:51Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":16,"startedAt":"2026-10-16T08:14:51Z","status":"completed"},{"completedAt":"2026-10-16T08:14:52Z","conclusion":"success","name":"Complete job","number":17,"startedAt":"2026-10-16T08:14:51Z","status":"completed"}],"url":"https://github.com/relux-works/curator/actions/runs/38869688515/job/116886090346"},{"completedAt":"2026-10-16T07:53:33Z","conclusion":"success","databaseId":116886090419,"name":"Go 1.25.5 driver (windows-latest)","startedAt":"2026-10-16T07:49:52Z","status":"completed","steps":[{"completedAt":"2026-10-16T07:49:54Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-16T07:49:53Z","status":"completed"},{"completedAt":"2026-10-16T07:50:11Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-16T07:49:54Z","status":"completed"},{"completedAt":"2026-10-16T07:50:14Z","conclusion":"success","name":"Checkout authoritative protocol suite","number":3,"startedAt":"2026-10-16T07:50:11Z","status":"completed"},{"completedAt":"2026-10-16T07:50:53Z","conclusion":"success","name":"Run actions/setup-go@v5","number":4,"startedAt":"2026-10-16T07:50:14Z","status":"completed"},{"completedAt":"2026-10-16T07:50:53Z","conclusion":"success","name":"Verify family toolchain identity","number":5,"startedAt":"2026-10-16T07:50:53Z","status":"completed"},{"completedAt":"2026-10-16T07:53:06Z","conclusion":"success","name":"Driver qualification and reproducibility","number":6,"startedAt":"2026-10-16T07:50:53Z","status":"completed"},{"completedAt":"2026-10-16T07:53:10Z","conclusion":"success","name":"Driver vet","number":7,"startedAt":"2026-10-16T07:53:06Z","status":"completed"},{"completedAt":"2026-10-16T07:53:26Z","conclusion":"success","name":"Build manager","number":8,"startedAt":"2026-10-16T07:53:10Z","status":"completed"},{"completedAt":"2026-10-16T07:53:26Z","conclusion":"success","name":"Post Run actions/setup-go@v5","number":14,"startedAt":"2026-10-16T07:53:26Z","status":"completed"},{"completedAt":"2026-10-16T07:53:29Z","conclusion":"success","name":"Post Checkout authoritative protocol suite","number":15,"startedAt":"2026-10-16T07:53:26Z","status":"completed"},{"completedAt":"2026-10-16T07:53:31Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":16,"startedAt":"2026-10-16T07:53:29Z","status":"completed"},{"completedAt":"2026-10-16T07:53:31Z","conclusion":"success","name":"Complete job","number":17,"startedAt":"2026-10-16T07:53:31Z","status":"completed"}],"url":"https://github.com/relux-works/curator/actions/runs/38869688515/job/116886090419"},{"completedAt":"2026-10-16T07:50:31Z","conclusion":"success","databaseId":116886090432,"name":"Go 1.27.1 driver (ubuntu-latest)","startedAt":"2026-10-16T07:49:52Z","status":"completed","steps":[{"completedAt":"2026-10-16T07:49:54Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-16T07:49:53Z","status":"completed"},{"completedAt":"2026-10-16T07:50:00Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-16T07:49:54Z","status":"completed"},{"completedAt":"2026-10-16T07:50:02Z","conclusion":"success","name":"Checkout authoritative protocol suite","number":3,"startedAt":"2026-10-16T07:50:00Z","status":"completed"},{"completedAt":"2026-10-16T07:50:14Z","conclusion":"success","name":"Run actions/setup-go@v5","number":4,"startedAt":"2026-10-16T07:50:02Z","status":"completed"},{"completedAt":"2026-10-16T07:50:14Z","conclusion":"success","name":"Verify family toolchain identity","number":5,"startedAt":"2026-10-16T07:50:14Z","status":"completed"},{"completedAt":"2026-10-16T07:50:17Z","conclusion":"success","name":"Driver qualification and reproducibility","number":6,"startedAt":"2026-10-16T07:50:14Z","status":"completed"},{"completedAt":"2026-10-16T07:50:19Z","conclusion":"success","name":"Driver vet","number":7,"startedAt":"2026-10-16T07:50:17Z","status":"completed"},{"completedAt":"2026-10-16T07:50:28Z","conclusion":"success","name":"Build manager","number":8,"startedAt":"2026-10-16T07:50:19Z","status":"completed"},{"completedAt":"2026-10-16T07:50:28Z","conclusion":"success","name":"Post Run actions/setup-go@v5","number":14,"startedAt":"2026-10-16T07:50:28Z","status":"completed"},{"completedAt":"2026-10-16T07:50:28Z","conclusion":"success","name":"Post Checkout authoritative protocol suite","number":15,"startedAt":"2026-10-16T07:50:28Z","status":"completed"},{"completedAt":"2026-10-16T07:50:28Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":16,"startedAt":"2026-10-16T07:50:28Z","status":"completed"},{"completedAt":"2026-10-16T07:50:28Z","conclusion":"success","name":"Complete job","number":17,"startedAt":"2026-10-16T07:50:28Z","status":"completed"}],"url":"https://github.com/relux-works/curator/actions/runs/38869688515/job/116886090432"},{"completedAt":"2026-10-16T08:10:32Z","conclusion":"success","databaseId":116886090523,"name":"Test (ubuntu-latest)","startedAt":"2026-10-16T07:49:53Z","status":"completed","steps":[{"completedAt":"2026-10-16T07:49:55Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-16T07:49:53Z","status":"completed"},{"completedAt":"2026-10-16T07:50:03Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-16T07:49:55Z","status":"completed"},{"completedAt":"2026-10-16T07:50:04Z","conclusion":"success","name":"Checkout authoritative protocol suite","number":3,"startedAt":"2026-10-16T07:50:03Z","status":"completed"},{"completedAt":"2026-10-16T07:50:12Z","conclusion":"success","name":"Run actions/setup-go@v5","number":4,"startedAt":"2026-10-16T07:50:04Z","status":"completed"},{"completedAt":"2026-10-16T07:50:13Z","conclusion":"success","name":"Run actions/setup-node@v4","number":5,"startedAt":"2026-10-16T07:50:12Z","status":"completed"},{"completedAt":"2026-10-16T07:50:13Z","conclusion":"success","name":"Verify pnpm pin matches Go constant","number":6,"startedAt":"2026-10-16T07:50:13Z","status":"completed"},{"completedAt":"2026-10-16T07:50:14Z","conclusion":"success","name":"Install pinned pnpm via npm","number":7,"startedAt":"2026-10-16T07:50:13Z","status":"completed"},{"completedAt":"2026-10-16T07:50:14Z","conclusion":"success","name":"Verify rust pin matches Go constant","number":8,"startedAt":"2026-10-16T07:50:14Z","status":"completed"},{"completedAt":"2026-10-16T07:50:21Z","conclusion":"success","name":"Install pinned Rust toolchain via rustup","number":9,"startedAt":"2026-10-16T07:50:14Z","status":"completed"},{"completedAt":"2026-10-16T07:50:21Z","conclusion":"success","name":"Verify Go toolchain identity","number":10,"startedAt":"2026-10-16T07:50:21Z","status":"completed"},{"completedAt":"2026-10-16T07:50:21Z","conclusion":"success","name":"Verify versioned Go install compatibility","number":11,"startedAt":"2026-10-16T07:50:21Z","status":"completed"},{"completedAt":"2026-10-16T07:50:21Z","conclusion":"success","name":"gofmt check","number":12,"startedAt":"2026-10-16T07:50:21Z","status":"completed"},{"completedAt":"2026-10-16T07:50:37Z","conclusion":"success","name":"go vet","number":13,"startedAt":"2026-10-16T07:50:21Z","status":"completed"},{"completedAt":"2026-10-16T07:50:40Z","conclusion":"success","name":"Ledger consistency","number":14,"startedAt":"2026-10-16T07:50:37Z","status":"completed"},{"completedAt":"2026-10-16T08:10:27Z","conclusion":"success","name":"go test + platform-case gate","number":15,"startedAt":"2026-10-16T07:50:40Z","status":"completed"},{"completedAt":"2026-10-16T08:10:29Z","conclusion":"success","name":"Upload gate evidence","number":16,"startedAt":"2026-10-16T08:10:27Z","status":"completed"},{"completedAt":"2026-10-16T08:10:29Z","conclusion":"success","name":"Post Run actions/setup-node@v4","number":29,"startedAt":"2026-10-16T08:10:29Z","status":"completed"},{"completedAt":"2026-10-16T08:10:29Z","conclusion":"success","name":"Post Run actions/setup-go@v5","number":30,"startedAt":"2026-10-16T08:10:29Z","status":"completed"},{"completedAt":"2026-10-16T08:10:29Z","conclusion":"success","name":"Post Checkout authoritative protocol suite","number":31,"startedAt":"2026-10-16T08:10:29Z","status":"completed"},{"completedAt":"2026-10-16T08:10:29Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":32,"startedAt":"2026-10-16T08:10:29Z","status":"completed"},{"completedAt":"2026-10-16T08:10:29Z","conclusion":"success","name":"Complete job","number":33,"startedAt":"2026-10-16T08:10:29Z","status":"completed"}],"url":"https://github.com/relux-works/curator/actions/runs/38869688515/job/116886090523"},{"completedAt":"2026-10-16T08:24:41Z","conclusion":"success","databaseId":116886090682,"name":"Race (macos-latest)","startedAt":"2026-10-16T07:49:58Z","status":"completed","steps":[{"completedAt":"2026-10-16T07:50:00Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-16T07:49:58Z","status":"completed"},{"completedAt":"2026-10-16T07:50:14Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-16T07:50:00Z","status":"completed"},{"completedAt":"2026-10-16T07:50:16Z","conclusion":"success","name":"Checkout authoritative protocol suite","number":3,"startedAt":"2026-10-16T07:50:14Z","status":"completed"},{"completedAt":"2026-10-16T07:50:52Z","conclusion":"success","name":"Run actions/setup-go@v5","number":4,"startedAt":"2026-10-16T07:50:16Z","status":"completed"},{"completedAt":"2026-10-16T07:50:53Z","conclusion":"success","name":"Run actions/setup-node@v4","number":5,"startedAt":"2026-10-16T07:50:52Z","status":"completed"},{"completedAt":"2026-10-16T07:50:53Z","conclusion":"success","name":"Verify pnpm pin matches Go constant","number":6,"startedAt":"2026-10-16T07:50:53Z","status":"completed"},{"completedAt":"2026-10-16T07:50:55Z","conclusion":"success","name":"Install pinned pnpm via npm","number":7,"startedAt":"2026-10-16T07:50:53Z","status":"completed"},{"completedAt":"2026-10-16T07:50:55Z","conclusion":"success","name":"Verify rust pin matches Go constant","number":8,"startedAt":"2026-10-16T07:50:55Z","status":"completed"},{"completedAt":"2026-10-16T07:51:02Z","conclusion":"success","name":"Install pinned Rust toolchain via rustup","number":9,"startedAt":"2026-10-16T07:50:55Z","status":"completed"},{"completedAt":"2026-10-16T07:51:03Z","conclusion":"success","name":"Verify Go toolchain identity","number":10,"startedAt":"2026-10-16T07:51:02Z","status":"completed"},{"completedAt":"2026-10-16T08:24:27Z","conclusion":"success","name":"go test -race + platform-case gate","number":11,"startedAt":"2026-10-16T07:51:03Z","status":"completed"},{"completedAt":"2026-10-16T08:24:29Z","conclusion":"success","name":"Upload race evidence","number":12,"startedAt":"2026-10-16T08:24:27Z","status":"completed"},{"completedAt":"2026-10-16T08:24:29Z","conclusion":"success","name":"Post Run actions/setup-node@v4","number":21,"startedAt":"2026-10-16T08:24:29Z","status":"completed"},{"completedAt":"2026-10-16T08:24:30Z","conclusion":"success","name":"Post Run actions/setup-go@v5","number":22,"startedAt":"2026-10-16T08:24:29Z","status":"completed"},{"completedAt":"2026-10-16T08:24:32Z","conclusion":"success","name":"Post Checkout authoritative protocol suite","number":23,"startedAt":"2026-10-16T08:24:30Z","status":"completed"},{"completedAt":"2026-10-16T08:24:33Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":24,"startedAt":"2026-10-16T08:24:32Z","status":"completed"},{"completedAt":"2026-10-16T08:24:38Z","conclusion":"success","name":"Complete job","number":25,"startedAt":"2026-10-16T08:24:33Z","status":"completed"}],"url":"https://github.com/relux-works/curator/actions/runs/38869688515/job/116886090682"},{"completedAt":"2026-10-16T07:54:06Z","conclusion":"success","databaseId":116886091052,"name":"Go 1.27.1 driver (windows-latest)","startedAt":"2026-10-16T07:50:32Z","status":"completed","steps":[{"completedAt":"2026-10-16T07:50:34Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-16T07:50:33Z","status":"completed"},{"completedAt":"2026-10-16T07:50:45Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-16T07:50:34Z","status":"completed"},{"completedAt":"2026-10-16T07:50:48Z","conclusion":"success","name":"Checkout authoritative protocol suite","number":3,"startedAt":"2026-10-16T07:50:45Z","status":"completed"},{"completedAt":"2026-10-16T07:51:18Z","conclusion":"success","name":"Run actions/setup-go@v5","number":4,"startedAt":"2026-10-16T07:50:48Z","status":"completed"},{"completedAt":"2026-10-16T07:51:19Z","conclusion":"success","name":"Verify family toolchain identity","number":5,"startedAt":"2026-10-16T07:51:18Z","status":"completed"},{"completedAt":"2026-10-16T07:53:38Z","conclusion":"success","name":"Driver qualification and reproducibility","number":6,"startedAt":"2026-10-16T07:51:19Z","status":"completed"},{"completedAt":"2026-10-16T07:53:41Z","conclusion":"success","name":"Driver vet","number":7,"startedAt":"2026-10-16T07:53:38Z","status":"completed"},{"completedAt":"2026-10-16T07:53:59Z","conclusion":"success","name":"Build manager","number":8,"startedAt":"2026-10-16T07:53:41Z","status":"completed"},{"completedAt":"2026-10-16T07:53:59Z","conclusion":"success","name":"Post Run actions/setup-go@v5","number":14,"startedAt":"2026-10-16T07:53:59Z","status":"completed"},{"completedAt":"2026-10-16T07:54:01Z","conclusion":"success","name":"Post Checkout authoritative protocol suite","number":15,"startedAt":"2026-10-16T07:53:59Z","status":"completed"},{"completedAt":"2026-10-16T07:54:04Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":16,"startedAt":"2026-10-16T07:54:01Z","status":"completed"},{"completedAt":"2026-10-16T07:54:04Z","conclusion":"success","name":"Complete job","number":17,"startedAt":"2026-10-16T07:54:04Z","status":"completed"}],"url":"https://github.com/relux-works/curator/actions/runs/38869688515/job/116886091052"},{"completedAt":"2026-10-16T07:53:11Z","conclusion":"success","databaseId":116886091165,"name":"Go 1.26.0 driver (windows-latest)","startedAt":"2026-10-16T07:49:52Z","status":"completed","steps":[{"completedAt":"2026-10-16T07:49:54Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-16T07:49:53Z","status":"completed"},{"completedAt":"2026-10-16T07:50:05Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-16T07:49:54Z","status":"completed"},{"completedAt":"2026-10-16T07:50:09Z","conclusion":"success","name":"Checkout authoritative protocol suite","number":3,"startedAt":"2026-10-16T07:50:05Z","status":"completed"},{"completedAt":"2026-10-16T07:50:37Z","conclusion":"success","name":"Run actions/setup-go@v5","number":4,"startedAt":"2026-10-16T07:50:09Z","status":"completed"},{"completedAt":"2026-10-16T07:50:37Z","conclusion":"success","name":"Verify family toolchain identity","number":5,"startedAt":"2026-10-16T07:50:37Z","status":"completed"},{"completedAt":"2026-10-16T07:52:43Z","conclusion":"success","name":"Driver qualification and reproducibility","number":6,"startedAt":"2026-10-16T07:50:37Z","status":"completed"},{"completedAt":"2026-10-16T07:52:47Z","conclusion":"success","name":"Driver vet","number":7,"startedAt":"2026-10-16T07:52:43Z","status":"completed"},{"completedAt":"2026-10-16T07:53:04Z","conclusion":"success","name":"Build manager","number":8,"startedAt":"2026-10-16T07:52:47Z","status":"completed"},{"completedAt":"2026-10-16T07:53:04Z","conclusion":"success","name":"Post Run actions/setup-go@v5","number":14,"startedAt":"2026-10-16T07:53:04Z","status":"completed"},{"completedAt":"2026-10-16T07:53:07Z","conclusion":"success","name":"Post Checkout authoritative protocol suite","number":15,"startedAt":"2026-10-16T07:53:04Z","status":"completed"},{"completedAt":"2026-10-16T07:53:09Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":16,"startedAt":"2026-10-16T07:53:07Z","status":"completed"},{"completedAt":"2026-10-16T07:53:09Z","conclusion":"success","name":"Complete job","number":17,"startedAt":"2026-10-16T07:53:09Z","status":"completed"}],"url":"https://github.com/relux-works/curator/actions/runs/38869688515/job/116886091165"},{"completedAt":"2026-10-16T08:20:48Z","conclusion":"success","databaseId":116886123572,"name":"Race (ubuntu-latest)","startedAt":"2026-10-16T07:50:09Z","status":"completed","steps":[{"completedAt":"2026-10-16T07:50:10Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-16T07:50:09Z","status":"completed"},{"completedAt":"2026-10-16T07:50:18Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-16T07:50:10Z","status":"completed"},{"completedAt":"2026-10-16T07:50:19Z","conclusion":"success","name":"Checkout authoritative protocol suite","number":3,"startedAt":"2026-10-16T07:50:18Z","status":"completed"},{"completedAt":"2026-10-16T07:50:28Z","conclusion":"success","name":"Run actions/setup-go@v5","number":4,"startedAt":"2026-10-16T07:50:19Z","status":"completed"},{"completedAt":"2026-10-16T07:50:29Z","conclusion":"success","name":"Run actions/setup-node@v4","number":5,"startedAt":"2026-10-16T07:50:28Z","status":"completed"},{"completedAt":"2026-10-16T07:50:29Z","conclusion":"success","name":"Verify pnpm pin matches Go constant","number":6,"startedAt":"2026-10-16T07:50:29Z","status":"completed"},{"completedAt":"2026-10-16T07:50:30Z","conclusion":"success","name":"Install pinned pnpm via npm","number":7,"startedAt":"2026-10-16T07:50:29Z","status":"completed"},{"completedAt":"2026-10-16T07:50:31Z","conclusion":"success","name":"Verify rust pin matches Go constant","number":8,"startedAt":"2026-10-16T07:50:30Z","status":"completed"},{"completedAt":"2026-10-16T07:50:38Z","conclusion":"success","name":"Install pinned Rust toolchain via rustup","number":9,"startedAt":"2026-10-16T07:50:31Z","status":"completed"},{"completedAt":"2026-10-16T07:50:38Z","conclusion":"success","name":"Verify Go toolchain identity","number":10,"startedAt":"2026-10-16T07:50:38Z","status":"completed"},{"completedAt":"2026-10-16T08:20:44Z","conclusion":"success","name":"go test -race + platform-case gate","number":11,"startedAt":"2026-10-16T07:50:38Z","status":"completed"},{"completedAt":"2026-10-16T08:20:45Z","conclusion":"success","name":"Upload race evidence","number":12,"startedAt":"2026-10-16T08:20:44Z","status":"completed"},{"completedAt":"2026-10-16T08:20:45Z","conclusion":"success","name":"Post Run actions/setup-node@v4","number":21,"startedAt":"2026-10-16T08:20:45Z","status":"completed"},{"completedAt":"2026-10-16T08:20:46Z","conclusion":"success","name":"Post Run actions/setup-go@v5","number":22,"startedAt":"2026-10-16T08:20:45Z","status":"completed"},{"completedAt":"2026-10-16T08:20:46Z","conclusion":"success","name":"Post Checkout authoritative protocol suite","number":23,"startedAt":"2026-10-16T08:20:46Z","status":"completed"},{"completedAt":"2026-10-16T08:20:46Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":24,"startedAt":"2026-10-16T08:20:46Z","status":"completed"},{"completedAt":"2026-10-16T08:20:46Z","conclusion":"success","name":"Complete job","number":25,"startedAt":"2026-10-16T08:20:46Z","status":"completed"}],"url":"https://github.com/relux-works/curator/actions/runs/38869688515/job/116886123572"},{"completedAt":"2026-10-16T07:50:20Z","conclusion":"success","databaseId":116886132197,"name":"Gate self-test (ubuntu-latest)","startedAt":"2026-10-16T07:49:52Z","status":"completed","steps":[{"completedAt":"2026-10-16T07:49:54Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-16T07:49:53Z","status":"completed"},{"completedAt":"2026-10-16T07:50:01Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-16T07:49:54Z","status":"completed"},{"completedAt":"2026-10-16T07:50:18Z","conclusion":"success","name":"Self-test the CI gates","number":3,"startedAt":"2026-10-16T07:50:01Z","status":"completed"},{"completedAt":"2026-10-16T07:50:18Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":6,"startedAt":"2026-10-16T07:50:18Z","status":"completed"},{"completedAt":"2026-10-16T07:50:18Z","conclusion":"success","name":"Complete job","number":7,"startedAt":"2026-10-16T07:50:18Z","status":"completed"}],"url":"https://github.com/relux-works/curator/actions/runs/38869688515/job/116886132197"},{"completedAt":"2026-10-16T07:51:07Z","conclusion":"success","databaseId":116886133989,"name":"Interop conformance gate","startedAt":"2026-10-16T07:50:47Z","status":"completed","steps":[{"completedAt":"2026-10-16T07:50:48Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-16T07:50:48Z","status":"completed"},{"completedAt":"2026-10-16T07:50:53Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-16T07:50:48Z","status":"completed"},{"completedAt":"2026-10-16T07:50:53Z","conclusion":"success","name":"Checkout authoritative protocol suite","number":3,"startedAt":"2026-10-16T07:50:53Z","status":"completed"},{"completedAt":"2026-10-16T07:51:02Z","conclusion":"success","name":"Run actions/setup-go@v5","number":4,"startedAt":"2026-10-16T07:50:53Z","status":"completed"},{"completedAt":"2026-10-16T07:51:02Z","conclusion":"success","name":"Verify Go toolchain identity","number":5,"startedAt":"2026-10-16T07:51:02Z","status":"completed"},{"completedAt":"2026-10-16T07:51:04Z","conclusion":"success","name":"Shared protocol suite","number":6,"startedAt":"2026-10-16T07:51:02Z","status":"completed"},{"completedAt":"2026-10-16T07:51:05Z","conclusion":"success","name":"Post Run actions/setup-go@v5","number":10,"startedAt":"2026-10-16T07:51:04Z","status":"completed"},{"completedAt":"2026-10-16T07:51:05Z","conclusion":"success","name":"Post Checkout authoritative protocol suite","number":11,"startedAt":"2026-10-16T07:51:05Z","status":"completed"},{"completedAt":"2026-10-16T07:51:05Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":12,"startedAt":"2026-10-16T07:51:05Z","status":"completed"},{"completedAt":"2026-10-16T07:51:05Z","conclusion":"success","name":"Complete job","number":13,"startedAt":"2026-10-16T07:51:05Z","status":"completed"}],"url":"https://github.com/relux-works/curator/actions/runs/38869688515/job/116886133989"},{"completedAt":"2026-10-16T07:50:46Z","conclusion":"success","databaseId":116886134849,"name":"Go 1.26.0 driver (ubuntu-latest)","startedAt":"2026-10-16T07:50:03Z","status":"completed","steps":[{"completedAt":"2026-10-16T07:50:05Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-16T07:50:04Z","status":"completed"},{"completedAt":"2026-10-16T07:50:08Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-16T07:50:05Z","status":"completed"},{"completedAt":"2026-10-16T07:50:09Z","conclusion":"success","name":"Checkout authoritative protocol suite","number":3,"startedAt":"2026-10-16T07:50:08Z","status":"completed"},{"completedAt":"2026-10-16T07:50:21Z","conclusion":"success","name":"Run actions/setup-go@v5","number":4,"startedAt":"2026-10-16T07:50:09Z","status":"completed"},{"completedAt":"2026-10-16T07:50:21Z","conclusion":"success","name":"Verify family toolchain identity","number":5,"startedAt":"2026-10-16T07:50:21Z","status":"completed"},{"completedAt":"2026-10-16T07:50:26Z","conclusion":"success","name":"Driver qualification and reproducibility","number":6,"startedAt":"2026-10-16T07:50:21Z","status":"completed"},{"completedAt":"2026-10-16T07:50:28Z","conclusion":"success","name":"Driver vet","number":7,"startedAt":"2026-10-16T07:50:26Z","status":"completed"},{"completedAt":"2026-10-16T07:50:42Z","conclusion":"success","name":"Build manager","number":8,"startedAt":"2026-10-16T07:50:28Z","status":"completed"},{"completedAt":"2026-10-16T07:50:42Z","conclusion":"success","name":"Post Run actions/setup-go@v5","number":14,"startedAt":"2026-10-16T07:50:42Z","status":"completed"},{"completedAt":"2026-10-16T07:50:42Z","conclusion":"success","name":"Post Checkout authoritative protocol suite","number":15,"startedAt":"2026-10-16T07:50:42Z","status":"completed"},{"completedAt":"2026-10-16T07:50:43Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":16,"startedAt":"2026-10-16T07:50:42Z","status":"completed"},{"completedAt":"2026-10-16T07:50:43Z","conclusion":"success","name":"Complete job","number":17,"startedAt":"2026-10-16T07:50:43Z","status":"completed"}],"url":"https://github.com/relux-works/curator/actions/runs/38869688515/job/116886134849"},{"completedAt":"2026-10-16T07:50:48Z","conclusion":"success","databaseId":116886136467,"name":"Gate self-test (macos-latest)","startedAt":"2026-10-16T07:50:23Z","status":"completed","steps":[{"completedAt":"2026-10-16T07:50:24Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-16T07:50:24Z","status":"completed"},{"completedAt":"2026-10-16T07:50:35Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-16T07:50:24Z","status":"completed"},{"completedAt":"2026-10-16T07:50:45Z","conclusion":"success","name":"Self-test the CI gates","number":3,"startedAt":"2026-10-16T07:50:35Z","status":"completed"},{"completedAt":"2026-10-16T07:50:45Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":6,"startedAt":"2026-10-16T07:50:45Z","status":"completed"},{"completedAt":"2026-10-16T07:50:46Z","conclusion":"success","name":"Complete job","number":7,"startedAt":"2026-10-16T07:50:45Z","status":"completed"}],"url":"https://github.com/relux-works/curator/actions/runs/38869688515/job/116886136467"},{"completedAt":"2026-10-16T07:50:07Z","conclusion":"success","databaseId":116886139040,"name":"Naming gate","startedAt":"2026-10-16T07:49:53Z","status":"completed","steps":[{"completedAt":"2026-10-16T07:49:54Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-16T07:49:53Z","status":"completed"},{"completedAt":"2026-10-16T07:50:03Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-16T07:49:54Z","status":"completed"},{"completedAt":"2026-10-16T07:50:04Z","conclusion":"success","name":"Employer name gate","number":3,"startedAt":"2026-10-16T07:50:03Z","status":"completed"},{"completedAt":"2026-10-16T07:50:05Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":6,"startedAt":"2026-10-16T07:50:04Z","status":"completed"},{"completedAt":"2026-10-16T07:50:05Z","conclusion":"success","name":"Complete job","number":7,"startedAt":"2026-10-16T07:50:05Z","status":"completed"}],"url":"https://github.com/relux-works/curator/actions/runs/38869688515/job/116886139040"},{"completedAt":"2026-10-16T07:52:27Z","conclusion":"success","databaseId":116886141826,"name":"Go 1.26.0 driver (macos-latest)","startedAt":"2026-10-16T07:51:00Z","status":"completed","steps":[{"completedAt":"2026-10-16T07:51:01Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-16T07:51:00Z","status":"completed"},{"completedAt":"2026-10-16T07:51:09Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-16T07:51:01Z","status":"completed"},{"completedAt":"2026-10-16T07:51:10Z","conclusion":"success","name":"Checkout authoritative protocol suite","number":3,"startedAt":"2026-10-16T07:51:09Z","status":"completed"},{"completedAt":"2026-10-16T07:51:21Z","conclusion":"success","name":"Run actions/setup-go@v5","number":4,"startedAt":"2026-10-16T07:51:10Z","status":"completed"},{"completedAt":"2026-10-16T07:51:21Z","conclusion":"success","name":"Verify family toolchain identity","number":5,"startedAt":"2026-10-16T07:51:21Z","status":"completed"},{"completedAt":"2026-10-16T07:52:15Z","conclusion":"success","name":"Driver qualification and reproducibility","number":6,"startedAt":"2026-10-16T07:51:21Z","status":"completed"},{"completedAt":"2026-10-16T07:52:16Z","conclusion":"success","name":"Driver vet","number":7,"startedAt":"2026-10-16T07:52:15Z","status":"completed"},{"completedAt":"2026-10-16T07:52:21Z","conclusion":"success","name":"Build manager","number":8,"startedAt":"2026-10-16T07:52:16Z","status":"completed"},{"completedAt":"2026-10-16T07:52:22Z","conclusion":"success","name":"Post Run actions/setup-go@v5","number":14,"startedAt":"2026-10-16T07:52:21Z","status":"completed"},{"completedAt":"2026-10-16T07:52:22Z","conclusion":"success","name":"Post Checkout authoritative protocol suite","number":15,"startedAt":"2026-10-16T07:52:22Z","status":"completed"},{"completedAt":"2026-10-16T07:52:22Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":16,"startedAt":"2026-10-16T07:52:22Z","status":"completed"},{"completedAt":"2026-10-16T07:52:23Z","conclusion":"success","name":"Complete job","number":17,"startedAt":"2026-10-16T07:52:22Z","status":"completed"}],"url":"https://github.com/relux-works/curator/actions/runs/38869688515/job/116886141826"},{"completedAt":"2026-10-16T08:15:27Z","conclusion":"success","databaseId":116886145726,"name":"Test (macos-latest)","startedAt":"2026-10-16T07:52:33Z","status":"completed","steps":[{"completedAt":"2026-10-16T07:52:35Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-16T07:52:34Z","status":"completed"},{"completedAt":"2026-10-16T07:52:47Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-16T07:52:35Z","status":"completed"},{"completedAt":"2026-10-16T07:52:49Z","conclusion":"success","name":"Checkout authoritative protocol suite","number":3,"startedAt":"2026-10-16T07:52:47Z","status":"completed"},{"completedAt":"2026-10-16T07:53:14Z","conclusion":"success","name":"Run actions/setup-go@v5","number":4,"startedAt":"2026-10-16T07:52:49Z","status":"completed"},{"completedAt":"2026-10-16T07:53:14Z","conclusion":"success","name":"Run actions/setup-node@v4","number":5,"startedAt":"2026-10-16T07:53:14Z","status":"completed"},{"completedAt":"2026-10-16T07:53:14Z","conclusion":"success","name":"Verify pnpm pin matches Go constant","number":6,"startedAt":"2026-10-16T07:53:14Z","status":"completed"},{"completedAt":"2026-10-16T07:53:16Z","conclusion":"success","name":"Install pinned pnpm via npm","number":7,"startedAt":"2026-10-16T07:53:14Z","status":"completed"},{"completedAt":"2026-10-16T07:53:16Z","conclusion":"success","name":"Verify rust pin matches Go constant","number":8,"startedAt":"2026-10-16T07:53:16Z","status":"completed"},{"completedAt":"2026-10-16T07:53:23Z","conclusion":"success","name":"Install pinned Rust toolchain via rustup","number":9,"startedAt":"2026-10-16T07:53:16Z","status":"completed"},{"completedAt":"2026-10-16T07:53:23Z","conclusion":"success","name":"Verify Go toolchain identity","number":10,"startedAt":"2026-10-16T07:53:23Z","status":"completed"},{"completedAt":"2026-10-16T07:53:24Z","conclusion":"success","name":"Verify versioned Go install compatibility","number":11,"startedAt":"2026-10-16T07:53:23Z","status":"completed"},{"completedAt":"2026-10-16T07:53:24Z","conclusion":"success","name":"gofmt check","number":12,"startedAt":"2026-10-16T07:53:24Z","status":"completed"},{"completedAt":"2026-10-16T07:53:36Z","conclusion":"success","name":"go vet","number":13,"startedAt":"2026-10-16T07:53:24Z","status":"completed"},{"completedAt":"2026-10-16T07:53:48Z","conclusion":"success","name":"Ledger consistency","number":14,"startedAt":"2026-10-16T07:53:36Z","status":"completed"},{"completedAt":"2026-10-16T08:15:14Z","conclusion":"success","name":"go test + platform-case gate","number":15,"startedAt":"2026-10-16T07:53:48Z","status":"completed"},{"completedAt":"2026-10-16T08:15:15Z","conclusion":"success","name":"Upload gate evidence","number":16,"startedAt":"2026-10-16T08:15:14Z","status":"completed"},{"completedAt":"2026-10-16T08:15:15Z","conclusion":"success","name":"Post Run actions/setup-node@v4","number":29,"startedAt":"2026-10-16T08:15:15Z","status":"completed"},{"completedAt":"2026-10-16T08:15:16Z","conclusion":"success","name":"Post Run actions/setup-go@v5","number":30,"startedAt":"2026-10-16T08:15:15Z","status":"completed"},{"completedAt":"2026-10-16T08:15:17Z","conclusion":"success","name":"Post Checkout authoritative protocol suite","number":31,"startedAt":"2026-10-16T08:15:16Z","status":"completed"},{"completedAt":"2026-10-16T08:15:18Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":32,"startedAt":"2026-10-16T08:15:17Z","status":"completed"},{"completedAt":"2026-10-16T08:15:23Z","conclusion":"success","name":"Complete job","number":33,"startedAt":"2026-10-16T08:15:18Z","status":"completed"}],"url":"https://github.com/relux-works/curator/actions/runs/38869688515/job/116886145726"},{"completedAt":"2026-10-16T07:50:55Z","conclusion":"success","databaseId":116886149092,"name":"Lint","startedAt":"2026-10-16T07:49:52Z","status":"completed","steps":[{"completedAt":"2026-10-16T07:49:55Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-16T07:49:53Z","status":"completed"},{"completedAt":"2026-10-16T07:50:05Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-16T07:49:55Z","status":"completed"},{"completedAt":"2026-10-16T07:50:13Z","conclusion":"success","name":"Run actions/setup-go@v5","number":3,"startedAt":"2026-10-16T07:50:05Z","status":"completed"},{"completedAt":"2026-10-16T07:50:13Z","conclusion":"success","name":"Verify Go toolchain identity","number":4,"startedAt":"2026-10-16T07:50:13Z","status":"completed"},{"completedAt":"2026-10-16T07:50:50Z","conclusion":"success","name":"Run golangci/golangci-lint-action@v7","number":5,"startedAt":"2026-10-16T07:50:13Z","status":"completed"},{"completedAt":"2026-10-16T07:50:51Z","conclusion":"success","name":"No broad suppression","number":6,"startedAt":"2026-10-16T07:50:50Z","status":"completed"},{"completedAt":"2026-10-16T07:50:51Z","conclusion":"success","name":"Verify GoReleaser rc channel values","number":7,"startedAt":"2026-10-16T07:50:51Z","status":"completed"},{"completedAt":"2026-10-16T07:50:51Z","conclusion":"success","name":"Post Run golangci/golangci-lint-action@v7","number":12,"startedAt":"2026-10-16T07:50:51Z","status":"completed"},{"completedAt":"2026-10-16T07:50:51Z","conclusion":"success","name":"Post Run actions/setup-go@v5","number":13,"startedAt":"2026-10-16T07:50:51Z","status":"completed"},{"completedAt":"2026-10-16T07:50:51Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":14,"startedAt":"2026-10-16T07:50:51Z","status":"completed"},{"completedAt":"2026-10-16T07:50:51Z","conclusion":"success","name":"Complete job","number":15,"startedAt":"2026-10-16T07:50:51Z","status":"completed"}],"url":"https://github.com/relux-works/curator/actions/runs/38869688515/job/116886149092"}],"status":"completed","url":"https://github.com/relux-works/curator/actions/runs/38869688515"}

```
