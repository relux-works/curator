# CocoaSkills Go E2E readiness audit

**Board item:** `BUG-260802-3ibgu1` (`csk-go-e2e-readiness-audit`)  
**Delivery target:** `TASK-260720-3pemm6` (`cross-platform-go-build-e2e`)  
**Dependency:** `TASK-260720-12r55p` (`shared-v6-vector-consumer`) / CocoaSkills PR 19  
**Audit date:** 2026-08-02  
**Mode:** read-only inspection; no CocoaSkills, curator-spec, Curator, checkout, branch, pin, tag, release, or delivery-task status was changed.

## 1. Executive finding

`TASK-260720-3pemm6` is implementation-ready as soon as `TASK-260720-12r55p` is reviewer-accepted and PR 19 is landed on CocoaSkills `main`, but none of its E2E acceptance criteria is already satisfied in full.

PR 19 is an important prerequisite, not the E2E implementation. At audit time it is open at exact head `6e7742f0d28ad95ddd7d8e92364b84062571ad0b` over `main` `dacccaaf3ed18740a4d501fe8a3bfec64644c03e`; the Ubuntu and macOS jobs plus strict mypy are green, while four Windows jobs are still in progress. Its diff adds rc.6 vector adapters and observed lifecycle bindings but adds no vendored skill fixture, no process-level install/launch E2E file, no Go setup step, and no CI environment that makes the existing native fixture mandatory.

The existing CocoaSkills suite has strong component and protocol evidence:

- `tests/test_builds_go_v1_fixture.py` performs a real native Go compile through the closed worker boundary on macOS/Windows, but creates source inline, requires `CSK_GO_V1_MANAGER_EXECUTABLE`, skips when the manager or Go is absent, and only asserts that the compiled artifact exists and was not executed.
- `tests/test_build_activation.py` proves real shim argument/exit behavior, but against synthetic cached artifact bytes rather than an artifact built by CocoaSkills install.
- installer/global/lifecycle suites exercise cache, mutation, transaction, rollback, recovery, concurrency, repair, status, GC, project/global/hybrid and two-project behavior, but the build seam is largely faked or the shared lifecycle observation is not a process-level vendored-skill run.
- `.github/workflows/ci.yml` already runs Python 3.11–3.14 on Ubuntu, macOS and Windows and pins byte-equivalent candidate-suite bytes at curator-spec `0c81c1f8d5321d822be2a2817b05aea03e656e15`, but it does not install a declared Go 1.25 toolchain or set the manager executable required by the native fixture.

The delivery should therefore add one small checked-in mixed skill and a focused process-level E2E module, reuse existing test helpers and product seams, and make the intended native/portable split explicit in CI. It should not duplicate the 1,000+ candidate-vector assertions introduced by PR 19.

## 2. Immutable provenance and release boundary

| Item | Exact evidence / constraint |
|---|---|
| CocoaSkills clean base at audit | `/Users/iv/Developer/Wildberries/cocoaskills`, clean `main`, `dacccaaf3ed18740a4d501fe8a3bfec64644c03e`, `+0/-0` against `origin/main`. |
| PR 19 candidate head | `6e7742f0d28ad95ddd7d8e92364b84062571ad0b`, branch `task/TASK-260720-12r55p-shared-v6-vectors`, open and mergeable. This is not the future task base until accepted and landed. |
| Existing Go driver/fixture provenance | `53b4eb0a8496a9cf35890a4b609d7266a851615e` (`feat(builds): add fixed Go compile driver`), already an ancestor of CocoaSkills `main`; it introduced `tests/test_builds_go_v1_fixture.py`. |
| Accepted rc.6 candidate source | curator-spec `432eb2ee1fe2d6b271e37269f867c8851c325539`, merged main commit for PR 15. |
| Candidate suite root | `/Users/iv/Developer/ReluxWorks/curator-spec/conformance/v1` or an independently materialized checkout at the same commit. |
| Candidate manifest | `sha256:12e58b82579645ba1ccafba49d3e2dd3216005ddf37ae63c68a9fafd46773071`. Direct `sha256sum conformance/v1/manifest.json` reproduced the digest. |
| Committed CI suite pin | `0c81c1f8d5321d822be2a2817b05aea03e656e15`. `git diff --quiet 0c81c1f8... 432eb2ee... -- conformance release/1.0.0-rc.6.json` exited 0, proving byte equivalence for the complete candidate suite plus rc.6 metadata. Preserve this pin in this task. |
| Release status | curator-spec has no upstream `v1.0.0-rc.6` tag and `gh release view v1.0.0-rc.6` exits 1 with `release not found`. `release/1.0.0-rc.6.json` records `committed_release_pin_advanced=false`, `claims_emitted=[]`, macOS/Windows `pending-downstream-native-evidence`, and Linux excluded. |
| Ownership boundary | This task may add fixture/tests and Go setup/candidate-input CI only. `TASK-260720-25d05o` qualifies a published release; only `TASK-260720-1utsx8` may advance/audit CocoaSkills committed release-suite refs. No tag, GitHub Release, released-suite pin, conformance claim, or generic Linux driver-success claim is authorized here. |

