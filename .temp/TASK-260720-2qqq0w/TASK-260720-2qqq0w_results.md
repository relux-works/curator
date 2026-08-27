# TASK-260720-2qqq0w — Document Curator compiled-build authoring

## Candidate provenance

- Task-owned worktree: `.temp/TASK-260720-2qqq0w/worktree`
- Base commit: `17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8` (origin/main, "Pin landed rc.3 protocol")
- Working tree copied byte-for-byte from the accepted integrated worktree
  `.temp/TASK-260729-2kaopg/worktree` via `rsync -a --delete --exclude=.git`,
  verified with `diff -r --exclude=.git` (identical; only the two
  `skill-go-testing-tools` symlink loops were not traversed).
- Baseline before this task: 37 modified files, +4358/-914.
- No accepted worktree was mutated. Nothing was staged, committed, published, or
  pinned.

## Deliverables

| Path | Change |
|---|---|
| `docs/compiled-builds.md` | New. Complete schema 6 / `go-v1` authoring and operations reference. |
| `README.md` | Schema range 1→6, compiled-commands bullet, corrected toolchain-selection precedence, compiled shim invocation, new "Tools and verification" section, build-driver conformance caveat. |
| `cmd/curator/docs_test.go` | New. Seven tests that exercise the shipped documentation. |

`LOGBOOK.md` (repository root, outside the candidate worktree) carries the
2026-07-29 1247 entry recording the CI-pin finding below.

## Scope coverage

| Required item | Where |
|---|---|
| Complete mixed script + build manifest example | `docs/compiled-builds.md` § A complete mixed package |
| `build_roots` context exclusion | § `build_roots` → Build roots never reach agent context |
| Vendor-only native `go-v1` prerequisites | § Go source prerequisites |
| Trusted toolchain precedence incl. `CURATOR_GO` | § The trusted toolchain; `README.md` |
| Cache and marker currentness | `README.md` § Compiled-command status…; guide § Logical identity and local paths |
| Dry-run outcomes | `README.md`; guide § Operating compiled commands |
| Install/upgrade repair | `README.md`; guide § Operating compiled commands |
| Locked GC + 24h grace | `README.md` § Maintenance…; guide § Operating compiled commands |
| Unix/Windows shim invocation | guide § Running a compiled command; `README.md` |
| Output untrusted, never run at install | guide § Trust boundary; `README.md` |
| Unsupported: hooks, argv/env, cgo, workspaces, downloads, external linking, root modules, generic drivers | guide § What the build is not allowed to do |
| Protocol links, no copied vectors | guide § Protocol status |
| Schema 1–5 guidance preserved | guide intro; `README.md` bullets |
| Portable logical identity vs Curator local paths | guide § Logical identity and local paths |
| Exact verification commands + artifact locations | guide § Verification; `README.md` § Tools and verification |

## Documentation tests

`cmd/curator/docs_test.go` parses the shipped documents rather than a copy:

1. `TestDocumentedJSONBlocksParse` — every ```json fence in `README.md` and the
   guide unmarshals.
2. `TestDocumentedMixedManifestLoads` — the documented example is materialized
   from the document's own fenced blocks and loaded through `skillspec.Load`;
   asserts schema 6, `agent-skill.json` as source file, `build_roots`,
   `runtime_roots`, and all three command types.
3. `TestDocumentedCurrentnessVocabulary` — every `currentnessCodes()` and
   `inputCauses()` value is documented, and no unknown build state is.
4. `TestDocumentedDryRunOutcomes` — all six `install.BuildOutcome` values appear
   in both documents.
5. `TestDocumentedToolchainSelection` — documented mechanisms equal
   `godriver.SelectionCuratorGo` / `SelectionGOROOT`, every
   `godriver.TestedFamilies()` entry is documented, and both documents state
   that `PATH` is never searched.
6. `TestDocumentedBoundaryCodesAreComplete` — the guide's failure-table Codes
   column is compared, in both directions, against every code the driver sources
   can emit. It caught `go_test_input_forbidden` as missing during development.
7. `TestDocumentedLinksResolve` — every relative link and in-document anchor
   resolves; mutation-checked by temporarily appending a bad link and a bad
   anchor, both of which failed the test, then reverting (`diff -q` identical).

The documented example was additionally compiled with the real toolchain under
the exact fixed argument vector:

```
go list  -mod=vendor -deps -json -buildvcs=false -compiler=gc -pgo=off ./cmd/indexer   -> exit 0
go build -mod=vendor -trimpath -buildvcs=false -buildmode=exe -compiler=gc -pgo=off \
         -ldflags='-linkmode=internal -libgcc=none' -o <tmp>/indexer ./cmd/indexer     -> exit 0, 2371298 bytes
