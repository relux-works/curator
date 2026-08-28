# Reviewer verdict for TASK-260811-tkurtl — round 2

Verdict: **changes requested -> to-dev**

- Reviewer run: `RUN-260823-b80fbf` (Claude claude-opus-5), not goal-bound
- Reviewed delivery: rework `RUN-260823-9ae808` against the three CONFIRMED
  findings of `TASK-260811-tkurtl_review-verdict_RUN-260823-fee71e.md`
- Reviewed outcome: `TASK-260811-tkurtl_rework-outcome_RUN-260823-fee71e.md`
- No code, staging, commit, reset, or clean was performed. Temporary probe
  tests and one throwaway `git worktree` at `HEAD` were created, executed, and
  removed; `git status --short` is 22 lines, byte-identical to the state
  received.

## Summary

All three round-1 findings reproduce as fixed against their original probes.
The `closuregraph` contract change is minimal, correctly shaped, and provably
absent-by-default. Acceptance is still blocked by three CONFIRMED defects, each
demonstrated by an executed probe:

- two of them show that round-1 finding 2 is only **half** closed — the scanner
  now fails closed on a non-literal *operand*, but still fails **open** on
  whether a line is a directive at all, and never scans transitively;
- one is a **new false rejection introduced by the finding-1 fix**, against
  scope item 1 which round 1 had accepted.

## Per-item verdict

| # | Scope item (round-2 brief) | Verdict |
| ---: | --- | --- |
| 1 | Finding 1 closed — selection neutrality (CGP05) | **accepted** |
| 2 | The `closuregraph` `ConsumesInteropPayload.Condition` contract change | **accepted** |
| 3 | Finding 2 closed — fail-closed include scanner | **changes requested** (findings A and B) |
| 4 | Finding 3 closed — `publicHeadersPath` honored | **accepted** |
| 5 | No regression of accepted items 1, 4, 6, 7, 8 | **changes requested** (finding C regresses item 1) |
| 6 | Evidence | **accepted** |

## Findings

### Finding A — CONFIRMED — the include scan is not transitive (scope item 3)

`internal/swiftpminterop/interop.go:379` scans exactly
`interop.Sources ++ headerPaths(interop.Headers)`, and `interop.Headers` is the
inventory of the **public-header root only** (`interop.go:367`). A header that
lives inside the target tree but outside that root — the conventional private
header layout SwiftPM explicitly permits — is admitted as an include *reference*
and then never opened. Every directive it declares is invisible.

Failure scenario, executed:

```go
"Sources/CLib/lib.c":     "#include \"CLib.h\"\n#include \"private.h\"\nint value(void) { return 1; }\n",
"Sources/CLib/private.h": "#include </etc/passwd>\n",
```

```
ACCEPTED includes=[{root CLib Sources/CLib/lib.c CLib.h false false}
                   {root CLib Sources/CLib/lib.c private.h false false}]
ACCEPTED: private.h include set never scanned
```

The closure succeeds. `private.h` is recorded as a resolved admitted reference;
`/etc/passwd` appears in no include set, in no `Resolution`, and raises no
`swiftpm_header_input_undeclared`. `Reads.Mode` is `not-observed`, so — exactly
as round 1 established — the declared static closure is the entire header proof
in the only mode reachable until `TASK-260811-2qfnai` supplies a
`ReadSetProvider`. This needs no exotic spelling: it is the ordinary C layout.

Required: scan the transitive closure of resolved admitted include references,
not just the declared sources and the public-header root, and fail closed on a
reference whose target this stage cannot open and scan.

### Finding B — CONFIRMED — directive *recognition* still fails open (scope item 3)

`headers.go:29`:

