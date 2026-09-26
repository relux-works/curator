# TASK-260922-cww1ov: envprofile-0017-production-entry-tests

## Description
F-C3: Production-entry tests on temporary stores covering both 0017 hazards, a stale link surviving shared to isolated and the unlink of a regular file at a link path, plus the dangling native target case where a recorded link points to a native file that does not exist, the migration and every repair refusal, each with a narrowing mutant proving the bound; runs on the three hosted lanes with declared skips only through the ledger vocabulary.

## Scope
internal/envprofile and the env CLI surface; docs/troubleshooting; CHANGELOG

## Acceptance Criteria
go test ./internal/envprofile/... passes with a mutant-killed row per refusal, conflict admitted means the test fails; the dangling-target case is a driven row; ledger rows registered
