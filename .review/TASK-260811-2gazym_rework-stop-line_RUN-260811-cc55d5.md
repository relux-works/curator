# TASK-260811-2gazym rework stop-line

Status: resolved by authoritative option-1 directive

Run: `RUN-260811-cc55d5`

Date: 2026-08-11

## Constraint

Accepted vector F14 and the accepted `artifact-manifest-v1` identity contract
cannot both be satisfied for two physically different archive serializations.

The accepted taxonomy requires the manifest to contain the immutable origin,
lock/checksum record, and exact raw payload SHA-256 and size, and requires the
final manifest digest to cover the canonical manifest (accepted taxonomy lines
485-505). It also says an allow receipt is valid only for the exact raw payload
(lines 507-510). F14 separately requires archive-order permutations of the
same logical members to have identical canonical manifest bytes and digest
(line 660).

Changing ZIP member order changes the raw ZIP byte sequence and therefore its
SHA-256. Because the raw digest, origin checksum, root-node digest, and final
manifest digest are all canonical manifest fields, two such captures cannot
have identical full manifest bytes or digest without dropping or lying about
the exact raw identity.

The reusable corpus demonstrates the contradiction with two deterministic
physical permutations:

- `F14/z-first` manifest digest:
  `sha256:e129bafec8dd039904f0040eecad89a8960956fcb12426491a76e742edd7f2ed`
- `F14/a-first` manifest digest:
  `sha256:ce03dd3148be9e1f06a78ded1571527863465b71d4cde5d40fd9505c9fcd2cc2`

The test proves their decision, primary diagnostic, classifications, virtual
paths, observations, accumulated accounting, and findings are identical after
removing only the exact capture-bound fields. It also deliberately proves that
the distinct physical bytes do not alias one full manifest identity
(`internal/artifactpolicy/conformance_test.go`, `assertF14CaptureAndOrderIdentity`).

## Rework state before the stop-line

The four reviewer findings from `RUN-260811-151b59` were otherwise repaired:

1. R1: caller-populated trust evidence was replaced by sealed, package-owned
   authorization capabilities bound to exact payload, path, selected root,
   fingerprint/time-of-use state, action/input identity, output expectation,
   and protected publication state. Forgery, replay, copied/hard-linked input,
   and drift tests reject before execution/publication.
2. R2: ZIP refuses the declared count before member accumulation; tar, ar, and
   live-directory walkers charge entries during enumeration, including invalid
   and duplicate entries; synthetic directories are charged once; diagnostic
   retention is bounded while a digest binds the full finding set. End-to-end
   100,001-entry and duplicate-flood tests are present.
3. R3: the manifest records accumulated traversal accounting and complete role
   facts. The codec rejects role/class/kind/decision mismatches, incomplete
   admitted nodes, invalid ancestor/final decisions, diagnostic-role drift,
   origin rewrites, accounting rewrites, and forged self-consistent rehashes.
4. R4 except the contradictory full-manifest-equality clause: the reusable
   `internal/artifactpolicy/conformance` corpus publishes 47 A01-A08,
   C01-C12/C01a-C01f, F01-F14, T01-T05, and V01 cases with stable path, class,
   node/final decision, primary code, authorization result, and exact manifest
   digest. Compound branches have public-service coverage, including all
   reviewer-named missing forms and early F08 limit refusals.

Kotlin remains excluded. Compiled dependency content remains fail-closed, and
`verified-binary-v1` remains unavailable.

## Failed assumptions and rejected workaround

The attempted secure interpretation compares an exact logical manifest
projection while preserving distinct full capture identities. It cannot be
claimed as literal F14 compliance because the accepted row explicitly says the
full canonical manifest bytes/digest are the same. Making the full values equal
would be a forced fit that weakens immutable raw-origin binding, so no such
normalization, exclusion flag, alternate digest priority, or test-only hook was
added.

## Viable options

1. **Preserve exact raw binding and amend F14 (recommended).** Require identical
   decision, primary diagnostic, canonical logical-node/evidence projection,
   and deterministic exact manifest for each physical payload; require distinct
   full manifest identities when raw bytes differ. This preserves the accepted
   source-closure trust invariant and current cache identity.
2. **Add dual identities in a new schema/policy version.** Keep an exact capture
   manifest/digest and add a separately domain-separated logical-content digest
   for order-independent comparison. This is explicit and safe but expands the
   schema and downstream checkpoint/cache contract.
3. **Remove raw/origin fields from full identity.** This makes literal F14
   possible but allows distinct immutable package payloads to alias and
   contradicts the governing security decision. Not recommended.

## Exact decision needed

Choose option 1 (amend/clarify F14) or option 2 (authorize a dual-identity
schema change). Without that architecture decision, the developer cannot both
check “Implementation matches AC” and preserve the exact raw-payload trust
boundary.

## Resolution

The orchestrator selected option 1 in directive
`RUN-260811-cc55d5:nudge:600053` on 2026-08-11. Exact raw/origin binding remains
mandatory. F14 now compares the canonical logical-node/evidence projection and
requires distinct full `artifact-manifest-v1` bytes and digests whenever the
physical archive bytes differ. The authoritative narrow amendment is recorded
in `.spec/TASK-260811-2gazym_f14-archive-order-amendment.md`.

The task returned to `development`; this stop-line is retained as historical
evidence rather than deleted.

## Validation state

Current-code validation completed before the stop-line:

- `go test -short -count=10 ./internal/artifactpolicy` — exit 0.
- Targeted 100,001-entry live-directory test — exit 0.

Earlier rework checkpoints were green for focused race/coverage, focused vet,
and pinned focused lint, but the final native-archive precedence correction was
made afterward. Per the stop-line order, repository-wide tests/build and final
current-code race/lint/vet gates were not claimed or checked off. They remain
mandatory after the architecture decision and before review handoff.

Development-loop red gates were reported at their real status and repaired:
the first full live-directory run exposed diagnostic precedence (exit 1), the
first repeated suite exposed premature native-archive class diagnostics (exit
1), and the first lint run exposed six lint findings (exit 1). Their focused
reruns exited 0 before this stop-line.
