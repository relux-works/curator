# Rework brief — TASK-260910-1wjst3, revision 2 (answers review verdict rev1)

Read `TASK-260910-1wjst3_review-verdict-rev1.md` (task outcome) first; it is
the authority for this round. Everything the reviewer marked "Pass" stays as
it is. Work in the same curator-spec Story worktree
(`/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/STORY-260910-2awkzu/worktree`,
your rev1 edits are still there, uncommitted); rules in
`remediation-spec-producer-rules.md` (note: attach the spec diff as
`TASK-260910-1wjst3_spec-patch_rev2.patch`, not under the change-request name).

## R1 — CLI guidance must carry both rollout revisions (medium)
`cli/curator.md` (the shell-hook paragraph, ≈ line 100) states the enforcing
behaviour only. Qualify it by the two labelled revisions exactly as
`profiles/manager.md` §8.5 does: Revision A (`A-warning`) keeps sourcing
unapproved/changed files and prints the warning with the migration hint
(`curator hook approve <path>`); Revision B (`B-enforcing`) skips them with the
warning. Same spellings as the profile.

## R2 — emitted-hook conformance surface must be reproducible (high)
The 12-case matrix in `conformance/v1/vectors/shell-hook-trust.json` is
declarative only: a consumer cannot reproduce it, and the reviewer showed that
flipping `sourced` to `true` in `changed-env-sh-B-enforcing-not-sourced` still
passes `tools/validate.py`. Make the vector a real input/expected-output
contract:
- fixture bytes for each candidate env file (`.agents/env.sh` and
  `.agents/env.ps1`), with their sha256 computed over exactly those bytes;
- the manager-state approval record present / absent / mismatched (record the
  exact `{ path, sha256, approved_by, approved_at }` used), and one case where
  a **project-supplied** forged approval record exists and MUST NOT authorize
  the bytes;
- the selected rollout profile (`A-warning` / `B-enforcing`) and the activation
  sequence (first activation, second activation in the same session) so the
  once-per-session warning count is observable;
- expected outcome per case: sourced yes/no, warning emitted yes/no with the
  path and the approval command, diagnostic id;
- the normative text in `profiles/manager.md` (§8.x conformance paragraph or
  §13-style surface statement) declaring the emitted hook a conformance-vector
  surface and describing the execution recipe (how a checker runs the emitted
  hook against the fixtures), OR — if this repository's `tools/validate.py` /
  Go generator cannot execute hooks — an explicit, labelled statement of the
  downstream execution binding (which implementation task owns running these
  vectors: `TASK-260910-1952mz` / `TASK-260910-3ungjy`) and a structural
  validation in `tools/validate.py` that at least checks each case's sha256
  matches its fixture bytes so the reviewer's mutant no longer passes. Say
  which of the two you did and why.
Keep the manifest registration and the `release/1.0.0-rc.9.json` derived pins
regenerated the way rev1 did.

## Evidence
`make validate` (repo venv on PATH, `set -o pipefail`, quote outputs and exit
codes) and, if you touched `tools/`, the unit tests it runs. Attach
`TASK-260910-1wjst3_spec-patch_rev2.patch` and update
`TASK-260910-1wjst3_evidence.md` (per-file changes, R1/R2 closure, validation
transcript). Tick the checklist items you satisfy, then
`task-board handoff TASK-260910-1wjst3 --role doc-writer`.
