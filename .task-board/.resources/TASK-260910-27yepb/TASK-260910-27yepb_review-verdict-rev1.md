# TASK-260910-27yepb review verdict — revision 1

Verdict: **accepted**. No blocking findings. Reviewer independently inspected and tested candidate tree `aa4b9b61ded7010182b790c85e2a6ba4de7e9256` against curator-spec `dced9b8317e0e8af79edf2d0539b32bd22b6c85b`.

## Identity and scope

Published patch SHA-256 verified: `b463dcf4adfec0340a76a045063a1047d506883b78796120258975091bf11af4`. Read the published patch, producer results, first-leaf revision-2 results and validation log. `git diff aa4b9b6 --exit-code` returned 0 in the managed worktree. The nine-path CR delta includes the accepted first leaf: current HEAD `feecd4b30878c9a63a5a1d59087fe239d98d5e31` is its checkpoint; this task adds seven changed paths relative to HEAD. No unrelated implementation, R2 memoization, key-rotation policy, manager, spec or deployment change.

Read profile sections 2/5/11, protocol sections 5/9/9.3, both v2 schemas and the authoritative pagination vectors. Architecture-diagrams skill not applicable: no diagram or architectural redesign requested.

## Per-item review

| Requirement | Evidence and result |
|---|---|
| Cursor page evaluated only at carried boundary | `src/csk_registry/app.py:242,304` decode carried state; `:250,312` pass `boundary=boundary` to records/log store methods. No current-snapshot fallback in either cursor branch. PASS. |
| Store fields and prefix verified | `src/csk_registry/store.py:594,640` call `_page_boundary_locked` under the paging lock; `:678-706` derives at carried log_size and compares the complete SnapshotBoundary. `:709-746` checks bounds and contiguous prefix and derives head, root and immutable timestamp. `:748` pre-check independently compares the same struct. PASS. |
| Disagreement/unavailability is 404 invalid_cursor | `app.py:259,319,516,528`; mismatch maps to the closed error code. `tests/test_registry.py:967,1029,1086,1134` cover active-key head/root forgeries, overlap-key forgery, unavailable future size and pruned prefix on both endpoints. PASS. |
| Shared vectors actually driven | `tests/test_protocol_conformance.py:362` drives 1/1 cursor_boundary_cases through 2/2 endpoints; `:450` covers 5/5 cursor_rejections, ten refusal probes. Unknown names fail explicit assertions. PASS. |
| No re-evaluation after append | `test_protocol_conformance.py:395` onward checks original records and boundary after the vector append; `test_registry.py:891,929` checks both endpoint chains after append. PASS. |
| Prior first-leaf guarantees | `test_registry.py:744` tests byte-identical boundaries through rotation/retirement on both endpoints; `test_protocol_conformance.py:178,199,222,250` uses real Draft 2020-12 schemas and a negative malformed-entry check. All passed. |
| Operator docs / release | CHANGELOG.md:36-49 names R1/P1 and dced9b8; README.md:24-31 and SECURITY.md:27-30 explain refusal. PASS. |
| Validation and hygiene | Independent 133-test suite and strict mypy pass below. `git diff --check` exit 0. No separate linter configured; pyproject.toml:49 has strict=true. Candidate untouched, tests/mutants/venv and logs all under /tmp. PASS. |

## Independent validation

Shell: zsh, `set -o pipefail`. Python `/tmp/csk-review-27yepb-venv/bin/python`, version **3.14.6**. Exact candidate exported with `git archive` into `/tmp/csk-review-27yepb`; venv is outside that tree. Installation `pip install -e '/tmp/csk-review-27yepb[dev]'` exited 0.

Setup anomaly, not a candidate failure: initial pytest and mypy calls were launched before pip finished. pytest exited 2 (missing fastapi, collection); mypy exited 1 (missing dependency imports, 22 consequent errors). Neither was treated as validation. Both rerun after observing pip exit 0, with results below.

Commands from the disposable candidate copy:
```sh
set -o pipefail
export CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1
/tmp/csk-review-27yepb-venv/bin/python -m pytest -q
/tmp/csk-review-27yepb-venv/bin/python -m mypy
```

