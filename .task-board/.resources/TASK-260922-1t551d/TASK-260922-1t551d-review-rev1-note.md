# Review note for TASK-260922-1t551d revision 1 (orchestrator, binding) — F-C1 credential-link repairs (0017)

Brief 1t551d-brief.md (rulings R1 fix-first repairs never destroy bytes; R2 dangling/mis-targeted
links reported detached by `env status`/`env resolve`, Pi native root `~/.pi/agent` for new
provisioning; R3 codex `isolated` admission by effective store with absent key ⇒ `file`; R4 frozen
v1 marker untouched; R5 rows + mutants). Normative source curator-spec main 05053cd7
(environments §7.4/§7.7/§8.4.1/§10.1/§10.4, manager §12.4/§12.5, decision 0017 + operator
corrections). Gate green: run 35678477038 — verify the gate commit resolves to the exact revision-1
tree. Patch: CHANGELOG, docs/troubleshooting.md, internal/envprofile/{managed.go,status.go,
credential_link_test.go,managed_test.go}, internal/envregistry/envregistry.go.

Judge with your own reruns (disposable clone; temporary stores; bounded commands):
1. The two 0017 hazards at file:line — (a) shared→isolated leaves no stale recorded link
   (`effectivePassthrough` empty for isolated + removal set from marker surfaces only); (b) the
   unconditional `os.Remove(full)` in `finalizeMarker` replaced by "unlink only a recorded symlink
   whose target is the declared/recorded native store; otherwise refuse with
   `environment_credential_conflict` naming the path, bytes untouched". Run the mutants (restore
   the unconditional Remove; skip recorded-link removal) — each must fail a row.
2. The operator's Pi case as a driven row: recorded link → `~/.pi/auth.json` missing, real
   credential at `~/.pi/agent/auth.json`; `env status` AND `env resolve` report it detached with
   conflict-class wording (mutant: silence → fails); Pi registry strategy now targets `~/.pi/agent`
   for new provisioning; existing wrong-target homes are reported, not moved (F-C2 owns migration).
3. Codex admission table: absent `cli_auth_credentials_store` ⇒ `file` ⇒ admitted; `keyring`/`auto`
   ⇒ `environment_isolated_unsupported`; unknown selector ⇒ `environment_credential_unsupported`;
   sharing defined by the native effective storage only (no config realignment).
4. Frozen v1 marker schema untouched (no new members; `envregistry.go` change is strategy/root
   only); no secret bytes read/copied anywhere new (grep the diff for reads of auth files).
5. Windows rows follow the existing envprofile symlink-privilege pattern; ledger vocabulary only;
   CHANGELOG + troubleshooting present and accurate.
Record exactly one verdict: accept_cr(TASK-260922-1t551d, revision=1, evidence=<your outcome
resource>) on ACCEPT, or changes_requested with file:line and reproduction. Do not write into the
control root's LOGBOOK.md.
