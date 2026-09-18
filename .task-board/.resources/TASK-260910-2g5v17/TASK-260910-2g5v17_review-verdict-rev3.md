# TASK-260910-2g5v17 — revision 3 exact-tree review

Verdict: **accepted**. Route CR revision 3 via accept_cr to integrating; reviewer does not close delivery.

Review scope: mechanical rebase of accepted P4 revision 2 over R6, faithful documentation union, removal of task artifact. Candidate unchanged by reviewer; all execution and mutations under /tmp.

## Per-item evidence

| Requirement | Result |
|---|---|
| Exact candidate | Temporary-index snapshot: `9b7d33115bd3ffe44c34c5a340de72aaab06d0cb`. Equals landed commit bf5cac1200cfa39dff0a6b0449072ff5b22f124d tree and freshly fetched origin/main tree. |
| Signature | Good git signature for bot@relux.works, ED25519 SHA256:qbALzjdB9BRgYJjDkX/p9EAPLEofB2AbskJc6Ftwhds. |
| Published patch | SHA256 5935295223f88063a4fa175cb3792f0f3d24478dd6ba5bff0b3ddc4388756431; independently applied to c7ef32c using temporary index, resulting in exact candidate tree. Export review/update SHA256 both verified. |
| Delta paths | Exactly 10 expected paths against prior tree: nine R6 paths plus deleted TASK-260910-2g5v17_results.md. No others. |
| Faithful merge | CHANGELOG, README, SECURITY, cli.py: all added/deleted lines in base→P4 exactly equal R6→landed; reciprocal base→R6 equals P4→landed. This checks both contributions, ignoring only hunk offsets/context. 4/4 merged files match; both changelog entries preserved verbatim. |
| R6 unchanged files | compose.yaml, __init__.py, keys.py, signing.py, test_key_passphrase.py all byte-identical to fb86420 (5/5). |
| P4 unchanged files | bundle.py, store.py, test_registry.py byte-identical to accepted rev2 (3/3). CLI union inspected: P4 import audit/flag remains; R6 KeyPassphraseError handling surrounds command dispatch. |
| Hygiene | Candidate git status contains only 12 expected source/test/doc/packaging paths. No results/logbook/build/coverage files in candidate tree or ignored-file scan. Pre-existing ignored mypy/pytest/bytecode caches observed, no reviewer additions. git diff --check exit 0. |
| Hosted gate | Read rev3-validation.log: run 35347197207, all six Python/OS jobs plus mypy, distribution and Docker successful, exit 0. Accepted attached evidence; did not invoke hosted gate. |

## Independent execution

Shell zsh, `set -o pipefail`, cwd `/tmp/p4-review3-copy`, detached at signed landed commit. `/tmp/csk-venv/bin/python` 3.14.6; pytest pythonpath=src selects disposable candidate. Strict mypy configured by pyproject.toml. Conformance worktree independently resolves to `47c3c8cbd5da5e3fd6d99b0327d382c6a36494fe`.

Commands:
```
CURATOR_CONFORMANCE_ROOT=/tmp/spec-47c3c8c/conformance/v1 /tmp/csk-venv/bin/python -m pytest -q
/tmp/csk-venv/bin/python -m mypy
```
Setup note: first invocation using newly created /tmp/p4-review3-venv preceded dependency installation and returned No module named pytest (exit 1). Used established external venv for actual independent validation. No candidate test result inferred from setup failure.

## Adversarial checks

Actual import_bundle production path exercised through committed tests in separate `/tmp/p4-review3-mutant`; source restored from saved bytes after each mutation.

- M1 narrows rollback rejection from `< current.version` to `< current.version - 1`: serial rollback and competing v3/v2 both fail DID NOT RAISE (2/2, exit 1).
- M2 narrows inconsistent rejection to unflagged imports by allowing the override for both diagnostic codes: serial inconsistent and flagged competing-first-import fail DID NOT RAISE; unflagged case remains green (2 expected failures/3 cases, exit 1).

2/2 narrowing mutants caught. Rev2 atomicity trigger and storage/backup probes accepted as historical evidence for byte-identical P4 modules, not claimed rerun in rev3. Full validation is independently rerun on the combined tree; prior review is not treated as validation of this changed tree.

## Validation decision and limits

First full suite: 205 passed, 1 failed, exit 1 (249.44s). Failure was the existing concurrent-store writer test at BEGIN IMMEDIATE with database is locked. The unchanged test passed alone (3.66s, exit 0). One full diagnostic rerun with other reviewer work finished passed **206/206 tests**, 2 dependency deprecation warnings, exit 0 (189.04s). Strict mypy: 14 files, exit 0. The first failure is retained below; load-related timing is an inference, not a proven root cause or claim that the test cannot flake. This is non-blocking for the exact mechanical rebase: no change to the failing test or serialized transaction entry; fresh full suite and hosted matrix are green.

