# Review note for TASK-260906-2t2t6w revision 1 (orchestrator, binding) — launcher SPEC 0.4.1-draft minors

Control root: curator-agent-launcher. Read the task brief `2t2t6w-brief.md` (precondition), the
producer's `TASK-260906-2t2t6w_results.md`, and the source findings
(`/Users/administrator/Developer/ReluxWorks/curator/curator/.task-board/.resources/TASK-260905-2czqqy/TASK-260905-2czqqy_review-findings-launcher-0.2.1.md`,
section "Minor findings", items 1–4). Gate: run 35719592315 (curator-agent-launcher CI) succeeded —
verify its head resolves to the exact candidate tree and reuse it; do not rerun the full suite.

Judge these four, each independently:
1. §4.1/§6 resolve pass-through for the widened `--repair` failure surface
   (`environment_marker_invalid`, `environment_surface_unmanaged_conflict`,
   `environment_backup_exists`, `environment_seed_unreadable`). The producer records a DELIBERATE
   deviation: Curator's stderr streams verbatim BEFORE the launcher's code line (transport
   constraint, pinned by TestRunDiagnosticsContract), whereas the finding asked for "beneath". Judge
   whether the operator-facing substance (verbatim code + message reach the operator, exactly one
   launcher diagnostic line, exit 1) is satisfied and whether the SPEC now states the form it
   actually implements. No new diagnostic mappings were to be added — confirm none were.
2. §4.6 pre-launch check order: SPEC states "binary check, then §4.5 stat, then §5.1 probe; the
   first failure is the one reported" and internal/execution was REORDERED to match. This is a
   behavioural change: verify the new order against §6's one-line rule, that the nil-Boundary guard
   stays a caller-contract error, and that the three-failures-at-once row plus both modes are
   actually driven. Attack it: a mutant that restores the old order (probe → binary → stat) must
   kill a named test.
3. §9 residual: whether `claude_code` (`--mcp-config <missing> --strict-mcp-config`) and `opencode`
   (`OPENCODE_CONFIG`) proceed silently on an absent MCP configuration is recorded as unverified,
   with the generalization rule. No invented facts.
4. `TestSpecVersionPinned` must now READ SPEC.md and README.md and assert `specVersion` appears in
   both. Re-run the producer's two mutants yourself (set SPEC.md, then README.md, back to
   `0.4.0-draft` in a disposable copy): each must fail the test. A test that still compares only a
   literal is a finding.

Also check: the three-file version bump is consistent (SPEC.md line 3, README, `specVersion` in
cmd/curator-run/main.go) at `0.4.1-draft`, the version-history row exists, `testdata/help.golden`
matches the new help output, CHANGELOG names the four minors, and NOTHING of the 0018
permission-interface text (F-L1, a separate leaf) was pre-empted. Run `make check` (or the narrow
packages) yourself with real exit codes and state the shell.

Record exactly one verdict: `accept_cr(TASK-260906-2t2t6w, revision=1, evidence=<your outcome
resource>)` on ACCEPT, or a changes-requested verdict routed with `set_status` naming file:line and
an executable reproduction. Do not write into any control root's LOGBOOK.md.
