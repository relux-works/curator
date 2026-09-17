# TASK-260910-35279p review verdict — revision 1

Verdict: accepted. No blocking findings. Route with accept_cr; integration remains producer-owned.

Reviewed candidate tree 559e178fe3099c3292977d00e0e1909fd65e74d6 against base aea81ccf298071a7746d277bea316bdc91309c43. Published patch SHA-256 verified: a8f372c684abeb139cfca7df66cc664a65814a697b837c7d45b435e90efcea06. Worktree diff against candidate is empty; original four modified paths remain unchanged. Tests and attacks ran in /tmp/csk-review-35279p, exported from the candidate tree. No files added to the managed worktree.

Read campaign rules, producer brief/results, published delta and runtime validation log, R2 finding, curator-spec dced9b8317e0e8af79edf2d0539b32bd22b6c85b protocol sections 5/6 and service profile sections 2–6/11 and shared service vectors.

| Review item | Evidence and result |
| --- | --- |
| Durable boundary reads | store.py:912 indexed primary-key lookups plus head/genesis anchors; no prefix scan or hash work. snapshot_boundary, boundary_available and carried-boundary validation share this path. O(1) row count per lookup (SQLite B-tree lookup cost is logarithmic, not literal constant CPU complexity). |
| Transaction and concurrency | store.py:551–596: BEGIN IMMEDIATE writer transaction covers log insert, frontier update and boundary insert; separate connections see durable state, not an in-memory frontier. Public read RLock discipline unchanged; nothing moved outside the lock. Historical rows immutable; one row added on head advance. Existing concurrent-writer and idempotency tests pass. |
| Startup proof | store.py:189 runs existing chain verification before _ensure_boundaries at 636; every stored boundary is compared with recomputed prefix. Disagreement refuses readiness. Missing rows backfill; frontier is recoverable derived state and rebuilt on disagreement, consistent with profile section 5. |
| Root correctness | Committed test at test_registry.py:1275 checks 100 prefixes; independent Store-entry probe checks 130/130 sizes 0–129, including odd and power-of-two transitions, against naive Merkle computation. |
| Operation counts | test_registry.py:1285 instruments actual _merkle_pair_hash: repeated records/log first and cursor pages, snapshot and store boundary reads perform zero pair hashes. Test at 1373 bounds append hashing and checks root correctness. No timing assertion. |
| Migration | schema-2 migration transaction creates schema 3; separate backfill transaction is retryable. Independent injected failure after frontier saving but before backfill commit leaves zero boundary/frontier rows, then two reopens reconstruct all 130/130 roots. This is exception/rollback injection, not a host-power-loss test. |
| Negative evidence | 2/2 narrowing mutants killed by committed tests, details below. This measures only these two mutants, not exhaustive mutation coverage. |
| Validation | Independently reran full pytest with authoritative conformance root: 139 passed, exit 0. Strict mypy: 13 files clean, exit 0. Restored-mutant targeted tests: 2 passed, exit 0. |
| Scope/documentation | Four intended paths only. CHANGELOG R2 names spec dced9b8; README explains migration and startup verification. No endpoint, protocol, deployment or health caching changes. git diff --check exit 0. No separate lint tool configured. |

## Validation commands and environment

Shell /bin/zsh, set -o pipefail, macOS, Python 3.14.6. Full pytest ran with existing external /tmp/csk-venv/bin/python while a fresh external review venv installed; pytest pythonpath selects the disposable copy's src. Final mypy and restored controls used /tmp/csk-review-35279p-venv/bin/python. Mypy was rerun after mutant restoration to eliminate overlap ambiguity. No Python 3.11 local rerun.

From /tmp/csk-review-35279p:

```
CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1 /tmp/csk-venv/bin/python -m pytest -q
# exit 0
/tmp/csk-review-35279p-venv/bin/python -m mypy
# exit 0; strict=true in pyproject.toml
/tmp/csk-review-35279p-venv/bin/python -m pytest -q tests/test_registry.py -k 'boundary_cache_disagreement or boundary_append_cost'
# exit 0
/tmp/csk-review-35279p-venv/bin/python /tmp/csk-review-probe.py
# exit 0
```

Hosted gate was not rerun. Accepted attached runtime evidence TASK-260910-35279p_change-request_rev1-validation.log: command exit 0, GitHub run 35241440695 success, all six OS/Python test jobs plus strict mypy, distribution and Docker builds successful. That is attached evidence, not an independent local replay of those platforms.

