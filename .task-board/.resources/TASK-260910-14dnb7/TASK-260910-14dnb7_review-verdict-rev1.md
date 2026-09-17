# TASK-260910-14dnb7 review verdict — revision 1

Verdict: changes_requested; route to `to-dev`. Candidate was not modified.

## Identity and method

Reviewed tree `7797df0209184dddab370e8e3a725f7005db3eda` against base
`b13ad660c8aa7fb0a484315fa9bed3d2dd9c9810`. Live worktree diff against the
candidate tree was empty. Published patch SHA-256 matched
`7533ee38b95ec45d00b5fd40d9a04f04af93acb22de89b04299bc24794e9f0df`.
Read campaign rules, producer brief/results, patch and hosted validation log,
security audit R1, registry protocol §§5/9/9.3, service profile §§2/5/11,
v2 schemas and referenced schemas, and shared pagination vectors.
Spec checkout HEAD: `dced9b8317e0e8af79edf2d0539b32bd22b6c85b`.
Tests ran from a git-archive disposable copy `/tmp/csk-review-14dnb7`, with
venv `/tmp/csk-review-14dnb7-venv`; no candidate files or caches were written.

## Findings requiring correction

1. **High: live cursor chains change boundary signatures across staged key rotation.**
   `src/csk_registry/app.py:260,310` signs every page with the currently active
   key. `_encode_cursor` at :398 stores only the boundary body; continuation
   has no retained original signed object. The existing staged-rotation test
   at `tests/test_registry.py:664` allows an old cursor with the overlap pin
   set but never compares its boundary. Adding that comparison to the test in
   the disposable copy fails: same snapshot body, different `sig.key_id` and
   `sig.signature`. Protocol §9.3 and profile §2 require byte-identical complete
   signed boundaries across a chain. Profile §5 permits replacing a snapshot's
   signature during rotation, but does not waive the specific chain rule.
   The producer's observation is therefore a known contract violation, not an
   acceptance exception. Preserve the original signed chain boundary during
   valid overlap continuation and test both endpoints across rotation. Keep
   snapshot reads free to use the new signer. Coordinate any required cursor
   representation adjustment with the sibling; do not silently waive equality
   or invalidate still-live cursors before their promised lifetime.

2. **Medium: the new conformance checker does not validate the named schemas.**
   `tests/test_protocol_conformance.py:170-195` loads only `required` and
   `additionalProperties`, then substitutes a partial handwritten validator.
   In particular :188-192 omit log-entry hash formats, seq minimum/maximum,
   and the referenced audit-record schema. A served response whose entry_hash
   is `"invalid"` passes the entire committed suite (see mutant transcript).
   `registry-log-entry-v1.schema.json`, referenced by log-response-v2, forbids
   that value. Use the actual Draft 2020-12 schemas with local reference
   resolution on served envelopes and registered schema-cases, or supply a
   demonstrably complete equivalent. Include a negative regression proving
   this malformed served envelope is rejected by the conformance harness.

## Per-item assessment

| Requirement | Assessment and evidence |
|---|---|
| Full signed boundary on both success envelopes | Pass for normal requests: app.py:260-261,310-311; snapshot.py:22-35 reuses evaluated immutable body and signature builder. |
| Same boundary across append/time | Append coverage passes on both endpoints in test_registry.py:782-853; timestamp sourced from immutable boundary, not wall clock. Rotation violates full equality (finding 1). |
| Equals snapshot, verifies service key | Pass for stable key: test_protocol_conformance.py:235-242,259-260,275-280 and unit endpoint tests. |
| Closed v2 envelopes and schema-cases | Runtime envelopes have exactly three members; schema-case registration is driven, but full schema validation is incomplete (finding 2). |
| Shared pagination flags | Both flags consumed at test_protocol_conformance.py:224-225 through actual TestClient endpoints, plus append vector and log chain. |
| CI pin and existing tests | ci.yml:30 exact required pin; all 124 tests pass, no edits to prior conformance cases. |
| Docs and results | CHANGELOG.md:36-40 names R1, boundary, v2 schemas and dced9b8; README.md:26-28,60-64 describes boundary. Producer results provide commands and exits. |
| Scope / architecture | Six scoped files; existing builder reuse fits architecture. No cursor disagreement refusal or Docker/deploy change introduced. |
| Error paths | Changes affect only successful returns; existing error tests pass. |
| Static/build | Independent strict mypy green; no separate linter configured; git diff --check clean. Hosted build/Docker/matrix evidence accepted from attached runtime log, not rerun. |

