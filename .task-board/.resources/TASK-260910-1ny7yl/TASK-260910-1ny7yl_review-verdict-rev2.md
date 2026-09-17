# TASK-260910-1ny7yl — revision 2 review verdict

Verdict: **changes_requested**; route to `to-dev`. No human decision or external blocker.

Reviewed candidate tree `ee0873925c60ab4662a0ce99390a4a903d6afca6` against base `131952da694cfad10d31ba4ac878dd45a5f6875f`. Published patch SHA256 verified: `b50ea50c16b895f7ccf3dacbc69dd55eed54fc59e119bf9a5b9810e29eb46dc2`. All tracked candidate blobs match the live worktree, verified again at review end. No candidate edits, caches, or reviewer files added. Tests and attacks ran in `/tmp/1ny7yl-review` and separate mutant copies, using `/tmp/1ny7yl-review-venv`.

Read campaign rules, task and review briefs, producer results, both published patches, revision-2 validation log, profile §§5/6/9/10/11 and protocol §5, and all checkpoint vectors. Specific task pin `47c3c8cbd5da5e3fd6d99b0327d382c6a36494fe` supersedes the campaign's older generic pin. Local spec HEAD matches it. `task-board spawn goal` reports this reviewer run is not goal-bound.

## Required corrections

1. **P2 — Required startup posture/outcome is silently dropped by the real CLI.** `src/csk_registry/app.py:725` logs no-checkpoint and successful comparisons at INFO, but `cli.py:355` evaluates `app_from_env()` before `uvicorn.run` configures logging. The audit logger inherits WARNING with no application handler. Running the installed console entry point in fresh subprocesses with no checkpoint and with a valid signed empty-store checkpoint reached Uvicorn startup/listener successfully but emitted zero `startup_checkpoint` events (0/2 expected). The tests use `caplog.at_level(INFO)` at `tests/test_protocol_conformance.py:878` and therefore hide the deployment failure. Configure a real audit/startup logging sink and level before startup enforcement, and add CLI subprocess assertions for both posture and successful comparison. Preserve structured refusal events. Profile §6/§9 explicitly requires these records.

2. **P2 — Above-checkpoint Merkle-root refusal has no discriminating negative test.** Removing only `and prefix.merkle_root == checkpoint.merkle_root` at `checkpoint.py:100` leaves all **162 tests passing**. Existing above-divergence fixtures change the head too; the comparator's extra diverged fixture (`tests/test_registry.py:913`) likewise changes only the head. Add a signed checkpoint with the genuine prefix head but a different Merkle root while the live store is above it; drive startup and assert inconsistent diagnostic, non-ready, and disabled writes. Require this narrowing mutant to fail. Production code currently contains the correct comparison; this is a demonstrated security-gate coverage gap, not a claim that the unmutated comparator admits it.

3. **P2 — The requested restored-database scenario does not traverse the CLI.** `tests/test_registry.py:790` directly constructs `TestClient(app_from_env())`; the CLI flag test at `:722` replaces `app_from_env` with a sentinel. Thus the restore scenario covers the factory but not CLI-to-factory wiring (0/1 required end-to-end CLI restore scenarios). Add a test that restores the older backup, invokes actual `main([...,'serve','--checkpoint',...])` or the console subprocess, leaves factory/enforcement real, and observes health/write refusal and unchanged history. It is acceptable to intercept only Uvicorn's final run boundary to inspect the real constructed app. Keep env fallback and flag precedence coverage.

## Per-item review

