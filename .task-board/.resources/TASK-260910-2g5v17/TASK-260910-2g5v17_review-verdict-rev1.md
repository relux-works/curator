# TASK-260910-2g5v17 — review verdict, revision 1

Verdict: **changes_requested**. Route to `to-dev`; no acceptance or integration.

Candidate tree: `376939af3147be39178c5b489b0a4a634f2114a1`.
Base: `c7ef32c75cc8dfda1abe38647af282a03175e8d3`.
Published patch SHA-256 verified: `31e8e27bc45ead7b5309a57dc84bef5327cc36beb94f35bbebfb703b6bdae907`.
Reviewed exact tree via git archive. Candidate tracked content and untracked results artifact match the tree; original worktree unchanged, no review files/caches created there.
`task-board spawn goal "$TASK_BOARD_RUN_ID"`: Active Goal none (not goal-bound).

## Required correction F1 — medium: concurrent comparison loses the import contract

Production sites: `src/csk_registry/bundle.py:149` reads high-water outside the write transaction; `bundle.py:195` calls `append_imports`; `src/csk_registry/store.py:1219-1237` rechecks it but raises a generic `ValueError("upstream high-water advanced during import")`. The store does not receive the override flag. The exception bypasses `bundle.py:196` outcome logging, and `_cmd_import_bundle` at `cli.py:271` prints that generic text.

Deterministic reproduction uses two real Store connections to the same database and real `import_bundle` calls. A wrapper pauses the outer operation at `append_imports`, commits a second import, then calls the original method. This represents a legal scheduling interleaving, not a fabricated high-water row.

1. Initially v1; outer call offers v2; second connection imports v3 before outer BEGIN IMMEDIATE. Without flag, refusal lacks `import_upstream_rollback`, key_id, both boundaries, and its structured refusal audit event.
2. Same schedule with `accept_older_upstream=True`: still fails with the same generic error instead of warning/importing without lowering v3.
3. Initially unknown upstream; competing first imports offer different v1 bodies. The second commits first. Outer call refuses generically instead of `import_upstream_inconsistent` with key_id/boundaries and audit event, including when the flag is set.

The guard preserves fail-closed state and does not lower high-water. This finding is a diagnostic/override/observability contract failure, not evidence that concurrent rollback records were accepted. The normative service profile §4 explicitly includes concurrent administrative imports; the brief requires closed refusals and flag semantics without a serial-only exception.

Correction: make the authoritative comparison under writer serialization carry the same outcomes as the ordinary import path, including current persisted boundary, offered boundary, key_id, override policy and no-op. Emit the documented refusal/warning audit event for the actual comparison. Preserve atomicity and never permit equal-version inconsistency. Add deterministic competing-writer regression tests for the three schedules above (and equal-identical no-op). No spec or CI pin change is requested.

## Per-item review

