# Review findings — stage (b) cycle 3, `feat/agent-environments-stage-b` @ `1a936e77`

Reviewer run `RUN-260906-0644ed`. Subject: curator worktree
`/Users/iv/Developer/ReluxWorks/.worktrees/curator-stage-b`, 22 signed commits past
`origin/main`, PR https://github.com/relux-works/curator/pull/60 (`headRefOid` =
`1a936e77490df3008efe08fa0aba473ff77575c6`, the exact head reviewed). Rework 2 is
`7bca4d4b..1a936e77` (5 commits, 9 files, +138/−78); the full stage is
`git diff origin/main..HEAD` (28 files, +7133/−40). Authority: curator-spec `f39f4a9`.

Everything below was driven. Every mutant was applied by me to a throwaway `rsync` copy under the
story worktree's `.temp/review3/mut/<name>/`; the producer's worktree was never written to and is
clean at `1a936e77`.

**repeat-of: none.** No cycle-1 or cycle-2 finding class recurs in rework 2. The cycle-2 meta-finding
(F3 → B1: *a gate line attested that the command does not print*) is the one I looked hardest for,
and it does not recur — see §5.

## Verdict: ACCEPT

Zero blocking, zero major. Three minors recorded below, none of them rework-2 regressions and none
of them grounds to return the work; two of the three are inherited from stage (a) and belong to the
orchestrator's routing, not to another revision of this leaf.

## Coverage ratio — 6 of 6 rework-2 rows driven

| row | how I drove it | production call site |
|---|---|---|
| B1 | `test-gate.sh` against both roots + 4 `suite-plan.sh` attacks on mutilated/empty roots | `suite-plan.sh` ← `test-gate.sh` ← `ci.yml:130,197,407` |
| B1b | 3 mutants on `CheckBoundary` + `GOOS=windows/linux/darwin go vet` | `CheckBoundary` ← `internal/envprofile/managed.go:1464` |
| M1 | my own 8-case probe through `run()` + the producer's mutant | `useLocked` ← `cmd/curator/profile.go:159` ← `run()` |
| m2 | my own whole-document probe through `run()` (58 key paths) | `cmd/curator/env.go:cmdEnvStatus` → `json.MarshalIndent` |
| m3 | predicate-emptying mutant | `envprofile.Resolve` ← `TestResolveDriftRepair` |
| m4 | 10 of the 13 report rows re-run as standalone processes | — |

Mutants: **6 applied, 5 killed, 1 survived** (m7 — chased to a redundant clause, not a coverage
hole). Four of the six are narrowing, not deleting: the boundary root is *widened by one level*
rather than removed, `TargetFor` is *weakened to ignore one argument* rather than deleted, the
`..` clause is dropped while the other two clauses stay, and the link predicate is *retargeted*
rather than removed.

---

## 1. B1 — the orchestrator's override, tested rather than accepted

The override required registration in `.github/ci/root-artifacts.tsv` with a `root-unset` ledger
class, over the reviewer's proposed `root-content` tolerated skip, on the reasoning that
`root-content` is policy `allow` in every lane including the candidate lane. **The reasoning holds
and I proved each half of it separately.**

*The mechanism, both roots.* Pinned root materialised read-only
(`git -C curator-spec archive 0ed5c691e9208eea52f21db2fc05e226ce3516fd conformance/v1 | tar -x`),
confirmed to publish neither family.

| invocation | observed | exit |
|---|---|---|
| `test-gate.sh` @ `SPEC_PIN` root | `served=70 deferred=2`; three rows `tol … (tolerated skip: root-unset)`; `28 skips recorded`; `platform-case gate: ok` | **0** |
| `test-gate.sh` @ curator-spec main, `CI_REQUIRE_FULL_ROOT=1` | `served=72 deferred=0`; three rows `ok` (served, passing); `19 skips recorded`; 0 `root-unset`, 0 `stage-deferred` | **0** |

Both numbers match the rework report's claims exactly (28 / 19).

*The attack the brief asked for.* A scratch copy of the curator-spec main root with one family
removed, through the candidate-lane invocation:

```
CI_REQUIRE_FULL_ROOT=1 suite-plan.sh <main root minus schema-cases/launch-env-fragment-v1>
  FAIL  internal/envfragment was deferred, but this lane requires a root that serves the whole module
  suite-plan: served=71 deferred=1 excluded=0
  suite-plan: FAILED                                                   EXIT=1
CI_REQUIRE_FULL_ROOT=1 suite-plan.sh <main root minus schema-cases/agent-environment-marker-v1>
  FAIL  internal/envmarker was deferred, …
  suite-plan: FAILED                                                   EXIT=1
```

