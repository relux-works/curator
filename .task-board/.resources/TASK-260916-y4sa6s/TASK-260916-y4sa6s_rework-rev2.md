# Rework brief — TASK-260916-y4sa6s, revision 2 (E1 signer allowlist + update confirmation)

Revision 1 was rejected with five corrections
(`TASK-260916-y4sa6s_review-verdict-rev1.md`, R1–R5). Two earlier reviewer
attempts of the same revision (cancelled by the machine before they could
record) found two more that the orchestrator adopts as R6–R7. Everything the
verdict table marks "Pass" stays byte-identical unless a correction touches it.

## Corrections (all required)
- **R1 — MCP trigger = complete canonical declaration.** §9.2: a moved `mcp`
  member triggers confirmation when the CCJ-1 bytes of its declaration object
  (as the lock/materialization reads it — `command`, `args`, `env_names`,
  `url` of an `http` declaration, the `environments` selector, everything)
  differ, not only the `command`/`args`/`env_names` triple. Update prose, CLI
  descriptions, `tools/validate.py` (recompute that rule) and its negative
  tests; add URL-only-change and selector-only-change vector cases with
  revision A warning / revision B refusal / flagged acceptance; reconcile
  `mcp-env-names-reorder-silent` with canonical bytes (CCJ-1 preserves array
  order, so a reorder IS a change — retarget or rename that case; no
  set-comparison exception survives).
- **R2 — reinstall takes the flag.** `profile install` on the reinstall path
  accepts `--confirm-system-delta` with identical per-invocation semantics
  (§9.1/§9.2 text, `cli/curator.md` install row, a vector or CLI row pinning
  it); delete the sentence "`profile install` takes no
  `--confirm-system-delta`". Keep: no configuration pre-confirmation, `--all`
  semantics. Revision A's `profile_update_system_delta` diagnostic MUST carry
  the migration hint (names the coming revision-B refusal and
  `--confirm-system-delta`); pin the hint in the conformance expectations
  (`e1_revision_outcomes` or wherever the gate records A/B outcomes) and a
  negative test.
- **R3 — fingerprint grammar exact.** `manager-config-v2` (and the
  `system-config-v2` reference): the GPG fingerprint is exactly 40 uppercase
  hex characters — anchor the pattern so a trailing newline (41 chars) is
  rejected (`minLength`/`maxLength` 40 or `\A…\z`-equivalent for Draft 2020-12:
  use both `pattern` and length constraints); add generated negative schema
  cases for the newline form in both configurations; verify with the real
  schema-validation entry.
- **R4 — revision selection needs commit evidence.** The fixture checker and
  the signature disjunction in `tools/validate.py` MUST refuse tag-only
  evidence for an exact `revision` selection (no tag exists); add the
  negative test/vector (`revision-selection-tag-only-evidence-rejected`);
  keep tag-OR-commit for tag selections.
- **R5 — remove the curator LOGBOOK delta.** In the curator Story worktree
  `<curator control root>/.temp/STORY-260916-ioemse/worktree` run
  `git checkout -- LOGBOOK.md` so the curator repository delta is empty again
  (the runtime's curator Change Request for a spec task is expected to be
  empty). Findings stay in `TASK-260916-y4sa6s_evidence.md` and task notes
  only. Never edit any LOGBOOK.md again.
- **R6 — SSH key identity ignores the comment.** An `ssh` signer entry's
  identity is the key type plus the base64 key material; the trailing comment
  of an OpenSSH public key line is not part of the identity (the same key with
  another comment matches; different material never does). Text (§12.1 entry
  shape), generator/oracle and vectors (same-key-different-comment accepted,
  different-material rejected).
- **R7 — update-confirmation posture.** §12 status text: `env status` reports
  the active update-confirmation revision (`A-warning` or `B-enforcing` — use
  the spelling the rollout section uses) and its behaviour, next to the signer
  posture rows; a conformance case pins the row; no persistent
  pre-confirmation knob is introduced.

## Validation and handoff
`make validate`, generator idempotence (state the exact command; note the
literal `make regenerate-check` is baseline-sensitive on an uncommitted tree —
quote the disposable-copy/`GIT_INDEX_FILE` form the reviewer used, or an
equivalent proof of zero regenerated-file drift). Update
`TASK-260916-y4sa6s_evidence.md` with a "Revision 2" section (per-correction
file:line, transcripts, the honest curator-delta statement), attach
`TASK-260916-y4sa6s_spec-patch_rev2.patch` (full `git diff origin/main` of the
curator-spec worktree with new files via `git add -N`), and hand off with
`task-board handoff TASK-260916-y4sa6s --role doc-writer`.

Worktree and rules unchanged (`TASK-260916-y4sa6s_brief.md`,
`remediation-spec-producer-rules.md`): curator-spec Story worktree
`/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/STORY-260916-ioemse/worktree`
only; no commits, no pushes.
