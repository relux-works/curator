# CocoaSkills integrated verification readiness audit

**Board item:** `BUG-260802-3fbn47` — csk-integrated-verification-readiness-audit  
**Delivery target:** `TASK-260720-3s27te` — Verify integrated CocoaSkills schema v6 and go-v1 implementation  
**Hard start dependency:** `TASK-260720-3pemm6` — Add cross-platform Go build end-to-end tests against rc.6  
**Audit date:** 2026-08-02  
**Mode:** read-only inspection of the board, CocoaSkills, and curator-spec; no product repository edit, tag, release, claim, pin change, or blocked-task status change.

## Executive finding

`TASK-260720-3s27te` has an executable verification plan, but it must not start yet. Its base SHA cannot be named honestly until `TASK-260720-3pemm6` is independently accepted (`status=done`), its accepted head is landed on CocoaSkills `origin/main`, and a clean canonical clone is fast-forwarded to that main.

The accepted protocol input is an **authenticated rc.6 candidate**, not a release:

- curator-spec commit `432eb2ee1fe2d6b271e37269f867c8851c325539` is the local and remote main commit and GitHub reports its commit signature as verified;
- `conformance/v1/manifest.json` hashes to `sha256:12e58b82579645ba1ccafba49d3e2dd3216005ddf37ae63c68a9fafd46773071`;
- `release/1.0.0-rc.6.json` says `committed_release_pin_advanced=false`, `claims_emitted=[]`, macOS and Windows are `pending-downstream-native-evidence`, and Linux is excluded;
- neither an upstream `v1.0.0-rc.6` tag nor GitHub Release exists;
- CocoaSkills CI remains pinned to immutable curator-spec `0c81c1f8d5321d822be2a2817b05aea03e656e15`, whose `conformance/` tree and `release/1.0.0-rc.6.json` are byte-equivalent to `432eb2ee...`. Candidate execution uses caller-supplied `CURATOR_CONFORMANCE_ROOT`; this task must not advance the committed pin.

Therefore the only defensible platform statement from this verification is: **native go-v1 behavior was evidenced on macOS and Windows at the recorded CocoaSkills head; Ubuntu evidenced portable non-driver behavior and source-aware go-v1 fail-closed behavior.** It is not an rc.6 release claim, conformance claim, or Linux driver-success claim.

## 1. Hard start gate and clean base

Run every block as standalone commands and preserve each real exit code. Do not create the verification worktree until all dependency assertions exit 0.

```bash
export EVIDENCE_ROOT=/absolute/evidence/TASK-260720-3s27te
mkdir -p "$EVIDENCE_ROOT"
task-board q --format json 'get(TASK-260720-3pemm6) { full }' > "$EVIDENCE_ROOT/dependency.json"
jq -e '.status == "done" and .isBlocked == false and (.outcomeResources | length > 0)' "$EVIDENCE_ROOT/dependency.json"
```

In the canonical CocoaSkills clone:

```bash
cd /Users/iv/Developer/Wildberries/cocoaskills
git fetch origin main
test -z "$(git status --porcelain)"
git switch main
git merge --ff-only origin/main
export CSK_BASE_SHA="$(git rev-parse HEAD)"
test "$CSK_BASE_SHA" = "$(git rev-parse origin/main)"
printf '%s\n' "$CSK_BASE_SHA" > "$EVIDENCE_ROOT/cocoaskills-base.sha"
```

The accepted `TASK-260720-3pemm6` outcome must name its signed/accepted head as `E2E_ACCEPTED_SHA`. Authenticate it rather than guessing it:

```bash
git merge-base --is-ancestor "$E2E_ACCEPTED_SHA" "$CSK_BASE_SHA"
git verify-commit "$E2E_ACCEPTED_SHA"
git log -1 --format='%H %P %s' "$CSK_BASE_SHA" > "$EVIDENCE_ROOT/cocoaskills-base.txt"
git status --porcelain=v2 --branch > "$EVIDENCE_ROOT/cocoaskills-base-status.txt"
```

Then create the isolated verification branch/worktree:

```bash
git worktree add -b task/TASK-260720-3s27te-integrated-verification \
  .temp/TASK-260720-3s27te/worktree "$CSK_BASE_SHA"
cd .temp/TASK-260720-3s27te/worktree
test -z "$(git status --porcelain)"
test "$(git rev-parse HEAD)" = "$CSK_BASE_SHA"
```

PowerShell equivalents for the native Windows checkout:

