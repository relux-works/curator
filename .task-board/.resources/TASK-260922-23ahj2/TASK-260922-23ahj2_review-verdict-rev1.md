# TASK-260922-23ahj2 — revision 1 review
Verdict: CHANGES_REQUESTED; route to to-dev for text/AC rework.

Reviewed base 937e7952c4da07b8c8f46147204c0c829a96b509 → candidate e095d4e676536eb8a4a53aa649cf89ca886dd71f. Working-tree diff against candidate is empty. Exactly six Markdown paths changed; no schema, vector, generator or executable change. No repository files modified by reviewer.

## Required rework against binding addendum and review note

1. C1 — decisions/0018-curator-run-permission-interface.md:350–361 and TASK-260922-23ahj2_results.md:114–119: mapping ownership and F-M1 are correct, but goldens are not explicitly positive AND negative interactive goldens, and argvguard rationale is missing. Add: "agents-management owns the LaunchRequest permission-mode member for LaunchModeInteractive with positive and negative interactive goldens per tool release; the module's argvguard forbids spelling argv grammar in two places, which is why F-M1 owns the mapping and F-L1 carries no argv grammar." Mirror in F-M1 AC.
   Also missing from 0018 Compatibility/results: "0017/0018 block nothing in task-board: tracked children spell bypass in the module's exec plugins themselves; 0018 serves curator run now and tracked sessions later through ax (F-A1)."

2. C2 — decisions/0017-environment-credential-modes.md:85 and :166; protocol/environments.md:1486; results.md:85–98: observation omits the exact managed path, report requirement omits detached, and F-C3 is explicitly left unchanged. Add to 0017 rationale and environments 7.4: "On the operator Mac, ~/.curator/environments/default/pi/auth.json points to ~/.pi/auth.json; that target does not exist, the real credential is ~/.pi/agent/auth.json, and env status does not show the dangling passthrough as detached." Extend F-C1: "env status and resolve report dangling or mis-targeted credential links as detached, with environment_credential_conflict-class wording instead of silence." Extend F-C3: "The production-entry suite covers a recorded link whose native target does not exist, with a narrowing mutant proving this bound." Preserve F-C2 bytes-preserved migration.

3. C3 — decisions/0017-environment-credential-modes.md:167; protocol/environments.md:1467; profiles/manager.md:2746: absent-key ⇒ file ⇒ isolated admission is present, but the requested operator evidence and design reliance are absent. Add in all three named sections: "The operator's machines have no cli_auth_credentials_store key in config.toml, so the platform default file applies and isolated is admitted; the operator's profile design relies on that platform default and on nothing more from 0017." Preserve keyring/auto refusal.

## Accepted portions and validation bounds
- Ownership grep requested by brief returns only 0018:319, the genuinely launcher-owned launch-record extension key. F-L1 narrowed; F-M1 present; F-A1 preserved. Explicit no-D5-violation/no-duplicate-grammar statement present.
- Headless marker set unchanged: CI/GITHUB_ACTIONS, TTY primary.
- Independent tools/validate.py: 62 schemas / 1124 vector files, exit 0. git diff --check clean.
- Accepted already-attached revision-1 validation log as full-suite evidence: 579 Python tests OK, Go passed after documented host SIGKILL retry, wrapper exit 0. Did not rerun full landing suite.
- Initial reviewer command using .venv failed (directory removed by producer); ambient python lacks jsonschema. Successfully used the configured gate interpreter at curator-spec/.temp/venv/bin/python3. This is an environment observation, not a product failure.
- No runtime behavior or new mutation coverage claimed: this leaf is text/AC only. Missing-target and interactive-golden runtime proofs belong to the follow-up leaves. Automated schema/vector checks do not prove completeness of these prose requirements.
- Findings recorded here only, per prohibition on LOGBOOK/control-root writes.

- Independent narrow rerun: EnvPassthroughVectorTests + ManagerConfigVectorTests + SystemConfigV2SchemaTests, 80/80 tests OK in 127.410s, exit 0 (zsh, gate virtualenv Python, run from tools).
- spawn goal queried: run is not goal-bound; no directives present.
