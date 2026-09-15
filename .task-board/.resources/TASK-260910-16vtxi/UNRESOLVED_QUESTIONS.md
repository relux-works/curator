# Open decisions after the visual discussion

Task: TASK-260910-16vtxi. Date: 2026-09-10.

Resolved: binary distribution means CLI packages required by skills. The
Curator manager installer is not the requested feature.

1. Project lock: final filename, local override/effective-state layout and
   refresh UX; exact pins and frozen collection membership are sufficient for
   an initial contract without a general semver solver.
2. Context activation: initial portable modes; adapter-specific scoped rules;
   explicit unsupported/fallback behavior. Keep semantic kind separate from
   root/system class and ordering.
3. Project instruction ownership: public tracked output versus generated
   ignored output, existing-file adoption and drift; explicit sync first or a
   proven per-launch project channel with concurrency isolation.
4. MCP: initial import formats and project/native writing scope; name-only auth
   bindings and unsupported-client behavior. Do not infer config discovery from
   a runtime MCP URL.
5. CLI package declaration: separate reusable provider package versus fields in
   agent-skill; candidate descriptor shape and command exports. Root/operator
   policy chooses acquisition preference and source fallback, not untrusted
   downloaded metadata.
6. Native packaging: initial target/ABI/minimum-OS matrix and accepted archive
   formats. Validate macOS standalone notarized CLI delivery online/offline;
   stapled container extraction into the protected store is not proven yet.
7. Trust ownership: accepted producer/builder identities, artifact-specific
   registry subject, key rotation/freshness/offline policy; reuse current
   Ed25519 infrastructure versus introducing full TUF/Sigstore profiles.
8. Runtime revocation: install/resolve admission only, or a new managed-exec
   gate? Specify cached/offline and direct-execution limits either way.
9. Source toolchain conformance: manager spec requires tested Go 1.23 support;
   checked implementation allowlists only Go 1.25. Reconcile separately from
   the new prebuilt route; do not silently change either contract here.
10. Local package input boundary: adopt physical overlap checks, define
    deterministic generated-output exclusion for root packages, and add a
    content identity that invalidates runtime/build state when only those
    inputs change. See `local-source-runtime-boundaries.md`.
11. Repository transport independence: exact logical identity syntax,
    per-machine endpoint/auth policy, safe translation of user Git/SSH
    settings, strict failure classes and acquisition-lane compatibility.
    Tracked separately as TASK-260910-3du5nd; see its resource
    `transport-neutral-repository-resolution.md`.

Recommendation: proceed first with P1 and typed-context contracts. Keep the
instruction writer and prebuilt native packaging as scoped decisions/spikes;
they do not block local skill collection installation.
