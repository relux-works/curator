# BUG-260921-1fpaij: gitignore-missing-conflates-spawn-failure-with-not-ignored

## Description
internal/gitignore Missing() treats ANY cmd.Run() error of git check-ignore as the entry being not ignored (gitignore.go:23-27), so a git spawn failure (the hosted macOS EACCES flake, or git absent/broken) is reported as a policy outcome: install skipped with generated paths are not ignored by git; missing entries: .claude/skills/. Seen on PR #83 run 35550731645 Test (macos-latest): internal/install TestAdapterLedgerCommitsAfterTheMirrorsItClaims skipped with exactly that message while .gitignore was correct. check-ignore exit 1 = not ignored; any other failure (spawn error, exit 128, non-ExitError) must be returned as an error so the caller refuses with a spawn/tool diagnostic instead of a misleading policy message. Same class as BUG-260921-30ycv0 (product spawn failure masked).

## Scope
internal/gitignore Missing/Ensure error classification; callers keep their fail-closed behaviour

## Acceptance Criteria
check-ignore exit status 1 still means not ignored (existing rows); a spawn failure or exit 128 returns an error carrying the tool diagnostic (row with an injected non-executable git and a row with a non-repository root); install refuses with that error, never the not-ignored message; mutant restoring the conflation fails the rows; no retry
