# TASK-260823-1wvgw8 — reviewer verdict: CHANGES REQUESTED (-> to-dev)

Reviewer run RUN-260824-471418. Target: PR 35 delta, impl commit `726ccb3`, merge `b869a90`
(current `origin/main` head). Reviewed against candidate spec `6001dc3` §4.2.3 and
`conformance/v1/vectors/module-roots.json`.

## Summary

The driver work is correct and I found no defect in it. The task is nevertheless not
acceptable, because one line of CI metadata added by the same commit takes the whole
repository red, and the branch was merged with three lanes already failing. `main` is red
right now for exactly that reason. The last acceptance criterion — "merged to main green" —
is not met, and the board note claiming "every lane verified green pre-merge" is factually
wrong.

The fix is one line. Nothing in `internal/godriver`, `internal/moduleroots` or
`internal/install` needs to change.

## BLOCKING — `root-artifacts.tsv` defers `internal/godriver` against the committed pin

`.github/ci/root-artifacts.tsv:36` (new in this commit):

```
internal/godriver	vectors/module-roots.json	TestModuleRootVectorsDriveTheWholeBuild reads the vector directly
```

The committed `SPEC_PIN` (`00b1688a9b2457ca397a0bb550acf47cad8ee967`) publishes no
`vectors/module-roots.json` — verified by `git ls-tree` over the pin. So `suite-plan.sh`
now moves `internal/godriver` from *served* to *deferred*, i.e. the package runs with
`CURATOR_CONFORMANCE_ROOT` **unset**. Reproduced locally in a clean worktree at `b869a90`
against the materialised pin root:

| suite-plan vs pin `00b1688` | served | deferred | `internal/godriver` |
| --- | ---: | ---: | --- |
| `77aafa0` (last green main) | 35 | 8 | served |
| `b869a90` (this delta) | 34 | 9 | **deferred** |
| `b869a90`, the one row removed | 35 | 8 | served |

With the root unset, `TestCandidateGoV1SourceAwareContract` skips with
`CURATOR_CONFORMANCE_ROOT is not set` instead of `... publishes no build-drivers vector`
(both reproduced locally). `skip-classes.tsv` classifies the first as `root-unset`;
`platform-cases.tsv:168` tolerates a skip of that case on darwin/windows only under class
`root-content`. `platform-case-gate.sh` therefore fails by name.

Observed, not inferred:

* PR 35 head `726ccb3`, run 32676699282 — `Test (macos-latest)` FAILURE 00:34:58Z,
  `Race (macos-latest)` FAILURE 00:38:58Z, `Test (windows-latest)` FAILURE 00:54:38Z.
  Each job's only failure is
  `FAIL  ledger case skipped for the wrong reason on <goos>: internal/godriver :: TestCandidateGoV1SourceAwareContract`.
  `go test` itself is exit 0 on every lane.
* The PR was merged at 00:54:58Z — 20 seconds after the Windows lane went red.
* `main` run 32678133350 on `b869a90`: `Test (macos-latest)` and `Race (macos-latest)`
  already FAILED; Windows still in flight and deterministic to fail the same way.

The row also contradicts the table's own stated contract
(`.github/ci/root-artifacts.tsv:20-24`):

> Only packages whose conformance tests FAIL rather than skip on a missing artefact need a
> row: a package that already guards with `t.Skipf("%s publishes no ... vector", root)` is
> safe under any root and is deliberately absent from this table.

`TestModuleRootVectorsDriveTheWholeBuild` does exactly that guard
(`internal/godriver/moduleroots_test.go`, `os.IsNotExist(err)` -> `t.Skipf`), so the row is
not merely harmful, it is unnecessary. The neighbouring `internal/moduleroots` row IS
legitimate — `internal/moduleroots/conformance_test.go:26` reads the vector unguarded, and
that package carries no `platform-cases.tsv` row, so its deferral is harmless.

### Requested change

Drop the `internal/godriver` row from `.github/ci/root-artifacts.tsv`, push a fix-forward
onto main, and confirm all lanes green on the new main head. Verified locally that this
restores the exact `35/8` partition of the last green baseline and that the candidate lane
still fully serves every package (`CI_REQUIRE_FULL_ROOT=1` vs `6001dc3`: served=41
deferred=0, `suite-plan: ok`).

Also reconcile the board note: "every lane verified green pre-merge" did not hold.

## What I verified and accept

