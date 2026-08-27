# TASK-260824-1n98b3: advance-consumer-pins

## Description
After rc.9 publishes: separate consumer PRs per the landing order step 9 — curator repo: bump the single SPEC_PIN in .github/workflows/ci.yml env block to the exact immutable rc.9 release commit, run default CI on all three OSes green, merge-commit per repo convention; cocoaskills: advance its released-suite pin/variables the same fail-closed way, rebase-merge, verify post-merge main green (its decisive Windows shard jobs run only post-merge). Executor: claude only.

## Scope
(define task scope)

## Acceptance Criteria
Both consumer pins on rc.9 with green default CI in both repos including post-merge main runs.
