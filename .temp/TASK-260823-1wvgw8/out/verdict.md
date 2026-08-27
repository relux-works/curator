# TASK-260823-1wvgw8 — reviewer cycle 2 (RUN-260824-9f95e5)

**Verdict: ACCEPTED.**

Scope of this cycle, per the coordinator's scope note: the PR 36 delta only
(`b869a90..c8ac575`, merged as `b00836c`). PR 35's substance was already
accepted on the merits by cycle 1 (`RUN-260824-471418`); the single blocking
finding was the `root-artifacts.tsv` row that took `main` red. This cycle
verifies that fix, the three non-blocking notes taken with it, and the AC clause
that failed last time — "merged to main green".

## 1. The blocking finding is fixed, and fixed in the shape the ledger prescribes

`.github/ci/root-artifacts.tsv` lost exactly one line:

    -internal/godriver	vectors/module-roots.json	TestModuleRootVectorsDriveTheWholeBuild reads the vector directly

Reproduced end to end in a clean worktree at `b00836c` against the pin root
materialised out of `relux-works/curator-spec` at the committed
`SPEC_PIN 00b1688a9b2457ca397a0bb550acf47cad8ee967`
(`git archive 00b1688 | tar -x`, root = `<pin>/conformance/v1`; confirmed that
root publishes **no** `vectors/module-roots.json`):

| Check | Result |
| --- | --- |
| `suite-plan.sh <pin> <ev>` | `served=35 deferred=8 excluded=0`, exit 0, `internal/godriver` **served** |
| deferred set | buildcache, buildsource, install, marker, moduleroots, scopes, skillspec, whitelist — the 8 that read unguarded, godriver absent |
| `CI_REQUIRE_FULL_ROOT=1 suite-plan.sh <candidate 6001dc3> <ev>` | `served=43 deferred=0 excluded=0`, exit 0 |

That restores the 35/8 partition of the last green main (`77aafa0`), which is
what cycle 1 asked for.

The removal is not merely benign, it is what `root-artifacts.tsv:20-24` states:
a package that already guards with `t.Skipf("%s publishes no ... vector", root)`
is *deliberately absent* from the table. `TestModuleRootVectorsDriveTheWholeBuild`
(`moduleroots_test.go:660-664`) has exactly that guard.

## 2. Skip-class correctness — the actual failure chain, re-observed

The cycle-1 chain was: row → godriver deferred → root UNSET →
`TestCandidateGoV1SourceAwareContract` skips with `CURATOR_CONFORMANCE_ROOT is
not set` (class `root-unset`) → `platform-cases.tsv:168` tolerates that case's
skip on darwin/windows only under `root-content` → `platform-case-gate.sh` fails
by name, with `go test` itself exit 0 on every lane.

Re-observed from the real `go test -json` stream at `b00836c` with the pin root
exported (`go test -json ./internal/godriver`, exit 0, **362 pass / 0 fail /
20 skip**), reading the verbatim reasons:

| Case | Reason printed | Class | Ledger policy |
| --- | --- | --- | --- |
| `TestCandidateGoV1SourceAwareContract` | `<root> publishes no build-drivers vector` | `root-content` | allow — and this is the class `platform-cases.tsv:168` demands |
| `TestModuleRootVectorsDriveTheWholeBuild` | `<root> publishes no module-roots vector` | `root-content` (`publishes no ` regex) | allow, under any root |
| `TestCandidateRC4ToolchainArtifacts` | `<root> publishes no expected/build-driver artifacts` | `root-content` | allow |
| `TestRealGoV1{ModuleRoots,Vendored}BuildIsBoundedAndNotLaunched` | `set CURATOR_REAL_GO_BUILD_TEST=1 …` | `opt-in` | allow, pre-existing class |

No `root-unset` reason anywhere in the godriver stream — the exact string that
went red is gone. No new skip class introduced.

