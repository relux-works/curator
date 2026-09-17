# Rework brief — TASK-260910-1b1ens, revision 2 (R1/P1 page boundary)

Revision 1 was rejected with one finding (`TASK-260910-1b1ens_review-verdict-rev1.md`,
F1): `protocol/registry.md` §9.3 orders the client steps signature → §5
rollback (with high-water advance) → chain comparison, while
`tools/validate.py::expected_page_boundary_verdict` and the generated
`page_boundary_cases` decide the chain mismatch BEFORE the rollback step and
never advance the high-water on a mismatching page. Passing gates therefore
protect a different behaviour from the written protocol. Everything else
passed review; keep it byte-identical.

## Settled resolution (orchestrator decision — implement exactly this)
The oracle's order is the intended one; the normative text is what changes.
The **first page's boundary is the chain boundary**; later pages carry no new
information and are compared to it before anything else. §9.3 client order
becomes:

1. presence and section-2 signature verification — absent or failing:
   `registry_page_boundary_missing`, no state change;
2. for every page after the first: the boundary MUST be byte-identical to the
   chain boundary, else `registry_page_boundary_mismatch` — **no state
   change**: a later page never advances or re-checks the high-water on its
   own, equality with the chain boundary already implies the check below;
3. for the first page: the section 5 rollback rules against the persisted
   high-water — below, or equal with a different `head`/`merkle_root`/
   `log_size`: `registry_page_boundary_stale`, no state change; equal and
   same body: accepted, nothing persisted; higher: the high-water is
   persisted atomically (section 5 discipline) BEFORE the page contributes
   anything, then accepted.

State the persistence rule explicitly in the text: only a first page that
passes steps 1 and 3 changes rollback state; a page rejected at any step
leaves it untouched. Diagnostic precedence for a later page is therefore
`missing` > `mismatch`, and `stale` can only be reported for a first page.
Keep the exclusion consequence, the closed three-diagnostic table, the
missing-boundary direct rollout and the posture sentence unchanged.

## Deliverable (revision 2)
1. `protocol/registry.md` §9.3: the ordered list rewritten as above with the
   persistence sentence; §5 wording ("Page boundaries advance and check the
   same rollback state") made consistent (the chain boundary advances it).
2. `profiles/registry-service.md`: only if a sentence there restates the
   client order — otherwise untouched.
3. `tools/validate.py`: the oracle docstring describes the same order as the
   text (the code order stays); the gate additionally requires the two new
   cases below to exist by name and pins their outcomes.
4. `tools/generate-vectors/main.go`: two new `page_boundary_cases`:
   - `stale-and-mismatch-reports-mismatch` (later page: `boundary_version`
     7, `stored_version` 8, `chain_boundary_equal: false` → rejected,
     `registry_page_boundary_mismatch`, `high_water_advanced: false`,
     excluded);
   - `higher-and-mismatch-never-advances` (later page: `boundary_version`
     9, `stored_version` 7, `chain_boundary_equal: false` → rejected,
     `registry_page_boundary_mismatch`, `high_water_advanced: false`,
     excluded).
   If the case shape needs a `first_page` boolean to express "later page",
   add it to every existing case (`true` for the seven existing ones except
   `chain-boundary-mismatch-rejected`, which is a later page) and to the
   oracle; otherwise document in the case comment/README that
   `chain_boundary_equal: false` only arises on a later page. Existing case
   names and outcomes stay unchanged.
5. `make regenerate` (manifest, index, rc.9 digest), `make validate` and
   `make regenerate-check` from the worktree — quote both transcripts.
6. `TASK-260910-1b1ens_evidence.md` updated with a "Revision 2" section
   (what changed per file, the two transcripts); attach
   `TASK-260910-1b1ens_spec-patch_rev2.patch` (full `git diff origin/main`
   with new files via `git add -N`) and hand off with
   `task-board handoff TASK-260910-1b1ens --role doc-writer`.

Worktree and rules as in `TASK-260910-1b1ens_brief.md` /
`remediation-spec-producer-rules.md`: work only in
`/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/STORY-260910-25yc0h/worktree`,
no commits, no pushes, no other edits.
