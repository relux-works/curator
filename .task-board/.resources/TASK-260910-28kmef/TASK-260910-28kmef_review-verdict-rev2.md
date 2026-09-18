# TASK-260910-28kmef — review verdict revision 2

Verdict: **accepted**. R1 is resolved; no required corrections remain.

Candidate: `fd0a94753d03dae85abbdbe0bdd28200e9da8ed1`, base `c7ef32c75cc8dfda1abe38647af282a03175e8d3`.
Published patch SHA-256 independently verified: `4a17ac429fbd383a973140a8aa65b0c991076eabb5ac2c12a807155ca1f4e78a`.
All published-tree files were compared byte-for-byte with the managed worktree before review and after validation. No candidate edits, new caches, commits, branches or hosted gates were made by this reviewer. Tests used a git archive of the exact candidate in a disposable directory; mutations used separate copies.

## Per-item review

| Requirement | Verdict and evidence |
|---|---|
| R1 canonicalization exception contract | PASS: `signing.py:32` inherits CanonicalError and ProtocolError, retaining ValueError. `errors.py:12` supplies the shared base without package imports; `protocol.py:10` explicitly re-exports it. Existing importers resolve the identical class, with no protocol/signing import cycle. Both entry points (`signing.py:41,51`) share `_canonical_document` at :60. |
| Actual depth and injected recursion | PASS: tests `test_registry.py:328` catch ProtocolError for actual over-depth input for both entry points. :349 covers both entry points × CCJ validation/serialization recursion (4/4 injected cases); 2/2 actual entry points are covered. `_canonical_document` maps recursion at `signing.py:66`; list/dict depth checks are at :86/:94. |
| load_json and internal callers | PASS: iterative O(n)-time/O(1)-space scan `protocol.py:38`; parsing and canonicalization failures mapped at :133–144. Direct reviewer injection into json.loads produced JSONDepthError(ProtocolError), with message naming 100. Validation wrappers at :147, :211, :259 preserve depth handling. Non-depth CanonicalError remains ValueError; external import sites unchanged. |
| HTTP behavior | PASS: only 1/1 body-parsing endpoints exists, POST /v1/records (`app.py:381,397`). Depth exceptions precede generic invalid_record handling at :416; idempotency digest handling at :438. Committed HTTP test `test_registry.py:375` covers 101 and 5000 levels and idempotency; malformed non-depth JSON stays invalid_record. Reviewer live Uvicorn probes: 6/6 returned 400 invalid_json, none 500 (see transcript). |
| Cursor mapping | PASS: `app.py:579` keeps 404 invalid_cursor; committed regression at `test_registry.py:404`. |
| Docs | PASS: 100-level bound `signing.py:19`, `protocol.py:95`, README.md:163. CHANGELOG.md:36 contains R5, mappings and actual pinned spec revision. |
| Existing store behavior and suite | PASS: independent full suite 175 passed using clean detached spec root 47c3c8cbd5da5e3fd6d99b0327d382c6a36494fe. No store regression observed. |
| Static/build gates | PASS: independent strict mypy, 15 source files, exit 0; git diff --check exit 0. No standalone linter configured. Build and Docker accepted from exact revision's attached hosted log (not rerun). |
| Revision scope | PASS: compared both published patches. Only R1 shared error/hierarchy/import/docstrings and direct tests differ, plus the producer-disclosed changelog reference correction dced9b8 → 47c3c8c required by corrected rules. app.py and README are byte-identical across revisions; no unrelated scope or CI pin changes. |
| Negative evidence | PASS: 2/2 narrowing mutants killed plus 1/1 old-hierarchy regression mutant. Details below. |

Read producer rules/brief/results, both CR patches, rev2 validation log, repository R5 finding, and pinned profiles/registry-service.md §§8–9 / protocol/registry.md §9.1. The spec requires 400 for malformed JSON and the existing generic error envelope; invalid_json is the requested stable code, not a newly changed schema. The corrected 47c3c8c root is clean and matches CI. No new spec gap or product decision identified.

## Independent validation

Shell: zsh, `set -o pipefail`. Interpreter: CPython 3.14.6. Disposable root: /var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/28kmef-review2-oxw9vqtf.
Environment was installed with `python3 -m venv "$ROOT/venv"` and `"$ROOT/venv/bin/pip" install -e "$ROOT/candidate[dev]"` (exit 0).
Commands run from `$ROOT/candidate`:

```sh
CURATOR_CONFORMANCE_ROOT=/tmp/spec-47c3c8c/conformance/v1 "$ROOT/venv/bin/python" -m pytest -q
# exit 0
"$ROOT/venv/bin/python" -m mypy
# exit 0
```

### pytest transcript

