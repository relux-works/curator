# BUG-261016-mlpvtn — self-hosted-linux-ci-flaky: rework logbook

Review revision 1 found shell-cache-override-unguarded: the workflow checker
rejected YAML env overrides but accepted shell exports in the first Go
consumer. The reviewed unmutated workflow itself had private state in all
six Linux-routable definitions.

This rework reserves the three private Go state variable names in inline
shell bodies and adds TestLinuxCIWorkflowRejectsShellCacheOverrides (27
literal variants with a valid workflow control) and
TestLinuxCIWorkflowShellOverrideBehavior (the real first Go consumer and Go
binary demonstrate the unsafe export; the same body must be rejected).
Simple literal quote/escape/continuation normalization is covered. Computed
names, eval and called script/action internals remain stated analysis bounds.

The narrowing matcher mutant admitting only GOMODCACHE fails the named
GOMODCACHE/export regression. A token-preserving shell override fails
TestLinuxCIWorkflow with the entire package, including behavioral tests,
executed. All 11 scoped mutants and the pre-fix control exited 1 for their
named intended failures. Local package tests, full build, vet, actionlint,
shell syntax and diff checks exited 0. All five earlier named tests remain
byte-identical.

Checkpoint a17e58156ac30a1629e3ff36efa259bab1555a15 contains the Landlock
recursion fix 6e85a912. No product deadline, new timeout flag, host setting or
repository variable was changed. Candidate stays uncommitted. The rework
relative to review is confined to tools/ciworkflow plus the supporting
inline-shell contract paragraph in docs/self-hosted-runner-setup.md.

At this checkpoint the fresh remote gate 39016112959 is still running on
signed snapshot 5e1d61470ec525b73c216f50c531cf407cffd93e. All eight self-hosted
Linux instances are green. Aggregate success is not claimed while remaining
hosted jobs run. Final handoff evidence will carry the terminal command exit
and per-job provenance. Provider records:
https://github.com/relux-works/curator/actions/runs/39016112959

Original AC ratio remains 0 of 3 fully driven; 6 of 6 job definitions are
checked. Historical ABI0 reproduction and arbitrary host-load qualification
are unestablished. Three self-hosted main successes remain the orchestrator's
post-integration obligation per the revised brief.