Local evidence, clean worktree at `b869a90`, `CURATOR_CONFORMANCE_ROOT` = candidate
`6001dc3` `conformance/v1`:

* `go build ./...` — 0.
* `go test ./internal/godriver/... ./internal/buildsource/... ./internal/moduleroots/... ./internal/install/...`
  — 0 (godriver 79.4s, buildsource 0.9s, moduleroots 1.2s, install 107.7s,
  install/atomicity 92.9s).
* `TestModuleRootVectorsDriveTheWholeBuild` — all 10 published vectors RUN and PASS, none
  skipped. The test asserts `go_list_started`/`go_build_started` off the stub launcher's
  call log, so a rejection arriving one fixed command late fails even with the right
  diagnostic. Good test.
* `CURATOR_REAL_GO_BUILD_TEST=1` — `TestRealGoV1ModuleRootsBuildIsBoundedAndNotLaunched`
  PASS (10.8s) and `TestRealGoV1VendoredBuildIsBoundedAndNotLaunched` PASS (8.4s).

Spec conformance, read line by line against §4.2.3:

* **Failure boundary.** `verifyModuleDeclaration` runs in `Build` before the worker exists
  and therefore before the fixed `go list`; `admitModuleRoots` runs inside
  `validatePackageGraph` after `go list` and before `go build`. Matches the vectors'
  `evaluation_order` and every `fails_before`.
* **Effective replace set.** Read only from `<R>/vendor/modules.txt`. `Module.Replace` in
  the stream is never the source of the set, and `Module.Replace.Dir`/`.GoMod` are never
  read at all. The two-token-left reconciliation is right: I checked it against the real
  `go mod vendor` output committed as `testdata/realmodules/tools/cli/vendor/modules.txt`,
  which carries both the selection annotation and the one-token directive.
* **Admitted form / bijection / containment.** Correct, including `path.Join`-based escape
  detection and the unconditional platform-path-mapping comparison, which is what makes the
  `windows-case-colliding` vector pass on a case-insensitive host.
* **Scan surface.** `firstParty` withholds both audited-vendor allowances (the below-`vendor`
  `SFiles` allowance and the `golang.org/x/sys` `cgo_import_dynamic` allowlist) from any
  result whose module carries a replacement — and the test proves both directions, that the
  allowance exists for an ordinary vendored module and is withheld for a replaced one.
  `scanDeclaredModules` classifies the declared copy conservatively from the tree, skipping
  only `testdata`, dot/underscore names and a nested `vendor` tree, which §4.2.3 explicitly
  says takes no part in resolution. Parsing the import block rather than byte-matching
  `import "C"` is the right call and is pinned by a test.
* **Identity.** `buildmeta.Input` untouched; no cache key, receipt, marker or artifact-path
  change, as §4.2.3 requires.
* **Closed surface.** `modules` is held to the same rule as `source_dir`, order-sensitive,
  with absent and empty normalising to one declaration. `internal/install` carries the
  declaration from the manifest to `StageRequest` and the test asserts it end to end.

Notes, non-blocking:

1. The driver's re-verification passes only the command's own build root as `buildRoots`,
   while §4.2.3 says a declared directory must not overlap *any* declared build root. The
   full rule is enforced at parse time (`internal/skillspec/parse.go:938` passes all build
   roots), and the driver check is documented as a fail-closed backstop, so this is
   correct as built — but `newVectorFixture` ignores the vectors' `build_roots` field and
   uses only `build_root`, so a future vector distinguishing the two would pass silently.
2. The deliberate tightening the producer flagged is spec-mandated ("A command with an
   absent or empty `modules` list therefore MUST have an empty effective replace set") and
   correctly implemented. It does change behaviour for a schema-6/7 skill carrying an
   *unused* `replace` directive, which previously built. Worth a CHANGELOG line.
3. `readVendorModules`' `info.Mode().IsRegular() && info.Mode()&fs.ModeSymlink == 0` is
   redundant — `IsRegular()` already excludes symlinks. Harmless.
4. The two pre-existing repairs are real and correctly scoped: `git ls-tree 77aafa0`
   confirms `testdata/realbuild/build/vendor/example.test/vendored/message/message.go` was
   never committed because `.gitignore`'s `*.test` matched the `example.test` module-path
   directory. `git check-ignore` now returns nothing for those paths and both opt-in
   real-toolchain integrations pass.
5. The producer's unowned finding about `registry.checkSnapshotsWithPolicy` not defaulting
   `clockSkew` is plausible and does need its own board item; it is out of this scope.
