# TASK-260910-2c7s0u review verdict — revision 1

Verdict: **changes_requested**, route **to-dev**. Documentation rework only; do not change service code or tests. Candidate tree `470a6884fcca2b44bee9cdbde0ac9b60f47fd28a`, base `bf5cac1200cfa39dff0a6b0449072ff5b22f124d`, reviewed 2026-09-18. No external blocker or human decision required. `task-board spawn goal "$TASK_BOARD_RUN_ID"` reports no active run goal.

## Required corrections

1. **F1, medium — attacker guarantees misdescribe the code (README.md:354–360).** An invalid-token request DOES reach token verification after its body is read (`app.py:402–408`). A valid token reaches the auditor limiter BEFORE signature validation (`app.py:411`, `:435–436`); a signature is not a prerequisite to reach that limiter. Replace the cannot paragraph with stage-specific statements: no valid token means no auditor-limiter access or record append; record append additionally requires valid schema/signature. The 15 s bound is scoped to `_read_request_body` (`app.py:611–623`), not total slot lifetime: acquisition precedes routing and release follows the route (`:148–175`). The size guard rejects after receiving the chunk that crosses the limit (`:616–619`); it does not guarantee at most 16 MiB on the wire. Describe accepted/buffered body bounds, not a hard inbound-bandwidth cap. Qualify 'changes no state' as no registry append/persistent registry mutation: network limiter accounting and audit happen on refused requests (`:131–163`). These are documentation inaccuracies, not requests to alter existing behavior.

   Independent real-entry reproduction, in the disposable candidate, using existing `_client` and spies wrapping the real `AuditorTokens.resolve` and `FixedWindowLimiter.allow`:
   ```text
   POST /v1/records, body {}, invalid Bearer token:
   401 invalid_token ['network:testclient', 'token verification']
   POST /v1/records, body {}, valid token, no valid signed record:
   400 invalid_record ['network:testclient', 'token verification', 'auditor:a1']
   exit 0
   ```

2. **F2, medium — nginx timeout model and worked example disagree (README.md:327,335,374–375).** `client_body_timeout 15s` is an idle gap between successive reads, not a 15 s total upload deadline. A trickling upload can outlive 15 s at nginx. `proxy_read_timeout` concerns upstream response reads, not incoming request bodies; the example sets 30 s while prose recommends body and read timeouts at or below 15 s. Correct the comment and prose, explain their distinct scopes, and make the example consistent. Explain that nginx request buffering defaults on: the complete body is received at the edge before forwarding, so slow-client occupancy is primarily at the proxy in this setup. Retain per-client connection and body-size caps without promising a total edge upload deadline these directives do not enforce. Official sources inspected:
   - https://nginx.org/en/docs/http/ngx_http_core_module.html#client_body_timeout
   - https://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_request_buffering
   - https://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_read_timeout

3. **F3, low — corrupted documentation sentence (README.md:344).** The actual file contains the literal placeholder `[REDACTED]` in 'the Bearer [REDACTED] the auditor limiter checked'. Restore ordinary prose, e.g. 'the bearer token is verified and the auditor limiter is checked'. This is not a secret and should not contain a redaction marker.

Update the producer results' 'zero mismatches' claims with these findings and the corrected code mapping. Re-run the required gates after rework and publish another revision for review. No code fixes requested.

## Acceptance and architecture review

