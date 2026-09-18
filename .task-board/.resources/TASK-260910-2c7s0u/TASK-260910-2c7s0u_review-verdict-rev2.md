# TASK-260910-2c7s0u review verdict — revision 2

Verdict: **accepted**. Route via accept_cr(revision=2) to integrating, not done. Reviewed 2026-09-18; no remaining correction or external blocker. Candidate tree `13cb0c91ffbed50e44e8a6ca85138e4d388d6db9`, base `bf5cac1200cfa39dff0a6b0449072ff5b22f124d`, checkpoint `7a8627bc314976b459c908a11e6c3bedd0d0e225`. Run goal queried before verdict: no active goal (not goal-bound).

## Findings and acceptance criteria

| Item | Verdict and independently inspected evidence |
|---|---|
| AC1: rate limits, trust, worked proxy | Pass. README:276–338 gives network host key, 600/min, auditor ID key, 120/min, global 128 slots and 0.1 s acquire. app.py:71–74,120–154,411–418; limits.py:18,27–39 support defaults, keys, fixed 60 s windows and Retry-After. cli.py:519–526 defaults proxy flag false and trusted list unset; :374–376 requires trusted list; :407–408 explicitly wires proxy_headers and forwarded_allow_ips. There is no service env equivalent. Untrusted headers ignored by installed Uvicorn middleware; only trusted proxy may rewrite client identity. Loopback nginx upstream and trusted 127.0.0.1 agree. |
| AC2/F1: body and authentication stages | Pass. README:342–374 now scopes 15 s to body read, acknowledges crossing chunk, and distinguishes invalid token from invalid signature. app.py:63,74,148–175,400–411,419–436,611–623: read then token then auditor limiter then schema/signature; timeout surrounds read only, semaphore surrounds routing; size is checked before chunks.append. No claim remains that token verification requires a valid token or that wire traffic/whole slot lifetime has those caps. Refusal wording qualifies registry persistence while acknowledging limiter accounting/audit. Overloaded occurs before routing with Retry-After 1. |
| AC2/F2: nginx and operator checklist | Pass. README:326–335,378–408 correctly separates body idle timeout from upstream response timeout and states default request buffering; retains body-size and per-source connection caps. Official nginx references below independently fetched. Checklist is actionable. Example is an http-context nginx fragment; no live nginx deployment test was performed. |
| F3: redaction placeholder | Pass. README:342–345 uses ordinary bearer-token verification prose; literal placeholder absent. |
| AC3 | Pass. SECURITY:85–96 links README#production-transport-and-limits; anchor exists. CHANGELOG:36–48 contains Unreleased R4 entry and spec pin. |
| AC4/results accuracy | Pass. Code references above verified; producer results Revision 2 explicitly supersedes inaccurate revision 1 claims. No code/doc mismatch requiring rework found. |
| AC5: unchanged tests and independent gates | Pass. Leaf delta checkpoint→candidate contains only CHANGELOG, README, SECURITY (159 insertions). Tests unchanged. Producer results include unchanged-green transcript and live outside tree. Independently reran 219 tests and mypy; transcripts below. |
| Revision scope | Pass. rev1 tree 470a6884→rev2 changes README only: 41 additions, 23 deletions in three relevant hunks; CHANGELOG/SECURITY and all other bytes identical. Inspected both published patches and their hashes, with rev2 hash matching assignment. |
| Story replay | Pass with explicit patch-ID context exceptions below. Exact ancestry bf5cac1→0b19c4a→b9847a1→7a8627b; all changed lines of all three checkpoints identical to originals; P4/R6 entries preserved and R7/R8/R5 entries retained without rewriting. Whole story delta has nine paths; this documentation leaf has three. |
| Architecture and scope | Pass. One existing README deployment section, SECURITY pointer, no duplicate deployment doc, implementation/spec/deploy changes absent from leaf. |
| Hygiene | Pass. All 30 tracked files byte-match frozen candidate. Before/after git status contains only three modified docs; no untracked files, results/logbook/build artifacts. Reviewer wrote nothing in candidate worktree; copies, caches and evidence under /tmp only. git diff --check and git diff 13cb0c91 --exit-code both exit 0. |
| Negative evidence | 2/2 selected narrowing mutants killed by committed tests; targeted coverage, not exhaustive mutation coverage. M1 narrows depth bound at helper boundary; M2 narrows invalid_json response mapping to idempotent submissions and drives real POST /v1/records, app.py:421–422. |