No merge or scope discrepancies. Revision-3 review requirements satisfied 4/4 (identity, accounting, independent validation, hygiene); 2/2 targeted mutants caught. This is not an exhaustive concurrency proof. Prior review's detailed atomicity/storage probes were read, not rerun.

Final temporary-index snapshot again equals 9b7d33115bd3ffe44c34c5a340de72aaab06d0cb; initial/final git status files compare identical. Run goal queried: Active Goal none (run not goal-bound); no directives.

## Transcripts

### identity.log

```text
bf5cac1200cfa39dff0a6b0449072ff5b22f124d
9b7d33115bd3ffe44c34c5a340de72aaab06d0cb
9b7d33115bd3ffe44c34c5a340de72aaab06d0cb
Good "git" signature for bot@relux.works with ED25519 key SHA256:qbALzjdB9BRgYJjDkX/p9EAPLEofB2AbskJc6Ftwhds
M	CHANGELOG.md
M	README.md
M	SECURITY.md
D	TASK-260910-2g5v17_results.md
M	compose.yaml
M	src/csk_registry/__init__.py
M	src/csk_registry/cli.py
M	src/csk_registry/keys.py
M	src/csk_registry/signing.py
A	tests/test_key_passphrase.py

```

### delta.log

```text
CHANGELOG.md: P4 added/deleted lines identical (100%); hunk offsets/context excluded
README.md: P4 added/deleted lines identical (100%); hunk offsets/context excluded
SECURITY.md: P4 added/deleted lines identical (100%); hunk offsets/context excluded
src/csk_registry/cli.py: P4 added/deleted lines identical (100%); hunk offsets/context excluded
compose.yaml: R6 byte-identical
src/csk_registry/__init__.py: R6 byte-identical
src/csk_registry/keys.py: R6 byte-identical
src/csk_registry/signing.py: R6 byte-identical
tests/test_key_passphrase.py: R6 byte-identical
src/csk_registry/bundle.py: accepted P4 byte-identical
src/csk_registry/store.py: accepted P4 byte-identical
tests/test_registry.py: accepted P4 byte-identical

```

### status-final.log

```text
 M CHANGELOG.md
 M README.md
 M SECURITY.md
 M compose.yaml
 M src/csk_registry/__init__.py
 M src/csk_registry/bundle.py
 M src/csk_registry/cli.py
 M src/csk_registry/keys.py
 M src/csk_registry/signing.py
 M src/csk_registry/store.py
 M tests/test_registry.py
?? tests/test_key_passphrase.py

```

### pytest.log

```text
........................................................................ [ 34%]
...................................................................F.... [ 69%]
..............................................................           [100%]
=================================== FAILURES ===================================
______________ test_concurrent_store_instances_serialize_writers _______________

tmp_path = PosixPath('/private/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/pytest-of-administrator/pytest-177/test_concurrent_store_instance0')

    def test_concurrent_store_instances_serialize_writers(tmp_path: Path):
        path = tmp_path / "r.db"
        Store(path).close()
        key = signing.generate_key()
        records = [
            key.sign_record(_body(name=f"skill-{index}", commit=f"{index:040d}"))
            for index in range(32)
        ]
    
        def append(index: int) -> int:
            store = Store(path)
            try:
                return store.append(
                    records[index],
                    created_at="2026-07-07T00:00:00Z",
                ).seq
            finally:
                store.close()
    
        with ThreadPoolExecutor(max_workers=8) as executor:
>           sequences = list(executor.map(append, range(len(records))))
                        ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

tests/test_registry.py:574: 
_ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ 
/Users/administrator/.local/share/uv/python/cpython-3.14.6-macos-x86_64-none/lib/python3.14/concurrent/futures/_base.py:645: in result_iterator
    yield _result_or_cancel(fs.pop())
          ^^^^^^^^^^^^^^^^^^^^^^^^^^^
/Users/administrator/.local/share/uv/python/cpython-3.14.6-macos-x86_64-none/lib/python3.14/concurrent/futures/_base.py:312: in _result_or_cancel
    return fut.result(timeout)
           ^^^^^^^^^^^^^^^^^^^
/Users/administrator/.local/share/uv/python/cpython-3.14.6-macos-x86_64-none/lib/python3.14/concurrent/futures/_base.py:454: in result
    return self.__get_result()
           ^^^^^^^^^^^^^^^^^^^
/Users/administrator/.local/share/uv/python/cpython-3.14.6-macos-x86_64-none/lib/python3.14/concurrent/futures/_base.py:396: in __get_result
    raise self._exception
/Users/administrator/.local/share/uv/python/cpython-3.14.6-macos-x86_64-none/lib/python3.14/concurrent/futures/thread.py:86: in run
    result = ctx.run(self.task)
             ^^^^^^^^^^^^^^^^^^
/Users/administrator/.local/share/uv/python/cpython-3.14.6-macos-x86_64-none/lib/python3.14/concurrent/futures/thread.py:73: in run
    return fn(*args, **kwargs)
           ^^^^^^^^^^^^^^^^^^^
tests/test_registry.py:566: in append
    return store.append(
src/csk_registry/store.py:870: in append
    with self._write_transaction():
         ^^^^^^^^^^^^^^^^^^^^^^^^^
/Users/administrator/.local/share/uv/python/cpython-3.14.6-macos-x86_64-none/lib/python3.14/contextlib.py:141: in __enter__
    return next(self.gen)
           ^^^^^^^^^^^^^^
_ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ 

self = <csk_registry.store.Store object at 0x106a37790>

    @contextmanager
    def _write_transaction(self) -> Iterator[None]:
>       self._conn.execute("BEGIN IMMEDIATE")
E       sqlite3.OperationalError: database is locked

src/csk_registry/store.py:704: OperationalError
=============================== warnings summary ===============================
../../../tmp/csk-venv/lib/python3.14/site-packages/fastapi/testclient.py:1
  /tmp/csk-venv/lib/python3.14/site-packages/fastapi/testclient.py:1: StarletteDeprecationWarning: Using `httpx` with `starlette.testclient` is deprecated; install `httpx2` instead.
    from starlette.testclient import TestClient as TestClient  # noqa

../../../tmp/csk-venv/lib/python3.14/site-packages/starlette/testclient.py:53
  /tmp/csk-venv/lib/python3.14/site-packages/starlette/testclient.py:53: DeprecationWarning: The anyio.abc.BlockingPortal alias is deprecated, use anyio.from_thread.BlockingPortal instead.
    _PortalFactoryType = Callable[[], AbstractContextManager[anyio.abc.BlockingPortal]]

-- Docs: https://docs.pytest.org/en/stable/how-to/capture-warnings.html
=========================== short test summary info ============================
FAILED tests/test_registry.py::test_concurrent_store_instances_serialize_writers
1 failed, 205 passed, 2 warnings in 249.44s (0:04:09)

```

