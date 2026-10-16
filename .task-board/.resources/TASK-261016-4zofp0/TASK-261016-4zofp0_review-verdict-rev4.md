# SNAP-LOCK revision 4 review verdict

**ACCEPT**, capture-stage scope. Blocking findings: **0**. Acceptance does not claim landing or completion of SNAP-CLOSURE/SNAP-CLI.

Element: TASK-261016-4zofp0 — implement-rc7-snapshot-nonrecovering-capture.
Base: `c29663024abfe6dc0353749018fd65b2bf1833e1`.
Candidate tree: `24299c092bc02d004545ff7103ffc192b541db37` (17 changed paths; repository delta present).

## Verdict data

```json
{
  "verdict": "accepted",
  "revision": 4,
  "candidate_tree": "24299c092bc02d004545ff7103ffc192b541db37",
  "findings": [],
  "notes": [
    "Capture-stage scope: 5 of 5 AC rows, 1 of 1 credential-store convergence requirement and 2 of 2 release-policy convergence properties driven; 0 of 2 CLI forms qualified.",
    "Revision 2 is ACCEPT with no findings; the generic rejection wording is inaccurate.",
    "No numeric review budget or catalog surface table supplied; all five mandatory paired rows swept.",
    "Callback bounds are cooperative; arbitrary callback blocking is not preempted.",
    "Windows POSIX sentinel, permission/mode and held-object replacement limits remain declared; local attacks ran on darwin/arm64."
  ],
  "surface_results": [
    {
      "row": "R1",
      "result": "held",
      "evidence": "TASK-261016-4zofp0_review-execution-rev4.md"
    },
    {
      "row": "R2",
      "result": "held",
      "evidence": "TASK-261016-4zofp0_review-execution-rev4.md"
    },
    {
      "row": "R3",
      "result": "held",
      "evidence": "TASK-261016-4zofp0_review-execution-rev4.md"
    },
    {
      "row": "R4",
      "result": "held",
      "evidence": "TASK-261016-4zofp0_review-execution-rev4.md"
    },
    {
      "row": "R5",
      "result": "held",
      "evidence": "TASK-261016-4zofp0_review-execution-rev4.md"
    }
  ],
  "free_hunt": []
}
```

## Scope and evidence

Coverage inspected before implementation review: **5 of 5 capture-stage AC rows driven**, **1 of 1 credential-store convergence requirement**, **2 of 2 release-policy convergence properties**. Named tests and production call sites are in `TASK-261016-4zofp0_rev4-coverage-map.md`; independently executed commands are in `TASK-261016-4zofp0_review-execution-rev4.md`.

The five paired rows are pending/recovery/non-provisioning, bounded lock/fresh selection, cooperating writers/torn generations, object identity/ABA/policy drift, and metadata-only vendor boundaries. Every row was attacked and recorded before starting the next row. Held means those attacks did not reproduce, not proof of absence.

Revision-4 supplied-observation and cache behavior was checked first. Capture and status reach `verifyHomeContent`; scoped observations cover assembly and passthrough policy derivation. `computeCredentialPolicy` remains the single derivation implementation. Cache success/error reuse checks profile, environment and supplied release. Explicit-store refusals occur before a capture callback; native policy and final verification retain source diagnostics. Capture intentionally verifies before and after the callback; no whole-invocation exactly-once assertion is made.

All **12 narrowing plants** were independently rerun in an isolated copy, each with expected exit 1 at the intended behavioral assertion. The harness exited 0. M-policy-probe ran the vendor executable; M-release-cache reached the forbidden capture callback/reused the wrong isolation projection; M3-lock-only accepted the equal-fragment ABA witness. The other nine plants also failed at their intended assertions. The restored copy and all 17 working candidate blobs match the exact tree. See `TASK-261016-4zofp0_review-mutant-logs-rev4.md` and the producer's exact harness, independently replayed with only its scratch output location changed.

Free hunt: late corrupt journal injection, repair/dry-run/preserve-lock flags, and a new native-selector mutation during the capture callback all refused with the intended diagnostic, no output and unchanged planted state beside valid controls. No blocking finding. Reviewer-only tests are attached as `TASK-261016-4zofp0_review-extra-rev4_test.go`; no product code was modified.

Normative source independently read: curator-spec main `1059b94b4135089d842bf017070b8ab47d79a536`, snapshots 1/5, environments 9.2/10.1, CIP-0014 consistency 1–5, CIP-0015 machine credential selection. No contradiction found in the capture-stage delta. The scope exceptions are justified shared verification/status, policy, managerlock and transaction seams, platform fixtures, README and signed gate snapshots. No deferred CLI or closure implementation is claimed.

## Independent validation

- Five row-specific focused test commands: exit 0 each, with `-count=1 -timeout=3m`.
- CredentialProjection/CredentialStorePolicy trunk regressions: exit 0.
- Supplemental reviewer overlays: exit 0; native metadata changed during the callback is re-observed and refused without repair.
- Twelve plant commands: expected exit 1 each, named failing assertions; harness exit 0.
- `git diff --check`: exit 0.
- `sh scripts/remote-gate.sh`: actual process **exit 0**. [Reviewer hosted run](https://github.com/relux-works/curator/actions/runs/38972219571) passed all **20 enabled jobs**, with two explicitly optional jobs skipped. Signed snapshot `98452a23e13ca4521353cbe3e26207755dec2170`, signature `G`, exact candidate tree. See `TASK-261016-4zofp0_review-gate-rev4.md`.

No prior green substitutes for these reruns. Windows skipped POSIX-specific cases remain bounds rather than passing runtime evidence. The cooperative local callback contract, closure inventory and actual CLI publication boundaries remain explicit.

Fresh upstream observation at review matched the base. Final working blobs match; Story HEAD and index are unchanged. Live checklist read returned 33/33 checked. Reviewer goal query returned no bound goal. Record acceptance with `accept_cr` revision 4, never reviewer `done` or `commit_ack`. Delivery owner retains PR/check/landing obligations.

Logbook: `TASK-261016-4zofp0_review-logbook-rev4.md`.
