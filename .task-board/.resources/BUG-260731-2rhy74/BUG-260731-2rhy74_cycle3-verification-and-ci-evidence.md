# BUG-260731-2rhy74 — cycle 3: verification, cross-platform CI, and two out-of-scope blockers

**Role:** developer · **Run:** `RUN-260731-b4fd97` · **Date:** 2026-07-31
**Board status at handoff:** `to-review`
**Answers:** reviewer verdict `changes_requested` from `RUN-260731-46c299`

This cycle published nothing new. The delta from cycle 2 was already committed and
pushed; this run independently re-verified it, completed the two gates cycle 2 left
as unfilled placeholders (`FULL_SUITE_EXIT`, `CI_SECTION`), and ran the cross-platform
CI matrix to completion. It also resolved the coordinator's "CI diagnosis" note, whose
two proposed remedies both turned out to be wrong.

---

## 1. Reviewer finding 1 — Windows CRLF — closed on real Windows

Cycle 2 changed `serialize_install_marker()` to return `bytes` and moved both
`installer.py` write sites to `Path.write_bytes`. Cycle 1's reviewer could only
*simulate* Windows. This cycle has the real thing.

**On all four Windows CI cells (3.11, 3.12, 3.13, 3.14), all four marker tests
PASSED:**

```
tests/test_install.py::test_install_writes_marker_schema_2_bytes_for_a_schema_1_skill PASSED
tests/test_install.py::test_serialize_install_marker_renders_utf8_lf_bytes            PASSED
tests/test_protocol_conformance.py::test_shared_fixture_legacy_marker_v1_stays_readable PASSED
tests/test_protocol_conformance.py::test_shared_fixture_context_hash_and_marker         PASSED
```

The first of those reads the real `.csk-install.json` with `read_bytes()` and asserts
`b"\r\n" not in raw`. That is the exact assertion the reviewer predicted would fail on
all four Windows cells before the fix. It passes.

Local reproduction of the reviewer's simulation against the published fixture:

```text
serializer_returns_bytes           = True
serializer_matches_fixture         = True
serializer_contains_CR             = False
windows_text_write_matches_fixture = False      <- the old write_text path
fixture bytes                      = 787
windows text-write bytes           = 825
CRLF sequences in text-write       = 38
```

787 / 825 / 38 reproduce the reviewer's numbers exactly. Production no longer takes the
text path. Audit of every marker write site in `src/csk` (`grep` for
`.csk-install.json`) finds exactly two writers, both `write_bytes`; every other hit is
a read.

---

## 2. Reviewer finding 2 — publication — closed

| Repo | Branch | Commit | Remote verified |
| --- | --- | --- | --- |
| curator-spec | `task/BUG-260731-2rhy74-marker-v2-fixture` (PR **15**) | `0c81c1f8d5321d822be2a2817b05aea03e656e15` | `git ls-remote` matches; worktree clean |
| CocoaSkills | `task/TASK-260720-3t8nr3-transactional-project-hybrid` (PR **16**) | `8a02e179fe35205490f081a7caa2e191b524e534` | PR head matches; worktree clean |

`.github/workflows/ci.yml` conformance `ref:` is `0c81c1f8…`, consistent with
`tests/test_build_metadata.py::EXPECTED_MANIFEST_SHA256`.

---

## 3. Fixture identities — re-verified this cycle

| File | SHA-256 |
| --- | --- |
| `conformance/v1/expected/marker.json` (frozen legacy-read) | `80989f850887814ec09c724a7dd891ac7e2422d5fef7e31f330be3554aa9b28a` |
| `conformance/v1/expected/marker-v2.json` (writer golden) | `22117126c8932769636bd5ca1f1623f26a3c55d4bb9da3266acf3dc3a3b7fc2c` |
| `conformance/v1/manifest.json` | `12e58b82579645ba1ccafba49d3e2dd3216005ddf37ae63c68a9fafd46773071` |

Legacy fixture byte-identity proved by diff, not just by hash:
`git diff origin/release/v1.0.0-rc.6 -- conformance/v1/expected/marker.json` → empty,
exit **0**.