## Independent validation

Shell: zsh; `set -o pipefail`; Python 3.14.6. Working directory
`/tmp/csk-review-14dnb7`.

```sh
export CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1
/tmp/csk-review-14dnb7-venv/bin/python -m pytest -q
# exit 0
/tmp/csk-review-14dnb7-venv/bin/python -m mypy
# exit 0; strict=true in pyproject.toml
```

An initial premature invocation while pip was still installing returned exit 1
(`No module named pytest` / `mypy`); after installation completed with exit 0,
the actual independent runs below completed successfully. No product change
was made to obtain green results.

### pytest

```text
........................................................................ [ 58%]
....................................................                     [100%]
=============================== warnings summary ===============================
../../../tmp/csk-review-14dnb7-venv/lib/python3.14/site-packages/fastapi/testclient.py:1
  /tmp/csk-review-14dnb7-venv/lib/python3.14/site-packages/fastapi/testclient.py:1: StarletteDeprecationWarning: Using `httpx` with `starlette.testclient` is deprecated; install `httpx2` instead.
    from starlette.testclient import TestClient as TestClient  # noqa

../../../tmp/csk-review-14dnb7-venv/lib/python3.14/site-packages/starlette/testclient.py:53
  /tmp/csk-review-14dnb7-venv/lib/python3.14/site-packages/starlette/testclient.py:53: DeprecationWarning: The anyio.abc.BlockingPortal alias is deprecated, use anyio.from_thread.BlockingPortal instead.
    _PortalFactoryType = Callable[[], AbstractContextManager[anyio.abc.BlockingPortal]]

-- Docs: https://docs.pytest.org/en/stable/how-to/capture-warnings.html
124 passed, 2 warnings in 31.83s

exit=0
```

### mypy

```text
Success: no issues found in 13 source files

exit=0
```

## Narrowing attacks and reproduction

Harness: `/tmp/csk-review-14dnb7-attacks.py` (attached with task-scoped name).
Each attack changed only the disposable app copy and restored exact saved bytes.
Two boundary mutants targeted distinct endpoint/continuation branches rather
than removing boundary support globally. Both were killed (2/2). A third
schema-invalid wire mutant survived (1/1), demonstrating finding 2. These
numbers cover only the named mutations, not exhaustive protocol coverage.

- records_cursor_missing: omit boundary only when records cursor is present;
  committed records chain test must fail, while log remains unmodified.
- log_cursor_latest: sign current snapshot only on log cursor pages; committed
  log append/chain test must fail, while first pages remain correct.
- log_bad_hash: serialize log entry_hash as `invalid`; full committed suite
  passes despite the referenced schema's hex256 requirement.
- rotation: no production mutation; add complete boundary equality to the
  existing staged key rotation test. It fails on signature differences.

### records_cursor_missing

