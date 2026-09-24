# Review note for TASK-260916-3dmjbc revision 2 (orchestrator, binding) — R1 script worker

Brief 3dmjbc-brief.md carries the rulings (R1 vertical slice, R2 admission honesty: preflight
refuses `script_execution_control_unavailable` until R2/R3 complete the mandatory control set,
no env/config switch opens the path; R3 evidence at real process boundaries; R4 no behaviour
change for declared-only/build/go-v1; R5 honest partial allowed). Revision 1 failed the gate
only on Windows fixtures (POSIX absolute path in config fixture; copied manager without
`.exe`); revision 2 fixed those (rework 3dmjbc-rework-1.md) and is green on all lanes: run
35579456976 — verify the gate commit resolves to the exact revision-2 tree. Spec at the CI
pin 87a0d006 (`profiles/manager.md` §3.1 eight-step order, fixed process graph, 11
mandatory controls; §2.2.1; Protocol Core §4.1.1/§4.3; vector
`conformance/v1/vectors/script-host-execution-policy.json`).

This is a large new package (`internal/scriptworker`, ~30 files) plus `internal/config`
(script_interpreters), `internal/scriptpolicy`, `cmd/curator/main.go`, `internal/godriver/identity.go`,
install rows, ledger rows, docs. Judge, with your own reruns in a disposable clone (build the
binary; bounded commands; retry once on host stalls):
1. Spec §3.1 order and graph: interpreter identity resolved ONLY from operator-trusted
   configuration (reject repository/runtime-root/`.agents/bin`/PATH/manifest — drive each
   rejection); manager self-identity (canonical regular file, symlink/reparse/hard-link
   rejection, strong identity + hash) reused from the go-v1 primitives (diff `godriver/identity.go`
   — what changed and why; no weakening of go-v1 rows); hidden-mode re-execution with a
   fresh nonce, identity recheck at the launch boundary, explicit stream binding, private
   runtime area, worker-domain termination and join; interpreter file identity verified by
   the worker; args forwarded verbatim; child exit status returned.
2. Admission honesty (R2): unsupported policies still `script_execution_policy_unsupported`;
   node-v1/python3-v1 → preflight → `script_execution_control_unavailable` naming the
   not-yet-implemented mandatory controls, shim not installed, NO way to launch an enforced
   script uncontained from any production entry (grep for switches, env, config, build tags;
   try the CLI). The implemented-controls table must be truthful (each "implemented" control
   backed by a row).
3. Process-boundary evidence: the worker mode exercised by spawning the built binary; negative
   rows (forged/copied manager, symlink/launcher-link substitution, tampered interpreter file,
   wrong/replayed nonce, PATH/manifest interpreter) refuse before the interpreter runs — run
   them; mutants (drop launch-boundary recheck, drop nonce check, accept PATH interpreter, skip
   teardown) killed — rerun at least three.
4. Windows: the process rows really run on windows-latest (check the gate's Windows evidence
   for the new rows; skip reasons in the ledger vocabulary; ledger rows registered).
5. Safety of the config seam (`script_interpreters`): absolute regular executable, no
   symlink resolution surprises, no per-package override; docs/troubleshooting section.
6. results.md: design, row table, mutant table, ratio line (vector keys consumed), bounds —
   check claims against your reruns; the six `opt_in_cases` classification unchanged.
Record exactly one verdict: accept_cr(TASK-260916-3dmjbc, revision=2, evidence=<your outcome
resource>) on ACCEPT, or changes_requested with file:line and reproduction. Non-blocking
residuals go to a separate list for the orchestrator (R2/R3/R5 will follow).
