# Decision needed: separately owned code and board repositories

Consumer: EPIC-260908-2wp8wn on Curator board; source owner: skill-project-management STORY-260908-3r282p.

## Required behavior
Code and signed PR heads belong to relux-works/curator-agent-launcher (later other package repositories); Story board-state commits belong to relux-works/curator main. The operator explicitly requires that topology and exact-reviewed-head PR delivery. Copying the board into the code repository or forcing done is not an equivalent implementation.

## Observed constraint
A0 CR-TASK-260908-qblycn-1 revision 1 is independently accepted, repository_delta=empty. A legitimate researcher integration run RUN-260907-6f5ac9 invoked worktree integrate and received integration_indeterminate: /Users/iv/Developer/ReluxWorks/curator/.task-board is outside the control root. No transaction opened and no ref moved. Full exact refusal is attached to A0 as TASK-260908-qblycn_integration-results.md.

Installed source at skill-project-management 074c249d: internal/integration/integrate.go invokes NewBoardLayout; paths.go:74 computes repoRelative(controlRoot, boardDir), rejecting an outside path. This is a one-repository commit ownership contract, not merely an incorrect setting.

## Clean paths investigated
1. Existing authoritative board remains external and run roots/binding remain explicit: spawn, evidence and review work, final integration refuses.
2. Attaching the run control checkout to its real main fixes the initially detached checkout condition without changing its binding, but cannot change the independent external-board refusal.
3. Existing approved managed-landing-clone design on TASK-260831-1xhtwg (architecture resource revision 14, lines 855-891) sets Request.BoardDir=<mlc>/<logicalBoardRepoPrefix>, explicitly placing board content inside the same landed repository tree. That feature is backlog and blocked by BUG-260903-m43v8h; even completed as designed, it does not give board-state commits a separate repository owner.
4. Hand editing CR/transaction/status records, widening path admission, or moving the shared board would violate the established contracts and were not attempted.

## Proposed product scope
Add separately declared and independently verified code-repository and board-repository bindings to task-board delivery. Preserve fresh protected-ref evidence for each. Code goes through its branch/PR/review/check/exact-head landing; board commit contains the frozen owned Story paths and ancestor aggregation only, signed in the board repository after code landing proof. Persist both phases so retries after either remote update resume idempotently; never infer Git authority from TASK_BOARD_DIR, and never advance an unrelated dirty checkout. Empty research CRs still need only a board-owned commit. Existing colocated behavior must stay unchanged.

This is an additional task-board product capability and a change beyond the already-approved MLC design, not a launcher fix or a safe one-line tool patch. Its protocol/ownership choice must be explicitly included in scope before implementation. Recommendation: include this source-level prerequisite while preserving the requested shared board topology. Alternative: leave accepted work integrating until upstream provides the capability; this cannot satisfy the current full delivery DoD yet.

## Human decision
Include this additional task-board product capability in the current workstream, or keep delivery pending on upstream? No source implementation or contract relaxation has been made.
