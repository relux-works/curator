# Reviewer verdict for TASK-260811-i3154q

Verdict: **changes requested -> to-dev**

## Goal and reviewed evidence

- Reviewer run: `RUN-260811-552b3a`
- Authoritative goal at the pre-verdict checkpoint: `GOAL-260811-c7033b`
  revision 1
- Resolved scope: `TASK-260811-i3154q`
- Review policy: `required`
- Directives: `nudge:a481f8`, acknowledged, kept review package-local while
  the sibling producer owned the repository-wide validation lane;
  `nudge:03d9dd` later released that lane and instructed this reviewer to use
  producer full-gate evidence unless a fresh repository-wide run was necessary
- Accepted contract:
  `TASK-260810-1uu9lk_accepted-cross-language-closure-graph-and-checkpoints.md`,
  SHA-256
  `874c3a40b9ff9fcf130f37f22e9f5aa2bdd6d9f246403e059c98e84736e27bbc`
- Latest producer evidence:
  `TASK-260811-i3154q_rework-evidence_RUN-260811-1cb000.md`, SHA-256
  `2ea1c04d87a885dd3a5a85d0625d0724ad7b4a9c65a4b9cc9ad8608baa60216e`
- Prior verdict:
  `TASK-260811-i3154q_review-verdict_RUN-260811-774fc7.md`, SHA-256
  `2556189a4167dd60995f34d409cb1a3b13be482581623b25927e303d16c6c8ec`
- Prior unchanged probes:
  `TASK-260811-i3154q_reviewer-probes_RUN-260811-774fc7.go`, SHA-256
  `6bddd1efc80da2b7cba4fdd799b4c6bf1f9714bca717e6ec662b82faba11cb52`
- New reviewer probes:
  `TASK-260811-i3154q_reviewer-probes_RUN-260811-552b3a.go`, SHA-256
  `7dc525a532027ac0908d3bbab01310955bff7d6f8af5a5709c7fe8c335c920cc`

This artifact records only the changes-requested branch. The defects are
ordinary, autonomous implementation rework; there is no external or
human-only Stop-The-Line boundary.

## Prior findings

The latest rework closes both findings from `RUN-260811-774fc7`:

1. The original unchanged undeclared-platform-role probe now rejects the
   invalid extra binding.
2. The original unchanged provenance-only duplicate-edge probe now rejects the
   duplicate semantic edge. `semanticEdgeKey` excludes `EvidenceOrigin`, and
   the permanent suite covers all six origin-bearing payload families and
   permutation-stable diagnostics.

The prior overlay therefore passes unchanged. Acceptance is withheld because
the independent adversarial pass found the two defects below.

## Required changes

### 1. Host fallback aliases a target-only declaration and creates two canonical identities

The accepted contract requires each `targets` edge to carry the exact role and
requires one explicit edge for each slot that declares that role. A
role-mismatched platform reference must fail canonically (accepted contract
lines 223-227 and edge definition line 278).

The implementation instead normalizes both the edge role and the node's
declared role before checking membership and exact cardinality
(`internal/closuregraph/validation.go:801-850`). With a target-only selection,
raw `host` therefore becomes `target` even when the source node declares only
`target`.

The new probe starts from exact CGP10, changes only the product's concrete
`targets` record from raw role `target` to raw role `host`, and retains the same
sole target-platform endpoint. `ProjectActive` accepts it. The original and
aliased bindings describe the same effective host-to-target platform mapping
but have distinct IDs:

- original:
  `sha256:7f404718cb92e903b650594515e373cdaf7643b4908225cc7671cf262c2c1578`
- host alias:
  `sha256:c36fa03925b89796a2cf88edf9df40a62a0d9b055aa14ec9350ced3d6f4b4b05`

This is both a role-validation failure and a noncanonical identity alias.
Host-to-target platform-ID fallback remains valid for a node that intrinsically
declares a `host` slot; it must not rewrite a raw `host` edge into a declared
`target` slot.

Required rework:

- validate the raw `TargetsPayload.BindingRole` against the source node's raw
  closed declared-role set;
- use host-to-target fallback only to resolve the destination platform ID when
  a genuinely declared host slot has no distinct host platform;
- count exact slot bindings without allowing raw host and target encodings to
  alias one declaration; and
- permanently cover a target-only product with a raw host edge, the resulting
  identity-alias case, and the positive host-declared action fallback.

### 2. Closed payload pointer forms pass record validation and panic downstream

