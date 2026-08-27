# TASK-260823-1wvgw8 — go-v1 builds with declared first-party module roots

**Branch:** `task/TASK-260823-1wvgw8-godriver-module-roots`, branched from `origin/main@77aafa0`.
**Candidate suite:** `relux-works/curator-spec@6001dc3` (`candidate/schema-8-rc.9`) —
`protocol/core.md` §4.2.3, `profiles/manager.md` §2.3 module-root table,
`conformance/v1/vectors/module-roots.json`.
**Predecessor:** TASK-260823-1vleh5 landed `internal/moduleroots` and the schema-8
parse surface. This task wires the driver.

## What the driver does now

### Before the fixed `go list` — declaration and containment

`godriver.Build` runs `moduleroots.ValidateDeclaration` against the frozen
snapshot before the worker exists (`internal/godriver/moduleroots.go`,
`verifyModuleDeclaration`). The parser already ran the same check over the same
snapshot; the driver runs it again because it is the component that starts Go,
so a caller that skipped or mis-plumbed the parse-time check cannot hand the
compiler a directory nothing validated. `BuildRequest` therefore carries
`RuntimeRoots` as well as `Modules` — the runtime roots exist for exactly one
purpose, the containment comparison of §4.2.3, and select nothing about the
build.

### After `go list`, before `go build` — form and bijection

`admitModuleRoots` reads `<build root>/vendor/modules.txt`, computes the
effective replace set with `moduleroots.EffectiveReplaceSet`, and checks
`moduleroots.ValidateBijection`. It returns the set of module paths whose
replacement the bijection admitted. `Module.Replace` in the `go list` stream is
never the source of the set, and `Module.Replace.Dir`/`Module.Replace.GoMod`
are never read as evidence that a path exists.

**This runs for every command, not only a declaring one.** §4.2.3 requires a
command with an absent or empty `modules` list to have an *empty* effective
replace set. That is a real tightening and it is deliberate: `go mod vendor`
writes a one-token-left annotation for an unused replace directive too
(reproduced on go1.25.5, see "Empirical checks"), so before this change a
schema-6 or schema-7 skill carrying an unused `replace` in its `go.mod` built
successfully — the directive reached no package result, so `validateModule`
never saw it. It is now rejected as `build_module_root_directive_undeclared`.
A *used* replacement was already rejected, only with `vendor_metadata_inconsistent`.

### `validateModule`

The unconditional `item.Module.Replace != nil` rejection at the old
`graph.go:306` is gone. A non-standard result whose module carries a
replacement is admitted exactly when `admitModuleRoots` accounted for that
module path, and it must still resolve from below `<build root>/vendor` — the
one rule §4.2.3 relaxes is versioning, because a replaced module is not
versioned there. A replacement outside the admitted set means `go list` and
`vendor/modules.txt` disagree, which is `vendor_metadata_inconsistent`. A main
module carrying a replacement is rejected outright.

### Scan surface