The future base SHA cannot honestly be named before PR 19 lands. The exact rule is: the task base is the clean, fast-forwarded CocoaSkills `origin/main` commit that contains accepted PR head `6e7742f0...`; record that resolved full SHA before creating the task worktree. Do not base directly on the PR branch or reuse its active worktree.

## 3. Acceptance-criterion map

### AC 1 — vendored networkless command, explicit launch, argv/exit, project/global shims

**Existing evidence**

- `tests/test_builds_go_v1_fixture.py:66-84` limits real source-aware success to macOS/Windows and `:93-167` constructs a vendored module inline, establishes a real toolchain, and calls `go_v1.build` through `manager-worker-v1`.
- `tests/test_build_activation.py:470-563` launches Unix and Windows project/global shims, proves argument forwarding, and preserves a non-zero status.
- `README.md:320-405` and `docs/skill-authoring.md:250-410` document the exact schema-6 mixed skill layout and the Go 1.25 implementation allowlist over the protocol's Go 1.23 floor.

**Concrete gap**

Add `tests/fixtures/skill_go_e2e/` containing at minimum `SKILL.md`, `agent-skill.json`, `scripts/<script>` plus `.cmd`, `build/go.mod`, `build/cmd/<command>/main.go`, and checked-in `build/vendor/` if importing a non-standard package. The Go command should serialize received argv, controlled output and chosen exit code without reading network or package-selected environment. Add `tests/test_go_build_e2e.py` that installs this fixture through public CocoaSkills project and global flows, asserts no artifact execution during install, invokes the emitted `.agents/bin/<command>` and global `.cmd`/Unix shim explicitly, and compares exact argv/stdout/stderr/exit status. Include a mixed script/build assertion and hybrid activation/collision path; do not call `go_v1.build` as the primary E2E entrypoint.

### AC 2 — cache/mutation/dry-run/build-two/rollback/recovery/concurrency/two-project preservation

**Existing evidence**

- Protected cache component coverage: `tests/test_build_cache_posix.py`, `tests/test_build_cache_windows.py`, `tests/test_build_gc.py`, `tests/test_build_currentness.py`.
- Transaction component coverage: `tests/test_installer_transactions.py`, `tests/test_global_install_transactions.py`; these currently monkeypatch `go_v1.build` for success/failure paths.
- PR 19 candidate-root lifecycle bindings: `tests/protocol_lifecycle_observations.py` and `tests/test_protocol_conformance.py`; these provide protocol breadth but are not a single real-install fixture trace.
- Project/global/hybrid and multi-project implementations already exist in `src/csk/installer.py`, `src/csk/global_install.py`, and `src/csk/status.py`.

**Concrete gap**

Bind the named acceptance cases to the vendored fixture and observable CocoaSkills state:

1. first install = one real compile; second identical install = protected cache hit;
2. relevant Go/vendor mutation = new key/build; irrelevant prompt/script mutation = unchanged build key/hit;
3. dry run leaves cache, receipt, marker, materialization and shim trees byte-for-byte unchanged and never launches worker/artifact;
4. two build commands stage/publish independently and cannot cross-bind artifacts;
5. injected target-swap/publication failure preserves old shim, marker, receipt and cache identity;
6. interrupted transaction is recovered on the next public command;
7. concurrent publishers converge on one valid identity without partial state;
8. updating one of two projects preserves the other project's active artifact and shared cache reference;
9. `repair`/`status` surface the expected applied/unavailable evidence after damage and repair.

