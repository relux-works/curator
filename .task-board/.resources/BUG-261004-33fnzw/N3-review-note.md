# Review note — BUG-261004-33fnzw (N3) security fix (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Same-provider review (operator rule, tb-R164). Brief: `n3-restore-brief.md`; acceptance criteria and checklist are on the element. Source: docs/security-audit-2026-10-inline.md (issue #106); N5 came from the cocoaskills comparison.

The producer worked in produce mode: targeted tests only; the hosted gate on the CR is the arbiter. Verify independently, with real exit codes and GOFLAGS=-work:
1. RED FIRST is real. Rerun the new regressions against the BASE (pre-fix) tree: they fail for the stated reason. Then confirm they pass on the candidate.
2. Every acceptance criterion holds at the production entry point named in the brief, not just at a helper.
3. The mutants named in the brief are killed. Add one mutant of your own that targets the most likely regression.
4. No weakening elsewhere: the control rows still behave, and the rc.14 conformance tests for the touched package stay green.
5. The hosted CR validation gate result: cite it. If it has not run or is red, say so.
6. Scope: no CHANGELOG.md or LOGBOOK.md edits, and no unrelated changes.

accept_cr if all hold; otherwise changes requested with concrete findings.

## HOSTED-EVIDENCE MODE (tb-keeper desk #52 decision, 2026-10-04; binding, overrides the "rerun" steps above)
The Mac mini has host-wide exec stalls. Do NOT run go build, go test, go vet or lint locally.
- Verify by reading code and the diff.
- Use the HOSTED evidence: the CR validation gate run on GitHub. Get its run id from the CR validation resource or `gh run list`. Download its artifacts with `gh run download <id> -R relux-works/curator` run from the worktree; go-test JSON or log artifacts show which tests ran and passed.
- Red-first: show from the diff that each new regression asserts the defect, i.e. it would fail on the base logic. Explain the assertion against the base code; do not execute it.
- Mutants: argue kill or survive by reading. Do not execute.
- If the hosted gate did not run the new tests, or is red, that is a finding (changes requested), not something to fix locally.
Keep board commands to the minimum needed: one outcome resource and the verdict.
