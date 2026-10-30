# Managed home leases

This package implements the Curator side of the `home-lease/1` refusal
contract in [curator-spec](https://github.com/relux-works/curator-spec/blob/3f8ffb7f15213872c66ceec28501b9e51d56ac5c/protocol/home-leases.md).
It is under qualification. The command entry point reads the public
`curator-run session-manager observe` contract through `SessionManagerObserver`.
That closed v1 contract lacks required independent proofs, so bind refuses
`environment_lease_authority_refused/authority_unavailable` and release of a
blocking lease refuses
`environment_lease_release_unconfirmed/session_state_unknown`.
An acquire therefore remains protected until independently qualified
reconciliation or an explicit audited operator break becomes possible.

The missing consumer contract must independently bind the request, Werkstatt,
manager registration, session, generation and OS process incarnation, and prove
no-start or completion of descendants and ownership reconciliation. The closed
curator-run observation at
[870e6235](https://github.com/relux-works/curator-agent-launcher/blob/870e62350a008f8b71d368a054140221cbc97925/schemas/session-manager/session-observation.schema.json)
does not supply all those proofs. A caller-supplied holder, PID existence,
launcher exit or elapsed time cannot substitute for them.

## Persistence and authority

The typed namespace is `state/home-leases/v1/`. Its canonical ledger indexes
the original acquire payload and current disposition. `bindings/` holds
immutable initial records; `events/` holds permanent terminal audit records.
All carriers use the accepted closed schemas, strict UTF-8 and duplicate-key
validation and CCJ-1 bytes. Missing ledgers, orphan artifacts, malformed records,
unsafe paths and linked authority refuse admission.

`persistence.Publish` is shared with the existing project-lease publisher. It
borrows one real `managerlock` acquisition and uses the existing transaction
journal for staging, forward publication, compensation and recovery. The
project-lease carrier and `project_view_busy` remain unchanged. Recovery checks
the complete pending inventory before replaying its first write. A failed
publication either completes rollback or retains pending intent for explicit
recovery. Completed rollback removes only unchanged, empty namespace directories
created by that publication; pre-existing admitted directory permissions and
unknown content remain.
A failed first acquisition can therefore be retried after completed rollback
without treating an unexplained missing ledger as authoritative absence.

`state.Guard` compares a command's complete affected target set against every
blocking record, including prepared and unknown holders. The same OS-lock
acquisition must survive until the last write. Physical file identity and
symlink-aware ancestry protect aliases; identifiers are not globally lowercased.
Collection unions retained build/runtime references before sweeping. Nothing
expires on a timer, launches a background reconciler, or breaks automatically.

## Commands

Recognized verbs emit one closed JSON line, including usage and refusals:

```text
curator env lease acquire --env <id> --profile <name> --lock-sha256 <digest> --request-id <id> --holder <json>
curator env lease bind --lease <id> --if-revision <revision> --holder <json>
curator env lease release --lease <id>
curator env lease list [--env <id>] [--profile <name>]
curator env lease break --lease <id> --if-revision <revision> --operator --reason <acknowledgement>
```

`--format json` is optional; other formats refuse. Success exits 0, typed
operational refusal exits 1, usage exits 2. Exact acquire replay returns the
current disposition with `changed:false`. Terminal replay returns
`already_released` or `already_broken` and never admits another session.
Historical successful bind retry cannot override a terminal disposition.
Operator break checks CAS and publishes one permanent audit event with the
tombstone; it makes no claim to have killed any process.

## Validation and remaining qualification

Run the store, schema, journal and writer regressions with:

```sh
bash .github/ci/isolated-test.sh go test ./internal/homelease/... ./internal/envprofile ./internal/transaction ./internal/projectlease ./internal/scopes ./cmd/curator
```

The envprofile and CLI package entry points additionally isolate native defaults
before tests and verify protected operator context/marker/skill identities on
exit. Their bounded negative controls use synthetic original homes only; see
[`testhomeguard`](../testhomeguard/README.md).

The exact accepted **131 paired schema vectors** run in
`state.TestAcceptedSchemaCorpus`. Runtime tests exercise real temporary stores,
OS locks, journal recovery and writer entry points; synthetic observer evidence
tests lifecycle decisions without qualifying a production observer. Refusals
check preimages immediately, before their positive controls. The repair/acquire
race uses a second process and a continuous real lock. Collector controls remove
unreferenced runtime and protected cache neighbors while retaining
prepared/unknown home references alongside durable project roots. A real pending
acquisition refuses the full collector and preserves the fixture preimage.

The [68-pair qualification map](qualification.md) names each semantic pair,
its actual Go tests and its remaining limits. These tests do **not** constitute
complete execution of all 68 semantic pairs.
Consumer vectors HL21–HL24, including launcher `op.json` settlement and zero
manager calls on terminal replay, require the actual consumer implementation.
Writer coverage still needs a complete corpus adapter for all variants of the
18 writer rows, especially global-add compensation and interrupted custom migration recovery.
Borrowed global continuation and a two-environment native MCP batch have paired
real writer controls; these do not cover every ownership combination. Cache
publication and sweep use the real APIs with synthetic artifact bytes, inputs,
portable receipts and independent authority. The production profile compiled
closure described in `docs/design/profile-skill-runtime.md` remains unqualified.
Remote installations whose final name is unresolved,
and several legacy whole-manager operations, currently guard conservatively;
their full disjoint-write qualification remains open. This package must not be
reported as fully conforming or enabled for hosted production sessions until
those gaps and the independent observation contract are resolved.