## Narrowing mutants

1. Startup comparison narrowed from `if cached != boundary` to `if index % 2 == 0 and cached != boundary`. Committed test_boundary_cache_disagreement_fails_startup_but_missing_row_backfills fails because corrupted size 3 no longer raises. Exit 1.
2. Append frontier persistence narrowed to even sequence numbers only. Committed test_boundary_append_cost_is_logarithmic fails at the next append with frontier/log-head mismatch. Exit 1.

Restored source from captured original after attacks, cmp matches managed candidate; both targeted controls then pass. Probe and mutant harnesses are appended for reproducibility.

## Limits and observations

Read-time interior tampering is no longer fully scanned, intentionally: startup provides the full chain proof and request-time anchors check head/genesis. This is documented in producer results and matches this task's startup-validation design; arbitrary live external database modification is not newly claimed detectable per request. Frontier repair is derived-index recovery, while conflicting stored snapshot boundaries fail startup. No new blocker, regression or unresolved decision found; no separate logbook CLI/file was available, and review observations are persisted in this task-scoped outcome. Conditional nonacceptance checklist item is not applicable because verdict is accepted.

Run goal queried before verdict: no active goal (run not goal-bound).

## Full pytest transcript

```
........................................................................ [ 51%]
...................................................................      [100%]
=============================== warnings summary ===============================
../../../tmp/csk-venv/lib/python3.14/site-packages/fastapi/testclient.py:1
  /tmp/csk-venv/lib/python3.14/site-packages/fastapi/testclient.py:1: StarletteDeprecationWarning: Using `httpx` with `starlette.testclient` is deprecated; install `httpx2` instead.
    from starlette.testclient import TestClient as TestClient  # noqa

../../../tmp/csk-venv/lib/python3.14/site-packages/starlette/testclient.py:53
  /tmp/csk-venv/lib/python3.14/site-packages/starlette/testclient.py:53: DeprecationWarning: The anyio.abc.BlockingPortal alias is deprecated, use anyio.from_thread.BlockingPortal instead.
    _PortalFactoryType = Callable[[], AbstractContextManager[anyio.abc.BlockingPortal]]

-- Docs: https://docs.pytest.org/en/stable/how-to/capture-warnings.html
139 passed, 2 warnings in 12.32s

```

## Final mypy transcript

```
Success: no issues found in 13 source files

```

## Restored controls

```
..                                                                       [100%]
=============================== warnings summary ===============================
../../../tmp/csk-review-35279p-venv/lib/python3.14/site-packages/fastapi/testclient.py:1
  /tmp/csk-review-35279p-venv/lib/python3.14/site-packages/fastapi/testclient.py:1: StarletteDeprecationWarning: Using `httpx` with `starlette.testclient` is deprecated; install `httpx2` instead.
    from starlette.testclient import TestClient as TestClient  # noqa

../../../tmp/csk-review-35279p-venv/lib/python3.14/site-packages/starlette/testclient.py:53
  /tmp/csk-review-35279p-venv/lib/python3.14/site-packages/starlette/testclient.py:53: DeprecationWarning: The anyio.abc.BlockingPortal alias is deprecated, use anyio.from_thread.BlockingPortal instead.
    _PortalFactoryType = Callable[[], AbstractContextManager[anyio.abc.BlockingPortal]]

-- Docs: https://docs.pytest.org/en/stable/how-to/capture-warnings.html
2 passed, 51 deselected, 2 warnings in 4.39s

```

## Independent probe

```
PASS: 130/130 historical roots; interrupted backfill rolls back both tables; two reopens recover all roots

```

## Mutant 1

