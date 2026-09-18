# TASK-260910-28kmef — revision 1 review

Verdict: **changes_requested**, route to **to-dev**. No candidate edits.

Reviewed tree `27a8d330d918043ef1268f3b8647cd7183b8fe85`, base `c7ef32c75cc8dfda1abe38647af282a03175e8d3`. Candidate working-tree diff against the published tree is empty. Published patch SHA-256 verified: `4521fcf78f1d5f028436b8b6162f261a9173fcf570e13386bc5706a4fee5147f`. Six changed paths match the assignment. No files or caches added to the managed worktree; tests and attacks ran in disposable copies under `/tmp/r5-review`.

## Required correction

**R1 — Canonicalization entry points do not satisfy the explicitly required ProtocolError contract.** `src/csk_registry/signing.py:30` defines `CanonicalDepthError(CanonicalError)`, where `CanonicalError` inherits only `ValueError`. Both `canonical_bytes` (:36) and `canonical_document_bytes` (:46) propagate that exception from `_canonical_document` (:55–64). Direct over-depth input and an injected RecursionError both reproduce `isinstance(error, ProtocolError) == False`. A caller following the brief and catching ProtocolError therefore does not catch these failures. The protocol wrappers correctly translate this exception, so the HTTP probes pass, but they do not fix the public canonicalization entry points themselves.

Make depth/recursion errors from both canonicalization entry points catchable as ProtocolError, while preserving existing CanonicalError/ValueError compatibility and other mappings. A shared error-definition module can avoid the current protocol-to-signing import dependency if needed. Add direct tests catching ProtocolError for both actual over-depth inputs and injected RecursionError from CCJ validation / JSON serialization. Existing tests at `tests/test_registry.py:328–345` assert only CanonicalDepthError and currently encode the incomplete contract.

## Validation anomaly requiring evidence reconciliation

The mandated exact `dced9b8317e0e8af79edf2d0539b32bd22b6c85b` conformance run is **not green**: 171 passed, 1 failed, exit 1. `tests/test_protocol_conformance.py:954` expects `checkpoint_cases`, absent from that revision's registry-service vectors. Independently reproduced the same failure on base `c7ef32c`, so this is **pre-existing, not an R5 regression**. The actual CI pin at `.github/workflows/ci.yml:30` is `47c3c8cbd5da5e3fd6d99b0327d382c6a36494fe`, not dced9b8; the producer already reported that discrepancy. Keep the pin untouched under the current scope. Correct the campaign validation instructions/evidence to identify a compatible authorized revision before claiming the required exact-pin check passed; do not weaken or skip checkpoint tests. This does not require blocking ordinary R1 rework.

## Per-item assessment

| Item | Result and evidence |
|---|---|
| Deep HTTP requests | PASS: `app.py:415–430`; 1/1 body-reading routes (`POST /v1/records`, body read :397). Four reviewer probes: 20,000-deep array, 20,000-deep object, nested audit member, nested signature-envelope member; 4/4 return 400 invalid_json. TestClient used real app routing with server exceptions disabled; no socket-level transport claim. |
| load_json ProtocolError | PASS: `protocol.py:35,124–147`; explicit bound scan and recursion conversion. |
| Canonicalization ProtocolError | FAIL: R1 above. |
| Bound and docs | PASS: 100 container levels; iterative text pre-scan; recursive canonical traversal bounded at 100. `signing.py:17,80–96`, `protocol.py:41–103`, README:162–167, CHANGELOG:36–43 (R5 and dced9b8). |
| Existing mappings | Non-depth ProtocolError remains invalid_record; cursor stays invalid_cursor. Reviewed diff and existing tests. |
| Tests / negative evidence | New endpoint/unit tests present; two narrowing mutants caught (2/2 attempted). This is not an exhaustive mutation score. |
| Store / regression | Diagnostic full-suite run against the current specification checkout recorded below; required old-pin failure independently attributed to baseline. |
| Static/lint | mypy strict exit 0, 14 source files. `git diff --check` exit 0. No separate linter configured. |
| Build / hosted validation | Accepted attached revision-1 validation log as evidence: hosted run 35340282852 success, six OS/Python test cells, mypy, Docker and distribution build. Did not rerun hosted gate/build. Does not establish dced9b8 conformance. |
| Architecture / scope | Depth checks are appropriately located at parsing/canonicalization boundaries; R1 exception hierarchy needs correction. R4/R7/R8, CI pin, spec, deployment unchanged. |

## Independent environment and commands

