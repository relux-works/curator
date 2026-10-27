# Managed home leases

This package implements the Curator side of the `home-lease/1` refusal
contract in [curator-spec](https://github.com/relux-works/curator-spec/blob/3f8ffb7f15213872c66ceec28501b9e51d56ac5c/protocol/home-leases.md).
It is under qualification. Production session observation is not connected:
the command entry point passes no `Observer`, so bind refuses
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
terminal publication retains pending intent until explicit recovery.

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
env -u GIT_DIR -u GIT_WORK_TREE -u SSH_AUTH_SOCK -u GOFLAGS go test ./internal/homelease/... ./internal/envprofile ./internal/transaction ./internal/projectlease ./internal/scopes ./cmd/curator
```

The exact accepted **131 paired schema vectors** run in
`state.TestAcceptedSchemaCorpus`. Runtime tests exercise real temporary stores,
OS locks, journal recovery and writer entry points; synthetic observer evidence
tests lifecycle decisions without qualifying a production observer. Refusals
check preimages immediately, before their positive controls. The repair/acquire
race uses a second process and a continuous real lock. Collector controls remove
an unreferenced runtime neighbor while retaining prepared/unknown references.

These tests do **not** constitute complete execution of all 68 semantic pairs.
Consumer vectors HL21–HL24, including launcher `op.json` settlement and zero
manager calls on terminal replay, require the actual consumer implementation.
Writer coverage still needs a complete corpus adapter for all variants of the
18 writer rows, especially borrowed legacy/global continuation, native MCP
ownership combinations and interrupted custom migration recovery. Build-key
marking is tested; compiled-artifact publication and sweep qualification belong
to the build-cache suite. Remote installations whose final name is unresolved,
and several legacy whole-manager operations, currently guard conservatively;
their full disjoint-write qualification remains open. This package must not be
reported as fully conforming or enabled for hosted production sessions until
those gaps and the independent observation contract are resolved.
