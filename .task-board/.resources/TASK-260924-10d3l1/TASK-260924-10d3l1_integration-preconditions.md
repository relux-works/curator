# TASK-260924-10d3l1 rev4 — landing preconditions (integration run RUN-260925-63a76a)

Status at check: integrating (left untouched; no status writes made).
Worktree: .temp/STORY-260924-1oyh2m/worktree, files unchanged by this run.

## 1. Candidate scope
- `git diff --name-only HEAD -- . :!.task-board` => exactly `internal/install/draftevidence_test.go`
- `git status --porcelain` => only `M  internal/install/draftevidence_test.go`
- No CHANGELOG change, no stray files.

## 2. Content vs accepted
- `git diff HEAD -- internal/install/draftevidence_test.go | git patch-id --stable` => `65e72d7524e3238fcbeb7d9565aedccc3bd51762` (not the rev1/rev2 `c7ce917d` id, as expected: trunk moved under the candidate per m28s6b).
- Trunk side preserved: `exact-admits`, `wrong-name`, `wrong-context` rows and marker-attestation assertion all intact in worktree.
- Candidate side present: new `wrong-repository-only` (source_identity-only) and `wrong-commit-only` rows, fail-closed at install.Project with the shared strict typed refusal, prior-state preservation over sourcelock/bindings/marker/SKILL.md/info.md, no Attestations on refusal, no endpoint/key leak (stub URL/pinned/ed25519/127.0.0.1). Nothing of trunk dropped.

## 3. Focused validation
- `go test ./internal/install -run ^TestDraftEvidenceExactMatch$ -count=1` => ok (exit 0, ~19.7s).

## 4. Integration boundary
- Did NOT run `task-board worktree integrate` and did NOT run `worktree checkpoint`; per the bound-runner assignment the runner performs the landing synchronously after this run. No files changed, no status/handoff writes made by this run apart from this evidence artifact.