```text
F.                                                                       [100%]
=================================== FAILURES ===================================
_______________ test_records_pages_carry_byte_identical_boundary _______________

tmp_path = PosixPath('/private/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/pytest-of-administrator/pytest-21/test_records_pages_carry_byte_0')

    def test_records_pages_carry_byte_identical_boundary(tmp_path: Path):
        client, registry_key, token = _client(tmp_path)
        auditor_key = client.app.state.auditor_key  # type: ignore[attr-defined]
        content_hash = "sha256:" + "1f" * 32
        for index in range(3):
            record = auditor_key.sign_record(_body(name=f"skill-{index}"))
            assert client.post(
                "/v1/records", json=record, headers={"Authorization": f"Bearer {token}"}
            ).status_code == 201
        snapshot = client.get("/v1/snapshot").json()
        first = client.get(
            "/v1/records", params={"content_sha256": content_hash, "limit": 1}
        ).json()
        assert set(first) == {"records", "next_cursor", "boundary"}
        assert first["boundary"] == snapshot
        assert signing.verify_signed(registry_key.public_pinned, first["boundary"])
        # An append after the first page must not move the chain's boundary.
        replacement = auditor_key.sign_record(_body("revoked", name="skill-0"))
        assert client.post(
            "/v1/records", json=replacement, headers={"Authorization": f"Bearer {token}"}
        ).status_code == 201
        chain = signing.canonical_document_bytes(first["boundary"])
        body = first
        names = [record["name"] for record in body["records"]]
        while body["next_cursor"] is not None:
            body = client.get(
                "/v1/records",
                params={"content_sha256": content_hash, "limit": 1, "cursor": body["next_cursor"]},
            ).json()
>           assert set(body) == {"records", "next_cursor", "boundary"}
E           AssertionError: assert {'next_cursor', 'records'} == {'boundary', ...r', 'records'}
E             
E             Extra items in the right set:
E             'boundary'
E             Use -v to get more diff

tests/test_registry.py:811: AssertionError
=============================== warnings summary ===============================
../../../tmp/csk-review-14dnb7-venv/lib/python3.14/site-packages/fastapi/testclient.py:1
  /tmp/csk-review-14dnb7-venv/lib/python3.14/site-packages/fastapi/testclient.py:1: StarletteDeprecationWarning: Using `httpx` with `starlette.testclient` is deprecated; install `httpx2` instead.
    from starlette.testclient import TestClient as TestClient  # noqa

../../../tmp/csk-review-14dnb7-venv/lib/python3.14/site-packages/starlette/testclient.py:53
  /tmp/csk-review-14dnb7-venv/lib/python3.14/site-packages/starlette/testclient.py:53: DeprecationWarning: The anyio.abc.BlockingPortal alias is deprecated, use anyio.from_thread.BlockingPortal instead.
    _PortalFactoryType = Callable[[], AbstractContextManager[anyio.abc.BlockingPortal]]

-- Docs: https://docs.pytest.org/en/stable/how-to/capture-warnings.html
=========================== short test summary info ============================
FAILED tests/test_registry.py::test_records_pages_carry_byte_identical_boundary
1 failed, 1 passed, 39 deselected, 2 warnings in 0.66s

exit=1

```

### log_cursor_latest

```text
.F                                                                       [100%]
=================================== FAILURES ===================================
_________________ test_log_pages_carry_byte_identical_boundary _________________

tmp_path = PosixPath('/private/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/pytest-of-administrator/pytest-22/test_log_pages_carry_byte_iden0')

    def test_log_pages_carry_byte_identical_boundary(tmp_path: Path):
        client, registry_key, token = _client(tmp_path)
        auditor_key = client.app.state.auditor_key  # type: ignore[attr-defined]
        for index in range(3):
            record = auditor_key.sign_record(_body("audited", commit=f"{index:040d}"))
            assert client.post(
                "/v1/records", json=record, headers={"Authorization": f"Bearer {token}"}
            ).status_code == 201
        snapshot = client.get("/v1/snapshot").json()
        first = client.get("/v1/log", params={"limit": 1}).json()
        assert set(first) == {"entries", "next_cursor", "boundary"}
        assert first["boundary"] == snapshot
        assert signing.verify_signed(registry_key.public_pinned, first["boundary"])
        late = auditor_key.sign_record(_body("audited", commit=f"{3:040d}"))
        assert client.post(
            "/v1/records", json=late, headers={"Authorization": f"Bearer {token}"}
        ).status_code == 201
        chain = signing.canonical_document_bytes(first["boundary"])
        body = first
        sequences = [entry["seq"] for entry in body["entries"]]
        while body["next_cursor"] is not None:
            body = client.get(
                "/v1/log",
                params={"limit": 1, "cursor": body["next_cursor"]},
            ).json()
            assert set(body) == {"entries", "next_cursor", "boundary"}
            assert signing.verify_signed(registry_key.public_pinned, body["boundary"])
>           assert signing.canonical_document_bytes(body["boundary"]) == chain
E           assert b'{"created_a...,"version":4}' == b'{"created_a...,"version":3}'
E             
E             At index 45 diff: b'8' != b'9'
E             Use -v to get more diff

tests/test_registry.py:847: AssertionError
=============================== warnings summary ===============================
../../../tmp/csk-review-14dnb7-venv/lib/python3.14/site-packages/fastapi/testclient.py:1
  /tmp/csk-review-14dnb7-venv/lib/python3.14/site-packages/fastapi/testclient.py:1: StarletteDeprecationWarning: Using `httpx` with `starlette.testclient` is deprecated; install `httpx2` instead.
    from starlette.testclient import TestClient as TestClient  # noqa

../../../tmp/csk-review-14dnb7-venv/lib/python3.14/site-packages/starlette/testclient.py:53
  /tmp/csk-review-14dnb7-venv/lib/python3.14/site-packages/starlette/testclient.py:53: DeprecationWarning: The anyio.abc.BlockingPortal alias is deprecated, use anyio.from_thread.BlockingPortal instead.
    _PortalFactoryType = Callable[[], AbstractContextManager[anyio.abc.BlockingPortal]]

-- Docs: https://docs.pytest.org/en/stable/how-to/capture-warnings.html
=========================== short test summary info ============================
FAILED tests/test_registry.py::test_log_pages_carry_byte_identical_boundary
1 failed, 1 passed, 39 deselected, 2 warnings in 0.81s

exit=1

```

