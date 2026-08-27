# BUG-260731-3a5q1p — independent review (RUN-260731-cefd98)

Reviewed: PR 14, head `d345420` "Bind the multi-project dry-run case", base `main` `3a047d5`.
Diff: `internal/install/dryrun_conformance_test.go` +727/-46, `.github/ci/root-artifacts.tsv` 1 line.

**Verdict: changes requested → `to-dev`.** The implementation itself is good work and the first
half of the AC is met and independently reproduced. It is sent back for one unrun verification
that the DoD names explicitly, and which — when I reproduced its mechanism locally — turns out to
hide a real cross-platform failure of the AC's second clause.

## 1. What I verified myself (host: darwin/arm64, Go 1.25.5 via GOTOOLCHAIN=auto)

Fresh `git archive` checkouts of `origin/main` and the PR head; rc.6 root = curator-spec
`b07ef1d51dc6ebcc04cf59a20375d191bf82f6bb` (= `origin/release/v1.0.0-rc.6`,
`protocol_version: 1.0.0-rc.6`); rc.3 root = `00b1688…` (the committed `SPEC_PIN`).

| # | check | result |
|---|---|---|
| 1 | repro on `main` + rc.6 root, `-run TestAuthoritativeDryRunCasesMutateNothingPersistent` | **FAIL** — `dryrun_conformance_test.go:189: published dry-run scope "multi-project" has no executable binding`. Bug premise confirmed. |
| 2 | PR head + rc.6 root, both dry-run tests `-v` | **PASS** — 3/3 cases incl. `compiled-cache-miss-is-read-only` (12.4 s), plus `TestDryRunEffectBindingsSeeWhatARealOperationWrites` |
| 3 | PR head + rc.6 root, `go test -count=1 ./internal/install/...` | **PASS, exit 0** — install 406.7 s, atomicity 365.5 s. **This is the AC's first clause.** |
| 4 | PR head + rc.3 root (the live `SPEC_PIN`), both dry-run tests | **PASS** — no regression at the current pin |
| 5 | both `default:` branches read directly | intact — `scope` (line 292) and `effect` (line 692) still `t.Fatalf`; the new case is not skipped and no assertion is special-cased |
| 6 | PR 14 CI vs `main` CI | identical: only `Test (windows-latest)` red, with a **byte-identical** failing-case set on both (7 × `cmd/curator`, `internal/buildsource TestFrozenTokenRejectsRootReplacement`, `internal/install TestEndToEndInstall`, 3 × `internal/install/atomicity`). All owned by BUG-260731-27h1yc / BUG-260731-fs3dht. **This PR regresses nothing.** Lint, Race, Interop, Naming, Gate self-test, ubuntu and macOS Test all green. |

The `root-artifacts.tsv` line is correct and live: rc.3 publishes `manager-lifecycle.json` but not
`build-drivers.json`, so `internal/install` still defers exactly as before (PR 14's own suite-plan
output confirms it), and a root missing the lifecycle vector now defers instead of `t.Fatal`ing on
an `open` error.

## 2. Blocking finding — the three-platform rc.6 proof was never run, and it hides a real failure

DoD item 3 — *"prove internal/install against the rc.6 root on macOS, Linux, and Windows CI"* — is
unchecked, and correctly so: no `candidate-conformance` run exists for
`task/BUG-260731-3a5q1p-multiproject-dryrun`. `gh run list` shows `workflow_dispatch` runs that day
on `ci/goenv-control-BUG-260731-3gm8kc` and `task/BUG-260731-33v6zz-windows-lane`, but none here.

This was **not blocked**. `.github/workflows/ci.yml` `candidate-conformance` takes a 40-character
immutable `candidate_ref`; rc.6 is exactly that (`b07ef1d51dc6ebcc04cf59a20375d191bf82f6bb`, tip of
the published `release/v1.0.0-rc.6`), and the evidence artifact already records the manifest sha256
for `candidate_manifest_sha256`. The earlier signing blocker was lifted before the branch was
pushed. So the only rc.6 evidence anywhere is macOS-local — mine and the implementer's.

**What that missing lane would have caught.** The new binding drives the *real* trusted-toolchain
boundary (`&goToolchain{...}`, `planEveryProject`, line 380). BUG-260731-fs3dht records, from a
native windows-latest run, that `godriver.Probe(ConfigFromEnvironment(...))` rejects the
actions/setup-go GOROOT with `go_toolchain_missing: trusted GOROOT is not a real directory`. When
`Probe` fails, `planBuilds` (`internal/install/plan.go`) returns a `toolchainInventory` whose rows
carry outcome `toolchain-unavailable` — and rc.6 does **not** admit that outcome
(`reported_build_outcomes` is `[cache-hit, would-preflight-and-build, would-rebuild-untrusted-cache,
corrupt, unsupported]`).

I reproduced that exact condition by running the compiled test binary with an unresolvable
trusted root (`GOROOT=/tmp/…/definitely-not-a-goroot`, rc.6 root, no other change):

```
--- FAIL: TestAuthoritativeDryRunCasesMutateNothingPersistent/compiled-cache-miss-is-read-only
    dryrun_conformance_test.go:291: project 0 reported build outcome "toolchain-unavailable",
    which the published case does not admit:
    [cache-hit would-preflight-and-build would-rebuild-untrusted-cache corrupt unsupported]
```