```text
........................................................................ [ 41%]
........................................................................ [ 82%]
...............................                                          [100%]
=============================== warnings summary ===============================
../../../../../../../../var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/28kmef-review2-oxw9vqtf/venv/lib/python3.14/site-packages/fastapi/testclient.py:1
  /var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/28kmef-review2-oxw9vqtf/venv/lib/python3.14/site-packages/fastapi/testclient.py:1: StarletteDeprecationWarning: Using `httpx` with `starlette.testclient` is deprecated; install `httpx2` instead.
    from starlette.testclient import TestClient as TestClient  # noqa

../../../../../../../../var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/28kmef-review2-oxw9vqtf/venv/lib/python3.14/site-packages/starlette/testclient.py:53
  /var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/28kmef-review2-oxw9vqtf/venv/lib/python3.14/site-packages/starlette/testclient.py:53: DeprecationWarning: The anyio.abc.BlockingPortal alias is deprecated, use anyio.from_thread.BlockingPortal instead.
    _PortalFactoryType = Callable[[], AbstractContextManager[anyio.abc.BlockingPortal]]

-- Docs: https://docs.pytest.org/en/stable/how-to/capture-warnings.html
175 passed, 2 warnings in 38.56s
PYTEST_EXIT=0
```

### mypy transcript

```text
Success: no issues found in 15 source files
MYPY_EXIT=0
```

Hosted evidence accepted from `TASK-260910-28kmef_change-request_rev2-validation.log`: run 35342221083 reports all six Python 3.11/3.14 × Windows/macOS/Ubuntu test jobs, strict mypy, distribution and Docker builds successful, command exit 0. I did not rerun the hosted gate. Local validation covers macOS/Python 3.14.6 only; this review does not claim independent execution on the other environments.

## Live HTTP attacks

Launched Uvicorn on an ephemeral loopback socket using the production create_app through the existing authenticated test fixture. Submitted raw 20,000-level arrays, 20,000-level objects, and a valid signed record with a 20,000-level extra signature member. Each ran without and with Idempotency-Key. Server was stopped and joined before completion.

```text
/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/28kmef-review2-oxw9vqtf/venv/lib/python3.14/site-packages/fastapi/testclient.py:1: StarletteDeprecationWarning: Using `httpx` with `starlette.testclient` is deprecated; install `httpx2` instead.
  from starlette.testclient import TestClient as TestClient  # noqa
array-20000 idempotent=False 400 {"error":{"code":"invalid_json","message":"JSON nesting exceeds maximum depth of 100"}}
array-20000 idempotent=True 400 {"error":{"code":"invalid_json","message":"JSON nesting exceeds maximum depth of 100"}}
object-20000 idempotent=False 400 {"error":{"code":"invalid_json","message":"JSON nesting exceeds maximum depth of 100"}}
object-20000 idempotent=True 400 {"error":{"code":"invalid_json","message":"JSON nesting exceeds maximum depth of 100"}}
signature-member-20000 idempotent=False 400 {"error":{"code":"invalid_json","message":"JSON nesting exceeds maximum depth of 100"}}
signature-member-20000 idempotent=True 400 {"error":{"code":"invalid_json","message":"JSON nesting exceeds maximum depth of 100"}}
HTTP probes: 6/6 passed; body-parsing endpoints covered: 1/1
PROBES_EXIT=0
```

Additional independent checks: dictionaries at depth 100 accepted by both canonicalization entry points; depth 101 raised an exception satisfying ProtocolError, CanonicalError and ValueError for both. Injected json.loads RecursionError mapped to JSONDepthError with maximum-depth message. Exit 0.

## Narrowing mutants

Every mutant ran committed tests in its own copy with PYTHONPATH pointing to that copy and bytecode writes disabled. The baseline stayed untouched.

1. `ccj-list-bound-only`: narrowed list rejection from depth >100 to >101, leaving dictionary checking intact. `test_canonicalization_rejects_over_deep_nesting` failed (DID NOT RAISE), pytest exit 1.
2. `recursion-only-for-lists`: narrowed RecursionError translation to list values, re-raising for dictionaries. All four parametrized entry point × fault tests failed with RecursionError, pytest exit 1.
3. `R1-old-hierarchy`: restored CanonicalDepthError(CanonicalError), removing ProtocolError compatibility. All five canonicalization tests failed, pytest exit 1.

Measured kill ratio: 2/2 requested narrowing mutants, 1/1 additional hierarchy regression. This is targeted mutation evidence, not exhaustive branch-mutation coverage. The HTTP probes establish the six tested shapes, not a claim that arbitrary failures can never produce 500.

