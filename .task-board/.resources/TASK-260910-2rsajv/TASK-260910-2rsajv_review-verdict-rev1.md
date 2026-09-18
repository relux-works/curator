# TASK-260910-2rsajv — review verdict revision 1

Verdict: accepted. No implementation corrections required.

Candidate tree `88118f3ebca6cee1386299fb5e87d10b89189921`, base/HEAD
`d7f424c3f8a850b2f1b0c14746a4797a6385e0a8`, branch
`task-board/story/STORY-260910-2xe3n2`. Published patch SHA-256
`58cf2438657ad00f07be44dd4d7e7df5e706bed50e038fe6943299e37a9c19f0`
matches `git diff HEAD` byte-for-byte before and after review.

| Requirement | Evidence | Result |
|---|---|---|
| 26 h retention and minimum/slack comment | app.py:65–70 quotes profile §4 and explains two-hour slack | Pass |
| Production wiring | app.py:454–461 passes constant to Store.append_idempotent; store.py:949–980 expires ledger rows and records now + ttl | Pass |
| Normative contract | Read profile §4:111–113 at CI pin 47c3c8c and brief-requested dced9b8; minimum wording identical | Pass |
| Retry inside slack and after expiry | tests/test_registry.py:1573,1596 drive POST /v1/records and assert replay/append and log cardinality | Pass |
| Boundary probes | Disposable test variants: 25h59m replay, 26h01m new append; selected suite 5 passed | Pass |
| Documentation | README.md:194 says 26-hour; SECURITY has no retention duration; CHANGELOG.md:36–43 has R8 | Pass |
| Architecture/scope | Only constant/comment, README, release note, two endpoint tests; unchanged store floor appropriately stays 24 h | Pass |
| Story baseline | HEAD is checkpointed R5 d7f424c; four-file delta exactly matches published patch | Pass |
| Independent tests and strict typing | Full logs below | Pass |
| Negative evidence | 2/2 selected narrowing mutants caught by behavioral assertion, not constant assertion | Pass |
| Hygiene | Initial/final git status only four expected tracked modifications; no untracked paths; git diff --check exit 0 | Pass |

The brief specifically asks to cite dced9b8; that historical citation is retained.
The authoritative CI/conformance pin remains 47c3c8c, and the cited minimum is
unchanged there. Historical 24-hour migration release note and store minimum
are not claims that current service retention is 24 hours.

## Independent validation

Shell: /bin/bash, set -o pipefail. Python: /tmp/csk-venv/bin/python,
Python 3.14.6. CWD: /tmp/TASK-260910-2rsajv-review/candidate, extracted with
`git archive 88118f3ebca6cee1386299fb5e87d10b89189921`.
Existing external venv reused without modifying it; PYTHONPATH explicitly points
at the disposable candidate src, and imported app.__file__ verified that path.
A separate fresh venv installation was started but was not used for these checks.
No tests, caches, files, or edits were made in the managed worktree.

Commands:
```bash
set -o pipefail
export CURATOR_CONFORMANCE_ROOT=/tmp/spec-47c3c8c/conformance/v1
export PYTHONPATH="$PWD/src"
/tmp/csk-venv/bin/python -m pytest -q
/tmp/csk-venv/bin/python -m mypy
```
Conformance checkout HEAD verified as 47c3c8cbd5da5e3fd6d99b0327d382c6a36494fe.
Strict mode comes from pyproject.toml. No separate lint tool configured.

## Mutants and probes

In a second disposable archive, changed only app.py:461 TTL argument to
`IDEMPOTENCY_TTL_SECONDS - 7200`, then separately `- 3600`. Constant stayed
26 hours. Both runs of committed candidate tests (`pytest -q tests/test_registry.py
-k idempotency_retry`) failed with replay HTTP 201 versus expected 200; expiry
control passed. Measured coverage: 2/2 attempted narrowing mutants killed.
This measures these two call-site mutations, not exhaustive store mutation coverage.
Restored source from saved bytes. Then changed only test offsets to 25h59m and
26h01m, ran `pytest -q tests/test_registry.py -k idempotency`, and restored tests.
All attack modifications stayed outside managed worktree.

## Reused hosted evidence

