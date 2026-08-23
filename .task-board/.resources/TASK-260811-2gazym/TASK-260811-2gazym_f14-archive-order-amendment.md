# F14 archive-order amendment for artifact-manifest-v1

Status: authoritative implementation clarification for `TASK-260811-2gazym`

Decision source: orchestrator directive `RUN-260811-cc55d5:nudge:600053`

Date: 2026-08-11

## Amendment

This amendment supersedes only the F14 expected-result sentence in the
accepted compiled-artifact taxonomy. It does not modify any other artifact
class, diagnostic, trust role, detector, limit, or conformance requirement.

For archive-order permutations containing the same normalized logical members,
F14 requires all of the following to be identical:

- final decision and primary diagnostic;
- canonical logical node paths and collision keys;
- node kinds, classes, decisions, rules, observations, container chains, and
  declared uses;
- deterministic diagnostics and complete-finding evidence;
- accumulated entry, container, emitted-byte, and expansion accounting; and
- the canonical logical-node/evidence projection after removing only fields
  that bind the exact physical payload.

The complete `artifact-manifest-v1` bytes and manifest digest remain bound to
the immutable origin and exact raw payload. They therefore **must differ** when
archive member order changes the physical payload bytes. Exact repetition of
the same physical payload must reproduce identical complete manifest bytes and
digest.

The capture-bound fields excluded only from the F14 logical comparison are:

- final manifest digest;
- raw payload SHA-256;
- immutable-origin checksum SHA-256 and its duplicated role-evidence value;
  and
- the root payload node SHA-256.

No field is excluded from normal manifest encoding, admission identity, cache
identity, or receipt validation. This clarification preserves the governing
requirement that an allow receipt is valid only for the exact raw payload and
prevents differently serialized package captures from aliasing.

## Required conformance proof

The shared F14 corpus contains deterministic `z-first` and `a-first` ZIP byte
fixtures. It must prove:

1. each physical payload has its own pinned full manifest digest;
2. the two full manifest byte sequences and digests differ;
3. repeated admission of identical bytes reproduces identical full manifest
   bytes and digest; and
4. the canonical logical-node/evidence projections are byte-identical.

This is option 1 from the stop-line evidence. No dual-identity schema is added,
and exact raw/origin binding is not weakened.
