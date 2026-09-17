# TASK-260910-3p2rbh review verdict — revision 1

Verdict: **changes_requested**, route **to-dev**. No candidate files changed.

Reviewed candidate tree `736fd87e1bbe49aa0a5990276be124fe9c0ffa1e`, base `aea81ccf298071a7746d277bea316bdc91309c43`; current leaf delta also inspected against checkpoint HEAD. Published patch SHA-256 verified: `8579d3e3b1bed82a816028d3fbde003d9dd019f752aa948c2a1874dd539ca90e`. `git diff --quiet 736fd87 -- .` returned 0 before/after review. Worktree status unchanged. All installs, tests, mutations, and artifacts were outside the managed worktree. Existing worktree caches were not touched.

## Required corrections

1. **High: valid concurrent append permanently trips integrity failure.** `src/csk_registry/store.py:1290-1305` reads frontier length/live size in one statement, then `_load_frontier(conn)` in another snapshot. Append between these reads makes `live_size == walked_size` refer to the old state but frontier tails refer to the new state. `_boundary_cache_errors` (`:1717`) reports corruption; refresh latches it (`:1379`), disabling writes until restart on a perfectly valid database. Deterministic reproduction interposes a real `store.append` immediately before the verifier's `_load_frontier`; `integrity_errors()` remains empty, yet refresh returns `merkle frontier disagrees with the committed log`. Fix cross-table snapshot consistency while preserving WAL append concurrency; add an entry-point refresh regression with this interleaving, not just pure comparator tests.

2. **High: incremental advancement attests a corrupted frontier and commits a wrong immutable boundary.** `_append_locked` (`store.py:679-703`) validates only frontier length/encoding before trusting tails, then advances the cached verified head. Change length-preserving level-0 tails after three entries and append a fourth: append succeeds, `/health` is 200, and the persisted boundary root differs from `_merkle_root` over the actual committed log. This is inherited frontier trust now used as justification for the new verified-head attestation; the brief explicitly requires incremental frontier consistency and that appends cannot mark corrupted state verified. Validate the needed frontier/head anchors cheaply before use, latch mismatch via the common failure path, and test corruption followed by an append through production entries. Do not fix by returning to full-chain hashing per request.

3. **Medium: rollback leaves the cached verdict ahead of committed state.** `_append_locked` updates `_health_verified_head/_log_size` before COMMIT (`store.py:702-703`); `_write_transaction` (`:507-516`) rolls back only SQLite. A BEFORE INSERT trigger on `imported_records` that raises ABORT after the log append reproduces: `append_imports` raises, durable head is size 0, but the ready verdict claims verified size 1 and its nonexistent head. Publish cached advancement only after successful transaction commit, or restore it on every failure, including a later import item, ledger insertion and COMMIT failure. Cover rollback via real `append_imports`/idempotent append call sites. Profile §3 atomicity and §9 verified-head agreement apply.

4. **Medium: stale verdict recovers without restart, contrary to the assigned settled behavior.** `store.py:1225-1253` treats staleness as a transient predicate; a successful refresh (`:1401-1406`) restores readiness. The reviewer brief explicitly says stale verdicts disable writes through the integrity path and readiness stays down until restart; the producer brief also says stays non-ready until restart. Frozen-clock reproduction at interval 10, age 20.001 gives HTTP 503 and write refusal, then refresh becomes ready without restart. The committed test at `tests/test_registry.py:1692-1721` deliberately protects the opposite behavior and uses a sleep. Implement the required latch, replace that expectation, and test deterministic time boundaries plus late completion of a stalled verifier. Restart must re-verify before recovery.

## Acceptance review