It fails **closed**, and `test-gate.sh` refuses to run at all on a non-zero plan
("refusing to run a suite whose shape is already wrong"). The counterfactual is established from the
tables, not assumed: `skip-classes.tsv` gives `root-unset` policy `deferred-only` and `root-content`
policy `allow`, and `CI_REQUIRE_FULL_ROOT` is read by **suite-plan.sh alone**
(`grep -rn CI_REQUIRE_FULL_ROOT .github/` → suite-plan.sh, gate-selftest.sh's two assertions,
ci.yml:408). Under the `root-content` fix the packages would have been *served* in the candidate
lane, skipped with an `allow`-policy reason, and the lane would have gone green with the family
absent. **The override is strictly stronger. It is correct.**

*The preserved intent, driven.* A root whose family directory exists but is empty is `served` —
and the three `t.Fatal`s fire, as the producer claimed:

```
suite-plan.sh <main root, both families emptied>  → served=72 deferred=0, ok
go test -run 'TestFragment…|TestParse…' ./internal/envfragment/ ./internal/envmarker/
--- FAIL: TestFragmentAuthoritativeSchemaCases/valid.json … no such file or directory   (+11 more)
```

*The ledger rows.* The three new rows follow the pre-existing seven `root-unset` rows
(`internal/skillspec`, `internal/marker`, `internal/moduleroots`, `internal/scriptpolicy` ×4) byte
for byte in shape. `platform-case-gate.sh` is untouched by this stage.

## 2. B1b — the boundary fixture, and the slash-semantics decision

`TestCheckBoundary` now builds `root = filepath.Join(t.TempDir(), "environments")` and derives every
value from it. **The assertions still assert what they did**, proven by mutants rather than by
reading:

