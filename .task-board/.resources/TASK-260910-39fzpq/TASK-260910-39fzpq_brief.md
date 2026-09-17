# Brief — TASK-260910-39fzpq: protected-boundary contract for the environments root and store (S5)

Story `STORY-260910-148pj1` (profile-store-protected-boundary), wave 3 of the
2026-09 security-audit remediation; spec first, then the manager task
`TASK-260910-32gki6`. `TASK-260916-3l60rn` (E6 path-kind admission) will
reference this contract for `path` overlay directories once it lands.
Worktree: `/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/STORY-260910-148pj1/worktree`
(branch `task-board/story/STORY-260910-148pj1`, forked from curator-spec `main` `684c9f1`).
Rules: `remediation-spec-producer-rules.md` (attached). Role: doc-writer of
normative spec text; no implementation.

## Finding (read it first)
`docs/security-audit-2026-09.md` S5 (Medium): core §9.3 gives the build cache
a normative ownership/permission/containment contract revalidated on every
lookup; environments §4 (profile store), §8.2 (marker), §10.1 (`env resolve`)
give the store, locks and markers none — "link-target identity is sufficient
currency" — so a same-user swap of store bytes is undetected at resolve; the
system prompt and root context are the sharpest surface and the least
protected. Read core §9.3 "Shared protected-cache rules" first: the contract
to mirror.

## Settled decisions (do not reopen)
- **The environments root and every store entry are protected state** in the
  core §9.3 sense: manager-created, manager-protected, resolved independently
  of package input. On every `env resolve` (and again under the manager-home
  mutation lock for mutating operations: install, update, use, sync, repair,
  GC) the manager MUST verify ownership (the operator), private mutation
  permissions or DACL, containment (every entry path resolves below the
  root without leaving it), regular file types and link safety (`lstat`,
  no symlink at the root, the entry root, or any component the manager did
  not create) for the environments root, the profile store root, each lock
  and marker file, and the store entry named by each lock member.
- **Surface hashes are verified, not trusted**: the lock/marker already
  record content hashes for system-prompt and root-context surfaces
  (§5.6 content-hash binding, §8.2); at resolve the manager MUST recompute
  the store entry's tree hash (its pin: `commit` tree bytes for `git`,
  `state_sha256` for `path`) or the recorded surface hashes and compare —
  choose the cheaper rule that still detects a same-user byte swap of a
  system-prompt or root-context file, and state its cost bound (the
  producer decides between full-entry hash per resolve vs per-surface
  hashes, with justification; per-surface hashes of the applied modules
  is the recommended minimum).
- **Outcomes**: a boundary that cannot be proven or a hash that does not
  match makes the profile `environment_store_untrusted` (error): resolve
  fails closed (no fragment), status is non-current, `env status` reports
  the row; a real operation rebuilds the entry from the revalidated
  snapshot into newly established protected state (mirroring
  "rebuilds from the revalidated snapshot" in core §9.3); dry-run reports
  `would-rebuild-untrusted-store`. Rollout is direct (impact row "S5").
- **`env resolve --repair` is not persistence** (E7 note): repair re-applies
  from a store entry only after that entry passed the contract; a
  non-trusted entry is never re-applied.
- Closed diagnostics: `environment_store_untrusted` and the dry-run
  outcome `would-rebuild-untrusted-store` (or the spellings the tables
  favour), spelled identically in §4, §8.4/§8.5, §10.4, §12 posture, vectors
  and CHANGELOG; no knob (the contract is not configurable).

## Deliverable
1. §4 "Profile store": the contract paragraph(s) mirroring core §9.3
   (ownership, permissions, containment, file types, link safety; when
   verified; what an implementation that cannot prove the boundary must do);
   §8.2 marker and §1.3 lock: the same verification applies to the files
   themselves; §10.1 `env resolve`: the verification step and its fail-closed
   outcome, the surface-hash rule; §10.4/§8.5 diagnostics rows; §12 posture
   row; §13 conformance surfaces.
2. Vectors: `env resolve` cases — swapped store entry bytes (system-prompt
   file replaced by same-user write) → untrusted; symlinked entry root →
   untrusted; wrong ownership/permissions → untrusted; intact → resolved;
   dry-run outcome case; registered in the manifest; generator if generated;
   existing vectors byte-identical; manifest/rc.9 pins regenerated.
3. `CHANGELOG.md` Unreleased entry "S5: …".

## Out of scope
Implementation (`TASK-260910-32gki6`), E6 (`TASK-260916-3l60rn`) — do not
define path-overlay admission here; E7 launcher config ownership; the build
cache (core §9.3 unchanged).

## Checklist and handoff
Tick the checklist items you satisfy; attach
`TASK-260910-39fzpq_spec-patch_rev1.patch` (`git diff HEAD` of the worktree
with new files via `git add -N`; NOT `git diff origin/main`, which moves) and
`TASK-260910-39fzpq_evidence.md` (with the `make validate` and regeneration
transcripts), then `task-board handoff TASK-260910-39fzpq --role doc-writer`.
