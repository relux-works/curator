# TASK-260922-23ahj2 addendum (orchestrator, binding) — precise operator evidence, 2026-09-22

The operator's relay `0017-0018-operator-relay-260922.md` (attached verbatim) refines C1–C3 of the
brief; make the text carry these exact facts:

C2 (Pi): the managed pi home links `~/.curator/environments/default/pi/auth.json → ~/.pi/auth.json`
and that target does not exist; the real credential is `~/.pi/agent/auth.json`; the passthrough is
dangling and `env status` does not show it as detached. Record the exact link path in the 0017
rationale and environments §7.4; F-C1 AC: `env status`/`resolve` report a dangling or mis-targeted
credential link as detached (never silent); F-C3 AC: the production-entry test suite covers exactly
this case — a recorded link whose native target does not exist — with a narrowing mutant.

C3 (Codex): the operator's machines carry no `cli_auth_credentials_store` in `config.toml` ⇒
platform default `file` ⇒ `isolated` admitted; the operator's profile design relies on the platform
default and on nothing more from 0017 — say so where the rule is stated (§7.4, manager §12.4, 0017
choice 4).

C1 (0018 flag spelling): the invariant is "the launcher never spells a provider flag; argv is the
system plugin's business" (launcher SPEC §1, Decision 0013 D5); the mapping
`yolo → --dangerously-skip-permissions | --dangerously-bypass-approvals-and-sandbox` lives in
agents-management as a `LaunchRequest` member for `LaunchModeInteractive` with interactive goldens
(positive AND negative); the module's `argvguard` forbids spelling the argv grammar in two places —
name it as the reason F-M1 exists and F-L1 must not carry argv. Add the task-board note: 0017/0018
block nothing there (tracked children spell the bypass in the module's exec plugins themselves);
0018's consumers are `curator run` now and tracked sessions later through ax (F-A1).