`internal/install` is never platform-excluded (`.github/ci/platform-exclusions.tsv` lists only
`internal/godriver`), and under rc.6 it is served on all three runners. So as things stand today,
**advancing `SPEC_PIN` to rc.6 does turn Curator CI red on this test on windows-latest** — which is
the second clause of the AC, verbatim.

This is not a demand to fix fs3dht here. It is that the task introduced a new, undeclared
dependency of `internal/install` on a working Windows trusted-toolchain probe, and neither the PR
nor fs3dht's own AC (scoped to `internal/godriver` + `cmd/curator`) records it. Linux is separately
unverified: I read the probe path and it does not go through `probeNativeControls`, so it should
work there, but I could not run it.

### What to do
1. Dispatch `candidate-conformance` on this branch with
   `candidate_ref=b07ef1d51dc6ebcc04cf59a20375d191bf82f6bb` and the recorded
   `candidate_manifest_sha256`, and attach the three per-platform evidence artifacts.
2. Report the Linux and Windows results as they land. If Windows fails as predicted, say so in the
   PR and on the board, add an explicit `blocked_by` on BUG-260731-fs3dht, and note on fs3dht that
   `internal/install`'s rc.6 dry-run case now also depends on its fix — so it is not merged as a
   silent landmine under the next `SPEC_PIN` bump.
3. Either way, item 3 gets checked with real evidence or the dependency is declared. Both are
   cheap; assuming it is not.

## 3. Non-blocking: the ordering rationale is wrong

`planEveryProject`'s doc comment (and §2.1 of the evidence artifact) says
`install --all` *"plan[s] every selected project target … and profiles/manager.md §2.5 takes those
targets in the unsigned byte order of their canonical project identities, which is the order used
here"*, and the board note adds *"a dry run takes no lock, so order is the only remaining 2.5
obligation"*.

Both readings are inverted:
- rc.6 `profiles/manager.md` §2.5 says *"A multi-project operation acquires **project locks** by
  canonical project identity in unsigned UTF-8 byte order."* It governs **lock acquisition**, which
  `managerlock.AcquireProjects` already implements via `CanonicalProjects`. It says nothing about
  planning-iteration order.
- Production `install --all` iterates targets in **alias** order — `selectProjectTargets`,
  `cmd/curator/main.go:264`, `sort.Strings(aliases)` — not canonical-path order.
- A dry run takes no lock, so §2.5's ordering obligation does not apply at all, rather than being
  "the only one left".

Nothing is asserted about order, so no test outcome depends on this and it is not a correctness
defect. But fidelity-to-production is the entire justification for calling two sequential
`Project()` calls "the multi-project operation", so the comment should say what is actually true:
one manager home, one skills root, one shared `FetchedRepos` set, deterministic order chosen by the
test. Worth correcting while the branch is open.

## 4. Non-blocking observations

- **Coarse effect bindings.** `revocation-state` binds to all of `<home>/state`, which subsumes
  `project-lock`, `cache-build-lock`, `manager-home-lock` and `journal`. Since the witness test only
  requires a binding to report *something*, those four are witnessed trivially by `<home>/state`
  existing. Same for `permission-repair`, which is `<home>/cache/build` — identical to
  `compiled-artifact-cache` and proving nothing about repair specifically. The code documents each
  as a consequence of Curator's layout and the implementer flagged it in §7; accepted as-is, but it
  is the weakest part of the anti-vacuity claim.
- **Order is constructed, not asserted.** `canonicalProjectOrder` computes an order the test then
  follows; no mutation would fail if the order changed. Fine given §3 above, but it means the
  ordering claim is decorative.
- `install --all` forces `Fetch = fetch && !opts.DryRun` (`cmd/curator/main.go`), so production dry
  runs never fetch. The test passes `Fetch: true`, which is *stronger* (it arms the surface the case
  forbids) and matches the pre-existing project/global cases. Intentional and correct.
- **`t.Parallel()` removal is justified.** `isolateTempDir` uses `t.Setenv`, which forbids a
  parallel ancestor, and that isolation is what makes `operation_private_state_after: absent`
  bindable. It is a scheduling change, and it strengthens the older two cases as well.
- **Anti-vacuity test is the right idea and it works.** I verified mutation 1's shape holds:
  `TestDryRunEffectBindingsSeeWhatARealOperationWrites` runs every binding backwards after real
  project + global installs and a real locked operation, and asserts completeness, so a future
  effect cannot gain a binding without a witness. 21 of 27 surfaces come from genuine production
  paths; the other 6 are produced through their owning packages' own allocators and reserved names,
  with the reason recorded inline. That is a defensible split.

## 5. Summary

AC clause 1 (`go test ./internal/install` green against an rc.6 root, no weakening) — **met, and
independently reproduced end to end.** AC clause 2 (`SPEC_PIN` → rc.6 does not turn Curator CI red
on this test) — **not demonstrated, and currently false on windows-latest** by a mechanism I
reproduced locally. Rework is one CI dispatch, one honest report of its result, one declared
dependency, and one corrected comment.
