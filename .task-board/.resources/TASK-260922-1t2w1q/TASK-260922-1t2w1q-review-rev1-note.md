# Review note for TASK-260922-1t2w1q revision 1 (orchestrator, binding) — F-C2 explicit credential migration

Brief 1t2w1q-brief.md (inspect → plan → apply under the manager lock; no secret copies; fail closed
on ambiguity; effective mode preserved; deterministic printed plan; apply refuses on plan drift;
`resolve --repair` never migrates silently; frozen v1 marker untouched). Normative: curator-spec
main 05053cd7 environments §7.4/§10.1, manager §12.4/§12.5, decision 0017 O3/choice 3. Base: the
Story branch with F-C1 checkpointed (mis-targeted vs dangling-to-declared states, Pi agent root).
Gate green: run 35696877595 — verify the gate commit resolves to the exact revision-1 tree. Patch:
cmd/curator/{env.go,envmigrate.go,env_migrate_test.go}, internal/envprofile/{migrate.go,
migrate_test.go,managed.go,credential_link_test.go}, internal/envregistry/envregistry.go, docs, CHANGELOG.

Judge with your own reruns (disposable clone; temporary stores; bounded commands):
1. CLI grammar and safety: which command/flags implement inspect/plan/apply; apply requires a
   plan and refuses on inventory drift (hash) — probe: change a link between plan and apply →
   refuse; apply takes the manager-home mutation lock; journaled + rollback on injected failure
   (row) restores the prior marker exactly.
2. No secret bytes copied/moved/rewritten anywhere (grep the production diff for file reads of
   credential paths; the row asserts byte identity and no new file with credential content);
   relink preserves the native file; the operator's Pi case (recorded `~/.pi/auth.json`, real
   `~/.pi/agent/auth.json`) migrates to the agent root with mode intact and the plan printed first.
3. Ambiguity fails closed: two live Pi credential files; regular file at a link path; unexpected
   target; isolated→shared account conflict — each refuses naming the operator choice; nothing
   changes on refusal.
4. `resolve --repair` does not migrate (row); status/resolve report "migration needed" pointing at
   the explicit step.
5. Mutants killed (skip drift check; copy instead of relink; migrate inside repair; skip the lock);
   Windows symlink pattern; ledger; CHANGELOG/docs accurate; nothing accepted in F-C1 regressed.
Record exactly one verdict: accept_cr(TASK-260922-1t2w1q, revision=1, evidence=<your outcome
resource>) on ACCEPT, or changes_requested with file:line and reproduction. Do not write into the
control root's LOGBOOK.md.