Use injection/fake toolchains only for failure timing that cannot be safely produced with a real compiler. Every positive build/cache/launch path must use the real Go 1.25 toolchain.

### AC 3 — native macOS/Windows matrix and Ubuntu fail-closed

**Existing evidence**

- `.github/workflows/ci.yml:13-21` already has all three OSes across Python 3.11–3.14.
- `src/csk/builds/toolchain.py:35` allowlists exactly Go family `1.25`; `protocol/core.md:236` sets the protocol floor at Go 1.23.
- `conformance/v1/vectors/go-host-execution-policy.json:911-919` declares the exhaustive native-control platforms to be exactly macOS and Windows.
- Existing native fixture success is skipped outside those hosts; it does not assert Ubuntu's public unavailable-control result.

**Concrete gap**

- Add deterministic Go 1.25 setup to the test matrix, then export the installed `csk` console launcher and resolved `go` executable.
- On macOS and Windows, run the focused real E2E slice for every supported Python version and fail if it skips.
- On Ubuntu, run portable non-driver E2E plus one source-aware public-install test that asserts `build_execution_control_unavailable`, no worker launch, no artifact/cache publication and no success/claim language.
- Preserve ordinary full pytest on Ubuntu; do not mark the entire Go E2E module skipped there because that would erase the negative contract.

### AC 4 — exact rc.6 evidence with no release/pin/claim change

**Existing evidence**

- PR 19 authenticates the caller-supplied candidate root and records the accepted revision/digest.
- Current workflow pin `0c81c1f8...` is byte-equivalent to accepted merged provenance `432eb2ee...` for `conformance/` and rc.6 metadata.
- `TASK-260720-3pemm6_release-boundary-plan.md` explicitly assigns pin promotion to `TASK-260720-1utsx8` after release qualification.

**Concrete gap**

The E2E job/report must print and verify the caller-supplied root commit and manifest digest, while a diff guard proves `.github/workflows/ci.yml` did not replace the existing curator-spec checkout ref. Any CI change must be limited to Go setup, E2E selection and candidate-root environment. Do not add or edit release/tag/claim files.

### AC 5 — full pytest, strict mypy, build, twine and diff gates

**Existing evidence**

- CI already defines full pytest, strict mypy, build and Twine jobs.
- PR 19's local evidence reports these gates green at its exact head, but those results predate this future E2E change and cannot satisfy its handoff.

**Concrete gap**

Run every gate again from the new task worktree and exact task head. Record each command and real exit code, including hosted macOS/Windows/Ubuntu results. Use a clean-tree/diff guard scoped against the recorded base and a specific no-pin/no-release guard. No checklist item tied to a gate should be checked until that exact command exits 0.

## 4. Exact worktree and fixture procedure

### 4.1 Common base precondition (macOS/Ubuntu shell)

Run only after `TASK-260720-12r55p` is accepted and PR 19 is merged:

```bash
cd /Users/iv/Developer/Wildberries/cocoaskills
git fetch origin main
test -z "$(git status --porcelain)"
git switch main
git merge --ff-only origin/main
BASE_SHA="$(git rev-parse HEAD)"
test "$BASE_SHA" = "$(git rev-parse origin/main)"
git merge-base --is-ancestor 6e7742f0d28ad95ddd7d8e92364b84062571ad0b "$BASE_SHA"
printf '%s\n' "$BASE_SHA"
git worktree add -b task/TASK-260720-3pemm6-cross-platform-go-e2e \
  .temp/TASK-260720-3pemm6/worktree "$BASE_SHA"
cd .temp/TASK-260720-3pemm6/worktree
git status --porcelain=v2 --branch
```

The printed `BASE_SHA` is the exact base to put in task notes/evidence. If the ancestry check fails, stop: the dependency has not landed on the selected base.

### 4.2 Equivalent Windows PowerShell procedure

