# Review brief — TASK-260918-24eazm (curator-spec revision: one closed per-manager, per-platform table for the §9.5 dotfile-manager heuristic, with vectors), review round 2

You are the independent reviewer of a curator-spec normative revision produced
for `TASK-260918-24eazm` (story `STORY-260916-12lbww`, wave 3 of the 2026-09 security-audit
remediation). Read, in this order: `remediation-spec-producer-rules.md`, the
producer brief `TASK-260918-24eazm_brief.md`, the producer's evidence
`TASK-260918-24eazm_evidence.md` and patch `TASK-260918-24eazm_spec-patch_rev2.patch`
(all task resources), and the finding in
`/Users/administrator/Developer/ReluxWorks/curator/curator-spec/docs/security-audit-2026-09.md`.

## Where the candidate is
The candidate tree is the curator-spec Story worktree
`/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/STORY-260916-12lbww/worktree`
(branch `task-board/story/STORY-260916-12lbww`, forked from curator-spec `main` `e8b53a0`).
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
Record exactly one verdict resource `TASK-260918-24eazm_review-verdict-rev2.md`
(per-item table with quotes and file:line, validation transcript, findings)
and then either
`task-board m 'accept_cr(TASK-260918-24eazm, revision=2, evidence=TASK-260918-24eazm_review-verdict-rev2.md)'`
or a changes-requested verdict routed with `set_status(TASK-260918-24eazm, status=to-dev)`,
listing the concrete corrections. Never accept on the producer's evidence
alone; never edit the candidate.

## Round specifics (revision 1)
Compare the patch against the worktree with `git diff HEAD` (base
`5146c7b`; `main` may have moved since — not the comparison base). Verify
in particular: §9.5 carries ONE closed table — one row per manager (at
least chezmoi, home-manager, yadm, stow, dotbot) × macOS / Linux / Windows
state location or an explicit `none` (stow and dotbot keep no state
directory of their own — check that the table says so rather than
inventing one), each cell labelled `verified` or `docs-confidence` with a
cited source (URL + version/date in the evidence) — spot-check at least
three cells against the managers' own documentation yourself (chezmoi's
`%LOCALAPPDATA%\chezmoi` on Windows and `~/.local/share/chezmoi` on POSIX;
yadm's `~/.local/share/yadm` / `$XDG_DATA_HOME`; home-manager's
`~/.config/home-manager` and Nix-only platforms) and flag any cell whose
label overstates its confidence; per-platform path resolution stated
precisely (home directory, `$XDG_DATA_HOME`/`$XDG_CONFIG_HOME` with
defaults where the manager honours them, `%LOCALAPPDATA%`/`%APPDATA%`), the
"present = existing directory" rule and its link discipline consistent with
§4/§8.4; the sentence that a conforming manager's list is exactly this
table in this order and that a table change is a spec revision; the
heuristic still never blocks (`environment_foreign_manager_suspected`,
warning; no new diagnostic); vectors per platform (present → suspected
naming the manager, absent → none, `none` on that platform → inert, XDG
override) pinned by a rule-7 validator gate — replay name-preserving
replacements (e.g. swap the platform or the manager of a "present" case)
and confirm refusal; existing vectors byte-identical; rollout direct;
EMPTY curator delta; leave NO files in the worktree.

## Round 2 specifics
Revision 1 was rejected with R1–R3 (`TASK-260918-24eazm_review-verdict-rev1.md`,
rework brief `TASK-260918-24eazm_rework-rev2.md`). Verify: (R1) with the
generated files STAGED (the campaign convention: `make regenerate-check`
diffs the working tree against the index, so staged generated bytes = a
green check; no Makefile change; the evidence shows the sequence and the
negative proof that a corrupted generated byte fails the check) —
reproduce it in a disposable copy that preserves the index state (copy the
worktree including `.git` pointers, or replay `git add` of the generated
paths), never call a clean-candidate check "drift"; (R2) §9.5 states that
the relative/empty-XDG-value → default rule is Curator's deliberate bounded
heuristic and that upstream managers differ (home-manager `4ac5a2a`
launcher 76–79 and go-xdg `477c957` 56–57 take the first non-empty value
without an absolute-path check; yadm `7eabaee` 1676–1680 rejects
non-absolute values), each `verified` label scoped to the default
location the source establishes, and a pinned home-manager
relative-`XDG_CONFIG_HOME` vector case exists — probe it (rule 7); (R3) the
CHANGELOG/evidence rollout wording says: default path spellings unchanged,
detection semantics changed (XDG relocation honoured, symlinked
directories excluded via lstat), no manager change in this leaf. Compare
`_spec-patch_rev1.patch` with `_rev2.patch`: every hunk difference traces
to R1–R3. Standing items as in round 1 (patch = `git diff HEAD` on base
`5146c7b`; untouched vectors byte-identical; `make validate`; EMPTY curator
delta; leave NO files in the worktree).
