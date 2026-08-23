# TASK-260728-1g0z69 — rework cycle 7

Closes the single blocking finding in
`TASK-260728-1g0z69_review-verdict-cycle-7.md`. No other change.

Candidate worktree: `/Users/iv/Developer/ReluxWorks/.temp/TASK-260728-1g0z69/curator-spec-worktree`.
Delta versus the accepted predecessor (`TASK-260728-2kp3tv`) is still exactly
three files: `CHANGELOG.md`, `decisions/0007-compiled-build-toolchain-preflight.md`,
`docs/compiled-build-toolchain-requirements.md`. Nothing staged, committed,
published or pinned; no platform validation claimed for rust/swift/kotlin/jdk.
The predecessor worktree is untouched (127 dirty entries at `57c1f56`, nothing
staged); the candidate is at the same commit with 128, the one extra entry being
untracked `decisions/0007-…md` — `docs/` is wholly untracked and git collapses it.

## 1. The finding, checked against upstream source rather than accepted

The reviewer's claim is that `probe/main.go:720-722` recognised *both* the exact
`invalid GOTOOLCHAIN "v"` form and **every** colon suffix after it, and that the
colon-bearing `Select` diagnostics quote the environment setting rather than the
value under test. Read from both probed toolchains' own
`src/cmd/go/internal/toolchain/select.go`:

| line | `base.Fatalf` | quotes | reachable under `GOTOOLCHAIN=local+path`? |
|---|---|---|---|
| 155 | `invalid GOTOOLCHAIN %q: invalid minimum toolchain %q` | `gotoolchain` = **env setting** | no — `min` is `local`, so the branch is skipped |
| 157 | `invalid GOTOOLCHAIN %q` | `gotoolchain` = **env setting** | no — same branch |
| 163 | `invalid GOTOOLCHAIN %q: only version suffixes are +auto and +path` | `gotoolchain` = **env setting** | no — the suffix is `path` |
| 198 | `invalid toolchain %q in %s` | the **go.mod** name | yes |
| 282 | `invalid GOTOOLCHAIN %q` | `gotoolchain`, **reassigned** from go.mod at line 171/208 | yes |
| 351 | `cannot find %q in PATH` | the selected name | yes |

So the reviewer is right on both points, and the second one sharpens the first:
the colon-bearing forms are not merely quoting the wrong thing, they are
*unreachable* under this probe's fixed setting. A branch matching
`invalid GOTOOLCHAIN "v":` plus anything was therefore answering `rejected` for a
family no host produces — which is the definition of an unearned verdict, and it
answered in the direction that hides behind an isolated-rejected value.

Measured directly on Go 1.25.1, `go version` under `GOTOOLCHAIN=local+path`, all
13 `toolchain` values: exit 0 for `default` and `go1`; exact
`go: cannot find "…" in PATH` for the four representable names; exact
`go: invalid toolchain "…" in go.mod` for six; exact
`go: invalid GOTOOLCHAIN "go2.0.0"` for one. No colon-bearing form occurs.
Identical on Go 1.25.5.

## 2. What changed

### Probe (`.temp/TASK-260728-1g0z69/probe`, attached)