```go
var directivePattern = regexp.MustCompile(`(?m)^[ \t]*#[ \t]*(include|import|include_next)\b([^\n]*)$`)
```

`literalIncludeOperand` correctly makes the *operand* fail closed — the
reviewer's original macro probe now rejects, verified below. But a line this
regex does not match is still silently invisible rather than rejected, which is
not the module-map parser's discipline the finding asked it to match. Four
spellings that the pinned Apple Clang on the accepted Darwin profile actually
honours pass straight through.

Failure scenario, executed — each body in place of `Sources/CLib/lib.c`, all
four ACCEPTED with the escaping include absent from the recorded include set:

| Spelling | Body fragment | Curator | `clang -std=c17 -H` |
| --- | --- | --- | --- |
| spliced keyword | `#inc\`⏎`lude </etc/passwd>` | ACCEPTED, dropped | reads the header |
| comment-prefixed | `/* */ #include </etc/passwd>` | ACCEPTED, dropped | reads the header |
| form-feed-prefixed | `\f#include </etc/passwd>` | ACCEPTED, dropped | reads the header |
| digraph `%:` | `%:include </etc/passwd>` | ACCEPTED, dropped | reads the header |

All four were confirmed against the real compiler, not inferred:

```
splice:   CLANG READS secret.h
comment:  CLANG READS secret.h
formfeed: CLANG READS secret.h
digraph:  CLANG READS secret.h
```

(Backslash-newline splicing is translation phase 2 and a comment becomes white
space in phase 3, both before directive recognition; `%:` is a C99/C++ digraph
for `#`; form feed is white space.) Two controls behave correctly and are worth
recording: `#include \`⏎`</etc/passwd>` and a CRLF literal both **reject** with
`swiftpm_header_input_undeclared`.

Same class, lower severity: `moduleImportPattern` (`headers.go:30`) requires
`@import` at line start, so `int x = 1; @import Secret;` is also dropped
silently (executed, ACCEPTED).

Required: recognize the directive on the phase-2/phase-3 translation the target
compiler actually performs (at minimum splice the backslash-newlines and strip
comments before matching, and accept the `%:` digraph and the full white-space
set), and reject any residual `#`-introduced line this grammar cannot classify,
so an unrecognized line is a rejection rather than a hole. Add H-family vectors
for the four spellings above.

### Finding C — CONFIRMED — a conditional `.interoperabilityMode(.Cxx)` hard-rejects a fully pruned declaration (scope item 5, regressing accepted item 1)

The rework outcome states the declaration-level gate "stays unconditional
because it is a declaration defect, not a destination fact". That premise does
not hold: the input it tests, `consumer.CxxInteropMode`, is itself
destination-evaluated — `cxxInteropSelected` (`platform.go:146-160`) resolves
the setting's `Condition` against `state.markers`. So `boundaries.go:85`

```go
case interfaceCxx && consumer.Kind == KindSwift:
    return Boundary{}, failFields(CodeInteropUndeclared, ...)
```

now fires on a destination where the entire C++ declaration is pruned, because
`deriveBoundaries` correctly walks the edge neutrally while `CxxInteropMode`
does not.

Failure scenario, executed — `App` declares both its `CxxLib` dependency and its
`.interoperabilityMode(.Cxx)` opt-in under `.when(platforms: [.macOS])`:

```
darwin err=<nil>
linux  err=closure_interop_undeclared: provider exposes a C++ interface but the
       Swift consumer does not declare interoperabilityMode(.Cxx); SwiftPM never
       propagates it implicitly
```

A package that is entirely valid on Linux — no C++ dependency is selected there
at all — cannot close. Capture neutrality is not even measurable on this shape
because one branch errors out. Before the rework this shape closed on both
destinations (the selected-only walk never classified `CxxLib` on Linux), so
this is a regression introduced by the finding-1 fix, on the interop-boundary
correctness item round 1 accepted.

Control, correctly unchanged: with the dependency **unconditional** and only the
opt-in conditional, Linux rejects with `swiftpm_target_platform_unsupported`
("C++ family has no accepted standard/toolchain profile") from
`requireLanguageProfile` — the right code for a C++ target actually selected on
a destination with no accepted profile.

Impact is a false rejection, not an admission hole, so it is the least severe of
the three — but it is a new defect in delivered behavior.

Required: make the declaration-level C++ gate condition-neutral (test whether the
consumer declares `interoperabilityMode(.Cxx)` at all, independent of markers),
or gate it on `boundary.Selected` like the two destination gates beside it.
Add a conditional-opt-in vector to the S05/S06 family.

## Accepted items — evidence

