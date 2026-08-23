# TASK-260728-1yhuqi — review verdict cycle 3

## Verdict

**CHANGES REQUESTED. Route to `analysis`.**

The cycle-2 repairs are reproducible and close the five cycle-2 findings, but
the contract still violates the accepted additional-language boundary in one
live compiler-macro path and does not make its job-plan parser fail closed over
unknown flag channels. These are decision/security-contract defects, not an
external blocker and not implementation-code rework, so the correct branch is
`analysis`.

## Finding 1 — the contract admits the compiler macros Decision 0008 requires it to reject

Accepted Decision 0008, lines 577–587, requires every additional-language
driver to reject every package-selected code-execution surface before the
compile phase and explicitly includes **procedural and compiler macros** and
compiler plugins. A surface that cannot be rejected deterministically
disqualifies the driver; it cannot be answered with a runtime allowance.

Decision 0011 contradicts that requirement:

- its exhaustive matrix, line 749, explicitly **admits** toolchain-supplied
  macros selected from source syntax, including `@Observable`;
- its residual-exposure section, lines 1144–1150, confirms that source syntax
  selects the macro implementation the frontend loads and treats that execution
  as an admitted compiler-front-end exposure;
- lines 1159–1161 additionally acknowledge that such a macro can read files
  during expansion.

This is live behavior rather than prose drift. Independent replay of the
attached probe reproduced structural check
`S7-toolchain-macro-loads-in-process`: source using `@Observable` compiled with
exit 0 and loaded `libObservationMacros.dylib` from the fingerprinted toolchain
root. Fingerprinting the implementation constrains provenance; it does not turn
source-selected macro execution into rejection.

### Required closure

Preserve the already-supported SwiftPM rejection and direct-`swiftc`
architecture, but make the macro rule conform to Decision 0008:

1. distinguish inert/search-only plugin paths from an actual macro/plugin load;
2. reject every actual compiler macro/plugin load at graph time before the
   compile permit, including in-process toolchain libraries and executable
   plugin servers;
3. change the matrix and canonical policy from `macros: "toolchain-only"` to a
   closed rejection, and remove the residual-exposure language that treats
   macro execution as admitted;
4. add a control proving the retired policy admits `@Observable`, plus a
   positive assertion that the replacement rejects its load plan before the
   compile phase;
5. if the intended product decision is instead to allow fingerprinted
   toolchain macros, explicitly reopen and amend Decision 0008 through its own
   reviewed architecture decision. Decision 0011 cannot silently override it.

Recommendation: reject actual macro/plugin loads. The plan already exposes the
load flags, so this preserves the portable boundary without requiring source
text scanning.

## Finding 2 — unknown flag channels are accepted, contrary to the fail-closed plan claim

Reference section 4.1 says anything the plan grammar does not account for
rejects, but section 4.1.2 narrows totality to tokens already recognized as
path-shaped and explicitly admits that an unrecognized future path carrier may
escape that recognition (lines 884–889).

The executable verifier confirms the gap. After checking the enumerated
executable/plugin/search/source/output flags and bare path-shaped positional
tokens, `VerifyPlan` reaches lines 455–460 and rejects only a non-flag relative
token naming an existing working-directory entry. An otherwise unknown token
beginning with `-` is accepted without a verdict. Thus joined or opaque forms
such as an unrecognized `-new-channel=/absolute/path` are neither rejected nor
classified. The claim that the failure direction is closed is false for an
embedded path channel.

This matters because registry compatibility admits the tested Swift `(6,3)`
family, not only the exact 6.3.2 bytes. A later compatible root has a different
toolchain digest but still passes the family gate; a new source-triggered
plugin/native/process flag shape must fail closed rather than silently pass the
graph permit.

### Required closure

1. Define a closed per-job flag/operand grammar for the measured plan and reject
   every unknown flag, joined form, opaque value carrier, and unexpected
   positional token.
2. Add adversarial vectors for unknown `-flag`, `-flag=value`,
   separator-embedded path/plugin values, relative opaque values, and an
   otherwise-valid plan carrying one extra unknown token.
3. Keep the current path buckets and permit-time re-binding; they remain useful
   after the token grammar is made closed.
4. If a future Swift patch emits a new plan token, require a measured contract
   update before adding it to the allowlist.

## Independent replay

Source: the attached
`TASK-260728-1yhuqi_probe.tar.gz`, extracted into a fresh reviewer directory.

- `gofmt -d .`: empty
- `go vet ./...`: exit 0
- `go test ./...`: exit 0
- `go build -o swiftboundaryprobe .`: exit 0
- native probe on macOS 26.5 arm64 / Apple Swift 6.3.2: exit 0
- cases: 23/23 matched, 0 divergences
- closure checks: 32, 0 verdicts
- expected-red controls: 12/12 failed as required
- structural checks: 44, 0 divergences
- executed P2 admission: one in-closure path, one base-installation path,
  zero rejections, P1 equals P2, green true
- all 12 controls were replayed individually and each exited 1

The repository Go suite also passed when invoked during an initial reviewer
working-directory mistake; that run is not used as Swift-contract evidence.
The producer's focused specification gate remains expected-red only for copied,
unrelated broken links. My direct local rerun could not start because the
review environment lacks the `jsonschema` Python dependency; the two delivered
documents are byte-identical to the worktree copies and their task-scoped link
check is already recorded as 6 links / 0 broken in producer evidence.

## Verified cycle-2 closures

The following are not blockers in this cycle:

- one whole-value compiler-banner grammar and typed rejection partition;
- executed P2 and the three-class runtime-library admission;
- portable `curator-swift-relpath-v1`, root ordinals, linker serialization and
  invoked/resolved process-closure receipt projections;
- exact `-###` insertion construction and LF-only physical-line grammar;
- explicit `blocked_by` trace to `TASK-260729-rhjxtx` and per-measurement
  tracing;
- local/external source-mode identity, deterministic executable extraction,
  cache/receipt/marker boundaries, macOS qualification scope, Windows
  implementation obligations, Linux follow-up, and deferred manager signing.

## Routing evidence

All three prerequisites are `done`:
`TASK-260728-2spy93`, `TASK-260728-1g0z69`, and
`TASK-260729-rhjxtx`. The findings require revision of the decision, reference,
probe, vectors, and controls, followed by another independent reviewer cycle.
They do not require a human-only decision unless the team chooses to reopen
Decision 0008's macro prohibition.
