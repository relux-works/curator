# TASK-260910-3kvq02 resume-2 results (developer, 2026-09-16)

## Preconditions verified
- `git log --oneline -3`: `c3d2067` (parser checkpoint TASK-260910-24cuys)
  on `18f0549` on `9213119` ("Set up Node on the rose-air lane…", post-PR72 main). Checkpoint present on current trunk.
- `git apply --check --reverse` of `TASK-260910-3kvq02_collections-candidate-rev0.patch`: exit 0 — the six owned paths were already present in the working tree byte-exact; no re-apply needed, zero conflicts, zero adaptations.
- `git status --short`: `M internal/closure/closure.go` + 5 untracked new files (`docs/draft-source-expansion.md`, `internal/closure/selections.go`, `internal/closure/selections_test.go`, `internal/manifest/expand.go`, `internal/manifest/expand_test.go`). Nothing committed; work remains uncommitted for the handoff snapshot.

## File hashes (sha256)
- 277c997c docs/draft-source-expansion.md
- 5c581d82 internal/closure/closure.go
- 94f65afc internal/closure/selections.go
- 07a23df0 internal/closure/selections_test.go
- c0cab236 internal/manifest/expand.go
- 215ab85a internal/manifest/expand_test.go

## Narrow verification (all run directly, real exit codes)
| Command | Exit |
|---|---|
| `go test -count=1 ./internal/manifest ./internal/closure ./internal/identifiers` | 0 (ok x3: 0.711s / 14.755s / 0.798s) |
| `go vet ./internal/manifest ./internal/closure ./internal/identifiers` | 0 |
| `golangci-lint run ./internal/manifest/... ./internal/closure/... ./internal/identifiers/...` | 0 (0 issues) |
| `gofmt -l` on the five owned Go files | 0 (no output) |
| `git diff --check` | 0 |

Note: an initial `gofmt -l` invocation wrongly included `docs/draft-source-expansion.md` (exit 2, markdown is not Go); rerun on the five Go files only is clean (exit 0). Full landing suite deliberately NOT run locally per wave note — it runs once via the remote gate on handoff.

## Narrowing mutant (case-collision guard)
- Mutation: `key := strings.ToLower(member.Decl.Name)` → `key := member.Decl.Name` in `internal/manifest/expand.go` (narrows the case-fold to exact match).
- `go test -count=1 -run 'TestExpandRefusals/case-collision' ./internal/manifest`: exit 1, FAIL — both `good` and `GOOD` admitted, `source_name_conflict` not raised. Mutant killed: the test exercises folding, not just exact duplicates.
- Restored from `/tmp/expand.go.orig`; sha256 back to `c0cab236…`; `git apply --check --reverse` passes again; `go test -run 'TestExpand' ./internal/manifest`: exit 0.

## Scope / bounds (unchanged from candidate rev0)
Implements only accepted draft-sources-v1 / repository-transport-v1 collection expansion: `manifest.Expand` (individual selectors + immediate-child `*`/literal expansion, exclusions, SKILL.md validation, traversal/output-boundary refusal, name/version collisions, deterministic ordering), `closure.BuildExpanded` + `AcquireSelection` boundary, legacy `Build` refusal of unexpanded selectors. No snapshot capturer, lock writer, transport resolver, publication path, or schema-2 install enablement.