```

`go mod vendor` on that stdlib-only module reports `no dependencies to vendor`
and writes nothing, and `-mod=vendor` still succeeds — which is why the guide
states `vendor/` is required only once something outside the standard library is
imported.

## Gate evidence

Every command run standalone in the candidate worktree; real exit codes.

| Gate | Exit | Note |
|---|---|---|
| `gofmt -l cmd internal` | 0 | empty output |
| `go build ./...` | 0 | |
| `go vet ./...` | 0 | |
| `~/go/bin/golangci-lint run` | 0 | `0 issues.` (`golangci-lint` is not on `PATH`; it is installed at `~/go/bin`) |
| `go test ./internal/skillspec/ -count=1` | 0 | |
| `go test ./cmd/curator/ -run TestDocumented -count=1` | 0 | 7/7 pass |
| `go test ./... -count=1` | FULLSUITE_EXIT | see below |

### Expected-red gate

`CURATOR_CONFORMANCE_ROOT=<rc.3 checkout>/conformance/v1 go test ./internal/buildsource/ ./internal/buildcache/ ./internal/buildmeta/ -count=1 -run 'Conformance|Vector|Candidate'`
exits **1**, and that is the correct result, not a passing gate:
`internal/buildsource` and `internal/buildcache` require
`vectors/build-drivers.json`, the published suite does not carry it, and a
missing vector file is a hard failure rather than a skip. `internal/buildmeta`
passes. With `CURATOR_CONFORMANCE_ROOT` unset every package passes.

## Finding: the committed CI pin cannot satisfy the build-driver conformance tests

- `.github/workflows/ci.yml:28` pins `relux-works/curator-spec` at
  `00b1688a9b2457ca397a0bb550acf47cad8ee967` (`1.0.0-rc.3`) and exports
  `CURATOR_CONFORMANCE_ROOT` for `go test ./...`.
- That published ref carries no `conformance/v1/vectors/build-drivers.json`, no
  `agent-skill-v6` schema, and no `install-marker-v2` schema; published tags stop
  at `v1.0.0-rc.3`. The schema-6 protocol text exists only as uncommitted work on
  a local `curator-spec` `agent/protocol-v6-core` branch.
- Consequence: the build-driver conformance tests fail against the committed pin.
  Reconciling the pin with a published suite that carries those vectors is
  `TASK-260720-1pvfj5` (enforce-cross-platform-ci-gates), which is blocked by this
  task. It is not a defect in the candidate implementation and is not something
  documentation can assert away.
- Handling here: the docs link only to paths that resolve at the published ref,
  state the compiled-build contract as accepted-but-unpublished, name the current
  pin explicitly, and document the expected failure. No pin was moved and no
  rc.4/rc.5 release is claimed.

## Corrections made to pre-existing text

The accepted baseline README stated that Curator selects the trusted Go
installation "only through `CURATOR_GO` … or `GOROOT`". `selectToolchain`
(`internal/godriver/session.go:465`) has a third branch: when neither variable is
set it falls back to `build.Default.GOROOT`, the GOROOT the manager binary was
built against. Both documents now state all three mechanisms in order, and keep
the accurate claim that `PATH` is never searched and no toolchain is downloaded.
