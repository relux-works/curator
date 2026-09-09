# Review verdict — TASK-260909-3d1589, CR revision 1

Verdict: **changes requested**. Route: **to-dev**. No acceptance or integration authorization.

## R1 — medium: published manual flow masks setup and test failure

Location: TASK-260909-3d1589_adoption.md, first bash block (archive/copy/test sequence, lines 36–54).
Repeat-of: TASK-260909-xtvqf3 review rev2 R1, manual-adoption execution class. The previous deleted-destination bug is repaired; this is a remaining fail-open behavior in the same delivery surface.

Executed the exact block extracted from the published Markdown, without editing its commands, in a fresh task-local Git clone sharing the source object database. The positive run exits 0 (8 named diagnostics tests and 2 framing tests). Repeating against its populated destination exits 1 and preserves it.

Then attacked the manual entry itself, preserving previous output by renaming reviewer-owned scratch directories:

1. Remove only the conformance attachment from the scratch attachment directory; keep framing present. The exact block reports `cp: ... No such file or directory`, diagnostics reports `[no tests to run]` with exit 0, framing passes, and the **whole block exits 0**.
2. Restore that attachment and inject `t.Fatal("review injected baseline failure")` at TestGateCoverageCounts entry in the scratch copy. The exact block reports that named diagnostics failure, continues to the passing framing command, and **again exits 0**.

Thus setup failure/absence and a real failing gate are both represented as successful manual completion. No production source was changed to demonstrate this. The automatic runner correctly rejects the equivalent failures; it does not protect the separate manual route.

Required: make every manual setup/extraction/copy/test failure terminate nonzero, check archive pipeline failures, and reject zero/missing named test selection. Execute the newly published block verbatim on fresh destinations with positive, missing-overlay, zero-selection, failing-diagnostics, and stale-destination cases. Preserve prior data and the currently passing automatic runner. A shell failure mode alone does not reject Go's successful zero-test selection.

## Confirmed work and independent evidence

| Check | Exit | Result |
| --- | ---: | --- |
| Exact manual block, fresh destination | 0 | 8 diagnostics + 2 framing named tests pass |
| Exact manual block, populated destination | 1 | Correct refusal; previous output retained |
| Exact manual block, missing conformance overlay | 0 | Incorrect success; no diagnostics tests ran |
| Exact manual block, failing diagnostics test | 0 | Incorrect success; final framing pass masks failure |
| Published automatic runner | 0 | Baselines and all 9 required narrowings verified |
| Each of 9 mutant Go runs | 1 | Named behavioral assertion, not a compile error |
| Published self-check | 0 | 5/5 negatives trip |
| Independent real-runner stale/missing/zero/baseline-failure/survivor | 1 each | Correct matching GATE-FAIL refusal |

Owner registry covers all five owners; positives exercise all 36 owner/code/form cells, including both fixed owners in direct/wrapped/joined forms. Four fixed-owner narrowing mutants fail the published named positive test, including the exact joined UsageError survivor. Existing full normative foreign matrices (44 pairs, 168 form cases including extras) and 23 nil cases remain present. The current-source coverage is complete; some count expressions are literal arithmetic rather than measured execution counters, so they should not be described as independent runtime measurement. This observation is not an additional blocking finding for the pinned package.

Framing probe reaches `run -> Resolver.Resolve (ExecRunner) -> diagnostics.Emit -> Line`. The equality-based single-detail exemption fails its named test; all three separate hostile companions remain green. CodeOf has no non-test production caller in the pinned source (only its declaration); owner tests prove API classification, not integrated main dispatch.

Independently verified all 8 published blob OIDs from the extracted restored source, SPEC SHA-256, 18 SPEC codes and JSON foreign complements. Tree: fbe90d5e60593a3a069721b2ad9e53cd071d8c02. Successful runner restores diagnostics.go byte-equal. Toolchain: Darwin arm64, Go 1.25.5. Executed the focused two-package suites through the runner; no unrelated suites, hosted CI, installs, tags, real ax/model execution or original diagnostics workspace writes. Historical producer gap-probe evidence was read as producer evidence, not claimed as independently rerun.

## Empty repository delta and preservation

The assigned CR delta from 3ff66a9421ff6ddf675a49fc0c2868309f6e3de3 to ff61be4a8bd43fa4ffb179d31aa38e41891d4313 is empty. That is the right delivery shape: this leaf explicitly requests an artifact-only conformance package, preserving diagnostics and previous resources. No repository change is required for this leaf. Rejection concerns the published adoption artifact, not emptiness. No commits, branch switches, rebases, or source edits were performed. Review writes are confined to task-local .temp and board resource mutations. Original and prior resource bytes are preserved.

Evidence: TASK-260909-3d1589_review-evidence-rev1.txt includes exact manual block, reproducible negative probe script, resource SHA-256 hashes, provenance, all logs and exit codes. Scratch path: .temp/TASK-260909-3d1589/review1/. Acceptance/AC/green-gate checklist stays unchecked because R1 is unresolved. This is ordinary producer rework, not an external Stop-The-Line blocker.
