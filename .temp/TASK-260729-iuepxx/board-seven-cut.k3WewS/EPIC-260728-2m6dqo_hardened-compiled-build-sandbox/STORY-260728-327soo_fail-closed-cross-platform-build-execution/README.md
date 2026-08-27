# STORY-260728-327soo: fail-closed-cross-platform-build-execution

## Description
Specify, implement and independently verify a hardened execution profile for untrusted compiled-skill source. The profile is additive to the portable manager-worker profile and may use reviewed native workers, signed helpers, containers or VMs where required by the host.

## Scope
Normative threat model and capability contract; Linux, macOS and Windows host adapters; Curator and csk integration; adversarial filesystem, network, executable, descendant and resource tests; cache, receipt and claim separation. No dependency edge may point from the portable compiled-build delivery to this follow-up story.

## Acceptance Criteria
For every positively claimed host profile, tests prove total network denial, read-only source and toolchain, private-build-root-only writes, hard aggregate descendant process/memory/disk/time/output bounds, exact executable allowlisting and fail-closed preflight. Unsupported platform/version/configuration combinations reject before compiler execution with stable diagnostics. Independent review confirms no portable output can be mistaken for hardened output.
