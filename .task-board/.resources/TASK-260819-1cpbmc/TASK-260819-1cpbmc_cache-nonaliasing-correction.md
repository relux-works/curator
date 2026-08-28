# Required correction: production cache and dispatch non-aliasing

Decision owner: goal orchestrator
Date: 2026-08-19
Applies to: `TASK-260819-1cpbmc`

This correction is mandatory for acceptance and refines the approved additive build-session decision.

1. `AssuredBuildCacheInput.ID` must derive a new typed cache identity for portable mode. Returning the historical `buildmeta.Input.CacheKey()` unchanged is forbidden because an assurance-blind legacy entry could satisfy the new portable lookup.
2. The typed identity must bind the logical build input plus the complete canonical `AssuranceBinding`: mode, policy, execution contract, provider identity where applicable, capability receipt where applicable, and the exact established capability record.
3. Both local `buildcache` and repository `buildrepo` paths must use explicit portable or verified assurance. Portable must never be represented by a missing or `nil` assurance field.
4. The same opaque `BuildAuthority` selected before lookup must gate the actual local and repository compiler session. Repository adapters must not discard `BuildSessionReceipt`; cache publication and hit adoption must verify it against the exact authority, logical input, toolchain, and artifact.
5. Add negative production tests proving that legacy assurance-blind entries, cross-mode entries, cross-provider entries, and capability-drifted entries cannot be adopted and that failed verified preflight causes zero cache adoption and zero process starts.

Green compatibility tests are insufficient if any of these properties is absent.
