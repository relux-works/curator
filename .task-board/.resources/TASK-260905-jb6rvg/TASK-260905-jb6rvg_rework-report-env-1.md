# Rework report 1: environments.md revision 1.1 (TASK-260905-jb6rvg)

Commit `db642b1` on `task-board/story/STORY-260905-1xwg3d` and `draft/environments-revision-1-1`
(both heads equal). Parent `ec695ba` (the checkpoint). One file: `protocol/environments.md`.

**Squash note (2026-09-05, RUN-260905-f20098).** The rework first landed as `8493c3c` on top of
the cycle-1 commit `4492b7e`; change-request construction rejected that shape
(`change_request_base_authority_mismatch`: the candidate must be exactly one single-parent
commit past checkpoint `ec695ba`). The branch was therefore squashed to a single signed commit.
`db642b1` has a tree byte-identical to `8493c3c` (`git diff 8493c3c db642b1` is empty) and is
one commit past `ec695ba`. The diff for cycle-2 review is `git diff ec695ba..db642b1`; the
rework-only delta is visible as `git diff 4492b7e 8493c3c` (both objects still reachable in
the story worktree's object store) — the content is identical.

Signature line (`git log --show-signature -1` at `db642b1`):
`Good "git" signature with ECDSA key SHA256:V6JiKG7J29mjsvikcLoSVp0bLa77VTsFy12gnLO81cM`
(the `allowed_signers` warning is the local verifier-config artifact the cycle-1 reviewer noted).

`make validate` re-run at `db642b1` (venv `.temp/venv`, log
`.temp/rework-1-reverify/make-validate-01.log`): exit 0 — 152 unittests OK, `go test ./tools/...` ok.

## Findings F1–F10 → disposition

| Finding | Section(s) | Disposition |
|---|---|---|
| F1 `name` member | §7.3, §10.2, §13 | grammar sentence added ("`argument: name` additionally carries `name`; absent for `path`/`contents`"); §10.2 member list extended; codex fragment shape note added after the `mcp` bullet; §13 says the schema requires `name` exactly when `argument` is `name` |
| F2 Linux row | §7.4 | relabeled `file-link` (write behavior still unverified); strategy paragraph now has the in-place-safe case (codex, pi) and the expected-to-detach case (rename-over or unverified writer, caught by liveness row and `--repair`) |
| F3 repair lock | §10.1 | "Under `--repair` the same lock-free verification runs first; a current home emits its fragment without touching any lock; only when stale does resolve take the repair lock" — residual kept |
| F4 `required_by` | §1.3 | "empty for the root and, for an overlay, the sorted names of any members that also require it" |
| F5 shims | §9.2 step 1 | re-points shims "when the switch is machine-scope — a section 9.3 scoped switch leaves the shims alone" |
| F6 `global update|upgrade` | §9.2 | fetch only: refresh candidates, report which pins `profile update` would move, change nothing |
| F7 codex TOML bytes | §5.8 | `args = ["a", "b"]`, `", "` separators, `[]` for empty; `<name>` emitted as a TOML bare key (core §2 grammar) |
| F8 `mcp_command_unresolved` | §5.7 | row annotated "(reported at resolution and audit, section 9.1, not at materialization)"; row stays tabled once |
| F9 `argument` required | §13 | stated for the schema batch; 0012 §9 example read as pre-revision |
| F10 stale wording | §9.2 step 5 | clause: seen at a bare `env resolve`; `curator run` always passes `--repair` and repairs instead |

## Verification-sprint items → section

| # | Item | Verdict carried | Section(s) | What the text now says |
|---|---|---|---|---|
| 1 | Keychain | verified/falsified | §7.4 table, isolation paragraph, matrix; §7.7; §7.9 registry; §12.1 | `oauth.claude.profile.*` removed everywhere. Service `Claude Code-credentials`, account `$USER`; suffix `-` + sha256[0:8] of the config-dir path under `CLAUDE_CONFIG_DIR`. Strategy `per-home-keychain`. `credential_scope` member recorded. `isolated` lifted for `claude_code`/macOS at ≥ 2.1.261; `environment_isolated_unsupported` kept for opencode and pre-gate releases; residual (fresh login writes the suffixed item) requires operator |
| 2 | `.claude.json` seed | verified | §7.4 seeds table + paragraph; §8.2 marker | seed is written, not copied: `{"hasCompletedOnboarding":true,"projects":{}}`; per-launch-directory project entry `hasTrustDialogAccepted:true` (+ `hasClaudeMdExternalIncludesApproved:true` under `referenced`), keyed by the literal cwd; marker member `seeded_projects` |
| 3 | codex global cap | falsified | §7.9 | global `$CODEX_HOME/AGENTS.md` not truncated (41 KB, cap narrowed to 1000); `global_context_cap: none`; advisory `32768` kept as a prompt-budget advisory, explicitly not tied to a cap |
| 4 | `auth.json` writes | verified | §7.4 table + strategy paragraph; §7.9 | codex truncate+write same inode 0600; pi in-place under lockfile; `file-link` safe for both; mid-write truncated-snapshot caveat recorded; liveness row stays; `auth_write` members |
| 5 | first run | verified | §7.4 seeds paragraph; §7.9 | first wall is authentication on all three; codex git check needs `--skip-git-repo-check` outside git and no seed lifts it (old "trust seed" claim corrected); pi no trust wall; `exec_flags` member |
| 6 | Xcode | bundle-verified / requires operator | §7.6 | home = Xcode-internal agentic dir set as `CLAUDE_CONFIG_DIR`/`CODEX_HOME`, override `IDEChatOverrideAgenticHomeDirectory`; `~/Library/Developer/Xcode/CodingAssistant` removed; env-var injection verified; "agents read the files" docs-confidence + requires operator |
| 7 | opencode Windows | not reproducible | — | the text carried no opencode-on-Windows claim to relabel; every opencode fact remains docs-confidence (§7.9) |
| 8 | claude `@path` | falsified | §5.3, §7.4, §10.1 | in-project targets need no approval; external targets loaded only when the project entry sets `hasClaudeMdExternalIncludesApproved:true`, else silently dropped (`-p` never asks); managed-home references are external by construction, so `referenced` requires the project entry and a launch directory lacking it is a staleness reason; content injection requires operator. N14 "behind the gate" wording replaced |
| 9 | codex `-p` | verified | §7.8, §5.8, §10.2 | single `-p` (second is an argument error; operator `-p` after `--` fails the launch — 0012 Open question 3 closed in the text); accepted before/after `exec`; missing layer silently ignored → launcher MUST stat; layer file with only `mcp_servers` verified; `profile_flag` member |

## Decisions taken beyond the Fix lines (decide-and-state)

- **`shared` unsupported for `claude_code` on macOS at the pinned release** (§7.4, §7.7, §12.1): the tool
  selects the Keychain item by config dir, and the manager may not handle credential material, so no
  passthrough can make a managed home share the native login. The adapter declares `isolated` as the
  platform default there, and an explicit `shared` is the new configuration error
  `environment_shared_unsupported` (tabled once, §7.7). Rationale: a silently unmet `shared` would be the
  same "silently shared home" class M4(c) forbids, mirrored.
- **Per-launch-directory project entry as a repair-time seed write** (§7.4, §8.2, §10.1): item 8 makes the
  `referenced` form depend on a per-project key the manager cannot know at provisioning; the entry is
  written by `env resolve --repair` for its launch directory under the repair lock, recorded in the marker
  (`seeded_projects`), and its absence is staleness only under `referenced`. Rationale: keeps seeds
  marker-recorded and never hashed, keeps the launcher read-only, and keeps `-p` launches under
  `monolithic` unaffected.
- **Codex advisory value** kept at `32768` although untied from any cap (§7.9), so the M5 warning stays a
  uniform prompt-budget advisory across adapters.

## Facts still labeled docs-confidence or requires-operator

`claude_code` Linux `.credentials.json` write behavior; fresh-login confirmation of the suffixed Keychain
item; the remaining `codex_cli` `config.toml` member shapes; `pi` `settings.json`/`models.json` shapes;
every `opencode` fact (not installed), incl. its MCP server-object shape and merge order; the `claude_code`
MCP server-object shape; Xcode agents reading the root context and the default internal home path;
`-c model_instructions_file=<path>` per-invocation override for codex; referenced-form content reaching
the model.

## Not done

Nothing in scope was skipped. No push, tag, or PR. Nothing written into the control root.
`tools/__pycache__/` appeared untracked from the unittest run and was not committed.
