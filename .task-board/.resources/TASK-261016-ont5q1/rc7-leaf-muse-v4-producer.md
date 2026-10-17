# M1 — Emit Muse v4 from verified Curator state

- **Owner repository:** curator.
- **Normative clauses:** CIP-0017 implementation leaf 5; environments §§5.4,
  7.1, 7.4 (Muse), 8.2–8.4.1, 10.1–10.3, 12; v4 schema and
  `conformance/environments-muse-v4/README.md`.
- **Attack surface:** manager-owned verified state becoming launch policy.
- **Size:** M; one producer change, using the store/admission result owned by
  the CIP-0015 leaf rather than independently resolving credentials.

## Acceptance

1. Successful current or repaired Muse resolve emits v4, effective
   `credential_store:{selection:file}`, four ordered XDG parents, and the
   exact `TBH_CREDENTIAL_BACKEND=file` literal. `HOME` is unchanged. JSON,
   env and shell project identical values; JSON alone carries typed policy.
2. Emit `root_context:{exclusive:true}` iff the selected profile's required
   root surface was materialized and verified. Linked, copied, already-current
   and repaired roots qualify. No declared root-context surface means no fact; a declared, verified header-only surface (zero applicable modules) still emits the fact (environments 5.4, 10.2).
   Missing/drifted/unreadable required content or failed repair emits no
   fragment; it cannot become an apparently context-free success.
3. Unset/native Muse selection maps to effective file without rewriting the
   machine selector. Explicit keychain refuses; shared requires a live native
   file captured before managed XDG substitution. Isolated creates no link.
   These are integration assertions against CIP-0015's common policy result.
4. Preserve existing conflict/readability guarantees and seed/config bytes.
   Status distinguishes 1.4.7 loader evidence, 1.4.8 backend evidence, the exact
   replay qualification, workspace residual, and unknown refresh/discovery.

## Mandatory tests and positive controls

| Negative / forbidden effect | Nearby positive control |
|---|---|
| Required missing/drifted/denied root must produce no fragment, including failed repair; never omit fact and proceed | Verified link, copy fallback, current root and successful repair each emit fact |
| No applicable root must not gain policy from an unrelated AGENTS.md or header-like file | A real declared and verified root emits fact |
| Extra env name, non-file constant, wrong suffix/parent, HOME or wrong adapter literal refuses | Exact five-key Muse object validates and renders in declared order |
| Backend present in JSON but absent in env/shell fails | Equality of the five parsed projections including the fifth value |
| Shared fork/retarget/dangling source/nonempty directory refuses without modifying either store; metadata denial stays unreadable | Intact live link; permitted absent-link/empty-directory repair with live native file |
| Stored native unexpectedly rewritten, or explicit keychain silently mapped, fails | Native remains stored native while effective file is reported |

Drive public resolve/status paths, retain the existing Muse liveness tests,
and plant a bounded omission of the fifth renderer value to prove the parity
check detects that defect. Do not read real credentials to assert any row.

## Dependencies and overlap

Depends on the plan shared source/store/isolation policy and containing
config publication when accepting the new machine knob. Environments §8.2
explicitly makes root_context a derived fact, not a new marker member;
nonsecret backend reporting does not widen released markers. Muse alone
does not require a successor marker. No dependency on nonempty Codex
settings, hosted runtime, token refresh, or setup-token delivery.
Release the producer/consumer pair only after M2 qualification and M3/M4
consumer readiness; merged dormant code is not a released capability.

Touch `internal/envfragment/envfragment.go` and schema/render tests;
`internal/envprofile/managed.go`, `muse.go`, `status.go` and their tests;
`internal/envregistry/envregistry.go`; environment docs. Serialize the
shared `managed.go`/registry/status edits with CIP-0015 and snapshot owners.
Do not make a second credential-store resolver in this leaf.