### retry.log

```text
.                                                                        [100%]
=============================== warnings summary ===============================
../../../tmp/csk-venv/lib/python3.14/site-packages/fastapi/testclient.py:1
  /tmp/csk-venv/lib/python3.14/site-packages/fastapi/testclient.py:1: StarletteDeprecationWarning: Using `httpx` with `starlette.testclient` is deprecated; install `httpx2` instead.
    from starlette.testclient import TestClient as TestClient  # noqa

../../../tmp/csk-venv/lib/python3.14/site-packages/starlette/testclient.py:53
  /tmp/csk-venv/lib/python3.14/site-packages/starlette/testclient.py:53: DeprecationWarning: The anyio.abc.BlockingPortal alias is deprecated, use anyio.from_thread.BlockingPortal instead.
    _PortalFactoryType = Callable[[], AbstractContextManager[anyio.abc.BlockingPortal]]

-- Docs: https://docs.pytest.org/en/stable/how-to/capture-warnings.html
1 passed, 2 warnings in 3.66s

```

### pytest-rerun.log

```text
........................................................................ [ 34%]
........................................................................ [ 69%]
..............................................................           [100%]
=============================== warnings summary ===============================
../../../tmp/csk-venv/lib/python3.14/site-packages/fastapi/testclient.py:1
  /tmp/csk-venv/lib/python3.14/site-packages/fastapi/testclient.py:1: StarletteDeprecationWarning: Using `httpx` with `starlette.testclient` is deprecated; install `httpx2` instead.
    from starlette.testclient import TestClient as TestClient  # noqa

../../../tmp/csk-venv/lib/python3.14/site-packages/starlette/testclient.py:53
  /tmp/csk-venv/lib/python3.14/site-packages/starlette/testclient.py:53: DeprecationWarning: The anyio.abc.BlockingPortal alias is deprecated, use anyio.from_thread.BlockingPortal instead.
    _PortalFactoryType = Callable[[], AbstractContextManager[anyio.abc.BlockingPortal]]

-- Docs: https://docs.pytest.org/en/stable/how-to/capture-warnings.html
206 passed, 2 warnings in 189.04s (0:03:09)

```

### mypy.log

```text
Success: no issues found in 14 source files

```

### M1.log