| Item | Result / evidence |
|---|---|
| AC1: single deployment section, limiter keys and trust settings | Present; README:276–338. Defaults and keys correct: app.py:63,71–74,120–154,411–418; limits.py:18. CLI defaults off/unset at cli.py:519–526, required pair at :374–376, Uvicorn wiring :407–408. Example loopback proxy and trusted address agree. Worked timeout explanation needs F2. |
| AC2: body bounds, semantics, mitigations, checklist | Numeric constants correct; app.py:400–408,611–623. Checklist present README:377–391. Cannot accept prose guarantees until F1/F2/F3 corrected. |
| AC3: SECURITY pointer and Unreleased R4 | Pass, SECURITY:85–96 and CHANGELOG:36–48. Local README anchor resolves. |
| AC4: every documented claim matches code | Fail, F1/F2; producer mismatch report must be corrected. No code patched by reviewer. |
| AC5: unchanged suite, results outside tree | Pass. Leaf delta from 7a8627b to candidate is exactly three docs, 141 insertions. 219 tests independently green. Producer results attached outside tree. |
| Story replay | Faithful content, with patch-ID context exceptions explicitly recorded below. |
| Architecture/scope | Existing one-place deployment documentation is appropriate. No new implementation, spec, deploy or packaging change in this leaf. Whole-story CR has nine files because it also contains R5/R8/R7. |
| Hygiene | Pass. Every tracked worktree file byte-matches candidate. Status before/after lists only CHANGELOG.md, README.md, SECURITY.md. No untracked files. Candidate filename scan found no results/logbook/coverage/build artifacts. Reviewer tests/caches/mutants only under /tmp. |
| Negative evidence | 3/3 selected narrowing mutants killed by committed tests. This is targeted coverage, not an exhaustive mutation score. M1/M2 exercise helpers; M3 drives real POST endpoint, production app.py:421. |
| Hosted validation | Existing attached log accepted as evidence only: run 35351139582, six OS/Python test jobs, mypy, distribution and Docker success; exit 0. Reviewer did not rerun hosted gate. |

## Replay evidence

Ancestry is bf5cac1 → 0b19c4a → b9847a1 → 7a8627b. Every checkpoint's added/deleted lines match the original on every changed path. Raw per-file stable patch IDs are identical except CHANGELOG and two context-only exceptions: R5 tests/test_registry.py includes newly landed `Callable` in surrounding import context; R7 README follows the newly landed P4 high-water section instead of the old startup heading. Their changed lines are identical; these are not lost edits. All CHANGELOG checkpoint additions are identical and preserved alongside P4/R6. This qualifies the review brief's expectation of identical non-CHANGELOG patch IDs rather than concealing a mismatch.

Published patch SHA-256 independently checked:
`1179c10dc6858972af06c262d0fe5ddcfaf6c0d62eac8df5540f0ff60a79bdb4`.


```text
d7f424c -> 0b19c4a CHANGELOG.md: 03ef4b39c2b4ba9f559b635a431055ce06c3804b 92eda67364607f47b3801d3019a0abbad7a087b1 equal=False
d7f424c -> 0b19c4a README.md: 74785ea5d3eda57b1339a2d683ba334ceda67f05 74785ea5d3eda57b1339a2d683ba334ceda67f05 equal=True
d7f424c -> 0b19c4a src/csk_registry/app.py: 370aadbf244bdb473c171b388437fb543bc6e779 370aadbf244bdb473c171b388437fb543bc6e779 equal=True
d7f424c -> 0b19c4a src/csk_registry/errors.py: 3025c0847982d45086db116ac14cb55de84953f4 3025c0847982d45086db116ac14cb55de84953f4 equal=True
d7f424c -> 0b19c4a src/csk_registry/protocol.py: a542cf0da6fd1bb7d132a57ee08e7b2fa8ef2d26 a542cf0da6fd1bb7d132a57ee08e7b2fa8ef2d26 equal=True
d7f424c -> 0b19c4a src/csk_registry/signing.py: 2d2d6b89228ec1366647a60fc1d7b11a3b810efa 2d2d6b89228ec1366647a60fc1d7b11a3b810efa equal=True
d7f424c -> 0b19c4a tests/test_registry.py: e40a41a12c0fa3a4152a97c000665efcd6ccda5e 3dd15b3b17896858669e38ceee1f2df084da86b6 equal=False
bd27d32 -> b9847a1 CHANGELOG.md: 25b2b4d71a71b9cc9e260e5bf5df169128a36951 d1ab250c07b8289b3e0815c961dfecde24f83312 equal=False
bd27d32 -> b9847a1 README.md: c90cdf8b455823a2f31b51a206cbc09eb0a8cebf c90cdf8b455823a2f31b51a206cbc09eb0a8cebf equal=True
bd27d32 -> b9847a1 src/csk_registry/app.py: fefb21bb74bca3262ff58de4b3178c3e4f3a62bb fefb21bb74bca3262ff58de4b3178c3e4f3a62bb equal=True
bd27d32 -> b9847a1 tests/test_registry.py: 297551b9d5fd364ec529f683e8778d219edc9f9f 297551b9d5fd364ec529f683e8778d219edc9f9f equal=True
5f3b028 -> 7a8627b CHANGELOG.md: 2183105b5513fd162f5f9de418491844410678c6 b26ab2dfe008c31a4146556120117909705300ff equal=False
5f3b028 -> 7a8627b README.md: d15a7a053a709e12935c69e2247ee1c58ed53198 155bc7ef1f3d18ab1a817f9e099f7dbb2d64799e equal=False
5f3b028 -> 7a8627b SECURITY.md: 8826396c68adb71e7179d4c5f770a7f295b0b035 8826396c68adb71e7179d4c5f770a7f295b0b035 equal=True
5f3b028 -> 7a8627b src/csk_registry/cli.py: 4fca1757f82e33934afc66b6b6d85ee7b8542c78 4fca1757f82e33934afc66b6b6d85ee7b8542c78 equal=True
5f3b028 -> 7a8627b tests/test_registry.py: 3c7eaa20ff6c68b82aa9b68a0d3279c88cadbb55 3c7eaa20ff6c68b82aa9b68a0d3279c88cadbb55 equal=True
candidate tracked byte mismatches: []
d7f424c changelog additions identical=True
bd27d32 changelog additions identical=True
5f3b028 changelog additions identical=True
d7f424c -> 0b19c4a: all added/deleted lines identical
bd27d32 -> b9847a1: all added/deleted lines identical
5f3b028 -> 7a8627b: all added/deleted lines identical
SECURITY.md local deployment anchor resolves
```

