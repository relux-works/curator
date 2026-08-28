# EPIC-260810-271m92 Stop-The-Line: Darwin authoritative execution observation

Owner checkpoint: GOAL-260817-ca1367 revision 1, RUN-260817-ffcba9.
Directive: nudge:f29e55, acknowledged at the current goal revision.

## Constraint and accepted foundation evidence

TASK-260811-27xisf remains blocked because reviewer requirement R1 requires authoritative OS-observed attempted and successful process, read, write, network, and output events. The detailed evidence is attached as TASK-260811-27xisf_stop-the-line-endpoint-security.md.

TASK-260818-3vfmjv is accepted done. Its independent verdict TASK-260818-3vfmjv_review-verdict_RUN-260817-6c44de.md proves the portable protected-execution substrate, R2-R6 rework, lossless provider interface, exact publication and cache checks, security-negative coverage, repository gates, and honest fail-closed Darwin behavior. It does not provide the missing entitled Darwin observer.

## Transitively gated active delivery tasks

The unavailable protected-execution boundary gates all remaining adapter delivery:
- TASK-260811-2h4m0s implement-cargo-source-capture-and-vendor-transform
- TASK-260811-3kbf3l implement-rust-offline-build-adapter
- TASK-260811-33ukne implement-swiftpm-source-resolution-and-closure
- TASK-260811-tkurtl implement-swiftpm-c-family-interop-validation
- TASK-260811-2qfnai implement-swiftpm-offline-build-adapter
- TASK-260811-3twayo implement-node-typescript-runtime-and-build-plan
- TASK-260811-1u42b9 implement-npm-source-closure-profile
- TASK-260811-3ksxig implement-pnpm-source-closure-profile
- TASK-260811-twq9ad implement-yarn-classic-source-closure-profile
- TASK-260811-32iojo implement-modern-yarn-source-closure-profile
- TASK-260811-x611eq integrate-cross-language-adapter-conformance

These tasks must not be launched while their common execution and receipt boundary cannot satisfy R1.

## Failed Darwin alternatives

- sandbox-exec can enforce a profile but has no caller-owned lossless audit stream; the attempted report modifier is rejected.
- Unified log inspection is global, asynchronous, potentially lossy, and does not enumerate the complete scoped set of allowed and denied descendant operations.
- fs_usage requires root and cannot serve as an unprivileged Curator boundary.
- Process-tree polling, output-directory scanning, child-emitted manifests, interposition, or copied permit declarations are incomplete or bypassable and would synthesize security evidence.

## Viable choices and tradeoffs

1. Provide a signed privileged Endpoint Security observer for macOS. This is recommended when Darwin support is required because it can supply authoritative process, file, and network events. It requires the Apple-granted com.apple.developer.endpoint-security.client entitlement plus a separately owned signed system extension or daemon, authenticated IPC, loss and fail-closed semantics, deployment, lifecycle, and ownership decisions.
2. Approve Linux-only initial protected execution. Supply the intended kernel enforcement and observation mechanism, required privilege model, and Linux validation environment while Darwin remains fail-closed. This avoids the Apple entitlement boundary but changes the promised initial delivery platform.

Enforcement-only sandbox-exec is not a viable third option because it weakens R1 and the accepted security contract.

## Recommendation and exact decision needed

Recommend choice 1: authorize and provide the ownership contract for a signed, entitled Endpoint Security observer if Darwin delivery is required.

To resume, decide exactly one:
- Darwin path: approve the signed Endpoint Security component, identify its owner, provide or authorize the entitlement and signing path, and approve the authenticated IPC plus event-loss fail-closed contract; or
- Linux path: approve Linux-only initial support, name the kernel enforcement and observation mechanism and privilege model, and provide a Linux validation runner.

No further producer or reviewer routing can honestly satisfy the delivery pool until one of these external platform authorities is supplied.