The other direction is proven too: against the candidate root
(`6001dc33281b94a4ec7442ab15278550dd0f51d9`, which does publish the vector),
`TestModuleRootVectorsDriveTheWholeBuild` **runs all 10 published cases and all
10 PASS**, none skipped:

    valid-declared-module-roots, replacement-target-escapes-snapshot,
    module-to-module-redirect, undeclared-directory-replacement,
    declared-module-without-replacement, nested-declared-module-roots,
    module-root-contained-by-build-root, module-root-contained-by-runtime-root,
    versioned-left-directory-replacement,
    windows-case-colliding-declared-module-roots

So the requested shape — "godriver stays served against the pin; the vector test
skips cleanly when the root lacks the file and runs when a candidate root
provides it" — holds in both directions.

## 3. The three non-blocking notes

**(1) `build_roots` plumbing.** `BuildRequest` gains `BuildRoots`; `install`
carries the whole declared set from `skillspec.Spec.BuildRoots` through
`PlannedBuild` (`plan.go:463,526`, both `toolchainInventory` and `planOne`) and
`StageRequest` (`stage.go:110`, `builddeps.go:215`) to
`verifyModuleDeclaration`, which **unions** it with the command's own build root
rather than substituting. Checked adversarially:

- the union can only get stricter — an unplumbed caller (e.g.
  `external.go:293`, which supplies no set) is checked exactly as before;
- `moduleroots.validateContainment` uses `buildRoots` for pure string-overlap
  comparison only and never stats them, so feeding it extra roots that may not
  exist in the frozen snapshot cannot raise a spurious
  `build_module_root_declaration_invalid`;
- no manifest that passes parse can newly fail here, because `parse.go:166`
  already runs the identical rule over the identical set.

`TestModuleRootsRejectAContainmentCollisionWithASiblingBuildRoot` asserts both
directions from one fixture — the *identical* declaration builds when `pkg` is
not a declared build root and is refused with
`build_module_root_containment_invalid` before `go list` when it is — so it
cannot pass by rejecting everything. `newVectorFixture` now reads the vector's
own `build_roots` field and fatals if `build_root` is not inside it, instead of
reconstructing a single-root set. PASS locally.

**(2) CHANGELOG.** Present under `## Unreleased`, both halves: an `Added` entry
for schema-8 module roots (§4.2.3) and a `Changed` entry naming
`build_module_root_directive_undeclared` and spelling out that a schema-6/7
skill carrying an **unused** directory `replace` built before and now fails,
with the remedy. That is the behaviour change cycle 1 asked to be documented.

**(3) Redundant symlink term.** `readVendorModules` is down to
`err == nil && info.Mode().IsRegular()`, with a comment stating why (`Lstat`
does not follow the final component; `IsRegular` is false for every non-regular
mode bit). Backed by two new cases rather than by assertion:
`TestVendorMetadataMustBeARegularFile` (a directory at `vendor/modules.txt`
yields `vendor_metadata_inconsistent` after exactly one `list` and no `build`)
and `TestALinkStandingInForVendorMetadataNeverReachesTheDriver` (the frozen
build source refuses the link a whole layer earlier). Both PASS. The second
degrades by `t.Logf` + return rather than `t.Skip` on a host that cannot create
a symlink, which correctly avoids adding a ledger class. Good call.

## 4. "Merged to main green" — the AC clause that failed last cycle

| Run | Commit | Result |
| --- | --- | --- |
| PR 36 CI `32683072667` (`pull_request`) | `c8ac575` | **success**, 02:54:22Z — all 12 jobs green incl. Test(macos), Race(macos), Test(windows), the exact three that were red on PR 35 |
| merge | `b00836c` at 02:54:34Z | 12 s **after** the green run, not 20 s before a red one |
| candidate dispatch `32683167886` | `c8ac575` | **success**, all 14 jobs green — Candidate suite on ubuntu, macos **and windows** |
| main push run `32684608758` | `b00836c` | **success**, all 12 jobs green — including Test(macos), Race(macos) and Test(windows), the three that were red on the `b869a90` main run `32678133350` |

So the clause holds on its own terms: `main` is green at `b00836c`, watched to
completion rather than predicted.

Corroborating: `git rev-parse c8ac575^{tree}` and `git rev-parse b00836c^{tree}` are
the **same tree** `0f98593bcc2766b9c555d1623904af247fd07411`. The fully green PR
run therefore executed against byte-identical content to the merged main head;
the push run is a re-run of the same tree.

