Supplementary findings E1–E6 for the environments / launch plane (posted on curator-spec PR #49 on 2026-09-16 as S8–S13; renumbered to E1–E6 in the document to avoid the manager-side S7/S8 identifiers).

### E1. Range resolution trusts every future tag of a source; no signature or provenance rule exists anywhere (High)

Decision 0012 makes root-context, skill and MCP requirements semver **ranges** (`^1.0`, `latest`) resolved to the highest satisfying `v`-tag; `profile update`/`--all` re-resolves them (0012 §8), and in-place surfaces (`~/.claude/CLAUDE.md`, `~/.codex/AGENTS.md`, skills) re-materialize on the spot. The lock pins commits only *after* resolution, and the document is explicit that "the lock is a record, not a signature" (environments §4, §8.2). Nothing in core, 0012 or environments requires or even allows verifying tag/commit signatures against an operator-pinned signer set. Whoever can push a tag to an in-range source — a compromised maintainer account, a hijacked fork used as a mirror — therefore ships new **system-prompt** and root-context bytes and new MCP `command`+`args` into every managed home and the operator's live global context on the next `profile update`, with only strict audit (secret detection) in the way. The strict-tag policy of 0012 §8 covers a *moved* tag, not a *new* one.

*Recommendation:* (1) an optional-but-lockable **signer allowlist per source** (SSH/GPG tag or commit signatures, verified by the manager before a candidate enters the lock), reported as posture like the other gates; (2) `profile update` MUST present the resolved-version delta and, when the delta introduces or changes a `class: system` module or an MCP declaration, refuse without an explicit per-run confirmation; (3) name the residual for `latest`. Pairs with `STORY-260910-2qmrb8` (hardened defaults) or its own story.

### E2. `class: system` modules are admitted transitively; the only control is an always-warn finding (High)

A `class: system` module replaces or appends the tool's system prompt (pi `SYSTEM.md` replaces it wholesale). Environments §12/§13 give it exactly one control: `context-system-module-present`, "an always-warn, never-blocking" finding. Any package anywhere in the closure — a dependency three edges below the umbrella, selected by a range — may carry one; weights order chapters, they do not gate admission. Composed with S8, a transitive dependency update rewrites the operator's system prompt with a warning nobody reads.

*Recommendation:* admit `class: system` modules only from packages the root (or a machine-config allowlist) names **directly**; make a transitive system module a resolution error (`context_system_module_transitive`) with an explicit waiver; carry the flag into the fragment's `works.relux.curator.system-modules` as today so `ax` resume refuses on drift.

### E3. The codex provisioning seed imports the native `mcp_servers` tables, so the MCP allowlist is enforced asymmetrically per adapter (Medium)

§7.4 seeds `codex_cli` with the native `config.toml` "copied whole (project trust, model, and MCP tables included)"; §7.8 then *layers* the profile's set over it with `-p curator-mcp`. A managed codex home therefore runs every native MCP server the operator ever configured, outside the profile's lock and outside the §2.2 allowlist, while `claude_code` gets `--strict-mcp-config` and runs only the profile's set. The document records the layering but not that the allowlist does not govern the seeded base.

*Recommendation:* strip `mcp_servers` from the codex seed (seed only trust/model/tui members) or state the residual in §7.4 and §7.8 and report the ungoverned entries in `env status`, as §7.6 already does for Xcode targets. Belongs with `STORY-260910-1lf0m5`.

### E4. S6 composes with umbrella discovery: a project `.agents/env.sh` can plant `curator-run` (High, composition)

§11 refuses a `curator-<name>` provider that resolves inside a manager-published or managed directory (`subcommand_provider_untrusted`). It does not refuse a provider found in a directory that a project-controlled shell hook (S6) prepended to `PATH`. `cd project && curator run claude_code` then executes the project's binary as the launcher with the operator's environment. The launcher SPEC inherits the same trust (§2 "the umbrella discovery of section 11 trusts PATH").

*Recommendation:* resolve providers only from the manager's own install directory plus an explicit machine-config provider directory list, never from the ambient `PATH`; or, at minimum, refuse providers in directories that are user-writable by anyone other than the operator, and record the resolved provider path in the launch stderr line-group. Pairs with `STORY-260910-2awkzu`.

### E5. Takeover replaces files but does not say "never write through a symlink" (Medium)

§9.5 detects a foreign-manager symlink (`environment_foreign_manager_detected`) and lets the operator take over with backup, but no sentence forbids writing *through* an existing link. The reference implementation had exactly this defect during the environments epic (takeover wrote through a chezmoi-style symlink into the foreign manager's source of truth; `os.WriteFile` follows links) and fixed it locally. A conforming implementation written from the text alone would reproduce it.

*Recommendation:* one normative sentence in §9.5/§8.3: a takeover or repair write replaces the directory entry (unlink, then create) and MUST NOT follow a symlink at the target path; `O_NOFOLLOW`-class semantics on every managed-surface write.

### E6. `path`-kind sources sit outside the source-identity allowlists and the store boundary (Medium)

Overlays and onboarding imports use the `path` kind (0012 §5/§9, environments §6/§9.6): pinned by state hash, no git identity. The MCP package allowlist and any future signer allowlist (E1) are over canonical source identities, which a `path` source does not have; S5's missing store boundary applies to the directory itself. If a `path`-kind MCP declaration package is admissible, the allowlist cannot name it and it is unbounded by construction; if it is not, the text should say so.

*Recommendation:* state explicitly which kinds may carry MCP declarations and `class: system` modules (git only is the safe answer); for `path` overlays, require the directory to pass the same ownership/permission/containment validation S5 proposes for the store.

### Minor

- **Launcher configuration family** (`defaults.json`, `ax.json`, in `/etc/curator-run/` and `$XDG_CONFIG_HOME/curator-run/`): the SPEC validates schema but not ownership; a user-writable machine file or a symlinked operator file flips tracking policy or locks. Same S5'style contract, one paragraph.
- **`env resolve --repair` on every launch** re-materializes from the store under the launcher's authority (launcher §4.1). Under S5, a tampered store is re-applied on every launch — repair doubles as persistence. Worth a sentence in S5's remediation.
- **`--strict-mcp-config` residual**: for `claude_code` the channel intentionally disables the managed home's own `.claude.json` servers (0012 D6 records this); for `codex_cli` the inverse holds (S10). The asymmetry should be one table row in §7.8.

I can fold E1–E6 into `docs/security-audit-2026-09.md` (findings + priority table + Appendix A rows) as a commit on this branch if you want them in the document rather than in the thread.