### pytest transcript
```text
........................................................................ [ 54%]
.............................................................            [100%]
=============================== warnings summary ===============================
../../../tmp/csk-review-27yepb-venv/lib/python3.14/site-packages/fastapi/testclient.py:1
  /tmp/csk-review-27yepb-venv/lib/python3.14/site-packages/fastapi/testclient.py:1: StarletteDeprecationWarning: Using `httpx` with `starlette.testclient` is deprecated; install `httpx2` instead.
    from starlette.testclient import TestClient as TestClient  # noqa

../../../tmp/csk-review-27yepb-venv/lib/python3.14/site-packages/starlette/testclient.py:53
  /tmp/csk-review-27yepb-venv/lib/python3.14/site-packages/starlette/testclient.py:53: DeprecationWarning: The anyio.abc.BlockingPortal alias is deprecated, use anyio.from_thread.BlockingPortal instead.
    _PortalFactoryType = Callable[[], AbstractContextManager[anyio.abc.BlockingPortal]]

-- Docs: https://docs.pytest.org/en/stable/how-to/capture-warnings.html
133 passed, 2 warnings in 12.45s

EXIT=0
```

### mypy transcript
```text
Success: no issues found in 13 source files

EXIT=0
```

## Independent narrowing attacks

Separate archive `/tmp/csk-review-27yepb-mutant-tree`; tests unchanged. Mutated both equality checks (`_page_boundary_locked` and `boundary_available`) together so one redundant check could not mask the weakened other check. Prefix availability, signatures, expiry, query binding and all other gates remain intact.

- M1: compare **only head**, retaining prefix verification. Merkle-root forgeries become admitted (HTTP 200), and both named committed tests fail.
- M2: compare **only merkle_root**, retaining prefix verification. Head forgeries become admitted (HTTP 200), and both named committed tests fail.

Measured result: **2/2 narrowing mutants killed, 0/2 survivors**. Each run: 2 failed / 131 deselected, pytest exit 1. The actual failed assertions are 200 versus required 404, not harness/import errors. Bounds: this is targeted head/root-class mutation coverage, not exhaustive mutation coverage of every predicate or interprocess race; size/pruning and append behavior additionally have passing committed negative/control tests.

Exact command for each mutated tree (same Python and conformance root):
```sh
python -m pytest -q tests/test_registry.py tests/test_protocol_conformance.py -k 'cursor_carried_boundary_disagreement or shared_service_cursor_boundary_cases'
```
Both mutant source edits restored from saved candidate bytes; full transcripts follow.

