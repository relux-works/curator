# Integration preflight — marker-reader-cross-field-validation

Bound integration run RUN-261002-23622f; developer / implementer; accepted CR-BUG-260923-2afgyq-1 revision 1. No source files changed, no status mutation, no generic handoff, checkpoint, or integration invoked. The runner owns synchronous landing after producer exit.

Board query confirms status integrating. Worktree status confirms accepted story_final, present repository delta, producer developer/implementer, active workspace lease held by this run, and candidate tree 4eded38359c54177b468b13d7c4934b4c52e4f7b. HEAD and checkpoint are 2cb29dac8a4c82c5a07d7ca2d107aa6e6e2c93e7. Exactly the expected four paths are modified/untracked.

Fresh checks performed here:
- syspolicyd launchctl state filter with pipefail: exit 0; running, successive crashes 355.
- Standalone git diff --quiet against accepted candidate for the three tracked changed files: exit 0.
- Untracked test file git hash-object and accepted-tree blob both cf9651f826024ad14e3ae505082868609ff86494; commands exit 0.
- Standalone git diff --check: exit 0.
- git ls-remote origin refs/heads/main: exit 0; fresh advertised main f40b77c19c01746bda8b9a610358d860f2ad20c5. This differs from the accepted base. The worktree status authority observation was older; it is not fresh authority proof. Overlap, ancestry and combined-tree validation are NOT established here. The runner must apply its normal fresh-authority/reparent/refusal rules; this evidence does not authorize bypassing them.
- Board resource reads, status and directive reads: exit 0; no run directives. Two exploratory schema lookups for unsupported change_request/cr operations exited 1; no mutation resulted.

Accepted evidence inspected, not rerun: BUG-260923-2afgyq_change-request_rev1-validation.log records sh scripts/remote-gate.sh exit 0, required command shards 1/1 green, hosted run 36958642909. BUG-260923-2afgyq_review-verdict-rev1.md accepts the exact candidate and records passing scoped marker, coverage, install and CLI validation plus 5/5 killed mutants. Full local install-package green remains UNKNOWN because broad attempts timed out; no fresh test/build or mutant run is claimed by this integration preflight. Reviewer accounting is five distinct cases and 20 physical ledger deletions across published suites/versions, with no unrelated deletions.

Preflight confirms accepted revision binding and unchanged candidate files. Landing has not occurred in this producer turn; fresh remote advance requires runner assessment. Board remains integrating.