| Requirement | Evidence and result |
| --- | --- |
| Per-key durable table, migration | PASS: store.py:57,158,286,588,1175; tests:2880,3112,3199. In registry.db, no side file. |
| Atomic records + high-water | PASS: store.py:617 BEGIN IMMEDIATE/rollback; 1217,1245,1253 update within one transaction. Independent failure at BEFORE INSERT ON upstream_high_water occurs after record writes and rolls back log, import ledger and high-water. |
| Rollback closed diagnostic | PASS serial test:2898 and CLI test:3149; FAIL competing writer F1. |
| Inconsistent refusal, including flag | PASS serial test:2925; FAIL closed diagnostic/audit in competing first-import F1. State still refuses. |
| Identical import no-op | PASS ordinary path bundle.py:163; test:2965 checks count 0, high-water unchanged, head unchanged and audit noop. Fingerprint ledger suppresses appends. |
| Newer advances | PASS bundle.py:187, store.py:1253; test:2993. |
| Older override preserves high-water | PASS serial bundle.py:173; tests:3013,3149; FAIL race F1. |
| Failed import untouched | PASS tests:3058,3086 and independent late-failure probe. Existing trigger test aborts before log insertion; reviewer additionally aborts after record writes. |
| First import, per-key isolation | PASS tests:2880,3112. |
| Security storage/backup | PASS db mode independently observed 0600; store.py:1936 private file and :1946 sidecar protection. backup_to :1918 uses SQLite backup; test:3234 preserves high-water. Documented verify-backup does not compare upstream state; acceptable optional choice, signed checkpoint lacks it. |
| Docs/release note | PASS README.md:104, SECURITY.md:38, CHANGELOG.md:36 P4 with dced9b8; no protocol/wire edits. |
| Spec diagnostic spelling | Pinned profiles/registry-service.md provides no import-upstream diagnostic to reuse. Brief spellings appropriate. protocol/registry.md §5 defines below/equal comparison. |
| Independent pytest/mypy | 178 passed on actual CI pin 47c3c8c, strict mypy 0. Requested dced9b8 yields 177 pass / 1 pre-existing missing-vector failure, reproduced on base; see below. |
| Scope, lint/build, hygiene | Only declared 8 paths. git diff --check exits 0. No separate lint configured. Hosted build/docker and platform matrix accepted from attached exact revision validation log, not rerun. No reviewer candidate edits. |

Seven of seven requested serial behavioral scenarios have committed tests and passed independently. Two of two narrowing mutants were detected. Three of three tested competing-writer schedules expose F1. These ratios describe those bounded scenarios, not exhaustive concurrency or corruption coverage.

## Independent validation and fixture anomaly

Interpreter `/tmp/p4-review/venv/bin/python`: Python 3.14.6. Shell `/bin/bash`, `set -o pipefail`; cwd `/tmp/p4-review/candidate`, archived from the candidate tree. Venv outside all repository trees, installed with `pip install -e '/tmp/p4-review/candidate[dev]'`.

Requested spec detached worktree `/tmp/p4-review/spec` at `dced9b8317e0e8af79edf2d0539b32bd22b6c85b`. Command:
`CURATOR_CONFORMANCE_ROOT=/tmp/p4-review/spec/conformance/v1 /tmp/p4-review/venv/bin/python -m pytest -q` (tee, pipefail): exit 1, 177 passed, 1 failed. `test_shared_service_startup_checkpoint_cases` requires `checkpoint_cases`, absent at this older revision. Reproduced the same targeted test on archived base c7ef32c: exit 1. This is not introduced by P4.

Actual CI pin is already `47c3c8cbd5da5e3fd6d99b0327d382c6a36494fe` in both base and candidate `.github/workflows/ci.yml:30`. Archived that spec into `/tmp/p4-review/spec-ci`. Full rerun justified by fixture mismatch:
`CURATOR_CONFORMANCE_ROOT=/tmp/p4-review/spec-ci/conformance/v1 /tmp/p4-review/venv/bin/python -m pytest -q`: exit 0, 178 passed, 2 dependency deprecation warnings.
`/tmp/p4-review/venv/bin/python -m mypy`: exit 0, 14 source files, strict configured in pyproject.toml.

Reviewer setup corrections: initial pytest/mypy started before pip finished and reported missing dependencies; rerun after pip exited 0. An initial base extraction was invoked from an archive without .git, so its targeted test found no file (exit 4); extracted from the actual repository and reran successfully to reproduce the expected missing-vector failure. Neither setup attempt is counted as candidate evidence.

Hosted validation resource `TASK-260910-2g5v17_change-request_rev1-validation.log` read: gate exit 0; GitHub run 35340311130, all six OS/Python tests, strict mypy, distribution and Docker success. This evidence is accepted as attached, not independently rerun. Producer results read but not used as the sole basis for any acceptance.

## Narrowing mutants

Disposable `/tmp/p4-review/mutant`, candidate snapshot, source restored byte-for-byte after each mutation. Real import entry point driven by committed tests.