Read TASK-260910-2rsajv_change-request_rev1-validation.log: runtime gate exit 0,
GitHub run 35344199613, all six OS/Python combinations, mypy, build and Docker
successful. Accepted that attached evidence for hosted build/matrix checks;
did not rerun hosted gate. Independently reran local full pytest and mypy.
Producer results were also read; acceptance is based on independent checks above.

## Logbook

No regression or external blocker found. The behavioral mutants establish that
the new tests detect narrowed retention at the actual store call site even when
the public constant remains correct. Goal query reports this run is not goal-bound.
The candidate is accepted for producer integration, not marked done by reviewer.

## Full pytest transcript

```text
........................................................................ [ 40%]
........................................................................ [ 81%]
.................................                                        [100%]
=============================== warnings summary ===============================
../../../../tmp/csk-venv/lib/python3.14/site-packages/fastapi/testclient.py:1
  /tmp/csk-venv/lib/python3.14/site-packages/fastapi/testclient.py:1: StarletteDeprecationWarning: Using `httpx` with `starlette.testclient` is deprecated; install `httpx2` instead.
    from starlette.testclient import TestClient as TestClient  # noqa

../../../../tmp/csk-venv/lib/python3.14/site-packages/starlette/testclient.py:53
  /tmp/csk-venv/lib/python3.14/site-packages/starlette/testclient.py:53: DeprecationWarning: The anyio.abc.BlockingPortal alias is deprecated, use anyio.from_thread.BlockingPortal instead.
    _PortalFactoryType = Callable[[], AbstractContextManager[anyio.abc.BlockingPortal]]

-- Docs: https://docs.pytest.org/en/stable/how-to/capture-warnings.html
177 passed, 2 warnings in 136.44s (0:02:16)
EXIT=0
```

## Strict mypy transcript

```text
Success: no issues found in 15 source files
EXIT=0
```

## Mutants and boundary probes transcript

