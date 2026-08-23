# TASK-260811-33ukne solution architecture

Status: prepared for review

Run: `RUN-260823-f8588f`

## Decision

Keep `TASK-260811-33ukne` as the single atomic implementation task. Extend its
authorized implementation surface across `internal/closureexec`,
`internal/artifactpolicy`, and `internal/swiftpmsource` only as needed to close
the SwiftPM source-control acquisition and mirror chain. Do not create a new
Story, research task, generic SCM framework, or verified-binary capability.

Remote source-control acquisition is not an offline evidence derivation. The
accepted contract explicitly gives acquisition its own network-capable,
manager-owned broker boundary at C2, while manifest, mirror, and metadata
derivations use admitted inputs and `network=none`. Therefore:

1. Add canonical `source-acquisition-permit-v1` and
   `source-acquisition-receipt-v1` records at the shared execution boundary.
   The permit binds the causal head, source profile, exact canonical origin and
   immutable requested revision, exact C0 Git executable/toolchain identity,
   argv, cwd, sanitized environment, allowed process family, quarantine-only
   writes, broker network policy, resource limits, expected object/snapshot
   evidence, and an immediate toolchain recheck. The receipt binds the issued
   permit, actual exit/evidence/output data, exact origin and revision, object
   and snapshot identities, diagnostics, assurance mode and actual
   capabilities, and the next causal head. Portable mode must not claim
   lossless host observation; verified-only claims still require the existing
   healthy provider contract. The broker runs no package code.
2. Extract the selected immutable source tree and admit that complete tree as
   `dependency_input` through the existing shared artifact service before its
   manifest can run. Compiled, opaque, ambiguous, unsafe, linked, or incomplete
   source-tree content remains deny-dominant. The quarantined remote Git object
   database is acquisition evidence, not an admitted dependency tree and not a
   substitute artifact manifest.
3. From the admitted source tree plus the acquisition receipt's exact Git
   commit/tree/object evidence, synthesize a minimal same-kind local repository
   in an absent destination under an ordinary C4 `derivation_permit` with
   `network=none`. Preserve the original pinned revision, use a shallow boundary
   where history is intentionally absent, generate only deterministic Git
   metadata, and verify the revision/tree/object closure using the exact C0 Git
   tool before issuing the derivation receipt.
4. Add a narrowly scoped artifact-policy authorization issuer for the
   `source-control-mirror-v1` local-output transform. It may authorize only an
   exact mirror output produced by a verified issued derivation receipt from
   the named admitted source tree and acquisition evidence. It does not admit
   arbitrary Git stores, cannot relabel dependency binaries, and does not make
   `verified-binary-v1` available.
5. Replace `internal/swiftpmsource` custom Git permit/receipt IDs and direct
   process authority with the two shared planes. The generated-lock receipt
   must reference the actual issued acquisition and derivation receipts; C2/C3
   must reference the exact intake/artifact evidence; C4 replay must consume
   only the authorized kind-preserving mirrors.
6. Reject dirty or escaping C0 Git relative paths after symlink-aware
   containment under the declared execution root before any process start.

## Development-ready verification

The existing task remains one clear deliverable: a trustworthy,
selection-neutral `swiftpm-source-v1` capture and active closure through C4.
The implementation is accepted only when all of the following are proven:

- mutable tag/range/branch resolution and remote acquisition have issued
  acquisition records, while exact-revision local inputs avoid unnecessary
  network;
- rejected root/dependency bytes, C0 path escape, origin drift, toolchain drift,
  and unauthorized acquisition cause zero affected process starts;
- every selected package tree is admitted before its manifest permit;
- every lock pin maps one-to-one to an admitted source snapshot and an
  authorized same-kind mirror preserving the exact revision and Git tree;
- mirror synthesis and verification run through issued `network=none`
  derivation permits; missing, extra, substituted, or tampered object/tree
  evidence rejects before SwiftPM replay;
- generated-lock evidence resolves transitively to the exact issued
  acquisition, intake, manifest, and mirror receipts;
- portable receipts state only manager-established capabilities, verified mode
  still fails before start without a compatible healthy provider, and receipt
  or cache identities cannot alias across assurance modes;
- `R01-R13`, `P01-P08`, `CGP05`, `CGP11`, `CGN09`, and `CGN15-CGN18` retain
  their required positive and zero-start negative outcomes, alongside focused
  race, full Go, build, vet, lint, canonical-golden, binary-deny, and Kotlin
  exclusion gates.

## Traceability and scope audit

This is not beyond-literal-spec scope. It directly implements:

- `.spec/skill-facing-cli-source-closure.md`: Current delivery scope;
  Source closure invariant 1-4 and 6; Vendored compiled artifact prohibition;
- accepted cross-language contract: C0 tool binding, C1 intake journal,
  C2 manager-owned acquisition broker, C3 artifact admission, C4 mirror
  derivation, pre-C5 records, first binding of source-control objects, and the
  explicit acquisition/execution boundary;
- accepted SwiftPM contract: controlled acquisition, immutable source closure,
  kind-preserving local mirrors, R01-R13, P01-P08, and exact revision replay;
- accepted portable assurance contract from `TASK-260819-3kwd8g` and
  `TASK-260819-1cpbmc`: honest portable capability evidence, explicit verified
  selection, no silent downgrade, and no cross-mode identity reuse.

The entire explicit out-of-scope set was checked. This decision adds no Python
adapter, Kotlin/Dart/.NET work, verified provider implementation,
verified-binary admission, non-SwiftPM native build system, active Swift
plugin/macro execution, arbitrary host library, or generic external-command
admission. No research task is justified because the accepted contracts answer
the boundary question.

## Proportionality audit

Considered splits into separate acquisition-schema, Git-mirror-policy, and
SwiftPM-wiring tasks were rejected. None independently satisfies a user-visible
or security-complete requirement: the broker without mirror admission cannot
replay, mirror authorization without broker evidence has no trusted input, and
adapter wiring without both repeats the reviewed forced fit. Keeping these as
one task-local causal chain is the smallest development-ready board shape.

No diagram was produced. The ordered trust-boundary list is smaller and more
precise than a separate architecture artifact.
