# TASK-261008-1j34ro results — revision 4 republish (no code changes)

The workspace was converged onto d7001974. `git diff HEAD` touches CHANGELOG.md (the N8 entry only), cmd/curator/main.go, internal/audit/audit.go and internal/hashing/hashing.go, plus the untracked cmd/curator/audit_allow_test.go and internal/audit/pin_digest_test.go. No conflict markers were found and no files were edited.

## Compile-only checks (local, R223)
- go vet ./... -> exit 0
- go build ./... -> exit 0
- gofmt -l cmd internal -> no output

## Hosted green (merged candidate)
- Scratch branch scratch/TASK-261008-1j34ro-green, commit 5a6b389c, which is d7001974 plus the exact worktree delta. CI was started with workflow_dispatch because pushes to scratch branches do not trigger CI.
- https://github.com/relux-works/curator/actions/runs/37916848969 -> conclusion success (`gh run watch --exit-status` exit 0). Test, Race, Gate self-test and the Go drivers all succeeded on ubuntu, macos and windows. Lint, the naming gate and the interop conformance gate also succeeded. rose-air and the candidate suite were skipped, which is the normal behaviour for a dispatch run.
- The scratch branch has been deleted.

## Red and ordering-mutant evidence
These come from revision 2 and still apply: the production fix and both test files are byte-identical. main.go and audit.go differ only by the clean merges with N9 and 4mzun5.
