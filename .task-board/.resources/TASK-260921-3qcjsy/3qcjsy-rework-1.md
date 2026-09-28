# TASK-260921-3qcjsy rework 1 (orchestrator, binding) — adopt 0017/0018

Verdict rev1: CHANGES_REQUESTED (TASK-260921-3qcjsy_review-verdict-rev1.md): schemas, generator,
vectors, gates, index, CHANGELOG, COMPATIBILITY are correct and in scope; three prose findings
block. Continue from the revision-1 tree (no checkout/clean/stash); fix exactly these, no
schema/vector/generator change:

F1 — `protocol/environments.md` §8.2 (~1887-1890) claims the schema-1 marker records the
passthrough credential record (isolation/source_role/backend/provenance); the frozen v1 marker
schema (`agent-environment-marker-v1.schema.json:68-80`, additionalProperties false, requires
path+strategy) rejects it, and 0017 Compatibility says frozen v1 schemas stay untouched. Keep
the §8.2 schema-1 bullet as before (strategy only); word §7.4 "Credential record" as the content
of the marker revision the F-S1 follow-up defines ("From the marker revision Decision 0017
choice 5 defines (follow-up), every passthrough record carries …; a schema-1 marker records
`path` and `strategy` only and is never rewritten to add the record"); align the 0017
Compatibility sentence and results.md.
F2 — `isolated` for `codex_cli` under `auto` storage is fail-open (on a keyring host it links
nothing and authenticates through the operator-global keyring = silently shared home, which
§7.4 forbids). Ruling (fail-closed per brief Q4): `isolated` for `codex_cli` is available under
`file` storage only; `auto` joins `keyring` in `environment_isolated_unsupported` (admission only
where the manager PROVES the effective store is `file`, probe-gated, may be named as the
follow-up's refinement). State it in §7.4 (paragraph + matrix row), manager §12.4 (paragraph +
table row), 0017 choice 4, results.md row 4.
F3 — proposal-era sentences contradict `Status: adopted`: 0017:55 and 0018:61 ("recorded below
as a proposal, not an adoption"), 0018:184 ("the adopting revision amends the fragment
schema" — the fragment is untouched; members are choice 7 / F-S2), 0018:273 ("stays open
question 7" → choice 7 follow-up), 0018:240 ("marker the adopting revision enumerates" → choice
7 fixed {CI, GITHUB_ACTIONS}; point item 5 at choice 7). Reword each to the adopted state;
`grep -n "not an adoption\|adopting revision amends\|stays open question" decisions/001[78]*.md`
must return nothing.
N1–N4 (cheap, do them): name the non-empty-directory repair diagnostic
(`environment_credential_conflict`); note the missing spec vector rows for
`environment_credential_conflict`/`environment_credential_unsupported` as part of follow-up F-S1;
mark 0018 choice 7's thin marker set {CI, GITHUB_ACTIONS} as an operator-confirmation item in
results.md (TTY prong primary); state that the launcher SPEC owns the choice-6 capability
encoding (F-L1) and align 0018 Compatibility wording.
Append "Revision 2" to results.md (F1–F3 resolution, N1–N4), rerun `make validate`, republish
only on a green gate.
