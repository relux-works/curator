# Review verdict: TASK-261017-yyg01d — preservation-triage

Accepted CR revision 1. No product files changed: this is the correct outcome for the explicitly read-only reproduction/spec-classification/fix-brief scope. Acceptance does not claim any product defect fixed or any landing completed.

Candidate base: fcd785ba93c5d00e2e80615f8add2baa93ccd73d; candidate tree: f1771911a5f3a0e6e4ae59e6fcf68113caa98f27. Exact tree diff is empty. Fresh implementation origin/main: 541d5e8ec32d84d71b902b54814e1d8bb26f1f01 (only board additions since base). Source/spec observations remain applicable. Spec producer pin: 1059b94b4135089d842bf017070b8ab47d79a536; fresh spec main: 9dfc557d247243e0e8ba60fb877afe6d1facfd06. Governing five files are byte-identical between these spec revisions.

AC coverage checked before code inspection: 4/4. AC1 driven by reproduce.py, supplement.py, verify-observations.py; AC2 by pinned clause comparison with captured results; AC3 by three bug/spec-gap briefs with named positive/negative controls and packages/risk; AC4 by the ordered priority list. Document ACs do not require production tests. The task asks for fix leaves as briefs; the deliverable explicitly does not pretend new board tasks or implementation exist.

Reviewer commands (all exit 0 unless specified):

```sh
go build -o .temp/TASK-261017-yyg01d-review/curator ./cmd/curator
python3 .temp/TASK-261017-yyg01d-review/reproduce.py
python3 .temp/TASK-261017-yyg01d-review/supplement.py
python3 .temp/TASK-261017-yyg01d-review/verify-observations.py
python3 .temp/TASK-261017-yyg01d-review/review_validate.py
```

The first three scripts are exact copies of the task's attached script resources under their documented replay filenames. Reproduction evidence includes all individual command exits and before/after inventories. 36 subprocesses: 32 exit 0; four intended exit 1 (B manifestless install, E beta switch, E alpha retry, C native collision). Fresh scratch HOMEs are mode 0700 with containment checked before every subprocess; configured native/XDG/TMP roots stay inside them. No real home was passed to Curator. This is configuration and observed fixture evidence, not an OS-level audit of arbitrary filesystem/network access.

The reviewer verification script initially expected 39 commands and failed before asserting behavior. Its corrected count is 36, enumerated from the complete command records. That initial setup failure is retained in the log and not counted as passed. A separate bounded control plant changed the captured native preservation hash; the producer observation verifier failed at the exact preservation equality assertion (exit 1), demonstrating that it detects this false preservation claim. No source mutants were needed for a research-only leaf.

Review B's A rerun requirement is met; all other supplied scenarios were rerun too. All seven bounded rows held. Findings describe flaws in the research deliverable, not the already-reported product bugs. The review claims only that the listed attacks did not disprove the bounded triage claims.

```json
{
  "verdict": "accepted",
  "revision": 1,
  "findings": [],
  "notes": [
    "Research acceptance only: A and C2 remain demonstrated product defects; B remains a format conflict needing coordinated follow-up.",
    "No catalog surface table was supplied. The seven review rows below cover the concrete brief, not the full catalog.",
    "Six producer Go tests were inspected as attached evidence, not rerun by reviewer. All CLI reproduction scripts and observation verifier were rerun.",
    "Current implementation main advanced only in board paths; governing spec files are unchanged on fresh spec main.",
    "Future named tests and narrowing mutants are implementation requirements, not executed mutation evidence."
  ],
  "surface_results": [
    {
      "row": "safety-hygiene",
      "result": "held",
      "evidence": "Explicit isolated environment, 0700 HOME assertions for all 36 subprocesses; attached payload inspection."
    },
    {
      "row": "A",
      "result": "held",
      "evidence": "Fresh build, three deleted unmanaged entries, dangling native link, managed positive control, exit 0."
    },
    {
      "row": "B",
      "result": "held",
      "evidence": "No-manifest check 0/install 1; minimal-manifest check/refresh/install 0; marker schema conflict verified."
    },
    {
      "row": "C",
      "result": "held",
      "evidence": "Selected collision loss, unrelated instructions preserved, native refusal preserves bytes, takeover backup exact."
    },
    {
      "row": "D",
      "result": "held",
      "evidence": "First installation activates without --use; later installation remains idle."
    },
    {
      "row": "E",
      "result": "held",
      "evidence": "Hash checks of all four roots, partial persistence, recorded-current unchanged, retry and takeover."
    },
    {
      "row": "spec-fix-coverage",
      "result": "held",
      "evidence": "Five pinned spec files verified against Git objects; each bug has a bounded fix brief with controls/packages/risk; A first."
    }
  ],
  "free_hunt": {
    "findings": [],
    "attacks": [
      "Public-path/identity scan and manual inspection",
      "Upstream code/spec delta comparison",
      "Log completeness and individual exit enumeration",
      "Control plant changes native-refusal preserved digest; verifier rejects at intended assertion, exit 1"
    ]
  }
}
```
