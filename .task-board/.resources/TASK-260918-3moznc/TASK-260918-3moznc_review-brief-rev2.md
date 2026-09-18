# Review brief — TASK-260918-3moznc (curator-spec revision: the absence-vs-read-failure discipline stated once, referenced everywhere, with vectors), review round 2

You are the independent reviewer of a curator-spec normative revision produced
for `TASK-260918-3moznc` (story `STORY-260916-1ll22r`, wave 3 of the 2026-09 security-audit
remediation). Read, in this order: `remediation-spec-producer-rules.md`, the
producer brief `TASK-260918-3moznc_brief.md`, the producer's evidence
`TASK-260918-3moznc_evidence.md` and patch `TASK-260918-3moznc_spec-patch_rev2.patch`
(all task resources), and the finding in
`/Users/administrator/Developer/ReluxWorks/curator/curator-spec/docs/security-audit-2026-09.md`.

## Where the candidate is
The candidate tree is the curator-spec Story worktree
`/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/STORY-260916-1ll22r/worktree`
(branch `task-board/story/STORY-260916-1ll22r`, forked from curator-spec `main` `e8b53a0`).
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
Record exactly one verdict resource `TASK-260918-3moznc_review-verdict-rev2.md`
(per-item table with quotes and file:line, validation transcript, findings)
and then either
`task-board m 'accept_cr(TASK-260918-3moznc, revision=2, evidence=TASK-260918-3moznc_review-verdict-rev2.md)'`
or a changes-requested verdict routed with `set_status(TASK-260918-3moznc, status=to-dev)`,
listing the concrete corrections. Never accept on the producer's evidence
alone; never edit the candidate.

## Round specifics (revision 1)
Compare the patch against the worktree with `git diff HEAD` (base
`5146c7b`; `main` may have moved since — not the comparison base). Verify
in particular: ONE general rule stated once (§8.4 or a referenced
subsection) — absent (`ENOENT`-class on the exact path) versus
present-but-unreadable/malformed (permission, I/O, type, encoding, parse,
schema, non-directory component, symlink where a regular file is required)
are distinct facts, a read failure is never reported/persisted/acted upon
as absence and never triggers an absence-shaped action, the affected row
is non-current with currency unknown, the operation fails closed; a closed
table of unreadable outcomes per file class (marker, recorded surface, lock,
seed, passthrough entry, ledger/backup, inventory candidate) reusing existing
codes where they apply and adding at most one code per uncovered class (the
evidence must show why an existing code did not apply — probe that claim);
every read-site section (§1.3, §7.4, §8.2, §8.3, §9.5, §9.6, §10.1 and any
other the evidence inventories) references the rule instead of restating
it; new codes present in the §9.7 table and every diagnostics list; §13
conformance surfaces; vectors exercising unreadable-but-present for
markers, seeds, locks and passthrough entries (permission denied, malformed
content, symlink-instead-of-file, directory-instead-of-file) pinned by a
rule-7 validator gate — replay name-preserving replacements yourself: an
absence-shaped outcome under an unreadable name MUST be refused; existing
vectors byte-identical; rollout direct; EMPTY curator delta; leave NO files
in the worktree.

## Round 2 specifics
Revision 1 was rejected with three corrections (`TASK-260918-3moznc_review-verdict-rev1.md`,
rework brief `TASK-260918-3moznc_rework-rev2.md`). Verify each: (F1) an
unreadable or malformed lock is `environment_store_untrusted` for that
profile — resolve fails closed with no fragment, status non-current with
currency unknown, and NO mutating operation (install, update, use, sync,
repair, GC) rebuilds, re-materializes or replaces anything from that
evidence; the precedence is stated once and referenced from §4, §8.4.1,
§10.1 and the manager mirrors; the earlier §4 wording that put a lock that
"cannot be read or parsed" into the rebuild-from-snapshot branch is gone;
a pinned vector covers unreadable lock + repair and + update → refused,
nothing rebuilt, nothing written — probe it (rule 7: replace the outcome
with a rebuild-shaped one under the same name → refused). (F2) §9.7 carries
the `environment_passthrough_unreadable` row cross-referencing the owning
§7.7 row. (F3) the backup-record row names its diagnostic — either an
existing code with the stated reason or exactly one new authorised code
(`environment_backup_record_unreadable`), spelled identically in the
table, §8.3, §9.7 and any status mirror, with a vector. (§9.4) the
migration note cites §8.4.1 and names the diagnostic for an unreadable
install record. Compare `_spec-patch_rev1.patch` with `_rev2.patch`: every
hunk difference traces to F1–F3 or the §9.4 note. Standing items as in
round 1: patch = `git diff HEAD` on base `5146c7b`; untouched vectors
byte-identical; `make validate` and `make regenerate-check` (generated
files are STAGED by convention so the check compares against the index —
do not report that as drift); EMPTY curator delta; leave NO files in the
worktree.
