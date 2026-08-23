# Architecture decision: additive assured build sessions

Decision owner: goal orchestrator
Decision date: 2026-08-19
Source blocker: `TASK-260819-1cpbmc_architecture-blocker_RUN-260819-6078f4.md`

## Decision

Proceed with Option A as an additive, platform-neutral build-session contract.
Do not change or weaken the existing `EnforceObserveProvider` / `host-execution-provider-v1` one-shot derivation contract. Add a distinct versioned build-session capability that composes with the existing authenticated `godriver` manager/worker protocol.

This decision follows the user's already-recorded architecture direction: portable execution is the default manager-owned path; verified execution is supplied by separately installed platform extensions; macOS, Linux, and Windows providers remain a separate future epic; provider contracts stay platform-neutral; vendored compiled binaries remain denied.

## Required production shape

1. The CLI assurance-selection boundary returns one opaque operation/session authority rather than only checking configuration and discarding the binding.
2. The same authority is mandatory for every production cache lookup/adoption, process dispatch, receipt validation, and cache publication in both local and repository build paths.
3. Portable mode adapts the existing Curator-owned `godriver` authenticated manager/worker session. It records only controls and observations the manager actually establishes; it must not claim lossless host observation.
4. Verified mode requires an installed extension implementing the new build-session capability plus the exact configured provider identity, trust evidence, health, and capability receipt. Without it, selection fails before cache lookup and before any child process. No portable fallback is allowed.
5. Cache and receipt identities are disjoint by assurance mode, policy, execution contract, provider identity, and negotiated capability evidence. Cross-mode, cross-provider, stale, or capability-drifted entries cannot be adopted.
6. The trusted Go toolchain is a separately declared, fingerprinted manager input, not a source replay mount. Do not copy or vendor `GOROOT`. Recheck its exact identity immediately before and after the build session and bind that identity and the build-session receipt into publication evidence.
7. Keep the existing one-shot pre-C5 `DerivationPermit` closed. Add build-specific permit/receipt types rather than widening its invocation kinds or read-root semantics.
8. Add production integration tests that run a real default-portable build and prove its receipt/cache binding, then inject verified provider success and health/identity/capability drift to prove exact dispatch and zero cache adoption/process starts on failure.

## Specification/release handling

The published `curator-spec v1.0.0-rc.8` and immutable rc.7 history must not be rewritten. Implement the additive contract with explicit versioned schemas and synchronize the normative source and compatibility mapping. If the new public provider contract changes the published normative surface, prepare a separately reviewed successor release; never mutate the rc.8 tag or assets.

## Rejected alternatives

- Do not accept cache-only binding while leaving compiler dispatch outside the assurance authority.
- Do not force the interactive build through the one-shot derivation permit.
- Do not add an extra helper layer that copies or ambiently reads the full toolchain without typed toolchain evidence.

This is an implementation authorization within the active goal, not a request to implement any macOS/Linux/Windows verified provider in the current epic.