---

## 4. Evidence — commands and real exit codes

### curator-spec, at published `0c81c1f8`, worktree `.temp/BUG-260731-2rhy74/spec`

| Command | Exit | Result |
| --- | ---: | --- |
| `make release-check VERSION=1.0.0-rc.6` | **0** | `validated 42 schemas and 448 vector files`; `Ran 79 tests … OK`; `ok …/generate-vectors`; regenerate diff clean; `release gate passed for 1.0.0-rc.6 at 0c81c1f8…` |
| `gofmt -l tools` | **0** | no output |
| `git diff --check` | **0** | — |
| `git status --porcelain` | **0** | 0 lines — regeneration is idempotent |

`make release-check` runs validate + unittests + `go test` + generator + regenerate-check
+ the rc.6 gate as one chain, against the real committed branch.

### CocoaSkills, at published `8a02e17`, worktree `.temp/TASK-260720-3t8nr3/worktree`

Conformance root = the published spec checkout `.temp/BUG-260731-2rhy74/pin-published/conformance/v1`.

| Command | Exit | Result |
| --- | ---: | --- |
| `pytest -q` (full suite) | **0** | `1304 passed, 49 skipped in 1351.74s` |
| `python -m mypy` (strict, `src/csk`) | **0** | `Success: no issues found in 67 source files` |
| `pytest -q tests/test_install.py -k marker` | **0** | `5 passed, 53 deselected` |
| `pytest -q tests/test_protocol_conformance.py tests/test_build_metadata.py` | **0** | `167 passed` |
| `pytest -q tests/test_protocol_conformance.py` against rc.5 pin `f5d7673` | **1** | **expected red** — `AssertionError: conformance root …/pin-rc5/conformance/v1 publishes no expected/marker-v2.json`, `1 failed, 106 passed`. Fail-closed: a root without the writer golden fails, it does not skip. |

1304 is one more than cycle 1's 1303 — `test_serialize_install_marker_renders_utf8_lf_bytes`
was added in cycle 2.

### The specification's own implementation gate, run against the PR 16 manager

| Command | Exit | Result |
| --- | ---: | --- |
| `python tools/run_pytest_no_skips.py -q <cocoaskills>/tests/test_protocol_conformance.py` | **0** | `107 passed` |

`run_pytest_no_skips.py` is the harness `implementations.yml` uses; it also fails on
skips. The PR 16 manager satisfies it against the PR 15 suite.

### Lint

CocoaSkills declares no linter — no `[tool.ruff]`, no `ruff.toml`; CI runs pytest, mypy
and the package build only. `mypy --strict` is the type gate and is clean. On the
specification side `gofmt -l tools` and `git diff --check` are real CI gates and pass.

---

## 5. Cross-platform CI — PR 16, run `30594273278` at `8a02e17`

Run conclusion: **failure**. Full matrix:

| Cell | 3.11 | 3.12 | 3.13 | 3.14 |
| --- | --- | --- | --- | --- |
| ubuntu-latest | pass | pass | pass | pass |
| macos-latest | pass | pass | pass | pass |
| windows-latest | **fail** | **fail** | **fail** | **fail** |

`Type check / mypy strict`: pass. `Build artifacts`: skipped (gated on tests).

**Every Ubuntu and macOS cell is green, on all four Python versions.** The Windows cells
fail — but on nothing this task touches. Per-cell totals: 3.11 `45 failed, 1157 passed,
151 skipped`; 3.12 `45 failed, 1157 passed, 151 skipped`; 3.13 `45 failed, 1157 passed,
151 skipped`; 3.14 `34 failed, 1168 passed, 151 skipped`.

Across all three Windows logs the count of failures in
`test_protocol_conformance.py`, `test_build_metadata.py`, or any marker-byte test is
**zero**. The failing files are:

```
test_activation_modes.py  test_audit_cli.py   test_closure_install.py
test_dev_substitution.py  test_gc.py          test_global_install.py
test_hybrid_scope.py      test_install.py     test_mcp_dependencies.py
test_status.py
```

and the failure signatures are:

