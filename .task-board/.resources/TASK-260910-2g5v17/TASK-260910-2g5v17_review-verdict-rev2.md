# TASK-260910-2g5v17 review verdict — revision 2

Verdict: **accepted**. F1 resolved; route via accept_cr revision=2 to integrating, not done.

Candidate tree `1dfd8513e036b3e2006dd6fc620e396ad2638ffb`, base `c7ef32c75cc8dfda1abe38647af282a03175e8d3`. Published patch SHA256 verified `ee3ca475c779a1a81faef2794a36017e214008fbb48bbb88aff493fa5196042a`. All eight changed files in the live candidate match the archived tree byte-for-byte. Review/testing uses `/tmp/p4-review2/candidate`; no candidate writes or caches. Original git status preserved.

Read campaign rules, producer brief/results, both published patches, revision-1 verdict/probes, revision-2 hosted validation, audit P4, and protocol §5/service profile at actual CI pin `47c3c8cbd5da5e3fd6d99b0327d382c6a36494fe` (`/tmp/spec-47c3c8c`, HEAD independently checked). Profile restore codes are not import codes; requested import diagnostic spellings are appropriate.

## Review mapping

| Requirement | Independent evidence / outcome |
|---|---|
| F1 authoritative writer comparison | PASS store.py:1360,1386-1407 uses BEGIN IMMEDIATE (703-712), then reads high-water, applies override solely to rollback, retains authoritative before-image in result/conflict. bundle.py:178-217 logs the deciding comparison. No advisory read remains. |
| Competing writers | PASS committed deterministic tests at tests/test_registry.py:3328,3373,3421,3488 replay real second-connection imports before outer serialized write. Four schedules / five cases cover rollback, flagged rollback, inconsistent first import with/without flag, identical no-op. Assertions cover diagnostic, boundaries, audit outcome and persisted state. |
| Per-key durable storage | PASS store.py:57,158,675,1259: table in registry.db; schema migration and per-key accessor. tests:2881,3121,3208. |
| Atomic advancement | PASS store.py:1386-1438 appends records/ledger and high-water in one transaction. Independent BEFORE INSERT high-water trigger aborts after records are written: log, ledger and high-water all roll back. |
| Closed rollback / inconsistent refusals | PASS store.py:251-271,1393-1404; bundle.py:184-188; cli.py:271. Error text names key_id and all four persisted/offered fields. Tests:2899,2926 plus races. |
| Identical no-op | PASS store.py:1405-1407 suppresses high-water rewrite; existing fingerprint ledger suppresses record writes. Tests:2966,3488 use distinct frozen timestamps and compare row/head. |
| Newer / first / older override | PASS store.py:1395-1400,1423-1438; tests:2881,3002,3022,3158 and race:3373. Flag cannot override inconsistent; accepted older does not lower state. |
| Invalid/failed import | PASS bundle.py:129-158 validates entire bundle first; tests:3067,3095 plus late-failure probe below. |
| Storage protection / backup | PASS database independently observed 0600. store.py:2112,2119 protects database/sidecars. backup_to:2087 uses SQLite backup; test:3243 confirms high-water preserved. verify-backup deliberately does not compare upstream state: signed checkpoint has no upstream field, documented optional choice. |
| Docs | PASS README.md:104, SECURITY.md:38, CHANGELOG.md:36 P4. Existing dced9b8 historical reference preserved as explicitly required by rev2 scope; actual CI/tests use 47c3c8c. |
| Scope / architecture | PASS rev1→rev2 changes only bundle.py/store.py/test_registry.py/results. README, SECURITY, CHANGELOG and CLI byte-identical to rev1. No spec, deployment, pin or client changes. Store owns serialization/persistence; bundle layer owns signature checks and audit. |
| Validation / hygiene | Full pytest and strict mypy transcripts below. git diff --check exit 0; no separate linter configured. Candidate untouched. |

Coverage: 7/7 specified serial scenarios, 4/4 competing schedules (5 cases), 2/2 narrowing mutants caught, 1/1 late-failure probe. These are bounded tested cases, not exhaustive concurrency/corruption proof. No new blocking findings.

