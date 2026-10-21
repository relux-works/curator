# Claude Code 2.1.305 project-source exclusion record

TASK-261021-wt8ve7, 2026-10-21. This records native behavior, not strict-adapter
admission. **The measured fresh-session sources are selectively excludable
with `--setting-sources user`; explicit MCP strictness also works. Previously
contaminated resume/fork history is not scrubbed. Arbitrary late hook,
plugin, MCP and settings reloads remain unqualified.** No product code,
registry pin or runtime capability changes.

## Identity and method

| Item | Exact identity |
|---|---|
| Native release | `2.1.305 (Claude Code)` |
| Executable SHA-256 | `ea3ea3c08d2a9634a4651e687fcf900d410c489bce2b00f22bd7352b91ba704f` |
| Platform | macOS 27.0 (26A428), arm64 |
| Toolchain | Go 1.25.5; Python 3; system Git and Apple sandbox-exec |
| Curator base | `cdcba3bd7a8de6028d3d1dbf621ce4cb19b130a7` |
| Spec read | curator-spec `e41f96387287f5e76604627f9e63662f7d6d86d4`, CIP-0002 |
| Harness identity | [source manifest](../../.research/project-exclusion-claude/evidence/source-sha256.txt) |

The exact versioned executable was copied to task-local `.temp/` and pinned
before execution. The updating launcher is never invoked. Missing opt-in
skips before fixture creation; a wrong digest is rejected before creating
one. The binary itself reports the exact release in a sandboxed version call.
CIP-0002's requirement is source admission rather than a different HOME:
approved copies must remain usable and unapproved native sources must not
automatically configure the session at entry, reload, nested discovery,
resume or child entry. The CIP remains Draft; these measurements do not
accept it or implement its adapter.

Every vendor process runs through `sandbox-exec -> env -i`, with disposable
HOME/CLAUDE_CONFIG_DIR, all four XDG parents, TMPDIR and CLAUDE_CODE_TMPDIR,
synthetic API credentials and one exact loopback provider port. The sandbox
denies every other network endpoint, writes outside the fixture, Keychain
files/security services, real native config/library directories, system
managed settings and discovery files above the synthetic ancestor. No
operator credentials, native auth command, remote inference or account API
are used. The local SSE server returns fixed text or fixed tool-use messages;
its text is never used as proof that the agent ignored instructions.

The harness inspects actual outbound prompt inputs, native init catalogs,
real Read/Skill/Task tool results, requested model, permission denials and
hook/MCP marker-file effects. Missing requests, init, successful terminal
result or unexpected process exits fail, rather than proving absence.
Disposable Git metadata is separate from the synthetic checkout. Native
auxiliary sockets/process-inspection attempts can be denied by the sandbox;
the successful measured turns and effects are recorded independently.
This is not OAuth, Keychain, unrestricted hostile-code containment, TUI,
Windows, Linux or another Claude build qualification.

## Exact launch controls

Every turn includes:

```text
$CLAUDE -p EXCLUSION_PROBE --output-format stream-json --verbose
  --debug-file $RUN/turn.debug.log --max-turns 5 --permission-mode dontAsk
```

The exclusion branch appends `--setting-sources user`. The baseline omits
that flag. Read, Skill and Task tests explicitly allow only the tool under
measurement through `--allowedTools`. Their execution is still inside the
outer sandbox. P8 skill reload uses `--input-format stream-json` with two
ordinary SDK turns and an intervening `/reload-skills`, rather than a positional
prompt. Native children receive fixed local provider responses too.

P5 independently crosses source selection with:

```text
--mcp-config $RUN/home/.claude/approved-mcp.json [--strict-mcp-config]
```

The file always exists, including `{"mcpServers":{}}`. Source selection alone
excludes the measured project `.mcp.json` in this build. Strict mode independently
excludes it when all setting sources are enabled. Both preserve an explicitly
supplied approved server. Keeping explicit strictness and an explicit empty
file is therefore a measured candidate for an exclusive MCP channel, rather
than relying on a missing MCP argument. Neither `--bare` nor `--safe-mode`
is used: this record measures the selective home-preserving source flag.