The reconciliation the review demanded was also made honestly: the fix-forward
notes and logbook `0508`/`0620` state plainly that "every lane verified green
pre-merge" was **factually wrong** for PR 35, name run `32676699282` and its
three FAILURE lanes, and record that `go test` was exit 0 throughout so the
failure was `platform-case-gate.sh` alone.

## 5. Local evidence at `b00836c`, each command standalone

| Command | Exit | Notes |
| --- | ---: | --- |
| `gofmt -l cmd internal` | 0 | empty |
| `go vet ./...` | 0 | |
| `golangci-lint run` | 0 | 0 issues |
| `go test -json ./internal/godriver` (pin root) | 0 | 362 pass / 0 fail / 20 skip, every skip a declared class |
| `go test -run TestModuleRootVectorsDriveTheWholeBuild ./internal/godriver` (candidate root) | 0 | 10/10 vector cases PASS |
| `go test ./internal/{moduleroots,buildsource,skillspec}` (root unset, as suite-plan defers them) | 0 | |
| `go test ./internal/{moduleroots,buildsource,skillspec,godriver}` (candidate root) | 0 | |
| `go test -run 'TestDeclaredModuleRootsReachTheBuilder\|TestModuleRoot' ./internal/install` | 0 | asserts `BuildRoots` reaches the builder |
| `gate-selftest.sh` | 0 | 81 passed, 0 failed |
| `ledger-consistency.sh` | 0 | 72 rows across linux/darwin/windows |
| `no-broad-suppression.sh` | 0 | |
| `suite-plan.sh <pin>` | 0 | 35/8, godriver served |
| `CI_REQUIRE_FULL_ROOT=1 suite-plan.sh <candidate>` | 0 | 43/0 |

Not run locally: the full `test-gate.sh`. It is exactly what the CI lanes
execute and prior reviewer runs timed out on it; the green PR run on the
identical tree is the authoritative check, per the coordinator's scope note.

## 6. Definition of Done

All items hold. Two that needed checking rather than reading:

- *Tests green* / *module-roots vectors consumed and green* — verified above in
  both root shapes, not taken on the note's word.
- *Findings recorded in logbook* — this was **unsupported** at cycle 1 and is
  now satisfied: working-tree `LOGBOOK.md` carries `0620` (the fix and the
  substantive containment-backstop finding), `0508` (the cycle-1 root cause),
  and `0627`, which records the divergence itself honestly rather than papering
  over it.

## 7. Carried forward, not blocking

1. **`origin/main:LOGBOOK.md` is a 4-line unrelated blob** while the maintained
   3257-line logbook exists only in the unpushed local `main` lineage; the two
   share no history. Logged as `0627`. Pre-existing, needs a repo-lineage
   ownership decision (force-carry the long file, or accept the short one and
   abandon the history). Not this task's to resolve, but every future DoD
   logbook claim on this repo is compromised until someone does.
2. **The registry `clockSkew` flake still has no board item.** Both cycles agree
   it is real: `registry.checkSnapshotsWithPolicy`
   (`internal/registry/snapshot.go:75`) defaults `maxAge` when zero but not
   `clockSkew`, and the install test env builds a bare `config.Config`, so the
   effective tolerance is 0 s instead of the production 300 s; a second boundary
   crossed between `install`'s `time.Now()` and the fixture's whole-second
   `created_at` trips `snapshot timestamp is too far in the future`. Correctly
   declared out of scope by the fix-forward. It needs its own item against
   `internal/registry` — the coordinator should file it.
3. `internal/install/external.go:293` builds its `StageRequest` with no
   `Modules`/`BuildRoots`/`RuntimeRoots`. Harmless today (the declaration half
   short-circuits on an empty `modules` list, and the post-`go list` replace-set
   check applies uniformly by design), and it predates PR 36 — noting it only so
   the next person to give the external-repository path a module-root
   declaration knows the plumbing stops there.

## 8. Handoff

Reviewer-archetype run: **no `commit_ack` supplied.** The scope is already
committed and merged as `b00836c`; the commit-owning mover makes the final
`done` transition with `commit_ack=scope_committed`.