**Item 1 — finding 1 closed, selection neutrality (CGP05).** Reproduced with the
reviewer's original probe verbatim (`.when(platforms: [.macOS])` on App→CLib,
same admitted closure, both accepted destinations):

```
INTEROP capture digests: darwin=sha256:b1c1e85c… linux=sha256:b1c1e85c… equal=true
darwin boundaries=1 targets=2 | linux boundaries=1 targets=2
darwin boundary[0] selected=true  cond=&{swiftpm-condition-v1 platform=macos}
linux  boundary[0] selected=false cond=&{swiftpm-condition-v1 platform=macos}
darwin target[1]=CLib selected=true | linux target[1]=CLib selected=false
active:   darwin=sha256:98e59246… linux=sha256:d02b3853… equal=false
evidence: darwin=sha256:523c7d1e… linux=sha256:e1f81c9b… equal=false
```

Byte-identical capture, opposite selected/pruned verdicts, differing binding,
active projection and evidence — exactly the invariant round 1 required. The
design is sound rather than incidentally passing: `declaredReach`
(`interop.go:281`) seeds from the destination-selected set and then ignores
every condition, and because the product's root targets are always selected and
neutral reach is monotone, `neutralReach(selected) = neutralReach(productRoots)`
on every destination. `CGP05` now has a real conditional branch
(`TestCGP05ConditionalEdgeKeepsInteropCaptureSelectionNeutral`), and
`TestPrunedConditionalDependencyIsCapturedAndProjectedPruned` moved its
assertions to the active projection as asked. `selected` appears only in the
evidence digest (`graph.go:386`, `graph.go:394`), never in a capture node; the
pruned and selected `provides_interop` branches both use `provider.NodeID` for a
non-system provider, so the capture edge is identical either way and only the
binding overlay differs (`graph.go:300-315`).

**Item 2 — the `closuregraph` contract change.** Minimal and correctly shaped:
`ConsumesInteropPayload.Condition *Condition` mirrors `RequiresPayload`
(`edge.go:170`) field for field, through `condition()`, `validate()`, `value()`,
and the decoder's optional-field list. No other payload's `condition()` was
widened — the table at `edge.go:261-272` still returns `nil` for the other nine
kinds. The stated justification is mechanically correct:
`selectionReachability` (`projection.go:179-181`) traverses `provides_interop`
in reverse, so an unconditional consumer side would keep the boundary — and
through it the provider — reachable on every destination, and `ProjectActive`
could not record the pruned verdict.

Absent-by-default canonicalization is proven, not argued. I built a throwaway
`git worktree` at `HEAD` (pre-rework) and hashed the same unconditional
`consumes_interop` edge in both trees:

```
HEAD (pre-rework):    sha256:473b312435a989887a0673fc8cec00fa2bcf2b615ae630811cdbd722e7a68d27
worktree (reworked):  sha256:473b312435a989887a0673fc8cec00fa2bcf2b615ae630811cdbd722e7a68d27
```

`internal/closuregraph/testdata/canonical-goldens.txt` is unmodified, the Ruby
oracle reports `canonical_goldens=pass labeled_records=53
cgp05_target_branches=2 cgp10_observation_branches=2` and
`canonical_references=pass cgp05_capture_reused=true
explicit_target_bindings=2 cgp10_all_refs_resolve=true`, and the generic
machinery treats the new conditional edge exactly like every other one
(`validation.go:334-403`, `validation.go:1136`, `checkpoint_evidence.go:169`).
The field has exactly one production writer outside the package —
`internal/swiftpminterop/graph.go:293` — so nothing else abuses it.

**Item 4 — finding 3 closed, `publicHeadersPath` honored.** The reviewer's
original probe (`Sources/CLib/pub/module.modulemap` with `header "/etc/passwd"`
alongside `pub/CLib.h`, no `publicHeadersPath` declared) now rejects instead of
being shadowed by a generated map for `include`:

```
PROBE3a rejected: swiftpm_modulemap_escape: admitted target tree contains a
                  module map outside the resolved public-header root
```

With the layout actually declared (`PublicHeadersPath: "pub"`), the real map is
located and parsed, and its escaping header rejects on the containment path:

```
PROBE3b rejected: swiftpm_modulemap_escape: reference names an absolute path
                  outside the admitted package
```