```text
COMMAND: /tmp/csk-venv/bin/python -m pytest -q tests/test_registry.py::test_import_rollback_bundle_refused tests/test_registry.py::test_concurrent_newer_first_refuses_outer_with_rollback
FF                                                                       [100%]
=================================== FAILURES ===================================
_____________________ test_import_rollback_bundle_refused ______________________

tmp_path = PosixPath('/private/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/pytest-of-administrator/pytest-178/test_import_rollback_bundle_re0')

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

tmp_path = PosixPath('/private/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/pytest-of-administrator/pytest-178/test_concurrent_newer_first_re0')
monkeypatch = <_pytest.monkeypatch.MonkeyPatch object at 0x1076122c0>
caplog = <_pytest.logging.LogCaptureFixture object at 0x106d33770>

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
INFO     csk_registry.audit:bundle.py:301 {"event":"import_bundle","imported":2,"key_id":"f13c4a46901e428e","offered":{"head":"6e2c5c4e3a8d668f9ff60b0674f16bd392c2560ae02ec6b9ff6f859bd76388b1","log_size":3,"merkle_root":"df4165ce7451e86b7b018257574f25821a8c4ad8c6bfe290651ad6b8d915f228","version":3},"persisted":{"head":"1353a8580eb17092a2ead7c181b759840716065a17df9cb6fe59b5a5fbb43830","log_size":1,"merkle_root":"1353a8580eb17092a2ead7c181b759840716065a17df9cb6fe59b5a5fbb43830","version":1},"result":"ok"}
INFO     csk_registry.audit:bundle.py:301 {"event":"import_bundle","imported":0,"key_id":"f13c4a46901e428e","offered":{"head":"f9e9bcbd12d09c733b302ad44779d2eb267bc6d80736cb313876d7c46d846247","log_size":2,"merkle_root":"c7d047f604a6c52206a9320130827849d44424dfb486d40375b146eae692a906","version":2},"persisted":{"head":"6e2c5c4e3a8d668f9ff60b0674f16bd392c2560ae02ec6b9ff6f859bd76388b1","log_size":3,"merkle_root":"df4165ce7451e86b7b018257574f25821a8c4ad8c6bfe290651ad6b8d915f228","version":3},"result":"ok"}
=============================== warnings summary ===============================
../../../tmp/csk-venv/lib/python3.14/site-packages/fastapi/testclient.py:1
  /tmp/csk-venv/lib/python3.14/site-packages/fastapi/testclient.py:1: StarletteDeprecationWarning: Using `httpx` with `starlette.testclient` is deprecated; install `httpx2` instead.
    from starlette.testclient import TestClient as TestClient  # noqa

../../../tmp/csk-venv/lib/python3.14/site-packages/starlette/testclient.py:53
  /tmp/csk-venv/lib/python3.14/site-packages/starlette/testclient.py:53: DeprecationWarning: The anyio.abc.BlockingPortal alias is deprecated, use anyio.from_thread.BlockingPortal instead.
    _PortalFactoryType = Callable[[], AbstractContextManager[anyio.abc.BlockingPortal]]

-- Docs: https://docs.pytest.org/en/stable/how-to/capture-warnings.html
=========================== short test summary info ============================
FAILED tests/test_registry.py::test_import_rollback_bundle_refused - Failed: ...
FAILED tests/test_registry.py::test_concurrent_newer_first_refuses_outer_with_rollback
2 failed, 2 warnings in 13.34s

EXIT:1

```

### M2.log