- M1: bundle.py first `offered["version"] < persisted.version` becomes `< persisted.version - 1`, allowing a one-version rollback. `python -m pytest -q tests/test_registry.py::test_import_rollback_bundle_refused` exits 1: DID NOT RAISE. Caught.
- M2: equal-version branch gains `and not accept_older_upstream`, narrowing proper inconsistency handling to unflagged calls. `python -m pytest -q tests/test_registry.py::test_import_inconsistent_bundle_refused_even_with_flag` exits 1: expected closed diagnostic, actual generic concurrent guard error. Caught, although store defense still prevents mutation. This mutant proves diagnostic coverage, not removal of all inconsistency defenses.

## Probe source and transcripts

### probes.py

```python
import importlib.util, tempfile
from pathlib import Path
from csk_registry.bundle import import_bundle, export_bundle
from csk_registry.store import Store
from csk_registry import signing
spec = importlib.util.spec_from_file_location('test_registry', '/tmp/p4-review/candidate/tests/test_registry.py')
t = importlib.util.module_from_spec(spec); spec.loader.exec_module(t)
with tempfile.TemporaryDirectory() as td:
 p=Path(td); uk, dk, v1, v2=t._upstream_pair(p); d=Store(p/'d.db')
 import_bundle(d,dk,v1,upstream_public_key=uk.public_pinned)
 d._conn.execute("CREATE TRIGGER fail_hw BEFORE INSERT ON upstream_high_water BEGIN SELECT RAISE(ABORT, 'after records before high-water'); END")
 try: import_bundle(d,dk,v2,upstream_public_key=uk.public_pinned)
 except Exception as e: print('transaction injection:',str(e))
 assert d.head()[0]==1 and d.get_upstream_high_water(uk.key_id).version==1
 assert d._conn.execute('SELECT count(*) FROM imported_records').fetchone()[0]==1
 print('PASS: log, import ledger and high-water all rolled back; db mode',oct((p/'d.db').stat().st_mode & 0o777)); d.close()
for override in (False,True):
 with tempfile.TemporaryDirectory() as td:
  p=Path(td); uk,dk,v1,v2=t._upstream_pair(p); u=Store(p/'new.db')
  for r in v2['records']: u.append(r,created_at='2026-07-07T00:00:00Z')
  u.append(uk.sign_record(t._body('audited',name='third')),created_at='2026-07-07T00:00:00Z'); v3=export_bundle(u,uk); u.close()
  d=Store(p/'d.db'); other=Store(p/'d.db'); import_bundle(d,dk,v1,upstream_public_key=uk.public_pinned)
  original=d.append_imports
  def race(*args,**kwargs):
   import_bundle(other,dk,v3,upstream_public_key=uk.public_pinned)
   return original(*args,**kwargs)
  d.append_imports=race
  try:
   result=import_bundle(d,dk,v2,upstream_public_key=uk.public_pinned,accept_older_upstream=override)
   print('race override',override,'returned',result)
  except Exception as e: print('race override',override,'ERROR',type(e).__name__,str(e))
  print('race final high-water',d.get_upstream_high_water(uk.key_id).version)
  d.close(); other.close()
with tempfile.TemporaryDirectory() as td:
 p=Path(td); uk,dk,v1,v2=t._upstream_pair(p); fork=Store(p/'fork.db')
 fork.append(uk.sign_record(t._body('audited',name='fork')),created_at='2026-07-07T00:00:00Z')
 vb=export_bundle(fork,uk); fork.close(); d=Store(p/'d.db'); other=Store(p/'d.db'); original=d.append_imports
 def race_equal(*args,**kwargs):
  import_bundle(other,dk,vb,upstream_public_key=uk.public_pinned)
  return original(*args,**kwargs)
 d.append_imports=race_equal
 try: import_bundle(d,dk,v1,upstream_public_key=uk.public_pinned,accept_older_upstream=True)
 except Exception as e: print('equal-version race ERROR',type(e).__name__,str(e))
 d.close(); other.close()

```

