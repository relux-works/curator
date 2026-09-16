# TASK-260916-dv7xv5 — review verdict rev2

Verdict: **rework / changes_requested**, route **analysis**. Research corrections only; no remediation requested or started.

Reviewed verify-e-findings-rev2.md and rev1 verdict, all seven sibling READMEs, and pinned source via git show: curator **80483355**, launcher **b34e1e27**. Independently reran all four rev2 grep commands against temporary git-archive snapshots of those pins (Markdown table pipe escaping decoded). No checkout, code edits, or tests. All commands exited 0, stderr empty. No test-suite evidence accepted; **Tests green: not applicable**. Run goal queried: none, not goal-bound.

## Rev1 closure

| Required correction | Closure |
|---|---|
| E1 update entry point and signer scope | Substantially closed: correct function and UpdateWithPolicy path, signer claim scoped to context resolution. Citation still needs correction: cmdProfileUpdate starts at profile.go:252, not :260. Range selection citation :511-515 is wrong (exact-commit branch); use :554-575. |
| E1 actual signer grep and dispositions | Closed: 13 locations reproduce and all are classified; these do not supply context-source tag signer admission. |
| E2 actual search output | **Not closed**: replacement command returns four lines, including envprofile.go:1138. The assertion “contextaudit.go declarations only” is false and omits a production warning caller. |
| E3 actual search output and dispositions | **Partially closed**: 15 matches reproduce, but internal/skillspec/types.go:98 is omitted from the claimed file list/disposition. Literal output is not supplied. |
| E7 strict-MCP asymmetry | Closed: curator registry :192/:215/:219 and launcher fragment.go:172-173 support the declared asymmetry, combined with whole-file seeding. Static channel/seed evidence, not a runtime launch observation. |
| E5/E6 bounds | Closed: E5 explicitly limits mitigation to pre-existing target links and acknowledges races/plain WriteFile. E6 separates git-only MCP dependencies from system-module admission and bounds directory ownership/permissions claims. |
| Seven sibling descriptions | Closed for recording verdicts/pins/narrowed scope: 7/7 now have verification paragraphs. Correct the propagated E1/E3 citations alongside artifact corrections. |

## Per-finding evidence check

| Finding | Review |
|---|---|
| E1 | Agree with scoped static confirmed verdict. cmd/curator/profile.go:252,292,301-305 traces update without delta confirmation; envprofile.go:887-897 forwards to updateLocked. contextresolve.go:489 sorts candidates and :554-575 chooses highest satisfying range; :511-515 instead annotates an exact commit. pkgversion.go:286-289 supports latest. Checked all 13 signer-search sites; registry envelopes, build policy and swiftpm existence probe are distinct from tag-signature admission. Correct citations in artifact and E1 sibling. |
| E2 | Agree with static transitive-admission verdict, disagree with reported grep evidence. contextmaterialize.go:248-265 loops emitted members and applicable system modules; managed.go:1785 calls it. contextaudit.go:109-115 only blocks Findings. envprofile.go:1138-1139 is the omitted production call that appends SystemModules as warnings. Record all four search lines and disposition this caller as warning surfacing, not admission. This is materially relevant evidence, not an unrelated match. |
| E3 | Agree with whole-file seed retention verdict, disagree with write-site citation and completeness of search evidence. envregistry.go:219 selects config.toml; managed.go:576-583 reads whole bytes. **Actual seed writes are managed.go:887-893**, looping seeds.files and calling WriteFile with payload. The cited :378-388 handles contextmaterialize.Referenced root-context documents and proves nothing about seed writes. contextmaterialize/mcp.go:231-240 renders the separate layer; mcp/mcp.go:230-253 reads home entries. Add skillspec/types.go:98 (McpServer type comment). Correct artifact and E3 sibling. |
| E4 | Agree. umbrella.go:30-38 uses LookPath; :43-63 excludes manager-published paths, :81-91 executes provider. Static acceptance of an otherwise executable ambient-PATH provider; no dynamic project-hook exploit reproduced. |
| E5 | Agree within explicit bound. switch.go:523-530 and :539-545 remove before writes; :694-696 removes before Symlink. :686 remains plain WriteFile. No race-free O_NOFOLLOW or all-write coverage claimed. Specification absence is inherited from supplied finding, not independently established by this implementation-only review. |
| E6 | Agree with partially-confirmed split. contextpkg.go:289-305 requires git dependencies; managed.go:163-170 loads MCP-kind lock members; contextresolve.go:481-483 applies a configured nonempty allowlist. class parsing :263 is kind-agnostic. envprofile.go:596-613 and :1243-1252 ingest via EnsureState; contextstore.go:90-131 checks directory/readability and snapshots, :146-207 rejects nonregular entries/hardlinks but supplies no source ownership/mode policy. Do not generalize the missing ownership boundary into absence of all filesystem validation. switch.go:760-777/:793-797 is provenance only. |
| E7 | Agree with static three-part verdict. Launcher axconfig/config.go:58-74 and defaults/defaults.go:83-118 follow regular-file symlink targets, permission grep only matches execution.go:197. fragment/resolve.go:75-80 appends repair unconditionally; persistence remains conditional on S5 store tampering. fragment/fragment.go:172-173 and curator envregistry.go:192,215,219 plus managed.go:576-583/:887-893 establish declared strict/layered channel asymmetry. defaults/lineup.go:271-290 prints model/effort only; ancillary provider logging is not a substitute for strict-MCP evidence. |