Nginx authorities inspected independently: [client_body_timeout](https://nginx.org/en/docs/http/ngx_http_core_module.html#client_body_timeout) limits the gap between body reads; [proxy_read_timeout](https://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_read_timeout) limits gaps reading upstream responses; [proxy_request_buffering](https://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_request_buffering) defaults on and reads the complete client body before forwarding. The prose and snippet now agree with these scopes.

Replay anomaly is unchanged from round 1: raw per-file patch IDs differ beyond CHANGELOG in R5 tests/test_registry.py (new Callable import in surrounding context) and R7 README (P4 heading replaces old following context). Inspected those diff contexts independently: actual additions/deletions identical. No edits lost or silently accepted on patch-ID assumption. Recorded in task-scoped logbook.

## Independent validation

Created immutable-input copy with `git archive 13cb0c91ffbed50e44e8a6ca85138e4d388d6db9 | tar -x -C /tmp/review-2c7s0u-r2/candidate`. Used existing external dependency venv `/tmp/csk-venv`, Python 3.14.6, zsh, set -o pipefail for pytest. Pytest configuration selects the copy's src; no candidate-worktree cache writes. Verified conformance root commit `47c3c8cbd5da5e3fd6d99b0327d382c6a36494fe` at /tmp/spec-47c3c8c.

Commands from copy:
- `CURATOR_CONFORMANCE_ROOT=/tmp/spec-47c3c8c/conformance/v1 /tmp/csk-venv/bin/python -m pytest -q 2>&1 | tee /tmp/review-2c7s0u-r2/pytest.log` — exit 0.
- `/tmp/csk-venv/bin/python -m mypy --cache-dir /tmp/review-2c7s0u-r2/mypy-cache` — success, no issues in 15 source files (output redirected then read).

Hosted evidence accepted from newly attached revision 2 validation log, not rerun: run 35353644646, six OS/Python test jobs plus mypy, distribution and Docker all success, gate exit 0. This is distinct from the independent local checks.

## Pytest transcript

```text
........................................................................ [ 32%]
........................................................................ [ 65%]
........................................................................ [ 98%]
...                                                                      [100%]
=============================== warnings summary ===============================
../../../../tmp/csk-venv/lib/python3.14/site-packages/fastapi/testclient.py:1
  /tmp/csk-venv/lib/python3.14/site-packages/fastapi/testclient.py:1: StarletteDeprecationWarning: Using `httpx` with `starlette.testclient` is deprecated; install `httpx2` instead.
    from starlette.testclient import TestClient as TestClient  # noqa

../../../../tmp/csk-venv/lib/python3.14/site-packages/starlette/testclient.py:53
  /tmp/csk-venv/lib/python3.14/site-packages/starlette/testclient.py:53: DeprecationWarning: The anyio.abc.BlockingPortal alias is deprecated, use anyio.from_thread.BlockingPortal instead.
    _PortalFactoryType = Callable[[], AbstractContextManager[anyio.abc.BlockingPortal]]

-- Docs: https://docs.pytest.org/en/stable/how-to/capture-warnings.html
219 passed, 2 warnings in 52.93s

```

## Mypy transcript

```text
Success: no issues found in 15 source files

```

## Replay / identity evidence

```text
d7f424c -> 0b19c4a CHANGELOG.md: 03ef4b39c2b4ba9f559b635a431055ce06c3804b 92eda67364607f47b3801d3019a0abbad7a087b1 equal=False; changed lines identical
d7f424c -> 0b19c4a README.md: 74785ea5d3eda57b1339a2d683ba334ceda67f05 74785ea5d3eda57b1339a2d683ba334ceda67f05 equal=True; changed lines identical
d7f424c -> 0b19c4a src/csk_registry/app.py: 370aadbf244bdb473c171b388437fb543bc6e779 370aadbf244bdb473c171b388437fb543bc6e779 equal=True; changed lines identical
d7f424c -> 0b19c4a src/csk_registry/errors.py: 3025c0847982d45086db116ac14cb55de84953f4 3025c0847982d45086db116ac14cb55de84953f4 equal=True; changed lines identical
d7f424c -> 0b19c4a src/csk_registry/protocol.py: a542cf0da6fd1bb7d132a57ee08e7b2fa8ef2d26 a542cf0da6fd1bb7d132a57ee08e7b2fa8ef2d26 equal=True; changed lines identical
d7f424c -> 0b19c4a src/csk_registry/signing.py: 2d2d6b89228ec1366647a60fc1d7b11a3b810efa 2d2d6b89228ec1366647a60fc1d7b11a3b810efa equal=True; changed lines identical
d7f424c -> 0b19c4a tests/test_registry.py: e40a41a12c0fa3a4152a97c000665efcd6ccda5e 3dd15b3b17896858669e38ceee1f2df084da86b6 equal=False; changed lines identical
bd27d32 -> b9847a1 CHANGELOG.md: 25b2b4d71a71b9cc9e260e5bf5df169128a36951 d1ab250c07b8289b3e0815c961dfecde24f83312 equal=False; changed lines identical
bd27d32 -> b9847a1 README.md: c90cdf8b455823a2f31b51a206cbc09eb0a8cebf c90cdf8b455823a2f31b51a206cbc09eb0a8cebf equal=True; changed lines identical
bd27d32 -> b9847a1 src/csk_registry/app.py: fefb21bb74bca3262ff58de4b3178c3e4f3a62bb fefb21bb74bca3262ff58de4b3178c3e4f3a62bb equal=True; changed lines identical
bd27d32 -> b9847a1 tests/test_registry.py: 297551b9d5fd364ec529f683e8778d219edc9f9f 297551b9d5fd364ec529f683e8778d219edc9f9f equal=True; changed lines identical
5f3b028 -> 7a8627b CHANGELOG.md: 2183105b5513fd162f5f9de418491844410678c6 b26ab2dfe008c31a4146556120117909705300ff equal=False; changed lines identical
5f3b028 -> 7a8627b README.md: d15a7a053a709e12935c69e2247ee1c58ed53198 155bc7ef1f3d18ab1a817f9e099f7dbb2d64799e equal=False; changed lines identical
5f3b028 -> 7a8627b SECURITY.md: 8826396c68adb71e7179d4c5f770a7f295b0b035 8826396c68adb71e7179d4c5f770a7f295b0b035 equal=True; changed lines identical
5f3b028 -> 7a8627b src/csk_registry/cli.py: 4fca1757f82e33934afc66b6b6d85ee7b8542c78 4fca1757f82e33934afc66b6b6d85ee7b8542c78 equal=True; changed lines identical
5f3b028 -> 7a8627b tests/test_registry.py: 3c7eaa20ff6c68b82aa9b68a0d3279c88cadbb55 3c7eaa20ff6c68b82aa9b68a0d3279c88cadbb55 equal=True; changed lines identical
All tracked worktree bytes equal frozen candidate; files=30
bf5cac1 fb86420
0b19c4a bf5cac1
b9847a1 0b19c4a
7a8627b b9847a1
rev1 patch sha256=1179c10dc6858972af06c262d0fe5ddcfaf6c0d62eac8df5540f0ff60a79bdb4
rev2 patch sha256=503f05eb6fa1d9f91c8743076725b8e531dca8d477207fc7908f3d70a02498ea
SECURITY local README deployment anchor resolves; no REDACTED placeholder
Artifact filename scan: []
 M CHANGELOG.md
 M README.md
 M SECURITY.md

```

## M1 narrowing mutant

```text
M1: src/csk_registry/signing.py
Replace 'if depth + 1 > MAX_JSON_DEPTH:' with 'if depth + 1 >= MAX_JSON_DEPTH:'
$ python -m pytest -q --tb=short tests/test_registry.py::test_canonicalization_rejects_over_deep_nesting
F                                                                        [100%]
=================================== FAILURES ===================================
_______________ test_canonicalization_rejects_over_deep_nesting ________________
tests/test_registry.py:331: in test_canonicalization_rejects_over_deep_nesting
    signing.canonical_document_bytes(_nested_list(bound))
src/csk_registry/signing.py:57: in canonical_document_bytes
    return _canonical_document(value)
           ^^^^^^^^^^^^^^^^^^^^^^^^^^
src/csk_registry/signing.py:62: in _canonical_document
    _validate_ccj(value)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:91: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:87: in _validate_ccj
    raise CanonicalDepthError(
E   csk_registry.signing.CanonicalDepthError: JSON nesting exceeds maximum depth of 100
=============================== warnings summary ===============================
../../../../tmp/csk-venv/lib/python3.14/site-packages/fastapi/testclient.py:1
  /tmp/csk-venv/lib/python3.14/site-packages/fastapi/testclient.py:1: StarletteDeprecationWarning: Using `httpx` with `starlette.testclient` is deprecated; install `httpx2` instead.
    from starlette.testclient import TestClient as TestClient  # noqa

../../../../tmp/csk-venv/lib/python3.14/site-packages/starlette/testclient.py:53
  /tmp/csk-venv/lib/python3.14/site-packages/starlette/testclient.py:53: DeprecationWarning: The anyio.abc.BlockingPortal alias is deprecated, use anyio.from_thread.BlockingPortal instead.
    _PortalFactoryType = Callable[[], AbstractContextManager[anyio.abc.BlockingPortal]]

-- Docs: https://docs.pytest.org/en/stable/how-to/capture-warnings.html
=========================== short test summary info ============================
FAILED tests/test_registry.py::test_canonicalization_rejects_over_deep_nesting
1 failed, 2 warnings in 4.09s

exit=1
Restored byte-identically.

```

## M2 narrowing mutant

```text
M2: src/csk_registry/app.py
Replace 'except JSONDepthError as exc:\n            raise APIError(400, "invalid_json", str(exc)) from exc' with 'except JSONDepthError as exc:\n            raise APIError(400, "invalid_json" if idempotency_key else "invalid_record", str(exc)) from exc'
$ python -m pytest -q --tb=short tests/test_registry.py::test_submit_rejects_deeply_nested_json_with_invalid_json
F                                                                        [100%]
=================================== FAILURES ===================================
___________ test_submit_rejects_deeply_nested_json_with_invalid_json ___________
tests/test_registry.py:380: in test_submit_rejects_deeply_nested_json_with_invalid_json
    assert resp.json()["error"]["code"] == "invalid_json"
E   AssertionError: assert 'invalid_record' == 'invalid_json'
E     
E     - invalid_json
E     + invalid_record
=============================== warnings summary ===============================
../../../../tmp/csk-venv/lib/python3.14/site-packages/fastapi/testclient.py:1
  /tmp/csk-venv/lib/python3.14/site-packages/fastapi/testclient.py:1: StarletteDeprecationWarning: Using `httpx` with `starlette.testclient` is deprecated; install `httpx2` instead.
    from starlette.testclient import TestClient as TestClient  # noqa

../../../../tmp/csk-venv/lib/python3.14/site-packages/starlette/testclient.py:53
  /tmp/csk-venv/lib/python3.14/site-packages/starlette/testclient.py:53: DeprecationWarning: The anyio.abc.BlockingPortal alias is deprecated, use anyio.from_thread.BlockingPortal instead.
    _PortalFactoryType = Callable[[], AbstractContextManager[anyio.abc.BlockingPortal]]

-- Docs: https://docs.pytest.org/en/stable/how-to/capture-warnings.html
=========================== short test summary info ============================
FAILED tests/test_registry.py::test_submit_rejects_deeply_nested_json_with_invalid_json
1 failed, 2 warnings in 1.53s

exit=1
Restored byte-identically.

```

## Attached hosted evidence

```text
$ sh scripts/remote-gate.sh
remote gate: attempt 1/3: pushing 465e82ac077e07742341babfbc46a3a4191c21ca as gate/STORY-260910-2xe3n2/260918-140103-67465-1
remote: 
remote: Create a pull request for 'gate/STORY-260910-2xe3n2/260918-140103-67465-1' on GitHub by visiting:        
remote:      https://github.com/relux-works/curator-skill-registry/pull/new/gate/STORY-260910-2xe3n2/260918-140103-67465-1        
remote: 
remote gate: run 35353644646 (https://github.com/relux-works/curator-skill-registry/actions/runs/35353644646)
remote gate: run 35353644646 finished: success
  Docker build: success  
  Type check / mypy strict: success  
  Tests / Python 3.11 on ubuntu-latest: success  
  Tests / Python 3.14 on ubuntu-latest: success  
  Tests / Python 3.11 on macos-latest: success  
  Tests / Python 3.14 on windows-latest: success  
  Tests / Python 3.11 on windows-latest: success  
  Tests / Python 3.14 on macos-latest: success  
  Build distribution: success  
[exit 0]
coverage_unit=exact_command_shard required=1 green=1 failed=0 missing=0 test_case_coverage=unknown

```
