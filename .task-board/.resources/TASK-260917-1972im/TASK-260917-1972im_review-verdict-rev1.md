# TASK-260917-1972im — landing review revision 1

Verdict: **accept-landing**. No landing correction required. This verdict approves the exact PR #62 tree; it does not merge the PR or close TASK-260916-1x0ogh.

## Identity and method

Reviewed head `0da40207a70d0b6c990c8f1bc8c79e59f218bf6d`, parent `f544a017fbabcbe40ba985ce6a52af3d4b3e9ddc`. `gh pr view 62 --repo relux-works/curator-spec --json headRefOid,baseRefOid,url` independently returned those same OIDs. Delivery worktree clean before review. Live diff SHA-256 `e653bf1ff742d049c659f39812a400e4e9931a9c6a705f012acc160cec16514c` matches the attached landing patch. Accepted rev2 patch SHA-256 `2d66b496c5779a4afa982369690aac5cb81362005099db146bedc4f9f406a38a`, applied to `07e2b41` for comparison. Read prior rev2 verdict; no prior gate result substituted for the runs below. Run is not goal-bound.

Compared every landing path and every added/deleted source line to the accepted patch; inspected all differing source deltas. The brief says “six merged sources” but enumerates eight physical files; all eight are covered below. Context-only means E4 added/deleted lines are identical while landed surrounding content remains. Generated files are independently reproduced, not excused from fidelity review.

## Per-file merge table

