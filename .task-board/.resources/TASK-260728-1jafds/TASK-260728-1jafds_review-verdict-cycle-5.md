# Review cycle 5 verdict: ACCEPTED

Route: `done`. Reviewer run is not goal-bound (`TASK_BOARD_RUN_ID` is unset).
No candidate specification, schema, vector, generator, validator, or product
file was edited. Nothing was staged, committed, published, pinned, or natively
qualified. This reviewer verdict supplies no `commit_ack`.

## Cycle-4 findings

### R4-1 — accepted

The nullable descriptive host build value has been replaced by the required
closed `curator-hardened-host-build-v1` record
`{algorithm, identifier, content_sha256}`. The construction is
domain-separated and length-framed over the observed kernel identity, bounded
release, platform build identifier, and the ordered closed set of declared
build-identity sources. Missing, empty, unreadable, ambiguous, bare-string, or
unreproducible identities fail with `hardened_tcb_identity_invalid` before
domain establishment and before package exposure.

Independent evidence:

- the prior `tcb.host.build = null` reviewer probe is rejected by both the
  shipped receipt schema and `check_tcb_record`;
- bare strings, missing members, unknown algorithms, extra members,
  cross-platform identifiers, malformed releases, and trailing newlines reject;
- an independent Node implementation reproduced all 7 published host-build
  fixture digests exactly;
- two macOS kernel instances with the same platform, release, and build
  identifier but different `kern.version` values produce different host-build
  digests, TCB digests, and cache keys;
- the same rotation is rejected against the base receipt, marker, and claim,
  while `package_visible_input_changed` remains false;
- every conforming observed-host record in the suite traces to a published
  recomputable fixture.

### R4-2 — accepted

The authoritative 17-phase model now orders
`domain-teardown` (15), `identity-reverification` (16), and `publication` (17).
Teardown destroys and joins the complete domain before re-verification. The
manager then re-observes canonical pinned identities, recomputes the complete
`hardened-tcb-v1` record, and requires byte-identical record and digest equality.
Subset checks, restated phase-5 records, and digest-against-itself comparisons
are explicitly forbidden.

Independent evidence:

- protocol, manager profile, schema phase enum, executable vector, and ordering
  invariants agree on teardown -> re-verification -> publication;
- the executable member set covers all 9 mutable TCB members plus the frozen
  source snapshot;
- each of the 10 members has an omission vector, with separate phase-order,
  restated-record, and changed-member cases;
- every re-verification failure case publishes nothing and writes no cache or
  marker state.

## Acceptance-criteria audit

- The executable profile contains exactly the six guarantees and the exhaustive
  11-class capability inventory. Every guarantee is kernel/hypervisor enforced,
  has explicit not-sufficient mechanisms, is not claimable under portable, and
  remains unestablished in this unqualified specification revision.
- Platform qualification, capability probing, toolchain freeze, and complete
  TCB verification precede domain entry. The in-domain self-test precedes
  `go-list` and any package exposure. No partial profile or portable fallback is
  permitted.
- The closed package-influence exclusion set prevents package control of
  executables, tool paths, argv, environment, working directory, hooks,
  network/trust policy, views, roots, allowlists, controls, session permits,
  resource bounds, evidence, reusable state, and publication.
- Execution-policy, hardened-profile, and concrete TCB identities bind cache,
  receipt, marker, and claim state. Capability evidence remains result-only.
- All Linux, macOS, and Windows declarations remain `unqualified`;
  `qualified_platforms` and `claims_emitted` remain empty, and adversarial
  evidence remains `pending-native-validation`.
- `conformance/v1`, `schemas/v1`, and `release/1.0.0-rc.5.json` are byte-identical
  to the accepted predecessor. Portable manifest SHA-256 remains
  `9ba9b8ecf6f06cafda1425aed3539dee8d12af43dabd36f803f4f420dbb1dacf`.
- Hardened JSON duplicate-key scan: 0 duplicate keys.

## Validation evidence

All commands ran in
`/Users/iv/Developer/ReluxWorks/.temp/TASK-260728-1jafds/curator-spec-worktree`.

- `PATH="$PWD/.venv/bin:$PATH" make validate` -> exit 0:
  42 portable schemas, 422 portable vector files, 6 hardened schemas,
  108 hardened suite files, 189 Python tests, and both Go tool packages.
- `go vet ./tools/...`, empty `gofmt -l tools`, and `git diff --check` -> exit 0.
- `probe-r4.py` -> exit 0.
- `validator-mutation-probe-r4.py` -> exit 0, 13/13 semantic adversarial
  instances rejected with 13 distinct messages and positive controls passing.
- `mutation-probe-r4.py` -> exit 0, 21/21 schema/vector/document rule mutations
  detected with 21 distinct messages.
- `construction-probe-r4.py` -> exit 0, 7/7 construction and phase-order
  mutations detected.
- Independent `diff -qr` against the accepted predecessor for
  `conformance/v1` and `schemas/v1`, plus `cmp` for
  `release/1.0.0-rc.5.json` -> exit 0.

A bare `make validate` invocation selected the system Python and failed because
that interpreter lacks `jsonschema`. The task worktree's recorded validation
environment is its `.venv`; putting `.venv/bin` on `PATH` produced the complete
green run above. This is an invocation-environment note, not a candidate defect.

## Verdict

ACCEPTED. The cycle-4 rework closes both prior blocking findings, preserves the
previously accepted hardened contracts and frozen rc.5 bytes, fits the
additive/non-gating architecture, and satisfies the task acceptance criteria.
