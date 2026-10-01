# BUG-261001-2n70px integration run — pre-landing verification (rev3, base f0119a8b)

Role binding: developer (implementer). Board already `integrating` for BUG and STORY; status untouched per integration assignment. No files changed; no `worktree integrate` executed (runner performs the bound landing synchronously after this run).

## Delta matches rev3
- `git diff --name-only`: 10 paths (gate-selftest.sh, platform-cases.tsv, skip-classes.tsv, CHANGELOG.md, cmd/curator/install_gitignore_test.go [new], docs/cli.md, internal/gitignore/gitignore.go, internal/gitignore/gitignore_test.go, internal/install/gitignore_spawn_test.go, internal/install/install.go).
- CHANGELOG.md: trunk E4 line kept, one new Fixed line added above it; no conflict markers.
- `git diff --check`: exit 0. Conflict-marker grep over all 10 paths: no matches (exit 1). No untracked files. LOGBOOK.md: no diff (untouched).
- Fix shape confirmed: exit 128 + repository-discovery diagnostic -> ErrNotRepository notice-and-continue (LC_ALL=C); exit 1 -> not-ignored; all other git failures fail closed. install.go keeps skip policy for NotIgnored, fail for tool errors.
- CI: no new skip class; two unexpected-exit rows skip on Windows only under existing `platform-control` with exact message; gate-selftest gains a dedicated ledger section with negative rows.

## Verification (real exit codes, standalone processes, no pipes)
- `go test ./cmd/curator -run InstallGitignore -count=1` -> exit 0 (ok 19.059s). Entry-point matrix: non-git positive + notice, git regression rows, broken-git/shim refusal rows.
- `go test ./internal/gitignore ./internal/install -count=1` -> exit 1: internal/gitignore ok (1.658s); internal/install hit the go-test 10m panic while TestDraftDeclaredTagBumpIsNotAMovedTag was running (no assertion failure). Isolated rerun `go test ./internal/install -run TestDraftDeclaredTagBumpIsNotAMovedTag -count=1 -timeout 9m` -> exit 0 (ok 14.684s). Unrelated draft-tag test; full-package timeout was host-load contention (concurrent gate-selftest builds + another run on this shared host). The rev3 diff adds no blocking, retry, or wait, so it cannot introduce a hang.
- `go test ./internal/install -run TestProjectRefusesOnGitignoreSpawnFailure|TestProjectNonGitWithDevSubstitution -count=1 -v` -> exit 0, PASS (11.236s).
- `bash .github/ci/gate-selftest.sh` -> exit 0: 301 passed, 0 failed (includes the new non-git install ledger section, all ok).
- `go build ./cmd/curator ./internal/install ./internal/gitignore` -> exit 0 (BUILD_OK).

## Landing preconditions: confirmed
Worktree holds exactly the accepted rev3 delta, uncommitted, on task-board/story/STORY-261001-1rrh6z. Ready for the runner-bound landing.