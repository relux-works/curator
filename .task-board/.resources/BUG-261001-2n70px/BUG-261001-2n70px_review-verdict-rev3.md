# Review verdict — BUG-261001-2n70px rev3 (CR-BUG-261001-2n70px-3): ACCEPTED

Reviewed delta f0119a8b..73c8b31e (10 paths) against nongit-brief / nongit-gate-note / nongit-review-note.

## Behaviour (real compiled `curator install`, throwaway HOME, non-git product folder with nested repo)
- non-git / non-git-fix rows: exit 0, skill bytes (SKILL.md, references/info.md) byte-equal, notice "gitignore hygiene check does not apply: project root is not a git repository"; no .git/.gitignore created, no undeclared surface, source tree digest unchanged.
- git regression rows unchanged: git-parent-ignored, git-ignored, git-missing-ignore (skipped, no bytes), git-fix-ignore, git-ignore-negation (skipped, no bytes).
- Fail closed: broken-git (corrupt `.git` file) exit 1; unexpected-exit-2 (even carrying the absence text) and unexpected-exit-128 (unrelated message) exit 1 with "git check-ignore failed".
- 128 discrimination (internal/gitignore/gitignore.go:~68-75,86-90): exit code must be exactly 128 AND stderr exactly Git's two stable repository-discovery diagnostics (LC_ALL=C forced, gitignore.go:~113); exit 1 = not ignored; everything else (incl. previously-"skipped" 128 cases: dubious ownership, corrupt gitdir, bare repo) now errors. Exit code and message must both agree; no "arbitrary error = not a repo" path. Dev-substitution second gate (install.go:~368) tolerates ErrNotRepository only; schema-1 lane covered by TestProjectNonGitWithDevSubstitution.

## Commands run by me (zsh; exit shown by `ok` line)
- go test ./cmd/curator -run InstallGitignore -count=1 -v → all 10 rows PASS, ok 24.4s
- go test ./internal/gitignore -count=1 → ok
- go test ./internal/install -run 'Gitignore|NonGit|Project' -count=1 → ok 101s (subset)
- Full `go test ./internal/gitignore ./internal/install -count=1` (run concurrently with gate-selftest + mutant): NOT completed locally — internal/install hit the default 10m go test timeout (FAIL 601.2s, load-induced; the same package's targeted subset passes in 101s). I did not rerun it with a longer -timeout; full-package result is accepted from the hosted gate (all lanes green on rev3 per the CR), the arbiter per the host-limits note. internal/gitignore ok.
- bash .github/ci/gate-selftest.sh → exit 0, "301 passed, 0 failed" (incl. the 4 new non-git ledger rows)

## Mutant (disposable rsync copy in /tmp, worktree untouched)
Restored old skip for ErrNotRepository ("skipped" + return) in install.go: non-git and non-git-fix FAIL (install exit 0, output "...not a git repository; skipped", bytes absent); broken-git still passes (exit 1). Positive row kills the mutant.

## CI policy
- No new skip class vocabulary: one new platform-control allow row in skip-classes.tsv, fully anchored regexp (^...$) matching exactly one reason string — cannot match other tests' skips. platform-cases.tsv: 11 rows, 9 run on all three OSes; only unexpected-exit-2/-128 skip on windows (POSIX `#!/bin/sh` PATH shim). Precedent: existing "test transport wrapper is POSIX-only" skips in cmd/curator. gate-selftest pins: Windows may skip exactly the two shim rows, cannot skip non-git or broken-git rows, Unix cannot skip the shim rows.
- Residual (non-blocking): the two shim rows could run on Windows with a Go-built fake git.exe; Windows still covers non-128 fail-closed via real corrupt `.git` row and the real not-a-repo path via non-git rows. Optional follow-up.

## Docs/hygiene
cli.md paragraph accurate (non-git folder, notice, no git init/.gitignore even with --fix-gitignore, lock at root, bindings under manager home source-bindings/); one CHANGELOG line, trunk 2772iz line preserved, no conflict markers; LOGBOOK untouched; no employer names.
