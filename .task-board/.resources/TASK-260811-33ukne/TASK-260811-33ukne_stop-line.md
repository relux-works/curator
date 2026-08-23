# TASK-260811-33ukne Stop-The-Line evidence

Date: 2026-08-23

## Constraint

The latest review requires every SwiftPM Git acquisition and mirror-verification
subprocess to run through the shared `closureexec` pre-C5 executor with complete
issued `DerivationPermit` / `DerivationReceipt` evidence and enforced network,
read, write, process, and environment policy.

The implemented shared executor cannot represent the acquisition half of that
contract:

- `closureexec.DerivationPermit.Validate` accepts only `Network == "none"`.
- Every derivation permit must name a non-empty set of already admitted input
  receipts and exact replay mounts.
- A remote Git origin is necessarily consumed before its returned snapshot can
  be admitted.
- A captured bare mirror cannot be supplied to the executor for verification
  before it is admitted, while the requested admission itself depends on that
  verification or on a receipted deterministic transform.
- `artifactpolicy` has no production issuer that can authorize the opaque Git
  object store as a local derived output, and its ordinary dependency-directory
  admission correctly rejects opaque/compressed Git object bytes rather than
  treating a hand-built inventory as policy authority.

This is an ownership and architecture boundary between the already accepted
shared protected-execution/artifact-policy leaves and the SwiftPM adapter. It
cannot be solved honestly inside `internal/swiftpmsource` by another custom
permit, a `network=none` label around `exec.CommandContext`, unconditional
mirror admission, or mock-only audit evidence. Those are the exact forced fits
rejected by reviews `RUN-260823-5da0b5` and its predecessors.

## Evidence and attempts

- Required lifecycle mutation succeeded: task entered `development` with exit
  code 0.
- Focused baseline `go test -count=1 ./internal/swiftpmsource` exited 0 in
  8.201s before any new product-code change.
- `internal/closureexec/models.go` requires a non-empty admitted-input set and
  rejects every network policy other than `none`.
- `internal/closureexec/executor.go` rechecks those admitted inputs before the
  process-start seam.
- `internal/artifactpolicy/types.go` intentionally exposes no adapter-usable
  `LocalOutputAuthorization` issuer.
- The latest rejected implementation still launches Git directly and invents
  partial Git permit/receipt IDs; retaining or elaborating it would not satisfy
  the accepted contract.

No product code, tests, README content, or existing user changes were modified
in this run.

## Viable options

1. **Recommended: extend the shared contracts at their owning boundary.** Add a
   typed acquisition-broker permit/receipt with exact origin allowlisting and
   lossless network/process/read/write audit, plus a manager-issued deterministic
   Git-mirror transform authorization (or semantic Git object-store admission)
   in `closureexec` / `artifactpolicy`. Then consume those issued records from
   `swiftpmsource`. This preserves remote SCM support and the accepted trust
   model, but materially changes shared code and conformance owned by already
   accepted prerequisite tasks.
2. Restrict `swiftpm-source-v1` to pre-captured/local repositories that can be
   admitted before any Git process. This fits the current offline executor but
   violates the accepted remote source-control acquisition scope and requires a
   formal scope revision.
3. Keep direct Git execution and custom evidence. This is not viable: it is the
   reviewed forced fit and does not establish enforcement or artifact authority.

## Required decision

Authorize option 1 as a cross-boundary extension of the shared
`closureexec`/`artifactpolicy` contracts, or revise the accepted SwiftPM profile
to option 2. Without one of those decisions, the four latest reviewer blockers
cannot all be satisfied truthfully.