```text
COMMAND: /tmp/csk-venv/bin/python -m pytest -q tests/test_registry.py::test_import_inconsistent_bundle_refused_even_with_flag tests/test_registry.py::test_concurrent_first_imports_with_different_bodies_refused
F.F                                                                      [100%]
=================================== FAILURES ===================================
____________ test_import_inconsistent_bundle_refused_even_with_flag ____________

tmp_path = PosixPath('/private/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/pytest-of-administrator/pytest-179/test_import_inconsistent_bundl0')

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
WARNING  csk_registry.audit:bundle.py:271 {"diagnostic":"import_upstream_inconsistent","event":"import_bundle","key_id":"968df12887edb45e","offered":{"head":"90f9a7543eee8bb340fb17f34f4565ee45ebbe0b1b49ed3766b285dccc2fdcf5","log_size":1,"merkle_root":"90f9a7543eee8bb340fb17f34f4565ee45ebbe0b1b49ed3766b285dccc2fdcf5","version":1},"persisted":{"head":"1353a8580eb17092a2ead7c181b759840716065a17df9cb6fe59b5a5fbb43830","log_size":1,"merkle_root":"1353a8580eb17092a2ead7c181b759840716065a17df9cb6fe59b5a5fbb43830","version":1},"result":"refused"}
WARNING  csk_registry.audit:bundle.py:299 {"accepted_older":true,"event":"import_bundle","imported":1,"key_id":"968df12887edb45e","offered":{"head":"90f9a7543eee8bb340fb17f34f4565ee45ebbe0b1b49ed3766b285dccc2fdcf5","log_size":1,"merkle_root":"90f9a7543eee8bb340fb17f34f4565ee45ebbe0b1b49ed3766b285dccc2fdcf5","version":1},"persisted":{"head":"1353a8580eb17092a2ead7c181b759840716065a17df9cb6fe59b5a5fbb43830","log_size":1,"merkle_root":"1353a8580eb17092a2ead7c181b759840716065a17df9cb6fe59b5a5fbb43830","version":1},"result":"ok","warning":"import_upstream_rollback"}
______ test_concurrent_first_imports_with_different_bodies_refused[True] _______

tmp_path = PosixPath('/private/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/pytest-of-administrator/pytest-179/test_concurrent_first_imports_1')
monkeypatch = <_pytest.monkeypatch.MonkeyPatch object at 0x1069f2780>
caplog = <_pytest.logging.LogCaptureFixture object at 0x106aa0f50>
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
INFO     csk_registry.audit:bundle.py:301 {"event":"import_bundle","imported":1,"key_id":"472739d6012269fc","offered":{"head":"90f9a7543eee8bb340fb17f34f4565ee45ebbe0b1b49ed3766b285dccc2fdcf5","log_size":1,"merkle_root":"90f9a7543eee8bb340fb17f34f4565ee45ebbe0b1b49ed3766b285dccc2fdcf5","version":1},"persisted":null,"result":"ok"}
WARNING  csk_registry.audit:bundle.py:299 {"accepted_older":true,"event":"import_bundle","imported":1,"key_id":"472739d6012269fc","offered":{"head":"1353a8580eb17092a2ead7c181b759840716065a17df9cb6fe59b5a5fbb43830","log_size":1,"merkle_root":"1353a8580eb17092a2ead7c181b759840716065a17df9cb6fe59b5a5fbb43830","version":1},"persisted":{"head":"90f9a7543eee8bb340fb17f34f4565ee45ebbe0b1b49ed3766b285dccc2fdcf5","log_size":1,"merkle_root":"90f9a7543eee8bb340fb17f34f4565ee45ebbe0b1b49ed3766b285dccc2fdcf5","version":1},"result":"ok","warning":"import_upstream_rollback"}
=============================== warnings summary ===============================
../../../tmp/csk-venv/lib/python3.14/site-packages/fastapi/testclient.py:1
  /tmp/csk-venv/lib/python3.14/site-packages/fastapi/testclient.py:1: StarletteDeprecationWarning: Using `httpx` with `starlette.testclient` is deprecated; install `httpx2` instead.
    from starlette.testclient import TestClient as TestClient  # noqa

../../../tmp/csk-venv/lib/python3.14/site-packages/starlette/testclient.py:53
  /tmp/csk-venv/lib/python3.14/site-packages/starlette/testclient.py:53: DeprecationWarning: The anyio.abc.BlockingPortal alias is deprecated, use anyio.from_thread.BlockingPortal instead.
    _PortalFactoryType = Callable[[], AbstractContextManager[anyio.abc.BlockingPortal]]

-- Docs: https://docs.pytest.org/en/stable/how-to/capture-warnings.html
=========================== short test summary info ============================
FAILED tests/test_registry.py::test_import_inconsistent_bundle_refused_even_with_flag
FAILED tests/test_registry.py::test_concurrent_first_imports_with_different_bodies_refused[True]
2 failed, 1 passed, 2 warnings in 18.95s

EXIT:1

```

### hosted.log

```text
$ sh scripts/remote-gate.sh
remote gate: attempt 1/3: pushing 65af014b81768dd4b1ce181444f6047faa21adf8 as gate/STORY-260910-stz5f0/260918-125449-61412-1
remote: 
remote: Create a pull request for 'gate/STORY-260910-stz5f0/260918-125449-61412-1' on GitHub by visiting:        
remote:      https://github.com/relux-works/curator-skill-registry/pull/new/gate/STORY-260910-stz5f0/260918-125449-61412-1        
remote: 
remote gate: run 35347197207 (https://github.com/relux-works/curator-skill-registry/actions/runs/35347197207)
remote gate: run 35347197207 finished: success
  Type check / mypy strict: success  
  Tests / Python 3.11 on macos-latest: success  
  Tests / Python 3.11 on windows-latest: success  
  Tests / Python 3.14 on macos-latest: success  
  Tests / Python 3.14 on windows-latest: success  
  Tests / Python 3.14 on ubuntu-latest: success  
  Docker build: success  
  Tests / Python 3.11 on ubuntu-latest: success  
  Build distribution: success  
[exit 0]
coverage_unit=exact_command_shard required=1 green=1 failed=0 missing=0 test_case_coverage=unknown

```

### three-way.log

