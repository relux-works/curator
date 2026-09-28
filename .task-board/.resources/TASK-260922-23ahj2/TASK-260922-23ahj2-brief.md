# TASK-260922-23ahj2 brief (orchestrator, binding) — operator corrections to the landed 0017/0018 adoption

The adoption (TASK-260921-3qcjsy revision 4) LANDED on curator-spec main as 937e7952 (PR #77) after the
operator confirmed it "as is, headless marker set minimal (TTY is the primary detector)" with three
corrections that THIS leaf lands as a follow-up revision (text/AC only; no schema, vector or generator
change). Work in this Story's fresh workspace on that trunk; publish revision 1 on a green gate
(the configured gate retries a SIGKILLed go test; do not stop on that symptom yourself).

C1 — 0018: who spells the provider flag. Launcher SPEC §1 and Decision 0013 D5 hold the
invariant "the launcher never spells a provider flag". The mapping
`yolo → --dangerously-skip-permissions` (Claude Code) / `--dangerously-bypass-approvals-and-sandbox`
(Codex) therefore MUST live in agents-management (repository skill-agents-management) as a
`LaunchRequest` member for `LaunchModeInteractive`, with goldens; the launcher only resolves and
passes the permission MODE. Fix every place in 0018 and the amended normative text that attributes
the flag spelling, the argv grammar, the per-tool-release verified mapping, or the "capability
table keyed by (environment, tool release) to a grammar version" to the launcher SPEC (choices 3,
4, 6, 7 and the Compatibility section; environments §10.1/§10.2 and cli/curator.md where they echo
it): ownership = agents-management for spelling/mapping/goldens; launcher = mode resolution
(flag > profile > global > built-in default), headless detector, transport of the resolved mode,
stderr provenance line and launch-record extension key; ax (F-A1) = tracked admission
representation. State explicitly that 0018 neither violates 0013 D5 nor duplicates argv grammar.
Add follow-up **F-M1 (skill-agents-management)**: "`LaunchRequest` permission-mode member for
`LaunchModeInteractive`, provider flag mapping + goldens per tool release, drift ⇒ `yolo` refused
first" with a one-line AC; narrow **F-L1** to mode resolution/transport/provenance (no argv
grammar), keep **F-A1** as is.

C2 — 0017 Pi root: evidence, not a hypothesis. On the operator's Mac the managed home links to
`~/.pi/auth.json`, which does not exist; the real credential is `~/.pi/agent/auth.json`; the link
is dangling and `env status` does not report it. Record this observation in 0017 (choice 3
rationale / Context) and in environments §7.4; add to the F-C1 AC: "`env status` (and `resolve`)
report a dangling or mis-targeted credential link with `environment_credential_conflict`-class
wording instead of silence"; keep F-C2 (migration to `~/.pi/agent`, bytes preserved).

C3 — 0017 Codex `isolated` under `file` only: the operator's config carries no
`cli_auth_credentials_store` key, i.e. the effective store is the default `file`, so
"different profiles = different Codex accounts" is admitted. Make the text explicit: an absent
`cli_auth_credentials_store` ⇒ effective `file` ⇒ `isolated` admitted; `keyring`/`auto` ⇒
`environment_isolated_unsupported` (unchanged). Put it in §7.4, manager §12.4 and 0017 choice 4;
add to F-C1's AC ("absent key resolves to `file`").

Headless marker set: keep {`CI`, `GITHUB_ACTIONS`} as adopted (operator-confirmed; TTY primary).
results.md: C1–C3 resolution and the amended follow-up table (start from the table in the attached TASK-260921-3qcjsy_results.md).
`grep -n "launcher SPEC owns\|launcher-SPEC-owned" decisions/0018*.md protocol/environments.md cli/curator.md`
must return only ownership that is genuinely the launcher's (mode resolution/provenance/record key).