```powershell
$CskSha = (git rev-parse HEAD).Trim()
if ($CskSha -ne (git rev-parse origin/main).Trim()) { throw 'checkout is not accepted origin/main' }
if ((git status --porcelain)) { throw 'checkout is dirty' }
git merge-base --is-ancestor $env:E2E_ACCEPTED_SHA $CskSha
if ($LASTEXITCODE -ne 0) { throw 'accepted E2E head is absent' }
$CskSha | Set-Content artifacts/TASK-260720-3s27te/cocoaskills-head.sha
```

**Stop:** dependency status is not `done`; outcome evidence is absent; accepted E2E head is not an ancestor; local main is dirty/diverged; signatures fail; or the accepted head has not landed. `to-review`, an open PR, or merely green CI is not acceptance.

## 2. rc.6 candidate authentication

Materialize curator-spec in a separate clean checkout at the exact commit, then run:

```bash
export CURATOR_SPEC=/absolute/path/to/curator-spec
export CURATOR_CONFORMANCE_ROOT="$CURATOR_SPEC/conformance/v1"
test "$(git -C "$CURATOR_SPEC" rev-parse HEAD)" = 432eb2ee1fe2d6b271e37269f867c8851c325539
test -z "$(git -C "$CURATOR_SPEC" status --porcelain)"
git -C "$CURATOR_SPEC" rev-parse HEAD > "$EVIDENCE_ROOT/curator-spec-head.sha"
git -C "$CURATOR_SPEC" status --porcelain=v2 --branch > "$EVIDENCE_ROOT/curator-spec-status.txt"
gh api repos/relux-works/curator-spec/commits/432eb2ee1fe2d6b271e37269f867c8851c325539 > "$EVIDENCE_ROOT/curator-spec-github-commit.json"
jq -e '.sha == "432eb2ee1fe2d6b271e37269f867c8851c325539" and .commit.verification.verified == true and .commit.verification.reason == "valid"' "$EVIDENCE_ROOT/curator-spec-github-commit.json"
python -c 'import hashlib,os,pathlib; p=pathlib.Path(os.environ["CURATOR_CONFORMANCE_ROOT"])/"manifest.json"; d=hashlib.sha256(p.read_bytes()).hexdigest(); print(d); assert d=="12e58b82579645ba1ccafba49d3e2dd3216005ddf37ae63c68a9fafd46773071"' > "$EVIDENCE_ROOT/manifest.sha256"
python "$CURATOR_SPEC/tools/validate.py"
git -C "$CURATOR_SPEC" diff --exit-code 0c81c1f8d5321d822be2a2817b05aea03e656e15 432eb2ee1fe2d6b271e37269f867c8851c325539 -- conformance release/1.0.0-rc.6.json
```

On Windows, replace the hash gate with:

```powershell
$Digest = (Get-FileHash -Algorithm SHA256 "$env:CURATOR_CONFORMANCE_ROOT\manifest.json").Hash.ToLowerInvariant()
if ($Digest -ne '12e58b82579645ba1ccafba49d3e2dd3216005ddf37ae63c68a9fafd46773071') { throw 'candidate digest mismatch' }
```

Local `git verify-commit` is an additional green gate only where the GitHub signing key is present; the GitHub API verification gate above is mandatory and portable. Evidence: `curator-spec-head.sha`, `curator-spec-status.txt`, `curator-spec-github-commit.json`, `manifest.sha256`, `validate.log`, and `candidate-byte-equivalence.log`.

**Stop:** any identity, digest, manifest validation, or byte-equivalence check fails. Never fall back to a branch, mutable tag, another worktree, or the committed rc.5 release root.

## 3. Exact platform matrix

Use Go family 1.25, the recorded CocoaSkills checkout, and the authenticated caller root. `TASK-260720-3pemm6` may refine the new E2E filename/selectors; substitute only the exact accepted paths recorded in its outcome, and record that substitution in the verification report.

### macOS native go-v1

Run for Python 3.11, 3.12, 3.13, and 3.14 (shown for 3.11):

```bash
python3.11 -m venv .venv-verify-311
.venv-verify-311/bin/python -m pip install -e '.[dev]'
case "$(go env GOVERSION)" in go1.25|go1.25.*) ;; *) exit 1 ;; esac
export CSK_GO_V1_MANAGER_EXECUTABLE="$PWD/.venv-verify-311/bin/csk"
export CSK_GO_V1_GO_EXECUTABLE="$(command -v go)"
.venv-verify-311/bin/python "$CURATOR_SPEC/tools/run_pytest_no_skips.py" -v \
  tests/test_go_build_e2e.py tests/test_builds_go_v1_fixture.py \
  --junitxml="$EVIDENCE_ROOT/macos-py311-native.xml"
```

