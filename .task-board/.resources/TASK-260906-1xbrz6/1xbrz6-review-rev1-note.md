# Review note for TASK-260906-1xbrz6 revision 1 (orchestrator, binding) — takeover closed set

Control root: curator-spec (separate board owner). Read the brief `1xbrz6-brief.md` (precondition),
the producer's `TASK-260906-1xbrz6_results.md`, and the source: review cycle 2 of TASK-260906-1hn93j
(two operations outside the closed set that can meet unmanaged files), environments §9.4/§9.5/§9.6
and manager §12.3. The board validation command (`spec-gate.sh` = `make validate`) ran at handoff —
reuse that evidence; run narrow checks yourself.

The orchestrator's ruling was: KEEP the five-member carrier set and resolve the question by text.
The producer added one exclusion sentence each to §9.5, §9.6, §9.4 and manager §12.3 (import
activation and the §9.4 global operations are outside the set by design; they fail closed with
`environment_surface_unmanaged_conflict` per §8.3; recovery is `profile sync --takeover` or
`profile use --takeover` then a retry), left `cli/curator.md` unchanged, and — since no vector family
enumerates the carriers — pinned the invariants in `tools/validate.py`
(`validate_takeover_closed_set_text`, wired into `main()`), with 19 unit rows.

Judge:
1. **Is the exclusion true?** Walk the actual normative text: can `profile import` activation or a
   §9.4 global operation reach a state where the named recovery does NOT unblock it (e.g. an
   import whose activation is the FIRST materialization of a scoped adapter; a `global install`
   under a profile that is current but whose surfaces were never materialized; a foreign symlink
   rather than a regular file)? The producer claims none exists — attack that claim with a concrete
   sequence read out of §9.1/§9.2/§9.4/§9.6, not from the summary. A real uncovered state is a
   finding and must come back as a decision packet, not a silent widening.
2. **Text-pin instrument**: a Python text validator is a weaker instrument than a schema case.
   Check it cannot pass vacuously (rename a section heading, move a sentence to another section,
   reword one component) and that its 19 rows really fail on widening, narrowing, reordering,
   deletion and scattering — re-run at least three of those mutants yourself. Also confirm the
   carrier list is pinned equal to the §9.5 onboarding-trigger list (drift either way must fail),
   and that the `cli/curator.md` rule ("no `--takeover` on any `profile import` / `global` line")
   is enforced rather than asserted.
3. **Consistency**: the four sentences must say the same thing with the same diagnostic name and
   the same recovery ops; manager §12.3 must not contradict §12.5 or §9.5. Check the editorial
   follow-up (TASK-260906-3o75d6) was not pre-empted.
4. `make validate` + regenerate-check green (reused), CHANGELOG entry present, no schema/vector
   change (the producer's justification for adding none must hold: verify no family enumerates the
   carriers).

Record exactly one verdict: `accept_cr(TASK-260906-1xbrz6, revision=1, evidence=<your outcome
resource>)` on ACCEPT, or a changes-requested verdict routed with `set_status` naming file:line and
an executable reproduction. Do not write into any control root's LOGBOOK.md.
