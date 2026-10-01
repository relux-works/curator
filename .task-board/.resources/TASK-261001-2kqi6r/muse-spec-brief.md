# TASK-261001-2kqi6r — curator-spec: the muse environment adapter (curator-spec#117) (THE ONLY CURRENT INSTRUCTION; priority)

curator-spec repository. Read issue relux-works/curator-spec#117 in full, including its comments: `gh issue view 117 -R relux-works/curator-spec --comments`.
Use tb-muse's latest probe results there on auth.json refresh semantics. If they are not posted, treat refresh as an open question and
spec the fail-safe described below. Read Decision 0010 §3 (the adapter registry), §7 and Decision 0017 (credential passthrough), Decision
0018 (the permission interface), and the existing `opencode` row, which also uses XDG parents.
1. Add the `muse` environment to the adapter registry:
   - home mechanism: XDG_CONFIG_HOME, XDG_DATA_HOME, XDG_STATE_HOME and XDG_CACHE_HOME set to `<home>/{config,data,state,cache}`;
   - the adapter MUST NOT replace HOME;
   - seeds: `config/muse/settings.json` and `trust.json` from the profile;
   - skills target and root-context target as the issue measured them; mark anything unverified as such.
2. Credential passthrough `shared`: `config/muse/auth.json` is a file-link to the native auth.json, as for codex_cli and pi. State the
   refresh rule:
   - if the probe shows writes go through the link, keep the plain link;
   - if refresh replaces the file by temp+rename, the manager MUST detect a broken or forked link at resolve and repair or refuse.

   Spec the detection either way.
3. Record the permission interface per Decision 0018 (`exec --yolo`; `serve --disable-sandbox/--trust-workspace` plus per-session
   approvalMode). Record the known isolation gap: foreign personal context loads while HOME is unchanged (muse-issues 013).
4. Add environment vectors/fixtures for muse in the families the repo uses for other adapters: registry row, home layout, passthrough
   link and the no-HOME-replacement rule, generated and not hand-edited. Released schemas stay byte-identical (the validate.py guard).
   If an enum of environment ids lives in a released schema, add a new schema version rather than editing it, and say which.
5. Run tools/validate.py and make regenerate-check, with the real exit codes. CHANGELOG under Unreleased. No LOGBOOK.md. Never spell
   any employer name.
Update the results, then run `task-board handoff TASK-261001-2kqi6r --role developer`, then END YOUR TURN.