```text
git diff -U0 c7ef32c 1dfd8513e036b3e2006dd6fc620e396ad2638ffb -- CHANGELOG.md
diff --git a/CHANGELOG.md b/CHANGELOG.md
index 0cb6f68..c588ace 100644
--- a/CHANGELOG.md
+++ b/CHANGELOG.md
@@ -35,0 +36,16 @@
+- P4: `import-bundle` now persists a per-upstream high-water (`version`,
+  `log_size`, `head`, `merkle_root` per upstream `key_id`) in a new
+  `upstream_high_water` table (schema version 4, migrated once at startup)
+  and compares every verified bundle against it under the client §5
+  rollback rules. A version below refuses with `import_upstream_rollback`;
+  an equal version with a different `head`/`merkle_root`/`log_size` refuses
+  with `import_upstream_inconsistent`; an equal identical boundary is an
+  accepted no-op; a higher version imports and advances the stored boundary
+  in the same transaction as the imported records. Refusals exit non-zero
+  naming the upstream `key_id` and both boundaries, and the `import_bundle`
+  audit event records the compared boundaries and the outcome.
+  `--accept-older-upstream` imports an older bundle with a warning without
+  lowering the high-water; the inconsistent case is never overridable. No
+  protocol or wire change; the table ships inside `registry.db` (so
+  `backup` copies it) and is not compared by `verify-backup` (curator-spec
+  `dced9b8`, implementation detail).
git diff -U0 fb86420 9b7d331 -- CHANGELOG.md
diff --git a/CHANGELOG.md b/CHANGELOG.md
index c562b1d..ee5d4ee 100644
--- a/CHANGELOG.md
+++ b/CHANGELOG.md
@@ -35,0 +36,16 @@
+- P4: `import-bundle` now persists a per-upstream high-water (`version`,
+  `log_size`, `head`, `merkle_root` per upstream `key_id`) in a new
+  `upstream_high_water` table (schema version 4, migrated once at startup)
+  and compares every verified bundle against it under the client §5
+  rollback rules. A version below refuses with `import_upstream_rollback`;
+  an equal version with a different `head`/`merkle_root`/`log_size` refuses
+  with `import_upstream_inconsistent`; an equal identical boundary is an
+  accepted no-op; a higher version imports and advances the stored boundary
+  in the same transaction as the imported records. Refusals exit non-zero
+  naming the upstream `key_id` and both boundaries, and the `import_bundle`
+  audit event records the compared boundaries and the outcome.
+  `--accept-older-upstream` imports an older bundle with a warning without
+  lowering the high-water; the inconsistent case is never overridable. No
+  protocol or wire change; the table ships inside `registry.db` (so
+  `backup` copies it) and is not compared by `verify-backup` (curator-spec
+  `dced9b8`, implementation detail).
git diff -U0 c7ef32c 1dfd8513e036b3e2006dd6fc620e396ad2638ffb -- README.md
diff --git a/README.md b/README.md
index 862726d..0c04139 100644
--- a/README.md
+++ b/README.md
@@ -35 +35,4 @@ Ed25519 keys before trusting it.
-  and Merkle root all match the pinned upstream snapshot.
+  and Merkle root all match the pinned upstream snapshot, and only when the
+  upstream snapshot is not below the persisted per-upstream high-water for
+  that upstream key (a rollback refuses; `--accept-older-upstream` imports
+  it with a warning without lowering the stored high-water).
@@ -85,0 +89 @@ curator-skill-registry --home ./data import-bundle <f> --upstream-key <k>  # imp
+curator-skill-registry --home ./data import-bundle <f> --upstream-key <k> --accept-older-upstream  # import below the high-water with a warning
@@ -99,0 +104,20 @@ encrypted, access-controlled secret storage.
+### Upstream import high-water
+
+`import-bundle` persists, per upstream public key (`key_id` of
+`--upstream-key`), the highest accepted upstream snapshot boundary
+(`version`, `log_size`, `head`, `merkle_root`) in the `upstream_high_water`
+table inside `registry.db`, updated in the same transaction as the imported
+records. The first import from an unknown upstream establishes it. Later
+imports compare under the client §5 rollback rules: a version below refuses
+with `import_upstream_rollback`; an equal version with a different `head`,
+`merkle_root`, or `log_size` refuses with `import_upstream_inconsistent`
+(never overridable); an equal identical boundary is accepted as a no-op;
+a higher version imports and advances the stored boundary. Refusals exit
+non-zero with a diagnostic naming the upstream `key_id` and both boundaries,
+and the `import_bundle` audit event (stderr, structured JSON) carries the
+compared `persisted`/`offered` boundaries and the outcome.
+`--accept-older-upstream` turns only the version-below refusal into a
+warning and imports without lowering the persisted high-water. The table is
+part of the database, so `backup` copies it; `verify-backup` does not compare
+it (the signed checkpoint carries no upstream state).
+
@@ -208,0 +233,7 @@ single-row lookups.
+Schema version 4 adds `upstream_high_water` (one
+`(key_id, version, log_size, head, merkle_root, updated_at)` row per upstream
+public key, the highest accepted upstream boundary). The first startup on a
+version 2 or 3 database creates the empty table and bumps the markers; log
+history and the boundary cache are unchanged, and existing databases simply
+establish each upstream's high-water on its next import.
+
git diff -U0 fb86420 9b7d331 -- README.md
diff --git a/README.md b/README.md
index 47c36e1..d892bb9 100644
--- a/README.md
+++ b/README.md
@@ -35 +35,4 @@ Ed25519 keys before trusting it.
-  and Merkle root all match the pinned upstream snapshot.
+  and Merkle root all match the pinned upstream snapshot, and only when the
+  upstream snapshot is not below the persisted per-upstream high-water for
+  that upstream key (a rollback refuses; `--accept-older-upstream` imports
+  it with a warning without lowering the stored high-water).
@@ -85,0 +89 @@ curator-skill-registry --home ./data import-bundle <f> --upstream-key <k>  # imp
+curator-skill-registry --home ./data import-bundle <f> --upstream-key <k> --accept-older-upstream  # import below the high-water with a warning
@@ -99,0 +104,20 @@ encrypted, access-controlled secret storage.
+### Upstream import high-water
+
+`import-bundle` persists, per upstream public key (`key_id` of
+`--upstream-key`), the highest accepted upstream snapshot boundary
+(`version`, `log_size`, `head`, `merkle_root`) in the `upstream_high_water`
+table inside `registry.db`, updated in the same transaction as the imported
+records. The first import from an unknown upstream establishes it. Later
+imports compare under the client §5 rollback rules: a version below refuses
+with `import_upstream_rollback`; an equal version with a different `head`,
+`merkle_root`, or `log_size` refuses with `import_upstream_inconsistent`
+(never overridable); an equal identical boundary is accepted as a no-op;
+a higher version imports and advances the stored boundary. Refusals exit
+non-zero with a diagnostic naming the upstream `key_id` and both boundaries,
+and the `import_bundle` audit event (stderr, structured JSON) carries the
+compared `persisted`/`offered` boundaries and the outcome.
+`--accept-older-upstream` turns only the version-below refusal into a
+warning and imports without lowering the persisted high-water. The table is
+part of the database, so `backup` copies it; `verify-backup` does not compare
+it (the signed checkpoint carries no upstream state).
+
@@ -272,0 +297,7 @@ single-row lookups.
+Schema version 4 adds `upstream_high_water` (one
+`(key_id, version, log_size, head, merkle_root, updated_at)` row per upstream
+public key, the highest accepted upstream boundary). The first startup on a
+version 2 or 3 database creates the empty table and bumps the markers; log
+history and the boundary cache are unchanged, and existing databases simply
+establish each upstream's high-water on its next import.
+
git diff -U0 c7ef32c 1dfd8513e036b3e2006dd6fc620e396ad2638ffb -- SECURITY.md
diff --git a/SECURITY.md b/SECURITY.md
index e24b091..cfe3f43 100644
--- a/SECURITY.md
+++ b/SECURITY.md
@@ -37,0 +38,11 @@ than re-evaluated at a newer boundary.
+Upstream import is fail-closed against rollback: `import-bundle` persists a
+per-upstream high-water (the highest accepted upstream `version`, with
+`log_size`, `head`, and `merkle_root`) in `registry.db` and refuses an old
+but validly signed bundle as `import_upstream_rollback`, or an equal version
+with a different body as `import_upstream_inconsistent`, instead of
+re-importing it as new. Only an explicit `--accept-older-upstream` imports
+an older bundle, with a warning and without lowering the stored high-water;
+the inconsistent case is never overridable. Without this, an attacker
+replaying a stale upstream bundle could resurrect revoked or superseded
+records as fresh imports.
+
git diff -U0 fb86420 9b7d331 -- SECURITY.md
diff --git a/SECURITY.md b/SECURITY.md
index ea57863..0812269 100644
--- a/SECURITY.md
+++ b/SECURITY.md
@@ -57,0 +58,11 @@ than re-evaluated at a newer boundary.
+Upstream import is fail-closed against rollback: `import-bundle` persists a
+per-upstream high-water (the highest accepted upstream `version`, with
+`log_size`, `head`, and `merkle_root`) in `registry.db` and refuses an old
+but validly signed bundle as `import_upstream_rollback`, or an equal version
+with a different body as `import_upstream_inconsistent`, instead of
+re-importing it as new. Only an explicit `--accept-older-upstream` imports
+an older bundle, with a warning and without lowering the stored high-water;
+the inconsistent case is never overridable. Without this, an attacker
+replaying a stale upstream bundle could resurrect revoked or superseded
+records as fresh imports.
+
git diff -U0 c7ef32c 1dfd8513e036b3e2006dd6fc620e396ad2638ffb -- src/csk_registry/cli.py
diff --git a/src/csk_registry/cli.py b/src/csk_registry/cli.py
index 4b664e8..ea71394 100644
--- a/src/csk_registry/cli.py
+++ b/src/csk_registry/cli.py
@@ -15 +15,5 @@ from .auth import Auditor, AuditorTokens
-from .bundle import export_bundle, import_bundle
+from .bundle import (
+    IMPORT_UPSTREAM_ROLLBACK,
+    export_bundle,
+    import_bundle,
+)
@@ -235,0 +240,14 @@ def _cmd_export_bundle(args: argparse.Namespace) -> int:
+def _ensure_import_audit_sink() -> None:
+    """Attach the structured audit sink so import outcomes reach stderr."""
+    import logging
+
+    audit = logging.getLogger("csk_registry.audit")
+    if audit.level == logging.NOTSET or audit.level > logging.INFO:
+        audit.setLevel(logging.INFO)
+    if not audit.handlers:
+        handler = logging.StreamHandler()
+        handler.setLevel(logging.INFO)
+        handler.setFormatter(logging.Formatter("%(message)s"))
+        audit.addHandler(handler)
+
+
@@ -239,0 +258 @@ def _cmd_import_bundle(args: argparse.Namespace) -> int:
+    _ensure_import_audit_sink()
@@ -245 +264,7 @@ def _cmd_import_bundle(args: argparse.Namespace) -> int:
-        count = import_bundle(store, key, bundle, upstream_public_key=args.upstream_key)
+        count = import_bundle(
+            store,
+            key,
+            bundle,
+            upstream_public_key=args.upstream_key,
+            accept_older_upstream=args.accept_older_upstream,
+        )
@@ -248,0 +274,6 @@ def _cmd_import_bundle(args: argparse.Namespace) -> int:
+    if args.accept_older_upstream:
+        print(
+            f"warning: {IMPORT_UPSTREAM_ROLLBACK} override flag was given; "
+            "an older bundle imports without lowering the persisted high-water",
+            file=sys.stderr,
+        )
@@ -437,0 +469,8 @@ def build_parser() -> argparse.ArgumentParser:
+    import_b.add_argument(
+        "--accept-older-upstream",
+        action="store_true",
+        help=(
+            "import a bundle below the persisted upstream high-water with a "
+            "warning, without lowering it (same-version-different-body stays refused)"
+        ),
+    )
git diff -U0 fb86420 9b7d331 -- src/csk_registry/cli.py
diff --git a/src/csk_registry/cli.py b/src/csk_registry/cli.py
index 71b62d1..26fabc9 100644
--- a/src/csk_registry/cli.py
+++ b/src/csk_registry/cli.py
@@ -15 +15,5 @@ from .auth import Auditor, AuditorTokens
-from .bundle import export_bundle, import_bundle
+from .bundle import (
+    IMPORT_UPSTREAM_ROLLBACK,
+    export_bundle,
+    import_bundle,
+)
@@ -236,0 +241,14 @@ def _cmd_export_bundle(args: argparse.Namespace) -> int:
+def _ensure_import_audit_sink() -> None:
+    """Attach the structured audit sink so import outcomes reach stderr."""
+    import logging
+
+    audit = logging.getLogger("csk_registry.audit")
+    if audit.level == logging.NOTSET or audit.level > logging.INFO:
+        audit.setLevel(logging.INFO)
+    if not audit.handlers:
+        handler = logging.StreamHandler()
+        handler.setLevel(logging.INFO)
+        handler.setFormatter(logging.Formatter("%(message)s"))
+        audit.addHandler(handler)
+
+
@@ -240,0 +259 @@ def _cmd_import_bundle(args: argparse.Namespace) -> int:
+    _ensure_import_audit_sink()
@@ -246 +265,7 @@ def _cmd_import_bundle(args: argparse.Namespace) -> int:
-        count = import_bundle(store, key, bundle, upstream_public_key=args.upstream_key)
+        count = import_bundle(
+            store,
+            key,
+            bundle,
+            upstream_public_key=args.upstream_key,
+            accept_older_upstream=args.accept_older_upstream,
+        )
@@ -249,0 +275,6 @@ def _cmd_import_bundle(args: argparse.Namespace) -> int:
+    if args.accept_older_upstream:
+        print(
+            f"warning: {IMPORT_UPSTREAM_ROLLBACK} override flag was given; "
+            "an older bundle imports without lowering the persisted high-water",
+            file=sys.stderr,
+        )
@@ -438,0 +470,8 @@ def build_parser() -> argparse.ArgumentParser:
+    import_b.add_argument(
+        "--accept-older-upstream",
+        action="store_true",
+        help=(
+            "import a bundle below the persisted upstream high-water with a "
+            "warning, without lowering it (same-version-different-body stays refused)"
+        ),
+    )

```
