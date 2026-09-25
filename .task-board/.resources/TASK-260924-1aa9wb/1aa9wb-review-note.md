# Review note — TASK-260924-1aa9wb Skillfile schema 2 default-on, CR revision 1 (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Review against `skillfile-default-on-handoff-20260924.md` (operator handoff, binding) and `1aa9wb-brief-v2.md`; disposable clone.
1. The switch is GONE: `grep -rn "DRAFT_SOURCES\|DraftSourcesV1\|DraftSourcesEnabled" cmd internal docs README*` empty except a row proving the
   variable is ignored; install Options / manifest ParseOptions / envprofile policy fields removed; schema 2 is the default reader path
   at every call site the handoff lists.
2. Schema-1 byte identity: a `schema_version: 1` Skillfile never enters the schema-2 path — check the base-vs-candidate binary diff
   evidence on the v1 corpus/golden outputs and re-run a sample yourself (two v1 projects: resolve/install/status outputs identical).
3. Schema-2 end to end with NO env var through the real CLI: path, git-tag and repository sources (local fixtures).
4. Global scope unchanged (`internal/install/global.go` "no draft lane" behaviour kept; row proves schema 2 is not a global input).
5. Draft wording removed from diagnostics/help/docs (docs/draft-source-expansion.md, docs/draft-transport-resolution.md, docs/cli.md,
   README, troubleshooting); tests that asserted refusal-without-switch now assert default acceptance — no assertion deleted.
6. Mutants (schema-1 routed into schema-2 path; schema 2 refused again) killed — re-apply one yourself. Validation log green.
accept_cr or changes requested with file:line. No LOGBOOK.md.