| File | Classification and evidence |
|---|---|
| `CHANGELOG.md` | Accepted E4 added/deleted lines byte-identical; context-only landed changes preserved. |
| `cli/curator.md` | Accepted E4 added/deleted lines byte-identical; context-only landed changes preserved. |
| `conformance/v1/manifest.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/index.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/manager-config-v2/invalid-backup-retention-negative.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/manager-config-v2/invalid-current-profile-grammar.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/manager-config-v2/invalid-form-value.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/manager-config-v2/invalid-in-place-mode-value.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/manager-config-v2/invalid-isolation-value.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/manager-config-v2/invalid-mcp-package-allowlist-duplicate.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/manager-config-v2/invalid-overlay-bare-drive-letter.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/manager-config-v2/invalid-overlay-default-weight-negative.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/manager-config-v2/invalid-overlay-directory-traversal.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/manager-config-v2/invalid-overlay-empty-source.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/manager-config-v2/invalid-overlay-file-url-with-form.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/manager-config-v2/invalid-overlay-file-url.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/manager-config-v2/invalid-overlay-git-scp-no-requirement-form.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/manager-config-v2/invalid-overlay-git-uppercase-no-requirement-form.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/manager-config-v2/invalid-overlay-negative-weight.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/manager-config-v2/invalid-overlay-no-requirement-form.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/manager-config-v2/invalid-overlay-path-directory.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/manager-config-v2/invalid-overlay-path-relative-requirement-form.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/manager-config-v2/invalid-overlay-path-requirement-form.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/manager-config-v2/invalid-overlay-path-windows-requirement-form.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/manager-config-v2/invalid-overlay-path-windows-slash-requirement-form.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/manager-config-v2/invalid-overlay-range-grammar.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/manager-config-v2/invalid-overlay-revision-grammar.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/manager-config-v2/invalid-overlay-scp-backslash-path.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/manager-config-v2/invalid-overlay-scp-host-grammar-no-user.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/manager-config-v2/invalid-overlay-scp-host-grammar-with-form.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/manager-config-v2/invalid-overlay-scp-host-grammar.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/manager-config-v2/invalid-overlay-tag-grammar.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/manager-config-v2/invalid-overlay-two-requirement-forms.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/manager-config-v2/invalid-overlay-unknown-scheme-with-form.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/manager-config-v2/invalid-overlay-unknown-scheme.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/manager-config-v2/invalid-overlays-allowed-type.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/manager-config-v2/invalid-passable-env-name-grammar.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/manager-config-v2/invalid-precedence-placement.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/manager-config-v2/invalid-precedence-winner.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/manager-config-v2/invalid-provider-directories-duplicate.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/manager-config-v2/invalid-provider-directories-relative.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/manager-config-v2/invalid-require-current-profile-grammar.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/manager-config-v2/invalid-schema-version-1.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/manager-config-v2/invalid-scoped-current-value.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/manager-config-v2/invalid-shadow-acknowledged-missing-path.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/manager-config-v2/invalid-system-module-waiver-missing-reason.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/manager-config-v2/invalid-system-module-waiver-package-grammar.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/manager-config-v2/invalid-system-module-waiver-unknown-field.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/manager-config-v2/invalid-system-prompt-files-value.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/manager-config-v2/invalid-target-consented-type.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/manager-config-v2/invalid-target-participation.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/manager-config-v2/invalid-transitive-system-modules-value.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/manager-config-v2/invalid-unknown-environments-field.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/manager-config-v2/invalid-unknown-overlay-field.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/manager-config-v2/invalid-unknown-precedence-field.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/manager-config-v2/invalid-unknown-shadow-field.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/manager-config-v2/invalid-unknown-system-prompt-environment.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/manager-config-v2/invalid-unknown-target-field.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/manager-config-v2/invalid-unknown-waiver-field.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/manager-config-v2/invalid-waiver-pin-grammar.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/manager-config-v2/invalid-waiver-span-arity.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/manager-config-v2/invalid-xdg-seed-entry-opencode.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/manager-config-v2/invalid-xdg-seed-entry-path.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/manager-config-v2/valid-overlay-git-git-uppercase.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/manager-config-v2/valid-overlay-git-http-uppercase.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/manager-config-v2/valid-overlay-git-https-uppercase.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/manager-config-v2/valid-overlay-git-scp-no-user.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/manager-config-v2/valid-overlay-git-scp.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/manager-config-v2/valid-overlay-git-single-letter-host.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/manager-config-v2/valid-overlay-git-ssh-uppercase.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/manager-config-v2/valid-overlay-path-colon-later-segment.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/manager-config-v2/valid-overlay-path-relative.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/manager-config-v2/valid-overlay-path-source.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/manager-config-v2/valid-overlay-path-windows-backslash.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/manager-config-v2/valid-overlay-path-windows-double-slash.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/manager-config-v2/valid-overlay-path-windows-lowercase-drive.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/manager-config-v2/valid-overlay-path-windows-slash.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/manager-config-v2/valid-provider-directories-windows-drive.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/manager-config-v2/valid-provider-directories.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/manager-config-v2/valid-system-module-waiver.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/manager-config-v2/valid.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/system-config-v2/invalid-isolation-isolated-direction.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/system-config-v2/invalid-isolation-profile-grammar.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/system-config-v2/invalid-isolation-value.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/system-config-v2/invalid-locked-bare-environments.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/system-config-v2/invalid-locked-duplicate.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/system-config-v2/invalid-locked-unknown-environments-key.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/system-config-v2/invalid-locked-unlockable-environments-key.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/system-config-v2/invalid-locked-unprefixed-knob.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/system-config-v2/invalid-mcp-package-allowlist-duplicate.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/system-config-v2/invalid-mcp-package-allowlist-empty-entry.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/system-config-v2/invalid-overlays-allowed-type.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/system-config-v2/invalid-passable-env-name-grammar.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/system-config-v2/invalid-precedence-placement.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/system-config-v2/invalid-precedence-winner.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/system-config-v2/invalid-provider-directories-duplicate.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/system-config-v2/invalid-provider-directories-relative.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/system-config-v2/invalid-require-current-profile-grammar.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/system-config-v2/invalid-schema-version-1.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/system-config-v2/invalid-transitive-system-modules-drop-direction.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/system-config-v2/invalid-transitive-system-modules-value.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/system-config-v2/invalid-unknown-environments-field.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/system-config-v2/invalid-unknown-precedence-field.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/system-config-v2/invalid-unlockable-environments-knob.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/schema-cases/system-config-v2/valid.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/vectors/manager-config-v2.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `conformance/v1/vectors/umbrella-provider-resolution.json` | Byte-identical accepted E4 vector; regenerated exactly (14 cases / 28 outcomes). |
| `profiles/manager.md` | Union: §1 includes both transitive_system_modules and provider_directories; see quotes below. |
| `protocol/environments.md` | Union: only status conjunction and lock-list wrapping differ from accepted E4 delta; E2 content retained. |
| `release/1.0.0-rc.9.json` | Regenerated exactly; combined source inputs; see fixture preservation and gate evidence. |
| `schemas/v1/manager-config-v2.schema.json` | Accepted E4 added/deleted lines byte-identical; context-only landed changes preserved. |
| `schemas/v1/system-config-v2.schema.json` | Accepted E4 added/deleted lines byte-identical; context-only landed changes preserved. |
| `tools/generate-vectors/main.go` | Accepted E4 added/deleted lines byte-identical; context-only landed changes preserved. |
| `tools/generate-vectors/manager_config.go` | Union: accepted E4 delta except gofmt alignment; E2 defaults, cases and vectors retained. |
| `tools/generate-vectors/manager_config_test.go` | Accepted E4 added/deleted lines byte-identical; context-only landed changes preserved. |
| `tools/generate-vectors/system_config.go` | Union: list insertion and gofmt alignment; E2 error-only policy retained. |
| `tools/generate-vectors/umbrella_provider.go` | Accepted E4 added/deleted lines byte-identical; context-only landed changes preserved. |
| `tools/test_validate.py` | Union: accepted umbrella tests, landed S6 tests, combined eight-key assertion and updated provider-key drift mutant; blank-line context only otherwise. |
| `tools/validate.py` | Union: accepted E4 default mapping, resolver model, validator and main registration; only blank-line context differs; landed E2/S6 checks retained. |

## Merge and semantic evidence

The following quotes use exact-head file:line references. The eight merged physical sources all preserve the union; no third behavior was introduced. `schemas/v1/system-config-v2.schema.json` and `tools/generate-vectors/manager_config_test.go` actually retain identical E4 change lines despite shifted E2 context.

- `protocol/environments.md:512`: the root's or an active overlay's `requires.contexts` entry; every other
- `protocol/environments.md:518`: only admitted system modules. The machine policy `transitive_system_modules`
- `protocol/environments.md:522`: materialization, and the manager MUST emit the warning
- `protocol/environments.md:527`: error `context_system_module_transitive` naming the package and the module
- `protocol/environments.md:2211`: - The provider trust roots, in search order, are exactly:
- `protocol/environments.md:2224`: `provider_directories` entry — is a read failure, never absence: lookup
- `protocol/environments.md:2225`: MUST fail with `subcommand_provider_root_unreadable`, naming the first
- `protocol/environments.md:2274`: - **Revision A (warning release).** Resolution is exactly the pre-change
- `protocol/environments.md:2291`: - **Revision B (flip release).** The ambient `PATH` never selects. The
- `protocol/environments.md:2295`: skipping non-executables and never descending. The five outcomes are
- `protocol/environments.md:2301`: holds the provider, the manager performs a diagnostic-only `PATH` probe
- `protocol/environments.md:2307`: trust roots consulted — a `PATH`-only match never reports missing;
- `protocol/environments.md:2319`: | unknown subcommand with no `curator-<name>` executable in the consulted search domain — revision A: no `PATH` match; revision B: no trust-root match and no `PATH`-probe match — and every trust root readable; names the trust roots consulted | `subcommand_provider_missing` |
- `protocol/environments.md:2320`: | `curator-<name>` candidate inside a manager-published or managed directory (revision A: the `PATH`-selected candidate; revision B: a trust-root match or a `PATH`-probe match); or, under revision B only, a `PATH`-probe match outside the trust roots; names the refused path and the trust roots consulted | `subcommand_provider_untrusted` |
- `protocol/environments.md:2321`: | revision A only (warning): `curator-<name>` `PATH`-selected outside the trust roots; names the resolved path, the trust roots consulted, and the `provider_directories` migration hint | `subcommand_provider_outside_trust_roots` |
- `protocol/environments.md:2322`: | a trust root that cannot be read — the install directory or a `provider_directories` entry — under either revision; names the first unreadable root in search order; never absence, never a fallback | `subcommand_provider_root_unreadable` |
- `protocol/environments.md:2352`: `transitive_system_modules` value with every dropped system module by
- `protocol/environments.md:2354`: absolute provider path and trust verdict for every `curator-<name>`
- `protocol/environments.md:2373`: provider row is non-current when the active revision refuses or fails
- `protocol/environments.md:2380`: Warnings —
- `protocol/environments.md:2424`: | `transitive_system_modules` | `drop`, `error` | `drop` | 3, 5.5 |
- `protocol/environments.md:2425`: | `system_module_waivers` | list of `{ package, reason }` | empty | 3, 5.5 |
- `protocol/environments.md:2429`: | `provider_directories` | list of absolute paths | `[]` | 11 |
- `protocol/environments.md:2454`: `passable_env_names`, `require_current_profile`, `transitive_system_modules`,
- `protocol/environments.md:2455`: `isolation`, and `provider_directories` — a locked provider list is fleet
- `protocol/environments.md:2465`: `transitive_system_modules` only to `error`, and `system_module_waivers`
- `profiles/manager.md:57`: `environments.transitive_system_modules`, and
- `profiles/manager.md:58`: `environments.provider_directories`.
- `profiles/manager.md:1095`: The project env files in scope are closed: exactly `.agents/env.sh`
- `profiles/manager.md:1102`: trusted. Bytes are trusted ONLY when the manager's shell-hook approval
- `profiles/manager.md:1131`: profile, package, and project surface. The manager MUST NOT read an
- `profiles/manager.md:1141`: - `approved_by` — exactly `manager` or `operator` (closed);
- `profiles/manager.md:1173`: | candidate project env file has no approval record (warning) | `shell_hook_env_unapproved` |
- `profiles/manager.md:1174`: | candidate project env file digest differs from the recorded digest; re-approval required (warning) | `shell_hook_env_changed` |
- `profiles/manager.md:1183`: closed profile set is exactly `A-warning` and `B-enforcing`:
- `profiles/manager.md:1207`: as non-current; an unapproved file is reported as a warning row.
- `schemas/v1/system-config-v2.schema.json:23`: "environments.transitive_system_modules",
- `schemas/v1/system-config-v2.schema.json:25`: "environments.provider_directories"
- `schemas/v1/system-config-v2.schema.json:52`: "transitive_system_modules": {
- `schemas/v1/system-config-v2.schema.json:53`: "enum": ["error"]
- `schemas/v1/system-config-v2.schema.json:55`: "provider_directories": {"$ref": "manager-config-v2.schema.json#/$defs/environments/properties/provider_directories"},
- `schemas/v1/manager-config-v2.schema.json:494`: "transitive_system_modules": {
- `schemas/v1/manager-config-v2.schema.json:501`: "system_module_waivers": {
- `schemas/v1/manager-config-v2.schema.json:531`: "provider_directories": {
- `schemas/v1/manager-config-v2.schema.json:543`: "default": []
- `tools/generate-vectors/manager_config.go:26`: "transitive_system_modules": "drop",
- `tools/generate-vectors/manager_config.go:27`: "system_module_waivers":     []any{},
- `tools/generate-vectors/manager_config.go:31`: "provider_directories":      []any{},
- `tools/generate-vectors/manager_config.go:67`: "transitive_system_modules": "error",
- `tools/generate-vectors/manager_config.go:68`: "system_module_waivers": []any{
- `tools/generate-vectors/manager_config.go:76`: "provider_directories":    []any{"/usr/local/lib/curator/providers", "/opt/curator/bin"},
- `tools/generate-vectors/manager_config.go:180`: {name: "valid-system-module-waiver", valid: true, instance: withKnob("system_module_waivers", []any{map[string]any{"package": "sysleaf", "reason": "reviewed leaf system prompt"}})},
- `tools/generate-vectors/manager_config.go:188`: {name: "valid-provider-directories", valid: true, instance: withKnob("provider_directories", []any{"/usr/local/lib/curator/providers", "/opt/curator/bin"})},
- `tools/generate-vectors/system_config.go:7`: "passable_env_names", "require_current_profile", "transitive_system_modules",
- `tools/generate-vectors/system_config.go:8`: "isolation", "provider_directories",
- `tools/generate-vectors/system_config.go:22`: "transitive_system_modules": "error",
- `tools/generate-vectors/system_config.go:24`: "provider_directories":      []any{"/usr/local/lib/curator/providers"},
- `tools/generate-vectors/system_config.go:87`: {name: "invalid-transitive-system-modules-drop-direction", instance: withKnob("transitive_system_modules", "drop")},
- `tools/generate-vectors/system_config.go:92`: {name: "invalid-provider-directories-relative", instance: withKnob("provider_directories", []any{"rel/providers"})},
- `tools/generate-vectors/manager_config_test.go:18`: "transitive_system_modules", "system_module_waivers",
- `tools/generate-vectors/manager_config_test.go:19`: "backup_retention", "require_current_profile", "in_place_mode", "provider_directories",
- `tools/validate.py:3003`: "transitive_system_modules": ("$defs", "environments", "properties", "transitive_system_modules", "default"),
- `tools/validate.py:3006`: "provider_directories": ("$defs", "environments", "properties", "provider_directories", "default"),
- `tools/validate.py:4699`: SHELL_HOOK_TRUST_DIAGNOSTICS = ("shell_hook_env_unapproved", "shell_hook_env_changed")
- `tools/validate.py:4883`: UMBRELLA_PROVIDER_DIAGNOSTICS = {
- `tools/validate.py:4884`: "subcommand_provider_missing",
- `tools/validate.py:4885`: "subcommand_provider_untrusted",
- `tools/validate.py:4886`: "subcommand_provider_outside_trust_roots",
- `tools/validate.py:4887`: "subcommand_provider_root_unreadable",
- `tools/validate.py:4969`: def _umbrella_expected(case: dict[str, Any]) -> dict[str, dict[str, Any]]:
- `tools/validate.py:5207`: if outcome.get("resolved") != want.get("resolved"):
- `tools/validate.py:5247`: validate_shell_hook_trust_vectors,
- `tools/validate.py:5250`: validate_umbrella_provider_vectors,
- `tools/test_validate.py:2093`: validate.validate_shell_hook_trust_vectors(changed)
- `tools/test_validate.py:2094`: class UmbrellaProviderVectorTests(unittest.TestCase):
- `tools/test_validate.py:2190`: def test_s6_revision_b_silent_resolution_mutant_fails(self) -> None:
- `tools/test_validate.py:2556`: def test_section_12_2_lists_the_seven_keys_in_order(self) -> None:
- `tools/test_validate.py:2560`: "require_current_profile", "transitive_system_modules", "isolation", "provider_directories"],
- `tools/test_validate.py:2658`: text = self.text.replace("`isolation`, and `provider_directories`", "and `isolation`", 1)
- `tools/test_validate.py:2660`: with self.assertRaisesRegex(validate.ValidationFailure, "schema-only \\['provider_directories'\\]"):
- `cli/curator.md:101`: manager-profile section 8. Under Revision A (`A-warning`, the warning
- `cli/curator.md:107`: hook sources a project `.agents/env.sh` or `.agents/env.ps1` only when
- `cli/curator.md:141`: resolves to an executable named `curator-<name>` as environments §11 states —
- `cli/curator.md:142`: revision A selects on the ambient `PATH` (warning outside the trust roots),
- `cli/curator.md:143`: revision B searches only the manager install directory and

E4 retains ordered install-directory / listed-directory trust roots, executable filtering, PATH selection plus outside-root warning in A, trust-root selection and diagnostic-only PATH refusal probe in B, mutually exclusive dispatch/refusal/missing/unreadable outcomes, and published/managed directory refusal. Read failure never becomes absence. Refused/failed provider status is non-current; A outside-root warning remains current. E2 retains direct/root/overlay admission, package waivers, drop-and-warn versus error-before-lock-write, status reporting, error-only system locking and non-lockable waivers. S6 retains manager-home digest approval, closed manager/operator provenance, A warning/B enforcement, unchanged two diagnostics and posture. These remain in their existing normative sections; merge did not duplicate sections or introduce alternate spellings. Cross-surface restatements are consistent, not additional competing rules.

Closed provider diagnostics: `subcommand_provider_missing`, `subcommand_provider_untrusted`, `subcommand_provider_outside_trust_roots`, `subcommand_provider_root_unreadable`. S6: `shell_hook_env_unapproved`, `shell_hook_env_changed`. E2: `context_system_module_dropped`, `context_system_module_transitive`. Lockable set now has eight keys; `system_module_waivers` remains intentionally absent from system config. One inherited test method name still says “seven” although its assertion correctly enumerates eight; this is a non-semantic naming nit, not a landing defect.

AST comparison additionally confirmed 7/7 accepted-only E4 validator functions, its test class, 5/5 landed-only E2/S6 functions and the S6 test class unchanged. Among landed validator functions only main changes; among landed test classes only SystemConfigV2SchemaTests changes as described. Vector diagnostics cover the exact validator set, 4/4. Exact-head normative heading counts: E4 1, S6 1, E2 admission 1. Post-regeneration blob verification remained 1,351/1,351.

Measured fixture checks: 102/102 pre-existing manager/system schema fixtures equal landed f544a01 JSON after stripping only E4's provider_directories field/lock entry; 6/6 new provider schema fixtures equal accepted rev2 after stripping only landed E2 fields/lock entry. Manager valid fixture carries provider_directories at line 58, system_module_waivers at 84, transitive_system_modules at 101. System valid fixture carries providers at 25, transitive policy at 29, both lock keys at 39/41; waivers correctly remain unavailable there. Regeneration verifies manifest/index, all cases, manager vector and release pins against the combined source. Unchanged landed S6/E2 vectors and other source files remain identical to f544a01 by the exhaustive landing-diff inventory.

## Independent gates

All final gates ran in a disposable copy whose 1,351/1,351 tracked files were verified against head Git blob OIDs, using its staged exact-byte baseline for `git diff`. Shell: zsh with `set -o pipefail`; PATH prepended `/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/venv/bin`. Original delivery files were never written. Regeneration completed before final validation. No background work was left pending.

Harness correction: initial `git archive` expanded the export-subst byte-exact fixture. Initial validation/regeneration consequently exited 2. A premature retry during copy restoration also exited 2. Those are discarded harness runs, not candidate failures or passing evidence. Copy was restored directly from Git blobs, all 1,351 hashes verified, index baseline restaged; only the final transcripts below establish gates.

### make regenerate-check

```text
go run ./tools/generate-vectors -root .
git diff --exit-code -- conformance/v1 release/1.0.0-rc.5.json release/1.0.0-rc.6.json release/1.0.0-rc.7.json release/1.0.0-rc.8.json release/1.0.0-rc.9.json
EXIT_CODE=0
```

### make validate

```text
python3 tools/validate.py
validated 60 schemas and 1066 vector files
python3 -B -m unittest discover -s tools -p 'test_*.py'
.....................................................................................................................................................................................................................................................................
----------------------------------------------------------------------
Ran 261 tests in 288.064s

