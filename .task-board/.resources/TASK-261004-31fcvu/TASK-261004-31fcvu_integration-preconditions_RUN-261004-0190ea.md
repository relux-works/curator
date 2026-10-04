# TASK-261004-31fcvu — rc3-release-notes: integration preconditions

Bound run RUN-261004-0190ea, developer / implementer. Board remains integrating. No repository edits, status mutations, generic handoff, checkpoint, or integration were performed by this run. The runner owns synchronous landing after this producer exits.

## Fresh checks (actual exit codes)

- task-board spawn status: 0; bound run running, developer / implementer.
- task-board q task status: 0; integrating.
- task-board q change-request activity: 0; revision 2 transitioned to accepted.
- task-board resource get revision-2 review verdict: 0; accepted, no outstanding findings.
- task-board worktree status STORY-261002-2327ef --json: 0; CR-TASK-261004-31fcvu-2 accepted, kind story_final, producer developer / implementer, changed paths only CHANGELOG.md. Active workspace and lease belong to this run; checkpoint reachable. Expected dirty worktree contains the accepted uncommitted candidate.
- Standalone Python byte/scope assertions: 0. Working CHANGELOG.md equals accepted candidate tree 6c30a1ff1e35ceda69bdaf86207f75e763cab835. Both candidate-vs-base and worktree-vs-base changed paths equal CHANGELOG.md only. No staged or untracked paths. rc.2 heading through EOF byte-identical to base; fresh empty Unreleased precedes rc.3. LOGBOOK.md untouched.
- git diff --check: 0.
- git ls-remote origin refs/heads/main: 0; observed 934952a45953587a1d4184b692b3fb4ee401e732.
- git diff --name-only between accepted base 876127f7c714e01092f43c8950dc879421c52461 and observed main: 0; three board-only paths.
- git diff --exit-code between those OIDs, excluding .task-board: 0; no product-file drift.
- task-board spawn directives: 0; none.
- Board-wide worktree status: 0 after waiting without retry; output truncated, so the scoped query above supplied the actual target evidence.
- CLI discovery attempts task-board cr --help and task-board change-request --help each exited 1 (unsupported command names); no mutation occurred. Used supported activity and worktree status reads instead.

CHANGELOG.md SHA-256: db6c4b97ea2ab17723b51e5e50d8a5f7052243766c57b2623588887c79ea3e76.
Worktree HEAD and accepted base: 876127f7c714e01092f43c8950dc879421c52461.

## Evidence bounds and runner handover

Accepted existing revision-2 review evidence includes 27/27 historical-entry byte proofs, exact two-fix reconstruction, expected-red rev1 validator exit 1, rev2 validator exit 0, three regression tests, and history accounting. These document validators and Go/build/platform suites were not rerun in this integration-only run; no code or documentation was changed.

Remote main advanced only in board files at the observation above. This is not a fresh-authority or landing attestation: the runner must perform its own authority/fetch, source-identity, lane and publication gates. Worktree status also reports unrelated board activity debt and holder_run_recorded=false / holder_run_record_known=false despite naming this run as lease holder; these observations are not bypassed or repaired here.

Local candidate and acceptance preconditions verified. Runner may now attempt the bound revision-2 story_final integration; successful landing is not claimed.
