# TASK-260728-1yhuqi — review verdict cycle 2

## Verdict

**CHANGES REQUESTED → `analysis`.**

Cycle-1 rework closes the four findings previously recorded: the manager
classifier is line-1-only; the plan verifier rejects the measured malformed
families and rechecks path bindings; `curator-swift-module-v1` is total over the
declared command-key grammar; and the matrix explicitly covers response files,
scripts, settings/configurations, and the platform-derived closure.

Acceptance is still withheld because the revised contract does not yet define
one enforceable toolchain/stdlib identity across the normative documents and
the implementation probe, and its Windows identity remains underspecified.

## Independent passing evidence

- Extracted `TASK-260728-1yhuqi_probe.tar.gz` into a reviewer-only temporary
  directory. `gofmt -l .`, `go vet ./...`, `go build ./...`, and
  `go test -count=1 ./...` all exited 0.
- Replayed the full native probe on macOS 26.5 arm64 / Apple Swift 6.3.2:
  23/23 cases matched; P1 and P2 alignment held with a non-empty security
  partition; 32 closure checks yielded no verdict; all 9 expected-red controls
  fired; all 30 structural checks matched; overall `green: true`.
- Replayed C1 through C9 individually. Each exited 1 as required.
- The reviewer fixture's summary, alignment, structural `(id, verdict)` set,
  and control `(id, failed, findings)` set are byte-equal after canonical JSON
  projection to the producer fixture.
- S17 in the executable probe correctly checks
  `graph_args = ["-###"] + compile_args`; S18 rejects all 20 supplied malformed
  plan families; S19–S21 exercise resolved symlink containment, absent-path
  appearance, and executable identity change against real filesystem state.
- No product file, producer artifact, release pin, staged file, commit, or
  platform claim was changed during review.

## Required rework

### R1 — one exact compiler-version normalization rule

The contract currently has three different rules:

1. decision 0011 section 11 admits
   `^Apple Swift version X.Y.Z \(.*\)$`, including empty parentheses;
2. the implementation reference section 1.2 admits
   `^Apple Swift version X.Y.Z \(.+\)$`, requiring a non-empty suffix; and
3. the probe's `NormalizeCompilerVersion` checks only the prefix and strict
   numeric token before the first space. It never validates the remaining
   suffix, parentheses, or whole value, so a value such as
   `Apple Swift version 6.3.2 x` is accepted by the probe while rejected by
   both documented regexes.

Choose one whole-value grammar, use it identically in the decision, reference,
registry contract, and probe, and add positive/negative vectors for empty,
missing, malformed, and trailing suffix forms. Add an expected-red control that
restores the current prefix-only parser so the new negatives prove the change.

### R2 — close and actually execute native runtime-library admission

Reference section 1.3 and decision 0011 section 5 require only that each P2
`runtimeLibraryPaths` entry *already inside* the toolchain root exists. They do
not reject an empty list or an existing entry outside every fingerprinted root.
That leaves the standard-library closure open in exactly the case the identity
is meant to exclude.

The supplied probe also does not run declared probe P2
`swiftc -print-target-info -target <native-triple>`. `readTargetInfo` runs P1
only, sets `StdlibPresent` when it happens to find one existing in-root P1 path,
and the fixture's green predicate never gates on `StdlibPresent`. The command
evidence contains no P2 invocation for `arm64-apple-macosx26.0`.

Define a non-empty, exhaustive P2 rule: every returned runtime-library path
must be absolute, resolve successfully, be a directory, and resolve inside a
named fingerprinted root (or explicitly define and justify a different closed
set). Run P2 with the exact compile triple, record its complete result, gate
green on it, and add negatives for empty, relative, absent, dangling, and
out-of-closure paths.

### R3 — make the Windows toolchain identity serializable before implementation

`curator-swift-toolchain-v1` currently serializes macOS constants:
`usr/bin/swiftc`, `usr/bin/swift-frontend`, `usr/bin/clang`,
`usr/bin/ld`, followed by exactly the macOS toolchain and one SDK root.
The runtime closure is instead defined structurally from job-plan executables
plus the linker P3 actually resolves. The Windows section correctly refuses to
guess a member count, but it also permits “one or more” `platform-sdk` roots and
does not define:

- how the P3-resolved Windows linker relpath is serialized when it is not
  `usr/bin/ld`;
- how additional plan-derived executable members enter identity;
- how multiple SDK roots are role-named, ordered, presented, and hashed;
- the Windows `link_support_roles` registry value; or
- the closed root-role/member schema that admission task
  `TASK-260728-251p01` must mint.

No Windows qualification is required here, but the identity algorithm and
registry shape must be implementable without inventing those semantics. Define
the platform-parametric serialization now; retain the explicit no-claim state
until a native Windows probe supplies the values.

### R4 — remove the remaining exact-vector ambiguity

The normative documents display `swiftc` in both command lines and then say the
vectors “differ in exactly one token, at index 0.” Read literally as complete
argv, the difference is an insertion after the executable, not a changed token
at index 0. The probe uses the correct argument-only rule:

```text
program = absolute swiftc
graph_args = ["-###"] + compile_args
```

State that rule verbatim in both normative documents and conformance vectors.
Also decide whether plan lines are LF-only or whether one terminal CR is
normalized: the reference grammar rejects every control byte, while
`VerifyPlan` silently removes one trailing CR from every line. The Windows
implementation must not have to choose between those rules.

### R5 — restore task traceability for the measured prerequisite

The task and both outcome documents rely on accepted
`TASK-260729-rhjxtx` measurements, but the board's `blockedBy` list contains
only `TASK-260728-2spy93` and `TASK-260728-1g0z69`. Add the missing dependency
or record a task-scoped justified-gap statement explaining why it is evidence
rather than a prerequisite. The current checked “Dependencies linked” item is
not supported by the board projection.

## Routing

These are research/design corrections with runnable evidence, not product-code
defects and not an external or human-only blocker. Route to `analysis`, revise
the decision/reference/probe and task traceability, replay the full native
fixture plus the new expected-red controls, then request another independent
review cycle.
