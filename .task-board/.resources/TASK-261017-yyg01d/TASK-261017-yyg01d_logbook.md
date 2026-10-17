# Preservation triage logbook

TASK-261017-yyg01d — preservation-triage. Source: a field report.

- Current-main research at implementation fcd785ba93c5d00e2e80615f8add2baa93ccd73d and spec 1059b94b4135089d842bf017070b8ab47d79a536. No tracked repository delta.
- A: silent deletion of three unmanaged canonical entries and dangling native symlink confirmed. Prioritize first; proposed refusal-before-publication fix leaf.
- C: selected canonical destination with no marker is replaced destructively; native adapter collision refuses and native takeover preserves exact manual bytes in backup. Unrelated .agents/.instructions/ content survives the tested install. Separate overwrite leaf from stale-cleanup leaf.
- B: context-only package without manifest checks/resolves but cannot install. Core optionality conflicts with source marker schema requiring version 1..8; format decision is a prerequisite to the implementation leaf.
- D/E reproduce but follow specified activation and per-entry partial-switch semantics. Document first-install exception and recovery rather than change runtime behavior.
- Setup mistakes and the first snapshot-verifier failure are retained in sandbox logs and excluded from green product evidence. Corrected snapshots plus six existing targeted Go tests passed; full CI/conformance/mutants were not run and are not claimed.
- Required outcome: research and proposed leaves, no product implementation. No surface-table supplied in brief; coverage is bounded to A-E fixtures, not a catalog-complete gate.
- See TASK-261017-yyg01d_triage.md, sandbox-logs.txt, validation.log and commands.txt for evidence, exact exits, controls, spec mapping, risks and priority. This task-scoped logbook travels as an outcome; the control-root LOGBOOK.md was not edited.