```powershell
Set-Location C:\path\to\cocoaskills
git fetch origin main
if ((git status --porcelain)) { throw 'base checkout is dirty' }
git switch main
git merge --ff-only origin/main
$BaseSha = (git rev-parse HEAD).Trim()
if ($BaseSha -ne (git rev-parse origin/main).Trim()) { throw 'main is not origin/main' }
git merge-base --is-ancestor 6e7742f0d28ad95ddd7d8e92364b84062571ad0b $BaseSha
if ($LASTEXITCODE -ne 0) { throw 'accepted PR19 head is not in the base' }
$BaseSha
git worktree add -b task/TASK-260720-3pemm6-cross-platform-go-e2e `
  .temp/TASK-260720-3pemm6/worktree $BaseSha
Set-Location .temp/TASK-260720-3pemm6/worktree
git status --porcelain=v2 --branch
```

Do not run both worktree-creation blocks against the same clone; Windows CI normally checks out the task commit directly. The PowerShell block is the native operator equivalent.

### 4.3 Fixture provenance checks

The new fixture should be authored in the task branch and recorded by path plus task commit SHA. It must derive its contract from:

- schema-6 example/layout: CocoaSkills `README.md:320-405` and `docs/skill-authoring.md:250-410` at the recorded base;
- real closed-worker test: `tests/test_builds_go_v1_fixture.py` from commit `53b4eb0a8496a9cf35890a4b609d7266a851615e`;
- candidate behavior: caller root at curator-spec `432eb2ee...`, manifest `sha256:12e58b...`;
- no network: checked-in standard-library-only Go source is sufficient; if a non-standard import is used, vendor every byte and set no fetch step.

Suggested deterministic tree:

```text
tests/fixtures/skill_go_e2e/
  SKILL.md
  agent-skill.json
  scripts/echo-script
  scripts/echo-script.cmd
  build/go.mod
  build/cmd/argv-exit/main.go
  build/cmd/second-tool/main.go
```

Before handoff, record:

```bash
git ls-tree -r --name-only HEAD tests/fixtures/skill_go_e2e tests/test_go_build_e2e.py
git diff --check "$BASE_SHA"..HEAD
git diff --name-status "$BASE_SHA"..HEAD
```

## 5. Exact platform and gate commands

Use a pre-materialized candidate checkout at `432eb2ee...`; `CURATOR_CONFORMANCE_ROOT` points to its `conformance/v1` directory. The committed CI pin remains unchanged.

### macOS

```bash
python3.11 -m venv .venv-e2e
.venv-e2e/bin/python -m pip install -e '.[dev]'
go version
test "$(go env GOVERSION | cut -d. -f1-2)" = 'go1.25'
export CURATOR_CONFORMANCE_ROOT=/absolute/path/to/curator-spec/conformance/v1
export CSK_GO_V1_MANAGER_EXECUTABLE="$PWD/.venv-e2e/bin/csk"
export CSK_GO_V1_GO_EXECUTABLE="$(command -v go)"
.venv-e2e/bin/python -m pytest -q tests/test_go_build_e2e.py tests/test_builds_go_v1_fixture.py
```

### Windows PowerShell

```powershell
py -3.11 -m venv .venv-e2e
.\.venv-e2e\Scripts\python.exe -m pip install -e '.[dev]'
go version
if ((go env GOVERSION) -notmatch '^go1\.25(?:\.|$)') { throw 'Go 1.25 family required' }
$env:CURATOR_CONFORMANCE_ROOT = 'C:\absolute\path\to\curator-spec\conformance\v1'
$env:CSK_GO_V1_MANAGER_EXECUTABLE = "$PWD\.venv-e2e\Scripts\csk.exe"
$env:CSK_GO_V1_GO_EXECUTABLE = (Get-Command go.exe).Source
.\.venv-e2e\Scripts\python.exe -m pytest -q tests/test_go_build_e2e.py tests/test_builds_go_v1_fixture.py
```

### Ubuntu

```bash
python3.11 -m venv .venv-e2e
.venv-e2e/bin/python -m pip install -e '.[dev]'
go version
test "$(go env GOVERSION | cut -d. -f1-2)" = 'go1.25'
export CURATOR_CONFORMANCE_ROOT=/absolute/path/to/curator-spec/conformance/v1
export CSK_GO_V1_MANAGER_EXECUTABLE="$PWD/.venv-e2e/bin/csk"
export CSK_GO_V1_GO_EXECUTABLE="$(command -v go)"
.venv-e2e/bin/python -m pytest -q \
  tests/test_go_build_e2e.py -k 'portable or unavailable_control'