## Independent commands and setup

Shell `/bin/bash`, `set -o pipefail`; interpreter `/tmp/p4-review/venv/bin/python`, Python 3.14.6; cwd `/tmp/p4-review2/candidate`. Reused the existing external review venv while a fresh venv installation completed. Pytest's configured pythonpath=src selects this candidate; independent probe prints the exact production module path. Mypy checks this candidate's src directory under pyproject.toml strict=true.

Commands:
- `CURATOR_CONFORMANCE_ROOT=/tmp/spec-47c3c8c/conformance/v1 /tmp/p4-review/venv/bin/python -m pytest -q`
- `/tmp/p4-review/venv/bin/python -m mypy`
- `PYTHONPATH=/tmp/p4-review2/candidate/src /tmp/p4-review/venv/bin/python /tmp/p4-review2/probe.py`

Setup-only initial invocation with the still-installing fresh venv reported No module named pytest (exit 1); not candidate evidence. Used the established external venv for independent gates instead. No hosted gate rerun: accept attached rev2 validation log, GitHub run 35342977861, six OS/Python matrix jobs, mypy, distribution and Docker all success, gate exit 0. Producer results are contextual, not sole acceptance evidence.

## Narrowing mutants

Disposable `/tmp/p4-review2/mutant`, source restored after each mutation; actual import_bundle entry point through committed tests.

M1: `_upstream_offer_diagnostic`: `< current.version` → `< current.version - 1`. Narrows rejection to rollbacks of at least two versions. Serial rollback + competing v3/v2 rollback tests both fail DID NOT RAISE; exit 1. Caught 2/2 targeted tests.

M2: serialized override condition accepts diagnostics in `(IMPORT_UPSTREAM_ROLLBACK, IMPORT_UPSTREAM_INCONSISTENT)` when flag is set. Narrows inconsistent rejection to unflagged calls. Serial flag test and competing first-import flagged case fail DID NOT RAISE; unflagged case passes; exit 1. Caught 2/3 targeted cases, with the third intentionally unaffected.

## Transcripts and probe source

Independent full suite: **183 passed, 2 dependency deprecation warnings, exit 0** (193.98s). Strict mypy: **14 source files, exit 0**. Producer-reported startup-concurrency flake did not reproduce in this full run; no broader claim about flake frequency.

### pytest.log

```text
........................................................................ [ 39%]
........................................................................ [ 78%]
.......................................                                  [100%]
=============================== warnings summary ===============================
../../../../tmp/p4-review/venv/lib/python3.14/site-packages/fastapi/testclient.py:1
  /tmp/p4-review/venv/lib/python3.14/site-packages/fastapi/testclient.py:1: StarletteDeprecationWarning: Using `httpx` with `starlette.testclient` is deprecated; install `httpx2` instead.
    from starlette.testclient import TestClient as TestClient  # noqa

../../../../tmp/p4-review/venv/lib/python3.14/site-packages/starlette/testclient.py:53
  /tmp/p4-review/venv/lib/python3.14/site-packages/starlette/testclient.py:53: DeprecationWarning: The anyio.abc.BlockingPortal alias is deprecated, use anyio.from_thread.BlockingPortal instead.
    _PortalFactoryType = Callable[[], AbstractContextManager[anyio.abc.BlockingPortal]]

-- Docs: https://docs.pytest.org/en/stable/how-to/capture-warnings.html
183 passed, 2 warnings in 193.98s (0:03:13)

```

### mypy.log

```text
Success: no issues found in 14 source files

```

### probe.py

