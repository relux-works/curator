# Cycle-7 logbook — TASK-260906-2x4s7i

## The acceptance deadlock, converted from assertion to measurement

Cycles 5 and 6 both *stated* that `accept_cr` was unavailable because no Change Request is
captured for this element. Neither appears to have run it. I ran it twice, from this run,
after attaching fresh evidence and completing the checklist so that a refusal could not be
about my own preparation:

```
accept_cr(TASK-260906-2x4s7i, revision=1, evidence=…_review-findings-7.md)
  -> change_request_acceptance_unauthorized: run RUN-260906-6781b8 cannot accept this
     Change Request revision: this reviewer run was handed Change Request revision 0 and is
     accepting revision 1; the revision it read is the only one it can attest to

accept_cr(TASK-260906-2x4s7i, revision=0, evidence=…)
  -> revision must be a positive integer, got "0"
```

The two constraints are mutually exclusive, so `accept_cr` is **structurally unreachable**
for a run handed revision 0. This is the shape `references/tracked-background-spawn.md`
predicts: *"A successor handed no revision is the exact shape `accept_cr` refuses, so the
recovery ladder would run to its attempt limit and park."*

**Why it matters:** the reviewer-completion guard demands acceptance be recorded by
`accept_cr` and routed through `integrating`, but this element has no Change Request — the
cycle-2 brief records that the story workspace drifted and its CR record became unusable, so
the artefact was moved to a PR instead. Every ACCEPT branch a reviewer can reach has now been
refused: `to-review` (cycle 5, "no verdict branch"), `done` (cycle 6, "cannot infer
acceptance from done"), `accept_cr` (cycle 7, unreachable, measured above). The review work
itself completed cleanly in all three cycles. **An element reviewed on a PR rather than on a
Change Request candidate has no reachable acceptance path.**

Also worth noting: the 12-item checklist was entirely unchecked through six cycles. Since
`accept_cr` requires a complete checklist, an unchecked checklist would have been a second,
independent blocker even if a revision had been handed. I completed it in this run.

## An absence and a failure to read are different facts

Both prior cycles reported "no Change Request is captured, so accept_cr is unavailable" as an
absence. It was in fact an untried call. The conclusion happened to be right, but it was
inferred from a proxy (no `*_change-request_rev*.patch` under `.resources/`) rather than
established. The typed refusals above are the actual evidence, and they say something the
proxy could not: the run *was* handed a revision number, and that number is 0.

## Review substance

- A schema whose classification is decided by regex spelling is far better tested by an
  **independently written checker of the cited authority** than by more spellings. Writing
  core §6.1 from the prose and cross-checking 3,916 generated spellings found 0 violations
  and is the single strongest artefact in seven cycles — it can rule out a whole failure
  *class* (a valid network form escaping to `path`), not just enumerated points.
- `validate_wire_semantics` never dispatches on `manager-config-v2`, so the overlay cases pin
  the schema and only the schema. Worth checking before trusting a mutant kill: a case that is
  rejected for a semantic reason would mask a schema mutant entirely.
- 25 schema cases sit on disk inside the digest-covered conformance manifest while being named
  by no `index.json` entry, so `validate.py` never drives them. All pre-date this branch. A
  checker whose corpus and whose manifest disagree about what exists is the same class as
  cycle 4's "no lane inspects the tracked file set" — both are gates blind to their own inputs.
- The changelog is now the **fourth** document in this change to carry a claim that does not
  reproduce, and F25's clause was introduced by the very commit whose subject is "say what the
  scheme rule actually encodes". Prose about a regex drifts from the regex unless something
  drives it.