## Independent validation transcripts

Frozen copy made with `git archive 470a6884fcca2b44bee9cdbde0ac9b60f47fd28a`, extracted at `/tmp/review-2c7s0u/candidate`. Existing external dependency venv `/tmp/csk-venv`, Python 3.14.6; pytest config puts this copy's `src` on sys.path. Shell zsh with `set -o pipefail`. Pinned conformance worktree independently verified at `47c3c8cbd5da5e3fd6d99b0327d382c6a36494fe`.

```text
$ CURATOR_CONFORMANCE_ROOT=/tmp/spec-47c3c8c/conformance/v1 /tmp/csk-venv/bin/python -m pytest -q 2>&1 | tee /tmp/review-2c7s0u/pytest.log
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
219 passed, 2 warnings in 116.88s (0:01:56)
exit 0
$ /tmp/csk-venv/bin/python -m mypy --cache-dir /tmp/review-2c7s0u/mypy-cache
Success: no issues found in 15 source files
exit 0
$ git diff --check
exit 0
```

## Narrowing attacks (disposable second copy)

M1 changes protocol depth rejection from >100 to >=100. M2 changes canonical depth rejection from >100 to >=100. Each is caught at the documented accepted boundary by a committed test. M3 narrows invalid_json mapping to idempotent requests, caught through the real non-idempotent POST path. All original bytes restored after each run; candidate never mutated. No claim of exhaustive gate coverage.

