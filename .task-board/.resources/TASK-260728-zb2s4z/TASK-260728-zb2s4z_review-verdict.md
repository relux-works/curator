# TASK-260728-zb2s4z independent review verdict

## Verdict

**CHANGES REQUESTED** — route to `to-dev`.

The candidate's provenance, generated identities, compatibility guards, and
executable gates are sound, but the normative portable execution contract is
internally inconsistent and its native-control/capability-evidence vocabulary is
not closed enough for interoperable implementation. These are specification
defects within the existing authorized architecture, so they are ordinary
rework rather than a stop-the-line decision.

## Findings

### R1 — Portable deferral conflicts with unconditional enforcement text

`protocol/core.md` lines 302-305 and `profiles/manager.md` lines 184-189 still
require that only the fingerprinted Go/GOROOT executable graph may start and
that source is read-only to every child. `SECURITY.md` lines 78-80 likewise says
a manager must reject when it cannot enforce source/context separation or the
process graph.

The new portable policy then says the opposite at the enforcement boundary:
`protocol/core.md` lines 370-377 and `SECURITY.md` lines 115-124 explicitly
defer kernel-enforced read-only source/toolchain and exact executable
allowlisting, and require that their absence not reject a portable build.
`SECURITY.md` line 134 repeats the unconditional "Source is read-only to
children" statement immediately before lines 136-140 say unavailable native
filesystem controls are merely reported.

This leaves two incompatible conforming-manager interpretations:

1. reject macOS/Windows when the absolute read-only/exact-process-graph
   requirements cannot be enforced, recreating the strict profile that this
   task was authorized to replace; or
2. permit the build without those guarantees, violating the unchanged absolute
   `MUST`/`may start` language.

The same ambiguity reaches the canonical policy: `common.schema.json` lines
339-340 still serializes `"network": "none"` while the portable policy
explicitly does not promise total network denial. The specification needs to
say unambiguously that this field denotes fixed Go module/VCS/network
configuration only, not a kernel/network-containment claim.

Required rework:

- reconcile the core, manager, SECURITY, decision, and authoring text so the
  fixed graph and source/toolchain integrity rules describe the exact portable
  mechanism actually required (for example fixed manager selection plus
  pre/post identity/currentness checks), without implying kernel-enforced
  read-only presentation or executable allowlisting;
- make the failure boundary explicit: missing mandatory portable controls
  reject before the worker, while absence of each of the six deferred hardened
  capabilities never does; and
- define the semantic meaning of policy `"network": "none"` so it cannot be
  mistaken for the deferred total-network-denial guarantee.

### R2 — Native-control and capability-evidence semantics are open-ended

`protocol/core.md` lines 365-368 requires "every native ... control the host
actually provides" and uses an illustrative `including` list. The manager
profile repeats "every available native control" at lines 210-213.
`SECURITY.md` lines 136-139 broadens this again to every filesystem, network,
time, memory, disk, process-count, and output control. The normative documents
do not close this set, define availability probes, or define the exact evidence
record.

By contrast, the executable vector fixes exactly five native-control names, and
the informative author guide says only that its table is what the "reference
matrix expects." A new OS primitive or a different interpretation of
"provides" would therefore change conformance without changing the
`manager-worker-v1` policy identity. Two managers can also emit different
capability evidence while both claiming conformance, because the record shape
and vocabulary are not normative.

This misses the task requirement to define exact portable mandatory controls
and capability evidence and makes `build_execution_control_unavailable` versus
non-rejecting unavailable evidence under-specified.

Required rework:

- make the rc.5 native-control inventory exhaustive and normative per platform,
  or explicitly make a named versioned vector section the exhaustive authority;
- define the closed evidence fields and allowed states for each control,
  including availability, applied/unavailable status, probe timing, and the
  contradiction/error rules;
- state where this result-only evidence is exposed and why it is excluded from
  cache/receipt/marker/claim identity while the execution-policy identity is
  bound into all four; and
- add negative guards for an unknown control, missing required control entry,
  contradictory availability/applied state, and any wording/metadata path that
  turns a deferred hardened capability into a portable rejection.

## Verified passing evidence

- Both predecessor and assigned worktrees remain detached at
  `57c1f56846d221ecc55786bd3c2467ec32f11730`, with zero commits after the pin
  and no staged changes.
- Protected manifest schemas 1-7 and the promised frozen receipt/marker/claim
  schema surfaces compare byte-for-byte with the accepted predecessor.
- External acquisition, external lifecycle, pack/index, and claim-v3
  qualification artifacts compare byte-for-byte.
- Independent CCJ-1 recomputation matched:
  - portable cache key
    `sha256:529370122ae11e2e961d5265b1a020e046bcd43165b2eb96b05e73a51187ac9b`;
  - reserved hardened key
    `sha256:13736230d33ce59de7f7323dcd4cffd510655ad8dabd5ee9e8b6cb182ec70037`;
  - pre-revision key
    `sha256:3fcd714a40e8918eb67dbd35d435875dcce6c9047da811a1fa26626e5e57be48`;
  - receipt-v2 cache key
    `sha256:07dd911a7edc29b906a021aa6e1449632ce91c2e5a3eb0ea4f851cb84fe5c492`;
  - receipt-v2 hash
    `sha256:11d2bf4df52638ef353b3286c426261eac2a73b0b64a32f85d78c04490072cea`.
- Independent candidate manifest SHA-256 is
  `sha256:bfe49f254332cb6f38d47b015679b4e6a4ec46eb13207dfa06e94171651b9124`;
  release metadata SHA-256 is
  `ead54508d5435a6e05f96e8fc399d5acfebab9c813a1e704faf4c8acec309766`.
- Direct gates passed:
  - `python -B tools/validate.py`: 42 schemas and 422 vector files;
  - Python unit tests: 22 passed;
  - `go test ./tools/...`;
  - `go vet ./tools/...`;
  - gofmt, `git diff --check`, unstaged-index, and zero-commit checks.
- In the byte-identical disposable clean probe:
  - two consecutive `make regenerate-check` runs passed;
  - `make release-check VERSION=1.0.0-rc.5` passed;
  - the probe remained clean and byte-identical to the assigned candidate.
- Candidate claim-v3 release tuples remain empty; no platform claim, release
  pin, commit, tag, or publication was fabricated.

## Reviewer boundary

No candidate, predecessor, schema, vector, tool, release, or product-code file
was modified by this review.