```

The Ubuntu selection must contain a positive portable non-driver test and an expected fail-closed source-aware test. It is a green test command because the product failure is asserted; it must prove zero worker launches and zero artifact publication.

### Candidate authentication and full handoff gates

Run each as a standalone process and record its real exit code:

```bash
test "$(git -C /absolute/path/to/curator-spec rev-parse HEAD)" = 432eb2ee1fe2d6b271e37269f867c8851c325539
```

```bash
test "$(sha256sum "$CURATOR_CONFORMANCE_ROOT/manifest.json" | awk '{print $1}')" = 12e58b82579645ba1ccafba49d3e2dd3216005ddf37ae63c68a9fafd46773071
```

```bash
.venv-e2e/bin/python -m pytest -q
```

```bash
.venv-e2e/bin/python -m mypy
```

```bash
.venv-e2e/bin/python -m build
```

```bash
.venv-e2e/bin/python -m twine check dist/*
```

```bash
git diff --check "$BASE_SHA"..HEAD
```

```bash
test "$(git show HEAD:.github/workflows/ci.yml | rg -o 'ref: [0-9a-f]{40}' | head -n1)" = 'ref: 0c81c1f8d5321d822be2a2817b05aea03e656e15'
```

```bash
git diff --exit-code "$BASE_SHA"..HEAD -- .github/workflows/release.yml CHANGELOG.md
```

On Windows replace `sha256sum` with `(Get-FileHash -Algorithm SHA256 ...).Hash.ToLowerInvariant()` and use the PowerShell executables shown above. Hosted evidence must identify exact task head and all 12 native OS/Python jobs; build and mypy jobs remain separate.

## 6. Ordered handoff plan

1. **Dependency gate:** wait for `TASK-260720-12r55p` to be reviewer-accepted and PR 19 to merge; do not change `TASK-260720-3pemm6` while it remains blocked.
2. **Freeze base:** fetch/fast-forward the clean CocoaSkills main clone, prove `6e7742f0...` ancestry, record the resulting full `BASE_SHA`, and create the dedicated task worktree/branch with the commands in §4.
3. **Add fixture first:** commit the minimal mixed vendored fixture and a fixture-integrity helper (manifest, module, vendor/no-network, deterministic argv/exit behavior).
4. **Add the thin positive E2E spine:** public project install → no launch → explicit project shim launch; then global, hybrid/mixed and Windows `.cmd` variants. Make real Go mandatory on macOS/Windows.
5. **Add lifecycle rows in risk order:** cache/relevant mutation/irrelevant mutation; dry-run purity; build-two isolation; rollback and interrupted recovery; concurrent publisher; two-project preservation; repair/status.
6. **Add Ubuntu negative:** public source-aware attempt returns unavailable-control before worker launch/publication, plus portable non-driver coverage. No Linux success assertion.
7. **Wire CI:** deterministic Go 1.25 setup and platform-specific E2E selections across the existing Python matrix; preserve curator-spec ref `0c81c1f8...` and authenticate candidate bytes through the caller root.
8. **Run focused gates:** macOS and Windows real fixture/E2E commands, Ubuntu portable/fail-closed command, and candidate provenance checks. Fix product/test defects without skips, xfails, fake positive toolchains or platform bypasses.
9. **Run full gates:** full pytest, strict mypy, build, exact Twine check, diff whitespace, clean worktree, no-pin/no-release guards, then hosted exact-head CI. Record every exit code and any expected-red characterization honestly.
10. **Handoff evidence:** attach task-scoped outcome with base/head, fixture tree, candidate root/digest, per-platform commands/results, CI run URL, AC-to-test mapping, release-boundary proof and remaining risks. Only then route `TASK-260720-3pemm6` to review.

## 7. Risks and stop conditions

- A future PR 19 merge commit may differ from its current head. That is normal; only ancestry of `6e7742f0...` is required. Do not guess the future base SHA.
- Current Windows PR 19 jobs are still running. PR 19 acceptance/landing is a hard start gate, not evidence to waive native Windows E2E.
- Do not treat `tests/test_builds_go_v1_fixture.py` as satisfying install/launch E2E: its success scenario calls the driver directly and explicitly asserts the artifact was not run.
- Do not expand candidate-vector replay into this task. PR 19 owns that layer; this task must connect public CocoaSkills workflows to real platform effects.
- If a positive native path requires a fake toolchain, skip, xfail, `os.name` bypass, or changes the Linux claim boundary, stop and record the platform/architecture decision instead of forcing the test.

## 8. Fact-check log (commands run during this audit)

| Command / check | Exit | Result |
|---|---:|---|
| mandated `set_status(BUG-260802-3ibgu1, analysis)` | 0 | audit lifecycle started |
| CocoaSkills `rg --files -g AGENTS.md` | 1 | expected absence; no repository-local agent instructions |
| CocoaSkills `git status --porcelain=v2 --branch` | 0 | clean `main` `dacccaaf...`, `+0/-0` |
| CocoaSkills `git worktree list --porcelain` | 0 | PR 19 has a separate active worktree; no E2E task worktree exists |
| `gh pr view 19 ... baseRefOid ...` | 1 | exploratory query used an unsupported gh JSON field; rerun corrected |
| corrected `gh pr view 19 --json ...` | 0 | open/mergeable exact head; Windows jobs in progress |
| CocoaSkills `git ls-remote` for main, PR branch and rc.6 tag | 0 | exact two branch refs; no CocoaSkills rc.6 tag |
| curator-spec `git status --porcelain=v2 --branch` | 0 | clean main at `432eb2ee...`, `+0/-0` |
| `sha256sum conformance/v1/manifest.json` | 0 | exact `12e58b...` digest |
| `git diff --quiet 0c81c1f8... 432eb2ee... -- conformance release/1.0.0-rc.6.json` | 0 | candidate suite/metadata byte-equivalent |
| curator-spec `git ls-remote` main and rc.6/rc.5 tags | 0 | main `432eb2ee...`; rc.5 tag exists; rc.6 tag absent |
| `gh release view v1.0.0-rc.6 -R relux-works/curator-spec` | 1 | expected-red release-boundary check: `release not found` |
| `git merge-base --is-ancestor 53b4eb0... dacccaaf...` | 0 | existing native fixture commit is on accepted CocoaSkills main |
| task-board spawn status/directives | 0 | read-only directive observed; no cancellation/reroute |
| `.research`/board-resource SHA-256 comparison | 0 | both copies were byte-identical at the pre-validation attachment checkpoint |
| `task-board validate` (standalone) | 0 | CLI returned 0 but reported 1,227 pre-existing board issues (legacy broken links/status mismatches/missing resource payloads); none was introduced or repaired by this read-only task |

No product validation suite was run because this was a read-only readiness audit, not a change or acceptance review. Existing test/CI results are cited as evidence, never reported as gates executed by this audit.

## 9. Sources

1. Board briefs/resources: `TASK-260720-3pemm6`, `TASK-260720-12r55p`, and `.task-board/.resources/TASK-260720-3pemm6/TASK-260720-3pemm6_release-boundary-plan.md`.
2. CocoaSkills exact repository: `/Users/iv/Developer/Wildberries/cocoaskills`, commits `dacccaaf...`, `53b4eb0...`, PR 19 head `6e7742f0...`; source files cited above.
3. CocoaSkills PR 19 and hosted run 30743353816: https://github.com/ivanopcode/cocoaskills/pull/19 and https://github.com/ivanopcode/cocoaskills/actions/runs/30743353816.
4. curator-spec exact repository: `/Users/iv/Developer/ReluxWorks/curator-spec`, commit `432eb2ee...`; `protocol/core.md`, `profiles/manager.md`, `conformance/v1/manifest.json`, `conformance/v1/vectors/go-host-execution-policy.json`, and `release/1.0.0-rc.6.json`.