Required evidence from the focused suite: real compile and explicit launch; exact argv/stdout/stderr/exit propagation; project/global/hybrid/mixed Unix shims; cache hit; relevant/irrelevant mutations; build-two isolation; compiler-free dry-run purity; target-swap rollback; interrupted recovery; concurrent publication identity; two-project preservation; status/repair; and native-control/capability evidence. The no-skips runner makes any skip in this required slice a gate failure.

### Windows native go-v1

Run for Python 3.11–3.14 (shown for 3.11):

```powershell
py -3.11 -m venv .venv-verify-311
.\.venv-verify-311\Scripts\python.exe -m pip install -e '.[dev]'
if ((go env GOVERSION) -notmatch '^go1\.25(?:\.|$)') { throw 'Go 1.25 family required' }
$env:CSK_GO_V1_MANAGER_EXECUTABLE = "$PWD\.venv-verify-311\Scripts\csk.exe"
$env:CSK_GO_V1_GO_EXECUTABLE = (Get-Command go.exe).Source
.\.venv-verify-311\Scripts\python.exe "$env:CURATOR_SPEC\tools\run_pytest_no_skips.py" -v `
  tests/test_go_build_e2e.py tests/test_builds_go_v1_fixture.py `
  --junitxml="$env:EVIDENCE_ROOT\windows-py311-native.xml"
```

The same behavioral rows are required through Windows `.cmd` project/global shims and the protected Windows cache. Fake toolchains are acceptable only for explicitly negative timing/failure injection; they are not positive native evidence.

### Ubuntu portable and fail-closed

Run for Python 3.11–3.14 (shown for 3.11):

```bash
python3.11 -m venv .venv-verify-311
.venv-verify-311/bin/python -m pip install -e '.[dev]'
case "$(go env GOVERSION)" in go1.25|go1.25.*) ;; *) exit 1 ;; esac
export CSK_GO_V1_MANAGER_EXECUTABLE="$PWD/.venv-verify-311/bin/csk"
export CSK_GO_V1_GO_EXECUTABLE="$(command -v go)"
.venv-verify-311/bin/python "$CURATOR_SPEC/tools/run_pytest_no_skips.py" -v \
  tests/test_go_build_e2e.py -k 'portable or unavailable_control' \
  --junitxml="$EVIDENCE_ROOT/ubuntu-py311-portable-fail-closed.xml"
```

This must be green because tests assert the product's refusal. Evidence must show `build_execution_control_unavailable` before worker launch, no compiler/artifact execution, and no cache/artifact/receipt/marker/shim publication. Do not run or describe a Linux positive source-aware go-v1 case.

## 4. Full repository gates

Run after the focused platform commands, from the same exact task head. Use `-ra` and JUnit so every skip is reviewable.

```bash
python -m pytest -v -ra --junitxml="$EVIDENCE_ROOT/pytest-full.xml"
```

```bash
python -m mypy
```

```bash
mkdir -p "$EVIDENCE_ROOT/dist"
test -z "$(find "$EVIDENCE_ROOT/dist" -mindepth 1 -maxdepth 1 -print -quit)"
python -m build --outdir "$EVIDENCE_ROOT/dist"
```

```bash
python -m twine check "$EVIDENCE_ROOT"/dist/*
```

```bash
git diff --check "$CSK_BASE_SHA"..HEAD
```

```bash
git status --porcelain=v2 --branch
```

Full-suite platform skips are not automatically defects, but must be inventoried from `pytest-full.xml` as `nodeid`, reason, platform, and classification. The only permitted classifications are an explicit opposite-platform test or a genuinely unavailable shell/filesystem host capability. Any skip in the required E2E slices, candidate-root tests, or a new/changed test is unexpected and stops the handoff. Do not hide it with `-k`, `xfail`, `--ignore`, fake positive toolchains, or an OS bypass.

## 5. Hosted exact-head CI gate

The CocoaSkills workflow currently defines 12 test jobs (three OSes × Python 3.11–3.14), strict mypy, and a build/Twine job. All 14 must be terminal-success for the exact `CSK_HEAD_SHA` being reported.

