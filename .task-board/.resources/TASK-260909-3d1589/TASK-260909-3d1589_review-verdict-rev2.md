# Review verdict — TASK-260909-3d1589, CR revision 2

Verdict: **changes requested**. Route: **to-dev**. No acceptance or integration authorization.

## R1 — medium: published automatic block masks a runner refusal

Location: `TASK-260909-3d1589_rev2_adoption.md`, second bash block, consecutive runner and `--self-check` commands.
Repeat-of: TASK-260909-3d1589 review rev1 R1, published adoption entry masks a failed command with a later success. The manual entry itself is now repaired; the revised automatic entry retains the same failure class.

Extracted both bash blocks verbatim from the published Markdown and executed them with Bash in a fresh task-local shared Git clone. After the automatic block successfully populated auto-work, replayed that exact block. The main runner printed `GATE-FAIL: refuses existing non-empty destination .../auto-work` and returned 1. The block continued to `--self-check`, whose five probes passed, and the **whole block returned 0**. The independent review harness correctly failed with exit 1 because it expected refusal from the block.

The runner guard exists and is called; the newly published shell composition discards its failure. This is a real repeated-use route, with no runner, production source, or guard mutation. A successful self-check establishes runner behavior, not successful adoption into the requested destination.

Required bounded repair: short-circuit the automatic block on failure of its primary runner invocation (for example `... "$GITROOT/.../auto-work" || exit $?` or chain the two commands with `&&`). Replay the exact corrected automatic block on a fresh destination and then on the populated destination; require nonzero for the latter and verify prior bytes are preserved. Keep the repaired manual entry, overlays, vectors and runner unchanged. No broad suite replay or production changes are needed.

## Independently rerun results

| Entry/scenario | Exit | Result |
| --- | ---: | --- |
| Manual, fresh destination | 0 | Baselines and all nine named narrowing mutants behave correctly |
| Manual, stale destination | 1 | Correct refusal; SHA-256 map of every prior file including binary sentinel unchanged |
| Manual, absent conformance overlay | 1 | Correct missing-overlay refusal |
| Manual, all conformance TestGate names disabled | 1 | Correct zero-selection refusal |
| Manual, TestGateCoverageCounts injected failure | 1 | Named diagnostics FAIL and correct baseline refusal |
| Automatic block, fresh destination | 0 | Primary gate and five self-check negatives pass |
| Automatic block, stale destination | **0** | **Incorrect success: primary refusal masked by self-check** |
| Review harness | 1 | Assertion detects the incorrect automatic-block exit |

Positive manual and automatic runs each executed the focused diagnostics and curator-run package suites, 8 named diagnostics gate tests and 2 named framing tests, and all 9 mutants with exit 1 and required named assertions. Both fixed owners were attacked in joined and wrapped forms. All 5 runner self-check negatives tripped. Baseline source and SPEC provenance were verified by the published runner against tree fbe90d5e60593a3a069721b2ad9e53cd071d8c02. Successful runner restored diagnostics.go byte-equal. No unrelated suite, CI, install, tag, ax/model call, or original diagnostics workspace write.

The conformance overlay, framing overlay and vectors are byte-identical to rev1 (independently compared). Prior review's registry/source analysis remains applicable: 5 owners, 36 owner/code/form positives, 23 nil cases, 44 normative foreign pairs and 168 form cases with extras. This review reran those tests through the runner; it relies on rev1 independent evidence for the full normative-source transcription analysis rather than claiming to repeat that analysis. Count formulas are not independent execution counters. CodeOf classification remains API-level coverage; framing reaches run -> Resolver.Resolve (ExecRunner) -> diagnostics.Emit -> Line.

## Empty repository delta and preservation

The exact assigned CR diff from 3ff66a9421ff6ddf675a49fc0c2868309f6e3de3 to ff61be4a8bd43fa4ffb179d31aa38e41891d4313 is empty (git diff exit 0). No repository change is the correct delivery shape because this leaf explicitly requests an artifact-only gate package and preservation of diagnostics source. Rejection concerns a published artifact, not the empty delta. Original resources were read-only; scratch overlays were restored, and all previous replay destinations were retained under distinct names. No branch changes or commits were made.

Evidence: TASK-260909-3d1589_review-evidence-rev2.txt includes exact command blocks in invocations, complete outputs and exit codes, hashes, provenance, and the reproducible replay script. Local scratch: .temp/TASK-260909-3d1589-review2/. This is bounded producer rework, not an external blocker.