```text
MUTANT head-only: both equality gates compare only head; prefix availability remains enforced.
FF                                                                       [100%]
=================================== FAILURES ===================================
_____ test_cursor_carried_boundary_disagreement_refused_on_both_endpoints ______

tmp_path = PosixPath('/private/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/pytest-of-administrator/pytest-47/test_cursor_carried_boundary_d0')

    def test_cursor_carried_boundary_disagreement_refused_on_both_endpoints(tmp_path: Path):
        client, registry_key, token = _client(tmp_path)
        auditor_key = client.app.state.auditor_key  # type: ignore[attr-defined]
        content_hash = "sha256:" + "1f" * 32
        for index in range(3):
            record = auditor_key.sign_record(_body(name=f"skill-{index}"))
            assert client.post(
                "/v1/records", json=record, headers={"Authorization": f"Bearer {token}"}
            ).status_code == 201
        records_first = client.get(
            "/v1/records", params={"content_sha256": content_hash, "limit": 1}
        ).json()
        log_first = client.get("/v1/log", params={"since": 0, "limit": 1}).json()
        assert records_first["next_cursor"] and log_first["next_cursor"]
        # Control: the genuine cursors continue the chain at the original boundary.
        genuine_records = client.get(
            "/v1/records",
            params={
                "content_sha256": content_hash,
                "limit": 1,
                "cursor": records_first["next_cursor"],
            },
        )
        assert genuine_records.status_code == 200
        assert genuine_records.json()["boundary"] == records_first["boundary"]
    
        cases = [
            (
                "/v1/records",
                "records",
                {"source_identity": "", "commit": "", "content_sha256": content_hash, "limit": 1},
                {"content_sha256": content_hash, "limit": 1},
                records_first["boundary"],
            ),
            (
                "/v1/log",
                "log",
                {"since": 0, "limit": 1},
                {"since": 0, "limit": 1},
                log_first["boundary"],
            ),
        ]
        for endpoint, name, query, params, boundary in cases:
            for field in ("head", "merkle_root"):
                forged = dict(boundary)
                forged[field] = _flip_hex(forged[field])
                # The forged body is genuinely re-signed with a key the service
                # accepts, so only the store comparison can refuse it.
                signed = registry_key.sign_record(forged)
                assert signing.verify_signed(registry_key.public_pinned, signed)
                forged_cursor = _encode_cursor(
                    registry_key,
                    endpoint=name,
                    query=query,
                    boundary_snapshot=signed,
                    offset=1,
                )
                refused = client.get(endpoint, params={**params, "cursor": forged_cursor})
>               assert refused.status_code == 404, (endpoint, field)
E               AssertionError: ('/v1/records', 'merkle_root')
E               assert 200 == 404
E                +  where 200 = <Response [200 OK]>.status_code

tests/test_registry.py:1025: AssertionError
__________________ test_shared_service_cursor_boundary_cases ___________________

tmp_path = PosixPath('/private/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/pytest-of-administrator/pytest-47/test_shared_service_cursor_bou0')

    def test_shared_service_cursor_boundary_cases(tmp_path: Path) -> None:
        vectors = _json("vectors/registry-service.json")
        pagination = vectors["pagination"]
        assert pagination["cursor_boundary_cases"], "no cursor_boundary_cases vectors"
    
        registry_key = signing.generate_key()
        store = Store(tmp_path / "cursor-boundary-registry.db")
        for item in vectors["records"]:
            store.append(
                registry_key.sign_record(dict(item["record"])),
                created_at="2026-07-13T00:00:00Z",
            )
        client = TestClient(
            create_app(store=store, signing_key=registry_key, tokens=AuditorTokens([]))
        )
    
        query = dict(pagination["query"])
        first = client.get("/v1/records", params=query)
        assert first.status_code == 200
        first_body = first.json()
        assert isinstance(first_body["next_cursor"], str) and first_body["next_cursor"]
        chain = signing.canonical_document_bytes(first_body["boundary"])
        records_query = {
            "source_identity": "",
            "commit": "",
            "content_sha256": query["content_sha256"],
            "limit": query["limit"],
        }
    
        # The chain is not re-evaluated after an append: reuse the shared
        # append_after_first_page flow as the control for every boundary case.
        appended = pagination["append_after_first_page"]
        store.append(
            registry_key.sign_record(dict(appended["record"])),
            created_at="2026-07-13T00:01:00Z",
        )
        assert client.get("/v1/snapshot").json()["log_size"] == (
            first_body["boundary"]["log_size"] + 1
        )
    
        for case in pagination["cursor_boundary_cases"]:
            assert case["name"] == "cursor-boundary-disagreement"
            assert case["reevaluate_at_newer_boundary"] is False
            # Control: the original cursor still serves the ORIGINAL boundary.
            continued = client.get(
                "/v1/records", params={**query, "cursor": first_body["next_cursor"]}
            )
            assert continued.status_code == 200
            assert [record["audit"]["case"] for record in continued.json()["records"]] == (
                pagination["expected_original_cursor_ids"]
            )
            assert signing.canonical_document_bytes(continued.json()["boundary"]) == chain
            # Disagreement: a carried boundary whose body differs from the store,
            # genuinely re-signed with the service key, is refused on /v1/records.
            forged = dict(first_body["boundary"])
            forged["head"] = _flip_boundary_hex(forged["head"])
            signed_forged = registry_key.sign_record(forged)
            assert signing.verify_signed(registry_key.public_pinned, signed_forged)
            bad_cursor = _encode_cursor(
                registry_key,
                endpoint="records",
                query=records_query,
                boundary_snapshot=signed_forged,
                offset=1,
            )
            refused = client.get("/v1/records", params={**query, "cursor": bad_cursor})
            assert refused.status_code == case["status"] == 404
            assert refused.json()["error"]["code"] == case["error"] == "invalid_cursor"
            # Same refusal on /v1/log, via the Merkle root this time.
            log_first = client.get("/v1/log", params={"since": 0, "limit": 2}).json()
            forged_log = dict(log_first["boundary"])
            forged_log["merkle_root"] = _flip_boundary_hex(forged_log["merkle_root"])
            signed_log = registry_key.sign_record(forged_log)
            assert signing.verify_signed(registry_key.public_pinned, signed_log)
            bad_log = _encode_cursor(
                registry_key,
                endpoint="log",
                query={"since": 0, "limit": 2},
                boundary_snapshot=signed_log,
                offset=1,
            )
            refused_log = client.get(
                "/v1/log", params={"since": 0, "limit": 2, "cursor": bad_log}
            )
>           assert refused_log.status_code == case["status"] == 404
E           assert 200 == 404
E            +  where 200 = <Response [200 OK]>.status_code

tests/test_protocol_conformance.py:446: AssertionError
=============================== warnings summary ===============================
../../../tmp/csk-review-27yepb-venv/lib/python3.14/site-packages/fastapi/testclient.py:1
  /tmp/csk-review-27yepb-venv/lib/python3.14/site-packages/fastapi/testclient.py:1: StarletteDeprecationWarning: Using `httpx` with `starlette.testclient` is deprecated; install `httpx2` instead.
    from starlette.testclient import TestClient as TestClient  # noqa

../../../tmp/csk-review-27yepb-venv/lib/python3.14/site-packages/starlette/testclient.py:53
  /tmp/csk-review-27yepb-venv/lib/python3.14/site-packages/starlette/testclient.py:53: DeprecationWarning: The anyio.abc.BlockingPortal alias is deprecated, use anyio.from_thread.BlockingPortal instead.
    _PortalFactoryType = Callable[[], AbstractContextManager[anyio.abc.BlockingPortal]]

-- Docs: https://docs.pytest.org/en/stable/how-to/capture-warnings.html
=========================== short test summary info ============================
FAILED tests/test_registry.py::test_cursor_carried_boundary_disagreement_refused_on_both_endpoints
FAILED tests/test_protocol_conformance.py::test_shared_service_cursor_boundary_cases
2 failed, 131 deselected, 2 warnings in 3.26s

EXIT=1
```

