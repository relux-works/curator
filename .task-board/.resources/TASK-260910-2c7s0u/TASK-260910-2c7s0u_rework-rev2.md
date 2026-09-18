# Rework brief — TASK-260910-2c7s0u, revision 2 (R4 docs)

Revision 1 was rejected with three documentation corrections
(`TASK-260910-2c7s0u_review-verdict-rev1.md`); the replay fidelity, hygiene
and validation items passed. Documentation only — no code or test changes.
Keep everything else byte-identical.

- **F1 — attacker guarantees must match the code (README ~354–360).** An
  invalid-token request DOES reach token verification after its body is
  read (`app.py:402–408`); a valid token reaches the auditor limiter BEFORE
  signature validation (`app.py:411`, `:435–436`). Replace the "cannot"
  paragraph with stage-specific statements: without a valid token — no
  auditor-limiter access and no record append; a record append additionally
  requires a valid schema and signature. Scope the 15 s bound to
  `_read_request_body` (`app.py:611–623`), not the slot lifetime (slot
  acquisition precedes routing, release follows the route, `:148–175`); say
  the size guard rejects after the chunk that crosses the limit
  (`:616–619`) — bound the accepted/buffered body, not inbound bandwidth;
  qualify "changes no state" as "no registry append or persistent registry
  mutation" (network-limiter accounting and audit events DO happen on
  refused requests, `:131–163`). Cite file:line in the results.
- **F2 — nginx timeouts (README ~327, 335, 374–375).** `client_body_timeout`
  is an idle gap between successive reads, not a total upload deadline;
  `proxy_read_timeout` is about upstream response reads, not the incoming
  body. Correct the comment and prose, explain the two scopes, make the
  example consistent with the prose (no "≤ 15 s" claim the directives do
  not enforce); state that nginx request buffering is on by default so the
  complete body is received at the edge before forwarding and slow-client
  occupancy sits primarily at the proxy in that setup; keep the per-client
  connection and body-size caps. Cite the official nginx directive docs the
  reviewer inspected (URLs in the results).
- **F3 — README ~344**: replace the literal `[REDACTED]` placeholder with
  ordinary prose ("the bearer token is verified and the auditor limiter is
  checked").

Validation as before (test suite unchanged and green; results "Revision 2"
section written OUTSIDE the worktree and re-attached); tick the checklist;
`task-board handoff TASK-260910-2c7s0u --role doc-writer` — this is the
story-final leaf: if the handoff refuses with `stale-anchor`, quote it, run
`task-board worktree refresh-candidate TASK-260910-2c7s0u` (with replay
resolutions as in your predecessor's run) and hand off again.
