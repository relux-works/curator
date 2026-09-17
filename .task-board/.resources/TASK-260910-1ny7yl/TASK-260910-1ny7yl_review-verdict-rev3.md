# TASK-260910-1ny7yl — revision 3 review verdict

Verdict: accepted. All three revision-2 corrections are closed. No blocking findings.

Candidate: CR-TASK-260910-1ny7yl-3, tree `48e7171819ea34b03c3945288e081ae448b1c694`, base `131952da694cfad10d31ba4ac878dd45a5f6875f`. Published patch SHA-256 independently matches `26b608bfef3353e3725a7e881d127a72984f388dcfd795391284f6892ad2625b`. Compared every tree file against the managed worktree: 28/28 byte-identical. Tests ran from a git archive of that tree in `/tmp/1ny7yl-review3`; venv outside the copy. No candidate edits or new candidate files.

Read campaign rules, task/rework briefs, producer results, published rev3 patch and validation log, security audit R3/P2, and curator-spec profile §5/6/9/10/11 plus protocol §5 and checkpoint vectors. Spec checkout HEAD is exactly `47c3c8cbd5da5e3fd6d99b0327d382c6a36494fe`; task-specific pin supersedes the older generic campaign pin.

| Review item | Evidence and result |
|---|---|
| Startup ordering and accepted keys | `app.py:747-763`: Store construction performs integrity verification and R2 boundary rebuild (`store.py:256`), then accepted `public_keys` and checkpoint enforcement, before returned app reaches `uvicorn.run` (`cli.py:357-359`). `app.py:717` verifies signature before `apply_startup_checkpoint`; boundary read beforehand is for diagnostics, not comparison. |
| Closed outcomes and latch | `checkpoint.py:85-103` handles below, equal-field mismatch and above-prefix reproduction; `store.py:1673-1712` latches through common integrity failure and preserves history; `app.py:209-223` exposes 503 diagnostic. Signature invalid and absent posture complete the four named diagnostics. Successful health envelope unchanged. |
| Real startup vectors | `test_protocol_conformance.py:769-968` constructs signed fixtures, measures discriminating predicates, runs `app_from_env`, health and authenticated writes, and validates logged boundaries/outcomes. 7/7 authoritative checkpoint vectors exercised; all previous recovery/R2 tests passed in full suite. |
| Logging correction | `app.py:653-669`, `cli.py:357`: JSON stderr sink at INFO established before enforcement. Committed console tests `test_registry.py:828,841` pass. Separate fresh-process installed-console probe independently observes 2/2 events (absent posture, successful signed comparison), before Uvicorn startup. |
| Root-only negative correction | `test_registry.py:1136` signs genuine prefix head with wrong root at boundary 8 while live is 10; real factory gives 503 inconsistent, authenticated write refuses, reopened history remains at 10. Root-only mutant is killed at HTTP assertion (200 instead of 503). |
| CLI restored-backup AC correction | `test_registry.py:954` restores backup at 7 below signed checkpoint 10, calls actual main with serve flag, intercepts only final Uvicorn run, checks health/write refusal and byte-identical database. Passed independently in full suite. Flag precedence/env fallback coverage at `:707` retained. |
| Rev1 harness repair | `test_registry.py:2250-2341` fixes manual epoch, stops existing verifier API before clock installation, refreshes then reads base; exact 20.0 and 20.001 assertions retained. No new production hook; clock and epoch live only in tests. Producer's float-rounding diagnosis is consistent with arithmetic and absent lifespan thread. |
| Scope/docs/CI | CI line 30 pins 47c3c8c; CHANGELOG line 36 names R3/P2 and diagnostics; README line 100 documents normative startup gate, checkpoint production/rotation, stderr; SECURITY line 23 and compose line 10 cover operation. verify-backup retained. No key rotation, P4, R4–R8 or spec implementation changes. |
| Rev2→rev3 scope | Applied published rev2 patch to separate base archive and compared: production delta only sink function and CLI call; tests add console probes, CLI restore, root mismatch; README/CHANGELOG clarify sink. |

## Independent validation

Shell `/bin/bash`, `set -o pipefail`; Python 3.14.6, `/tmp/1ny7yl-review3-venv/bin/python`; PATH prepends that venv so console probes use its installed script. Working directory `/tmp/1ny7yl-review3`.

```
export CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1
python -m pytest -q
166 passed, 2 warnings in 31.39s
exit 0
python -m mypy
Success: no issues found in 14 source files
exit 0
git diff --check  # managed worktree, read-only
exit 0
```

Warnings are third-party Starlette/httpx and AnyIO deprecations. No separate linter configured. Hosted build/Docker and six OS/Python lanes accepted from attached rev3-validation.log, run 35265453703, all success, gate exit 0; reviewer did not rerun hosted gate. Local runtime coverage is Python 3.14.6/macOS only.

## Narrowing attacks

Independent disposable copies, no mutations in candidate or clean validation copy. Commands use same interpreter/root and `python -m pytest -q tests/test_registry.py tests/test_protocol_conformance.py` with masks below.

| Mutation | Committed test result |
|---|---|
| Drop only above-prefix merkle_root equality, retaining head comparison | `-k startup_checkpoint`: 1 failed, 2 passed, 163 deselected; exit 1. `test_startup_checkpoint_above_with_root_only_mismatch_refuses` catches HTTP 200 vs 503. |
| Require signature verification only when checkpoint version differs from live (equal-version bypass) | Same mask: 1 failed, 2 passed, 163 deselected; exit 1. Shared real-startup vectors catch invalid signature admitted at equal version. |
| Drop only above-prefix head equality, retaining root comparison | Initial startup-only mask: 3 passed, exit 0. Broadened `-k checkpoint`: 1 failed, 8 passed, 157 deselected; exit 1, `test_compare_checkpoint_closed_diagnostics`. Head-only discrimination is currently helper-level; shared startup divergence changes both head and root. Reported coverage bound, not a claim of independent startup head-only negative coverage. |

3/3 attempted mutants killed by committed tests; 2/3 killed through real startup. This is a bounded attack set, not exhaustive mutation coverage. Required root-specific and signature attacks both discriminate at production startup.

## Console reproduction

Separate probe initializes a real three-record store, launches installed console command twice on port 0, captures output, requires completed startup, terminates each server boundedly. No app/factory/enforcement mocks. Events observed:

```
absent: {"configured":false,"event":"startup_checkpoint","posture":"checkpoint_not_configured"}
valid: {"checkpoint":{"head":"f398d7a031255c349758daa9333fe3661125295b8ac26c2b09991d2dcb56c374","log_size":3,"version":3},"configured":true,"event":"startup_checkpoint","live":{"head":"f398d7a031255c349758daa9333fe3661125295b8ac26c2b09991d2dcb56c374","log_size":3,"version":3},"result":"ok"}
console startup events: 2/2
probe exit 0
```

Full console and mutant transcripts attached separately. Unreadable/malformed checkpoint remains explicit startup configuration failure, never absence fallback. Reads during checkpoint refusal retain preexisting common integrity posture as explicitly required by task (readiness 503 and writes disabled).

Lifecycle: queried `task-board spawn goal "$TASK_BOARD_RUN_ID"` before verdict: run not goal-bound. Acceptance routes to integrating; no commit_ack, commit, done transition, or integration performed by reviewer.