### log_bad_hash

```text
........................................................................ [ 58%]
....................................................                     [100%]
=============================== warnings summary ===============================
../../../tmp/csk-review-14dnb7-venv/lib/python3.14/site-packages/fastapi/testclient.py:1
  /tmp/csk-review-14dnb7-venv/lib/python3.14/site-packages/fastapi/testclient.py:1: StarletteDeprecationWarning: Using `httpx` with `starlette.testclient` is deprecated; install `httpx2` instead.
    from starlette.testclient import TestClient as TestClient  # noqa

../../../tmp/csk-review-14dnb7-venv/lib/python3.14/site-packages/starlette/testclient.py:53
  /tmp/csk-review-14dnb7-venv/lib/python3.14/site-packages/starlette/testclient.py:53: DeprecationWarning: The anyio.abc.BlockingPortal alias is deprecated, use anyio.from_thread.BlockingPortal instead.
    _PortalFactoryType = Callable[[], AbstractContextManager[anyio.abc.BlockingPortal]]

-- Docs: https://docs.pytest.org/en/stable/how-to/capture-warnings.html
124 passed, 2 warnings in 7.75s

exit=0

```

### rotation

```text
F                                                                        [100%]
=================================== FAILURES ===================================
______ test_staged_key_rotation_preserves_snapshot_body_and_live_cursors _______

tmp_path = PosixPath('/private/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/pytest-of-administrator/pytest-24/test_staged_key_rotation_prese0')

    def test_staged_key_rotation_preserves_snapshot_body_and_live_cursors(tmp_path: Path):
        home = tmp_path / "home"
        assert main(["--home", str(home), "genkey"]) == 0
        old_key = load_active_key(home)
        store = Store(home / "registry.db")
        content_hash = "sha256:" + "1f" * 32
        for index in range(2):
            store.append(
                old_key.sign_record(_body(name=f"skill-{index}")),
                created_at=f"2026-07-13T00:00:0{index}Z",
            )
        before = TestClient(
            create_app(
                store=store,
                signing_key=old_key,
                tokens=AuditorTokens([]),
                verification_keys=public_keys(home, old_key),
            )
        )
        first = before.get(
            "/v1/records",
            params={"content_sha256": content_hash, "limit": 1},
        )
        assert first.status_code == 200
        cursor = first.json()["next_cursor"]
        old_snapshot = before.get("/v1/snapshot").json()
    
        assert main(["--home", str(home), "genkey", "--force"]) == 1
        assert main(["--home", str(home), "prepare-key-rotation"]) == 0
        assert len(public_keys(home, old_key)) == 2
        assert main(["--home", str(home), "activate-key-rotation"]) == 1
        assert main(
            ["--home", str(home), "activate-key-rotation", "--confirm-pins-deployed"]
        ) == 0
    
        new_key = load_active_key(home)
        retained = public_keys(home, new_key)
        assert new_key.key_id != old_key.key_id
        assert {old_key.public_pinned, new_key.public_pinned} == set(retained)
        after = TestClient(
            create_app(
                store=store,
                signing_key=new_key,
                tokens=AuditorTokens([]),
                verification_keys=retained,
            )
        )
        new_snapshot = after.get("/v1/snapshot").json()
        assert {field: value for field, value in old_snapshot.items() if field != "sig"} == {
            field: value for field, value in new_snapshot.items() if field != "sig"
        }
        continued = after.get(
            "/v1/records",
            params={"content_sha256": content_hash, "limit": 1, "cursor": cursor},
        )
        assert continued.status_code == 200
>       assert continued.json()["boundary"] == first.json()["boundary"]
E       AssertionError: assert {'schema_vers...c7cb48c', ...} == {'schema_vers...c7cb48c', ...}
E         
E         Omitting 6 identical items, use -vv to show
E         Differing items:
E         {'sig': {'key_id': 'ded85d726b2a3bc9', 'algorithm': 'ed25519', 'signature': '0CNKNyUsz6tqigjKX66rzbSr1AT/dhvWWAKh3vwzHfmz65N+pi8FeZGjaCPn6KS1LSbaIGyVwdQgVxJ22iFoCQ=='}} != {'sig': {'key_id': '467f9dd4b061db3f', 'algorithm': 'ed25519', 'signature': 'gD+2QJerPR+uAvHOBnKx+ECJ5fCCa2MmjHsNrn6JopZYKgXnaKvoEk6f9ECt/DTCPwyCiOu9mR8LkmZbr0f9CQ=='}}
E         Use -v to get more diff

tests/test_registry.py:720: AssertionError
----------------------------- Captured stdout call -----------------------------
{
  "key_id": "467f9dd4b061db3f",
  "public_key": "ed25519:+8vXLVCzp1C+pR/Kpu5te5/gD6rOK0i7swunOkjminw=",
  "path": "/private/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/pytest-of-administrator/pytest-24/test_staged_key_rotation_prese0/home/signing-key.pem"
}
{
  "active_key_id": "467f9dd4b061db3f",
  "active_public_key": "ed25519:+8vXLVCzp1C+pR/Kpu5te5/gD6rOK0i7swunOkjminw=",
  "next_key_id": "ded85d726b2a3bc9",
  "next_public_key": "ed25519:Bpo288TMId12g9WpTHoFGQREdP1EdaeG1lW4zpNkNPs=",
  "next_step": "deploy the expanded out-of-band pin set, then run activate-key-rotation"
}
{
  "active_key_id": "ded85d726b2a3bc9",
  "active_public_key": "ed25519:Bpo288TMId12g9WpTHoFGQREdP1EdaeG1lW4zpNkNPs=",
  "retained_key_id": "467f9dd4b061db3f",
  "next_step": "wait through the overlap and cursor-retention window before retire-key"
}
----------------------------- Captured stderr call -----------------------------
refusing to replace the signing key for a non-empty registry; use staged rotation
activation requires --confirm-pins-deployed
=============================== warnings summary ===============================
../../../tmp/csk-review-14dnb7-venv/lib/python3.14/site-packages/fastapi/testclient.py:1
  /tmp/csk-review-14dnb7-venv/lib/python3.14/site-packages/fastapi/testclient.py:1: StarletteDeprecationWarning: Using `httpx` with `starlette.testclient` is deprecated; install `httpx2` instead.
    from starlette.testclient import TestClient as TestClient  # noqa

../../../tmp/csk-review-14dnb7-venv/lib/python3.14/site-packages/starlette/testclient.py:53
  /tmp/csk-review-14dnb7-venv/lib/python3.14/site-packages/starlette/testclient.py:53: DeprecationWarning: The anyio.abc.BlockingPortal alias is deprecated, use anyio.from_thread.BlockingPortal instead.
    _PortalFactoryType = Callable[[], AbstractContextManager[anyio.abc.BlockingPortal]]

-- Docs: https://docs.pytest.org/en/stable/how-to/capture-warnings.html
=========================== short test summary info ============================
FAILED tests/test_registry.py::test_staged_key_rotation_preserves_snapshot_body_and_live_cursors
1 failed, 40 deselected, 2 warnings in 1.72s

exit=1

```

## Evidence reuse and remaining bounds

Hosted validation resource `TASK-260910-14dnb7_change-request_rev1-validation.log`
reports successful run 35228541270: all six OS/Python combinations, mypy,
distribution build and Docker; exit 0. I inspected that attached evidence and
did not invoke the hosted gate. Independently executed only Python 3.14.6 on
macOS; no claim of independent 3.11 or other-OS execution.

No spec edits, commits, pushes, or candidate modifications. Producer's historical
filesystem-write hygiene outside this candidate cannot be independently proven;
reviewer writes were confined to /tmp and authorized board resource/status APIs.
The run goal query returned no active goal. This is ordinary implementation/test
rework, not an external blocker or human-only decision. Another producer and
reviewer cycle is required before acceptance.
