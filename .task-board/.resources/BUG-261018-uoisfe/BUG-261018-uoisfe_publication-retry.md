# Publication retry evidence


## Publication retry — round 4

No source, test, documentation, configuration or candidate file changed in this
run. The entire materialized worktree still equals signed candidate
`3e00fcc8abe75e411a88773c65c4042ae2448eb0`, tree
`2fe068da665b203ed48e52cfcb7fba74de96e37e`; signature verification reports `G`.
Story HEAD remains at checkpoint
`5d8e6b07f7f485557d9bbb065c9692348ca065c2`; the candidate stays uncommitted.
The isolated-index whole-tree comparison and signature check exited 0. The
real index and Story branch were not modified.

Full local, hosted and mutant evidence from round 3 is reused on the unchanged
candidate; those suites were not rerun by this publication-only producer. The
configured post-turn validator reruns automatically after handoff. Its result
is pending and is not claimed green here. The approved host environment was
sourced for this run's signing/identity check and handoff; the round-4 brief
states that the validator launch environment is repaired. Both earlier runtime
snapshot-signing failures (CR revisions 2 and 3, exit 1) remain recorded above.

No refresh-candidate was requested by the board at preflight. This retry
preserves the exact candidate identity. Coverage remains the recorded 12 of 12
surface rows and 5 of 5 AC rows; no accepted row was reopened. Muse multi-root,
adapter-ledger coexistence and opencode remain the recorded open spec questions.
No product operation or test used the real HOME as a write target. Only board
CLI mutations update the authoritative board; no control-root file was edited.

Identity check (exit 0):
```json
{
  "candidate": "3e00fcc8abe75e411a88773c65c4042ae2448eb0",
  "candidate_tree": "2fe068da665b203ed48e52cfcb7fba74de96e37e",
  "worktree_tree": "2fe068da665b203ed48e52cfcb7fba74de96e37e",
  "signature": "G",
  "story_head": "5d8e6b07f7f485557d9bbb065c9692348ca065c2",
  "unchanged": true
}
```


### Publication retry refusal — checklist sequencing blocker

The prescribed-environment `task-board handoff BUG-261018-uoisfe --role developer`
ran synchronously and exited **1**. It refused because checklist item 24 is
unchecked: "Post-turn Change Request validation succeeds under the prescribed
runtime signing environment; two prior signing failures remain recorded".
A fresh compact board read exited 0 and confirms items 1-23 checked, item 24
unchecked, and status development. The handoff also warned that it could not
compare board content for the current run and recorded no progress for that
command; that warning is retained, not interpreted as a clean audit.

The candidate is unchanged and no source repair is needed. The earlier two
signing failures remain above. This retry has no successful post-turn validator
result: a future validation is neither a passing command nor evidence that can
truthfully check item 24. No checklist item was checked by inference. No
refresh-candidate refusal occurred; refresh is unrelated to this blocker.

Constraint: the handoff prerequisite demands an outcome described by the brief
as occurring automatically after the producer turn. The approved signing
environment fixes the former signing issue but cannot supply that future
outcome before this handoff. Altering the requirement, falsely checking it or
substituting prior manual hosted evidence would weaken the acceptance contract.

Required external input: the runtime/board owner must either run and record the
actual configured CR validation under the approved environment and check item
24 from that evidence, or correct the checklist/lifecycle sequencing so the
post-turn gate is enforced after producer handoff. Recommendation: keep the
validation mandatory and repair its sequencing. Then retry handoff of the same
unchanged candidate. No request to weaken signing, checks or review is made.
Status is routed to blocked on this external lifecycle prerequisite; no review
handoff or successful new CR validation is claimed.