OK
go test ./tools/...
ok  	github.com/relux-works/curator-spec/tools/generate-vectors	1.499s
EXIT_CODE=0
```

### Original round-1 mutant at production entry

In a separate exact-head copy, changed only `s6-planted-path-provider-warns-then-refuses.revision_b` to `{"resolved":"/home/operator/work/acme/.bin/curator-run","diagnostic":null}`; refreshed its manifest digest and both rc.9 manifest pins. Ran venv `python3 tools/validate.py` (main registers the semantic gate at line 5250). The failure is semantic, not integrity masking: **1/1 requested mutants rejected**.

```text
validation failed: umbrella provider case s6-planted-path-provider-warns-then-refuses revision_b resolves '/home/operator/work/acme/.bin/curator-run' but the §11 model expects None
EXIT_CODE=1
```

Scope bound: this is specification/conformance-model validation, not execution of a manager/launcher against real OS permissions, symlinks or Windows. Prior accepted E4 semantic design is preserved; landing review does not widen its runtime claim. Architecture remains normative docs + schemas + deterministic conformance generators and validators; no runtime implementation or proposal scope added.

## Logbook entry

TASK-260917-1972im: E4 landing union with S6/E2 verified at 0da4020. Reviewer fixture-copy lesson: Git archive is not a byte copy where export-subst applies; use blob-verified bytes before validating exact-byte fixture digests. No code/logbook repository edit made under the read-only reviewer constraint; this task-scoped outcome and board notes persist the entry for coordinator incorporation.