```text
narrow24h: ['/tmp/csk-venv/bin/python', '-m', 'pytest', '-q', 'tests/test_registry.py', '-k', 'idempotency_retry']
F.                                                                       [100%]
=================================== FAILURES ===================================
__________ test_idempotency_retry_within_slack_window_is_deduplicated __________

tmp_path = PosixPath('/private/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/pytest-of-administrator/pytest-165/test_idempotency_retry_within_0')
monkeypatch = <_pytest.monkeypatch.MonkeyPatch object at 0x10ceb5350>

    def test_idempotency_retry_within_slack_window_is_deduplicated(
        tmp_path: Path, monkeypatch: pytest.MonkeyPatch
    ) -> None:
        import csk_registry.app as app_module
    
        assert app_module.IDEMPOTENCY_TTL_SECONDS == 26 * 3600
        client, _, token = _client(tmp_path)
        auditor_key = client.app.state.auditor_key  # type: ignore[attr-defined]
        clock = _ManualAppClock(1_800_000_000.0)
        monkeypatch.setattr(app_module, "time", clock)
        headers = {"Authorization": f"Bearer {token}", "Idempotency-Key": "slack-key"}
        record = auditor_key.sign_record(_body("audited"))
        first = client.post("/v1/records", json=record, headers=headers)
        assert first.status_code == 201
        # A retry at 25 h — past the 24 h contract minimum but inside the 26 h
        # retention — replays the original response without a second append.
        clock.now += 25 * 3600
        replay = client.post("/v1/records", json=record, headers=headers)
>       assert replay.status_code == 200
E       assert 201 == 200
E        +  where 201 = <Response [201 Created]>.status_code

tests/test_registry.py:1591: AssertionError
=============================== warnings summary ===============================
../../../../tmp/csk-venv/lib/python3.14/site-packages/fastapi/testclient.py:1
  /tmp/csk-venv/lib/python3.14/site-packages/fastapi/testclient.py:1: StarletteDeprecationWarning: Using `httpx` with `starlette.testclient` is deprecated; install `httpx2` instead.
    from starlette.testclient import TestClient as TestClient  # noqa

../../../../tmp/csk-venv/lib/python3.14/site-packages/starlette/testclient.py:53
  /tmp/csk-venv/lib/python3.14/site-packages/starlette/testclient.py:53: DeprecationWarning: The anyio.abc.BlockingPortal alias is deprecated, use anyio.from_thread.BlockingPortal instead.
    _PortalFactoryType = Callable[[], AbstractContextManager[anyio.abc.BlockingPortal]]

-- Docs: https://docs.pytest.org/en/stable/how-to/capture-warnings.html
=========================== short test summary info ============================
FAILED tests/test_registry.py::test_idempotency_retry_within_slack_window_is_deduplicated
1 failed, 1 passed, 88 deselected, 2 warnings in 6.62s

EXIT=1
narrow25h: ['/tmp/csk-venv/bin/python', '-m', 'pytest', '-q', 'tests/test_registry.py', '-k', 'idempotency_retry']
F.                                                                       [100%]
=================================== FAILURES ===================================
__________ test_idempotency_retry_within_slack_window_is_deduplicated __________

tmp_path = PosixPath('/private/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/pytest-of-administrator/pytest-166/test_idempotency_retry_within_0')
monkeypatch = <_pytest.monkeypatch.MonkeyPatch object at 0x1069888a0>

    def test_idempotency_retry_within_slack_window_is_deduplicated(
        tmp_path: Path, monkeypatch: pytest.MonkeyPatch
    ) -> None:
        import csk_registry.app as app_module
    
        assert app_module.IDEMPOTENCY_TTL_SECONDS == 26 * 3600
        client, _, token = _client(tmp_path)
        auditor_key = client.app.state.auditor_key  # type: ignore[attr-defined]
        clock = _ManualAppClock(1_800_000_000.0)
        monkeypatch.setattr(app_module, "time", clock)
        headers = {"Authorization": f"Bearer {token}", "Idempotency-Key": "slack-key"}
        record = auditor_key.sign_record(_body("audited"))
        first = client.post("/v1/records", json=record, headers=headers)
        assert first.status_code == 201
        # A retry at 25 h — past the 24 h contract minimum but inside the 26 h
        # retention — replays the original response without a second append.
        clock.now += 25 * 3600
        replay = client.post("/v1/records", json=record, headers=headers)
>       assert replay.status_code == 200
E       assert 201 == 200
E        +  where 201 = <Response [201 Created]>.status_code

tests/test_registry.py:1591: AssertionError
=============================== warnings summary ===============================
../../../../tmp/csk-venv/lib/python3.14/site-packages/fastapi/testclient.py:1
  /tmp/csk-venv/lib/python3.14/site-packages/fastapi/testclient.py:1: StarletteDeprecationWarning: Using `httpx` with `starlette.testclient` is deprecated; install `httpx2` instead.
    from starlette.testclient import TestClient as TestClient  # noqa

../../../../tmp/csk-venv/lib/python3.14/site-packages/starlette/testclient.py:53
  /tmp/csk-venv/lib/python3.14/site-packages/starlette/testclient.py:53: DeprecationWarning: The anyio.abc.BlockingPortal alias is deprecated, use anyio.from_thread.BlockingPortal instead.
    _PortalFactoryType = Callable[[], AbstractContextManager[anyio.abc.BlockingPortal]]

-- Docs: https://docs.pytest.org/en/stable/how-to/capture-warnings.html
=========================== short test summary info ============================
FAILED tests/test_registry.py::test_idempotency_retry_within_slack_window_is_deduplicated
1 failed, 1 passed, 88 deselected, 2 warnings in 8.17s

EXIT=1
Boundary probes 25h59m and 26h01m: ['/tmp/csk-venv/bin/python', '-m', 'pytest', '-q', 'tests/test_registry.py', '-k', 'idempotency']
.....                                                                    [100%]
=============================== warnings summary ===============================
../../../../tmp/csk-venv/lib/python3.14/site-packages/fastapi/testclient.py:1
  /tmp/csk-venv/lib/python3.14/site-packages/fastapi/testclient.py:1: StarletteDeprecationWarning: Using `httpx` with `starlette.testclient` is deprecated; install `httpx2` instead.
    from starlette.testclient import TestClient as TestClient  # noqa

../../../../tmp/csk-venv/lib/python3.14/site-packages/starlette/testclient.py:53
  /tmp/csk-venv/lib/python3.14/site-packages/starlette/testclient.py:53: DeprecationWarning: The anyio.abc.BlockingPortal alias is deprecated, use anyio.from_thread.BlockingPortal instead.
    _PortalFactoryType = Callable[[], AbstractContextManager[anyio.abc.BlockingPortal]]

-- Docs: https://docs.pytest.org/en/stable/how-to/capture-warnings.html
5 passed, 85 deselected, 2 warnings in 15.95s

EXIT=0
```

