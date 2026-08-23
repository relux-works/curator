# Reviewer verdict: changes requested

Run: RUN-260819-9ed2da
Route: to-dev

## Blocking findings

### R1 — Parsed assurance selection is not wired into CLI execution

internal/config/config.go parses and stores Config.Execution, but repository-wide Go usage finds no production consumer of Config.Execution. NewAssuredExecutor and NewManagerProcessRunner likewise have no production caller outside internal/closureexec. Therefore default CLI execution has not been demonstrated through portable mode, and explicit verified configuration cannot control an actual CLI process path. This misses the acceptance criterion that default CLI execution succeeds in portable mode and verified selection fails closed before any process.

Required rework: connect machine configuration and any supported CLI override to one production assurance-selection boundary, instantiate exactly the portable runner or configured verified provider, and add command/integration tests proving portable default, explicit verified refusal, unknown-mode refusal, and zero process starts on verified preflight failure.

### R2 — The protected cache identity and lookup remain assurance-agnostic

internal/closureexec/store.go:51-56 stores only ExpectedCacheInputID and ExecutionReceiptID. Publish at lines 61-153 and Inspect at lines 157-220 accept only closuregraph.ExpectedCacheInput; neither API receives or validates assurance mode, policy, provider contract, provider identity, binary digest, or capability receipt. DerivedCacheReceipt.ValidateFor is detached from ProtectedStore and is referenced only by its unit test. Consequently a protected entry published under portable mode is addressable by the same expected-cache ID during verified operation, so cross-mode and cross-provider cache reuse is not closed at the actual cache boundary.

Required rework: introduce a typed cache input/key used by publish and inspect that binds the exact mode, policy, execution policy, provider contract/id/binary and fresh capability receipt for verified mode, while preserving the portable manager-worker-v1 identity. Persist the binding in the protected entry and validate it on every hit. Add negative tests that first publish a portable entry and then perform a verified lookup, plus cross-provider and capability-receipt drift lookups, all producing misses or stable fail-closed diagnostics.

### R3 — Verified preflight is not ordered before cache lookup

Provider negotiation exists only in Executor.Commit and Executor.Execute at internal/closureexec/executor.go:116-190. ProtectedStore.Inspect is independent of Executor and can return a hit without provider resolution, binary/trust verification, health, or capability negotiation. The normative assurance contract requires compatible healthy-provider negotiation before cache lookup or execution.

Required rework: move or expose verified preflight at an operation boundary that dominates both cache lookup and process dispatch. Bind the resulting receipt to the cache identity and permit. Prove an unhealthy, incompatible, stale, or identity-drifted provider causes zero cache adoption and zero process starts with the specified stable diagnostic.

### R4 — The concrete portable runner does not yet establish the assigned portable execution boundary

ManagerProcessRunner at internal/closureexec/portable_runner.go:40-75 validates protected inputs and then discards their paths, launches one command with exec.CommandContext, and reports an output root. It does not itself construct the declared immutable replay mounts, bind declared write paths to OutputRoot, derive a deadline from ResourceLimits, bound combined output, or implement complete worker-domain termination/join. There are no ManagerProcessRunner production callers or integration tests. Fake-runner tests prove receipt shaping, not usable portable execution.

Required rework: implement and integrate the manager-owned portable controls that this substrate can establish, with manager-prepared immutable replay, operation-private roots, deadline/output bounds, declared-output placement and post-run identity rechecks. Keep lossless OS observation claims absent. Add real-runner security-negative tests for input mutation, mount/write escape, timeout, undeclared output, nonzero exit, and descendant cleanup as applicable to the portable contract.

### R5 — Malformed provider fields can be silently erased in portable config

internal/config/config.go:481-488 ignores failed string type assertions for provider_version, provider_binary_sha256, and provider_trust_evidence. In portable mode, a non-string value becomes an empty string and can pass as if the field were absent. This is a silent configuration weakening.

Required rework: reject every present field with the wrong type in both modes, reject explicit null where the field is present, validate the closed provider identifier/version/trust shape, and add table tests for malformed values.

## Independent validation

- go test -count=1 ./internal/closureexec ./internal/config — passed
- go test -race -count=1 ./internal/closureexec — passed
- Source inspection confirms the absence of production consumers for Config.Execution, NewAssuredExecutor, and NewManagerProcessRunner.
- Accepted normative assurance protocol and decision were checked against the provider-before-cache, disjoint-cache-identity, no-downgrade, and portable-control requirements.

The green focused tests do not cover R1-R4. This is ordinary implementation rework, not a human-only or external blocker, so the correct verdict branch is changes requested to to-dev.