```bash
CSK_HEAD_SHA="$(git rev-parse HEAD)"
gh run list --repo ivanopcode/cocoaskills --workflow CI --commit "$CSK_HEAD_SHA" \
  --json databaseId,headSha,status,conclusion,url > "$EVIDENCE_ROOT/ci-runs.json"
jq -e --arg sha "$CSK_HEAD_SHA" '[.[] | select(.headSha == $sha and .conclusion == "success")] | length >= 1' \
  "$EVIDENCE_ROOT/ci-runs.json"
gh run view "$CI_RUN_ID" --repo ivanopcode/cocoaskills \
  --json headSha,status,conclusion,url,jobs > "$EVIDENCE_ROOT/ci-run.json"
jq -e --arg sha "$CSK_HEAD_SHA" \
  '.headSha == $sha and .status == "completed" and .conclusion == "success" and (.jobs | length == 14) and all(.jobs[]; .status == "completed" and .conclusion == "success")' \
  "$EVIDENCE_ROOT/ci-run.json"
```

If `TASK-260720-3pemm6` legitimately changes the job count, replace `14` only with a checked-in workflow-derived inventory and explain the delta. Never accept a green run for a different SHA, cancelled job, skipped job, neutral conclusion, partial matrix, or superseded run.

## 6. Coverage map: criterion to command and evidence

| Integrated criterion | Exact command / selector | Required evidence artifact |
|---|---|---|
| Clean accepted base | dependency `jq` gate; clean fast-forward; accepted-head ancestry; `git verify-commit` | `dependency.json`, `cocoaskills-base.sha`, `cocoaskills-base-status.txt`, signature/ancestry log |
| Exact rc.6 candidate | §2 SHA, signature, manifest hash, validator, byte-equivalence commands | candidate SHA/status/signature, `manifest.sha256`, validator and equivalence logs |
| Schema 6 and legacy parsing | full pytest plus `tests/test_protocol_conformance.py::test_rc6_generated_schema_case_is_consumed` and `tests/test_skillspec.py` | JUnit rows and AC matrix |
| Required build rejections | `test_rc6_build_driver_rejection_case`, condition/observed-field mutation guards, `tests/test_builds_go_v1.py` graph/directive/environment negatives | candidate JUnit containing all 77 rejection parameter rows and mutation-sensitive guards |
| Source/toolchain/cache identities | `test_rc6_build_source_case`, `test_rc6_toolchain_case`, execution-policy identity tests, build metadata/cache suites | JUnit plus digest/key/receipt/marker row mapping |
| Portable execution policy | mandatory-control, identity/protocol, package-influence, failure-boundary tests | candidate JUnit; Ubuntu fail-closed JUnit |
| Native controls/capabilities | inventory-entry/exhaustiveness, capability-evidence, deferred-hardened guards | candidate JUnit; macOS/Windows native JUnit; no hardened-profile claim |
| Lifecycle ordering | `test_rc6_manager_lifecycle_case` and 378-leaf mutation-sensitivity gate | candidate JUnit with all 32 lifecycle cases and mutation guard |
| Dry-run purity | lifecycle dry-run/planning/upgrade transient-write bindings plus accepted E2E dry-run case | candidate and platform JUnit; before/after state digest in task report |
| Transactions/rollback/recovery | rollback mutation guard, transaction post-restore guard, recovery/generation/backup bindings, accepted E2E cases | candidate and platform JUnit; injected-failure state inventory |
| Cache/currentness/status/repair/GC | lifecycle bindings plus `test_build_currentness.py`, cache/GC suites, accepted E2E status/repair | full/platform JUnit and state/JSON snapshots |
| Launch and shims | `tests/test_build_activation.py`, real fixture, accepted E2E project/global/hybrid/mixed cases | macOS/Windows no-skip JUnit with argv and exit evidence |
| Ubuntu boundary | accepted E2E `portable or unavailable_control` selection | Ubuntu no-skip JUnit proving refusal and zero launch/publication |
| Full quality gates | §4 standalone pytest/mypy/build/Twine/diff/status commands | one raw log per command, real exit code, dist hashes |
| Hosted matrix | §5 exact-head `gh`/`jq` gates | `ci-runs.json`, `ci-run.json`, run URL, exact head, 14-success inventory |
| No unexpected skips | no-skips runner for required slices; full JUnit skip inventory and base/head set comparison | per-platform required JUnit; `skip-inventory.tsv`; `skip-diff.tsv` empty or justified |
| No pin/release/claim | §7 guards | before/after refs/releases plus protected-path and committed-pin logs |

