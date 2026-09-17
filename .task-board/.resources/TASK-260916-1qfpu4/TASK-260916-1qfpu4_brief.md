# Brief — TASK-260916-1qfpu4: managed-surface writes never follow a symlink (E5)

Story `STORY-260916-73a5zg` (managed-write-nofollow-rule), wave 3 of the
2026-09 security-audit remediation; spec first, then the manager task
`TASK-260916-19shmj`.
Worktree: `/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/STORY-260916-73a5zg/worktree`
(branch `task-board/story/STORY-260916-73a5zg`, forked from curator-spec `main` `684c9f1`).
Rules: `remediation-spec-producer-rules.md` (attached). Role: doc-writer of
normative spec text; no implementation.

## Finding (read it first)
`docs/security-audit-2026-09.md` E5 (Medium) and Appendix B (E5: mitigated in
code — remove-then-create on every managed-surface write in
`internal/envprofile/switch.go` — but "not atomic, not `O_NOFOLLOW`", and the
rule is absent from the text). Environments §9.5 detects a foreign-manager
symlink (`environment_foreign_manager_detected`) and lets the operator take
over with backup, but nothing forbids writing THROUGH an existing link at a
managed-surface path; a conforming implementation written from the text
reproduces the defect (takeover wrote through a dotfile-manager symlink into
the foreign source of truth). Text to revise: §8.3 (ledger discipline and
backups — the write discipline), §9.5 (onboarding/takeover), §8.4 (drift)
where it reads through links, §5.x/§7.5 where materialization writes surfaces,
§13 conformance surfaces; vectors: the family that holds materialization /
takeover cases (`environments.json` or `manager-lifecycle.json` — find it).

## Settled decisions (do not reopen)
- **One normative rule, stated once and referenced everywhere a managed
  surface is written**: a materialization, takeover, repair or backup write
  to a managed-surface path replaces the directory entry — create the new
  regular file or link under an operation-private name in the same
  directory, then rename over the target (atomic replace) — and MUST NOT
  follow a symbolic link at the target path or at any path component below
  the managed root that the manager itself did not create in this operation;
  `O_NOFOLLOW`-class semantics on open, `lstat`-class semantics on inspection.
  A pre-existing symlink at the target is either the foreign-manager stop of
  §9.5 (link pointing outside the managed root) or a manager-owned link to
  be replaced as an entry; in neither case is the link's target opened for
  writing. The rule covers copies, links and generated files alike
  (`copied`/`linked`/`managed-home` modes, backups of §8.3, marker and
  ledger writes).
- Rollout is direct (impact row "E5": under the hood — a tampered or
  symlinked target now yields a refusal/replacement, never a write through).
- **Diagnostics**: reuse the existing closed set where the situation is
  already named (`environment_foreign_manager_detected`,
  `environment_surface_unmanaged_conflict`); add at most one new code for a
  write that would have to follow a link the manager does not own and that
  no existing stop covers (working name `environment_write_would_follow_link`,
  error) — only if the producer shows an uncovered case; otherwise no new
  code. Closed sets stay closed.
- **Conformance vector**: a materialization/takeover case whose target path
  is a symlink to a file outside the managed root: expected = the link is
  replaced by the managed entry (or the operation stops with the §9.5
  diagnostic when takeover is not authorized), and the link's former target
  file is byte-identical afterwards ("target untouched"); plus a case with a
  symlinked parent directory component below the managed root (refused, not
  followed); plus a repair case (`env resolve --repair`) with a link planted
  after provisioning.

## Deliverable
1. §8.3 (or a new §8.3.1 "Write discipline") carrying the rule with RFC 2119
   keywords; §9.5 takeover text referencing it (takeover = backup, then the
   §8.3 replace; never a write through the link); §8.4/§10.1 repair text
   referencing it; §5/§7.5 pointers where surfaces are written.
2. §13 conformance surfaces naming the new cases; vectors added to the right
   family with the "target untouched" expectation expressed in the case
   shape; manifest registration; generator if generated; existing vectors
   byte-identical; manifest/rc.9 pins regenerated.
3. `CHANGELOG.md` Unreleased entry "E5: …".

## Out of scope
Implementation (`TASK-260916-19shmj`), E7 launcher config ownership
(`STORY-260916-33vuzm`), S5 store boundary (`TASK-260910-39fzpq`, running in
parallel — do not define ownership/permission validation here; reference
"the section 4 boundary contract" generically if you must).

## Checklist and handoff
Tick the checklist items you satisfy; attach
`TASK-260916-1qfpu4_spec-patch_rev1.patch` (`git diff HEAD` of the worktree
with new files via `git add -N`; NOT `git diff origin/main`, which moves) and
`TASK-260916-1qfpu4_evidence.md` (with the `make validate` and regeneration
transcripts), then `task-board handoff TASK-260916-1qfpu4 --role doc-writer`.
