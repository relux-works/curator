# Managed home leases

This package implements the Curator side of the `home-lease/1` refusal
contract in [curator-spec](https://github.com/relux-works/curator-spec/blob/264244b321c7a2272450fc29bd9952f988d746d7/protocol/home-leases.md).
It is under qualification. Bind reads the public
`curator-run session-manager observe --schema v2` contract through
`SessionManagerObserver` and succeeds only when the requested bound holder,
the authoritative v2 process tuple and Curator's own kernel reading are
byte-equal: `/proc` stat/status plus boot id on Linux, `proc_pidinfo`
`PROC_PIDTBSDINFO` plus `kern.bootsessionuuid` on macOS. Platforms without
a native observer refuse bind as unavailable authority. The first use adds
no release authority, so release of a blocking lease still refuses
`environment_lease_release_unconfirmed/session_state_unknown` until an
authoritative reconciliation integration (terminal session, descendants
and ownership, or proven no-start) is qualified. An acquire therefore
remains protected until that reconciliation exists or an explicit audited
operator break ends it. A caller-supplied holder, a manager assertion, PID
existence, launcher exit or elapsed time cannot substitute for these
proofs; the manager's tuple is an input to the observer, never the proof.

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
project-lease carrier and `project_view_busy` remain unchanged. Native profile
writers also preflight busy project homes under the same held acquisition, using
physical path overlap even when no home-lease ledger exists. Codex selection has
real busy-refusal and post-reconciliation write controls, including preservation
of a disjoint home lease; composition and session evidence are synthetic.
Recovery checks the complete pending inventory before replaying its first write. A failed
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

The exact accepted **198 paired schema vectors** run in
`state.TestAcceptedSchemaCorpus`. Runtime tests exercise real temporary stores,
OS locks, journal recovery and writer entry points; synthetic observer evidence
tests lifecycle decisions without qualifying a production observer. Refusals
check preimages immediately, before their positive controls. The repair/acquire
race uses a second process and a continuous real lock. Collector controls remove
unreferenced runtime and protected cache neighbors while retaining
prepared/unknown home references alongside durable project roots. A real pending
acquisition refuses the full collector and preserves the fixture preimage.

The [qualification map](qualification.md) names semantic pairs, their actual
Go tests and their remaining limits. These tests do **not** constitute
complete execution of all semantic pairs in the pinned corpus.
Consumer vectors HL21–HL24, including launcher `op.json` settlement and zero
manager calls on terminal replay, require the actual consumer implementation.
Writer coverage still needs a complete corpus adapter for all variants of the
18 writer rows, especially remaining legacy disjoint variants and owner-specific recovery.
Borrowed global continuation and a two-environment native MCP batch have paired
real writer controls. Actual standalone/borrowed global addition extends the lock
and projects a Git-backed skill; a failed continuation restores mutable records,
native trees and the declaration before a successful retry. Immutable admitted
entries may remain. Shared-store reconstruction with two consumers, unmanaged
takeover/backup, first provisioning and resolve/use/explicit identity migration
also have real controls with disjoint protection retained. Native Codex selection
protects the old and new profiles, including scoped clear and direct pointer
setters; these fixtures do not cover every environment or physical native alias.
Selected explicit migration includes sibling environments; an actual legacy recovery journal
retains its conservative manager-wide guard. Interrupted apply/rollback fixtures
exercise actual generated credential-migration intent, refuse leased or foreign
recovery writes, and restore prior state before fresh acquisition. These controls do not cover every
ownership combination. Codex and Muse credential-link repairs also have real
refusal, ownership-conflict and relinking controls; Codex seed preservation is
checked, while Claude Keychain and other credential backends remain unqualified.
Install/import activation also refuses before publishing a new profile, including
first activation with native-home physical overlap; real controls verify every
native entry and the recorded choice after protection ends. Rendered-document
batch repair is also paired across the admitted environments and both Claude
forms, with actual document bytes restored after protection ends.
Cache publication and sweep use the real APIs with synthetic artifact bytes, inputs,
portable receipts and independent authority. The production profile compiled
closure described in `docs/design/profile-skill-runtime.md` remains unqualified.
Remote installations whose final name is unresolved,
and several legacy whole-manager operations, currently guard conservatively;
their full disjoint-write qualification remains open. This package must not be
reported as fully conforming or enabled for hosted production sessions until
those gaps and the independent observation contract are resolved.