The task report must also include every `STORY-260720-1uv5gi` criterion: shared schema/build vectors; valid build/launch; fixed Go environment; cache identity and rebuild rules; dry-run purity; rollback; unsafe/unsupported fail-closed; full Python tests and strict typing. The selectors above map each criterion without relying on an earlier task's claim.

## 7. No-release, no-pin, and no-claim guards

Capture external state before any verification work and compare it at handoff:

```bash
git ls-remote --tags https://github.com/ivanopcode/cocoaskills.git > "$EVIDENCE_ROOT/csk-tags.before"
gh release list --repo ivanopcode/cocoaskills --limit 100 --json tagName,isDraft,isPrerelease,publishedAt > "$EVIDENCE_ROOT/csk-releases.before.json"
git ls-remote --tags https://github.com/relux-works/curator-spec.git > "$EVIDENCE_ROOT/spec-tags.before"
gh release list --repo relux-works/curator-spec --limit 100 --json tagName,isDraft,isPrerelease,publishedAt > "$EVIDENCE_ROOT/spec-releases.before.json"
```

Repeat into `.after` files, then:

```bash
cmp "$EVIDENCE_ROOT/csk-tags.before" "$EVIDENCE_ROOT/csk-tags.after"
cmp "$EVIDENCE_ROOT/csk-releases.before.json" "$EVIDENCE_ROOT/csk-releases.after.json"
cmp "$EVIDENCE_ROOT/spec-tags.before" "$EVIDENCE_ROOT/spec-tags.after"
cmp "$EVIDENCE_ROOT/spec-releases.before.json" "$EVIDENCE_ROOT/spec-releases.after.json"
```

Guard repository surfaces and the committed suite pin:

```bash
git diff --exit-code "$CSK_BASE_SHA"..HEAD -- .github/workflows/release.yml CHANGELOG.md
```

```bash
python - <<'PY'
import re, subprocess
expected = '0c81c1f8d5321d822be2a2817b05aea03e656e15'
base = subprocess.check_output(['git', 'show', f'{subprocess.check_output(["git", "merge-base", "HEAD", "origin/main"], text=True).strip()}:.github/workflows/ci.yml'], text=True)
head = subprocess.check_output(['git', 'show', 'HEAD:.github/workflows/ci.yml'], text=True)
pattern = re.compile(r'ref:\s*([0-9a-f]{40})')
assert expected in pattern.findall(base)
assert expected in pattern.findall(head)
PY
```

```bash
bash -o pipefail -c 'git diff --unified=0 "$CSK_BASE_SHA"..HEAD -- . | rg -n "^\\+.*(claims_emitted|conformance.claim|v1\\.0\\.0-rc\\.6|softprops/action-gh-release|git tag|gh release)"'
```

The last command is an **expected-red diagnostic** when no suspicious additions exist: its correct result is exit 1 with no matches, and it must be recorded as a non-zero expected absence check, not a passing green gate. Any match requires manual disposition and is a stop until proven harmless. The four `cmp` commands are the authoritative green external-state gates.

## 8. Ordered handoff

1. Wait for `TASK-260720-3pemm6` to be accepted `done`, with outcome evidence and its accepted signed head landed on CocoaSkills main.
2. Freeze the clean base and verify accepted-head ancestry; create a fresh task worktree. Never reuse PR 19 or E2E task worktrees.
3. Authenticate curator-spec `432eb2ee...`, manifest `12e58b...`, candidate validation, and byte equivalence to the preserved committed pin.
4. Snapshot tag/release state and the protected repository surfaces.
5. Run candidate-root conformance and the macOS/Windows/Ubuntu focused gates; required platform slices use the no-skips runner.
6. Run full pytest with JUnit/skip inventory, strict mypy, build, exact Twine, diff, and clean-tree gates as standalone processes.
7. Verify exact-head hosted CI: every matrix, typecheck, and build job terminal-success.
8. Re-run provenance/no-release/no-pin/no-claim guards and compare before/after state.
9. Publish `TASK-260720-3s27te_results.md` with base/head SHAs, candidate SHA/digest, commands and real exits, log/JUnit hashes, skip inventory, hosted run, platform-boundary wording, and the criterion matrix.
10. Route any substantive defect to its owning task. Only if every gate and evidence row is satisfied may `TASK-260720-3s27te` be handed to review; it still must not tag, release, change the committed suite pin, or emit a conformance claim.

## 9. Stop conditions

