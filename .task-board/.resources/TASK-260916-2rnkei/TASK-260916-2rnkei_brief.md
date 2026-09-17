# Brief — TASK-260916-2rnkei: the codex provisioning seed strips `mcp_servers` (E3)

Story `STORY-260916-1i1gfo` (codex-seed-mcp-tables-residual), wave 3 of the
2026-09 security-audit remediation; spec first, then the manager task
`TASK-260916-33abdk`.
Worktree: `/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/STORY-260916-1i1gfo/worktree`
(branch `task-board/story/STORY-260916-1i1gfo`, forked from curator-spec `main` `684c9f1`).
Rules: `remediation-spec-producer-rules.md` (attached). Role: doc-writer of
normative spec text; no implementation.

## Finding (read it first)
`docs/security-audit-2026-09.md` E3 (Medium) and Appendix B (E3 confirmed:
codex seeds `config.toml` whole, nothing strips `mcp_servers`). Environments
§7.4 "Provisioning seeds" table seeds `codex_cli` with the native
`config.toml` "copied whole (project trust, model, and MCP tables included)";
§7.8 layers the profile's MCP set over it with `-p curator-mcp`. A managed
codex home therefore runs every native MCP server the operator ever
configured, outside the profile lock and outside the §2.2 allowlist, while
`claude_code` runs only the profile set under `--strict-mcp-config`. Text to
revise: §7.4 seeds table and prose, §7.8 table (asymmetry row), §7.7
diagnostics if a diagnostic is added, §12 status posture, §13 conformance
surfaces; vectors `conformance/v1/vectors/environments.json` (or the file the
provisioning-seed cases live in — find them).

## Settled decisions (do not reopen)
- **Strip at provisioning, warn first** (operator decision; impact row "E3
  codex seed": warn-first with `env status` listing the dropped entries).
  Two explicitly labelled rollout revisions:
  - **Revision A (warning release)**: the seed is still copied whole, but
    provisioning emits `mcp_native_servers_ungoverned` (warning) naming every
    native `mcp_servers` entry the home inherited and stating that the next
    revision stops inheriting them (migration hint: declare the server in the
    profile's MCP set, or accept the loss); `env status` lists those entries
    per managed codex home as ungoverned (outside the lock and the §2.2
    allowlist).
  - **Revision B (flip release)**: the `codex_cli` seed copies `config.toml`
    with the `mcp_servers` table and every `mcp_servers.*` sub-table
    removed — exactly the trust, model and TUI members the current text
    names; provisioning reports the stripped names once
    (`mcp_native_servers_not_inherited`, warning); `env status` lists them
    as not inherited. A managed codex home runs only the profile's MCP set
    through the §7.8 channel.
- **Asymmetry row** in §7.8: one table row (or a short table) stating per
  adapter what the channel does to the home's own MCP configuration —
  `claude_code`: `--strict-mcp-config` disables every other MCP
  configuration; `codex_cli`: `-p` layers over the seeded base (revision A:
  incl. native servers, reported; revision B: none); `opencode`: documented
  merge order (state the residual: project `opencode.json`/`.opencode/`
  servers remain, as today); `pi`: none.
- Existing managed codex homes: the seed rule applies at provisioning only
  (seeds are "thereafter owned by the tool"); an existing home keeps its
  bytes; the marker's seed record carries the seed rule revision so `env
  status` can report a home provisioned under an older revision as
  `mcp_seed_unstripped` (warning) with the repair hint (re-provision).
- Closed diagnostics: exactly `mcp_native_servers_ungoverned`,
  `mcp_native_servers_not_inherited`, `mcp_seed_unstripped` (or the
  spellings the tables favour), in the §7.7 table, §12 posture, vectors and
  CHANGELOG; the seed-source snapshot that names the entries is taken from
  the native file at provisioning time and recorded in the marker seed
  record (names only, never server commands or env values).

## Deliverable
1. §7.4: seeds table row for `codex_cli` rewritten (what is copied, what is
   stripped, evidence row kept honest — mark member shapes still
   docs-confidence), prose paragraph on the rule and the marker seed record.
2. §7.8: the asymmetry row/table; §7.7 diagnostics; §12 posture rows; §8.2
   marker: the seed record gains the rule revision field if the marker
   schema admits it (check `agent-environment-marker-v1.schema.json` —
   if a schema change is needed, follow the frozen-schema versioning rule
   the repository uses and say so; otherwise record it in the closed seed
   record shape already defined).
3. Vectors: provisioning cases under both revisions — native `config.toml`
   with `mcp_servers` tables → A: copied, `mcp_native_servers_ungoverned`
   names them, posture lists them; B: seeded file lacks them,
   `mcp_native_servers_not_inherited` names them; without `mcp_servers` →
   no warning under either; pre-rule home → `mcp_seed_unstripped` posture; registered in the manifest; generator updated if the family is
   generated; existing vectors byte-identical.
4. `CHANGELOG.md` Unreleased entry "E3: …" naming both rollout revisions, the diagnostics
   and the manager/README follow-up (a managed codex home runs only the
   profile MCP set; native `~/.codex/config.toml` servers are not inherited).

## Out of scope
Implementation (`TASK-260916-33abdk`), E7 launcher notes (`STORY-260916-33vuzm`),
S4 (landed), the opencode merge-order residual beyond stating it.

## Checklist and handoff
Tick the checklist items you satisfy; attach
`TASK-260916-2rnkei_spec-patch_rev1.patch` (`git diff HEAD` of the worktree —
its base commit — with new files via `git add -N`; NOT `git diff origin/main`,
which moves) and `TASK-260916-2rnkei_evidence.md` (with the `make validate`
and regeneration transcripts), then
`task-board handoff TASK-260916-2rnkei --role doc-writer`.