```text
ccj-list-bound-only exit=1
===================
_______________ test_canonicalization_rejects_over_deep_nesting ________________

    def test_canonicalization_rejects_over_deep_nesting():
        bound = signing.MAX_JSON_DEPTH
        signing.canonical_document_bytes(_nested_list(bound))
>       with pytest.raises(signing.CanonicalDepthError) as excinfo:
             ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
E       Failed: DID NOT RAISE CanonicalDepthError

tests/test_registry.py:331: Failed
=============================== warnings summary ===============================
../../../../../../../../var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/28kmef-review2-oxw9vqtf/venv/lib/python3.14/site-packages/fastapi/testclient.py:1
  /var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/28kmef-review2-oxw9vqtf/venv/lib/python3.14/site-packages/fastapi/testclient.py:1: StarletteDeprecationWarning: Using `httpx` with `starlette.testclient` is deprecated; install `httpx2` instead.
    from starlette.testclient import TestClient as TestClient  # noqa

../../../../../../../../var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/28kmef-review2-oxw9vqtf/venv/lib/python3.14/site-packages/starlette/testclient.py:53
  /var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/28kmef-review2-oxw9vqtf/venv/lib/python3.14/site-packages/starlette/testclient.py:53: DeprecationWarning: The anyio.abc.BlockingPortal alias is deprecated, use anyio.from_thread.BlockingPortal instead.
    _PortalFactoryType = Callable[[], AbstractContextManager[anyio.abc.BlockingPortal]]

-- Docs: https://docs.pytest.org/en/stable/how-to/capture-warnings.html
=========================== short test summary info ============================
FAILED tests/test_registry.py::test_canonicalization_rejects_over_deep_nesting
1 failed, 87 deselected, 2 warnings in 1.45s

recursion-only-for-lists exit=1
boom

tests/test_registry.py:361: RecursionError
=============================== warnings summary ===============================
../../../../../../../../var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/28kmef-review2-oxw9vqtf/venv/lib/python3.14/site-packages/fastapi/testclient.py:1
  /var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/28kmef-review2-oxw9vqtf/venv/lib/python3.14/site-packages/fastapi/testclient.py:1: StarletteDeprecationWarning: Using `httpx` with `starlette.testclient` is deprecated; install `httpx2` instead.
    from starlette.testclient import TestClient as TestClient  # noqa

../../../../../../../../var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/28kmef-review2-oxw9vqtf/venv/lib/python3.14/site-packages/starlette/testclient.py:53
  /var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/28kmef-review2-oxw9vqtf/venv/lib/python3.14/site-packages/starlette/testclient.py:53: DeprecationWarning: The anyio.abc.BlockingPortal alias is deprecated, use anyio.from_thread.BlockingPortal instead.
    _PortalFactoryType = Callable[[], AbstractContextManager[anyio.abc.BlockingPortal]]

-- Docs: https://docs.pytest.org/en/stable/how-to/capture-warnings.html
=========================== short test summary info ============================
FAILED tests/test_registry.py::test_canonicalization_maps_recursion_error_to_protocol_error[validate_ccj-canonical_bytes]
FAILED tests/test_registry.py::test_canonicalization_maps_recursion_error_to_protocol_error[validate_ccj-canonical_document_bytes]
FAILED tests/test_registry.py::test_canonicalization_maps_recursion_error_to_protocol_error[json_dumps-canonical_bytes]
FAILED tests/test_registry.py::test_canonicalization_maps_recursion_error_to_protocol_error[json_dumps-canonical_document_bytes]
4 failed, 84 deselected, 2 warnings in 2.98s

R1-old-hierarchy exit=1
= warnings summary ===============================
../../../../../../../../var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/28kmef-review2-oxw9vqtf/venv/lib/python3.14/site-packages/fastapi/testclient.py:1
  /var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/28kmef-review2-oxw9vqtf/venv/lib/python3.14/site-packages/fastapi/testclient.py:1: StarletteDeprecationWarning: Using `httpx` with `starlette.testclient` is deprecated; install `httpx2` instead.
    from starlette.testclient import TestClient as TestClient  # noqa

../../../../../../../../var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/28kmef-review2-oxw9vqtf/venv/lib/python3.14/site-packages/starlette/testclient.py:53
  /var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/28kmef-review2-oxw9vqtf/venv/lib/python3.14/site-packages/starlette/testclient.py:53: DeprecationWarning: The anyio.abc.BlockingPortal alias is deprecated, use anyio.from_thread.BlockingPortal instead.
    _PortalFactoryType = Callable[[], AbstractContextManager[anyio.abc.BlockingPortal]]

-- Docs: https://docs.pytest.org/en/stable/how-to/capture-warnings.html
=========================== short test summary info ============================
FAILED tests/test_registry.py::test_canonicalization_rejects_over_deep_nesting
FAILED tests/test_registry.py::test_canonicalization_maps_recursion_error_to_protocol_error[validate_ccj-canonical_bytes]
FAILED tests/test_registry.py::test_canonicalization_maps_recursion_error_to_protocol_error[validate_ccj-canonical_document_bytes]
FAILED tests/test_registry.py::test_canonicalization_maps_recursion_error_to_protocol_error[json_dumps-canonical_bytes]
FAILED tests/test_registry.py::test_canonicalization_maps_recursion_error_to_protocol_error[json_dumps-canonical_document_bytes]
5 failed, 83 deselected, 2 warnings in 2.59s

Killed 2/2 narrowing mutants and 1/1 hierarchy regression mutant
MUTANTS_EXIT=0
```

