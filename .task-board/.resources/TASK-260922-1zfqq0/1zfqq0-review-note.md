# Review note for TASK-260922-1zfqq0 rev 2 (orchestrator, binding) — F-L1a launcher SPEC 0.5.0-draft

Review the Change Request revision 2 against `1zfqq0-brief.md` (incl. the 2026-09-23 addendum) and
`1zfqq0-rework-1.md`. Read-only, in a disposable copy of the candidate tree if you run anything.

Judge independently:
1. Every numbered deliverable 1–9 of the brief is present in SPEC.md/README/CHANGELOG, with the F-S2
   names exactly as landed (`launch-env-fragment-v2`; `permissions {mode, locked, source}`; the
   `locked iff source==global` lattice) and the F-M1 v0.5.18 names (`LaunchRequest.PermissionMode`,
   `permission-grammar-v1`). A name that is invented or differs from the landed sources is a finding
   (check curator-spec `protocol/environments.md` §10 and the skill-agents-management v0.5.18 tree).
2. Decision 0013 D5: no provider flag string anywhere in the launcher docs. Re-run your own grep
   (e.g. `--dangerously`, `--yolo` as provider flag vs launcher alias, `--full-auto`,
   `bypassPermissions`, `--approval`, `--sandbox`) — the launcher's own `--yolo` alias is allowed.
3. Revision 2: exactly three diagnostics codes were added to the Go registry without behaviour, so
   SPEC §6 == registry (21) and `internal/diagnostics` gate conformance passes; no other Go change
   beyond `specVersion` + help golden. Any behaviour change is out of scope (F-L1b owns it).
4. §4.6 headless detector closed set {CI, GITHUB_ACTIONS}; tracked refusal has no untracked fallback;
   the lock sits above the precedence.
5. Gate: hosted CI green on the candidate (read the runtime's gate result); `make check` exit 0 in
   the results.

Findings → `reject_cr` with precise, checkable items. Otherwise `accept_cr`. No LOGBOOK.md writes.