```text
MUTANT merkle-only: both equality gates compare only merkle_root; prefix availability remains enforced.
FF                                                                       [100%]
=================================== FAILURES ===================================
_____ test_cursor_carried_boundary_disagreement_refused_on_both_endpoints ______

tmp_path = PosixPath('/private/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/pytest-of-administrator/pytest-48/test_cursor_carried_boundary_d0')

    def test_cursor_carried_boundary_disagreement_refused_on_both_endpoints(tmp_path: Path):
        client, registry_key, token = _client(tmp_path)
        auditor_key = client.app.state.auditor_key  # type: ignore[attr-defined]
        content_hash = "sha256:" + "1f" * 32
        for index in range(3):
            record = auditor_key.sign_record(_body(name=f"skill-{index}"))
            assert client.post(
                "/v1/records", json=record, headers={"Authorization": f"Bearer {token}"}
            ).status_code == 201
        records_first = client.get(
            "/v1/records", params={"content_sha256": content_hash, "limit": 1}
        ).json()
        log_first = client.get("/v1/log", params={"since": 0, "limit": 1}).json()
        assert records_first["next_cursor"] and log_first["next_cursor"]
        # Control: the genuine cursors continue the chain at the original boundary.
        genuine_records = client.get(
            "/v1/records",
            params={
                "content_sha256": content_hash,
                "limit": 1,
                "cursor": records_first["next_cursor"],
            },
        )
        assert genuine_records.status_code == 200
        assert genuine_records.json()["boundary"] == records_first["boundary"]
    
        cases = [
            (
                "/v1/records",
                "records",
                {"source_identity": "", "commit": "", "content_sha256": content_hash, "limit": 1},
                {"content_sha256": content_hash, "limit": 1},
                records_first["boundary"],
            ),
            (
                "/v1/log",
                "log",
                {"since": 0, "limit": 1},
                {"since": 0, "limit": 1},
                log_first["boundary"],
            ),
        ]
        for endpoint, name, query, params, boundary in cases:
            for field in ("head", "merkle_root"):
                forged = dict(boundary)
                forged[field] = _flip_hex(forged[field])
                # The forged body is genuinely re-signed with a key the service
                # accepts, so only the store comparison can refuse it.
                signed = registry_key.sign_record(forged)
                assert signing.verify_signed(registry_key.public_pinned, signed)
                forged_cursor = _encode_cursor(
                    registry_key,
                    endpoint=name,
                    query=query,
                    boundary_snapshot=signed,
                    offset=1,
                )
                refused = client.get(endpoint, params={**params, "cursor": forged_cursor})
>               assert refused.status_code == 404, (endpoint, field)
E               AssertionError: ('/v1/records', 'head')
E               assert 200 == 404
E                +  where 200 = <Response [200 OK]>.status_code

tests/test_registry.py:1025: AssertionError
__________________ test_shared_service_cursor_boundary_cases ___________________

tmp_path = PosixPath('/private/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/pytest-of-administrator/pytest-48/test_shared_service_cursor_bou0')

    def test_shared_service_cursor_boundary_cases(tmp_path: Path) -> None:
        vectors = _json("vectors/registry-service.json")
        pagination = vectors["pagination"]
        assert pagination["cursor_boundary_cases"], "no cursor_boundary_cases vectors"
    
        registry_key = signing.generate_key()
        store = Store(tmp_path / "cursor-boundary-registry.db")
        for item in vectors["records"]:
            store.append(
                registry_key.sign_record(dict(item["record"])),
                created_at="2026-07-13T00:00:00Z",
            )
        client = TestClient(
            create_app(store=store, signing_key=registry_key, tokens=AuditorTokens([]))
        )
    
        query = dict(pagination["query"])
        first = client.get("/v1/records", params=query)
        assert first.status_code == 200
        first_body = first.json()
        assert isinstance(first_body["next_cursor"], str) and first_body["next_cursor"]
        chain = signing.canonical_document_bytes(first_body["boundary"])
        records_query = {
            "source_identity": "",
            "commit": "",
            "content_sha256": query["content_sha256"],
            "limit": query["limit"],
        }
    
        # The chain is not re-evaluated after an append: reuse the shared
        # append_after_first_page flow as the control for every boundary case.
        appended = pagination["append_after_first_page"]
        store.append(
            registry_key.sign_record(dict(appended["record"])),
            created_at="2026-07-13T00:01:00Z",
        )
        assert client.get("/v1/snapshot").json()["log_size"] == (
            first_body["boundary"]["log_size"] + 1
        )
    
        for case in pagination["cursor_boundary_cases"]:
            assert case["name"] == "cursor-boundary-disagreement"
            assert case["reevaluate_at_newer_boundary"] is False
            # Control: the original cursor still serves the ORIGINAL boundary.
            continued = client.get(
                "/v1/records", params={**query, "cursor": first_body["next_cursor"]}
            )
            assert continued.status_code == 200
            assert [record["audit"]["case"] for record in continued.json()["records"]] == (
                pagination["expected_original_cursor_ids"]
            )
            assert signing.canonical_document_bytes(continued.json()["boundary"]) == chain
            # Disagreement: a carried boundary whose body differs from the store,
            # genuinely re-signed with the service key, is refused on /v1/records.
            forged = dict(first_body["boundary"])
            forged["head"] = _flip_boundary_hex(forged["head"])
            signed_forged = registry_key.sign_record(forged)
            assert signing.verify_signed(registry_key.public_pinned, signed_forged)
            bad_cursor = _encode_cursor(
                registry_key,
                endpoint="records",
                query=records_query,
                boundary_snapshot=signed_forged,
                offset=1,
            )
            refused = client.get("/v1/records", params={**query, "cursor": bad_cursor})
>           assert refused.status_code == case["status"] == 404
E           assert 200 == 404
E            +  where 200 = <Response [200 OK]>.status_code

tests/test_protocol_conformance.py:428: AssertionError
=============================== warnings summary ===============================
../../../tmp/csk-review-27yepb-venv/lib/python3.14/site-packages/fastapi/testclient.py:1
  /tmp/csk-review-27yepb-venv/lib/python3.14/site-packages/fastapi/testclient.py:1: StarletteDeprecationWarning: Using `httpx` with `starlette.testclient` is deprecated; install `httpx2` instead.
    from starlette.testclient import TestClient as TestClient  # noqa

../../../tmp/csk-review-27yepb-venv/lib/python3.14/site-packages/starlette/testclient.py:53
  /tmp/csk-review-27yepb-venv/lib/python3.14/site-packages/starlette/testclient.py:53: DeprecationWarning: The anyio.abc.BlockingPortal alias is deprecated, use anyio.from_thread.BlockingPortal instead.
    _PortalFactoryType = Callable[[], AbstractContextManager[anyio.abc.BlockingPortal]]

-- Docs: https://docs.pytest.org/en/stable/how-to/capture-warnings.html
=========================== short test summary info ============================
FAILED tests/test_registry.py::test_cursor_carried_boundary_disagreement_refused_on_both_endpoints
FAILED tests/test_protocol_conformance.py::test_shared_service_cursor_boundary_cases
2 failed, 131 deselected, 2 warnings in 2.98s

EXIT=1
```

## Reused evidence and limits

Read `TASK-260910-27yepb_change-request_rev1-validation.log`: hosted `sh scripts/remote-gate.sh` exit 0, run 35234396569, all six Python 3.11/3.14 × OS combinations, strict mypy, distribution build and Docker build success; coverage unit reports 1/1 required command shard green (test-case coverage unknown). Accepted this runtime-attached evidence; did not rerun the hosted gate or build locally. Locally reran Python 3.14.6 only. Producer reports 133 tests and 13 mypy source files, independently reproduced. Its statement about previously undriven rejection names is consistent with the checkpoint tests; existing ad hoc cursor tests existed, while the vector-name-driven harness is new.

No spec gap or product/architecture decision found. No code modified in the managed worktree and no files added there. Final git diff against candidate empty and status unchanged. No new substantive finding requiring a separate logbook entry; logbook CLI/tool is unavailable in this run. All review observations are persisted here. `task-board spawn goal` reports this run is not goal-bound; no directives present.

Acceptance evidence is this resource. Route with accept_cr revision=1 to integrating; reviewer does not commit, supply commit_ack or set done.