## Reproducible attack scripts

```python
import sys, tempfile, json, socket, threading, time
from pathlib import Path
sys.path.insert(0, str(Path.cwd() / "tests"))
from test_registry import _client, _body
import httpx, uvicorn
from csk_registry import signing
from csk_registry.protocol import ProtocolError
with tempfile.TemporaryDirectory() as td:
    client, key, token = _client(Path(td))
    sock = socket.socket(); sock.bind(("127.0.0.1", 0)); port=sock.getsockname()[1]
    server=uvicorn.Server(uvicorn.Config(client.app, log_level="error"))
    thread=threading.Thread(target=server.run, kwargs={"sockets":[sock]}); thread.start()
    try:
        for _ in range(200):
            if server.started: break
            time.sleep(.05)
        assert server.started
        valid=client.app.state.auditor_key.sign_record(_body())
        array="["*20000+"0"+"]"*20000
        obj='{"k":'*20000+"0"+"}"*20000
        nested=json.dumps(valid).replace('"sig": {', '"sig": {"extra":'+array+',', 1)
        for name, body in [("array-20000",array),("object-20000",obj),("signature-member-20000",nested)]:
            for idem in [False,True]:
                headers={"Authorization":f"Bearer {token}","Content-Type":"application/json"}
                if idem: headers["Idempotency-Key"]="probe-"+name
                resp=httpx.post(f"http://127.0.0.1:{port}/v1/records",content=body,headers=headers)
                print(name,"idempotent="+str(idem),resp.status_code,resp.text)
                assert resp.status_code==400 and resp.json()["error"]["code"]=="invalid_json"
    finally:
        server.should_exit=True; thread.join(10); sock.close()
        assert not thread.is_alive()
print("HTTP probes: 6/6 passed; body-parsing endpoints covered: 1/1")
```

```python
import subprocess, pathlib, shutil, os, sys
root=pathlib.Path(__file__).parent
source=root/"candidate"
mutants=[
("ccj-list-bound-only", "signing.py", "if depth + 1 > MAX_JSON_DEPTH:", "if depth + 1 > MAX_JSON_DEPTH + 1:", "test_canonicalization_rejects_over_deep_nesting"),
("recursion-only-for-lists", "signing.py", "except RecursionError as exc:\n        raise CanonicalDepthError(", "except RecursionError as exc:\n        if not isinstance(value, list):\n            raise\n        raise CanonicalDepthError(", "test_canonicalization_maps_recursion_error_to_protocol_error"),
("R1-old-hierarchy", "signing.py", "class CanonicalDepthError(CanonicalError, ProtocolError):", "class CanonicalDepthError(CanonicalError):", "canonicalization"),
]
for name, file, old, new, selection in mutants:
    target=root/name; shutil.copytree(source,target,ignore=shutil.ignore_patterns("__pycache__",".pytest_cache",".mypy_cache","*.egg-info"))
    path=target/"src/csk_registry"/file
    text=path.read_text(); assert old in text; path.write_text(text.replace(old,new,1))
    env={**os.environ,"PYTHONPATH":str(target/"src"),"PYTHONDONTWRITEBYTECODE":"1"}
    proc=subprocess.run([sys.executable,"-m","pytest","-q","tests/test_registry.py","-k",selection],cwd=target,env=env,text=True,stdout=subprocess.PIPE,stderr=subprocess.STDOUT)
    (root/(name+".log")).write_text(proc.stdout)
    print(name,"exit="+str(proc.returncode)); print(proc.stdout[-1800:])
    assert proc.returncode==1
print("Killed 2/2 narrowing mutants and 1/1 hierarchy regression mutant")
```

## Lifecycle

Queried `task-board spawn goal "$TASK_BOARD_RUN_ID"`: run is not goal-bound. Checklist is fully checked. Verdict artifact is attached before accept_cr revision=2, which routes accepted work to integrating. No done transition or commit_ack supplied. Round-1 R1 resolution also recorded in the task-scoped review logbook.