```
F                                                                        [100%]
=================================== FAILURES ===================================
___ test_boundary_cache_disagreement_fails_startup_but_missing_row_backfills ___

tmp_path = PosixPath('/private/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/pytest-of-administrator/pytest-55/test_boundary_cache_disagreeme0')

    def test_boundary_cache_disagreement_fails_startup_but_missing_row_backfills(
        tmp_path: Path,
    ) -> None:
        path = tmp_path / "cache.db"
        store = Store(path)
        key = signing.generate_key()
        for index in range(3):
            store.append(
                key.sign_record(_body(name=f"skill-{index}")),
                created_at=f"2026-07-13T00:00:0{index}Z",
            )
        genuine = store.snapshot_boundary()
        store.close()
    
        connection = sqlite3.connect(path)
        try:
            connection.execute(
                "UPDATE boundaries SET merkle_root = ? WHERE log_size = 3", ("ff" * 32,)
            )
            connection.commit()
        finally:
            connection.close()
>       with pytest.raises(StoreIntegrityError, match="disagrees with the committed log"):
             ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
E       Failed: DID NOT RAISE StoreIntegrityError

tests/test_registry.py:1421: Failed
=============================== warnings summary ===============================
../../../tmp/csk-venv/lib/python3.14/site-packages/fastapi/testclient.py:1
  /tmp/csk-venv/lib/python3.14/site-packages/fastapi/testclient.py:1: StarletteDeprecationWarning: Using `httpx` with `starlette.testclient` is deprecated; install `httpx2` instead.
    from starlette.testclient import TestClient as TestClient  # noqa

../../../tmp/csk-venv/lib/python3.14/site-packages/starlette/testclient.py:53
  /tmp/csk-venv/lib/python3.14/site-packages/starlette/testclient.py:53: DeprecationWarning: The anyio.abc.BlockingPortal alias is deprecated, use anyio.from_thread.BlockingPortal instead.
    _PortalFactoryType = Callable[[], AbstractContextManager[anyio.abc.BlockingPortal]]

-- Docs: https://docs.pytest.org/en/stable/how-to/capture-warnings.html
=========================== short test summary info ============================
FAILED tests/test_registry.py::test_boundary_cache_disagreement_fails_startup_but_missing_row_backfills
1 failed, 2 warnings in 1.81s

exit=1

```

## Mutant 2

```
F                                                                        [100%]
=================================== FAILURES ===================================
___________________ test_boundary_append_cost_is_logarithmic ___________________

tmp_path = PosixPath('/private/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/pytest-of-administrator/pytest-56/test_boundary_append_cost_is_l0')
monkeypatch = <_pytest.monkeypatch.MonkeyPatch object at 0x10a4f6c40>

    def test_boundary_append_cost_is_logarithmic(
        tmp_path: Path, monkeypatch: pytest.MonkeyPatch
    ) -> None:
        import csk_registry.store as store_module
    
        store = Store(tmp_path / "append.db")
        key = signing.generate_key()
>       _seed_paginated_store(store, key, 64)

tests/test_registry.py:1380: 
_ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ 
tests/test_registry.py:1268: in _seed_paginated_store
    store.append(
src/csk_registry/store.py:552: in append
    return self._append_locked(record, created_at=created_at)
           ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
_ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ 

self = <csk_registry.store.Store object at 0x10a619010>
record = {'schema_version': 1, 'name': 'skill-001', 'source_identity': 'gitlab.example.com/skills/skill-tracker', 'commit': '8c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9d', ...}
created_at = '2026-07-13T00:00:01Z'

    def _append_locked(self, record: dict[str, Any], *, created_at: str) -> LogEntry:
        for key in ("name", "source_identity", "commit", "content_sha256", "status"):
            if not isinstance(record.get(key), str) or not record[key]:
                raise ValueError(f"record requires a non-empty string {key!r}")
        record_bytes = canonical_bytes(record)
        row = self._conn.execute(
            "SELECT seq, entry_hash FROM log ORDER BY seq DESC LIMIT 1"
        ).fetchone()
        previous_seq = int(row["seq"]) if row else 0
        prev_hash = str(row["entry_hash"]) if row else _GENESIS
        entry_hash = hashlib.sha256(prev_hash.encode("ascii") + record_bytes).hexdigest()
        cursor = self._conn.execute(
            "INSERT INTO log (entry_hash, prev_hash, name, source_identity, commit_hash, "
            "content_sha256, status, record_json, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
            (
                entry_hash,
                prev_hash,
                record["name"],
                record["source_identity"],
                record["commit"],
                record["content_sha256"],
                record["status"],
                json.dumps(record, sort_keys=True, separators=(",", ":"), ensure_ascii=False),
                created_at,
            ),
        )
        seq = int(cursor.lastrowid or 0)
        if seq != previous_seq + 1:
            raise StoreIntegrityError("log sequence is not contiguous")
        frontier = self._load_frontier()
        if (frontier[0].length if frontier else 0) != previous_seq:
>           raise StoreIntegrityError("merkle frontier does not match the log head")
E           csk_registry.store.StoreIntegrityError: merkle frontier does not match the log head

src/csk_registry/store.py:585: StoreIntegrityError
=============================== warnings summary ===============================
../../../tmp/csk-venv/lib/python3.14/site-packages/fastapi/testclient.py:1
  /tmp/csk-venv/lib/python3.14/site-packages/fastapi/testclient.py:1: StarletteDeprecationWarning: Using `httpx` with `starlette.testclient` is deprecated; install `httpx2` instead.
    from starlette.testclient import TestClient as TestClient  # noqa

../../../tmp/csk-venv/lib/python3.14/site-packages/starlette/testclient.py:53
  /tmp/csk-venv/lib/python3.14/site-packages/starlette/testclient.py:53: DeprecationWarning: The anyio.abc.BlockingPortal alias is deprecated, use anyio.from_thread.BlockingPortal instead.
    _PortalFactoryType = Callable[[], AbstractContextManager[anyio.abc.BlockingPortal]]

-- Docs: https://docs.pytest.org/en/stable/how-to/capture-warnings.html
=========================== short test summary info ============================
FAILED tests/test_registry.py::test_boundary_append_cost_is_logarithmic - csk...
1 failed, 2 warnings in 2.45s

exit=1

```

