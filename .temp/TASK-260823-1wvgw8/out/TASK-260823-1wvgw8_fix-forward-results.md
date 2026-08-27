# TASK-260823-1wvgw8 — fix-forward after review (RUN-260824-8660aa)

Branch `fix/TASK-260823-1wvgw8-root-artifacts-row`, commit `c8ac575` on top of `b869a90`.
PR https://github.com/relux-works/curator/pull/36.

## 1. Reconciliation of the earlier board claim

The note attached when PR #35 landed said "every lane verified green pre-merge". That is
**factually wrong** and is corrected here.

| Run | Head | Lane | Result |
| --- | --- | --- | --- |
| 32676699282 (PR #35) | `726ccb3` | Test (macos-latest) | FAILURE |
| 32676699282 (PR #35) | `726ccb3` | Race (macos-latest) | FAILURE |
| 32676699282 (PR #35) | `726ccb3` | Test (windows-latest) | FAILURE |
| 32678133350 (main) | `b869a90` | Test (macos), Race (macos), Test (windows) | FAILURE |

`go test` itself was exit 0 on every one of those lanes. The failure is
`platform-case-gate.sh` alone, by name. The acceptance criterion "merged to main green" did
not hold at `b869a90`; PR #36 is what makes it hold.

## 2. Blocking fix — the `root-artifacts.tsv` row

PR #35 added:

```
internal/godriver	vectors/module-roots.json	TestModuleRootVectorsDriveTheWholeBuild reads the vector directly
```

The committed `SPEC_PIN` (`00b1688a9b2457ca397a0bb550acf47cad8ee967`) publishes no
`vectors/module-roots.json`, so `suite-plan.sh` moved `internal/godriver` from *served* to
*deferred* — the package ran with `CURATOR_CONFORMANCE_ROOT` **unset**.

The chain, each link verified locally, not inferred:

1. root unset → `TestCandidateGoV1SourceAwareContract` skips with
   `CURATOR_CONFORMANCE_ROOT is not set` instead of `... publishes no build-drivers vector`.
2. `skip-classes.tsv:44` classifies the first as `root-unset`; `:45` classifies the second
   as `root-content`.
3. `platform-cases.tsv:168` tolerates a skip of that case on darwin/windows only under
   `root-content`.
4. `platform-case-gate.sh` fails by name.

Reproduced against a freshly materialised pin root
(`git archive 00b1688 | tar -x`, out of `relux-works/curator-spec`):

| `suite-plan.sh` vs pin `00b1688` | served | deferred | `internal/godriver` | exit |
| --- | ---: | ---: | --- | ---: |
| `b869a90` table (via `CI_ROOT_ARTIFACTS`) | 34 | 9 | deferred | 0 |
| this branch | 35 | 8 | **served** | 0 |

Both skip reasons reproduced by running the case directly:

```
root SET   -> build_conformance_test.go:466: <pin>/conformance/v1 publishes no build-drivers vector
root UNSET -> build_conformance_test.go:461: CURATOR_CONFORMANCE_ROOT is not set
```

And the guard that makes the row unnecessary in the first place:

```
CURATOR_CONFORMANCE_ROOT=<pin> -run TestModuleRootVectorsDriveTheWholeBuild
  -> moduleroots_test.go:664: <pin>/conformance/v1 publishes no module-roots vector
  -> SKIP, package ok
```

`.github/ci/root-artifacts.tsv:20-24` says in as many words that a package guarding with
`t.Skipf("%s publishes no ... vector", root)` is deliberately absent from the table. The
neighbouring `internal/moduleroots` row stays — `conformance_test.go:26` reads the vector
unguarded and that package carries no `platform-cases.tsv` row, so its deferral is harmless.

Candidate lane unaffected: `CI_REQUIRE_FULL_ROOT=1 suite-plan.sh` vs candidate `6001dc3`
gives served=43, deferred=0, exit 0.

## 3. Reviewer's non-blocking notes 1–3, closed in the same commit

**Note 1 — the containment backstop verified a narrower rule than the one it mirrors.**
§4.2.3 says a declared module directory may not overlap *any* declared build root; the
driver re-check passed only the build root of the command being compiled.
`BuildRequest` gains `BuildRoots`; `internal/install` carries the manifest's whole declared
set through `PlannedBuild` → `StageRequest` → `verifyModuleDeclaration`, which **unions** it
with the command's own root rather than substituting, so an unplumbed caller is checked
exactly as strictly as before and never less. `newVectorFixture` now reads the vector's
`build_roots` instead of reconstructing a single-root set, and fails the vector outright if
`build_root` is absent from it. `TestModuleRootsRejectAContainmentCollisionWithASiblingBuildRoot`
builds and refuses the identical declaration across the two shapes from one fixture, so it
cannot pass by rejecting everything. `TestDeclaredModuleRootsReachTheBuilder` asserts the
set end to end through install.

**Note 2 — CHANGELOG.** Records the schema-8 addition and, under *Changed*, the behaviour
change for a schema-6/7 skill carrying an **unused** directory `replace` directive: it used
to build and is now `build_module_root_directive_undeclared`.

**Note 3 — redundant symlink term.** `readVendorModules`' `IsRegular() && Mode()&ModeSymlink == 0`
reduced to `IsRegular()`: `os.Lstat` does not follow the final component and `IsRegular` is
false for every non-regular mode bit. Two tests pin what is left — a directory standing at
`vendor/modules.txt` is `vendor_metadata_inconsistent` after exactly one `go list` and no
`go build`, and `TestALinkStandingInForVendorMetadataNeverReachesTheDriver` asserts the
frozen build source refuses a link there a whole layer earlier, so the reasoning that made
the term removable stays checked instead of assumed.

Note 4 (`.gitignore` repair) needed nothing. Note 5 (registry `clockSkew`) is out of scope
and still needs its own board item.

## 4. Local evidence

Each command run as its own process, unpiped, with its real exit code. Package tests ran
with `CURATOR_CONFORMANCE_ROOT` = candidate `6001dc3` `conformance/v1`.

| Command | Result | Exit |
| --- | --- | ---: |
| `gofmt -l cmd internal` | empty | 0 |
| `go build ./...` | — | 0 |
| `go vet ./...` | — | 0 |
| `golangci-lint run` (v2.12.2) | 0 issues | 0 |
| `go test ./internal/godriver/...` | ok 55.246s | 0 |
| `go test ./internal/moduleroots/... ./internal/buildsource/... ./internal/skillspec/...` | ok 0.381s / 0.828s / 1.456s | 0 |
| `go test ./internal/install/...` | ok 151.821s + 152.199s | 0 |
| `CURATOR_REAL_GO_BUILD_TEST=1 go test -run TestRealGoV1 ./internal/godriver` | both integrations PASS | 0 |
| `.github/ci/gate-selftest.sh` | 81 passed, 0 failed | 0 |
| `suite-plan.sh <pin>` | served=35 deferred=8 excluded=0 | 0 |
| `CI_REQUIRE_FULL_ROOT=1 suite-plan.sh <candidate 6001dc3>` | served=43 deferred=0 | 0 |

**Not run locally:** the full `.github/ci/test-gate.sh`. It is exactly what the red CI lanes
execute, and local attempts time out (this run lost a 10-minute foreground budget to one
combined `go test` invocation before chunking). PR #36's CI is the authoritative check, and
the merge waits on it rather than on a local substitute.

## 5. Out of scope, recorded

`origin/main`'s `LOGBOOK.md` is a four-line file added whole by `9a5f7f6` (#27); the
~3000-line logbook the project actually writes to exists only in the local `main` lineage,
which is 1 ahead / 25 behind `origin/main` and has never been pushed. The two files share no
history. Any logbook entry written from a task worktree therefore either lands in the
four-line file or never leaves the local checkout. Not caused by this task; recorded as
logbook entry `0627` and needs a lineage-ownership decision, not a drive-by fix.
