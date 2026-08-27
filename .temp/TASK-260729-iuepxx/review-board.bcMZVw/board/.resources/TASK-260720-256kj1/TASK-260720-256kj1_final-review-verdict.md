# Final review verdict: accepted

Task: TASK-260720-256kj1 — Implement immutable build-source identity
Verdict date: 2026-07-21
Verdict: accepted → done

## Review findings

The implementation matches the task scope and acceptance criteria. The new internal/buildsource package implements the exact curator-build-source-v1 domain prefix and uint64 big-endian framing, unsigned UTF-8 path ordering, root .csk-install.json inclusion, portable and platform-collision validation, link and special-file rejection, and a retained-root token with before/after rechecks. Directory structure participates in frozen-state equivalence without changing the file-only protocol digest. internal/hashing has no diff, preserving legacy marker-excluding ContentSHA256 behavior.

The first-cycle atomic-repair defect is resolved. Snapshot repair now archives and validates the immutable repository source before cache reuse, serializes commit-keyed publication with OS locks, publishes validated immutable sibling generations, atomically selects the live regular-file generation, preserves the historical snapshot directory, and retains the immediately retired generation. Static review found no remaining stop-ship correctness or architecture defect in scope.

## Independent validation

- Candidate conformance plus focused packages: CURATOR_CONFORMANCE_ROOT=.../conformance/v1 go test ./internal/buildsource ./internal/snapshot ./internal/hashing -count=1 — pass.
- Concurrent present-tamper repair: go test ./internal/snapshot -run TestConcurrentGetRepairsPresentTamperedSnapshotAtomically -count=100 — pass, zero caller failures or missing historical live-path observations.
- Focused race detector: go test -race ./internal/buildsource ./internal/snapshot ./internal/hashing -count=1 — pass.
- Repository gate: make check — pass, including go vet, full go test, and gofmt validation.
- Coverage: buildsource 82.0 percent; snapshot 72.4 percent.
- Native curator build — pass.
- Windows amd64 buildsource and snapshot test compilation — pass; Linux amd64 snapshot test compilation — pass.
- git diff --check and gofmt diff — clean.
- internal/hashing diff and status — empty.

## Scope note

The conformance input remains the explicitly authorized candidate suite at SHA-256 70cd274150d629f16ca04f2073a8eae5a0f3fff4584f905d807775040af21eae; this verdict is implementation conformance evidence, not release or pin evidence. The worktree-local untracked task-board.config.json is orchestration state and is outside the implementation diff.
