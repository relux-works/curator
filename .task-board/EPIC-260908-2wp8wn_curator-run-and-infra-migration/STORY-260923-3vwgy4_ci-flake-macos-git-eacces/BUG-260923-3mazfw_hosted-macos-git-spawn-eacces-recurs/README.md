# BUG-260923-3mazfw: hosted-macos-git-spawn-eacces-recurs

## Description
The hosted macos-latest runner still intermittently fails to exec git: fork/exec /opt/homebrew/bin/git: permission denied. BUG-260920-3vfwch diagnosed and bounded it and BUG-260921-1fpaij made the failure a named error instead of a silent not-ignored, so it now fails candidates loudly on packages they do not touch. Latest: gate run 35855550263 (TASK-260907-187z6x republish), internal/install TestBuildRootExcludedBeforeLocaleRenderingWithoutCompilerExecution: dry-run error git check-ignore failed for .claude/skills/: fork/exec /opt/homebrew/bin/git: permission denied. Each occurrence costs a republish cycle.

## Scope
Test/runner determinism for git spawns on hosted macOS (retry on EACCES at the spawn seam in tests or a bounded production retry if justified); CI workflow if the runner git can be pinned or pre-warmed. No behaviour change on real permission errors.

## Acceptance Criteria
1) root cause stated with evidence (e.g. Homebrew updating git under the runner, or quarantine/xattr on first exec); 2) a bounded, named mitigation that retries only EACCES on the git spawn and still fails closed on a persistent permission error, with a test for both; 3) the hosted macOS lane passes the affected packages repeatedly; 4) CHANGELOG
