# BUG-261016-mlpvtn \u2014 self-hosted Linux CI: task logbook

2026-10-17: Resumed after the Story convergence onto a17e58156ac30a1629e3ff36efa259bab1555a15, containing the separately landed Landlock recursion fix 6e85a912. Preserved the workflow isolation implementation and tests without any source edits in this run.

Fresh signed snapshot 3cf597414f1285714c976361ff1a8d6f72de725c passed the full normal remote gate: run 39008448258, command exit 0, 20 successful jobs and two expected optional-lane skips. All eight self-hosted Linux instances passed; no probe or fingerprint deadline failure was observed. No product deadline, new flag, host setting or repository variable was changed. Windows suite retained its existing 120m per-package limit.

Fresh focused isolation/fork-routing tests, full build, actionlint, shell syntax and whitespace validation exited 0. Reused the pre-fix expected-red control and eight killed narrowing mutants only after verifying unchanged source/test/dependency identity and intact log digests. The control fixture also equals the current checkpoint workflow exactly. Historical ABI0 race reproduction and arbitrary-loaded-host qualification remain stated bounds; three consecutive green main runs remain the orchestrator's post-integration obligation. No formal surface table was supplied in the brief.

Public evidence uses runner aliases and authoritative job links; exact runner registrations remain in a private task-local artifact to preserve the public-board boundary. The candidate remains uncommitted on the recorded Story checkpoint. See BUG-261016-mlpvtn_converged-evidence.md for the complete coverage ratio, named tests, mutant table, command exits and job mapping.