### probes.log

```text
/tmp/p4-review/venv/lib/python3.14/site-packages/fastapi/testclient.py:1: StarletteDeprecationWarning: Using `httpx` with `starlette.testclient` is deprecated; install `httpx2` instead.
  from starlette.testclient import TestClient as TestClient  # noqa
transaction injection: after records before high-water
PASS: log, import ledger and high-water all rolled back; db mode 0o600
race override False ERROR ValueError upstream high-water advanced during import
race final high-water 3
race override True ERROR ValueError upstream high-water advanced during import
race final high-water 3
equal-version race ERROR ValueError upstream high-water advanced during import

```

### pytest-final.log

```text
........................................................................ [ 40%]
.F...................................................................... [ 80%]
..................................                                       [100%]
=================================== FAILURES ===================================
_________________ test_shared_service_startup_checkpoint_cases _________________

tmp_path = PosixPath('/private/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/pytest-of-administrator/pytest-125/test_shared_service_startup_ch0')
monkeypatch = <_pytest.monkeypatch.MonkeyPatch object at 0x105206520>
caplog = <_pytest.logging.LogCaptureFixture object at 0x104fa3b60>

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
../../../../tmp/p4-review/venv/lib/python3.14/site-packages/fastapi/testclient.py:1
  /tmp/p4-review/venv/lib/python3.14/site-packages/fastapi/testclient.py:1: StarletteDeprecationWarning: Using `httpx` with `starlette.testclient` is deprecated; install `httpx2` instead.
    from starlette.testclient import TestClient as TestClient  # noqa

../../../../tmp/p4-review/venv/lib/python3.14/site-packages/starlette/testclient.py:53
  /tmp/p4-review/venv/lib/python3.14/site-packages/starlette/testclient.py:53: DeprecationWarning: The anyio.abc.BlockingPortal alias is deprecated, use anyio.from_thread.BlockingPortal instead.
    _PortalFactoryType = Callable[[], AbstractContextManager[anyio.abc.BlockingPortal]]

-- Docs: https://docs.pytest.org/en/stable/how-to/capture-warnings.html
=========================== short test summary info ============================
FAILED tests/test_protocol_conformance.py::test_shared_service_startup_checkpoint_cases
1 failed, 177 passed, 2 warnings in 39.05s

```

### base-pinned.log

```text
F                                                                        [100%]
=================================== FAILURES ===================================
_________________ test_shared_service_startup_checkpoint_cases _________________

tmp_path = PosixPath('/private/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/pytest-of-administrator/pytest-129/test_shared_service_startup_ch0')
monkeypatch = <_pytest.monkeypatch.MonkeyPatch object at 0x1090510f0>
caplog = <_pytest.logging.LogCaptureFixture object at 0x109145160>

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
../../../../tmp/p4-review/venv/lib/python3.14/site-packages/fastapi/testclient.py:1
  /tmp/p4-review/venv/lib/python3.14/site-packages/fastapi/testclient.py:1: StarletteDeprecationWarning: Using `httpx` with `starlette.testclient` is deprecated; install `httpx2` instead.
    from starlette.testclient import TestClient as TestClient  # noqa

../../../../tmp/p4-review/venv/lib/python3.14/site-packages/starlette/testclient.py:53
  /tmp/p4-review/venv/lib/python3.14/site-packages/starlette/testclient.py:53: DeprecationWarning: The anyio.abc.BlockingPortal alias is deprecated, use anyio.from_thread.BlockingPortal instead.
    _PortalFactoryType = Callable[[], AbstractContextManager[anyio.abc.BlockingPortal]]

-- Docs: https://docs.pytest.org/en/stable/how-to/capture-warnings.html
=========================== short test summary info ============================
FAILED tests/test_protocol_conformance.py::test_shared_service_startup_checkpoint_cases
1 failed, 2 warnings in 2.79s

```

### pytest-ci-pin.log