Shell zsh; CPython 3.14.6; fresh venv `/tmp/r5-review/venv`, installed editable from disposable exact-tree archive with `pip install -e '/tmp/r5-review/candidate[dev]'` (exit 0). A detached spec worktree at `/tmp/r5-review/spec` is exactly dced9b8. Commands run from candidate with `set -o pipefail`:

```
CURATOR_CONFORMANCE_ROOT=/tmp/r5-review/spec/conformance/v1 /tmp/r5-review/venv/bin/python -m pytest -q
# exit 1
/tmp/r5-review/venv/bin/python -m mypy
# exit 0
```

Diagnostic rerun justified by the missing-vector failure:

```
# baseline archive c7ef32c, targeted reproduction
CURATOR_CONFORMANCE_ROOT=/tmp/r5-review/spec/conformance/v1 /tmp/r5-review/venv/bin/python -m pytest -q tests/test_protocol_conformance.py::test_shared_service_startup_checkpoint_cases
# exit 1, same KeyError
# candidate, current spec HEAD 5146c7b9ed4b0c07b840ab58f9908f197d667478
CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1 /tmp/r5-review/venv/bin/python -m pytest -q
```

Mutants independently altered only disposable copies of signing.py: (1) list depth check `>` → `>=`, caught by `test_canonicalization_rejects_over_deep_nesting`; (2) object depth check `>` → `>=`, caught by `test_load_json_rejects_over_deep_nesting`. Both pytest processes exited 1 because depth 100 was incorrectly refused. Candidate remained unchanged.

## Captured transcripts

Diagnostic current-spec full suite: **172 passed, 2 warnings, exit 0**. Warnings are third-party deprecations.

### pytest.log

```text
........................................................................ [ 41%]
.F...................................................................... [ 83%]
............................                                             [100%]
=================================== FAILURES ===================================
_________________ test_shared_service_startup_checkpoint_cases _________________

tmp_path = PosixPath('/private/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/pytest-of-administrator/pytest-118/test_shared_service_startup_ch0')
monkeypatch = <_pytest.monkeypatch.MonkeyPatch object at 0x11093fce0>
caplog = <_pytest.logging.LogCaptureFixture object at 0x10e9ab770>

    def test_shared_service_startup_checkpoint_cases(
        tmp_path: Path,
        monkeypatch: pytest.MonkeyPatch,
        caplog: pytest.LogCaptureFixture,
    ) -> None:
>       cases = _json("vectors/registry-service.json")["checkpoint_cases"]
                ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
E       KeyError: 'checkpoint_cases'

tests/test_protocol_conformance.py:954: KeyError
=============================== warnings summary ===============================
../../../../tmp/r5-review/venv/lib/python3.14/site-packages/fastapi/testclient.py:1
  /tmp/r5-review/venv/lib/python3.14/site-packages/fastapi/testclient.py:1: StarletteDeprecationWarning: Using `httpx` with `starlette.testclient` is deprecated; install `httpx2` instead.
    from starlette.testclient import TestClient as TestClient  # noqa

../../../../tmp/r5-review/venv/lib/python3.14/site-packages/starlette/testclient.py:53
  /tmp/r5-review/venv/lib/python3.14/site-packages/starlette/testclient.py:53: DeprecationWarning: The anyio.abc.BlockingPortal alias is deprecated, use anyio.from_thread.BlockingPortal instead.
    _PortalFactoryType = Callable[[], AbstractContextManager[anyio.abc.BlockingPortal]]

-- Docs: https://docs.pytest.org/en/stable/how-to/capture-warnings.html
=========================== short test summary info ============================
FAILED tests/test_protocol_conformance.py::test_shared_service_startup_checkpoint_cases
1 failed, 171 passed, 2 warnings in 40.00s

```

### mypy.log

```text
Success: no issues found in 14 source files

```

### baseline-pin.log

