# Review brief — TASK-260910-1tvf2t (curator-spec revision, S2 bootstrap checkpoint and equivocation residual), review round 2

You are the independent reviewer of a curator-spec normative revision produced
for `TASK-260910-1tvf2t` (story `STORY-260910-6bo7ej`, wave 3 of the 2026-09 security-audit
remediation). Read, in this order: `remediation-spec-producer-rules.md`, the
producer brief `TASK-260910-1tvf2t_brief.md`, the producer's evidence
`TASK-260910-1tvf2t_evidence.md` and patch `TASK-260910-1tvf2t_spec-patch_rev2.patch`
(all task resources), and the finding in
`/Users/administrator/Developer/ReluxWorks/curator/curator-spec/docs/security-audit-2026-09.md`.

## Where the candidate is
The candidate tree is the curator-spec Story worktree
`/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/STORY-260910-6bo7ej/worktree`
(branch `task-board/story/STORY-260910-6bo7ej`, forked from curator-spec `main` `4a2fa3e`).
The Change Request revision the runtime published for this task carries an
EMPTY repository delta in the curator repository — that is expected: the spec
lives in curator-spec and the reviewed artifact is the worktree plus the
attached patch. Do not edit the worktree; read only.

## What to verify (all read-only)
1. **Patch = worktree.** `git -C <worktree> add -N . && git -C <worktree> diff origin/main`
   (or `git diff main` — same base) equals the attached patch resource
   (`git patch-id --stable` on both). Any difference is a finding.
2. **Brief conformance, item by item**: every deliverable line of the producer
   brief (sections named, rule content, settled decisions honoured, diagnostics
   in the tables, §12.1/§12.2 rows, vectors + manifest registration, CHANGELOG
   entry, warn-first two-step where marked user-visible, `env status` posture
   row). Quote the exact text for each.
3. **Closed sets stay closed**: every new diagnostic / knob / lock key is
   spelled identically in prose, tables, schema, vectors and CLI rows; no
   open-ended wording ("and similar", "etc."); no frozen v1 surface changed
   beyond what the brief authorises.
4. **Validation, independently**: from the worktree run
   `PATH="/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/venv/bin:$PATH" make validate`
   (`set -o pipefail`, quote the three gate outputs and exit code). If the
   repository has `make regenerate-check`, run it too. A red gate is a rework
   finding whatever the producer's evidence says.
5. **Vectors actually exercise the rule**: open each new vector; confirm the
   positive and negative cases correspond to the rule's branches (e.g. approved
   / unapproved / changed; install dir / listed dir / PATH-only / published
   dir; direct / transitive-drop / transitive-error / waived; both rollout
   profiles). A vector that only restates a default is a finding.
6. **Consistency with the settled operator decisions** in the brief; a
   deviation is a finding unless the evidence names a real spec gap and
   leaves the decision intact.
7. **Scope discipline**: no implementation code, no unrelated edits, no
   proposal 0014–0018 content; `git -C <worktree> status --short` lists only
   spec, schema, vector, manifest and CHANGELOG files.

## Verdict
Record exactly one verdict resource `TASK-260910-1tvf2t_review-verdict-rev2.md`
(per-item table with quotes and file:line, validation transcript, findings)
and then either
`task-board m 'accept_cr(TASK-260910-1tvf2t, revision=<N>, evidence=TASK-260910-1tvf2t_review-verdict-rev2.md)'`
or a changes-requested verdict routed with `set_status(TASK-260910-1tvf2t, status=to-dev)`,
listing the concrete corrections. Never accept on the producer's evidence
alone; never edit the candidate.

## Round specifics (S2, revision 2)
Compare the patch against the worktree with `git diff HEAD` (base `4a2fa3e`),
not `origin/main` (now `1ca4b3d` with E6). Check in particular: the bootstrap
object is exactly a signed `registry-snapshot-v1` supplied out of band
(reusing the R3/P2 object, not respecified) in ONE closed configuration form
next to `public_keys`; on first use it is verified against the pinned keys
and persisted as the initial high-water BEFORE any network response is
accepted, and the first snapshot/page boundary must satisfy §5 against it;
without it TOFU stays but is reported once (`registry_bootstrap_tofu`) and
the per-registry posture says bootstrapped-from-checkpoint vs first use;
rebootstrap never lowers state (`registry_checkpoint_regression`); the
equivocation residual is stated precisely and the OPTIONAL cross-registry
`merkle_root` comparison at the same `log_size` is specified with a closed
mirror relation and `registry_view_divergence` (warning under advisory,
error under strict) without changing resolution and without a quorum;
rollout direct; the registry entry extension follows the frozen-schema
rule (v1 byte-identical; v2 definition extends it); `bootstrap_cases` cover
every branch the brief lists with a scenario-pinning validator gate (rule 7 —
probe with a replacement); pre-existing `rollback_state_cases` and
`snapshot_transitions` byte-identical; no restatement of the S1/S3
hardened profile (in flight on the same base); EMPTY curator delta.

## Round 2 specifics
Revision 1 was rejected with F1–F3 (`TASK-260910-1tvf2t_review-verdict-rev1.md`)
plus a CHANGELOG sentence. Verify each correction against the rev-2 rework
brief `TASK-260910-1tvf2t_rework-rev2.md`:
- F1: replay the reviewer's five removal probes (the rev-1 probe is a task
  outcome) through `validate.main()` — all five MUST be refused — and run
  your own exhaustive same-name substitution/removal sweep over
  `bootstrap_cases`: every discriminating input (`signature_valid`,
  `candidate_same_body`, `same_log_size`, `roots_equal`, …) must be
  required present with its exact type, explicit `False` where false is the
  discriminator; missing/null/wrong-type negatives and self-consistent
  replacement negatives exist (incl. a main-entry negative); the 15 scenario
  names and the valid vectors are unchanged.
- F2: manager §10 states the status outcome PER ROW (successful checkpoint
  bootstrap = current/informational; TOFU = warning; checkpoint regression =
  error, non-current; divergence = warning under advisory / non-current
  error under strict) with the refusal's status mapping pinned in a vector
  case and NO new diagnostic.
- F3: registry §5, §8 and the diagnostic table agree on first fixation: a
  cached snapshot fixes first-use high-water only if it was itself accepted
  online under §5 (persisted state), a never-validated cache entry never
  fixes state, and the permitted path's posture is pinned in a vector.
- CHANGELOG sentence acknowledges the TOFU warning.
Compare `_spec-patch_rev1.patch` with `_rev2.patch`: every hunk difference
must trace to F1–F3 or the CHANGELOG fix. Standing items: patch = `git diff
HEAD` on base `4a2fa3e` (`main` is `1ca4b3d`, not the comparison base);
byte-identity of untouched vectors; `make validate`; regeneration proof;
EMPTY curator delta; leave NO files in the worktree.
