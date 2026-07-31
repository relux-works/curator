# Review cycle 4 verdict: CHANGES REQUESTED

Route: `to-dev`. Reviewer run: `RUN-260729-a1e74e`. No candidate specification,
schema, vector, generator, validator, or product file was edited. Nothing was
staged, committed, published, pinned, or natively qualified.

## Blocking finding R4-1: the observed-host record can alias materially different kernels

`protocol/hardened-execution.md` lines 190 and 197-200 says the TCB identifies
the observed kernel, version, and build, and normatively requires that two
materially different trusted computing bases cannot produce the same
`hardened-tcb-v1` record. The wire contract does not establish that property:

- `schemas/hardened/v1/hardened-common.schema.json` lines 170-184 requires the
  `build` key but permits its value to be `null`.
- `tools/validate_hardened.py` lines 1270-1279 likewise accepts a null build.
- `host.version` and a non-null `host.build` are unconstrained observed strings,
  not a platform-specific immutable kernel/build identity or digest.

Independent probe: start with the valid receipt-v3 schema case, set
`tcb.host.build` to `null`, then run both the shipped receipt schema and
`check_tcb_record`. Result: `schema_errors=0`, semantic check `accepted`.

Two different kernel/backend implementations that expose the same platform,
version string, and absent build value can therefore produce the same TCB
record, cache key, receipt, marker, and claim identity. That contradicts the
profile's complete-TCB MUST and leaves the reusable-output identity guarantee
dependent on an optional descriptive value.

Required rework: define a canonical, platform-specific strong host/kernel build
identity (for example a documented immutable build identifier and/or digest),
make absence fail closed for a claimed complete TCB, and add schema, semantic,
rotation, cache-divergence, receipt, marker, and claim mutants. If the intended
contract is only “same declared host tuple” rather than concrete TCB identity,
the stronger no-alias claim must be narrowed explicitly; that weaker alternative
does not satisfy the current task contract and is not recommended.

## Blocking finding R4-2: full-TCB stability is neither ordered nor reverified through domain exit

`protocol/hardened-execution.md` lines 320-323 requires every trusted component
to remain byte-for-byte unchanged until the last domain member exits. The
authoritative phase list instead puts `identity-reverification` at phase 15 and
`domain-teardown` at phase 16 (lines 959-960), so the re-verification happens
before the operation has proved that every domain member exited.

The manager obligation is narrower still:
`profiles/manager-hardened.md` line 58 re-verifies only the supervisor, worker,
source snapshot, and toolchain. It omits the manager parent, observed host,
backend version/configuration, and every additional trusted component that
phase 5 placed in the hashed TCB. This conflicts with line 48's complete-record
and component-change requirements.

The current executable vector and validator lock this incomplete order in rather
than detecting it. A component can change after phase 15 but before phase 16, or
an omitted TCB member can change without any specified end-of-operation check,
while publication still attributes the artifact to the initial digest.

Required rework: either (recommended) tear down and join the whole domain first,
then re-verify the complete TCB record from canonical pinned identities before
publication, or define an equivalent immutable-handle/snapshot construction
that proves every TCB member cannot change throughout its last use and through
the last domain-member exit. In both cases the executable phase model must
recheck every mutable TCB member and add phase-order plus omitted-member
adversarial mutants.

## Cycle-3 closures independently verified

- Reimplemented `curator-hardened-component-file-v1` and
  `curator-hardened-component-tree-v1` independently in Node: all 10 published
  fixture digests reproduced exactly.
- Independently parsed and compared all 13 `hardened-backend-version-v1` cases:
  all validity, comparability, and satisfaction outcomes matched.
- Host/platform contradiction: schema rejected and semantic validator rejected.
- Observed backend below minimum: semantic validator rejected.
- Trailing-newline backend version: schema and semantic validator rejected.
- Component algorithm/kind mismatch: schema and semantic validator rejected.
- The six guarantees remain individually kernel/hypervisor-enforced with
  not-sufficient lists; package influence exclusions cover executable, paths,
  argv, environment, hooks, network policy, trust roots, controls, and
  publication; all capability, qualification, and identity rejection paths
  precede compiler/package exposure with no fallback.

## Mechanical evidence

Green:

- `make validate`: exit 0; 42 portable schemas, 422 portable vector files,
  6 hardened schemas, 92 hardened suite files, 151 Python tests, and both Go
  tool packages.
- `go test -count=1 ./tools/...`, `go vet ./tools/...`, `gofmt -l tools`, and
  `git diff --check`: all exit 0; formatter output empty.
- Independent scratch-copy `make regenerate` plus
  `make regenerate-hardened`: exit 0 and byte-identical to the candidate.
- `conformance/v1`, `schemas/v1`, and `release/1.0.0-rc.5.json` are byte-identical
  to the accepted predecessor. Portable manifest SHA-256 remains
  `9ba9b8ecf6f06cafda1425aed3539dee8d12af43dabd36f803f4f420dbb1dacf`.
- `release/hardened-1.0.0-rc.1.json` keeps `qualified_platforms` and
  `claims_emitted` empty; every platform remains unqualified and no native
  evidence or guarantee is claimed.

The green gates do not cover R4-1 or R4-2, so they do not change the verdict.