| mutant (applied to `CheckBoundary`) | outcome |
|---|---|
| `root := filepath.Dir(filepath.Clean(envRoot))` | **killed** — `envfragment_test.go:167: a value outside the environments root must fail the boundary` |
| same mutant with the `outside` assertion deleted | **killed** — `:167: a value below the environments root parent but outside the root must fail the boundary` (F4's own case is live independently) |
| drop the `..` segment clause | **killed** — `:163: a .. traversal must fail the boundary even when it resolves inside` |

The `..` kill is the direct proof the brief asked for: the traversal value still carries a real `..`
segment (it is built by concatenation with `filepath.Separator`, not `filepath.Join`, which would
have cleaned it), and without that clause the prefix check admits it. The out-of-root case is
genuinely outside: `filepath.Join(base, "evil")` against `root = base/environments`.

*The slash-semantics judgement:* **right, and it does not move a platform assumption anywhere.**
`fragment_schema_test.go:checkPath` has exactly one entry point, `validateFragmentCase`, called at
`:306` with the literal root `/manager/environments` over values decoded from the published vector
JSON (`instance.Env[…]`, `SystemPrompt.Path`, `MCP.Path`, `PathPrepend`). No host path can reach it.
The previous `filepath.IsAbs` was not stricter on Windows — it rejected every *valid* vector case
there, so the change makes the same assertion run identically on all three runners rather than
hiding one. Host-path absoluteness was never this function's job and is covered by production
`CheckBoundary` (call site `internal/envprofile/managed.go:1464`) with the now-portable fixture.

*Which half the Windows runner actually proves.* `TestCheckBoundary` is required on all three
runners with **no** tolerated class (`platform-cases.tsv`: `must=linux,darwin,windows skipok=-
class=-`) and it needs no conformance root, so a *deferred* `internal/envfragment` still executes it:
I observed `ok internal/envfragment :: TestCheckBoundary` in **both** my lanes, and the hosted
Windows lane's own uploaded evidence carries the same line (§6). B1b's reported defect is therefore
hosted-proven on the runner that failed it in cycle 2, not proven by construction. Its second half — `fragment_schema_test.go:checkPath` — is **not**: that test is
one of the three deferred at this `SPEC_PIN` (m6), so its slash semantics are proven by construction
plus my curator-spec-main run, not by a Windows execution. Correct call, stated bound.

*Portability sweep of my own.* `GOOS=windows|linux|darwin go vet ./...` → exit 0 for all three (vet
compiles the test files). Grep of every added POSIX-rooted literal in the delta: the remaining ones
are byte-comparison fixtures in `envfragment` (`testFragment`, `TestFragmentJSONCanonical`,
`TestFragmentCodexNameChannel`, `TestFragmentEnvCarriesOnlyEnvObject`) whose values never reach
`path/filepath`; `managed_test.go:67`'s `/tmp/<profile>-source` is the JSON body of a source
descriptor; `managed_test.go:312`'s `/rendered/` match applies `filepath.ToSlash` before the
comparison. Nothing new of the class.

## 3. M1 — per-adapter target resolution, driven through `run()`

Driven through the production entry point (`cmd/curator.run()`), not the helper, with my own probe
(`zz_r3_target_probe_test.go`):

```
--env pi          --target xcode-coding-assistant -> 1  environment_target_unknown: undeclared target "xcode-coding-assistant" for pi
--env opencode    --target xcode-coding-assistant -> 1  environment_target_unknown: undeclared target "xcode-coding-assistant" for opencode
--env claude_code --target xcode-coding-assistant -> 1  secondary fixed-home target … writes are deferred
--env codex_cli   --target xcode-coding-assistant -> 1  secondary fixed-home target … writes are deferred
                  --target xcode-coding-assistant -> 1  secondary fixed-home target … writes are deferred
                  --target ghost-target           -> 1  environment_target_unknown: undeclared target "ghost-target"
--env pi          --target ghost-target           -> 1  environment_target_unknown: undeclared target "ghost-target" for pi
--env ghost-env   --target xcode-coding-assistant -> 1  environment_unknown: explicit operand names the unregistered environment "ghost-env"
```

The producer's own mutant reproduced: `TargetFor` ignoring `adapterID` and delegating to
`TargetByID` → **killed** by `TestProfileUseTargetBoundToAdapter` at `profile_test.go:249`
("use --env pi … = 1, want environment_target_unknown"), through `run()`.

*The `--env`-absent branch is right, not a remaining hole.* §7.6's stated condition is exactly
"A target identifier **not declared by the registry** is `environment_target_unknown`" — the
identifier alone, with no adapter qualifier. Revision 1 declares `xcode-coding-assistant` twice, one
row per adapter, and each row names its own adapter and home, so a `target:`-scoped switch with no
`--env` implicates no adapter that did not declare it. Admitting it is the spec sentence, not a gap.
`splitScope`'s resync path (`envprofile.go:1149`) is unaffected: a scope carries either
`env:` or `target:`, never both, so the new branch is reachable only from `profile use --env … --target …`.

## 4. m2, m3

*m2 — the wire shape is snake_case at **every** level, not just the sampled top.* The producer's
`TestEnvStatusMatrix` asserts the 9 top-level members, which I checked is the *complete* set
(`Status` has exactly 9 fields). I walked the whole document through `run()` instead
(`zz_r3_status_probe_test.go`, after a real `env resolve claude_code --repair`): **58 distinct key
paths observed, 0 non-snake_case**, covering all nine structs including
`homes[].surfaces[].{key,paths,form,state,detail}` and `profiles[].precedence.{winner,placement}`.
Only `StatusRequest` carries no tags, correctly — it is an input struct with func members and is
never marshalled. `git diff` on `status.go` is 58 removed untagged field lines and 58 added tagged
ones: pure tag addition, no behaviour change.

*m3 — the non-vacuity guard fires.* Mutant: the write-through-link predicate changed to match
nothing (`"/rendered/"` → `"/rendered-NEVER/"`) → **killed** —
`managed_test.go:342: claude_code  matched no rendered symlink surface`.

## 5. m4 and the meta-finding — the gate table is true

Cycle 2's instruction was that this revision's gate table is itself under review. I re-ran it. Every
row I re-ran matched the report's own output; nothing was taken on assertion except where stated.

| gate (exact command) | root | my observed result | exit | source |
|---|---|---|---|---|
| `go build ./...` | — | (no output) | 0 | re-run |
| `go vet ./...` | — | (no output) | 0 | re-run |
| `GOOS=windows/linux/darwin go vet ./...` | — | (no output) ×3 | 0 | **added by me** |
| `gofmt -l cmd internal` | — | (no output) | 0 | re-run |
| `golangci-lint run ./...` (v2.12.2) | — | `0 issues.` | 0 | re-run |
| `bash .github/ci/gate-selftest.sh` | — | `81 passed, 0 failed` | 0 | re-run |
| `bash .github/ci/no-broad-suppression.sh` | — | `no-broad-suppression: ok` | 0 | re-run |
| `bash .github/ci/ledger-consistency.sh` | — | `131 rows checked across linux darwin windows` / `ok` | 0 | re-run |
| `bash .github/ci/test-gate.sh` | `SPEC_PIN 0ed5c691` | `go test exit=0, platform-case gate exit=0`; 28 skips; 3 rows `tol root-unset`; 0 `stage-deferred` | 0 | re-run |
| `bash .github/ci/test-gate.sh`, `CI_REQUIRE_FULL_ROOT=1` | curator-spec `f39f4a9` | `go test exit=0, platform-case gate exit=0`; 19 skips; 3 rows `ok`; 0 `root-unset`, 0 `stage-deferred` | 0 | re-run |
| `suite-plan.sh` × 4 mutilated/empty roots | scratch | fails closed / stays fatal (§1) | 1 / n/a | **added by me** |
| `go test ./cmd/curator` | main | `ok … 129s` (my probe run) and inside the two gates | 0 | re-run |
| `gh pr checks 60` | — | see §6 — the report quoted the **old** head's red state and labelled it as such, which was accurate | — | re-run |

The rework report's CI section is honest: it quotes `gh pr checks 60` verbatim, states plainly that
those results are the *reviewed* head `7bca4d4b`'s and not this revision's, names the local runs as
darwin/arm64 only, and lists what it did not run (full `-race` on `cmd/curator`, no candidate-root
lane). Its B1b claim is labelled "established by construction … not by a Windows execution here".
That is the standard cycle 2 asked for. **m4 is fixed and the F3 class does not recur.**

## 6. Hosted lanes on the exact accepted head

`gh pr checks 60`, head `1a936e77` (run 34020466962):

**Every lane green.** `gh pr checks 60` on head `1a936e77` (run 34020466962), verbatim:

```
Test (ubuntu-latest)            pass   3m8s
Test (macos-latest)             pass   8m2s
Test (windows-latest)           pass   40m30s
Race (ubuntu-latest)            pass   9m57s
Race (macos-latest)             pass   18m50s
Lint                            pass   41s
Gate self-test (ubuntu-latest)  pass   6s
Gate self-test (macos-latest)   pass   11s
Gate self-test (windows-latest) pass   22s
Interop conformance gate        pass   23s
Naming gate                     pass   9s
Candidate suite (${{ matrix.os }})  skipping  0     (conditional job; workflow_dispatch only)
```

From the Windows lane's own uploaded evidence (`test-evidence-windows-latest`, artifact
9985904326) — the two things that had to be true there and could not be checked locally:

```
suite-plan: GOOS=windows
suite-plan: root=D:\a\curator\curator/protocol-spec/conformance/v1
defer internal/envfragment    … publishes none of: schema-cases/launch-env-fragment-v1
defer internal/envmarker      … publishes none of: schema-cases/agent-environment-marker-v1
suite-plan: served=70 deferred=2 excluded=0 / ok

ok    internal/envfragment :: TestCheckBoundary
tol   internal/envmarker   :: TestParseAuthoritativeEnvMarkerSchemaCases (tolerated skip: root-unset)
tol   internal/envfragment :: TestFragmentAuthoritativeSchemaCases       (tolerated skip: root-unset)
tol   internal/envfragment :: TestFragmentEmissionMatchesReference       (tolerated skip: root-unset)
platform-case gate: 84 skips recorded … platform-case gate: ok
```

`TestCheckBoundary` **executed and passed on the real Windows runner** — B1b is hosted-proven, not
proven by construction; and B1's registration behaves on GOOS=windows exactly as it does locally.

The regression cycle 2 reported — three schema drivers red on all five Test/Race lanes plus
`TestCheckBoundary` red on Windows — is gone.

## 7. Regression re-drive (the cycle-1 "verified, held" set)

From gate B's own `go test -json` stream at curator-spec `f39f4a9`: **25 environments conformance
cases pass, zero skipped**, including every set this stage was to deliver —
`referenced-claude-code-composed`, `referenced-opencode`, `referenced-opencode-zero-modules`,
`system-prompt-composed`, `mcp-claude-code`, `mcp-codex-cli`, `mcp-opencode`, `mcp-pi-none` — plus
`TestFragmentAuthoritativeSchemaCases`, `TestFragmentEmissionMatchesReference` and
`TestParseAuthoritativeEnvMarkerSchemaCases` served and passing. `stage-deferred` occurs **0** times
in either lane's `skips-observed.tsv`; the class usage this stage removed stays removed.

## 8. Scope and signatures

22 commits, every one `G` (good signature), author `Ivan Oparin <oparin@me.com>`. The delta touches
`.github/ci/`, `cmd/curator/`, `internal/` only — no board files, no `LOGBOOK.md`, no control-root
writes. Producer worktree clean at `1a936e77`; PR head equals it.

---

## MINOR