**Recognition is now whole-line and exact, in both positions.** The classifier is
no longer a chain of predicates over the output; it is a list of *expected
lines*, each built before the command runs from the value under test plus the
constants the probe itself fixes (`runContext`: module file name, directive line
number, `GOTOOLCHAIN` setting, the probed toolchain's own local version). A
non-zero outcome is recognised when one whole trimmed output line equals one
expected line; two expected lines of different states matching inside one output
is unknown rather than first-wins.

| Position | Recognised (exact whole line) | State |
|---|---|---|
| `toolchain` | exit 0 | accepted |
| | `go: cannot find "v" in PATH` | accepted |
| | `go: invalid toolchain "v" in go.mod` | rejected |
| | `go: invalid GOTOOLCHAIN "v"` | rejected |
| `go` | exit 0 | accepted |
| | `go: go.mod requires go >= v (running go L; GOTOOLCHAIN=local)` | too-new |
| | `go.mod:3: invalid go version 'v': must match format 1.23.0` | rejected |
| | `panic: go: internal error: missing go root module` | rejected |
| either | anything else | **unknown → fail** |

This removes four open families, not one. The reviewer flagged the
`invalid GOTOOLCHAIN "v":` lead; `invalid toolchain "v" in <anything>` was a lead
too, and at the `go` position `invalid go version 'v':` matched as a substring
anywhere in the output while the TooNew matcher matched upstream's lead and tail
with anything between them. All four are the same defect, and the three the
reviewer did not name were left open only because nothing had exercised them yet.

The `panic:` abort stays recognised literally and is the one form that names no
value, because it is a fixed internal abort with nothing to echo. That exception
is now stated in the reference, is excluded from cross-feeding by name, and is
not extended to any value-bearing form.

**Closure is now measured rather than asserted.** A new `classifier closure`
section runs per toolchain and classifies 331 outcomes that are deliberately
*outside* the recognised set, requiring each to yield no verdict:

- **measured, unrelated** — a real command, a real non-zero exit, nothing in it
  about the value: `go build ./...` module-loader refusal for the `toolchain`
  position (the cycle-6 outcome), a package compile error for the `go` position.
  Each is classified once against an isolated-accepted value and once against an
  isolated-rejected one, so both directions are covered by real text;
- **measured** — every measured value-bearing outcome classified against every
  other case's value. Outcomes that name no value are excluded and the exclusion
  is printed with its reason;
- **measured, extended** — a measured recognised diagnostic with the structural
  change a later release makes to a message: a tail appended, a wrapper in front,
  the line embedded in a longer one.

The third kind is constructed, and the run labels it so. It has to be: a
fail-closed property is a claim about outcomes that do not exist yet, so no host
can emit them. Taking text upstream did emit and changing it the way a later
release would is the honest form of the check; the alternative — asserting
closure from the recognised set — is not a check.

Every row reports whether a wrong answer would have **laundered**: a fabricated
verdict whose `representable()` equals the isolated measurement compares equal in
the crosscheck, so the row goes green for a reason nobody measured. The summary
splits the count by direction, A (`accepted` on an isolated-accepted value) and
B (`rejected` on an isolated-rejected value).

**New expected-red control** `-red open-classifier` restores the four retired
families behind the exact classifier, so the cycle-7 defect is reproducible from
this binary rather than from a hand-edited copy of it.

### `docs/compiled-build-toolchain-requirements.md`

- §4.2.1.2: new normative bullet requiring whole-line exact recognition against a
  form predicted before the command runs, naming the family problem and the
  conflicting-match rule; both recognised sets restated in their exact rendered
  form, with the environment-quoting `Select` calls explicitly excluded and the
  reason they are unreachable; the `panic:` abort named as the one exception and
  the exception forbidden from spreading; new bullets requiring closure to be
  measured, requiring both directions to be covered and reported separately, and
  requiring the constructed checks to be labelled as constructed.
- §4.2.1.2 run record: 58 measured cases **plus 331 closure checks per
  toolchain**; four carrying measurements become five, the fifth being the
  environment-quoting reading above and the 24 fabrications the restored families
  produce per toolchain.
- §4.2.1.2 control table: four controls become five, with why the last two do not
  substitute for each other — the fourth reaches the acceptance direction with a
  real command form, the fifth is the only one that reaches the rejection
  direction, because every family it restores names `rejected`.
- §8: 127d updated to five controls and to covering both directions; **127e** is
  new (recognition is exact and its closure is measured).

### `decisions/0007-compiled-build-toolchain-preflight.md`

- §4 gains two paragraphs: a closed set is a set of outcomes rather than of
  leads, with the `Select` reading that settles it; and closure is measured
  rather than asserted, with the disclosure that the extended checks are
  constructed and why that is the honest form.
- Two new rejected alternatives: recognising a lead plus whatever tail follows,
  and stating fail-closedness as a property of the recognised set.
- The "superseded classifications" alternative now says five and explains why two
  of them guard closure at different layers.
- `TASK-260728-2jaw7h` obligation: five controls, whole-line exact recognition
  with families forbidden, and a measured closure section reporting both
  directions.
- Reserved rust/swift/kotlin/jdk obligation: each recognised entry must be one
  whole diagnostic matched exactly rather than a lead, and each driver must
  extend the closure section to its own forms and report both directions —
  `cargo`, `swift build` and Gradle all render diagnostics that share a lead with
  unrelated ones, which is the surface that produced this defect.

### `CHANGELOG.md`

One bullet extended: the closed command classifier "recognises whole diagnostics
exactly rather than message leads … and whose closure is itself measured in both
laundering directions". Still one bullet, still a design record.

## 3. Gates — each run standalone, real exit codes

| Gate | Exit | Note |
|---|---|---|
| `validate.py` | **0** | 42 schemas, 422 vector files |
| `python -m unittest discover -s tools` | **0** | 29 tests |
| `go test ./...` | **0** | |
| `go vet ./...` | **0** | |
| `gofmt -l tools` | **0** | no output |
| `git diff --check` | **0** | no output |
| `go run ./tools/generate-vectors -root .` | **0** | |
| probe, green, both toolchains | **0** | 58 cases + 331 closure checks each, 0 failures |
| probe `gofmt -l .` / `go vet ./...` | **0** | no output |
| clean probe `e6bb0bd`: `make regenerate-check` ×2 | **0**, **0** | |
| clean probe `e6bb0bd`: `make release-check VERSION=1.0.0-rc.5` | **0** | `release gate passed for 1.0.0-rc.5 at e6bb0bd…` |

Byte identity against the accepted predecessor worktree after regeneration:
`conformance/v1` identical (`diff -r` exit **0**) and `release/1.0.0-rc.5.json`
identical (`diff` exit **0**).

Reported truthfully as failing, because it is: `make regenerate-check` **inside
the task worktree** exits **2** (make's own code; the underlying
`git diff --exit-code -- conformance/v1 release/1.0.0-rc.5.json` exits **1**),
because it diffs against `57c1f56` while the whole rc.5 candidate is
uncommitted. Determinism is therefore established by the predecessor
byte-comparison plus the clean git-init probe, exactly as in cycles 1–6.

Python gates use the existing `validation-venv`; ambient `python3` lacks
`jsonschema`.

## 4. Expected-red controls — reported as failing, because they are

| Control | Exit | Fails with |
|---|---|---|
| probe `-red open-classifier` | **1** | `outcomes outside the recognised set produce no verdict: MISMATCH`; 24 laundered fabrications per toolchain — **9 direction A, 15 direction B**; 68 failures — **new this cycle** |
| probe `-red unrelated-command-failure` | **1** | `command outcomes inside the recognised closed set: VIOLATED by default, go1` on both toolchains; 4 failures |
| probe `-semantic tidy-exit` | **1** | P1 VIOLATED by `1.26.0`, `1.99.0`, `1.99rc1` on both; 12 failures |
| probe `-red patch-prerelease-compared` | **1** | P1 VIOLATED by `1.23.4rc1`, `1.24.0alpha1`, `1.21.3beta2` on both; 12 failures |
| probe `-red c-equals-upstream` | **1** | security partition subtracts: MISMATCH on both; 2 failures |
| injected broken local link in the reference | **1** | `broken local link: ./does-not-exist-rework-07.md` |
| injected retired descriptor stem in the reference | **1** | `docs/…:1863: retired repository descriptor name is not an alias and must be absent` |

The new control is the direct answer to required rework item 3. It proves both
directions from one run, and the rows name themselves:

- direction **B** (the cycle-7 defect):
  `tail appended on go1x` → `go: invalid GOTOOLCHAIN "go1x": unrelated failure` →
  `VERDICT FABRICATED: rejected; LAUNDERED, direction B`. `go1x` is
  isolated-rejected, so the fabrication agrees and the crosscheck stays green.
  Also fires for `go1.23/../evil`, `go1.`, `go1.99.0rc1x`, `godefault`, `1.23.4`
  at the `toolchain` position and for `1.023`, `v1.23`, `1.23/4` at the `go` one;
- direction **A**: `1.26.0`, `1.99.0`, `1.99rc1` under all three mutations →
  `VERDICT FABRICATED: too-new; LAUNDERED, direction A`, from the restored
  lead-and-tail TooNew matcher.

Fabrications the crosscheck *would* have caught are printed as
`would surface as a crosscheck disagreement` and are counted separately, so the
control cannot be read as covering a direction it does not.

Both injected-defect probes were reverted; the reference file is byte-identical
to its pre-probe copy (`cmp` exit 0) and `validate.py` exits **0** afterwards.

## 5. Preserved

Cycle-6 closure state and narrowest-command rule, `go version` as the
`toolchain`-position command, and the four earlier controls; cycle-5
representability-versus-TooNew separation and the isolated `gover` harness;
P1/P2 and the forbidden-selector partition; the tested-release-family gate;
twelve diagnostics and their totality; revisioned guidance IDs and catalog
transitions; the wire-versus-source-metadata disposition split; Stage A
sub-steps and Stage B ordering; `go 1.23.4rc1` / `1.24.0alpha1` / `1.21.3beta2`
still rejected and `go 1.23rc1` still permitted; `toolchain go1.99.0-custom`
still upstream-accepted and `forbidden`; no auto-install. 16 `go` + 13
`toolchain` values, 58 measured cases, unchanged.

## 6. Reviewer focus

- Is the recognised set now genuinely exact for both positions, or does one of
  the eight forms still admit more than one string — in particular, is
  `go.mod:3:` safe to fix given the probe writes the module file itself, and is
  `; GOTOOLCHAIN=local` the complete `TooNewError` explain tail under this
  probe's settings (`Startup.AutoFile` is empty because `mode` is neither `auto`
  nor `path` at the `go` position)?
- Is exempting `panic: go: internal error: missing go root module` from the
  cross-feed correct, given it is the one form that names no value, or does the
  exemption itself need a control?
- Do the three mutations cover the shapes a later release actually produces, or
  is there a fourth shape — a reordered rendering, a different quoting — that
  would slip past whole-line matching?
- Is 331 closure checks per toolchain the right scope, or does cross-feeding
  every pair add noise that would hide a real failure in the listing?