```
AssertionError: ['transaction target changed while digesting: C:\...\.csk-materialization-plan-...\home\runtime\<skill>\<sha>\bin\<cmd>.cmd']
assert not ["[WinError 5] Access is denied: 'C:\\Users\\runneradmin\\...'"]
AssertionError: assert not ['cache_publication_invalid: publication artifact source is not ...']
```

Every `transaction target changed while digesting` target is a `.cmd` command shim under
`home\runtime\…\bin\`. Not one is `.csk-install.json`.

### This is a pre-existing regression from PR 16's base commit, not from this task

| Fact | Evidence |
| --- | --- |
| CocoaSkills `main` `b3a5031` is green on **all four Windows cells** | CI run `30556125542`, conclusion `success` |
| The transaction engine (`721ca47`, `edbc871`) is already on `main` | `git merge-base --is-ancestor` → yes |
| `c4131bd` "feat(installer): make project installs transactional" (TASK-260720-3t8nr3) is **not** on `main` | `git merge-base --is-ancestor` → no |
| This task's commit `8a02e17` touches 6 files: `ci.yml`, `install_marker.py`, `installer.py`, and 3 test files | `git show --stat` |
| `8a02e17` touches **no** file in the failing subsystem — `transactions.py`, `shims.py`, runtime materialization, cache publication | same |
| The four marker tests pass on every completed Windows cell | job logs above |
| The previous PR 16 run at `c4131bd` had **every** Windows cell `cancelled` | run `30589736936` |

PR 16 = `main` + `c4131bd` + `8a02e17`. `main` is Windows-green; `8a02e17` touches
nothing in the failing area and its own tests pass on Windows. The regression came in
with `c4131bd`, and this is simply the first PR 16 run in which Windows cells were
allowed to finish rather than being cancelled.

**Raised as a separate board item** — see §7. Not fixed here: it is a different
subsystem, a different commit, and a different task, and folding a Windows
file-handle/locking fix for the transaction engine into a conformance-fixture bug is
exactly the kind of scope creep the stop-the-line rule exists to prevent.

---

## 6. The coordinator's "CI diagnosis" note — both remedies are wrong

The board note added before this run asked to fix curator-spec PR 15's red
`Implementations` job by advancing two pins in `.github/workflows/implementations.yml`.
Both were tested. Neither works, and neither is needed.

### 6a. Advancing the Curator pin does not fix it

The job fails at its *first* step, `Go manager shared suite`, in
`TestManagerLifecycleVectors`. `internal/interop/golden_test.go:487` hard-codes:

```go
if len(vector.LauncherCases) != 2 || len(vector.BootstrapCases) != 3 ||
   len(vector.UpgradeCases) != 3 || len(vector.DryRunCases) != 2 {
```

The rc.6 base commit `671888e` expanded `conformance/v1/vectors/manager-lifecycle.json`
`dry_run_cases` from 2 to 3, adding `compiled-cache-miss-is-read-only`. curator-spec
`main` still has 2, which is why the gate is green there.

Tested locally at the proposed pin — released Curator `main`/v0.13.0 `cfffd7cd`, in a
fresh worktree with submodules, against the PR 15 conformance root:

```
go test -count=1 ./internal/interop ./internal/closure ./internal/skillspec
→ exit 1
--- FAIL: TestManagerLifecycleVectors
    golden_test.go:488: manager lifecycle vector is incomplete: {... DryRunCases:[project-upgrade global-upgrade compiled-cache-miss-is-read-only]}
ok   .../internal/closure
ok   .../internal/skillspec
```

Surveying every pushed Curator branch, only two carry `TestManagerLifecycleVectors` —
`main` `cfffd7cd` and `agent/seamless-manager-lifecycle` `a78545cd` — and **both** assert
`!= 2`. There is no pinnable Curator commit that accepts the rc.6 vector. A Curator-side
code change must land first.

### 6b. This failure is inherited from rc.6, not introduced here

- `git diff --name-only 671888e HEAD` on the task branch touches **no** vector file.
- PR **14** (`release/v1.0.0-rc.6` → `main`), which does not contain the marker-v2
  commit, fails the same three `Implementations` jobs identically.

PR 15's `Specification` (ubuntu/macos/windows), `Formatting` and `Links` checks — the
ones this bug's AC names — all **pass**.

### 6c. Advancing the CocoaSkills pin is not needed, and would be wrong now

Tested: CocoaSkills `6fc2fd9` (the current pin) against the marker-v2 root →
exit **0**, `106 passed`. Its conformance test only ever reads `expected/marker.json`,
so the new writer golden goes *unexercised* rather than failing. This task breaks
nothing there.

Advancing it to PR 16 head `8a02e179` was evaluated and rejected for now:

- `6fc2fd9` is on CocoaSkills `main`; `8a02e179` is an unmerged PR head. The workflow's
  convention is to pin merged commits.
- CocoaSkills PR 16 already pins curator-spec PR 15's commit. Pinning PR 16's head back
  from PR 15 would make two unmerged PRs pin each other.
- PR 16 is currently Windows-red for the reason in §5; pinning it would import that
  redness into the specification repo.

Correct sequencing: PR 16 lands on CocoaSkills `main` → a separate change advances the
`implementations.yml` pin to that merged commit, which also starts exercising
`expected/marker-v2.json`. Proof it will pass when it happens is in §4:
`run_pytest_no_skips.py` → `107 passed`, exit 0.

---

## 7. Follow-up items raised

1. **`BUG-260731-1rldqv` windows-transactional-install-regression** (§5) — 34–45 failures
   per Windows cell, entering with `c4131bd`; blocks PR 16 merge.
2. **`BUG-260731-3gm8kc` curator-interop-lifecycle-vector-gate** (§6a) — blocks
   `Implementations` on both PR 15 and PR 14; needs a published Curator revision.

Both are children of `STORY-260720-35dck7`, created by this run with the full evidence
above in their descriptions.

---

## 9. What this hands to review

Delivered and verified:

- The AC's specification half in full: frozen legacy fixture, distinct generated
  marker-v2 writer fixture, generator, manifest hashes, validator tests,
  regenerate-check and the rc.6 release gate — all green, published at `0c81c1f8`.
- The AC's consumer half in full: PR 16 reads the new fixture directly, compares
  byte-semantic writer output, fails closed without it, and keeps legacy marker-v1
  readable — all green, published at `8a02e17`.
- Local focused, full-suite and mypy gates: green.
- Cross-platform: the marker scope is green on Ubuntu, macOS **and Windows**, all four
  Python versions.

Not delivered, with reasons:

- **PR 16's overall CI is not green**, because all four Windows cells fail on the
  pre-existing transactional-install regression from `c4131bd` (§5). Out of scope for a
  conformance-fixture bug; tracked as `BUG-260731-1rldqv`.
- **PR 15's `Implementations` job is not green**, because the rc.6 lifecycle vector has
  a third dry-run case that no published Curator commit accepts (§6a). Inherited from
  rc.6, fails PR 14 identically, needs a Curator-side change; tracked as
  `BUG-260731-3gm8kc`. PR 15's `Specification`, `Formatting` and `Links` checks pass.
- **No pin in `implementations.yml` was changed** (§6a, §6c) — neither proposed advance
  is correct right now.

---

## 8. Scratch artifacts

- `.temp/BUG-260731-2rhy74/spec` — curator-spec worktree at `0c81c1f8`
- `.temp/BUG-260731-2rhy74/pin-published` — clean checkout of `0c81c1f8`, used as the conformance root
- `.temp/BUG-260731-2rhy74/pin-rc5` — rc.5 checkout used to prove fail-closed
- `.temp/BUG-260731-2rhy74/curator-main` — Curator worktree at `cfffd7cd` (§6a)
- `/Users/iv/Developer/intranet/cocoaskills/.temp/BUG-260731-2rhy74/pin-old` — CocoaSkills `6fc2fd9` (§6c)
- `.temp/BUG-260731-2rhy74/venv` — task venv (cocoaskills `[dev]` + `jsonschema==4.25.1`)
- `.temp/BUG-260731-2rhy74/logs/` — every log referenced above, including the three
  Windows CI job logs