### m5 — `internal/interop`'s environments vectors are not registered, so the candidate lane can still skip the whole conformance subset

`.github/ci/root-artifacts.tsv` (no `internal/interop` row);
`internal/interop/context_materialization_test.go:75`.

This is the *same class* the orchestrator's B1 override closed for the two schema families, still
open for the vectors that carry this stage's headline deliverable. Observed in my `SPEC_PIN` run:

```
internal/interop  TestConformanceEnvironmentsMonolithic  root-content  allowed-root-content
   conformance root … publishes no vectors/environments.json (pre-environments suite; root-content)
internal/interop  TestConformanceEnvironmentsHeader      root-content  allowed-root-content
```

`root-content` is policy `allow` in **every** lane, the candidate lane included, so a candidate root
that dropped `vectors/environments.json` would pass green with all 25 environments cases silently
skipped — exactly the outcome the override rejected.

**Not a rework-2 regression and not this leaf's defect.** The skip was introduced by stage (a)
(`4b5cd059`, landed on main), and the same shape covers five further interop vector families
(`context-detectors`, `context-versions` ×2, `snapshot-acquisition`) that are wholly outside stage
(b). Fixing only the environments row here would be arbitrary. **Recommendation:** the orchestrator
routes one CI leaf that registers `internal/interop`'s vector families in `root-artifacts.tsv` and
converts those `root-content` skips to `root-unset`, in stage (c) or alongside it. Recorded here
because this review is where it was measured, not because it blocks this head.

### m6 — the three schema drivers assert nothing on any automatic hosted lane at the current `SPEC_PIN`

Consequence of B1 done correctly, and a bound the rework report should have stated in one sentence.
`ci.yml:44` pins `SPEC_PIN: 0ed5c691`, which publishes neither family, so on every push/PR lane the
two packages are deferred and `TestFragmentAuthoritativeSchemaCases`,
`TestFragmentEmissionMatchesReference` and `TestParseAuthoritativeEnvMarkerSchemaCases` skip. Their
assertions run only in the manually dispatched `candidate-conformance` job, or locally against a root
that publishes the families (my gate B: all three `ok`). The skips are recorded in
`skips-observed.tsv`, so this is visible rather than hidden — but "these three do not execute in CI
until `SPEC_PIN` is bumped past `f39f4a9`" is a stated bound the report omits, and the bump is the
thing that must be remembered.

### m7 — the `CheckBoundary` absoluteness clause has no killing test, because it is subsumed

`internal/envfragment/envfragment.go:243`. My third mutant — delete the `filepath.IsAbs` clause,
keep the rest — **survived**: `TestCheckBoundary` still passes. I chased it rather than reporting the
survival as a coverage hole, and it is not one: for any *absolute* `envRoot`, a value passing
`value == root || strings.HasPrefix(value, root+Separator)` is necessarily absolute, so the clause
changes the error message and never the accept/reject set. It would bite only with a relative
manager home, and `Home()` is `filepath.Dir(config.Path)` off `os.UserHomeDir()`. Recorded so the
next reviewer who runs this mutant does not re-derive the reasoning, and so the survival is on the
record rather than absent from it. No action.

---

## Observations, not findings

- The `env status` text renderer still prints an empty `form` for unprovisioned rows (cycle 2 noted
  it); the same empty field shows through the m3 mutant's own failure message
  (`claude_code  matched no rendered symlink surface`). Cosmetic.
- `platform-case-gate.sh` evaluates the ledger's `tol` branch *before* the `deferred-only` policy
  check (`:215` vs `:234`), so a ledger row that tolerates a class on a platform tolerates it whether
  or not the package was deferred. Unreachable here — a `root-unset` skip needs an unset variable,
  which the served stage always exports — and the file is untouched by this stage.
- `fragment_schema_test.go:checkPath` is a test-local re-implementation of the §10.3 path rule rather
  than a call into `CheckBoundary`. Deliberate (it validates vector text, not host paths) and the
  production gate is separately driven; noting the duplication only so a future change to one is
  known to not propagate to the other.

## Probe artifacts

`TASK-260906-2g0bgq_review-probes-stage-b-3.tar.gz` — my two in-package probes
(`zz_r3_target_probe_test.go`, `zz_r3_status_probe_test.go`, both driving `cmd/curator`'s `run()`),
the three runner scripts, the full two-root `test-gate.sh` log, both lanes' `suite-plan.txt` and
`skips-observed.tsv`, the four mutilated-root `suite-plan.sh` logs, the mutant log and mutant table.
Scratch lived under the story worktree's `.temp/review3/`; the producer's worktree and the control
root were never written to.