### M1
```text
if depth > MAX_JSON_DEPTH: -> if depth >= MAX_JSON_DEPTH:
$ python -m pytest -q tests/test_registry.py::test_load_json_rejects_over_deep_nesting
F                                                                        [100%]
=================================== FAILURES ===================================
___________________ test_load_json_rejects_over_deep_nesting ___________________

    def test_load_json_rejects_over_deep_nesting():
        bound = signing.MAX_JSON_DEPTH
>       assert isinstance(load_json("[" * bound + "]" * bound), list)
                          ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

tests/test_registry.py:303: 
_ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ 
src/csk_registry/protocol.py:123: in load_json
    _check_json_depth(raw)
src/csk_registry/protocol.py:87: in _check_json_depth
    bump(True)
_ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ 

opening = True

    def bump(opening: bool) -> None:
        nonlocal depth
        if opening:
            depth += 1
            if depth >= MAX_JSON_DEPTH:
>               raise JSONDepthError(
                    f"JSON nesting exceeds maximum depth of {MAX_JSON_DEPTH}"
                )
E               csk_registry.protocol.JSONDepthError: JSON nesting exceeds maximum depth of 100

src/csk_registry/protocol.py:54: JSONDepthError
=============================== warnings summary ===============================
../../../../tmp/csk-venv/lib/python3.14/site-packages/fastapi/testclient.py:1
  /tmp/csk-venv/lib/python3.14/site-packages/fastapi/testclient.py:1: StarletteDeprecationWarning: Using `httpx` with `starlette.testclient` is deprecated; install `httpx2` instead.
    from starlette.testclient import TestClient as TestClient  # noqa

../../../../tmp/csk-venv/lib/python3.14/site-packages/starlette/testclient.py:53
  /tmp/csk-venv/lib/python3.14/site-packages/starlette/testclient.py:53: DeprecationWarning: The anyio.abc.BlockingPortal alias is deprecated, use anyio.from_thread.BlockingPortal instead.
    _PortalFactoryType = Callable[[], AbstractContextManager[anyio.abc.BlockingPortal]]

-- Docs: https://docs.pytest.org/en/stable/how-to/capture-warnings.html
=========================== short test summary info ============================
FAILED tests/test_registry.py::test_load_json_rejects_over_deep_nesting - csk...
1 failed, 2 warnings in 7.04s

exit=1
restored byte-identically

```

### M2
```text
if depth + 1 > MAX_JSON_DEPTH: -> if depth + 1 >= MAX_JSON_DEPTH:
$ python -m pytest -q tests/test_registry.py::test_canonicalization_rejects_over_deep_nesting
F                                                                        [100%]
=================================== FAILURES ===================================
_______________ test_canonicalization_rejects_over_deep_nesting ________________

    def test_canonicalization_rejects_over_deep_nesting():
        bound = signing.MAX_JSON_DEPTH
>       signing.canonical_document_bytes(_nested_list(bound))

tests/test_registry.py:331: 
_ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ 
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
_ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ 

value = [], depth = 99

    def _validate_ccj(value: Any, depth: int = 0) -> None:
        if value is None or isinstance(value, bool):
            return
        if isinstance(value, int):
            if not -MAX_SAFE_INTEGER <= value <= MAX_SAFE_INTEGER:
                raise CanonicalError(f"CCJ-1 integer outside safe range: {value}")
            return
        if isinstance(value, float):
            raise CanonicalError("CCJ-1 numbers must be integers")
        if isinstance(value, str):
            if any(0xD800 <= ord(character) <= 0xDFFF for character in value):
                raise CanonicalError("CCJ-1 strings must not contain lone surrogates")
            return
        if isinstance(value, list):
            if depth + 1 >= MAX_JSON_DEPTH:
>               raise CanonicalDepthError(
                    f"JSON nesting exceeds maximum depth of {MAX_JSON_DEPTH}"
                )
E               csk_registry.signing.CanonicalDepthError: JSON nesting exceeds maximum depth of 100

src/csk_registry/signing.py:87: CanonicalDepthError
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
1 failed, 2 warnings in 3.78s

exit=1
restored byte-identically

```

### M3
```text
Narrow invalid_json mapping to idempotent submit only (app.py:421).
F                                                                        [100%]
=================================== FAILURES ===================================
___________ test_submit_rejects_deeply_nested_json_with_invalid_json ___________

tmp_path = PosixPath('/private/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/pytest-of-administrator/pytest-185/test_submit_rejects_deeply_nes0')

    def test_submit_rejects_deeply_nested_json_with_invalid_json(tmp_path: Path):
        client, _, token = _client(tmp_path)
        bound = signing.MAX_JSON_DEPTH
        headers = {"Authorization": f"Bearer {token}", "Content-Type": "application/json"}
        # Just over the explicit bound (far below the interpreter limit).
        over = "[" * (bound + 1) + "]" * (bound + 1)
        resp = client.post("/v1/records", content=over, headers=headers)
        assert resp.status_code == 400
>       assert resp.json()["error"]["code"] == "invalid_json"
E       AssertionError: assert 'invalid_record' == 'invalid_json'
E         
E         - invalid_json
E         + invalid_record

tests/test_registry.py:380: AssertionError
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
1 failed, 2 warnings in 7.19s

exit=1; restored byte-identically

```