```text
F                                                                        [100%]
=================================== FAILURES ===================================
_________________ test_shared_service_startup_checkpoint_cases _________________

tmp_path = PosixPath('/private/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/pytest-of-administrator/pytest-119/test_shared_service_startup_ch0')
monkeypatch = <_pytest.monkeypatch.MonkeyPatch object at 0x1111dd0f0>
caplog = <_pytest.logging.LogCaptureFixture object at 0x111209160>

    def test_shared_service_startup_checkpoint_cases(
        tmp_path: Path,
        monkeypatch: pytest.MonkeyPatch,
        caplog: pytest.LogCaptureFixture,
    ) -> None:
>       cases = _json("vectors/registry-service.json")["checkpoint_cases"]
                ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
E       KeyError: 'checkpoint_cases'

tests/test_protocol_conformance.py:954: KeyError
=============================== warnings summary ===============================
../../../../tmp/r5-review/venv/lib/python3.14/site-packages/fastapi/testclient.py:1
  /tmp/r5-review/venv/lib/python3.14/site-packages/fastapi/testclient.py:1: StarletteDeprecationWarning: Using `httpx` with `starlette.testclient` is deprecated; install `httpx2` instead.
    from starlette.testclient import TestClient as TestClient  # noqa

../../../../tmp/r5-review/venv/lib/python3.14/site-packages/starlette/testclient.py:53
  /tmp/r5-review/venv/lib/python3.14/site-packages/starlette/testclient.py:53: DeprecationWarning: The anyio.abc.BlockingPortal alias is deprecated, use anyio.from_thread.BlockingPortal instead.
    _PortalFactoryType = Callable[[], AbstractContextManager[anyio.abc.BlockingPortal]]

-- Docs: https://docs.pytest.org/en/stable/how-to/capture-warnings.html
=========================== short test summary info ============================
FAILED tests/test_protocol_conformance.py::test_shared_service_startup_checkpoint_cases
1 failed, 2 warnings in 1.66s

```

### pytest-current-spec.log

```text
........................................................................ [ 41%]
........................................................................ [ 83%]
............................                                             [100%]
=============================== warnings summary ===============================
../../../../tmp/r5-review/venv/lib/python3.14/site-packages/fastapi/testclient.py:1
  /tmp/r5-review/venv/lib/python3.14/site-packages/fastapi/testclient.py:1: StarletteDeprecationWarning: Using `httpx` with `starlette.testclient` is deprecated; install `httpx2` instead.
    from starlette.testclient import TestClient as TestClient  # noqa

../../../../tmp/r5-review/venv/lib/python3.14/site-packages/starlette/testclient.py:53
  /tmp/r5-review/venv/lib/python3.14/site-packages/starlette/testclient.py:53: DeprecationWarning: The anyio.abc.BlockingPortal alias is deprecated, use anyio.from_thread.BlockingPortal instead.
    _PortalFactoryType = Callable[[], AbstractContextManager[anyio.abc.BlockingPortal]]

-- Docs: https://docs.pytest.org/en/stable/how-to/capture-warnings.html
172 passed, 2 warnings in 32.95s

```

### probes.log

```text
/tmp/r5-review/venv/lib/python3.14/site-packages/fastapi/testclient.py:1: StarletteDeprecationWarning: Using `httpx` with `starlette.testclient` is deprecated; install `httpx2` instead.
  from starlette.testclient import TestClient as TestClient  # noqa
array 400 {'error': {'code': 'invalid_json', 'message': 'JSON nesting exceeds maximum depth of 100'}}
object 400 {'error': {'code': 'invalid_json', 'message': 'JSON nesting exceeds maximum depth of 100'}}
nested-audit 400 {'error': {'code': 'invalid_json', 'message': 'JSON nesting exceeds maximum depth of 100'}}
nested-sig 400 {'error': {'code': 'invalid_json', 'message': 'JSON nesting exceeds maximum depth of 100'}}
load_json JSONDepthError is ProtocolError: True
canonical_document_bytes CanonicalDepthError is ProtocolError: False
canonical_bytes CanonicalDepthError is ProtocolError: False

```

### forced-recursion.log

```text
canonical_document_bytes CanonicalDepthError ProtocolError= False cause= RecursionError
canonical_bytes CanonicalDepthError ProtocolError= False cause= RecursionError

```

### list-bound-narrowed.log

```text
F                                                                        [100%]
=================================== FAILURES ===================================
_______________ test_canonicalization_rejects_over_deep_nesting ________________

    def test_canonicalization_rejects_over_deep_nesting():
        bound = signing.MAX_JSON_DEPTH
>       signing.canonical_document_bytes(_nested_list(bound))

tests/test_registry.py:330: 
_ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ 
src/csk_registry/signing.py:52: in canonical_document_bytes
    return _canonical_document(value)
           ^^^^^^^^^^^^^^^^^^^^^^^^^^
src/csk_registry/signing.py:57: in _canonical_document
    _validate_ccj(value)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:86: in _validate_ccj
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

src/csk_registry/signing.py:82: CanonicalDepthError
=============================== warnings summary ===============================
../../../../tmp/r5-review/venv/lib/python3.14/site-packages/fastapi/testclient.py:1
  /tmp/r5-review/venv/lib/python3.14/site-packages/fastapi/testclient.py:1: StarletteDeprecationWarning: Using `httpx` with `starlette.testclient` is deprecated; install `httpx2` instead.
    from starlette.testclient import TestClient as TestClient  # noqa

../../../../tmp/r5-review/venv/lib/python3.14/site-packages/starlette/testclient.py:53
  /tmp/r5-review/venv/lib/python3.14/site-packages/starlette/testclient.py:53: DeprecationWarning: The anyio.abc.BlockingPortal alias is deprecated, use anyio.from_thread.BlockingPortal instead.
    _PortalFactoryType = Callable[[], AbstractContextManager[anyio.abc.BlockingPortal]]

-- Docs: https://docs.pytest.org/en/stable/how-to/capture-warnings.html
=========================== short test summary info ============================
FAILED tests/test_registry.py::test_canonicalization_rejects_over_deep_nesting
1 failed, 2 warnings in 4.00s

EXIT=1

```

