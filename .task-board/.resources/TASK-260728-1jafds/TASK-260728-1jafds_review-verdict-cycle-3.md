# Review cycle 3 verdict: CHANGES REQUESTED

Route: to-dev. Reviewer run: RUN-260728-6a2bde. No candidate product, schema, vector, generator, or validator file was edited.

## Blocking finding R3-1: trusted-component digest algorithms are names without normative constructions

protocol/hardened-execution.md lines 191-205 requires two materially different trusted computing bases to produce different hardened-tcb-v1 records and calls each trusted component a closed cryptographic record. It names curator-hardened-component-file-v1 and curator-hardened-component-tree-v1, but the candidate never defines either algorithm. The repository-wide references only enumerate or use the names; there is no domain-separated byte construction, file byte rule, tree traversal and ordering, relative-path encoding, file-type and mode treatment, symbolic-link or hard-link rule, or fail-closed handling for unreadable or changing components.

Impact: content_sha256 is not independently reproducible under one normative contract. In particular, different implementations may hash different projections of an installed package tree while claiming the same algorithm identity, so the new trusted_components member does not yet establish the concrete-TCB guarantee required before native implementation.

Coverage also overstates completeness. tools/generate-hardened/main.go lines 1181-1273 rotates a trusted-component digest and adds one component, but never rotates a valid component kind, name, or digest algorithm. tools/test_validate_hardened.py lines 678-687 treats any rotation of the top-level trusted_components array as coverage for all of those mutable subfields. This does not satisfy the cycle-3 directive to mutate component kind and digest.

Required rework: normatively define both digest algorithms, including domain separation and the full file/tree canonical byte stream and link/type/error semantics; add independently recomputable fixtures and expected digests; add valid rotations for component kind, name, algorithm, content, path/tree membership, type, and link substitution, with cache divergence and receipt/marker/claim binding checks.

## Blocking finding R3-2: claim qualification does not enforce its own minimum-version or observed-host relations

protocol/hardened-execution.md lines 1026-1028 says observed tcb.backend.version MUST be at or above the claim entry minimum_version and that an unsatisfied qualification is invalid. Claim schema 4 lines 35-46 only requires both as non-empty strings. tools/validate_hardened.py lines 1455-1488 checks operating-system membership, backend equality, and required configuration, but never compares backend.version with minimum_version.

Independent adversarial probe against the shipped schema and semantic validator:
- backend-below-minimum: set tcb.backend.version to 0 and minimum_version to 999999; schema errors = 0; check_claim_qualification accepted.
- host-platform-contradiction: keep platform linux but set tcb.host.identity to windows and version to 1.0; schema errors = 0; check_claim_qualification accepted.

The first instance directly contradicts the normative MUST. The second shows that the newly bound observed-host identity is not related to the platform/backend qualification it is supposed to identify. These gaps are dormant only because every platform is currently unqualified; the task defines the contract future native implementations and claims will use, so accepting optimistic claims now is not safe.

Required rework: define an unambiguous, backend-specific version identity and comparison rule; enforce minimum-version satisfaction in the conformance validator with below/equal/above and malformed-version mutants; relate the canonical observed host identity to platform, or explicitly narrow and justify the host record contract; add negative claim vectors for both relations.

## Rechecked cycle-1 and cycle-2 properties

The earlier profile/TCB build-input binding and phase-order fixes remain present. The 17-phase protocol list is authoritative, manager-hardened is phase-keyed, domain-entry precedes every in-domain actor and self-test, and package exposure begins only after self-test acceptance. Package-influence exclusions remain closed. The TCB record now includes parent, supervisor, worker, host, backend, toolchain, and closed component records, and platform/backend plus target/platform schema relations reject the cycle-2 probes.

All platforms remain unqualified; release qualified_platforms and claims_emitted are empty; every platform declaration has native_evidence absent; adversarial escape evidence remains pending-native-validation.

## Mechanical evidence

GREEN:
- PATH=.venv/bin:$PATH make validate: exit 0; 42 portable schemas, 422 portable vector files, 6 hardened schemas, 79 hardened suite files, 113 Python tests, and both Go tool packages.
- go test -count=1 ./tools/..., go vet ./tools/..., gofmt -l tools, and git diff --check: all exit 0.
- Exact diff -r/cmp against the accepted rc.5 predecessor: conformance/v1, schemas/v1, and release/1.0.0-rc.5.json unchanged. Portable manifest SHA-256 remains 9ba9b8ecf6f06cafda1425aed3539dee8d12af43dabd36f803f4f420dbb1dacf.
- Independent clean probe /tmp/TASK-260728-1jafds-review.5C7a8z: make validate, make regenerate-check, make regenerate-hardened-check, and release_gate.py --version 1.0.0-rc.5 --commit HEAD all exit 0.

No native guarantee is accepted, and nothing was staged, committed, or published.