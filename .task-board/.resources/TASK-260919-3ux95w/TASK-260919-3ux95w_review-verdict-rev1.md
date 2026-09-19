# Review verdict — TASK-260919-3ux95w rev1 (CR-TASK-260919-3ux95w-1)

**Verdict: ACCEPT** — recorded via `accept_cr(TASK-260919-3ux95w, revision=1, evidence=TASK-260919-3ux95w_review-verdict-rev1.md)`.

Reviewer: [reviewer] reviewer (claude-opus-5, RUN-260919-f2b5a2). Read-only: no bytes in the Story worktree were modified; every mutant ran in a disposable clone under `$TMPDIR`.
Shell for every command below: bash invoked through the harness Bash tool (zsh login), `set -o pipefail` where a pipe is involved; exit codes are the real ones.

## 1. Exact candidate tree (verified, not taken from the producer)

| Fact | Value | How |
|---|---|---|
| Base OID | `6f173778c9c1dcd6312d3cb91320df3df458ad7e` | matches `HEAD` of the Story worktree |
| Candidate tree OID | `355ae2ca84960573f0f1c13e0a8ebf5018294e12` | `git write-tree` of a temp index built from the worktree (`read-tree HEAD` + `add -A`) reproduces it byte-for-byte; `git diff 355ae2ca… -- .` is empty; no untracked files |
| Delta shape | `1 file changed, 9 insertions(+), 6 deletions(-)` — `.github/workflows/ci.yml` only | `git diff --stat <base> <tree>` |
| Delta content | (a) Test job `GO_TEST_TIMEOUT: ${{ runner.os == 'Windows' && '60m' \|\| '30m' }}` → `'120m'`; (b) Candidate suite same expression → `'120m'`; (c) Test-job budget comment rewritten with 3516 s / run 35424565415 / 97.7% / rev2 run 35418260010 timed out / ~79 min extrapolation, keeps "A hang is still bounded and still fatal."; (d) Candidate-suite comment "the same 60m as the default lane" → "the same 120m" | `git diff <base> <tree>` read in full |
| Anything else | nothing — no other path, no other hunk | same diff |
| Rose-air lane (line 337) | untouched at fixed `30m` (macOS ARM64 self-hosted, no Windows branch; AC says non-Windows stays 30m) | grep + ruby YAML parse |
| Stale `'60m'` anywhere under `.github/` | none | `grep -rnE "GO_TEST_TIMEOUT|'60m'" .github/` (only test-gate.sh default `30m` and the two 120m/30m expressions + rose-air 30m remain) |

## 2. Hosted gate on the exact candidate tree (verified against GitHub, not the validation log alone)

- CR validation log: `sh scripts/remote-gate.sh` pushed `71c9093627b8b1800feb418fd23923d3db3cff23` as `gate/STORY-260915-3w11un/260919-072454-92633-1` → run **35429257975** → `success`, `[exit 0]`, `coverage_unit=exact_command_shard required=1 green=1 failed=0 missing=0`.
- `git rev-parse 71c9093…^{tree}` = `355ae2ca84960573f0f1c13e0a8ebf5018294e12` (the candidate tree); parent = `6f173778…` (the base). The gate ran on exactly the reviewed tree.
- `gh run view 35429257975 --json headSha,...`: headSha `71c9093…`, conclusion `success`; jobs: Test (ubuntu/macos/windows) success, Race (ubuntu/macos) success, Gate self-test (ubuntu/macos/windows) success, Lint / Naming / Interop success; Test (rose-air) skipped; Candidate suite skipped.
- **Windows Test job (id 105860762390) log**, step `go test + platform-case gate`, env block:
  `GO_TEST_TIMEOUT: 120m` (log line 758) and `test-gate: flags=<none> timeout=120m` (line 760). The job ran 07:25:00Z → 08:06:48Z (41m48s wall); the served stage 07:28:08Z → 08:06:39Z; `test-gate: go test overall exit=0`; platform-case gate rows all `ok`.
- **Non-Windows branch of the expression** on the same run: ubuntu Test job log `GO_TEST_TIMEOUT: 30m` / `timeout=30m` (lines 757/759); macOS Test job log `GO_TEST_TIMEOUT: 30m` / `timeout=30m` (lines 750/752). Both branches of the expression are therefore exercised on hosted runners.
- Windows `test-evidence-windows-latest` artifact (id 10580792509), `test/go-test.json`, package-level results: 75 packages, 0 fail. Slowest: `internal/install` **2201.0 s**, `internal/install/atomicity` 1542.9 s, `cmd/curator` 1271.7 s, `internal/envprofile` 1011.6 s. On this (faster) runner internal/install used 61% of the old 60m and 31% of the new 120m; the motivating measurement (3516 s, run 35424565415) is 49% of 120m and the ~79 min extrapolation is 66% — the variance between 2201 s and 3516 s on the same package is exactly what the new headroom absorbs.

## 3. "A hang is still bounded and still fatal" — verified against the consumer

- `.github/ci/test-gate.sh:38` `GO_TEST_TIMEOUT="${GO_TEST_TIMEOUT:-30m}"`, passed as `go test -json -count=1 -timeout "$GO_TEST_TIMEOUT" …` at lines 104/115/125 — a per-package `go test` deadline; a hung package panics its test binary → non-zero → `test-gate` fails the lane.
- Job-level `timeout-minutes` in ci.yml: only `test-self-hosted` (rose-air, line 247, 90m, with a fixed 30m budget). `test` and `candidate-conformance` have none → GitHub's 360m default applies, so `go test`'s 120m remains the effective bound and is not pre-empted by a job kill.

