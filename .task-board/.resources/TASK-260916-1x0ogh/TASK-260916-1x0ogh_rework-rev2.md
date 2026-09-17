# Rework brief — TASK-260916-1x0ogh, revision 2 (answers review verdict rev1)

Read `TASK-260916-1x0ogh_review-verdict-rev1.md` (task outcome) first; it is the
authority for this round. Everything it marks passing stays as it is. Work in
the same curator-spec Story worktree
(`/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/STORY-260916-2otjbn/worktree`,
your rev1 edits are still there, uncommitted); rules in
`remediation-spec-producer-rules.md` (attach the spec diff as
`TASK-260916-1x0ogh_spec-patch_rev2.patch`).

## F2 — revision A keeps ambient-PATH selection (high; fix first, it shapes the rest)
Settled decision: revision A (`warning release`) resolves exactly as today —
the ambient `PATH` selects the provider — and only WARNS
(`subcommand_provider_outside_trust_roots`, naming the resolved path, the trust
roots consulted and the migration hint to list its directory in
`provider_directories`) when that selected executable lies outside the trust
roots; nothing else changes under A. Revision B (`flip release`) resolves
install-directory-first, then the configured `provider_directories` in order,
never ambient `PATH`, and refuses a PATH-only provider with
`subcommand_provider_untrusted`. Rewrite §11 (≈2213–2226) accordingly; add the
vector where PATH selects a different provider while a trusted one exists
(A: PATH provider runs + warning; B: trusted provider runs); fix
`listed-directory-provider-resolved` and any other both-profile vector that
encoded B's order under A; align `cli/curator.md:123` with the two profiles;
CHANGELOG consistent.

## F3 — disjoint outcomes under B (medium)
Define mutually exclusive conditions and diagnostics: trusted candidate found
(dispatch); candidate in a manager-published/managed directory (refuse,
`subcommand_provider_untrusted`); PATH-only candidate (refuse,
`subcommand_provider_untrusted`, naming path and roots); true absence
(`subcommand_provider_missing`, roots consulted); read failure (see F4). State
explicitly whether B performs a diagnostic-only PATH probe (recommendation:
yes, to name the untrusted path in the refusal) and that the `missing` row
excludes PATH-only matches. Closed set; vectors per branch.

## F4 — unreadable is not absence (medium)
A `provider_directories` entry (or the install directory) that cannot be read
MUST NOT contribute "no candidate": it is a read failure, reported as such
(admit one diagnostic into the closed set, e.g.
`subcommand_provider_root_unreadable`, naming the directory), never treated as
absence and never activating a fallback. Specify the outcome under both
profiles and the `env status` posture row; replace the
`unreadable-listed-directory-contributes-nothing` vector with the failure
outcome. This follows environments §8.4 ("unreadable evidence is reported as
unreadable, never as absence", ≈2280–2281).

## F1 — semantic enforcement of the §11 vectors (high)
`tools/validate.py` (≈4644–4688) checks shape only. Implement the resolver
model in the checker: from each case's `install_dir`, `provider_directories`,
`path_entries`, `present`, `published_dirs`, `managed_dirs` (+ readability),
derive the expected `resolved`/`diagnostic` for BOTH profiles per the rules
above (precedence, executable filtering, published/managed refusal,
missing/untrusted/unreadable distinction, migration hint fields) and compare
with the declared outcomes; refusal outcomes carry the path and consulted
roots. Add unit tests in `tools/test_validate.py`, including the reviewer's
mutant (`s6-planted-path-provider-warns-then-refuses.revision_b` resolved
instead of refused) which MUST fail.

## Validation and handoff
`make validate` and `make regenerate-check` (repo venv on PATH,
`set -o pipefail`, quote outputs and exit codes). Attach
`TASK-260916-1x0ogh_spec-patch_rev2.patch`, update
`TASK-260916-1x0ogh_evidence.md` (F1–F4 closure with file:line, transcripts),
tick the checklist items you satisfy, then
`task-board handoff TASK-260916-1x0ogh --role doc-writer`.
