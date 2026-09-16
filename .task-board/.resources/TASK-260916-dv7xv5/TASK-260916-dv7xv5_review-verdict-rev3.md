# TASK-260916-dv7xv5 — independent review rev3

Verdict: **rework / changes_requested**, route **analysis**. One literal-transcript correction remains; no remediation, code changes, or tests requested.

Reviewed current verify-e-findings-rev3.md, rev1/rev2 verdicts, producer rev3 logbook, seven live sibling descriptions, and pinned sources via git show (curator 80483355; curator-agent-launcher b34e1e27). Independently reran all four quoted grep commands in temporary git-archive snapshots. All exited 0 with empty stderr; 3/4 outputs match byte-for-byte. Run goal queried: none (not goal-bound).

## Rev2 correction closures

| Correction | Result |
|---|---|
| 1 E1 citations and sibling | Closed: profile.go:252 is entry, contextresolve.go:489 ordering and :554-575 range selection; :511-515 is no longer range evidence. ioemse carries corrected citations. Minor explanatory typo: rev3 introduction calls profile.go:260 the return, but the return is :259 and :260 its closing brace; correct while updating the artifact. |
| 2 E2 output and caller | Closed: literal four-line output reproduces exactly; envprofile.go:1138-1139 is correctly classified as production warning surfacing, not admission. |
| 3 E3 writes, matches, sibling | Substantively closed: managed.go:887-893 writes seed payloads, skillspec/types.go:98 is dispositioned, and 1i1gfo carries the corrected write site. Literal-output aspect remains open below. |
| 4 Literal outputs and retained bounds | Partially closed: E1/E2/E7 exact, E3 has one missing tab on skillspec/parse.go:697 (artifact two tabs; actual three). All 15 matching locations and content otherwise agree. E5/E6 limits and E7 asymmetry retained. |

## Per-finding evidence check

| Finding | Review and pinned evidence |
|---|---|
| E1 | Agree, static confirmed on the context/profile path. cmd/curator/profile.go:252,292,301-305 and internal/envprofile/envprofile.go:887-897 trace update; internal/contextresolve/contextresolve.go:79-94,489,554-575 selects tag candidates without signer admission; internal/pkgversion/pkgversion.go:286-289 handles latest. Checked the 13 signer-search locations: registry signature envelopes, external-build policy/comments and swiftpm cat-file existence probe do not implement context tag signer admission. |
| E2 | Agree, static confirmed. internal/contextmaterialize/contextmaterialize.go:248-265 includes emitted system modules, called at internal/envprofile/managed.go:1785. internal/contextaudit/contextaudit.go:109-115 blocks Findings; internal/envprofile/envprofile.go:1138-1139 surfaces system warnings. |
| E3 | Agree with static seed verdict. internal/envregistry/envregistry.go:219 declares config.toml, internal/envprofile/managed.go:576-583 reads whole bytes, :887-893 writes payload unchanged. internal/contextmaterialize/mcp.go:231-240 renders separate layer; internal/mcp/mcp.go:230-253 reads home entries; internal/skillspec/types.go:98 is a type comment. Transcript defect does not alter the substantive verdict. |
| E4 | Agree, static confirmed. cmd/curator/umbrella.go:30-38 uses ambient LookPath, :43-63 excludes manager-published locations, :81-91 executes provider. No dynamic project-hook exploit claimed. |
| E5 | Agree within bound. internal/envprofile/switch.go:523-530,539-545 removes pre-existing target links before writes; :694-696 removes before Symlink. :686 still uses plain WriteFile; no race-free or all-writes guarantee. Specification absence is inherited from supplied finding, not independently re-audited here. |
| E6 | Agree, partially confirmed. internal/contextpkg/contextpkg.go:289-305 requires git dependencies; internal/envprofile/managed.go:163-170 loads MCP lock members; internal/contextresolve/contextresolve.go:481-483 enforces a configured nonempty allowlist regardless of root kind. Class parsing :263 is kind-agnostic. internal/envprofile/envprofile.go:596-613,1243-1252 ingests path state; internal/contextstore/contextstore.go:90-131,146-199 checks directory/readability, rejects hardlinks/nonregular entries, but supplies no source owner/mode policy. internal/envprofile/switch.go:760-777,793-797 is provenance, not admission. |
| E7 | Agree, static confirmed for three minors. Launcher internal/axconfig/config.go:58-74 and internal/defaults/defaults.go:83-118 accept regular symlink targets; permission search matches only execution.go:197. internal/fragment/resolve.go:75-80 adds repair unconditionally, persistence conditional on S5. internal/fragment/fragment.go:172-173 and curator internal/envregistry/envregistry.go:192,215,219 plus managed.go:576-583,887-893 establish declared strict/layered MCP asymmetry. No runtime tool behavior tested. |