| Requirement | Evidence / result |
|---|---|
| §5 before checkpoint, before bind | Store verification completes in `store.py:255-283`; factory calls gate at `app.py:744`; CLI constructs app before Uvicorn at `cli.py:356`. Pass by inspection. |
| Signature against accepted staged set before comparison | `app.py:697-707`, keys from `:743`; key loading unchanged. Pass. |
| Below/equal/above comparisons | `checkpoint.py:87-103`, R2 prefix read at `store.py:1704`. Correct implementation; negative-test gap above. |
| Refusal stays non-ready, writes disabled, latch survives refresh | `store.py:1673-1715`, `app.py:213`, committed refusal/latch tests pass. No truncation. |
| No checkpoint and compared boundary/outcome logs | Event contents tested, but real CLI silently drops INFO events. FAIL, finding 1. |
| Frozen health success schema / offline verification retained | Health success unchanged; verify-backup untouched. Pass. |
| Shared checkpoint vectors | 7/7 executed through real app factory + HTTP at `test_protocol_conformance.py:949`; fixture predicates asserted. Pass, with narrowing blind spot identified. |
| Older database refusal through CLI | Restore test proves factory refusal and byte preservation, not CLI entry. FAIL, finding 3. |
| Revision-1 gate repair | Applying rev1 patch to base and comparing rev2 shows only `tests/test_registry.py` differs. No new production hook. Existing stop API used before clock setup at `:2047-2055`; fixed epoch and exact-bound assertions retained at `:2058-2088`. Pass. Producer's floating-point diagnosis is plausible; exact CI historical cause not independently proven. |
| Independent pytest / strict mypy | 162 passed / 14 source files clean, both exit 0. |
| Narrowing attacks | 2/3 killed; root-only omission survives full suite. See below. |
| CI / existing recovery and R2 suites | Full local suite green; CI pin correct at `.github/workflows/ci.yml:31`. Attached hosted run reports all six test lanes, mypy, build and Docker green. Accepted as existing hosted evidence, not rerun. |
| Docs / scope | README startup gate, SECURITY restore paragraph, compose commented mount/env, CHANGELOG R3/P2 present. No key rotation, P4/R4–R8 or spec modifications. |
| Hygiene / lint | `git diff --check 131952d ee087392` exit 0. No dedicated lint gate configured. Candidate unchanged. |

## Validation commands and measured bounds

Shell `/bin/bash`, `set -o pipefail`; Python 3.14.6. Working directory `/tmp/1ny7yl-review`:

```
CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1 /tmp/1ny7yl-review-venv/bin/python -m pytest -q
# 162 passed, 2 warnings in 27.12s; exit 0
/tmp/1ny7yl-review-venv/bin/python -m mypy
# Success: no issues found in 14 source files; exit 0
```

Initial validation was inadvertently launched before pip finished and produced missing-dependency collection/type errors (pytest 2, mypy 1). Those runs were discarded as environment setup failures and rerun after successful installation; final transcripts below. No source changes between attempts. Only local Python 3.14/macOS was independently run; other platforms/build/Docker rely on attached hosted evidence. No hosted gate rerun.

Mutants each use independent copies and the committed tests:

```
CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1 /tmp/1ny7yl-review-venv/bin/python -m pytest -q tests/test_registry.py tests/test_protocol_conformance.py -k checkpoint
```

- `prefix-head`: omit only prefix head equality, retain root check. Killed: comparator test fails, 1 failed/6 passed, exit 1.
- `signature-equal`: narrow signature refusal to `live.version < checkpoint.version and not any(verify_signed(...))`. Killed through startup conformance HTTP: 1 failed/6 passed, exit 1.
- `prefix-root`: omit only prefix root equality, retain head check. Survives: 7 passed, exit 0. Full suite rerun in this mutant copy: **162 passed, 2 warnings in 35.01s, exit 0**. Coverage measured 2/3 mutants killed; no claim of exhaustive gate coverage.

CLI log probe uses installed console entry point, real factory, real Uvicorn, empty initialized database and auditors, ephemeral loopback port. Each process was deliberately terminated after 3 seconds and reaped (exit -15); these are bounded observation runs, not successful-exit assertions. Both reached application startup and listener. No-checkpoint and valid-checkpoint startup records: 0/2 observed. The initial attempted `python -m csk_registry` invocation failed because the package has no `__main__`; corrected to the supported console entry point before drawing conclusions.

## Logbook

2026-09-17 review rev2: discovered production startup INFO suppression masked by caplog; demonstrated an above-prefix Merkle-root-only narrowing mutant survives all 162 tests; identified missing CLI restore integration coverage. Ordinary implementation/test rework required. Preserve correct comparison/latch code, repair observability and add discriminating entry-point tests, then republish for another reviewer cycle. No standalone logbook tool is installed/exposed in this session; this durable task-scoped logbook entry is also attached separately on the board.

## Transcript: Pytest