## Probe harness

```
import sys, tempfile, sqlite3
from pathlib import Path
sys.path.insert(0,'/tmp/csk-review-35279p/src')
from csk_registry.store import Store, StoreIntegrityError, _merkle_root
p=Path(tempfile.mkdtemp())/'db'
s=Store(p)
for i in range(129):
 s.append(dict(name=str(i),source_identity='example/a',commit='a'*40,content_sha256='sha256:'+'b'*64,status='audited'),created_at='2026-09-17T00:00:00Z')
leaves=[r[0] for r in s._conn.execute('SELECT entry_hash FROM log ORDER BY seq')]
for n in range(130):
 assert s.snapshot_boundary(n).merkle_root==_merkle_root(leaves[:n])
s.close()
with sqlite3.connect(p) as c:
 c.execute('DROP TABLE boundaries'); c.execute('DROP TABLE merkle_frontier')
 c.execute("UPDATE metadata SET value='2' WHERE key='schema_version'"); c.execute('PRAGMA user_version=2')
original=Store._save_frontier
def interrupted(self, frontier):
 original(self,frontier)
 raise StoreIntegrityError('injected interruption before backfill commit')
Store._save_frontier=interrupted
try:
 try: Store(p)
 except StoreIntegrityError as e: assert 'injected interruption' in str(e)
 else: raise AssertionError('injection not reached')
finally: Store._save_frontier=original
with sqlite3.connect(p) as c:
 assert c.execute('SELECT count(*) FROM boundaries').fetchone()[0]==0
 assert c.execute('SELECT count(*) FROM merkle_frontier').fetchone()[0]==0
for _ in range(2):
 s=Store(p)
 for n in range(130): assert s.snapshot_boundary(n).merkle_root==_merkle_root(leaves[:n])
 s.close()
print('PASS: 130/130 historical roots; interrupted backfill rolls back both tables; two reopens recover all roots')

```

## Mutation harness

```
from pathlib import Path
import subprocess, os
p=Path('/tmp/csk-review-35279p/src/csk_registry/store.py')
original=p.read_text()
env=dict(os.environ, CURATOR_CONFORMANCE_ROOT='/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1')
mutants=[('startup-even-only','if cached != boundary:', 'if index % 2 == 0 and cached != boundary:', 'test_boundary_cache_disagreement_fails_startup_but_missing_row_backfills'),('frontier-even-only','        self._save_frontier(frontier)\n        try:', '        if seq % 2 == 0:\n            self._save_frontier(frontier)\n        try:', 'test_boundary_append_cost_is_logarithmic')]
try:
 for name,old,new,test in mutants:
  assert original.count(old)==1
  p.write_text(original.replace(old,new))
  r=subprocess.run(['/tmp/csk-venv/bin/python','-m','pytest','-q','tests/test_registry.py::'+test],cwd='/tmp/csk-review-35279p',env=env,text=True,stdout=subprocess.PIPE,stderr=subprocess.STDOUT)
  Path('/tmp/csk-review-'+name+'.log').write_text(r.stdout+'\nexit='+str(r.returncode)+'\n')
  print(name, 'exit=',r.returncode, r.stdout[-1200:])
finally:
 p.write_text(original)

```
