# Review note for TASK-260916-1l44nd revision 5 (orchestrator, binding) — R3 probes, evidence, preflight

Brief 1l44nd-brief.md (rulings R1 admission completes: with all eleven mandatory controls
implemented and the probe finding them available, enforced node-v1/python3-v1 scripts launch
end-to-end in production — the first such revision; R2 deny by default, host-conditional only when
probed present, no label/cache/config substitution; R3 14 evidence + 5 preflight cases driven,
mutant per control, Linux rows on ubuntu, Windows job-object rows on windows-latest; R4 no change
to declared-only/build/go-v1; R5 honest partial). Revisions 1–4 failed on Linux Landlock
(add_rule EINVAL on /dev/null; restrict_self without no_new_privs/locked thread; exec-denial
probe↔evidence divergence; /dev/null under confinement; missing-path diagnostic) and lint;
revision 5 is green on all lanes: run 35685793652 — verify the gate commit resolves to the exact
revision-5 tree. Read results.md's done/undone/bounds first (R-e stream model may be deferred).

This is a security-critical slice: judge with your own reruns (disposable clone; build the
binary; Linux-specific behaviour is proven by the hosted ubuntu evidence — extract the Linux row
names/results from `test-evidence-ubuntu-latest` and `race-evidence-ubuntu-latest`; on this macOS
host rerun the macOS/portable rows, the evidence-record validation rows and the mutants that are
platform-independent; bounded commands; retry once on host stalls; do not write into the control
root's LOGBOOK.md):
1. Probe: per invocation, pre-worker-launch, per platform mechanism as the vector says (Linux:
   teardown, cgroup v2 pids.max/memory.max delegation, rlimit-fsize, close-on-exec + release,
   Landlock execute/write rights, netns; macOS: teardown, rlimit-fsize, handle restriction, the rest
   unavailable with the vector's reasons; Windows: job objects where the vector says); no cached
   result, no host label, no config standing in (mutant `cached-probe-result` → row fails);
   fixed-unavailable controls never reject.
2. Landlock enforcement design: in the worker only, locked OS thread, `no_new_privs`, ABI-probed
   `handled_access_fs`, object-typed rules, `/dev/null` file-typed rule, fork/exec from the same
   thread; never `restrict_self` in the manager/test process; the derived path set carried for
   write rights (R2's R-a). Confinement effect rows on ubuntu (write outside denied, inside
   allowed; exec denial when applied) — read them in the Linux evidence.
3. Evidence record: exactly one closed result-only `script-capability-evidence-v1` record per
   invocation, one entry per inventory control, pre-launch timing, probe/status consistency,
   validated by the parent BEFORE `permit`; all 14 `capability_evidence_cases` (missing/duplicate/
   extra entries, unknown versions, contradicts-probe, cached probe, second record, foreign
   build record/policy, deferred guarantees) driven through production paths with mutants; the
   record never reaches stdout/stderr/markers/receipts/cache keys/conformance claims (grep).
4. Preflight: the 5 `preflight_cases` at production entries (install + native launcher); on a
   host with all mandatory controls available the enforced install + launch succeeds end-to-end
   through the CLI (`TestEnforcedInstallAndLaunchAtCLIEntry`-class rows) — this is the first
   revision where production launches an enforced script: confirm the R2 table-injection seam is
   removed/inert and no other switch opens or bypasses the path.
5. R2 residuals picked up (R-a path set, R-f DerivationReport surfaced, R-b/R-c/R-h rows, R-g
   hardening, R-i, R-m hygiene) or honestly listed as undone with an owner (R4/R5).
6. Windows: job-object rows on windows-latest (extract from the Windows evidence); declared skips
   only via the ledger vocabulary; ledger rows registered; CHANGELOG/docs.
Record exactly one verdict: accept_cr(TASK-260916-1l44nd, revision=5, evidence=<your outcome
resource>) on ACCEPT, or changes_requested with file:line and reproduction.