## Per-row results

Every row has an approved managed-home positive control. Baseline project
copies use the same native formats and must load or execute before their
absence counts as exclusion evidence. The bounded late-hook observation is
the explicit exception and is classified unmeasured for exclusion.

| Row | Real observation and positive control | Verdict and exact control |
|---|---|---|
| P1 root and ancestors | Root `CLAUDE.md`, ancestor `CLAUDE.md`, `.claude/CLAUDE.md`, `CLAUDE.local.md`, `.claude/rules/*.md` and AGENTS.md fallback each enter the actual baseline input. Each is absent with exclusion; home `CLAUDE.md` remains in both inputs. | Excludable for these six entry cases with `--setting-sources user`. Imports, additional-directory discovery and symlink aliases are not independently qualified. |
| P2 nested/lazy | Real Read of `sub/ordinary.txt` returns the ordinary file in both branches. Nested `CLAUDE.md`, nested AGENTS.md fallback and path-scoped `.claude/rules` text enter baseline follow-up requests and disappear with exclusion. Home instructions remain. | Excludable on these three Read paths with `--setting-sources user`; this is not a proof for every native tool that might enter a directory. |
| P3 skills | Project skill metadata appears in the catalog/input and Skill loads its body in baseline. With exclusion its catalog entry/body disappear and invocation returns `Unknown skill: foreign-probe`. The home skill is advertised and its body loads in both branches. | Excludable for project `.claude/skills` with `--setting-sources user`, preserving approved home discovery and invocation. |
| P4 hooks/plugins | Project SessionStart hook executes only in baseline; the home hook executes in both. A project-enabled, project-scoped synthetic local plugin contributes catalog entries and executes its plugin hook only in baseline; the home-approved plugin contributes its skill and hook in both. | Excludable for these startup hooks and installed local-plugin activations with `--setting-sources user`. Fixtures register local marketplaces; no download/sync or other plugin type is qualified. |
| P5 MCP | All 16 combinations of empty/nonempty project set, source selection, strictness and empty/nonempty approved set are driven. Allowed stdio processes create admission witnesses and connect; forbidden project processes leave no witness/catalog entry. Explicit approved server connects with either exclusion control. | Excludable with `--setting-sources user`, or independently `--strict-mcp-config --mcp-config APPROVED_FILE`. An empty approved set is measured explicitly. |
| P6 settings/permissions | Project and local settings each change the baseline requested model to Haiku and deny a specific Read. Exclusion restores the managed-home Sonnet selection and that file's actual readable contents. An approved home deny still rejects its neighboring Read and the protected bytes never enter a request. | Excludable for both `.claude/settings.json` and `.claude/settings.local.json` model/Read-policy atoms with `--setting-sources user`. Arbitrary setting keys and environment/PATH assignment are not individually qualified. |
| P7 resume/children | Repeating the flag on a session started with exclusion preserves root/skill exclusion and approved home sources. A real Task child uses an approved home agent definition and receives approved home context; the project custom-agent catalog entry/root text is excluded. Child completion is observed in native stats. | Excludable on the measured safe-origin resume and one async/background native child path, with the same source flag on each launch. Resumed hooks/MCP, foreground children and other child modes remain unmeasured. |
| P7 unsafe-origin resume | A session first started without exclusion has project text in its initial input. A subsequent excluded `--resume`, with or without `--fork-session`, still sends that text from history. Approved home context still loads. | Not excludable by the source flag alone for preexisting contaminated transcripts. Require proven excluded origin/bound context, a fresh session, or refusal; no transcript-scrubbing adapter is implemented here. |
| P8 late files | A nested instruction created after request 1 enters the baseline real Read follow-up and is excluded with the flag; work data/home context remain. Late home and project skills are created after request 1, then `/reload-skills` and another SDK turn run: the approved home skill appears, the project one appears only in baseline. | Excludable for these late instruction and explicit skill-reload paths with `--setting-sources user`. |
| P8 late hooks and remaining reloads | A new project PreToolUse hook is written during request 1. At the next Read after 800ms it does not execute even in baseline; a startup-approved home PreToolUse hook does execute. The discriminating positive control for newly loaded hook settings is therefore absent. | Unmeasured for late hook/settings exclusion. Late plugin/MCP discovery, broader settings reload and arbitrary timing are also unmeasured; startup absence is not promoted to their qualification. |