| Requirement | Evidence / outcome |
| --- | --- |
| Cached health; zero chain/Merkle/canonical work | PASS: `app.py:201`, `store.py:1230`; committed operation counters pass for both probe batches. |
| First verdict comes from startup verification | PASS: `store.py:268-278`; committed startup test passes. |
| Background lifecycle and no full-walk write lock | PARTIAL: app lifespan starts/stops daemon; refresh uses its own query-only SQLite connection without store lock, publishing under `self._lock`. WAL walk does not hold the app write lock, but inconsistent statement snapshots cause finding 1. |
| Failed refresh latches; corruption and restart repair | Existing committed corruption/boundary/frontier and repair tests PASS; incremental corruption and rollback have findings 2/3. |
| Staleness fails closed and stays down until restart | HTTP 503 and write refusal reproduced at age 20.001; restart-only latch FAILS (finding 4). Exact-bound and stalled-completion committed coverage missing. |
| Existing recovery / restore / concurrency conformance | PASS in full independent suite, including 8/8 recovery vector rows. They do not exercise the new verifier interleaving. |
| Incremental verified head | Happy path passes; corruption/rollback invariants FAIL (findings 2/3). |
| Audit refresh outcomes | Existing success/failure log tests pass; structured duration/log_size/result present. |
| Docs and scope | README interval/env/staleness guidance, SECURITY paragraph, compose probe comment, completed CHANGELOG R2 and frozen schema unchanged. SECURITY/compose do not themselves spell out env/default; align operations wording with corrected latch. No spec edits. |
| Independent pytest/mypy | PASS: 150 tests; strict mypy 13 source files. |
| Negative evidence | 2/2 effective narrowing mutants killed by committed assertions; 5/5 reviewer attack assertions fail on original candidate, demonstrating four defects (two attacks cover finding 2). This is targeted coverage, not exhaustive concurrency or corruption proof. |

## Validation commands and results

Shell: zsh, `set -o pipefail`. Interpreter `/tmp/csk-review-3p2rbh-venv/bin/python`, Python 3.14.6. CWD `/tmp/csk-review-3p2rbh`, populated by `git archive 736fd87e1bbe49aa0a5990276be124fe9c0ffa1e`. Venv outside tree; editable install targets only disposable copy.

```
export CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1
/tmp/csk-review-3p2rbh-venv/bin/python -m pytest -q
150 passed, 2 warnings in 11.72s
exit 0
/tmp/csk-review-3p2rbh-venv/bin/python -m mypy
Success: no issues found in 13 source files
exit 0
```

Full transcripts attached as `TASK-260910-3p2rbh_review-transcripts-rev1.log`. Pre-existing Starlette/httpx and anyio deprecation warnings only. `git diff --check` exit 0. No separate linter configured; mypy is not claimed to be a linter.

Hosted gate NOT rerun. Read published `TASK-260910-3p2rbh_change-request_rev1-validation.log`: runtime gate exit 0, GitHub run 35249305219, six Python 3.11/3.14 OS jobs plus mypy/build/docker successful. That evidence is accepted for those environments; independently ran Python 3.14.6 only.

Environment discrepancy: spec checkout is now `684c9f1`, not the brief's `dced9b8`. Read-only comparison against dced9b8 shows no changes in the cited registry protocol/profile, service/client/behavior vectors or records/log v2 schema cases; intervening conformance changes concern manager/system configuration. Recorded actual checkout rather than falsely claiming the whole tree is pinned. Shared suite was run with the exact required CURATOR_CONFORMANCE_ROOT above.

## Narrowing mutants (disposable copy only)

- M1: narrow staleness to `verified_log_size > 1` as well as age; committed `test_health_stale_verifier_fails_closed_and_refresh_recovers` fails at HTTP 200 != 503 (exit 1).
- M2: retain only refresh errors containing `wrong entry hash` before verdict publication; committed `test_health_refresh_detects_boundary_cache_tampering` fails because ready=True (exit 1). This narrows the corruption class to chain hashes rather than deleting the whole gate.
- An initial M2 placement produced AttributeError instead of the intended semantic failure; it was discarded and replaced by the effective filter above. It is not counted in the 2/2 ratio.

Source restored from saved original after each mutant. Attached reproduction tests contain no product-code changes; copy them into the disposable candidate's `tests/test_review_attacks.py` and run:

```
/tmp/csk-review-3p2rbh-venv/bin/python -m pytest -q tests/test_review_attacks.py
5 failed, 2 warnings in 1.34s
exit 1
```

These expected review failures prove the deficiencies; they are not passing regression tests. Producer should retain/adapt them and make them pass. No sleeps used in review attacks.

## Routing

Ordinary implementation rework, no external blocker or human decision required. `task-board spawn goal "$TASK_BOARD_RUN_ID"` reported not goal-bound. Persist verdict, reproduction, transcripts, and task-scoped logbook before `set_status(..., status=to-dev)`. No acceptance, commit_ack, done transition, code edits, commits, or hosted gate invocation by reviewer.