```text
........................................................................ [ 44%]
........................................................................ [ 88%]
..................                                                       [100%]
=============================== warnings summary ===============================
../../../tmp/1ny7yl-review-venv/lib/python3.14/site-packages/fastapi/testclient.py:1
  /tmp/1ny7yl-review-venv/lib/python3.14/site-packages/fastapi/testclient.py:1: StarletteDeprecationWarning: Using `httpx` with `starlette.testclient` is deprecated; install `httpx2` instead.
    from starlette.testclient import TestClient as TestClient  # noqa

../../../tmp/1ny7yl-review-venv/lib/python3.14/site-packages/starlette/testclient.py:53
  /tmp/1ny7yl-review-venv/lib/python3.14/site-packages/starlette/testclient.py:53: DeprecationWarning: The anyio.abc.BlockingPortal alias is deprecated, use anyio.from_thread.BlockingPortal instead.
    _PortalFactoryType = Callable[[], AbstractContextManager[anyio.abc.BlockingPortal]]

-- Docs: https://docs.pytest.org/en/stable/how-to/capture-warnings.html
162 passed, 2 warnings in 27.12s

```

## Transcript: Mypy

```text
Success: no issues found in 14 source files

```

## Transcript: CLI startup log probe

```text
absent ['/tmp/1ny7yl-review-venv/bin/curator-skill-registry', '--home', '/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/1ny7yl-live-6z2no0_7', 'serve', '--port', '0'] exit -15 
INFO:     Started server process [15825]
INFO:     Waiting for application startup.
INFO:     Application startup complete.
INFO:     Uvicorn running on http://127.0.0.1:59770 (Press CTRL+C to quit)
INFO:     Shutting down
INFO:     Waiting for application shutdown.
INFO:     Application shutdown complete.
INFO:     Finished server process [15825]

startup_checkpoint present: False
valid ['/tmp/1ny7yl-review-venv/bin/curator-skill-registry', '--home', '/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/1ny7yl-live-6z2no0_7', 'serve', '--port', '0', '--checkpoint', '/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/1ny7yl-live-6z2no0_7/checkpoint.json'] exit -15 
INFO:     Started server process [15917]
INFO:     Waiting for application startup.
INFO:     Application startup complete.
INFO:     Uvicorn running on http://127.0.0.1:59773 (Press CTRL+C to quit)
INFO:     Shutting down
INFO:     Waiting for application shutdown.
INFO:     Application shutdown complete.
INFO:     Finished server process [15917]

startup_checkpoint present: False

```

## Transcript: Prefix-root full suite

```text
........................................................................ [ 44%]
........................................................................ [ 88%]
..................                                                       [100%]
=============================== warnings summary ===============================
../../../tmp/1ny7yl-review-venv/lib/python3.14/site-packages/fastapi/testclient.py:1
  /tmp/1ny7yl-review-venv/lib/python3.14/site-packages/fastapi/testclient.py:1: StarletteDeprecationWarning: Using `httpx` with `starlette.testclient` is deprecated; install `httpx2` instead.
    from starlette.testclient import TestClient as TestClient  # noqa

../../../tmp/1ny7yl-review-venv/lib/python3.14/site-packages/starlette/testclient.py:53
  /tmp/1ny7yl-review-venv/lib/python3.14/site-packages/starlette/testclient.py:53: DeprecationWarning: The anyio.abc.BlockingPortal alias is deprecated, use anyio.from_thread.BlockingPortal instead.
    _PortalFactoryType = Callable[[], AbstractContextManager[anyio.abc.BlockingPortal]]

-- Docs: https://docs.pytest.org/en/stable/how-to/capture-warnings.html
162 passed, 2 warnings in 35.01s

```

## Transcript: Hosted validation (accepted existing evidence)

```text
$ sh scripts/remote-gate.sh
remote gate: attempt 1/3: pushing 770a6d4bb2d5f4329156bf78686fe6f666cf59b0 as gate/STORY-260910-35tbgb/260917-190851-93619-1
remote: 
remote: Create a pull request for 'gate/STORY-260910-35tbgb/260917-190851-93619-1' on GitHub by visiting:        
remote:      https://github.com/relux-works/curator-skill-registry/pull/new/gate/STORY-260910-35tbgb/260917-190851-93619-1        
remote: 
remote gate: run 35263136115 (https://github.com/relux-works/curator-skill-registry/actions/runs/35263136115)
remote gate: run 35263136115 finished: success
  Docker build: success  
  Tests / Python 3.11 on ubuntu-latest: success  
  Tests / Python 3.11 on macos-latest: success  
  Tests / Python 3.14 on windows-latest: success  
  Type check / mypy strict: success  
  Tests / Python 3.11 on windows-latest: success  
  Tests / Python 3.14 on macos-latest: success  
  Tests / Python 3.14 on ubuntu-latest: success  
  Build distribution: success  
[exit 0]
coverage_unit=exact_command_shard required=1 green=1 failed=0 missing=0 test_case_coverage=unknown

```
