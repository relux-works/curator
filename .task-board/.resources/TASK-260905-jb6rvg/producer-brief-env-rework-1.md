# Producer brief: environments.md revision 1.1 — rework 1

Worktree `/Users/iv/Developer/ReluxWorks/.worktrees/curator-spec-env-1-1`, branch
`draft/environments-revision-1-1`, head `4492b7e` (the story branch carries the same commit;
commit on the story worktree as before and fast-forward the draft branch, or vice versa — both
must end at the same head). Edit ONLY `protocol/environments.md`. Apply ALL findings F1–F10 of
`TASK-260905-jb6rvg_review-findings-env-1.md` exactly as the reviewer's Fix lines say (author
decisions: F1 add the `name` member to the §7.3 grammar and §10.2 member list plus a codex
fragment note; F2 relabel the claude_code Linux row `file-link` and add the file-link-under-
rename-over case to the strategy paragraph; F3 lock-free verification first, repair lock only
when stale; F4–F8 as proposed, F6 = `global update|upgrade` fetch only and report that pins move
through `profile update`, F7 = `args = ["a", "b"]` with `", "` separators and `[]` for empty;
F9 note kept for the schema batch (state in §13 that `argument` is required on every `flag`
descriptor); F10 one clause).

Then fold in the verification-sprint evidence (`TASK-260905-3jq1so_verification-sprint.md` on
TASK-260905-3jq1so; binaries claude 2.1.261, codex 0.153.2, pi 0.84.2) — each item replaces a
docs-confidence claim with the verified statement, quoting the sprint's implied text, and the
adapter registry rows (§7.x) gain the named members:

1. **Keychain (item 1, verified/falsified)**: drop every mention of `oauth.claude.profile.*` as a
   Claude scheme (it is another app's). State: Claude Code stores OAuth credentials in the login
   Keychain as service `Claude Code-credentials` (account `$USER`); with `CLAUDE_CONFIG_DIR` set the
   service name is suffixed with `-` + the first 8 hex of sha256 of the config-dir path, so each
   managed home owns a separate Keychain item and needs its own login. Registry: `claude_code`
   `credential_scope: per CLAUDE_CONFIG_DIR (keychain service suffix sha256[0:8])`, verified from
   the 2.1.261 bundle strings and the existing Keychain items. Consequence for M4(c): lift
   `environment_isolated_unsupported` for `claude_code` on macOS **behind the pinned-release gate
   (≥ 2.1.261)** with the residual recorded that the login-flow confirmation (a fresh login writing
   the suffixed item) requires an operator; keep `opencode` unsupported; keep the
   `environment_isolated_unsupported` diagnostic for the pre-gate release and for opencode.
2. **`.claude.json` seed (item 2, verified)**: the minimal seed is
   `{"hasCompletedOnboarding":true,"projects":{"<abs project path>":{"hasTrustDialogAccepted":true}}}`
   with `"hasClaudeMdExternalIncludesApproved":true` in the project entry when the referenced form
   is materialized (item 8); the project key is the literal cwd path as used, not a realpath.
3. **codex global cap (item 3, falsified)**: `project_doc_max_bytes` applies to the project-doc
   chain only; the global `$CODEX_HOME/AGENTS.md` is not truncated (verified with a 41 KB file on
   0.153.2 via `codex debug prompt-input`). Registry `codex_cli` `global_context_cap: none`; keep
   `root_context_size_advisory_bytes` as an advisory only, with the codex value no longer tied to
   the cap (say why).
4. **auth.json write mode (item 4, verified)**: codex and pi rewrite `auth.json` in place
   (truncate + write, 0600), neither uses temp+rename; a snapshot mid-write can observe a truncated
   file — copy under the tool's lock or when idle. Registry `codex_cli`/`pi` `auth_write: in-place`
   (pi with lockfile). Adjust the §7.4 passthrough strategy rows accordingly (file-link is safe
   from detachment for in-place writers; the liveness row stays).
5. **First run (item 5)**: in non-interactive mode each tool's first wall is authentication;
   claude reaches it before any trust/onboarding prompt; codex's only other wall is the git check
   (`--skip-git-repo-check` needed outside a git repo for `exec`, which `projects.*.trust_level`
   does not lift); pi has no trust wall. Correct the seeds text where it claims a codex trust seed
   removes the `exec` wall.
6. **Xcode (item 6, requires operator + bundle-verified)**: name the home as the Xcode-internal
   `CLAUDE_CONFIG_DIR`/`CODEX_HOME` set by Xcode (override `IDEChatOverrideAgenticHomeDirectory`),
   not `~/Library/Developer/Xcode/CodingAssistant`; env-var injection verified from the Xcode 26.5
   bundle; the "agents read the root context" claim stays docs-confidence/requires-operator.
7. **opencode on Windows (item 7)**: not reproducible (host down) — keep docs-confidence.
8. **claude `@path` (item 8, falsified)**: §5.3 must say that referenced files outside the project
   are loaded only when the project entry sets `hasClaudeMdExternalIncludesApproved: true`
   (the interactive dialog sets it; `-p` never asks and silently skips); in-project targets need
   no approval; the referenced-form seed carries the key. Replace the N14 "behind the gate" wording
   with this verified rule; behavioral confirmation of injected content remains requires-operator.
9. **codex `-p` (item 9, verified)**: `-p <name>` layers `$CODEX_HOME/<name>.config.toml`, accepts
   exactly one value (a second is an argument error, not last-wins), is accepted before and after
   `exec`, and a missing layer file is silently ignored — so the adapter/launcher MUST stat the
   layer file before launch. Update §5.8/§7.8 and close Decision 0012 Open question 3 in the text.

Deliverables: one additional signed commit; `make validate` green (venv);
`TASK-260905-jb6rvg_rework-report-env-1.md` with finding → disposition and sprint item →
section table; `task-board handoff TASK-260905-jb6rvg --role developer`. No push, no PR.
Never write into the control root.
