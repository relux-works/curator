# TASK-260720-2jfnz6 review verdict — RUN-260730-bb92a2

## Verdict

CHANGES REQUESTED. Route to to-dev.

The R3 identity-stable publication rework closes the previously reported post-rename/pre-seal race, and all functional, race, type, and full-suite gates pass. The candidate does not satisfy the explicit task Definition of Done item Lint clean: Ruff fails on a file introduced and owned by this task. This is ordinary implementation rework, not a stop-the-line blocker.

## Exact candidate

- Product worktree: /Users/iv/Developer/Wildberries/cocoaskills/.temp/TASK-260720-2jfnz6/worktree
- Branch: task/TASK-260720-2jfnz6-protected-posix-build-cache
- Recorded base and merge base: 495ad021847529ce5a544dba415ca2fe19949539
- HEAD: 0d6ad16fce35c1bd8854511e13766cd236908e3b
- Current product delta is exactly two unstaged files: src/csk/builds/cache_posix.py and tests/test_build_cache_posix.py. The interface file src/csk/builds/cache.py was added by this task in HEAD and does not exist at the recorded base, so it remains part of the task-wide review scope.
- SHA-256 cache.py: d5936a40bc6628f37a7b5e485ff5651d11cd38ec939b27cb81e2cf4b9c6f0821
- SHA-256 cache_posix.py: be370f6f4b63d355d7082b31d0f753ddcd8ad89449588130ae4edc2140d1aa36
- SHA-256 test_build_cache_posix.py: c7724b619a8074d36f44d17aacea1124f0627d5130bcc8532e378dffe3f1d13c

## Finding R4-1 — task-wide lint is not clean

Command: uvx ruff check src/csk/builds/cache.py src/csk/builds/cache_posix.py tests/test_build_cache_posix.py

Result: exit 1. Ruff 0.16.0 reports I001 at src/csk/builds/cache.py:10 because the import block is separated from _SHA256_IDENTITY by one extra blank line. Ruff diff proposes deleting that single blank line. The candidate files and index remained unchanged.

The producer R3 evidence ran Ruff only on cache_posix.py and test_build_cache_posix.py and called cache.py out of scope because it is unchanged from HEAD. That is valid rework-delta hygiene but not task-scope hygiene: cache.py is one of the three files explicitly owned by TASK-260720-2jfnz6 and was created by its HEAD commit relative to base. Therefore the checked Lint clean DoD item is not currently evidenced. No project-specific Ruff configuration or CI lint job exists, but Ruff is the lint gate selected and asserted by the producer; the all-task-file invocation must pass before acceptance.

## Required rework

1. Remove the extra blank line identified by Ruff in src/csk/builds/cache.py, or otherwise make the exact all-task-file Ruff command pass without suppressing the finding.
2. Rerun Ruff over all three task-owned files, POSIX-focused pytest, strict python -m mypy, and the relevant candidate validation gates on the revised bytes.
3. Preserve the R3 cache_posix.py and test behavior; no functional/security change is requested.

## Passing evidence retained

- Fresh reviewer POSIX-focused pytest: 41 passed, 1 native-Linux-only skip, exit 0.
- Fresh reviewer strict mypy: Success, no issues in 63 source files, exit 0.
- Fresh reviewer full suite with the pinned accepted conformance corpus: 890 passed, 7 skipped in 85.36 seconds, exit 0.
- Fresh reviewer race stress: the ordinary identical/different winner tests plus both deterministic paused-after-rename tests ran 20 times; all 80 test executions passed.
- git diff --check and cache.py unchanged-from-HEAD checks pass.
- Producer native Ubuntu evidence is tied to the exact current cache_posix.py and test hashes: CPython 3.11, 3.12, 3.13, and 3.14 each passed all 42 focused tests with no skips.
- Static review found no remaining functional or security acceptance-criterion gap: protected-boundary-first parsing, rooted no-follow access, exact input/receipt/artifact validation, read-only inspection, protected fresh publication, atomic identical/different winner handling, immutable live entries, and locked quarantine are implemented coherently.

The reviewer changed no source, test, index, commit, branch, remote, or PR state and supplied no commit_ack. This run is not goal-bound. The operator nudge was observed at the final checkpoint and is satisfied by this explicit bounded verdict.