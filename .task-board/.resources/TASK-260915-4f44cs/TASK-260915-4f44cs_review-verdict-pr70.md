# Review: PR #70 (chore/worker-policy-and-rose-air-lane)

Head SHA reviewed: 4d240bac6a30aea2585f566068b55a2b325cbef7
Worktree: /Users/administrator/Developer/ReluxWorks/curator/curator/.temp/bootstrap/worktree (`git rev-parse HEAD` = 4d240bac6a30aea2585f566068b55a2b325cbef7)

## Check 1: diff scope
Command: `git diff origin/main..HEAD --stat`
Result: exactly two files -- `.github/workflows/ci.yml` (+64) and `task-board.config.json` (+36/-20). Diff content matches the commit message (codex/claude ceilings and role defaults replaced with gpt-6-astra:low / claude-fable-5-1:low; muse block untouched; one new `test-self-hosted` job). No unrelated changes.

## Check 2: task-board.config.json
- `python3 -c "json.load(...)"` -> valid JSON.
- Key diff vs `origin/main:task-board.config.json`: removed keys: none. Added keys: `spawn/ceilings/codex/roles/{doc-writer,researcher,solution-architect,tester}` only. All other structure identical.
- `task-board --no-update-check --board-dir .../.task-board q 'project_config(view="spawn-preflight", role="developer", agent="codex")'` -> `admitted_pairs.models = [{"id":"gpt-6-astra","efforts":["low"]}]`, contract spawn-policy-v3, authority explicit_allow_set.
- Same with `role="reviewer", agent="claude"` -> `admitted_pairs.models = [{"id":"claude-fable-5-1","efforts":["low"]}]`.
- `q 'models()'`: gpt-6-astra supportedEfforts [low,medium,high,xhigh,max,ultra]; claude-fable-5-1 supportedEfforts [low,medium,high,xhigh,max]. Both admit `low`. Both lifecycle `current`.
- Note: the task prompt's unquoted query form (`view=spawn-preflight`) is a parse error in this task-board version; string args must be quoted. Not a PR issue.

## Check 3: new CI job `Test (rose-air)`
Step-by-step vs hosted `test` job: checkout (submodules) / spec checkout at SPEC_PIN / setup-go from go.mod / toolchain-identity.sh / gofmt / go vet / ledger-consistency.sh / test-gate.sh with CURATOR_CONFORMANCE_ROOT and GO_TEST_TIMEOUT=30m / upload evidence. Identical except: (a) release-source-gate.sh omitted (intended; workflow gate requires exactly one invocation); (b) the `runner.os != 'Windows'` guard on gofmt and the Windows 60m timeout ternary are dropped, correct for a macOS-only lane. `timeout-minutes: 90` added (hosted job has none) -- reasonable for a self-hosted runner.
- Scripts referenced exist under .github/ci/: toolchain-identity.sh, ledger-consistency.sh, test-gate.sh (plus suite-plan.sh, platform-case-gate.sh called transitively).
- `bash .github/ci/release-workflow-gate.sh` -> "release workflow gate: protected CI and publication paths verified", exit 0.
- `bash .github/ci/gate-selftest.sh` -> "gate-selftest: 140 passed, 0 failed", exit 0.
- Artifact name `test-evidence-rose-air` is unique (others: test-evidence-${{ matrix.os }}, race-evidence-*, candidate-evidence-*).
- test-gate.sh: no use of RUNNER_TEMP, GITHUB_*, HOME or /tmp (grep across test-gate, toolchain-identity, ledger-consistency, platform-case-gate, suite-plan, candidate-suite, excluded-packages returned nothing). All writes go to the evidence dir argument (`.temp/ci-evidence/test`, workspace-relative) plus Go's own build/test cache. Requires only `go`, `awk`, `sed`, `bash`. No hosted-only assumptions.

## Check 4: commit provenance
Command: `git log --show-signature -1`
Result: `Good "git" signature with ECDSA key SHA256:V6JiKG7J29mjsvikcLoSVp0bLa77VTsFy12gnLO81cM` (SSH signing; "No principal matched" is the local allowed_signers lookup, not a signature failure). Author: Ivan Oparin <oparin@me.com>.

## Findings
Blocking: none.
Non-blocking:
1. Codex roles now include `doc-writer` while the claude block uses `technical-writer`; the two provider blocks name the writer role differently. Harmless (role defaults are per-provider lookups) but worth aligning later.
2. The self-hosted lane keeps the checkout's `.temp/ci-evidence/` from a previous run on the same runner unless actions/checkout cleans it (it does by default with `clean: true`); no action needed.
3. The rose-air lane does not set CI_REQUIRE_FULL_ROOT, same as the hosted lane; consistent.

VERDICT: ACCEPT
