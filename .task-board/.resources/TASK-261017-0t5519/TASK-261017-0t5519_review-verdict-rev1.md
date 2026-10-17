# Review verdict, revision 1

Task: TASK-261017-0t5519 — windows-test-duration-analysis.
Candidate: CR-TASK-261017-0t5519-1 revision 1; base b2a4c13321429e82148b07c886765a788de16495; tree f23fd51b4b11024fcf0394fae00060e4137d68b5.

## AC coverage checked before source inspection

4/4 research AC rows have named driving checks: AC1 `review_checks.py` hosted metadata/hash/CSV reconciliation and fresh Windows artifact spot checks; AC2 pinned-source cause audit plus event-scheduling/cache checks; AC3 numerical bounds and coverage-preservation audit; AC4 sizing-versus-measurement audit. No product test requirement is invented for a read-only research leaf. The brief supplies neither a catalog surface table nor numerical review/free-hunt budgets; the following task-specific surfaces cover every supplied review criterion.

## Surface sweep

| Row | Result | Attack and evidence |
| --- | --- | --- |
| R1 provenance and timing completeness | held | Fresh GitHub metadata establishes success/main and matching SHA for three runs and all nine jobs. All nine original stream hashes match the attached manifest. Fresh C Windows artifact also matches. Re-ran all 11,824 CSV rows against raw events; missing-terminal and forged-duration controls fail with intended assertions. Three independent timing spot checks: CLI 2841.923s, compiled-status 450.64s, activation matrix 262.98s. Job totals independently recomputed: Windows 3429/2981/3122s, macOS 1052/1315/965s, Linux 140/158/126s. See review-checks log. |
| R2 causes versus unsupported attribution | held | Fresh C events reproduce CLI serial phase 2301.881936s and entirely serial envprofile; CLI ends 315.305835s after envprofile. Fresh C log confirms exact cache hit and both gate exits 0. Pinned C source inspection confirms 36 activation combinations, real install/Git fixtures, operation-private probe layout, subprocesses and toolchain fingerprinting. OS-level CPU/I/O/Defender fractions are explicitly unknown rather than inferred from slow executables. |
| R3 ranked options and coverage | held | Independent C greedy-bin replay reproduces all six reported shard-bin values; faster-runner arithmetic reproduces 28.35–36.24m. Options distinguish estimates from measured gains, preserve native Windows feature coverage, name PR deferrals, and explicitly warn that these are per-package bins rather than a global four-worker schedule. No silent Linux substitution or verdict-cache reuse is proposed. |
| R4 sizing recommendation | held | Challenged the recommendation against absent utilization/RSS/disk telemetry: report correctly labels 8 vCPU/32 GB/250 GB as a starting pilot, not a measured minimum or guaranteed acceleration; it gives a cheaper pilot and requires three unchanged-suite samples. Serial bottlenecks support prioritizing single-thread and disk performance over merely increasing cores. |
| R5 empty delta and public hygiene | held | Exact base-to-tree diff is empty and repository status is clean. Read-only analysis explicitly calls for this outcome: changes would exceed scope. Inspected all nine authored outcomes; pattern scan flagged six synthetic public test-name paths (config operator-path cases and Unicode home-path cases), not private machine paths. Manual classification prevents treating a scanner hit as a leak. No private host, person, credential or local operator path was found in the authored deliverables. |

## Free hunt and limits

No additional blocking finding reproduced. Checked parent/subtest double counting, skip-versus-pass treatment, scheduling wall time versus Elapsed, cross-SHA sample comparability, heavy-package bottleneck movement, and overly optimistic sharding interpretation. The report explicitly bounds all of them. Hosted stages are reused evidence, not reviewer-rerun product tests. Nine original streams were hash-verified from existing scratch; C Windows was downloaded anew. Fresh API metadata and C job logs were independently fetched. Source was inspected at C for fixture/probe claims; no new product build or Windows benchmark was run.

Reproductions run by reviewer: `python3 .temp/review-windur/review_checks.py` (exit 0), including original `verify.py` green exit 0 and negative modes `missing-terminal`/`wrong-timing` exit 1 with intended assertions. Independent bin replay and runner arithmetic are in the options log. The initial hygiene pattern scan produced an expected-to-investigate false positive; the six matching names are synthetic test fixtures, manually resolved, not suppressed evidence of a private path. Git exact-tree diff and final status separately checked.

## Verdict

```json
{
  "verdict": "accepted",
  "revision": 1,
  "findings": [],
  "notes": [
    "No surface table or numeric review budgets were supplied; all five concrete research-review surfaces were swept and held.",
    "Sizing and speedup figures remain conditional planning estimates; OS causal attribution requires a future profiled pilot.",
    "No repository change is the correct result for this expressly read-only research leaf. Acceptance does not claim integration or landing."
  ]
}
```

Recommendation: accept revision 1. This review asserts only the checks and attacks recorded above, not proof of absence of all errors.
