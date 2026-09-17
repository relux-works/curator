# Review brief — TASK-260910-39fzpq (curator-spec revision, S5 environments root and store protected boundary), review round 2

You are the independent reviewer of a curator-spec normative revision produced
for `TASK-260910-39fzpq` (story `STORY-260910-148pj1`, wave 3 of the 2026-09 security-audit
remediation). Read, in this order: `remediation-spec-producer-rules.md`, the
producer brief `TASK-260910-39fzpq_brief.md`, the producer's evidence
`TASK-260910-39fzpq_evidence.md` and patch `TASK-260910-39fzpq_spec-patch_rev2.patch`
(all task resources), and the finding in
`/Users/administrator/Developer/ReluxWorks/curator/curator-spec/docs/security-audit-2026-09.md`.

## Where the candidate is
The candidate tree is the curator-spec Story worktree
`/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/STORY-260910-148pj1/worktree`
(branch `task-board/story/STORY-260910-148pj1`, forked from curator-spec `main` `684c9f1`).
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
Record exactly one verdict resource `TASK-260910-39fzpq_review-verdict-rev2.md`
(per-item table with quotes and file:line, validation transcript, findings)
and then either
`task-board m 'accept_cr(TASK-260910-39fzpq, revision=2, evidence=TASK-260910-39fzpq_review-verdict-rev2.md)'`
or a changes-requested verdict routed with `set_status(TASK-260910-39fzpq, status=to-dev)`,
listing the concrete corrections. Never accept on the producer's evidence
alone; never edit the candidate.

## Round specifics (S5, revision 2)
Compare the patch against the worktree with `git diff HEAD` (base `684c9f1`),
not `origin/main`. Check in particular: the contract mirrors core §9.3
(ownership, private permissions/DACL, containment, regular file types, link
safety with `lstat` semantics) for the environments root, the store root,
each lock/marker file and each store entry named by the lock; when it is
verified (every `env resolve`, and again under the manager-home mutation
lock for mutating operations); the surface-hash verification rule and its
stated cost bound; the outcomes (`environment_store_untrusted` fail-closed,
no fragment, status non-current, posture row; rebuild from the revalidated
snapshot into newly established protected state; dry-run
`would-rebuild-untrusted-store`); repair never re-applies an untrusted
entry; no knob; closed spellings identical in §4, §8.4/§8.5, §10.1/§10.4,
§12, §13, vectors and CHANGELOG; the vectors cover swapped bytes, symlinked
entry root, wrong ownership/permissions, intact, dry-run; pre-existing
vectors byte-identical; scope excludes E6 path-overlay admission and E7.
Two other spec revisions (E3 `TASK-260916-2rnkei`, E5 `TASK-260916-1qfpu4`)
are in flight on the same base; they are not part of this candidate.

## Round 2 specifics
Revision 1 was rejected with F1–F4 (`TASK-260910-39fzpq_review-verdict-rev1.md`);
the orchestrator settled the design in `TASK-260910-39fzpq_rework-rev2.md`:
store integrity is verified against the PIN (git tree hash of the resolved
commit / `state_sha256`), needing no marker and running before provisioning
and on every resolve (absent vs unreadable/malformed marker distinct); home
currency (marker surface hashes vs the current lock) is a separate stale-home
check (`environment_home_stale`, repaired from the verified store); two
failure classes with an ordering — an unprovable enclosing boundary
(environments root, store root) refuses everything before its first write
with nothing rebuilt, while an individual entry/lock/marker failure inside a
proven boundary is untrusted and rebuilt from the revalidated snapshot
(dry-run `would-rebuild-untrusted-store`); the validator pins scenarios to
object × check with the five-to-one narrowing refused; root cases and the
rebuild case added. Verify each closure with your own probes (replay the
five-to-one narrowing and an intact-updated-store + old-marker case through
`validate.main()`), the stale-vs-untrusted split in the vectors, that the
pin-hash rule detects a same-user byte swap of a system-prompt file with no
marker present, and the standing items (patch = `git diff HEAD` on base
`684c9f1`, byte-identity of untouched vectors, make validate, regeneration
proof, empty curator delta).
