# TASK-261018-70enqt: qualify-codex-0-162-1

## Description
Re-verify the codex_cli registry facts on Codex 0.162.1 in temporary homes, record the pilot behaviours, bump VerifiedRelease only if every fact holds, opt-in replay.

## Scope
.research/codex162, internal/envregistry verified release and test, CHANGELOG line. No spec text, no product behaviour change.

## Acceptance Criteria
Every codex_cli registry fact re-verified on 0.162.1 with raw evidence in temporary homes; trust prompt and .git sandbox recorded; VerifiedRelease bumped only if all facts hold; opt-in replay skipped by default; real HOME and root README untouched.