```text
........................................................................ [ 40%]
........................................................................ [ 80%]
..................................                                       [100%]
=============================== warnings summary ===============================
../../../../tmp/p4-review/venv/lib/python3.14/site-packages/fastapi/testclient.py:1
  /tmp/p4-review/venv/lib/python3.14/site-packages/fastapi/testclient.py:1: StarletteDeprecationWarning: Using `httpx` with `starlette.testclient` is deprecated; install `httpx2` instead.
    from starlette.testclient import TestClient as TestClient  # noqa

../../../../tmp/p4-review/venv/lib/python3.14/site-packages/starlette/testclient.py:53
  /tmp/p4-review/venv/lib/python3.14/site-packages/starlette/testclient.py:53: DeprecationWarning: The anyio.abc.BlockingPortal alias is deprecated, use anyio.from_thread.BlockingPortal instead.
    _PortalFactoryType = Callable[[], AbstractContextManager[anyio.abc.BlockingPortal]]

-- Docs: https://docs.pytest.org/en/stable/how-to/capture-warnings.html
178 passed, 2 warnings in 43.78s

```

### mypy-final.log

```text
Success: no issues found in 14 source files

```

### M1.log

```text
F                                                                        [100%]
=================================== FAILURES ===================================
_____________________ test_import_rollback_bundle_refused ______________________

tmp_path = PosixPath('/private/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/pytest-of-administrator/pytest-126/test_import_rollback_bundle_re0')

    def test_import_rollback_bundle_refused(tmp_path: Path) -> None:
        from csk_registry.bundle import import_bundle
    
        up_key, down_key, bundle_v1, bundle_v2 = _upstream_pair(tmp_path)
        downstream = Store(tmp_path / "down.db")
        assert (
            import_bundle(downstream, down_key, bundle_v2, upstream_public_key=up_key.public_pinned)  # type: ignore[arg-type,union-attr]
            == 2
        )
        before = downstream.head()[0]
>       with pytest.raises(ValueError, match="import_upstream_rollback"):
             ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
E       Failed: DID NOT RAISE ValueError

tests/test_registry.py:2908: Failed
------------------------------ Captured log call -------------------------------
WARNING  csk_registry.audit:bundle.py:283 {"accepted_older":true,"event":"import_bundle","imported":0,"key_id":"99441d942c1066ed","offered":{"head":"1353a8580eb17092a2ead7c181b759840716065a17df9cb6fe59b5a5fbb43830","log_size":1,"merkle_root":"1353a8580eb17092a2ead7c181b759840716065a17df9cb6fe59b5a5fbb43830","version":1},"persisted":{"head":"f9e9bcbd12d09c733b302ad44779d2eb267bc6d80736cb313876d7c46d846247","log_size":2,"merkle_root":"c7d047f604a6c52206a9320130827849d44424dfb486d40375b146eae692a906","version":2},"result":"ok","warning":"import_upstream_rollback"}
=============================== warnings summary ===============================
../../../../tmp/p4-review/venv/lib/python3.14/site-packages/fastapi/testclient.py:1
  /tmp/p4-review/venv/lib/python3.14/site-packages/fastapi/testclient.py:1: StarletteDeprecationWarning: Using `httpx` with `starlette.testclient` is deprecated; install `httpx2` instead.
    from starlette.testclient import TestClient as TestClient  # noqa

../../../../tmp/p4-review/venv/lib/python3.14/site-packages/starlette/testclient.py:53
  /tmp/p4-review/venv/lib/python3.14/site-packages/starlette/testclient.py:53: DeprecationWarning: The anyio.abc.BlockingPortal alias is deprecated, use anyio.from_thread.BlockingPortal instead.
    _PortalFactoryType = Callable[[], AbstractContextManager[anyio.abc.BlockingPortal]]

-- Docs: https://docs.pytest.org/en/stable/how-to/capture-warnings.html
=========================== short test summary info ============================
FAILED tests/test_registry.py::test_import_rollback_bundle_refused - Failed: ...
1 failed, 2 warnings in 4.62s

EXIT:1

```