```text
import importlib.util, tempfile
from pathlib import Path
from csk_registry.bundle import import_bundle
from csk_registry.store import Store
import csk_registry.store
print("Production module:", csk_registry.store.__file__)
spec = importlib.util.spec_from_file_location("test_registry", "/tmp/p4-review2/candidate/tests/test_registry.py")
t = importlib.util.module_from_spec(spec); spec.loader.exec_module(t)
with tempfile.TemporaryDirectory() as td:
 p=Path(td); uk,dk,v1,v2=t._upstream_pair(p); d=Store(p/"d.db")
 import_bundle(d,dk,v1,upstream_public_key=uk.public_pinned)
 before=d.get_upstream_high_water(uk.key_id)
 d._conn.execute("CREATE TRIGGER fail_hw BEFORE INSERT ON upstream_high_water BEGIN SELECT RAISE(ABORT, 'after records before high-water'); END")
 try: import_bundle(d,dk,v2,upstream_public_key=uk.public_pinned)
 except Exception as e: print("Injected failure:",e)
 else: raise AssertionError("failure injection not reached")
 assert d.head()[0]==1 and d.get_upstream_high_water(uk.key_id)==before
 assert d._conn.execute("SELECT count(*) FROM imported_records").fetchone()[0]==1
 print("PASS: log, ledger and high-water rolled back; mode", oct((p/"d.db").stat().st_mode & 0o777))
 d.close()

```

### probe.log

```text
/tmp/p4-review/venv/lib/python3.14/site-packages/fastapi/testclient.py:1: StarletteDeprecationWarning: Using `httpx` with `starlette.testclient` is deprecated; install `httpx2` instead.
  from starlette.testclient import TestClient as TestClient  # noqa
Production module: /tmp/p4-review2/candidate/src/csk_registry/store.py
Injected failure: after records before high-water
PASS: log, ledger and high-water rolled back; mode 0o600

```

### M1.log

```text
COMMAND: /tmp/p4-review/venv/bin/python -m pytest -q tests/test_registry.py::test_import_rollback_bundle_refused tests/test_registry.py::test_concurrent_newer_first_refuses_outer_with_rollback
FF                                                                       [100%]
=================================== FAILURES ===================================
_____________________ test_import_rollback_bundle_refused ______________________

tmp_path = PosixPath('/private/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/pytest-of-administrator/pytest-161/test_import_rollback_bundle_re0')

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

tests/test_registry.py:2909: Failed
___________ test_concurrent_newer_first_refuses_outer_with_rollback ____________

tmp_path = PosixPath('/private/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/pytest-of-administrator/pytest-161/test_concurrent_newer_first_re0')
monkeypatch = <_pytest.monkeypatch.MonkeyPatch object at 0x111562190>
caplog = <_pytest.logging.LogCaptureFixture object at 0x11174b770>

    def test_concurrent_newer_first_refuses_outer_with_rollback(
        tmp_path: Path, monkeypatch: pytest.MonkeyPatch, caplog: pytest.LogCaptureFixture
    ) -> None:
        from csk_registry.bundle import import_bundle
    
        up_key, down_key, bundle_v1, bundle_v2, bundle_v3 = _upstream_triple(tmp_path)
        downstream = Store(tmp_path / "down.db")
        competitor = Store(tmp_path / "down.db")
        try:
            assert (
                import_bundle(downstream, down_key, bundle_v1, upstream_public_key=up_key.public_pinned)  # type: ignore[arg-type,union-attr]
                == 1
            )
            competitor_state: dict[str, object] = {}
    
            def commit_v3_first() -> None:
                assert (
                    import_bundle(competitor, down_key, bundle_v3, upstream_public_key=up_key.public_pinned)  # type: ignore[arg-type,union-attr]
                    == 2
                )
                competitor_state["high_water"] = competitor.get_upstream_high_water(up_key.key_id)  # type: ignore[union-attr]
                competitor_state["head"] = competitor.head()
    
            _pause_outer_import_at_write(monkeypatch, downstream, commit_v3_first)
            with caplog.at_level(logging.INFO, logger="csk_registry.audit"):
>               with pytest.raises(ValueError, match="import_upstream_rollback") as exc_info:
                     ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
E               Failed: DID NOT RAISE ValueError

tests/test_registry.py:3353: Failed
------------------------------ Captured log call -------------------------------
INFO     csk_registry.audit:bundle.py:301 {"event":"import_bundle","imported":2,"key_id":"101cbd59edb417fc","offered":{"head":"6e2c5c4e3a8d668f9ff60b0674f16bd392c2560ae02ec6b9ff6f859bd76388b1","log_size":3,"merkle_root":"df4165ce7451e86b7b018257574f25821a8c4ad8c6bfe290651ad6b8d915f228","version":3},"persisted":{"head":"1353a8580eb17092a2ead7c181b759840716065a17df9cb6fe59b5a5fbb43830","log_size":1,"merkle_root":"1353a8580eb17092a2ead7c181b759840716065a17df9cb6fe59b5a5fbb43830","version":1},"result":"ok"}
INFO     csk_registry.audit:bundle.py:301 {"event":"import_bundle","imported":0,"key_id":"101cbd59edb417fc","offered":{"head":"f9e9bcbd12d09c733b302ad44779d2eb267bc6d80736cb313876d7c46d846247","log_size":2,"merkle_root":"c7d047f604a6c52206a9320130827849d44424dfb486d40375b146eae692a906","version":2},"persisted":{"head":"6e2c5c4e3a8d668f9ff60b0674f16bd392c2560ae02ec6b9ff6f859bd76388b1","log_size":3,"merkle_root":"df4165ce7451e86b7b018257574f25821a8c4ad8c6bfe290651ad6b8d915f228","version":3},"result":"ok"}
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
FAILED tests/test_registry.py::test_concurrent_newer_first_refuses_outer_with_rollback
2 failed, 2 warnings in 9.51s

EXIT:1

```

