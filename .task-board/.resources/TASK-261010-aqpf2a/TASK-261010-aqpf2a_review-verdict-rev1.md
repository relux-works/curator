# Review verdict — TASK-261010-aqpf2a, Change Request rev1

**Verdict: CHANGES REQUESTED (one required change). The content of `.research/261010_pi-opencode-tool-lockdown.md` is otherwise ACCEPTED.**

## Required change (the only one)

Drop the `LOGBOOK.md` hunk; keep the `.research` file byte-identical.

The rev1 candidate (tree `c3285cf15b77bedfabbd7a458ad533ce688c9114` against base `5ec599f0b7eddeb3403fe89104b7b800fd0a55db`) changes two paths: `.research/261010_pi-opencode-tool-lockdown.md` and `LOGBOOK.md` (+7 lines). Campaign rule: producers never edit `LOGBOOK.md`; the Change Request must contain only the `.research` file.

The next review only has to confirm that delta: the candidate diff is exactly one path, and that path's blob is byte-identical to the rev1 blob (506 lines).

## Items 2–5 — all hold

2. **Structure against the brief.** The verdict table is first. Sections A–F follow, with a source and verification record at the end. Project identity is evidenced: Pi is `earendil-works/pi` (formerly `badlogic/pi-mono`), package `@earendil-works/pi-coding-agent`. OpenCode is `anomalyco/opencode`, and the archived Go `opencode-ai/opencode` is explicitly excluded. Release tags and commits are given for Pi 1.1.0 and 0.84.2 and for OpenCode 1.18.35. Draft qualification rows and the hosted-runner plan are present. Unsupported, conditional and candidate verdicts are separated and hedged correctly. No cell is claimed as qualified.
3. **Citation sample (16 claims, read-only `gh api`, at the cited commits).** All matched.
   - §B1 Pi, 1.1.0 `docs/cli.md` and `src/core/agent-session.ts` (`abe508e1…`):
     - `--tools` takes a list with `*` patterns.
     - `--exclude-tools` applies after all other selection, MCP included.
     - `--no-tools` disables built-in, extension, custom and MCP tools.
     - `--no-builtin-tools` retains extension and custom tools.
     - `pi -ne -e builtin:mcp` keeps only built-in MCP.
     - MCP tools stay unless an entry starts with `mcp__`; `allowedToolNames` handling at `agent-session.ts:502-507`; `baseToolsOverride` exists at `agent-session.ts:294`.
     - `src/core/mcp-servers.ts` shows the `mcp__<server>__<tool>` naming and `-`→`_` replacement.
   - §A2 Pi tool registries: the 1.1.0 `ToolName` union has eight names including `powershell`; the 0.84.2 `allToolNames` set (`914cf147…`) has seven. `src/extensions/index.ts` lists the llama.cpp, codemode, tool-search and mcp built-ins.
   - §B2 OpenCode (`53d1eabb…`):
     - `permission/index.ts` uses `findLast` and `disabled()`, which hides a tool only when the last matching rule is pattern `*` with deny. `edit`, `write` and `apply_patch` map to the `edit` permission.
     - `session/llm/request.ts:210-215` `resolveTools` filters by `Permission.disabled` and `user.tools[k] !== false`.
     - `core/src/npm.ts:147-151` `install` returns when the directory is not writable.
     - `mcp/catalog.ts:117-119` `toolName = sanitize(client)+"_"+sanitize(name)`.
     - `core/src/v1/config/mcp.ts` has local `command: string[]` and `enabled`.
     - `tool/shell/id.ts` keeps `ToolID="bash"`.
   - The §A1 tag-to-commit claims were re-resolved by API: Pi `v1.1.0` is `abe508e1…`, and OpenCode `v1.18.35` is `53d1eabb…`.
   - The §A4 count of 17 fixed OpenCode tool IDs reproduces: `tool/registry.ts` has 16 `Tool.init` entries plus the conditional `execute`. No citation drift was found in the sample.
4. **No execution on this host.** The report states that no harness was installed or run, and there was no test suite, login or credential probe. The only local checks listed are a document and source audit, `git diff --check`, and one nonzero `--no-index` result that is explicitly not counted as a gate. This is consistent. This reviewer also ran nothing but read-only API reads.
5. **Hygiene.** There are no home or personal paths, e-mail addresses, host names or employer names in the report or the patch. The only path is the neutral example `/opt/curator/qualified/relay/swarma-remote-tools`, labelled as not a host path. Repository names are public upstream names.

```verdict-findings
{
  "findings": [
    {
      "id": "producer-edits-logbook",
      "row": "Change Request scope: producer-owned paths",
      "invariant": "Producers never edit LOGBOOK.md; the Change Request contains only .research/261010_pi-opencode-tool-lockdown.md",
      "mechanism": "Candidate tree c3285cf15b77bedfabbd7a458ad533ce688c9114 adds a 7-line entry at the top of LOGBOOK.md (hunk @@ -3,6 +3,13 @@).",
      "reproductions": [
        {
          "test_file": "LOGBOOK.md",
          "command": "git diff --stat 5ec599f0b7eddeb3403fe89104b7b800fd0a55db c3285cf15b77bedfabbd7a458ad533ce688c9114 -- LOGBOOK.md",
          "expected_failure": "output lists LOGBOOK.md with 7 insertions; the empty output required by the campaign rule is not produced",
          "pinned_blobs": ["sha256:cd43b27578ff417f730bd06a756ca13201ebd1a426a2f96709f31c2ac040caef"]
        }
      ],
      "severity": "regression",
      "severity_reason": "P1 process-rule violation (campaign rule), trivially fixable; content itself unaffected",
      "repeat-of": "none"
    }
  ],
  "notes": [],
  "surface_results": [
    {"row": "verdict table first, sections A-F", "result": "held"},
    {"row": "project identity evidence", "result": "held"},
    {"row": "citation sample B/C2/C4", "result": "held"},
    {"row": "no execution on host", "result": "held"},
    {"row": "hygiene (secrets, paths, hosts, employers)", "result": "held"},
    {"row": "Change Request scope", "result": "broken"}
  ],
  "free_hunt": []
}
```
