# TASK-260728-1jafds reviewer verdict, cycle 1

Verdict: changes requested; route to to-dev.

## Findings

1. Hardened identity binding does not satisfy the task contract. protocol/hardened-execution.md lines 157-165 says profile and TCB identities are absent from receipt bytes and the hashed input. Lines 707-709 exclude the TCB from cache keys and hit comparison, and lines 729-731 forbid it in receipts. The executable vector conformance/hardened/v1/vectors/hardened-identity-separation.json lines 4-10 explicitly excludes both hardened profile identity and the TCB record from hashed identity. Closed receipt schemas v3 and v4 contain only schema_version, cache_key, input, and artifact, and the suite treats TCB in a receipt as invalid. This conflicts with the DoD requiring hardened TCB and profile identities to bind cache, receipt, marker, and claim state and with the AC requiring profile identity in every reusable output. One-to-one prose mapping to execution_policy does not bind a concrete TCB to cache reuse.

2. The fail-closed self-test sequence is not executable under its own process graph. protocol/hardened-execution.md lines 243-253 defines the domain-root worker as the first process inside the build domain. Lines 638-642 require phase 6 self-test from inside that domain, while lines 620-629 and the conformance vector put domain-entry later at phase 7. profiles/manager-hardened.md repeats self-test before worker domain entry at steps 7 and 8. There is no specified actor that can run the in-domain phase-6 test without violating either the graph or the phase order.

3. The two normative exact phase lists conflict. protocol/hardened-execution.md lines 614-636 omits the package-independent toolchain-probe/freeze phase and separates build-permit from go-build. profiles/manager-hardened.md lines 34-102 adds that probe as step 3 and combines permit plus go-build as step 11. The executable vector follows the protocol list only. Implementers cannot satisfy both exact sequences.

## Verified evidence

- PATH=.venv/bin:$PATH make validate: exit 0; 42 portable schemas and 422 portable vector files, 6 hardened schemas and 42 hardened suite files, 64 Python tests, and both Go tool packages green.
- git diff --check and gofmt -l tools: exit 0 with no output.
- Independent no-commit reviewer probe regenerated portable and hardened suites and compared before versus after: exit 0, byte-stable.
- conformance/v1, schemas/v1, and release/1.0.0-rc.5.json are byte-identical to the accepted predecessor; portable manifest SHA-256 remains 9ba9b8ecf6f06cafda1425aed3539dee8d12af43dabd36f803f4f420dbb1dacf.
- Package-influence exclusions cover executable, paths, argv, environment, hooks, network policy, and trust roots; all platform declarations remain unqualified and claims_emitted is empty.

## Required rework

- Define and mechanically validate one identity-binding model that satisfies the accepted cache, receipt, marker, and claim requirements for both hardened profile and TCB identity. Revise schemas and negative vectors accordingly; do not preserve the current exclusion merely to retain the reserved example key if that contradicts the task contract.
- Define an executable pre-package process state machine: identify the actor that enters the build domain, performs the guarantee self-test, and then permits package bytes, or reorder and rename the phases so domain entry precedes the in-domain self-test while compiler/package exposure still remains fail-closed.
- Make one normative ordered phase list authoritative, have the manager profile reference it, regenerate vectors, and add validator/unit mutants that reject the current impossible ordering and phase-list drift.
- Re-run full validation and portable byte-parity evidence, then request another independent review.