## Acceptance criteria and checklist

- Outcome structure: 7/7 rows, both pins, file:line evidence present. Rev2 is the current artifact, rev1 history. Accuracy is not yet accepted because of the E1/E2/E3 corrections above.
- Sibling descriptions: read ioemse, 2d9coh, 1i1gfo, 2otjbn, 73a5zg, wgt8vz, 33vuzm READMEs under EPIC-260910-2hw1xb. All now carry pins and verdicts, with E5 limits and E6 split. Item 2 can be checked for recorded verdicts; E1/E3 citation repair remains required by evidence accuracy.
- Read-only: story worktree clean, launcher clean, curator control checkout changes confined to .task-board metadata/resources at observation. No product/test modifications in this review. This is present-tree evidence, not proof of every past producer operation.
- Architecture: research follows existing resolver/materializer/launcher boundaries; no product architecture change.
- Implementation matches AC remains unchecked until evidence corrections. Tests green remains unchecked as not applicable, explicitly no suites run.
- Rework outcome attached before analysis transition. No human-only blocker.

## Required producer corrections

1. Correct E1 entry line to profile.go:252 and range selection to contextresolve.go:554-575 (include :489 for ordering). Propagate to E1 sibling.
2. Replace E2 “declarations only” with literal four-line output and classify envprofile.go:1138-1139 as production warning surfacing. Preserve confirmed verdict.
3. Replace E3 seed-write citation :378-388 with :887-893; include literal 15-line output and disposition skillspec/types.go:98. Propagate seed-write citation to E3 sibling.
4. Supply the literal output blocks requested by the brief instead of compressed/inaccurate output paraphrases; keep E5/E6 limits and E7 asymmetry. Resubmit research for review without code changes/tests/remediation.

## Independent command transcripts

curator @ 80483355

```sh
grep -rnE 'verify-tag|verify-commit|allowed_signers|gpg|ssh-keygen|Signature|signer' internal cmd --include='*.go' | grep -v _test.go | cut -d: -f1,2 | sort -u
```
```text
internal/buildrepo/pipeline.go:177
internal/buildrepo/pipeline.go:33
internal/buildrepo/pipeline.go:34
internal/install/external.go:24
internal/install/install.go:57
internal/registry/registry.go:181
internal/registry/registry.go:198
internal/registry/registry.go:212
internal/registry/registry.go:223
internal/registry/registry.go:245
internal/registry/registry.go:260
internal/registry/snapshot.go:478
internal/swiftpmsource/git.go:514
exit=0; stderr=''
```

curator @ 80483355

```sh
grep -rn 'context_system_module_transitive\|SystemModules' internal --include='*.go' | grep -v _test.go
```
```text
internal/contextaudit/contextaudit.go:104:	SystemModules []SystemModule
internal/contextaudit/contextaudit.go:251:// SystemModules lists every class: system module of a context manifest as
internal/contextaudit/contextaudit.go:253:func SystemModules(packageName string, manifest *contextpkg.Manifest) []SystemModule {
internal/envprofile/envprofile.go:1138:				for _, system := range contextaudit.SystemModules(resolved.Name, manifest) {
exit=0; stderr=''
```

curator @ 80483355

```sh
grep -rn 'mcp_servers' internal cmd --include='*.go' | grep -v _test.go
```
```text
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

curator-agent-launcher @ b34e1e27

```sh
grep -rnE 'Sys\(\)|Uid|Gid|Perm\(\)' --include='*.go' | grep -v _test.go
```
```text
./internal/execution/execution.go:197:		if !info.Mode().IsRegular() || info.Mode().Perm()&0111 == 0 {
exit=0; stderr=''
```
