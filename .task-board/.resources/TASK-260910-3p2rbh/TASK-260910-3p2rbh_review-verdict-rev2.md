# TASK-260910-3p2rbh — revision 2 review verdict

Verdict: **accepted**. No blocking findings. Route with accept_cr(revision=2); integration remains producer-owned.

Candidate tree: d1fb1e405b1e6bf96cc8868f44d24de71b2756cf; base aea81ccf298071a7746d277bea316bdc91309c43. Published patch SHA-256 independently matched 9abbc92e55405a1a4e867f2b1628477abe1a969fb529ddabfc568c950b53a4f4. `git diff <candidate> --exit-code` returned 0; no untracked files. Reviewed and tested an archive of that exact tree at /tmp/csk-review-r2-copy. Candidate worktree was never modified.

Read campaign rules, briefs/re-decision, producer results including Revision 2, prior boundary-task results, published patch and hosted validation, protocol registry sections 5/6, service profile sections 2/5/9/11, R2 finding, recovery/transaction/restore vectors. The revision-2 transient-staleness decision supersedes the original latch-on-staleness wording.

| Review item | Evidence and conclusion |
|---|---|
| Cached health and initial verdict | app.py:201 calls only health_verdict; store.py:273 establishes startup verdict. Committed test at test_registry.py:1564 measures zero SHA/CCJ/Merkle operations over both five-probe batches, including probes after appends. Startup-verdict test passes. |
| Failed/stale refresh and write refusal | store.py:638,1346,1351,1526 enforce stale refusal and latched failure. Deterministic committed tests at test_registry.py:1702,1755 pass. Reviewer additionally blocked a real verifier thread with Events, advanced injected time to age=20 and 20.001, checked /health and append refusal, then released success/failure: both pass. Success recovers; failure stays latched on subsequent refresh. |
| Snapshot consistency correction | store.py:1492 begins one read transaction across chain, ledgers, boundaries and frontier; rollback ends it. The walk takes no store lock, and WAL permits appends. Publication briefly takes the store RLock. Original reviewer reproduction interposes a real append before frontier load: green and no integrity errors. |
| Frontier anchors correction | store.py:744 checks last boundary/head, last leaves, inter-level links and root before insertion, latching through store.py:645. Original length-preserving level-0 tamper at size 3 now refuses append; committed endpoint-health assertion passes. No full-chain request work reintroduced. |
| Post-commit correction | store.py:672,975,1088 publish only after successful transaction exit. Original imported_records ABORT-trigger reproduction passes; committed idempotency ABORT-trigger regression passes. Reviewer additionally replayed an earlier idempotency key after another append and verified the cached head neither advances falsely nor rewinds. |
| Corruption and recovery | Committed log-hash, boundary-cache, frontier, repair/restart tests pass. Reviewer changed canonical record JSON after startup: cached 200 before refresh, then 503 and append refusal after explicitly driven refresh. |
| Validation and conformance | Independent full suite: 155 passed, including vector-driven recovery, concurrency and restore; strict mypy clean. Hosted matrix/build/Docker evidence accepted from attached runtime log, not rerun. |
| Operations/scope | README:173 documents flag, env, default, interval sizing, stale recovery and failure latch; SECURITY:32 documents split; compose:12 notes cheap probes/default; CHANGELOG:36 completes R2 with dced9b8. Frozen health schema and protocol/profile untouched. No configured separate linter. |

## Independent validation

Shell /bin/zsh; interpreter /tmp/csk-venv/bin/python, Python 3.14.6; cwd /tmp/csk-review-r2-copy. Existing external venv reused; pytest's configured src path loads the disposable candidate. All shell test calls used `set -o pipefail` and explicit captured exit codes.

```
export CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1
/tmp/csk-venv/bin/python -m pytest -q
155 passed, 2 warnings in 11.93s
exit=0
/tmp/csk-venv/bin/python -m mypy
Success: no issues found in 13 source files
exit=0
/tmp/csk-venv/bin/python -m pytest -q tests/test_review_attacks.py -k 'concurrent_append_between or append_does_not or rollback_does_not'
3 passed, 2 deselected, 2 warnings in 0.93s
exit=0
/tmp/csk-venv/bin/python -m pytest -q tests/test_review_r2_extra.py
4 passed, 2 warnings in 1.36s
exit=0
```

The two deselected prior attacks encode superseded stale-latching policy and the old successful-corrupt-append shape, respectively. They are not claimed green. New actual-thread stale tests use the revised policy. Two warnings are dependency deprecations. Python 3.11 and other OSes were not independently rerun; runtime validation log reports six OS/Python jobs, strict mypy, Docker and distribution builds success, exit 0 (run 35252150306).

## Narrowing mutants and coverage bounds

Mutations only in disposable copy; original bytes restored in finally blocks.

1. Stale refusal narrowed from age > 2×interval to age > 4×interval. Committed test_health_stale_verifier_fails_closed_and_refresh_recovers fails (expected 503, got 200), exit 1.
2. Refresh cache failures narrowed to frontier errors only, dropping boundary disagreement. Committed test_health_refresh_detects_boundary_cache_tampering fails, exit 1.
3. Additional exploratory mutant narrows level-0 tail comparison to even sizes: survives, exit 0, because inter-level hash consistency independently catches the same injected size-3 corruption. This is redundant enforcement for that fixture, not evidence of missing corruption refusal.

Measured: 2/2 selected independent narrowing obligations killed; 2/3 total exploratory mutants killed. This is not exhaustive guard mutation coverage. Four rework corrections verified; no blanket claim of exhaustive concurrency interleavings. Producer states an upper odd-level second-last-tail detection limitation; full refresh compares all tails. No reproduced corrupt-boundary advancement remains from the required size-3 attack.

Full background walks have no execution deadline and pin a WAL snapshot until completion; readiness has the explicit stale bound, and shutdown joins boundedly. Accepted design uses the brief's snapshot alternative rather than holding the write lock for a full walk. External DB replacement/checkpoint enforcement remains the separate R3/P2 task.

Evidence: TASK-260910-3p2rbh_review-transcripts-rev2.log; TASK-260910-3p2rbh_review-attacks-rev2.py; TASK-260910-3p2rbh_review-logbook-rev2.md. Goal query reported this run is not goal-bound.