### object-bound-narrowed.log

```text
F                                                                        [100%]
=================================== FAILURES ===================================
___________________ test_load_json_rejects_over_deep_nesting ___________________

raw = '{"a":{"a":{"a":{"a":{"a":{"a":{"a":{"a":{"a":{"a":{"a":{"a":{"a":{"a":{"a":{"a":{"a":{"a":{"a":{"a":{"a":{"a":{"a":{"...":{"a":{"a":{"a":1}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}'

    def load_json(raw: bytes | str) -> Any:
        """Parse protocol JSON, rejecting over-deep documents.
    
        Documents nested deeper than ``MAX_JSON_DEPTH`` (100 levels of
        objects/arrays) raise :class:`JSONDepthError`, a ``ProtocolError``;
        ``RecursionError`` from the parser or canonicalization is mapped to
        the same error so callers only handle ``ProtocolError``.
        """
        if (isinstance(raw, bytes) and raw.startswith(b"\xef\xbb\xbf")) or (
            isinstance(raw, str) and raw.startswith("\ufeff")
        ):
            raise ProtocolError("protocol JSON must not contain a byte-order mark")
    
        def object_pairs(pairs: list[tuple[str, Any]]) -> dict[str, Any]:
            result: dict[str, Any] = {}
            for key, value in pairs:
                if key in result:
                    raise ProtocolError(f"duplicate JSON object key: {key!r}")
                result[key] = value
            return result
    
        def parse_integer(text: str) -> int:
            value = int(text)
            if text == "-0" or not -MAX_SAFE_INTEGER <= value <= MAX_SAFE_INTEGER:
                raise ProtocolError(f"JSON integer is not shortest-form or safe: {text}")
            return value
    
        def reject_number(text: str) -> None:
            raise ProtocolError(f"registry JSON does not allow non-integer number {text!r}")
    
        try:
            _check_json_depth(raw)
            value = json.loads(
                raw,
                object_pairs_hook=object_pairs,
                parse_int=parse_integer,
                parse_float=reject_number,
                parse_constant=reject_number,
            )
>           canonical_document_bytes(value)

src/csk_registry/protocol.py:134: 
_ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ 
src/csk_registry/signing.py:52: in canonical_document_bytes
    return _canonical_document(value)
           ^^^^^^^^^^^^^^^^^^^^^^^^^^
src/csk_registry/signing.py:57: in _canonical_document
    _validate_ccj(value)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
src/csk_registry/signing.py:97: in _validate_ccj
    _validate_ccj(item, depth + 1)
_ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ 

value = {'a': 1}, depth = 99

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
            if depth + 1 > MAX_JSON_DEPTH:
                raise CanonicalDepthError(
                    f"JSON nesting exceeds maximum depth of {MAX_JSON_DEPTH}"
                )
            for item in value:
                _validate_ccj(item, depth + 1)
            return
        if isinstance(value, dict):
            if depth + 1 >= MAX_JSON_DEPTH:
>               raise CanonicalDepthError(
                    f"JSON nesting exceeds maximum depth of {MAX_JSON_DEPTH}"
                )
E               csk_registry.signing.CanonicalDepthError: JSON nesting exceeds maximum depth of 100

src/csk_registry/signing.py:90: CanonicalDepthError

The above exception was the direct cause of the following exception:

    def test_load_json_rejects_over_deep_nesting():
        bound = signing.MAX_JSON_DEPTH
        assert isinstance(load_json("[" * bound + "]" * bound), list)
>       assert isinstance(load_json('{"a":' * bound + "1" + "}" * bound), dict)
                          ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

tests/test_registry.py:303: 
_ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ 

raw = '{"a":{"a":{"a":{"a":{"a":{"a":{"a":{"a":{"a":{"a":{"a":{"a":{"a":{"a":{"a":{"a":{"a":{"a":{"a":{"a":{"a":{"a":{"a":{"...":{"a":{"a":{"a":1}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}'

    def load_json(raw: bytes | str) -> Any:
        """Parse protocol JSON, rejecting over-deep documents.
    
        Documents nested deeper than ``MAX_JSON_DEPTH`` (100 levels of
        objects/arrays) raise :class:`JSONDepthError`, a ``ProtocolError``;
        ``RecursionError`` from the parser or canonicalization is mapped to
        the same error so callers only handle ``ProtocolError``.
        """
        if (isinstance(raw, bytes) and raw.startswith(b"\xef\xbb\xbf")) or (
            isinstance(raw, str) and raw.startswith("\ufeff")
        ):
            raise ProtocolError("protocol JSON must not contain a byte-order mark")
    
        def object_pairs(pairs: list[tuple[str, Any]]) -> dict[str, Any]:
            result: dict[str, Any] = {}
            for key, value in pairs:
                if key in result:
                    raise ProtocolError(f"duplicate JSON object key: {key!r}")
                result[key] = value
            return result
    
        def parse_integer(text: str) -> int:
            value = int(text)
            if text == "-0" or not -MAX_SAFE_INTEGER <= value <= MAX_SAFE_INTEGER:
                raise ProtocolError(f"JSON integer is not shortest-form or safe: {text}")
            return value
    
        def reject_number(text: str) -> None:
            raise ProtocolError(f"registry JSON does not allow non-integer number {text!r}")
    
        try:
            _check_json_depth(raw)
            value = json.loads(
                raw,
                object_pairs_hook=object_pairs,
                parse_int=parse_integer,
                parse_float=reject_number,
                parse_constant=reject_number,
            )
            canonical_document_bytes(value)
            return value
        except JSONDepthError:
            raise
        except CanonicalDepthError as exc:
>           raise JSONDepthError(str(exc)) from exc
E           csk_registry.protocol.JSONDepthError: JSON nesting exceeds maximum depth of 100

src/csk_registry/protocol.py:139: JSONDepthError
=============================== warnings summary ===============================
../../../../tmp/r5-review/venv/lib/python3.14/site-packages/fastapi/testclient.py:1
  /tmp/r5-review/venv/lib/python3.14/site-packages/fastapi/testclient.py:1: StarletteDeprecationWarning: Using `httpx` with `starlette.testclient` is deprecated; install `httpx2` instead.
    from starlette.testclient import TestClient as TestClient  # noqa

../../../../tmp/r5-review/venv/lib/python3.14/site-packages/starlette/testclient.py:53
  /tmp/r5-review/venv/lib/python3.14/site-packages/starlette/testclient.py:53: DeprecationWarning: The anyio.abc.BlockingPortal alias is deprecated, use anyio.from_thread.BlockingPortal instead.
    _PortalFactoryType = Callable[[], AbstractContextManager[anyio.abc.BlockingPortal]]

-- Docs: https://docs.pytest.org/en/stable/how-to/capture-warnings.html
=========================== short test summary info ============================
FAILED tests/test_registry.py::test_load_json_rejects_over_deep_nesting - csk...
1 failed, 2 warnings in 4.60s

EXIT=1

```

### hosted.log

```text
$ sh scripts/remote-gate.sh
remote gate: attempt 1/3: pushing fb8780e1d6f3d5e347ed892477f70baf2d8d4a1d as gate/STORY-260910-2xe3n2/260918-113413-64236-1
remote: 
remote: Create a pull request for 'gate/STORY-260910-2xe3n2/260918-113413-64236-1' on GitHub by visiting:        
remote:      https://github.com/relux-works/curator-skill-registry/pull/new/gate/STORY-260910-2xe3n2/260918-113413-64236-1        
remote: 
remote gate: run 35340282852 (https://github.com/relux-works/curator-skill-registry/actions/runs/35340282852)
remote gate: run 35340282852 finished: success
  Tests / Python 3.11 on macos-latest: success  
  Type check / mypy strict: success  
  Docker build: success  
  Tests / Python 3.14 on windows-latest: success  
  Tests / Python 3.11 on ubuntu-latest: success  
  Tests / Python 3.11 on windows-latest: success  
  Tests / Python 3.14 on macos-latest: success  
  Tests / Python 3.14 on ubuntu-latest: success  
  Build distribution: success  
[exit 0]
coverage_unit=exact_command_shard required=1 green=1 failed=0 missing=0 test_case_coverage=unknown

```
