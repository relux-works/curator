# Verified provider architecture boundary

Verified providers implement one versioned platform-neutral contract but use platform-specific security primitives. The common contract covers provider identity and signature, authenticated IPC, capability negotiation, session challenge, immutable workload permit, atomic start, enforcement and observation stream, event-loss detection, final provider-bound receipt, health, cancellation, upgrades, rollback prevention, revocation, and stable fail-closed diagnostics.

Platform plans:

- macOS: qualify Endpoint Security for process and filesystem authorization or observation and Network Extension for required network flows; package signed and notarized system extensions with required entitlements.
- Linux: qualify a conservative kernel and distribution matrix and compose namespaces, cgroups v2, seccomp, Landlock, LSM or eBPF, fanotify, and network controls only where each required guarantee is provable.
- Windows: qualify a supported build matrix and compose a signed service or drivers with process containment, filesystem or registry filters, WFP, and other supported primitives; telemetry alone never substitutes for enforcement.

Each platform has an independent research gate before implementation. Unsupported configurations reject verified mode instead of degrading. All three providers must pass the same mandatory conformance harness plus platform-specific adversarial, lifecycle, signing, update, rollback, revocation, event-loss, and recovery tests before a joint release.

The providers are separately installed Curator host components. They are never vendored in skills and do not authorize vendored compiled artifacts.