The `swiftpmsource` side is genuinely additive: `Target.PublicHeadersPath`
(`types.go:70`) decoded from `dump-package` (`executor_runtime.go:418`) and
bound into `manifestDigest` under `public_headers_path` (`manifest.go:115`),
with no existing field or behavior changed. That last one shifts every
`swiftpm-normalized-manifest-v1` digest value, which is correct here — no
persisted golden or receipt of that record exists anywhere in the repository,
and binding the declaration means a silent public-header relocation is now
manifest-replay drift. `publicHeaderRoot` (`headers.go:66`) rejects absolute,
Windows-drive, backslash, control-character, and parent-escaping declarations
with `swiftpm_target_platform_unsupported` rather than substituting a default,
and `confineModuleMapLayout` (`headers.go:99`) is the conservative guard round 1
proposed. `H10` and `H11` cover eight subtests including both silent-pass
layouts.

**Items 5 partial — seam, guard, honesty, and scope hygiene preserved.**
`grep -l exec.Command` over `internal/closureexec`, `internal/swiftpmsource`,
`internal/swiftpminterop` production files still returns exactly
`closureexec/acquisition.go` and `closureexec/portable_runner.go`; no production
file in `internal/swiftpminterop` imports `os/exec`; the guard allowlist in
`guard_test.go` is unchanged. `verifyReads` (`boundaries.go:151`) is still
honest: `not-observed` without a provider, `observed` only when every selected
target reports observed evidence with its issued receipt, and `AssuranceVerified`
still fails closed without a full observed set. Skipping pruned targets there is
correct — a pruned target is never compiled, so it neither supplies nor weakens
read evidence. No Kotlin or Gradle reference anywhere in the package, no
non-`.go` file so no vendored compiled payload, and no creep into
`TASK-260811-2qfnai` or `TASK-260811-x611eq`. `S02`-`S09`, `H01`-`H08`,
`P01`-`P09`, `CGN03`, `CGN09`, `CGN15` all present and green.

**Item 6 — evidence.** Every gate I reran reproduced:

| Gate | Result |
| --- | --- |
| `go test -count=1 -cover ./internal/swiftpminterop/ ./internal/swiftpmsource/ ./internal/closuregraph/ ./internal/closureexec/` | exit 0; interop coverage **86.0%** |
| Test matrix, `internal/swiftpminterop` | **62** top-level, **114** including subtests (was 58/96) |
| New rework vectors, verbose | `TestCGP05Conditional…`, `TestH09…` (6 subtests), `TestH10…` (2), `TestH11…` (6), `TestConditionalProduct…`, `TestPrunedConditional…` — all PASS |
| `go test -timeout 9m -count=1 $(go list ./... \| grep -v cmd/curator)` | exit 0, **51 ok**, 0 FAIL |
| `go test -race` over the three changed packages | exit 0 |
| `golangci-lint run ./internal/swiftpminterop/... ./internal/swiftpmsource/... ./internal/closuregraph/...` | `0 issues.` |
| `go build ./...`, `go vet ./...`, `gofmt -l internal/ cmd/` | clean |
| `ruby .research/260811_cross-language-closure-canonical-golden-verifier.rb` | both lines pass, 53 labeled records |
| `task-board validate` | `Board is valid. No issues found.` |
| Orchestrator `TASK-260811-tkurtl_full-go-02.log` | SHA-256 **matches** `1ea1d064…b69f89` in the resource description; `EXIT:0`, **52 ok**, 0 FAIL, `cmd/curator 470.844s` |

Accepted rather than rerun: the monolithic full suite, on the hash-verified
orchestrator log above. My 51-package run plus that log's `cmd/curator` entry
account for all 52 packages. The producer's own `-race`/`-cover` logs were
independently reproduced rather than accepted.

## Routing

`TASK-260811-tkurtl` -> `to-dev`. Findings A, B, and C each have a concrete
executed reproduction and a concrete required change. Round-1 findings 1 and 3
are closed and should not be reopened; round-1 finding 2 is half closed — keep
`literalIncludeOperand` and `H09`, and extend the scanner to directive
recognition and transitive scanning.

As a reviewer-archetype run this supplies no `commit_ack`.
