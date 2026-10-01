# Review note — TASK-261001-2yvag1 rev6 (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Review the newest CR (rev6, re-applied onto current trunk e87d488b, 29 paths, gate green; rev5 snapshot refs/campaign/2yvag1-rev5-20261001 for comparison) against `muse-curator-brief.md` and the binding scope split.
- This task owns ONLY the resolver side, i.e. what `curator env resolve muse --repair --format json` emits as launch-env-fragment-v3. The curator-run mapping and the v3 reader already landed separately (launcher ee66c107).
- Normative source: curator-spec candidate branch `spec-muse-environment`, d373078a:
  - protocol/environments.md, the muse sections;
  - Decision 0018 additions;
  - schemas/v1/launch-env-fragment-v3.schema.json.

Check:
1. **Spec conformance.**
   - The emitted fragment validates against the v3 schema.
   - XDG variables only (four, one managed parent); HOME is never set.
   - Link-state rows 16/16 and fragment rows 36/36 are exercised through the production entry, not helpers only.
2. **Muse auth inspection** (`internal/envprofile/muse.go` inspectMuseAuth).
   - Reads go through the stateread seam: the guard test passes without a new allowlist entry, or with a justified one.
   - No secret value is read into output or logs.
   - A symlinked or unreadable auth path fails closed.
3. **Status semantics.** A profile that never enabled muse stays "current". Earlier fixture edits must reflect intended semantics, not hide a regression. Check the six previously failing tests and say why each now passes.
4. **Mutants.** Kill at least two yourself, with real exit codes: HOME emitted; seam bypass.
5. **Hygiene.** No LOGBOOK; CHANGELOG at most one Unreleased line; no stray files; never spell any employer name. The unrelated askpass EPIPE flake is tracked as BUG-261001-2772iz and is out of scope.

Verdict: accept_cr, or changes requested with file:line.