The plugin fixture intentionally needs known-marketplace and installed-plugin
metadata to make its baseline real. Initial missing-marketplace/missing-directory
attempts were fixture failures, not exclusion evidence. Likewise the initial
late-hook activation failure is retained as a non-discriminating observation,
not relabeled a successful exclusion row.

## Evidence, controls and validation

The [harness README](../../.research/project-exclusion-claude/README.md)
documents commands and output locations. [Observations](../../.research/project-exclusion-claude/observations.json)
retain per-invocation argv, exits, catalogs, explicit marker verdicts for every
actual request, bounded prompt/tool excerpts, loader messages and file effects.
Full sanitized request/debug bodies stay in the private task-local run root.
Ordinary and flattened native task-path locators are redacted before publication.

The gated matrix has **64 leaf cases**: 60 native/intake/boundary cases and four
publication-checker controls. It records 60 successful provider turns plus
one sandboxed identity call. Intake tests also drive missing/wrong-digest child
entries without running a vendor. These are evidence-collection results;
unsafe-origin resume is a green failure witness and the non-discriminating
late-hook cases confer no exclusion qualification.

| Check | Expected result and evidence |
|---|---|
| Ungated harness | Exit 0; all vendor/publication leaves skip before fixture creation. [Log](../../.research/project-exclusion-claude/evidence/ungated.log). |
| Gated harness | Exit 0; 64 leaf cases with per-test claims. [Log](../../.research/project-exclusion-claude/evidence/replay.log). |
| Source-control narrowing | Drop only `--setting-sources user`; P1/root/exclude-true exits 1 with the foreign-marker assertion. |
| MCP-control narrowing | Keep all setting sources enabled and drop only strict MCP; the forbidden server starts and the matching row exits 1. |
| Offline boundary narrowing | Dropping network denial admits only the synthetic unrelated-loopback control and fails its rejection assertion; dropping Keychains denial similarly exposes only the synthetic file and fails its assertion. No vendor runs under either weakened policy. |
| Control runner | Exit 0 only after all four intended failures, including their exact diagnostics. [Summary](../../.research/project-exclusion-claude/evidence/mutants.log). |
| Publication checker | Neutral synthetic input passes. Ordinary/flattened private locators and an empty run are refused with exact diagnostics, with no partial output. Included in the gated log. |
| Build/static validation | `go build ./...` and `go vet ./... ./.research/project-exclusion-claude`; commands/exits recorded in the task outcome. |
| Default suite | Runs in a disposable normal Git clone at the exact handover head, with short TMPDIR and `GIT_DIR`/`GIT_WORK_TREE` unset; package/shard commands, exits, limitations and cleanup are recorded in `TASK-261021-wt8ve7_results-r1.md`, the board outcome resource. |

The boundary instrument pairs a reachable exact provider with a denied second
loopback endpoint and no extra hit, readable ordinary bytes with denied
synthetic Keychains bytes and unchanged contents, and a permitted fixture
write with a rejected outside-fixture write whose original bytes remain.
These controls supplement the real product-source negatives; they do not
substitute for them. The default suite excludes hidden research, so the
explicit ungated and gated runs are separately necessary.

## Admission consequence and open scope

This build supplies a promising, measured source-selection control that was
not established at the older CIP evidence pin. A downstream adapter can use
these observations only within the tested build/platform/mode and must retain
approved home delivery. This leaf does not certify a universal strict tuple:
provenance-bound resume, all remaining late/reload paths, other plugin/tool
variants, arbitrary settings and other platforms still need qualification.
Keep unsupported strict entry paths refused rather than treating a green
research harness or a new HOME as whole-session isolation.
