# Review brief: manager-profile environments sections

## Subject
- Worktree `~/Developer/ReluxWorks/.worktrees/curator-spec-env-manager`, branch `draft/environments-manager-profile`, head `6697c1e`, base exactly `c3b29b1` (= main). Delta: `profiles/manager.md` (+276, new §12 + §7 pointer), `cli/curator.md` (+37).
- Producer notes resource on TASK-260901-2pho68 (read all resources on the task) — includes a filed normative gap: install-time ref selection for `profile install` unspecified in environments.md §9.1. Judge that gap: real or covered elsewhere; if real, it stays a filed backlog item (do NOT fix protocol prose here).

## Dimensions
1. Consistency with `protocol/environments.md` at c3b29b1: every §12 claim cites or matches the normative doc; NO byte rules restated (spot-check the riskiest: header, chapters, opencode.json, collision rule — must be referenced, not duplicated).
2. Coverage vs task AC: registry incl. secondary targets + shadowing warns; modes/marker/ledger/drift; lifecycle incl. scoped switching + rev-1 onboarding; credential passthrough per platform; env resolve verify-and-repair; always-strict audit + context-secret-material; status/GC live roots.
3. House style vs existing manager.md sections; diagnostics referenced correctly (codes exist in environments.md tables).
4. cli/curator.md: informative tone kept, command rows sane, umbrella note accurate, example runs mentally.
5. Cross-references verified both ways; links validation.
6. Git facts: signature, base, delta files exactly the two named.

## Constraints and verdict
Read-only. `review-findings-env-manager-1.md` on TASK-260901-2pho68; blocking/major -> development; else ACCEPT explicit, leave to-review. Do not mark done. Ignore tools/__pycache__ and .temp/venv.
