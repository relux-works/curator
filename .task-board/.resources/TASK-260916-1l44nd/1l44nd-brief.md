# TASK-260916-1l44nd brief (orchestrator, binding) — R3 native probes, evidence record, preflight

Story STORY-260822-2h0v9j; builds on R1 (3dmjbc) and R2 (1h82gq), both checkpointed on the Story
branch. Read first: results.md of both, the R1 rev2/rev4 and R2 rev3 review verdicts — verdict
R2 §8 "What R3 must pick up" is binding input: (1) the two remaining table rows
`inventory-controls-applied` and `closed-script-capability-evidence-record`; (3) carry the derived
path set (R-a) for Landlock write rights; (4) surface the invocation `DerivationReport` (R-f);
(5) rows R-b, R-c, R-h, hardening R-g, cosmetic R-i, test hygiene R-m; (2) stream model R-e:
implement pass-through binding only if you can do it as a bounded protocol change with rows,
otherwise state the bound in docs (R5 decides). Spec at the CI pin 87a0d006: `profiles/manager.md`
§3.1 steps 2 and 6 and the inventory section; vector `script-host-execution-policy.json` keys
`native_control_inventory` (8 controls × {linux, macos, windows}; states available /
host-conditional / unavailable with mechanisms and `unavailable_reason`s; probe_scope
per-invocation, probe_timing pre-worker-launch), `capability_evidence_record` (record/entry
fields, cardinality, `result_only`, `excluded_from`, `probe_timings`, `inventory_version`,
`record_version`), `capability_evidence_cases` (14), `preflight_cases` (5).

## Scope (R3 exactly)
- Per-invocation inventory probe BEFORE the worker starts, without host labels, cached results
  or configuration standing in for the probe: for each of the eight controls decide
  available / host-conditional-present / unavailable on this host with the vector's mechanism
  (Linux: process-group+session teardown, delegated cgroup v2 pids.max/memory.max, rlimit-fsize,
  close-on-exec + explicit release, Landlock execute/write rights, network namespace without
  interfaces; macOS: teardown, rlimit-fsize, handle restriction, the rest `unavailable` with the
  vector's reason; Windows: job-object based where the vector says so, else unavailable) — the
  probe result is the truth, and fixed-`unavailable` controls never reject an invocation
  (`fixed-unavailable-control-does-not-reject`).
- Apply exactly the controls the probe found available/present (`inventory-controls-applied`),
  reusing the go-v1 worker's primitives (job objects, cgroup delegation, Landlock, rlimits) —
  never fork them; carry the derived path set for Landlock write rights (R-a).
- Exactly ONE closed `script-capability-evidence-v1` record per invocation, result-only (never
  stdout/stderr, markers, receipts, cache keys, conformance claims), one entry per inventory
  control, pre-worker-launch timing, truthful probe/status consistency, validated by the parent
  BEFORE `permit` (frame is in place, client.go:171-190); reject missing/duplicate/extra
  entries, cached probes, foreign versions/policies, forbidden hardened claims — drive all 14
  `capability_evidence_cases` through production paths (forged/missing evidence and probe
  contradictions included).
- Preflight: mandatory-control unavailable at install and at invocation → refuse
  `script_execution_control_unavailable` without starting the worker; the Linux pids.max
  probe-available/unavailable pair with evidence applied/unavailable and the invocation
  succeeding — the five `preflight_cases` at production entries (`install.Project`/CLI +
  the native launcher).
- Surface the invocation `DerivationReport` and the evidence record through the
  operator-selected diagnostic destination (R-f) — result-only.

## Rulings
R1 Admission honesty completes here: when all eleven mandatory controls are implemented and the
probe finds them available on the host, install/invoke PROCEED end-to-end for node-v1/python3-v1
through the production entries (this is the first revision where an enforced script launches in
production) — the six `opt_in_cases` and the launch vectors are driven at the CLI entry; on hosts
where a mandatory control is unavailable, preflight refuses as the vector says. The
table-injection seam of R2 is removed or proven inert.
R2 Deny by default, fail closed; `host-conditional` controls apply only when the probe found them
present; no substitution of labels/config for probes (mutant: cached probe → row
`cached-probe-result` fails).
R3 Evidence: 14 evidence + 5 preflight cases driven; a narrowing mutant per new control/check;
Linux-specific rows (cgroup/Landlock/netns) run on ubuntu-latest and are declared skips elsewhere
only via the ledger vocabulary; Windows job-object rows on windows-latest; the hosted lanes are
the arbiter; ledger rows registered.
R4 No behaviour change for declared-only/build/go-v1; CHANGELOG `### Added` extended; docs
(troubleshooting for `script_execution_control_unavailable` per control, config reference).
R5 Large leaf: results.md in the first 10 minutes and updated as you go (the 150-minute launcher
timeout is real); publish an honest partial if the budget runs out.
results.md: probe design per platform, applied-control mapping, evidence record schema, row
table (production boundary per row), mutant table, ratio line (vector keys consumed), Windows
proof status, bounds (R-e if deferred).
