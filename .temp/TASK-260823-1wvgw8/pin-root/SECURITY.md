# Security policy

## Reporting

Report suspected vulnerabilities privately to `ivan@relux.works`. Include the
affected protocol section or schema, a minimal reproducer, and the security
impact. Do not include secrets, private registry records, or production keys.

Acknowledgement is targeted within three business days. Coordinated disclosure
and release timing depend on severity and whether deployed implementations need
updates before publication.

## Security model

The protocol treats skill repositories, manifests, registry responses, cache
entries, and bundles as untrusted input. Conforming implementations MUST apply
the parsing limits and validation rules in the normative documents before
using paths, signatures, references, or configuration values.

The protocol provides integrity, provenance, revocation, rollback detection,
and deterministic installation. Capability declarations and source auditing
are review and policy surfaces; they are not runtime sandboxes. A successful
audit or registry attestation does not make skill-provided code safe to execute
without the consuming agent's own isolation and authorization controls.

Client snapshot high-water state is security state. It is stored separately
from disposable registry responses, written atomically before acceptance, and
included in protected machine backups. Existing corruption or a failed write
is fail-closed. Loss of all local state still requires an out-of-band
authenticated checkpoint or explicit operator rebootstrap; signatures alone
cannot prove that a newly presented snapshot is not an old valid view.

Registry trust anchors are distributed out of band. Removing a key from the
pinned set revokes trust in signatures made solely by that key. Key rotation
and incident behavior are defined in `protocol/registry.md`. The production
registry threat model, including replay, rollback, equivocation, cursor abuse,
resource exhaustion, credential compromise, crash recovery, and backup
rollback, is normative in `profiles/registry-service.md`.

## Release review

Stable protocol releases require:

1. schema and vector CI on all supported operating systems;
2. both conforming clients passing the same released vectors;
3. review of changes to canonicalization, hashing, signatures, snapshots,
   transparency logs, source identities, and path handling;
4. a signed release tag and immutable release artifacts.

`1.0.0-rc.2` is not promoted to stable until an independent security review
and an independent interoperability review conform to
`reviews/review-report-v2.schema.json`. Each report identifies the reviewed
candidate commit and a public authorship trail. Stable release CI rejects open
critical or high findings and rejects normative changes made after either
review. See `RELEASE.md` for the complete gate.