- `TASK-260720-3pemm6` is not accepted `done`, its outcome is absent, or its accepted head is not on the clean base.
- Candidate SHA/signature/digest/validator/equivalence disagrees with the authenticated values.
- Any required native or Ubuntu slice skips, xfails, is deselected unintentionally, uses a fake positive toolchain, or runs on the wrong platform.
- macOS/Windows evidence lacks real Go 1.25 compile plus explicit launch; Ubuntu evidence reports a driver success instead of fail-closed behavior.
- Full pytest, strict mypy, build, Twine, diff, clean-tree, or exact-head hosted CI exits non-zero or is incomplete.
- A new/unexplained skip appears, a required test disappears, or the task changes tests merely to avoid real behavior.
- A tag/release/ref snapshot changes; the committed curator-spec checkout ref changes; release workflow/changelog/claim surfaces change; or language implies rc.6 is released/qualified.
- A defect requires redesign outside narrow integration correction. Record it against the owning task rather than weakening verification.

## 10. Fact-check record

| Command / source | Exit | Finding |
|---|---:|---|
| required `set_status(BUG-260802-3fbn47, analysis)` | 0 | research lifecycle started |
| board full queries for bug, target, dependency, docs task, parent story | 0 | target is blocked by `TASK-260720-3pemm6`; docs dependency is accepted; exact AC/scope captured |
| CocoaSkills `git status`, `rev-parse HEAD`, `origin/main`, signature | 0 | clean main at `dacccaaf...`, equal to local origin/main at audit time; future verification base remains unknown |
| `gh pr view/list` for CocoaSkills PR 19 | 0 | open exact head `6e7742f...`; Ubuntu/macOS/mypy green and Windows still running at observation time |
| curator-spec SHA/status/signature inspection | 0 | clean main `432eb2ee...`; local GPG key unavailable, but GitHub API reports `verified=true`, `reason=valid` |
| manifest `shasum -a 256` | 0 | exact digest `12e58b...` |
| `git diff --quiet 0c81c1f8... 432eb2ee... -- conformance release/1.0.0-rc.6.json` | 0 | candidate bytes/metadata equivalent across the two immutable commits |
| curator-spec remote rc.6 tag lookup | 0 | no matching tag |
| curator-spec GitHub Release list | 0 | rc.5 is latest listed; no rc.6 release |
| first `gh pr list` with unsupported `baseRefOid` field | 1 | exploratory query error; corrected query exited 0 |
| initial `python -m pytest --collect-only ...` in PR 19 worktree | 127 | expected environment failure: `python` is absent from PATH; no gate was claimed; source/test inventory was inspected directly |
| `task-board validate` | 0 | diagnostic reported 1,226 pre-existing board issues, chiefly legacy broken links/status mismatches and missing historical spawn-log payloads; none was created or repaired by this read-only audit |
| `task-board spawn directives "$TASK_BOARD_RUN_ID"` | 0 | read-only/no-release directive acknowledged |

## References

- Board: `STORY-260720-1uv5gi`, `TASK-260720-3pemm6`, `TASK-260720-3s27te`, and `TASK-260720-akf5kh` structured `full` queries on 2026-08-02.
- Board outcomes: `.task-board/.resources/STORY-260720-1uv5gi/STORY-260720-1uv5gi_decomposition-plan.md`, `STORY-260720-1uv5gi_17-task-audit.md`, and `.task-board/.resources/TASK-260720-3pemm6/TASK-260720-3pemm6_release-boundary-plan.md`.
- CocoaSkills: `pyproject.toml`, `.github/workflows/ci.yml`, `README.md`, and accepted PR 19 tests at exact inspected head `6e7742f0d28ad95ddd7d8e92364b84062571ad0b`.
- Candidate consumer tests: `tests/test_protocol_conformance.py`, `tests/test_build_metadata.py`, `tests/test_builds_go_v1.py`, `tests/test_builds_go_v1_fixture.py`, and `tests/protocol_lifecycle_observations.py` in the PR 19 worktree.
- curator-spec: commit `432eb2ee1fe2d6b271e37269f867c8851c325539`; `conformance/v1/manifest.json`; `release/1.0.0-rc.6.json`; `conformance/v1/vectors/go-host-execution-policy.json`; `conformance/v1/vectors/conformance-claim-v3-qualification.json`; `tools/validate.py`; `tools/run_pytest_no_skips.py`; `.github/workflows/ci.yml`; `.github/workflows/release.yml`.
