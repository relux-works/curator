# TASK-260906-19gjyw: consume-the-landed-path-overlay-rule

## Description
Port the landed git-source-only overlay form rule into the Go reader so a path overlay is declarable from all three production surfaces, and make the candidate lane green against curator-spec main. Stage (c) was written against authority 550579d and passes there; the reconciliation landed as 87a0d00 after that authority was frozen, so this is consumption work rather than a defect in stage (c).

## Scope
curator repository, branch feat/consume-overlay-rule in worktree /Users/iv/Developer/ReluxWorks/.worktrees/curator-overlay-consume, base 7c74a49216f1da820ca562eb2643353ccea520b1 (stage (c) landed). Authority curator-spec main 87a0d006.

## Acceptance Criteria
A path overlay declared in machine configuration is reachable from all three production surfaces (the config reader, resolveOverlay, and profile compose add) and joins the closure with its weight, driven through run(). One exported source-kind discriminator matches the landed manager-config-v2 schema and never probes the filesystem. A path overlay carrying range, tag, branch, revision or directory stays profile_source_invalid. profile install operand classification uses the same helper with no source silently changing kind. The candidate lane against curator-spec main 87a0d006 with CI_REQUIRE_FULL_ROOT=1 is green on all three runners, closing the 30 currently failing internal/config subcases. The stale path-overlay bound and ledger rows 303-305 are retired and truthful. Every new refusal is driven through run() and proven by a narrowing mutant that kills a named test.