### M2.log

```text
COMMAND: /tmp/p4-review/venv/bin/python -m pytest -q tests/test_registry.py::test_import_inconsistent_bundle_refused_even_with_flag tests/test_registry.py::test_concurrent_first_imports_with_different_bodies_refused
F.F                                                                      [100%]
=================================== FAILURES ===================================
____________ test_import_inconsistent_bundle_refused_even_with_flag ____________

tmp_path = PosixPath('/private/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/pytest-of-administrator/pytest-162/test_import_inconsistent_bundl0')

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
E           Failed: DID NOT RAISE ValueError

tests/test_registry.py:2953: Failed
------------------------------ Captured log call -------------------------------
WARNING  csk_registry.audit:bundle.py:271 {"diagnostic":"import_upstream_inconsistent","event":"import_bundle","key_id":"702d6583a300e8b1","offered":{"head":"90f9a7543eee8bb340fb17f34f4565ee45ebbe0b1b49ed3766b285dccc2fdcf5","log_size":1,"merkle_root":"90f9a7543eee8bb340fb17f34f4565ee45ebbe0b1b49ed3766b285dccc2fdcf5","version":1},"persisted":{"head":"1353a8580eb17092a2ead7c181b759840716065a17df9cb6fe59b5a5fbb43830","log_size":1,"merkle_root":"1353a8580eb17092a2ead7c181b759840716065a17df9cb6fe59b5a5fbb43830","version":1},"result":"refused"}
WARNING  csk_registry.audit:bundle.py:299 {"accepted_older":true,"event":"import_bundle","imported":1,"key_id":"702d6583a300e8b1","offered":{"head":"90f9a7543eee8bb340fb17f34f4565ee45ebbe0b1b49ed3766b285dccc2fdcf5","log_size":1,"merkle_root":"90f9a7543eee8bb340fb17f34f4565ee45ebbe0b1b49ed3766b285dccc2fdcf5","version":1},"persisted":{"head":"1353a8580eb17092a2ead7c181b759840716065a17df9cb6fe59b5a5fbb43830","log_size":1,"merkle_root":"1353a8580eb17092a2ead7c181b759840716065a17df9cb6fe59b5a5fbb43830","version":1},"result":"ok","warning":"import_upstream_rollback"}
______ test_concurrent_first_imports_with_different_bodies_refused[True] _______

tmp_path = PosixPath('/private/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/pytest-of-administrator/pytest-162/test_concurrent_first_imports_1')
monkeypatch = <_pytest.monkeypatch.MonkeyPatch object at 0x110baa780>
caplog = <_pytest.logging.LogCaptureFixture object at 0x110c58cd0>
accept_older = True

    @pytest.mark.parametrize("accept_older", [False, True])
    def test_concurrent_first_imports_with_different_bodies_refused(
        tmp_path: Path,
        monkeypatch: pytest.MonkeyPatch,
        caplog: pytest.LogCaptureFixture,
        accept_older: bool,
    ) -> None:
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
        competitor = Store(tmp_path / "down.db")
        try:
            down_key = signing.generate_key()
            competitor_state: dict[str, object] = {}
    
            def commit_fork_first() -> None:
                assert import_bundle(competitor, down_key, bundle_fork, upstream_public_key=up_key.public_pinned) == 1
                competitor_state["high_water"] = competitor.get_upstream_high_water(up_key.key_id)
                competitor_state["head"] = competitor.head()
    
            _pause_outer_import_at_write(monkeypatch, downstream, commit_fork_first)
            with caplog.at_level(logging.INFO, logger="csk_registry.audit"):
>               with pytest.raises(ValueError, match="import_upstream_inconsistent") as exc_info:
                     ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
E               Failed: DID NOT RAISE ValueError

tests/test_registry.py:3462: Failed
------------------------------ Captured log call -------------------------------
INFO     csk_registry.audit:bundle.py:301 {"event":"import_bundle","imported":1,"key_id":"66a88215c0a88de8","offered":{"head":"90f9a7543eee8bb340fb17f34f4565ee45ebbe0b1b49ed3766b285dccc2fdcf5","log_size":1,"merkle_root":"90f9a7543eee8bb340fb17f34f4565ee45ebbe0b1b49ed3766b285dccc2fdcf5","version":1},"persisted":null,"result":"ok"}
WARNING  csk_registry.audit:bundle.py:299 {"accepted_older":true,"event":"import_bundle","imported":1,"key_id":"66a88215c0a88de8","offered":{"head":"1353a8580eb17092a2ead7c181b759840716065a17df9cb6fe59b5a5fbb43830","log_size":1,"merkle_root":"1353a8580eb17092a2ead7c181b759840716065a17df9cb6fe59b5a5fbb43830","version":1},"persisted":{"head":"90f9a7543eee8bb340fb17f34f4565ee45ebbe0b1b49ed3766b285dccc2fdcf5","log_size":1,"merkle_root":"90f9a7543eee8bb340fb17f34f4565ee45ebbe0b1b49ed3766b285dccc2fdcf5","version":1},"result":"ok","warning":"import_upstream_rollback"}
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
FAILED tests/test_registry.py::test_concurrent_first_imports_with_different_bodies_refused[True]
2 failed, 1 passed, 2 warnings in 15.65s

EXIT:1

```

### hosted.log

```text
$ sh scripts/remote-gate.sh
remote gate: attempt 1/3: pushing 64a2423200f1239552499dd02adaca6eb3a284e6 as gate/STORY-260910-stz5f0/260918-120653-3482-1
remote: 
remote: Create a pull request for 'gate/STORY-260910-stz5f0/260918-120653-3482-1' on GitHub by visiting:        
remote:      https://github.com/relux-works/curator-skill-registry/pull/new/gate/STORY-260910-stz5f0/260918-120653-3482-1        
remote: 
remote gate: run 35342977861 (https://github.com/relux-works/curator-skill-registry/actions/runs/35342977861)
remote gate: run 35342977861 finished: success
  Type check / mypy strict: success  
  Docker build: success  
  Tests / Python 3.11 on macos-latest: success  
  Tests / Python 3.11 on windows-latest: success  
  Tests / Python 3.14 on ubuntu-latest: success  
  Tests / Python 3.11 on ubuntu-latest: success  
  Tests / Python 3.14 on macos-latest: success  
  Tests / Python 3.14 on windows-latest: success  
  Build distribution: success  
[exit 0]
coverage_unit=exact_command_shard required=1 green=1 failed=0 missing=0 test_case_coverage=unknown

```
