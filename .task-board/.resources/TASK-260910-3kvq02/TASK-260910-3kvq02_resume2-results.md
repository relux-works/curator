# TASK-260910-3kvq02 — resume-2 results (collections replay after PR72)

Run: RUN-260916-8899e0 (holds the STORY-260910-197y84 lease). Shell: bash. Go 1.26.0 darwin/amd64.

## Base verification (resume-2 gate 1)

`git log --oneline -3` in the Story worktree:

```
c3d2067 TASK-260910-24cuys: TASK-260910-24cuys: parse-opt-in-skillfile-v2-sources
18f0549 Record STORY-260916-jfyr3f board state
9213119 Set up Node on the rose-air lane and fix its comment
```

Parser checkpoint on top of current main (post-PR72). `task-board worktree status
STORY-260910-197y84`: tree clean before apply, branch
task-board/story/STORY-260910-197y84, rev 1 of TASK-260910-24cuys checkpointed.

## Patch replay (resume-2 gate 2)

Resource `.temp/resources/TASK-260910-3kvq02/TASK-260910-3kvq02_collections-candidate-rev0.patch`
(41713 bytes, materialized via `task-board resource get`):

- `git apply --check`: clean, exit 0
- `git apply`: clean, exit 0 — **zero conflicts, no adaptation**

Resulting tree (exactly the six owned paths, no collateral):

```
M internal/closure/closure.go
?? docs/draft-source-expansion.md
?? internal/closure/selections.go
?? internal/closure/selections_test.go
?? internal/manifest/expand.go
?? internal/manifest/expand_test.go
```

sha256 after apply (and after mutant restore — identical):

```
277c997c docs/draft-source-expansion.md
5c581d82 internal/closure/closure.go
94f65afc internal/closure/selections.go
07a23df0 internal/closure/selections_test.go
c0cab236 internal/manifest/expand.go
215ab85a internal/manifest/expand_test.go
```

## Narrow verification (resume-2 gate 3; real exit codes)

| Command | Exit | Result |
|---|---|---|
| `go test -count=1 ./internal/manifest ./internal/closure ./internal/identifiers` | 0 | ok manifest 1.033s, ok closure 25.183s, ok identifiers 1.510s |
| `go vet ./internal/manifest ./internal/closure ./internal/identifiers` | 0 | clean |
| `golangci-lint run ./internal/manifest/... ./internal/closure/... ./internal/identifiers/...` | 0 | 0 issues |
| `gofmt -l` on the five Go files | 0 | empty (clean) |
| `git diff --check` | 0 | clean |
| `go build ./...` | 0 | whole module compiles |

Full landing suite (`sh scripts/remote-gate.sh`) intentionally NOT run manually —
it runs exactly once at handoff on hosted CI.

## Narrowing mutant: case-collision guard (resume-2 gate 3)

Mutant: `key := strings.ToLower(member.Decl.Name)` → `key := member.Decl.Name`
in `internal/manifest/expand.go:71` (narrows the guard to exact-case matches).

- `go test -count=1 -run 'TestExpandRefusals/case-collision|TestExpandRefusals/version-collision' ./internal/manifest`
  → exit 1, `TestExpandRefusals/case-collision` FAILS (`GOOD` vs `good` no longer
  collides), `version-collision` (exact duplicate) still passes. Mutant killed:
  the negative test covers the case-folding class, not just exact duplicates.
- Restored via inverse sed; sha256 `c0cab236…` matches pre-mutant hash exactly.
- `go test -count=1 -run 'TestExpandRefusals' ./internal/manifest` after restore
  → exit 0.

## Bounds

- Evidence covers the three touched packages plus a whole-module build. Conformance
  suites, race detector, and other platforms are covered by the remote gate at handoff.
- Coverage claim: `manifest.Expand`, `closure.BuildExpanded`, and the legacy `Build`
  refusal path are driven through the committed tests listed above; acquisition
  fixtures are immutable temp dirs and are not evidence of snapshot capture, lock
  replay, transport auth, or installation (per docs/draft-source-expansion.md).
- No LOGBOOK.md edit (worktree rule forbids it); no findings beyond this evidence.
