# Brief — TASK-260910-1b1ens: spec revision for the records/log page boundary (R1/P1)

Story `STORY-260910-25yc0h` (records-boundary-binding), wave 2 of the 2026-09
security-audit remediation. The service story `STORY-260910-3rvvxh`
(curator-skill-registry, `TASK-260910-14dnb7`/`-27yepb`) and the client task
`TASK-260910-2n0233` implement what THIS revision specifies, so the shapes you
define are the contract for both.
Worktree: `/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/STORY-260910-25yc0h/worktree`
(branch `task-board/story/STORY-260910-25yc0h`, forked from curator-spec `main` `23dafa7`).
Rules: `remediation-spec-producer-rules.md` (attached). Role: doc-writer of
normative spec text; no implementation.

## Finding (read it first)
`docs/security-audit-2026-09.md` "R1 / P1. Records pages carry no snapshot
boundary" (High). Records/log responses (`schemas/v1/records-response-v1`,
`log-response-v1`) carry no boundary; §9 only says the first page "captures
one signed snapshot boundary" server-side; §5 rollback state binds snapshot
versions, not page boundaries; nothing requires log replay. A key-holding
registry can serve a fresh `/v1/snapshot` and evaluate `/v1/records` at an
older boundary, hiding a `revoked` record. Text to revise: `protocol/registry.md`
§5 (rollback state), §9 (HTTP service, pagination), the registry-service
profile `profiles/registry-service.md` §2 (stable pagination and cursors),
§5 (snapshots) and §11 (conformance); the client-side rule where §5 already
speaks to clients.

## Settled decisions (do not reopen)
- **Boundary = the signed snapshot object.** Every `/v1/records` and `/v1/log`
  success envelope carries a REQUIRED `boundary` field whose value is the
  registry snapshot (`registry-snapshot-v1`, all fields incl. `sig`) at which
  the page was evaluated. All pages of one cursor chain carry a byte-identical
  boundary (the profile §2 rule becomes visible on the wire).
- **Envelope schemas are versioned, v1 stays frozen**: add
  `records-response-v2.schema.json` and `log-response-v2.schema.json`
  (v1 fields plus the required `boundary`, `additionalProperties: false`),
  register them the way `schemas/v1/README.md` describes for a next-version
  schema, switch the §9 endpoint table to the v2 schemas, and state that a
  client validating the v1 envelope treats the added `boundary` as ignorable
  (the §9 "unknown fields are ignored" rule, made explicit for this field).
- **Client rule (R1)**: before a page contributes records, a conforming client
  verifies the boundary signature under §2, then applies the §5 rollback rules
  to it: a boundary whose `version` is below the persisted high-water for
  that registry URL is rejected (`registry_page_boundary_stale`); an equal
  version must equal the stored `head`, `merkle_root`, `log_size`; a higher
  version advances the high-water exactly as an accepted snapshot does. Every
  page of the chain must carry the same boundary as the first page
  (`registry_page_boundary_mismatch`). A rejected page excludes that registry
  from the resolution like a reachable invalid snapshot (§5).
- **Missing boundary is reported, never silently accepted** (rollout is
  direct, not warn-first — impact row "R1"): a page without a valid `boundary`
  is an invalid response, reported as `registry_page_boundary_missing` naming
  the registry URL; the registry contributes no record for that operation
  (the §4 unreachable-registry wording). No legacy-accept mode, no knob.
- **P1 (cursor/boundary agreement, service side)**: a cursor is bound to the
  boundary of its first page; a service MUST refuse (`404 invalid_cursor`, the
  existing status row) to serve a cursor page at a boundary that differs from
  the cursor's, and MUST NOT re-evaluate a cursor at a newer boundary. Name
  the case in registry-service §2.
- Diagnostics are the closed spellings above unless the existing tables
  favour another form; add them where client diagnostics of this protocol
  live (registry.md tables and, if the manager profile or environments
  status text lists registry client diagnostics, there too). Posture:
  `env status` / `curator status` reports per trusted registry the persisted
  high-water (version, log_size) and whether the last page boundary was
  verified.

## Deliverable
1. `protocol/registry.md`: §9 endpoint table → v2 envelope schemas; a new
   §9.3 "Page boundary" (shape, chain equality, client verification order,
   the three diagnostics, the exclusion rule); §5 extended so the rollback
   state is advanced/checked by page boundaries as well as snapshots; §10 or
   the transparency wording amended so the boundary is the stated inclusion
   evidence and log replay stays optional but now has a stated purpose.
2. `profiles/registry-service.md` §2/§5: the boundary is emitted on every
   page, the cursor is bound to it (P1), the refusal case; §11 conformance
   note.
3. Schemas: `schemas/v1/records-response-v2.schema.json`,
   `log-response-v2.schema.json`; `schemas/v1/README.md` row(s);
   `conformance/v1/schema-cases/{records,log}-response-v2/{valid,invalid}.json`
   registered in `conformance/v1/manifest.json` like the v1 cases.
4. Vectors: extend `conformance/v1/vectors/registry-client.json` with a
   `page_boundary_cases` block (fresh boundary accepted and high-water
   advanced; equal-version-same-body accepted; equal-version-different-body
   rejected; below-high-water rejected as stale; second page with a different
   boundary rejected as mismatch; missing boundary reported/excluded;
   bad signature rejected) and `registry-service.json` `pagination` with the
   boundary-on-every-page and cursor-boundary-disagreement cases; keep the
   existing cases byte-identical. Register nothing new in the manifest unless
   a new file is added.
5. `CHANGELOG.md` Unreleased entry "R1/P1: …" naming the envelope v2, the
   client rule, the missing-boundary reporting and the stories.

## Out of scope
Implementation (service `TASK-260910-14dnb7`/`-27yepb`, client
`TASK-260910-2n0233`), S2 (TOFU/equivocation, `STORY-260910-6bo7ej`), inclusion
proofs beyond the signed snapshot, R2–R8, tags/releases.

## Checklist and handoff
Tick the checklist items you satisfy; attach
`TASK-260910-1b1ens_spec-patch_rev1.patch` (git diff against origin/main with
new files via `git add -N`) and `TASK-260910-1b1ens_evidence.md` (with the
`make validate` transcript), then
`task-board handoff TASK-260910-1b1ens --role doc-writer`.