## Acceptance criteria and empty repository delta

- Outcome: rev3 attached, 7/7 findings, both pins and file:line evidence. Accuracy of the explicitly required literal transcript remains incomplete.
- Siblings: 7/7 live descriptions read through task-board q with description projection: ioemse, 2d9coh, 1i1gfo, 2otjbn, 73a5zg, wgt8vz, 33vuzm. All record verdicts, pins and narrowed scope; E1/E3 citations corrected.
- Empty CR delta is the right repository outcome for this read-only research task. Independently inspected diff --stat between base 23cb9e2aa03527da4d852679a55feee8ed527faa and candidate 28513f24f299350d6a967088a2a2ad19dd8883c4: empty. Story worktree and launcher status are clean. No code/test edits or checkouts performed by this reviewer. This establishes candidate/present-tree state, not every historical operation.
- Architecture fit: research respects resolver, materializer and launcher boundaries; no architecture changes.
- No tests run or passing evidence accepted, as explicitly required. Live checklist unexpectedly contains a checked Tests green item (13), contrary to brief; uncheck as unverified/not applicable. Do not run tests to satisfy it.
- Findings, highlights, source checks, sibling consequences and producer logbook are present. Uncheck item 4 (Implementation matches AC) and item 9 (Fact-checking performed — claims verified, sources cited) pending literal-evidence correction; other applicable checks remain satisfied.

## Required correction

Replace E3 literal output with the exact independently captured block below (preserve three tabs after internal/skillspec/parse.go:697:). Also fix the introductory explanation of profile.go:260 to closing brace / return at :259. Preserve all substantive verdicts and siblings. Resubmit research only. Checklist Tests green must remain explicitly not applicable rather than attesting a suite run.

## Independent E3 transcript

```text
grep -rn 'mcp_servers' internal cmd --include='*.go' | grep -v _test.go
internal/contextmaterialize/mcp.go:231:// codexMCP renders the TOML layer document whose only table is mcp_servers,
internal/contextmaterialize/mcp.go:232:// one [mcp_servers.<name>] table per server in sorted name order, keys in
internal/contextmaterialize/mcp.go:240:		fmt.Fprintf(&out, "[mcp_servers.%s]\n", server.Name)
internal/skillspec/types.go:98:// McpServer is a dependencies.mcp_servers entry (Spec §5.8).
internal/skillspec/parse.go:514:	if _, present := obj["mcp_servers"]; present && schema < 5 {
internal/skillspec/parse.go:515:		return nil, nil, nil, verr.New("dependencies.mcp_servers", "requires schema_version 5")
internal/skillspec/parse.go:522:		allowed["mcp_servers"] = true
internal/skillspec/parse.go:532:	mcpServers, err := parseMcpServers(obj["mcp_servers"])
internal/skillspec/parse.go:692:		return nil, verr.New("dependencies.mcp_servers", "must be an object")
internal/skillspec/parse.go:695:		label := "dependencies.mcp_servers." + name
internal/skillspec/parse.go:697:			return nil, verr.New("dependencies.mcp_servers", "MCP server names must be non-empty strings")
internal/mcp/mcp.go:253:		servers = data["mcp_servers"]
internal/marker/marker.go:173:	McpServers         map[string][]string   `json:"mcp_servers,omitempty"`
internal/marker/marker.go:253:		"git", "requirements", "mcp_servers", "attestation", "activation", "requirers", "substituted",
internal/marker/marker.go:300:	if _, present := raw["mcp_servers"]; present {
exit=0; stderr=''
```