## 4. Gate self-test — independent rerun and gate attack

Producer claim: the self-test pins the budget relationally, no literal row. Verified by reading `.github/ci/gate-selftest.sh:680-740`: row 159 "every windows test-gate.sh lane declares GO_TEST_TIMEOUT" (awk over lanes), row 160 "windows per-package budget is larger than the unix one" (`minutes(win) > minutes(other)`), row 161 "unix per-package budget still matches the gate default" (`minutes(other) == minutes(test-gate.sh default)`). `win_timeout`/`other_timeout` are read from the **first** `GO_TEST_TIMEOUT:.*Windows` match (`exit` after the first), i.e. the Test job line. No self-test edit was needed for 120m/30m; none was made.

Reruns in the Story worktree (candidate tree):
- `bash -n .github/ci/gate-selftest.sh` → exit 0; `bash -n .github/ci/test-gate.sh` → exit 0
- `bash .github/ci/gate-selftest.sh` → **187 passed, 0 failed, exit 0**; rows 159/160/161 all `ok`
- `ruby -ryaml YAML.load_file(ci.yml)` → YAML-OK; parsed `jobs.test` step env `GO_TEST_TIMEOUT` = `${{ runner.os == 'Windows' && '120m' || '30m' }}`; `jobs.candidate-conformance` step env = same; `jobs.test-self-hosted` = `30m`. (yamllint and pyyaml are not installed on this host — same bound the producer reported.)

Gate attack (disposable clone at base `6f17377…` with the candidate ci.yml checked out from tree `355ae2ca…`, blob `380468f…` confirmed; each mutant restored with `git checkout -- ci.yml` afterwards; host load avg ≈16 so each run took ~5 min):

| Mutant | Change | Self-test result |
|---|---|---|
| M0 control | none | 187 passed, 0 failed, exit 0 |
| M1 | Test-job Windows budget narrowed `120m` → `20m` (below unix 30m) | **FAIL** row 160 "windows per-package budget is larger than the unix one" (`windows='20m' other='30m'`) — 186/1, exit 1 |
| M2 | Test-job unix branch drifted `30m` → `45m` (Windows 120m) | **FAIL** row 161 "unix per-package budget still matches the gate default" — 186/1, exit 1 |
| M3 | Test-job `GO_TEST_TIMEOUT` line deleted (unbudgeted Windows lane) | **FAIL** row 159 "every windows test-gate.sh lane declares GO_TEST_TIMEOUT" (`lanes running on windows-latest with no GO_TEST_TIMEOUT: test`) — 186/1, exit 1 |
| M4 | Candidate-suite Windows budget narrowed `120m` → `20m`, Test job left at 120m | **survives** — 187 passed, 0 failed, exit 0 (rows 159/160/161 all `ok`): the self-test does not see the second expression; recorded as a measured bound in §5 |

Every narrowing of the exercised expression is caught; the three budget rows are load-bearing, not decorative.

## 5. Stated bounds (what this review does not prove)

1. **Candidate-suite lane at runtime.** `candidate-conformance` runs only on `workflow_dispatch` with `candidate_ref`/`candidate_root` (ci.yml:561-563), so it was `skipped` on the push gate and its 120m was **not exercised on a hosted runner**. It is verified statically: the expression string is byte-identical to the Test-job one that did run with 120m, and the YAML parse returns it from the candidate-conformance step env. AC wording ("is 120m in both jobs") is a configuration statement and is satisfied.
2. **Self-test blind spot on the second expression.** The self-test's budget rows read the first Windows expression only; a future drift confined to the candidate-suite line is not visible to rows 160/161 (row 159 still passes as long as *some* `GO_TEST_TIMEOUT` is declared). M4 measures this: **the mutant survives (187/0, exit 0)** — the self-test rows are blind to the candidate-suite line. This is a pre-existing bound of the self-test, outside this task's scope ("gate-selftest rows that pin the expression if any; nothing else"); the candidate-suite line is covered here by direct read (§1) and YAML parse (§4). Worth a follow-up self-test row if the orchestrator wants the second lane pinned.
3. **Headroom is a measurement on two runs, not a bound.** 2201 s (this run) and 3516 s (run 35424565415) on Windows internal/install; the ~79 min figure is an extrapolation from the F-W1 verdict, not a measurement. 120m covers all three with ≥34% margin; it does not guarantee a future slower runner.

## 6. AC / DoD mapping

- Windows GO_TEST_TIMEOUT = 120m in the Test job and the Candidate suite job — **yes** (§1, §4 YAML parse; Test job proven at runtime §2)
- non-Windows stays 30m — **yes** (expression else-branch; ubuntu/macos job logs `30m`; rose-air fixed `30m` untouched)
- comment states 3516 s on run 35424565415, ~79 min extrapolation, hang bounded and fatal — **yes** (diff hunk 1; the 97.7% and rev2 run 35418260010 are also stated)
- gate self-test updated if it checks the expression — **n/a, correctly**: relational rows need no edit; suite green 187/0 locally and on all three hosted self-test lanes
- hosted gate green — **yes**, run 35429257975 on the exact candidate tree
- Scope discipline — only ci.yml changed; no code, no tests to write for a CI env expression (no Go behaviour changed); lint job green on the gate

## 7. Disposition

`accept_cr(TASK-260919-3ux95w, revision=1, evidence=TASK-260919-3ux95w_review-verdict-rev1.md)`; the element routes to `integrating`. No `commit_ack`, no `done` from here. Integration by a tracked run with the producer role/archetype (`developer`/`implementer`), per the CR contract.
