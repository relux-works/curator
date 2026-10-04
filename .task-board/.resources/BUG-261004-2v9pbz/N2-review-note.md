# Review note — BUG-261004-2v9pbz (N2) security fix (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Same-provider review (operator rule, tb-R164). Brief: `n2-budget-brief.md`; acceptance criteria and checklist are on the element. Source: docs/security-audit-2026-10-inline.md (issue #106); N5 came from the cocoaskills comparison.

The producer worked in produce mode: targeted tests only; the hosted gate on the CR is the arbiter. Verify independently, with real exit codes and GOFLAGS=-work:
1. RED FIRST is real. Rerun the new regressions against the BASE (pre-fix) tree: they fail for the stated reason. Then confirm they pass on the candidate.
2. Every acceptance criterion holds at the production entry point named in the brief, not just at a helper.
3. The mutants named in the brief are killed. Add one mutant of your own that targets the most likely regression.
4. No weakening elsewhere: the control rows still behave, and the rc.14 conformance tests for the touched package stay green.
5. The hosted CR validation gate result: cite it. If it has not run or is red, say so.
6. Scope: no CHANGELOG.md or LOGBOOK.md edits, and no unrelated changes.

accept_cr if all hold; otherwise changes requested with concrete findings.