`NodePayload`, `EdgePayload`, and `CheckpointPayload` are exported sealed
interfaces. Their concrete methods have value receivers, so pointers to every
closed payload type also satisfy the interfaces. The record validators accept
those pointer forms (`internal/closuregraph/node.go:271-285`,
`internal/closuregraph/edge.go:269-289`, and
`internal/closuregraph/checkpoint.go:207-248`), while downstream validation
uses unchecked assertions to value types.

Three independent probes replace one already-valid fixture payload with a
non-nil pointer to the same value. Construction/record validation accepts each
form, then validation panics instead of returning a canonical diagnostic:

- action node: `interface conversion: closuregraph.NodePayload is
  *closuregraph.ActionPayload, not closuregraph.ActionPayload` at the active
  action-slot projection (`internal/closuregraph/validation.go:640-646`);
- `targets` edge: `interface conversion: closuregraph.EdgePayload is
  *closuregraph.TargetsPayload, not closuregraph.TargetsPayload` during
  platform validation (`internal/closuregraph/validation.go:470-484`); and
- C0 checkpoint: `interface conversion: closuregraph.CheckpointPayload is
  *closuregraph.C0ProfilePayload, not closuregraph.C0ProfilePayload` during
  checkpoint-chain validation (`internal/closuregraph/checkpoint.go:376-385`).

A closed codec/model boundary must reject an unsupported dynamic
representation deterministically or support it consistently; it must not admit
it and later panic. Typed-nil pointers have the same interface-shape hazard and
must also fail canonically.

Required rework:

- enforce an exact supported dynamic payload representation in
  `Node.Validate`, `Edge.Validate`, and `Checkpoint.Validate`, or remove every
  unsafe value assertion by consistently supporting pointer representations;
- return a stable canonical schema/reference/checkpoint error for unsupported
  and typed-nil payload forms before graph or chain projection; and
- add table-driven non-nil-pointer and typed-nil regression coverage across all
  ten node payloads, eleven edge payloads, and C0-C7/C3a/C3b checkpoint
  payloads, proving no panic and permutation-independent diagnostics where
  applicable.

## Independent verification

Positive package-local gates passed against the reviewed source:

| Gate | Result |
| --- | --- |
| Original `RUN-260811-774fc7` reviewer overlay | pass |
| `go test -count=1 ./internal/closuregraph` | pass, 12.197s |
| Exact CGP05/CGP10, SCC, cycle, permutation, checkpoint, and Go selectors | pass |
| Accepted Ruby verifier against the authoritative contract | pass, 53 records and all references |
| Accepted Ruby verifier against the implementation corpus | pass, 53 records and all references |
| `go test -race -count=1 ./internal/closuregraph` | pass, 109.900s |
| `go test -shuffle=on -count=10 ./internal/closuregraph` | pass, 100.107s |
| `go test -count=1 -cover ./internal/closuregraph` | pass, 82.1% |
| `go vet ./internal/closuregraph` | pass |
| `go build ./internal/closuregraph` | pass |
| `gofmt -l internal/closuregraph` | pass, no files listed |
| pinned `golangci-lint` v2.12.2 on `./internal/closuregraph/...` | pass, `0 issues.` |
| canonical implementation corpus SHA-256 | `fed9657b33297650e64178d605c7ee7f47dde640fefd12a99dec3651bb9cadcb` |
| sorted 27-file package manifest SHA-256 | `7491158c1521b20abfe19464c306c1603b4ca26c90e748462afdb348cb7b0c88` |

The new reviewer overlay fails exactly four tests: the platform-role identity
alias and the node, edge, and checkpoint pointer panics. This is the expected
red evidence for the requested rework.

No repository-wide gate was launched by this reviewer. Directive
`nudge:a481f8` initially reserved that lane for sibling producer
`RUN-260811-799b42`; final directive `nudge:03d9dd` released it after the
sibling full gate completed source-stably and explicitly permitted reliance on
producer evidence unless a fresh run was necessary. The deterministic failing
review probes already decide this branch, so another expensive full run was not
necessary. The latest graph producer evidence records a green full suite at Go-source
fingerprint
`152935e2a15928239815c36851b597fb37c4d284cb878900ab17777b4bc72423`;
that producer result is acknowledged but not restated as an independent
reviewer run.

No product code was modified by this reviewer. The task must return to
development and receive another independent reviewer cycle after these
regressions are fixed.
