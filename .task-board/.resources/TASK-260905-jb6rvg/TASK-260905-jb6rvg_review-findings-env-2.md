# Review findings: environments.md revision 1.1, cycle 2 (TASK-260905-jb6rvg)

Subject: Change Request `CR-TASK-260905-jb6rvg-2` revision 2, commit `3ce0d5a` on
`task-board/story/STORY-260905-1xwg3d` (cherry-pick of `db642b1`; base `ec695ba`, candidate tree
`412edf0` = HEAD tree, `git diff db642b1 HEAD` empty). One file, `protocol/environments.md`
(+1511/−414). Rework-only delta reviewed as `git diff 4492b7e HEAD` (+159/−66). Reviewer run
`RUN-260905-80d190`, 2026-09-05.

## Verdict

**ACCEPT** revision 2. All ten cycle-1 findings F1–F10 are applied exactly as the Fix lines and the
author decisions in `producer-brief-env-rework-1.md` say; all nine verification-sprint items are
folded in with the sprint's implied text and the registry members; every fact the text now states as
verified reproduces on the installed binaries. The two `env-sprint-addenda.md` notes are absent, as
the cycle-2 brief predicted — recorded below as minors for the next edit, not blocking.

## Gates re-run by this reviewer

| Gate | Result | Evidence |
|---|---|---|
| Candidate identity | HEAD `3ce0d5a`, tree `412edf05…` = CR candidate tree; `git diff db642b1 HEAD` empty; `git status` clean | `git rev-parse HEAD^{tree}` |
| Commit signature | Good `git` signature, author Ivan Oparin, ECDSA `SHA256:V6JiKG…` (`allowed_signers` warning = local verifier config, as in cycle 1) | `git log --show-signature -1` |
| Scope | exactly one path, `protocol/environments.md` | `git diff --stat ec695ba HEAD` |
| `make validate` | exit 0: 57 schemas, 780 vectors, 152 unittests OK, go test ok | `.temp/review-env-2/make-validate-01.log` |
| Cross-references | 57 headings; every `section N.N` reference resolves | script |
| Byte identity vs `ec695ba` | §1.2, §5.2, §7 body, §7.2, §8 body, §8.4, §9 body, §10 body identical | heading-split diff |
| claude 2.1.261 | `--system-prompt-file`, `--append-system-prompt-file` accepted; `--mcp-config`, `--strict-mcp-config` in help | direct invocation |
| codex 0.153.2 | `-p a -p b` → clap error "cannot be used multiple times"; `-p nonexistent` silently ignored (exit 0); a layer file with only `[mcp_servers.*]` is applied via `-p curator-mcp` and listed; `--strict-config` unsupported for `codex mcp` (so the addendum's strict-config claim could not be re-checked on this subcommand) | direct invocation, scratch `CODEX_HOME` |
| pi 0.84.2 | `--append-system-prompt <text>` ("text or file contents"); `--system-prompt-file` → "Unknown option"; no MCP flag | direct invocation |
| opencode | not installed; every opencode fact stays docs-confidence in the text | `command -v opencode` |

## F1–F10 → verification

| Finding | Section(s) | Result |
|---|---|---|
| F1 `name` member | §7.3, §10.2, §13 | verified: grammar sentence present; member list "`flag`, `argument`, `name` when `argument` is `name`, OPTIONAL `with`"; codex fragment shape note after the `mcp` bullet; §13 requires `name` exactly when `argument` is `name` |
| F2 Linux row | §7.4 | verified: relabeled `file-link`, write behavior still `unverified`; strategy paragraph carries the in-place-safe case (codex, pi) and the expected-to-detach case (rename-over or unverified) caught by the liveness row and `--repair` |
| F3 repair lock | §10.1 | verified: lock-free verification first; current home emits without touching any lock; repair lock only when stale; residual kept |
| F4 `required_by` | §1.3 | verified, matches §6 rule 4 |
| F5 shims | §9.2 step 1 | verified: machine-scope only; scoped switch leaves shims alone; consistent with §9.4 |
| F6 `global update\|upgrade` | §9.2 | verified: fetch only, report which pins `profile update` would move, change nothing |
| F7 codex TOML bytes | §5.8 | verified: `args = ["a", "b"]`, `", "` separators, `[]`; `<name>` as bare key under core §2 |
| F8 `mcp_command_unresolved` | §5.7 | verified: annotated "(reported at resolution and audit, section 9.1, not at materialization)", tabled once |
| F9 | §13 | verified: `argument` required on every `flag` descriptor; 0012 §9 example read as pre-revision |
| F10 | §9.2 step 5 | verified: bare `env resolve` surfaces it; `curator run` repairs |

## Sprint items → verification

| # | Item | Section(s) | Result |
|---|---|---|---|
| 1 | Keychain | §7.4 table + isolation paragraph, §7.7, §7.9, §12.1 | verified: `oauth.claude.profile.*` absent from the text (grep 0); service `Claude Code-credentials`, account `$USER`, `-`+sha256[0:8] suffix under `CLAUDE_CONFIG_DIR`; `credential_scope` member; `isolated` lifted for claude_code/macOS ≥ 2.1.261 with the login-flow residual marked requires-operator; opencode and pre-gate keep `environment_isolated_unsupported`. Author decision `environment_shared_unsupported` (claude_code/macOS at pinned release) is consistent: the manager may not handle credential material (§7.4), so no `shared` passthrough exists; tabled once in §7.7; knob default in §12.1 updated |
| 2 | `.claude.json` seed | §7.4, §8.2, §10.1 | verified: written `{"hasCompletedOnboarding":true,"projects":{}}`; per-launch-directory `hasTrustDialogAccepted` (+ `hasClaudeMdExternalIncludesApproved` under `referenced`), literal cwd key; marker `seeded_projects`; under `referenced` a missing entry is staleness, repaired under the repair lock. Note: the sprint's minimal seed carries the project entry at seed time; the text splits it into provisioning + repair-time entry because the launch directory is unknown at provisioning — a justified refinement, seeds still never hashed and marker-recorded |
| 3 | codex global cap | §7.9 | verified: `global_context_cap: none`, advisory kept at 32768 as prompt-budget advisory, reason stated |
| 4 | `auth.json` writes | §7.4, §7.9 | verified: in-place for codex (0600, same inode) and pi (lockfile); `file-link` safe under in-place writers; truncated-snapshot caveat; liveness row stays; `auth_write` members |
| 5 | first run | §7.4, §7.9 | verified: auth wall first on all three; codex git wall not lifted by `trust_level`; `exec_flags` member; old trust-seed claim gone |
| 6 | Xcode | §7.6 | verified: `CodingAssistant` path gone (grep 0); Xcode-internal `CLAUDE_CONFIG_DIR`/`CODEX_HOME`, `IDEChatOverrideAgenticHomeDirectory`; env-var injection verified; "agents read the files" docs-confidence + requires operator |
| 7 | opencode Windows | — | no opencode-on-Windows claim exists to relabel; all opencode facts docs-confidence |
| 8 | claude `@path` | §5.3, §7.4, §10.1 | verified: in-project targets need no approval; external targets need `hasClaudeMdExternalIncludesApproved: true` else silently dropped; managed-home references are external by construction → `referenced` requires the entry; N14 "behind the gate" wording replaced; content injection requires operator |
| 9 | codex `-p` | §7.8, §5.8, §10.2 | verified on the binary this cycle (single `-p`, before/after `exec`, silent-if-missing → launcher MUST stat); 0012 Open question 3 closed in the text; `profile_flag` member |

## Minor findings (next edit; not blocking, per `review-brief-env-2.md`)

### F11 — minor — §5.3: addendum 1 absent (linked user-level `CLAUDE.md` skipped without the approval key)

Quote (§8.1): "managed homes always link from the store". Missing rule: with `hasClaudeMdExternalIncludesApproved` unset for the project, Claude Code 2.1.261 also skips a user-level `$CLAUDE_CONFIG_DIR/CLAUDE.md` that is itself a symlink or a hard link (`nlink > 1`). Under the `monolithic` form the seed does not set the key, so a store-linked `CLAUDE.md` in a managed `claude_code` home would be silently ignored. Fix (addendum wording): state it beside the external-include rule; the `claude_code` root-context surface in a `linked`/managed home MUST be a regular file (copy), never a link, unless the project entry sets the key. Flagging it for the producer as the more consequential of the two minors: it touches the §8.1 "managed homes always link" default for one adapter's one surface.

### F12 — minor — §5.8/§7.8: addendum 2 absent (`-p` layering under `--strict-config`)

Add: `-p <name>` layers `$CODEX_HOME/<name>.config.toml` even under `--strict-config`; a missing layer is still silently ignored, so the stat-before-launch rule stands under strict config too. (Not re-checkable via `codex mcp` on this machine — `--strict-config` is rejected for that subcommand; the sprint reviewer's evidence stands.)

### F13 — nit — §7.9 members table: `claude_code` `global_context_cap: none recorded` and `pi` `none recorded` are not labeled verified or docs-confidence

Both are "adopted default" facts; one word ("docs-confidence") each would keep the §7.3 discipline uniform in that table.

## Diagnostics hygiene

57 tabled rows by this reviewer's regex. Duplicates are exactly the five pre-existing at `ec695ba` (`profile_unknown`, `context_manifest_invalid`, `environment_surface_unmanaged_conflict`, `environment_unknown`, `environment_import_lossy`); no new duplicate. The one new diagnostic, `environment_shared_unsupported`, is tabled once (§7.7). `environment_isolated_unsupported` row rewritten in place. Absence-vs-unreadable kept.

## Decision 0013 consistency

§10.2 fragment: `profile.lock_sha256`, `precedence` object, `mcp` with sorted `env_names` union, `path_prepend` reserved, codex `mcp.channels` shape now representable (F1). §10.3 `env_names` reach a plan only through the launcher allowlist; the literal-wins collision rule is referenced, not restated. §11 `curator run` sentences match D6.4 (opencode unsupported). Unchanged from cycle 1.

## Held under attack (review §5)

Profile-influence boundary, closed registry, no templating, byte-exact determinism, always-strict audit, inert system prompt, fire-vs-manage, two modes, onboarding import — all still stand; the rework touched none of them.