Two halves, matching profile step 4 ("the scan surface above run with the
declared directories included and the vendored exceptions withheld from
replaced modules"):

1. **Withheld allowances.** A result whose module carries a replacement no
   longer gets the vendored-`SFiles` allowance or the
   `golang.org/x/sys` `//go:cgo_import_dynamic` allowlist.
2. **Declared directories.** `scanDeclaredModules` walks each declared module
   directory. The vendor copy is covered by the `go list` stream; the declared
   copy is not in that stream at all — it takes no part in `-mod=vendor`
   resolution — and the fixed argument vectors admit exactly one `go list`, so
   the toolchain cannot classify it. It is classified conservatively from the
   tree: every native-input extension (`.c .cc .cpp .cxx .m .h .hh .hpp .hxx
   .f .F .for .f90 .swig .swigcxx .s .S .sx .syso`) is rejected with the same
   diagnostic the corresponding `go list` field would raise, and every non-test
   Go file is rejected if it imports `"C"` or carries the exact
   `//go:cgo_import_dynamic` bytes.

Two interpretation calls are made explicit here because a reviewer should
check them rather than discover them:

- **Conservative superset.** "Active inputs" of the declared directory cannot
  be computed without a second `go list`, which the protocol forbids. Every
  Go file is treated as active, which over-approximates. That is the
  fail-closed direction, §4.2.3 withholds every audited allowance here on
  purpose, and the profile says first-party code "can simply not use those
  constructs".
- **Skipped subtrees.** `testdata`, names beginning with `.` or `_`, and a
  nested `vendor` tree are skipped, because Go compiles from none of them
  under any configuration. The nested `vendor` case is named by §4.2.3 itself:
  it "takes no part in resolution" at the build root and the manager must not
  read it as a dependency source, so it holds no active input of this build.
- **cgo detection parses the import block** (`go/parser`, `ImportsOnly`) rather
  than byte-matching, so the bytes `import "C"` inside a string or comment
  cannot reject a build that has no cgo in it. A file that does not parse is a
  Go input in no build, so it cannot introduce cgo; it is still byte-scanned
  for the directive.

### Package-controlled surface

`modules` joins `type`, `driver`, `source_dir` as the fourth admitted key of
the build-command object, and is held to the same rule as `source_dir`: the
declared list must be exactly the validated one, so a list the manager never
checked cannot reach the directories the driver trusts. Absent and empty are
one declaration. Both the `[]string` the manager rebuilds and a decoded
manifest's `[]any` are accepted spellings.

### Plumbing

`skillspec.Command.Modules` and `Spec.RuntimeRoots` now reach the driver
through `install.PlannedBuild` → `install.StageRequest` →
`godriver.BuildRequest`, and `install.commandObject` emits `modules` only when
the schema-8 command declared a non-empty list, so a schema-6 or schema-7
command still presents exactly three fields.

### Not changed, deliberately

- **Cache identity.** `buildmeta.Input` is untouched. §4.2.3: "No algorithm,
  domain separator, framing, receipt identity, install marker, or
  artifact-relative path changes." §8.1 already binds the whole validated
  snapshot, so an edit below a declared directory already moves the key.
- **Context exclusion.** `whitelist.ContextExcludedRoots` still excludes build
  roots and runtime roots only. A declared module directory is not a build
  root and §4.2.3 changes no §8.1 rule, so its files stay agent-facing.
- **Vendor-copy reconciliation.** The two copies are never compared. §4.2.3
  forbids it and says a divergence is not an error.

## Suite consumption

| Family | Consumer | Result |
| --- | --- | --- |
| `vectors/module-roots.json` (10 cases) | `godriver.TestModuleRootVectorsDriveTheWholeBuild` | all 10 pass |
| `vectors/module-roots.json` (10 cases) | `moduleroots.TestModuleRootVectors` (predecessor) | all 10 pass |

The driver-level consumer is not a second copy of the unit-level one.
`internal/moduleroots` asserts each vector against the half of §4.2.3 that owns
it. The driver test runs each vector through `Build` itself and reads
`go_list_started` / `go_build_started` off the stub launcher's own call log, so
a rejection that arrives one fixed command too late fails there even when it
carries the right diagnostic. `.github/ci/root-artifacts.tsv` declares the
dependency for `internal/godriver`.

## Behavioral tests added

`internal/godriver/moduleroots_test.go`:

- a multi-module vendored snapshot builds, with exactly one `go list` and one
  `go build`;
- every declaration failure class rejects with **zero** source-aware commands;
- every bijection failure class rejects after **exactly one** `go list`;
- a replacement with no declaration is refused (the schema-6 rule);
- declared roots without `vendor/modules.txt` are `vendor_metadata_inconsistent`;
- a `go list` replacement outside the admitted set is refused;
- each audited-vendor allowance is proved to *apply* for an ordinary vendored
  module and to be *withheld* from a replaced one — so the test proves the
  withholding, not merely that something failed;
- eleven declared-directory scan classes reject, and seven files Go would never
  compile do not;
- the closed `modules` surface, in both spellings and seven mismatch classes.

`internal/godriver/testdata/realmodules/` is a real `go mod vendor` multi-module
fixture (two first-party modules replaced by one build root).
`TestRealGoV1ModuleRootsBuildIsBoundedAndNotLaunched` compiles it with the real
toolchain — the only way to prove the path against Go's own vendor consistency
check rather than a canned stream — and then proves the same snapshot is refused
without the declaration.

`internal/install/moduleroots_test.go` pins that a schema-8 manifest's modules
and the skill's runtime roots reach the staging boundary, and that a schema-6
command declares neither.

## Verification

Every command run standalone, unpiped; the real exit code is reported.

| Gate | Exit |
| --- | ---: |
| `gofmt -l cmd internal` | 0, no output |
| `go build ./...` | 0 |
| `go vet ./...` | 0 |
| `golangci-lint run` (v2.12.2, CI's pin) | 0 — "0 issues." |
| `go test ./internal/... -count=1 -timeout 30m` | 0 (42 packages ok) |
| `go test ./cmd/... -count=1 -timeout 30m` | 0 (`cmd/curator` 284.356s) |
| `CURATOR_REAL_GO_BUILD_TEST=1 go test ./internal/godriver -run TestRealGoV1` | 0 (both integrations) |
| `bash .github/ci/gate-selftest.sh` | 0 — 81 passed, 0 failed |
| `bash .github/ci/suite-plan.sh <candidate root> …` | 0 — served=43 deferred=0 excluded=0 |
| `CURATOR_CONFORMANCE_ROOT=<6001dc3> CI_REQUIRE_FULL_ROOT=1 bash .github/ci/test-gate.sh` | see below |

`GOOS=windows` / `GOOS=linux` native execution was not run from this session;
the three-OS matrix runs on the PR.

## Empirical checks, not taken from prose

Run against a real `go mod vendor` with go1.25.5 darwin/arm64:

- a directory replacement of a required module writes **both** the selection
  annotation (`# m v0.0.0 => ../../pkg/m`) and the one-token-left directive
  annotation (`# m => ../../pkg/m`), which is what the reconciliation in
  `EffectiveReplaceSet` depends on;
- an **unused** replace directive writes the one-token-left annotation too,
  which is what makes "none can hide from validation by going unused" real and
  what makes the empty-list rule a genuine tightening;
- for a replaced module under `-mod=vendor`, `go list` reports the package
  `Dir` inside `<build root>/vendor`, `Module.Version` as `v0.0.0`,
  `Module.Dir`/`Module.GoMod` **absent**, and `Module.Replace.Dir`/`GoMod`
  pointing **outside** the build root — which is exactly why validating those
  two fields against the vendor root would reject every legitimate declared
  build, and why the code does not look at them.

## Repository defects repaired on the way

Both are pre-existing on `origin/main@77aafa0` and both block this task's
acceptance criterion, so they are fixed here rather than filed.

1. **`.gitignore` silently swallowed the build-driver vendor fixtures.**
   Line 5's `*.test` (Go test binaries) also matches the reserved `example.test`
   module path used by `internal/godriver/testdata/**/vendor/example.test/…`.
   `!**/testdata/**/example.test/` re-includes those directories.
2. **`testdata/realbuild` was missing its vendored package.** Because of (1),
   `build/vendor/example.test/vendored/message/` was never committed, so
   `TestRealGoV1VendoredBuildIsBoundedAndNotLaunched` failed with
   `vendor_dependency_missing` for anyone who set `CURATOR_REAL_GO_BUILD_TEST=1`.
   Verified red on `origin/main@77aafa0` in a clean worktree and green after the
   file was restored. The gate is opt-in and never runs in CI, which is why it
   went unnoticed.

## Findings not fixed here

**`internal/install` registry snapshot tests are time-sensitive and flaked once
under the full gate.** `TestStrictRegistryPolicyFailsUnknown` failed one
`test-gate.sh` run with `registry test-reg snapshot timestamp is too far in the
future`, then passed on a rerun and in 48 repeat runs on `origin/main`.
Mechanism, from the code: `registry.checkSnapshotsWithPolicy`
(`internal/registry/snapshot.go:75`) defaults `maxAge` when it is zero but does
**not** default `clockSkew`, and the install test env
(`internal/install/commit_test.go:1067`) builds a bare `config.Config`, so the
effective tolerance is 0s rather than the production
`DefaultSnapshotClockSkewSeconds = 300`. The fixture stamps `created_at` inside
the HTTP handler, after `install` captured its own `time.Now()`, and truncates
to whole seconds — so any second boundary crossed between the two calls trips
`CreatedAt.After(now)`. Under a saturated gate run the window widens. This is
registry/config surface, untouched by this task, and the fix (default
`clockSkew` too, or give the test env real audit defaults) changes behaviour
outside it. Needs its own board item.

**Carried forward from the predecessor's review, still unowned:** the marker
v3/v4 cross-field validation gap, the two REQUIRED audit warning classes
(`script-command-declared-only`, `script-command-unfiltered-declared-network`),
`cmd/curator/builds.go:365` treating every marker above v2 as unsupported, and
`README.md:36` still saying schemas 1 through 5 while `SupportedSchemaVersions`
spans 1–8. None is touched here.
