# Addenda for environments.md 1.1 from the verification-sprint review (TASK-260905-3jq1so, ACCEPT)

Two additive text notes to fold into the next edit of `protocol/environments.md` (cycle-2 reviewer: treat their absence as a minor; the next producer applies them):

1. **§5.3 referenced/linked forms (sprint item 8 gap, claude 2.1.261 bundle):** with `hasClaudeMdExternalIncludesApproved` unset for the project, Claude Code silently skips not only external `@` targets but also a user-level `$CLAUDE_CONFIG_DIR/CLAUDE.md` that is itself a symlink (depth 0) or a hard link (`nlink > 1`). State it next to the external-include rule; the `linked` mode's `CLAUDE.md` surface must therefore be a regular file (copy), never a link, unless the seed sets the key.
2. **§5.8/§7.8 codex layer (sprint item 9):** `-p <name>` layers `$CODEX_HOME/<name>.config.toml` even under `--strict-config`; a missing layer file is still silently ignored, so the stat-before-launch rule stands under strict config too.

Compliance observation recorded, not a text change: the researcher's run window shows an mtime change on the real `~/.codex/auth.json` with an unknown cause (the researcher's probes were metadata-only). No content change was demonstrated; noted for the operator.
