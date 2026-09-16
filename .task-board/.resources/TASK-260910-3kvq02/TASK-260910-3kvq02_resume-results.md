# TASK-260910-3kvq02 refreshed candidate results

Restored the preserved rev0 patch without edits or conflicts onto parser checkpoint 824494690e2c3171d79b2cd9f220126993af439a, directly above b0e905d. Both git apply --check and git apply exited 0. Six owned paths remain uncommitted for runtime snapshotting.

Read previous results and task logbook; their implementation description and explicit integration bounds remain applicable. Read current landed source/transport contracts. No new product behavior was introduced during this resume.

## Direct verification
All commands ran as standalone zsh processes, without pipelines, on Go 1.26.0 darwin/amd64:

- go test -count=1 ./internal/manifest ./internal/closure ./internal/identifiers — exit 0 (1.211s, 16.358s, 2.099s).
- go vet ./internal/manifest ./internal/closure ./internal/identifiers — exit 0.
- golangci-lint run ./internal/manifest/... ./internal/closure/... ./internal/identifiers/... — exit 0; 0 issues.
- go build ./cmd/curator — exit 0. Binary moved into ignored .temp.
- git diff --check — exit 0.
- test -z "$(gofmt -l internal/closure/closure.go internal/closure/selections.go internal/closure/selections_test.go internal/manifest/expand.go internal/manifest/expand_test.go)" — exit 0.

## Negative gate attack
Narrowed the destination collision guard from any lowercase-key collision to only exact-case equal names. Ran go test -count=1 ./internal/manifest -run '^TestExpandRefusals/case-collision$': exit 1, expected failure. Assertion observed both good and GOOD accepted instead of source_name_conflict. Restored exact bytes from a copy; cmp exited 0. Identical test command then exited 0.
Measured on this resume: 1/1 narrowing mutants killed. Earlier 8/9 mutation evidence remains historical evidence of the preserved candidate, not a rerun on this refreshed base; no comprehensive mutation coverage claimed.

## Scope and bounds
Tests drive manifest.Expand, closure.BuildExpanded and legacy closure.Build, including positive, refusal and legacy package regression cases. Alias grammar remains supplied by the accepted parser. Dependency units and provider-first ordering remain skill-level. The trusted acquisition callback is tested with immutable temporary fixtures; it is not production snapshot capture. No CLI schema-2 installation, lock replay, transport authentication, root-input admission, publication, all-73-case or cross-platform claim. External conformance vectors were not supplied. Full landing suite is reserved for the handoff runtime exactly once; independent review remains required.

## Resume logbook
Base-authority refresh resolved by exact patch replay; no forced-fit adaptation or unrelated changes. Campaign forbids LOGBOOK.md edits, so this outcome records the resume event. Narrow verification is fresh; prior full evidence is not represented as a new run.

## Exact source bytes
Base plus these six files identifies the candidate; the handoff runtime records its tree/revision.

- internal/closure/closure.go: `5c581d82ba4864ac02dddea4661534faf7febe124654cc1638feb28018451c98`
- internal/closure/selections.go: `94f65afc8b4d78516a6dd329badb9d6216f2872e42fcc050a53a7a157ab9228d`
- internal/closure/selections_test.go: `07a23df0aa6926f8214d0778e1749d9ea57597c4f89cd41e26427ba51b635338`
- internal/manifest/expand.go: `c0cab2363f9044186d70a819ce02dabd1207e1af470f1e0b0a04bdd9fbaf1b3d`
- internal/manifest/expand_test.go: `215ab85a57a52166beea1d5f543161e55accc8d08296c0a26550ee6e15bfe548`
- docs/draft-source-expansion.md: `277c997c4d2eb839f4f3e1cc2d59aa5b27c1185d4e8c9e4b7cbc61d52363d02b`
