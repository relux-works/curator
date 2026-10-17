# Review B brief: BUG-261017-8v470p, audit verdict cache must bind effective capabilities

Decide accept or reject. Read the bug, `gw-AUDITCACHE-brief.md`, curator-spec core §4.4 and skillfile-sources §4, and the producer handoff.

1. Cache reuse requires equal effective capabilities on every field the detectors read; identical inputs still hit.
2. Crossover controls exist in both orders across skill, context and MCP paths and fail without the fix (run the narrowing mutant yourself).
3. Old records keep their meaning; no silent reinterpretation.
4. Every new rejection (null prompt_scope, null capability objects) is backed by the spec, with valid, absent and empty-object controls; anything not spec-backed is out.
5. Full remote gate green on the exact candidate; public hygiene holds.

Reject with numbered concrete fixes; accept only when all five hold.

## Additional cross-lane input (2026-10-17)

The external lane sent its patch rebased onto 51e8fca (`BUG-261017-8v470p_candidate-rebased-51e8fca.patch`, sha256 9d608106b3525bb2a8f4f2d848b2875314838dedd695012098f02e22e03f9432) with an observation: in production paths the cross-kind crossover is already masked by package identity (repository, commit, directory), and only direct cache-input controls reproduce the missing capability equality. Check the producer's controls against this: a fix must still bind capabilities in the cache key (latent defect), and its tests must assert package/commit/directory binding independently so that capability equality does not mask the existing package-binding mutation.