### M2.log

```text
F                                                                        [100%]
=================================== FAILURES ===================================
____________ test_import_inconsistent_bundle_refused_even_with_flag ____________

tmp_path = PosixPath('/private/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/pytest-of-administrator/pytest-127/test_import_inconsistent_bundl0')

    def test_import_inconsistent_bundle_refused_even_with_flag(tmp_path: Path) -> None:
        from csk_registry.bundle import export_bundle, import_bundle
    
        upstream = Store(tmp_path / "up.db")
        up_key = signing.generate_key()
        upstream.append(
            up_key.sign_record(_body("audited", name="skill-a")),
            created_at="2026-07-07T00:00:00Z",
        )
        bundle_a = export_bundle(upstream, up_key)
        upstream.close()
        fork = Store(tmp_path / "fork.db")
        fork.append(
            up_key.sign_record(
                _body("audited", name="skill-fork", commit="2" * 40, content_sha256="sha256:" + "3f" * 32)
            ),
            created_at="2026-07-07T00:00:00Z",
        )
        bundle_fork = export_bundle(fork, up_key)
        fork.close()
        assert bundle_a["snapshot"]["version"] == bundle_fork["snapshot"]["version"] == 1
        assert bundle_a["snapshot"]["head"] != bundle_fork["snapshot"]["head"]
    
        downstream = Store(tmp_path / "down.db")
        down_key = signing.generate_key()
        assert import_bundle(downstream, down_key, bundle_a, upstream_public_key=up_key.public_pinned) == 1
        for accept_older in (False, True):
>           with pytest.raises(ValueError, match="import_upstream_inconsistent"):
                 ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
E           AssertionError: Regex pattern did not match.
E             Expected regex: 'import_upstream_inconsistent'
E             Actual message: 'upstream high-water advanced during import'

tests/test_registry.py:2952: AssertionError
------------------------------ Captured log call -------------------------------
WARNING  csk_registry.audit:bundle.py:255 {"diagnostic":"import_upstream_inconsistent","event":"import_bundle","key_id":"d3013201941ce0fb","offered":{"head":"90f9a7543eee8bb340fb17f34f4565ee45ebbe0b1b49ed3766b285dccc2fdcf5","log_size":1,"merkle_root":"90f9a7543eee8bb340fb17f34f4565ee45ebbe0b1b49ed3766b285dccc2fdcf5","version":1},"persisted":{"head":"1353a8580eb17092a2ead7c181b759840716065a17df9cb6fe59b5a5fbb43830","log_size":1,"merkle_root":"1353a8580eb17092a2ead7c181b759840716065a17df9cb6fe59b5a5fbb43830","version":1},"result":"refused"}
=============================== warnings summary ===============================
../../../../tmp/p4-review/venv/lib/python3.14/site-packages/fastapi/testclient.py:1
  /tmp/p4-review/venv/lib/python3.14/site-packages/fastapi/testclient.py:1: StarletteDeprecationWarning: Using `httpx` with `starlette.testclient` is deprecated; install `httpx2` instead.
    from starlette.testclient import TestClient as TestClient  # noqa

../../../../tmp/p4-review/venv/lib/python3.14/site-packages/starlette/testclient.py:53
  /tmp/p4-review/venv/lib/python3.14/site-packages/starlette/testclient.py:53: DeprecationWarning: The anyio.abc.BlockingPortal alias is deprecated, use anyio.from_thread.BlockingPortal instead.
    _PortalFactoryType = Callable[[], AbstractContextManager[anyio.abc.BlockingPortal]]

-- Docs: https://docs.pytest.org/en/stable/how-to/capture-warnings.html
=========================== short test summary info ============================
FAILED tests/test_registry.py::test_import_inconsistent_bundle_refused_even_with_flag
1 failed, 2 warnings in 1.25s

EXIT:1

```
